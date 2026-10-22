package sentry

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
)

// Sentryクライアントが本来ネットワーク送信するイベントをすべて収集する
// テスト用Transport。
type slogCaptureTransport struct {
	mu     sync.Mutex
	events []*sentry.Event
}

func (t *slogCaptureTransport) Configure(_ sentry.ClientOptions)        {}
func (t *slogCaptureTransport) Flush(_ time.Duration) bool              { return true }
func (t *slogCaptureTransport) FlushWithContext(_ context.Context) bool { return true }
func (t *slogCaptureTransport) Close()                                  {}
func (t *slogCaptureTransport) SendEventWithContext(_ context.Context, e *sentry.Event) {
	t.SendEvent(e)
}

func (t *slogCaptureTransport) SendEvent(event *sentry.Event) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = append(t.events, event)
}

func (t *slogCaptureTransport) Events() []*sentry.Event {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]*sentry.Event, len(t.events))
	copy(out, t.events)
	return out
}

// テストごとに独立したHub + slogCaptureTransportを作る。グローバルHub
// には一切触らない。
func newSlogTestHub(t *testing.T) (*sentry.Hub, *slogCaptureTransport) {
	t.Helper()
	transport := &slogCaptureTransport{}
	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:       "https://public@example.com/1",
		Transport: transport,
	})
	if err != nil {
		t.Fatalf("sentry.NewClient()のエラー = %v", err)
	}
	return sentry.NewHub(client, sentry.NewScope()), transport
}

// 本番と同じbeforeSendフックをテスト用Sentry Clientに組み込む。
// slog → アプリケーションのイベントハンドラー → beforeSendのパイプラインを
// End-to-Endで検証できるため、ハンドラーのタグマッピング挙動が変わった場合に
// テストで検知できる。
func newSlogTestHubWithBeforeSend(t *testing.T) (*sentry.Hub, *slogCaptureTransport) {
	t.Helper()
	transport := &slogCaptureTransport{}
	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:        "https://public@example.com/1",
		Transport:  transport,
		BeforeSend: beforeSend,
	})
	if err != nil {
		t.Fatalf("sentry.NewClient()のエラー = %v", err)
	}
	return sentry.NewHub(client, sentry.NewScope()), transport
}

// recordingHandlerが捕捉した1レコード。Withチェーンで蓄積された属性も
// attrsに平坦化してテストで検証しやすくする。
type recordedLog struct {
	level slog.Level
	msg   string
	attrs map[string]string
}

// 受け取ったslog.Recordをすべて保持するテスト用ハンドラー。
// WithAttrs / WithGroupで派生したハンドラーも同じバッファを共有するため、
// テスト側はルートハンドラーから全レコードを取り出せる。
type recordingHandler struct {
	mu      *sync.Mutex
	records *[]recordedLog
	attrs   []slog.Attr
}

func newRecordingHandler() *recordingHandler {
	return &recordingHandler{
		mu:      &sync.Mutex{},
		records: &[]recordedLog{},
	}
}

func (h *recordingHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *recordingHandler) Handle(_ context.Context, r slog.Record) error {
	attrs := make(map[string]string)
	for _, a := range h.attrs {
		attrs[a.Key] = a.Value.String()
	}
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.String()
		return true
	})
	h.mu.Lock()
	defer h.mu.Unlock()
	*h.records = append(*h.records, recordedLog{
		level: r.Level,
		msg:   r.Message,
		attrs: attrs,
	})
	return nil
}

func (h *recordingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := &recordingHandler{
		mu:      h.mu,
		records: h.records,
	}
	next.attrs = append(next.attrs, h.attrs...)
	next.attrs = append(next.attrs, attrs...)
	return next
}

func (h *recordingHandler) WithGroup(_ string) slog.Handler { return h }

func (h *recordingHandler) Snapshot() []recordedLog {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]recordedLog, len(*h.records))
	copy(out, *h.records)
	return out
}

