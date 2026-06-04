package sentry

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
)

// slogCaptureTransport collects events that the Sentry client would otherwise
// ship over the network so the tests can assert against them in-process.
//
// [Ja] Sentry クライアントが本来ネットワーク送信するイベントをすべて収集する
// テスト用 Transport。
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

// newSlogTestHub builds an isolated Sentry Hub backed by slogCaptureTransport
// so each test sees its own event stream and the global hub stays untouched.
//
// [Ja] テストごとに独立した Hub + slogCaptureTransport を作る。グローバル Hub
// には一切触らない。
func newSlogTestHub(t *testing.T) (*sentry.Hub, *slogCaptureTransport) {
	t.Helper()
	transport := &slogCaptureTransport{}
	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:       "https://public@example.com/1",
		Transport: transport,
	})
	if err != nil {
		t.Fatalf("sentry.NewClient: %v", err)
	}
	return sentry.NewHub(client, sentry.NewScope()), transport
}

// newSlogTestHubWithBeforeSend wires the production beforeSend hook into the
// test Sentry Client. This lets the slog → sentryslog → beforeSend pipeline be
// exercised end-to-end so the suite catches drift in sentryslog's Tag-mapping
// behavior on future upgrades.
//
// [Ja] 本番と同じ beforeSend フックをテスト用 Sentry Client に組み込む。
// slog → sentryslog → beforeSend のパイプラインを End-to-End で検証できるため、
// sentryslog の Tag マッピング挙動が将来のバージョンアップで変わった場合に
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
		t.Fatalf("sentry.NewClient: %v", err)
	}
	return sentry.NewHub(client, sentry.NewScope()), transport
}

// recordedLog is a single slog record captured by recordingHandler with all
// accumulated WithAttrs/WithGroup state flattened into attrs for easy assert.
//
// [Ja] recordingHandler が捕捉した 1 レコード。With チェーンで蓄積された属性も
// attrs に平坦化してテストで検証しやすくする。
type recordedLog struct {
	level slog.Level
	msg   string
	attrs map[string]string
}

// recordingHandler stores every slog.Record it sees. WithAttrs / WithGroup
// derivatives share the same underlying records buffer so the test can
// retrieve everything from the root handler regardless of how the logger was
// derived.
//
// [Ja] 受け取った slog.Record をすべて保持するテスト用ハンドラー。
// WithAttrs / WithGroup で派生したハンドラーも同じバッファを共有するため、
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
		t.Fatalf("expected 1 Sentry event, got %d", len(events))
	}
	got := events[0]
	if got.Message != "operation failed" {
		t.Errorf("event.Message = %q, want %q", got.Message, "operation failed")
	}
	if got.Tags["user_id"] != "u123" {
		t.Errorf("event.Tags[user_id] = %q, want %q", got.Tags["user_id"], "u123")
	}
	if len(got.Exception) != 1 || got.Exception[0].Value != "disk full" {
		t.Errorf("expected exception with value %q, got %+v", "disk full", got.Exception)
	}

	recs := base.Snapshot()
	if len(recs) != 1 {
		t.Fatalf("expected 1 base record, got %d", len(recs))
	}
	if recs[0].level != slog.LevelError || recs[0].msg != "operation failed" {
		t.Errorf("base record = %+v, want Error/operation failed", recs[0])
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
		t.Errorf("expected 0 Sentry events for Info level, got %d", got)
	}

	recs := base.Snapshot()
	if len(recs) != 1 {
		t.Fatalf("expected 1 base record, got %d", len(recs))
	}
	if recs[0].level != slog.LevelInfo || recs[0].msg != "request received" {
		t.Errorf("base record = %+v, want Info/request received", recs[0])
	}
}

func TestSlogHandler_BaseHandlerReceivesAllLevels(t *testing.T) {
	t.Parallel()

	// The base handler is the structured-log sink: every level must reach it
	// regardless of whether Sentry captures the record. This guarantees
	// developers can still tail stderr to see Info/Warn messages while Sentry
	// stays focused on errors only.
	//
	// [Ja] base ハンドラーは構造化ログの出力先なので、Sentry が取り上げるか
	// どうかに関わらず全レベルが届く必要がある。これにより、開発者は標準
	// エラー出力で Info/Warn を引き続き追える一方、Sentry は Error 以上に
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
		t.Errorf("expected exactly 1 Sentry event (Error only), got %d", got)
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
		t.Fatalf("expected %d base records, got %d", len(want), len(recs))
	}
	for i, w := range want {
		if recs[i].level != w.level || recs[i].msg != w.msg {
			t.Errorf("base record[%d] = (%v, %q), want (%v, %q)", i, recs[i].level, recs[i].msg, w.level, w.msg)
		}
	}
}