func TestSlogHandler_ErrorIsCapturedToSentry(t *testing.T) {
	t.Parallel()

	hub, transport := newSlogTestHub(t)
	base := newRecordingHandler()
	logger := slog.New(NewSlogHandler(base))

	ctx := sentry.SetHubOnContext(context.Background(), hub)
	logger.ErrorContext(ctx, "operation failed", "user_id", "u123", "error", errors.New("disk full"))

	hub.Flush(2 * time.Second)
	events := transport.Events()
	if len(events) != 1 {
		t.Fatalf("Sentryイベントの件数 = %d、期待値 = 1", len(events))
	}
	got := events[0]
	if got.Message != "operation failed" {
		t.Errorf("event.Message = %q、期待値 = %q", got.Message, "operation failed")
	}
	if got.Tags["user_id"] != "u123" {
		t.Errorf("event.Tags[user_id] = %q、期待値 = %q", got.Tags["user_id"], "u123")
	}
	if len(got.Exception) != 1 || got.Exception[0].Value != "disk full" {
		t.Errorf("例外のvalueの期待値 = %q、実測値 = %+v", "disk full", got.Exception)
	}

	recs := base.Snapshot()
	if len(recs) != 1 {
		t.Fatalf("ベースレコードの件数 = %d、期待値 = 1", len(recs))
	}
	if recs[0].level != slog.LevelError || recs[0].msg != "operation failed" {
		t.Errorf("ベースレコード = %+v、期待値 = Error/operation failed", recs[0])
	}
}

func TestSlogHandler_InfoIsNotCapturedToSentry(t *testing.T) {
	t.Parallel()

	hub, transport := newSlogTestHub(t)
	base := newRecordingHandler()
	logger := slog.New(NewSlogHandler(base))

	ctx := sentry.SetHubOnContext(context.Background(), hub)
	logger.InfoContext(ctx, "request received", "path", "/")

	hub.Flush(2 * time.Second)
	if got := len(transport.Events()); got != 0 {
		t.Errorf("InfoレベルのSentryイベントの件数 = %d、期待値 = 0", got)
	}

	recs := base.Snapshot()
	if len(recs) != 1 {
		t.Fatalf("ベースレコードの件数 = %d、期待値 = 1", len(recs))
	}
	if recs[0].level != slog.LevelInfo || recs[0].msg != "request received" {
		t.Errorf("ベースレコード = %+v、期待値 = Info/request received", recs[0])
	}
}

func TestSlogHandler_BaseHandlerReceivesAllLevels(t *testing.T) {
	t.Parallel()

	// baseハンドラーは構造化ログの出力先なので、Sentryが取り上げるか
	// どうかに関わらず全レベルが届く必要がある。これにより、開発者は標準
	// エラー出力でInfo/Warnを引き続き追える一方、SentryはError以上に
	// 集中できる。
	hub, transport := newSlogTestHub(t)
	base := newRecordingHandler()
	logger := slog.New(NewSlogHandler(base))

	ctx := sentry.SetHubOnContext(context.Background(), hub)
	logger.DebugContext(ctx, "debug msg")
	logger.InfoContext(ctx, "info msg")
	logger.WarnContext(ctx, "warn msg")
	logger.ErrorContext(ctx, "error msg")

	hub.Flush(2 * time.Second)
	if got := len(transport.Events()); got != 1 {
		t.Errorf("Sentryイベントの件数 = %d、期待値 = 1 (Errorのみ)", got)
	}

	recs := base.Snapshot()
	want := []struct {
		level slog.Level
		msg   string
	}{
		{slog.LevelDebug, "debug msg"},
		{slog.LevelInfo, "info msg"},
		{slog.LevelWarn, "warn msg"},
		{slog.LevelError, "error msg"},
	}
	if len(recs) != len(want) {
		t.Fatalf("ベースレコードの件数の期待値 = %d、実測値 = %d", len(want), len(recs))
	}
	for i, w := range want {
		if recs[i].level != w.level || recs[i].msg != w.msg {
			t.Errorf("ベースレコード[%d] = (%v, %q)、期待値 = (%v, %q)", i, recs[i].level, recs[i].msg, w.level, w.msg)
		}
	}
}

func TestSlogHandler_DropsReverseProxySourceEventsEndToEnd(t *testing.T) {
	t.Parallel()

	// End-to-Endの確認: SourceAttrKey=ReverseProxySourceを付けた
	// slog.ErrorはbeforeSendで破棄され、transportに到達しないこと。
	// アプリケーションのイベントハンドラーがslog属性をevent.Tagsに乗せる
	// 挙動 (本判定の前提) が将来変わった場合に本テストで検知できる。
	hub, transport := newSlogTestHubWithBeforeSend(t)
	base := newRecordingHandler()
	logger := slog.New(NewSlogHandler(base))

	ctx := sentry.SetHubOnContext(context.Background(), hub)
	logger.ErrorContext(ctx, "rails proxy failure",
		SourceAttrKey, ReverseProxySource,
		"error", errors.New("upstream 502"),
	)

	hub.Flush(2 * time.Second)
	if got := len(transport.Events()); got != 0 {
		t.Errorf("source=%sのイベント件数 = %d、期待値 = 0 (beforeSendで捨てられること)", ReverseProxySource, got)
	}

	// baseのテキストハンドラーには通常通り届くこと (デバッグ用に標準
	// エラー出力には残す)。
	recs := base.Snapshot()
	if len(recs) != 1 {
		t.Fatalf("ベースレコードの件数 = %d、期待値 = 1", len(recs))
	}
	if recs[0].attrs[SourceAttrKey] != ReverseProxySource {
		t.Errorf("ベースレコードのattrs[%s] = %q、期待値 = %q", SourceAttrKey, recs[0].attrs[SourceAttrKey], ReverseProxySource)
	}
}