func TestSlogHandler_DropsReverseProxySourceEventsEndToEnd(t *testing.T) {
	t.Parallel()

	// End-to-end check: a slog.Error tagged with SourceAttrKey=ReverseProxySource
	// must be filtered out by beforeSend before reaching the transport. This
	// catches drift in sentryslog's behavior of stamping slog attributes onto
	// event.Tags (which beforeSend relies on).
	//
	// [Ja] End-to-End の確認: SourceAttrKey=ReverseProxySource を付けた
	// slog.Error は beforeSend で破棄され、transport に到達しないこと。
	// sentryslog が slog 属性を event.Tags に乗せる挙動 (本判定の前提) が
	// 将来変わった場合に本テストで検知できる。
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
		t.Errorf("source=%s のイベントは beforeSend で drop されるべき: got %d events", ReverseProxySource, got)
	}

	// The base text handler still receives the record so developers can tail
	// stderr for debugging.
	//
	// [Ja] base のテキストハンドラーには通常通り届くこと (デバッグ用に標準
	// エラー出力には残す)。
	recs := base.Snapshot()
	if len(recs) != 1 {
		t.Fatalf("expected 1 base record, got %d", len(recs))
	}
	if recs[0].attrs[SourceAttrKey] != ReverseProxySource {
		t.Errorf("base attrs[%s] = %q, want %q", SourceAttrKey, recs[0].attrs[SourceAttrKey], ReverseProxySource)
	}
}

func TestSlogHandler_KeepsOtherSourceEventsEndToEnd(t *testing.T) {
	t.Parallel()

	// Counterpart to TestSlogHandler_DropsReverseProxySourceEventsEndToEnd:
	// only the exact ReverseProxySource value triggers a drop. Other values
	// on SourceAttrKey must still reach Sentry, so the filter does not silently
	// over-broaden.
	//
	// [Ja] TestSlogHandler_DropsReverseProxySourceEventsEndToEnd の対称ケース。
	// SourceAttrKey に乗っていても、値が ReverseProxySource 以外のときは
	// drop せず Sentry に届くこと (フィルタが意図せず広く効いてしまっていない
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
		t.Fatalf("source=other のイベントは drop されず Sentry に届くこと: got %d events", len(events))
	}
	if events[0].Tags[SourceAttrKey] != "other_subsystem" {
		t.Errorf("event.Tags[%s] = %q, want %q", SourceAttrKey, events[0].Tags[SourceAttrKey], "other_subsystem")
	}
}

func TestSlogHandler_MasksSensitiveTagsEndToEnd(t *testing.T) {
	t.Parallel()

	// End-to-end check: an email address logged as a slog attribute must be
	// masked by beforeSend before the event reaches the transport, while the
	// base handler still receives the original value so developers can debug
	// via stderr.
	//
	// [Ja] End-to-End の確認: slog 属性として載せたメールアドレスは beforeSend
	// でマスクされてから transport に届くこと。base ハンドラーには元の値が
	// 届き、標準エラー出力でのデバッグには引き続き使えること。
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
		t.Fatalf("expected 1 Sentry event, got %d", len(events))
	}
	tags := events[0].Tags
	if tags["email"] != "[FILTERED]" {
		t.Errorf("event.Tags[email] = %q, want %q", tags["email"], "[FILTERED]")
	}
	if tags["user_id"] != "u123" {
		t.Errorf("event.Tags[user_id] = %q, want %q", tags["user_id"], "u123")
	}

	recs := base.Snapshot()
	if len(recs) != 1 {
		t.Fatalf("expected 1 base record, got %d", len(recs))
	}
	if recs[0].attrs["email"] != "user@example.com" {
		t.Errorf("base attrs[email] = %q, want %q", recs[0].attrs["email"], "user@example.com")
	}
}

func TestSlogHandler_WithAttrs_PropagatesToBothHandlers(t *testing.T) {
	t.Parallel()

	// Attributes added via Logger.With must reach both the base text handler
	// and the Sentry handler. Otherwise request-scoped fields (e.g. request_id)
	// would be missing from one side, breaking correlation between stderr
	// logs and Sentry events.
	//
	// [Ja] Logger.With で追加した属性は base と Sentry の両方に届く必要がある。
	// さもないとリクエスト単位の情報 (例: request_id) が片方に欠落し、
	// 標準エラー出力のログと Sentry イベントを突き合わせられなくなる。
	hub, transport := newSlogTestHub(t)
	base := newRecordingHandler()
	logger := slog.New(NewSlogHandler(base)).With("request_id", "r123")

	ctx := sentry.SetHubOnContext(context.Background(), hub)
	logger.ErrorContext(ctx, "boom", "user_id", "u456")

	hub.Flush(2 * time.Second)
	events := transport.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 Sentry event, got %d", len(events))
	}
	tags := events[0].Tags
	if tags["request_id"] != "r123" {
		t.Errorf("event.Tags[request_id] = %q, want %q", tags["request_id"], "r123")
	}
	if tags["user_id"] != "u456" {
		t.Errorf("event.Tags[user_id] = %q, want %q", tags["user_id"], "u456")
	}

	recs := base.Snapshot()
	if len(recs) != 1 {
		t.Fatalf("expected 1 base record, got %d", len(recs))
	}
	if recs[0].attrs["request_id"] != "r123" {
		t.Errorf("base attrs[request_id] = %q, want %q", recs[0].attrs["request_id"], "r123")
	}
	if recs[0].attrs["user_id"] != "u456" {
		t.Errorf("base attrs[user_id] = %q, want %q", recs[0].attrs["user_id"], "u456")
	}
}