func TestSlogHandler_KeepsOtherSourceEventsEndToEnd(t *testing.T) {
	t.Parallel()

	// TestSlogHandler_DropsReverseProxySourceEventsEndToEndの対称ケース。
	// SourceAttrKeyに乗っていても、値がReverseProxySource以外のときは
	// dropせずSentryに届くこと (フィルタが意図せず広く効いてしまっていない
	// ことの担保)。
	hub, transport := newSlogTestHubWithBeforeSend(t)
	base := newRecordingHandler()
	logger := slog.New(NewSlogHandler(base))

	ctx := sentry.SetHubOnContext(context.Background(), hub)
	logger.ErrorContext(ctx, "other subsystem error",
		SourceAttrKey, "other_subsystem",
		"error", errors.New("real failure"),
	)

	hub.Flush(2 * time.Second)
	events := transport.Events()
	if len(events) != 1 {
		t.Fatalf("source=otherのイベント件数 = %d、期待値 = 1 (捨てられずSentryへ届くこと)", len(events))
	}
	if events[0].Tags[SourceAttrKey] != "other_subsystem" {
		t.Errorf("event.Tags[%s] = %q、期待値 = %q", SourceAttrKey, events[0].Tags[SourceAttrKey], "other_subsystem")
	}
}

func TestSlogHandler_MasksSensitiveTagsEndToEnd(t *testing.T) {
	t.Parallel()

	// End-to-Endの確認: slog属性として載せたメールアドレスはbeforeSend
	// でマスクされてからtransportに届くこと。baseハンドラーには元の値の
	// レコードが届き、標準エラー出力で伏せるのはNewBaseHandlerの書き出し
	// (TestNewBaseHandler_MasksSensitiveAttrs) が受け持つこと。
	hub, transport := newSlogTestHubWithBeforeSend(t)
	base := newRecordingHandler()
	logger := slog.New(NewSlogHandler(base))

	ctx := sentry.SetHubOnContext(context.Background(), hub)
	logger.ErrorContext(ctx, "email send failure",
		"email", "user@example.com",
		"user_id", "u123",
		"error", errors.New("send failure"),
	)

	hub.Flush(2 * time.Second)
	events := transport.Events()
	if len(events) != 1 {
		t.Fatalf("Sentryイベントの件数 = %d、期待値 = 1", len(events))
	}
	tags := events[0].Tags
	if tags["email"] != "[FILTERED]" {
		t.Errorf("event.Tags[email] = %q、期待値 = %q", tags["email"], "[FILTERED]")
	}
	if tags["user_id"] != "u123" {
		t.Errorf("event.Tags[user_id] = %q、期待値 = %q", tags["user_id"], "u123")
	}

	recs := base.Snapshot()
	if len(recs) != 1 {
		t.Fatalf("ベースレコードの件数 = %d、期待値 = 1", len(recs))
	}
	if recs[0].attrs["email"] != "user@example.com" {
		t.Errorf("ベースレコードのattrs[email] = %q、期待値 = %q", recs[0].attrs["email"], "user@example.com")
	}
}

func TestNewBaseHandler_MasksSensitiveAttrs(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(newBaseHandler(&buf))
	logger.Info("確認コードを再送信しました",
		"email", "user@example.com",
		"token_digest", "digest-value",
		"user_id", "u123",
		"ip_address", "203.0.113.10",
	)
	// WithAttrsで先に載せた属性とグループ内の属性も伏せること。
	logger.With("sign_up_email", "group@example.com").WithGroup("form").Info("フォームを受け付けました",
		"password", "plain-password",
	)

	out := buf.String()
	for _, secret := range []string{"user@example.com", "digest-value", "group@example.com", "plain-password"} {
		if strings.Contains(out, secret) {
			t.Errorf("標準エラー出力に %qが含まれている: %s", secret, out)
		}
	}
	for _, want := range []string{
		"email=[FILTERED]",
		"token_digest=[FILTERED]",
		"sign_up_email=[FILTERED]",
		"form.password=[FILTERED]",
		"user_id=u123",
		"ip_address=203.0.113.10",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("標準エラー出力に %qが含まれていない: %s", want, out)
		}
	}
}

func TestSlogHandler_WithAttrs_PropagatesToBothHandlers(t *testing.T) {
	t.Parallel()

	// Logger.Withで追加した属性はbaseとSentryの両方に届く必要がある。
	// さもないとリクエスト単位の情報 (例: request_id) が片方に欠落し、
	// 標準エラー出力のログとSentryイベントを突き合わせられなくなる。
	hub, transport := newSlogTestHub(t)
	base := newRecordingHandler()
	logger := slog.New(NewSlogHandler(base)).With("request_id", "r123")

	ctx := sentry.SetHubOnContext(context.Background(), hub)
	logger.ErrorContext(ctx, "boom", "user_id", "u456")

	hub.Flush(2 * time.Second)
	events := transport.Events()
	if len(events) != 1 {
		t.Fatalf("Sentryイベントの件数 = %d、期待値 = 1", len(events))
	}
	tags := events[0].Tags
	if tags["request_id"] != "r123" {
		t.Errorf("event.Tags[request_id] = %q、期待値 = %q", tags["request_id"], "r123")
	}
	if tags["user_id"] != "u456" {
		t.Errorf("event.Tags[user_id] = %q、期待値 = %q", tags["user_id"], "u456")
	}

	recs := base.Snapshot()
	if len(recs) != 1 {
		t.Fatalf("ベースレコードの件数 = %d、期待値 = 1", len(recs))
	}
	if recs[0].attrs["request_id"] != "r123" {
		t.Errorf("ベースレコードのattrs[request_id] = %q、期待値 = %q", recs[0].attrs["request_id"], "r123")
	}
	if recs[0].attrs["user_id"] != "u456" {
		t.Errorf("ベースレコードのattrs[user_id] = %q、期待値 = %q", recs[0].attrs["user_id"], "u456")
	}
}

func TestSlogHandler_GroupedAttrsAndBranches(t *testing.T) {
	t.Parallel()

	// グループの展開方法が変わっても、機密値のマスクと派生ロガー間の属性分離を保つ。
	hub, transport := newSlogTestHubWithBeforeSend(t)
	ctx := sentry.SetHubOnContext(context.Background(), hub)
	root := slog.New(NewSlogHandler(newRecordingHandler())).With("request_id", "r1")
	grouped := root.WithGroup("request").With("method", "POST")
	left := grouped.WithGroup("left").With("id", "a")
	right := grouped.WithGroup("right").With("id", "b")

	var wg sync.WaitGroup
	for _, logger := range []*slog.Logger{left, right} {
		wg.Go(func() {
			logger.ErrorContext(ctx, "失敗",
				slog.Group("", slog.String("email", "private@example.com")),
				slog.Group("payload", slog.String("token", "secret"), slog.Int("count", 2)),
				slog.Attr{},
			)
		})
	}
	wg.Wait()

	events := transport.Events()
	if len(events) != 2 {
		t.Fatalf("Sentryイベントの件数 = %d、期待値 = 2", len(events))
	}
	seenIDs := make(map[string]bool, len(events))
	for _, event := range events {
		prefix, other := "request.left.", "request.right."
		wantID := "a"
		if event.Tags[prefix+"id"] == "" {
			prefix, other = other, prefix
			wantID = "b"
		}
		if seenIDs[wantID] {
			t.Errorf("派生ロガーのイベントが重複: id = %q", wantID)
		}
		seenIDs[wantID] = true
		wantTags := map[string]string{
			"request_id":             "r1",
			"request.method":         "POST",
			prefix + "id":            wantID,
			prefix + "email":         maskedValue,
			prefix + "payload.token": maskedValue,
			prefix + "payload.count": "2",
		}
		for key, want := range wantTags {
			if event.Tags[key] != want {
				t.Errorf("タグ[%s] = %q、期待値 = %q", key, event.Tags[key], want)
			}
		}
		if len(event.Tags) != len(wantTags) {
			t.Errorf("タグの件数 = %d、期待値 = %d: %v", len(event.Tags), len(wantTags), event.Tags)
		}
		if _, ok := event.Tags[other+"id"]; ok {
			t.Errorf("別のハンドラーの属性が混入: %v", event.Tags)
		}
	}
}

type mutableLogValue struct{ value string }

func (v *mutableLogValue) LogValue() slog.Value { return slog.StringValue(v.value) }

func TestSlogHandler_WithAttrsResolvesValues(t *testing.T) {
	t.Parallel()

	// LogValuerの値はslog標準のハンドラーと同じくLogger.Withの時点で解決する。
	hub, transport := newSlogTestHub(t)
	ctx := sentry.SetHubOnContext(context.Background(), hub)
	value := &mutableLogValue{value: "追加時"}
	logger := slog.New(NewSlogHandler(newRecordingHandler())).With("state", value)
	value.value = "送信時"
	logger.ErrorContext(ctx, "失敗", "current", value)

	events := transport.Events()
	if len(events) != 1 {
		t.Fatalf("Sentryイベントの件数 = %d、期待値 = 1", len(events))
	}
	if events[0].Tags["state"] != "追加時" || events[0].Tags["current"] != "送信時" {
		t.Errorf("属性の解決時点が異なる: %v", events[0].Tags)
	}
}

func TestSlogHandler_DropsCanceledErrorsEndToEnd(t *testing.T) {
	t.Parallel()

	// slog経由のイベントでも元のエラーをヒントで渡し、beforeSendのshouldDropErrorで
	// 中断エラーを破棄する。
	for _, key := range []string{"error", "err"} {
		t.Run(key, func(t *testing.T) {
			t.Parallel()

			hub, transport := newSlogTestHubWithBeforeSend(t)
			ctx := sentry.SetHubOnContext(context.Background(), hub)
			logger := slog.New(NewSlogHandler(newRecordingHandler()))

			logger.ErrorContext(ctx, "処理中断", key, fmt.Errorf("中断: %w", context.Canceled))
			if got := len(transport.Events()); got != 0 {
				t.Fatalf("中断エラーのイベントは破棄されるべき: Sentryイベントの件数 = %d", got)
			}

			logger.ErrorContext(ctx, "失敗", key, errors.New("障害"))
			events := transport.Events()
			if len(events) != 1 {
				t.Fatalf("Sentryイベントの件数 = %d、期待値 = 1", len(events))
			}
			if len(events[0].Exception) != 1 || events[0].Exception[0].Value != "障害" {
				t.Errorf("例外 = %+v、期待値 = 障害", events[0].Exception)
			}
			if _, ok := events[0].Tags[key]; ok {
				t.Errorf("例外にしたエラーがタグにも追加された: %v", events[0].Tags)
			}
		})
	}
}

func TestSlogHandler_ErrorAttrsWithoutErrorValue(t *testing.T) {
	t.Parallel()

	// error型でないerror属性はタグに残し、例外にするのは最初のエラーだけにする。
	hub, transport := newSlogTestHub(t)
	ctx := sentry.SetHubOnContext(context.Background(), hub)
	logger := slog.New(NewSlogHandler(newRecordingHandler()))
	logger.ErrorContext(ctx, "失敗", "error", "文字列のエラー", "err", errors.New("一つ目"), "error", errors.New("二つ目"))

	events := transport.Events()
	if len(events) != 1 {
		t.Fatalf("Sentryイベントの件数 = %d、期待値 = 1", len(events))
	}
	if events[0].Tags["error"] != "文字列のエラー" {
		t.Errorf("event.Tags[error] = %q、期待値 = %q", events[0].Tags["error"], "文字列のエラー")
	}
	if len(events[0].Exception) != 1 || events[0].Exception[0].Value != "一つ目" {
		t.Errorf("例外 = %+v、期待値 = 一つ目のみ", events[0].Exception)
	}
}

// countingLogValueは無効なSentryで属性を解決しないことを確かめるため、解決回数を記録する。
type countingLogValue struct{ calls int }

func (v *countingLogValue) LogValue() slog.Value {
	v.calls++
	return slog.StringValue("解決済み")
}

func TestSlogHandler_DisabledHub(t *testing.T) {
	t.Parallel()

	// Sentryが未初期化 (noop client) のとき、基底ログを残しつつイベントの組み立てを省く。
	base := newRecordingHandler()
	value := &countingLogValue{}
	ctx := sentry.SetHubOnContext(context.Background(), sentry.NewHub(nil, sentry.NewScope()))
	slog.New(NewSlogHandler(base)).ErrorContext(ctx, "失敗", "state", value)

	if got := len(base.Snapshot()); got != 1 {
		t.Fatalf("ベースレコードの件数 = %d、期待値 = 1", got)
	}
	if value.calls != 0 {
		t.Fatalf("無効なHubでの属性解決回数 = %d、期待値 = 0", value.calls)
	}
}
