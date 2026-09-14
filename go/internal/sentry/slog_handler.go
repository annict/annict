package sentry

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/getsentry/sentry-go"
	sentryslog "github.com/getsentry/sentry-go/slog"
)

// SourceAttrKeyはログイベントの発生源をタグ付けするslog属性のキー名。本キーに既知の
// 値 (例: ReverseProxySource) を載せたエラーログは、beforeSend側でその値を
// 検出してSentry送信を抑止する。
//
// 別の開発者が無関係な目的で汎用的な "source" 属性を追加した場合に衝突しない
// よう、"annict_" プレフィックスで名前空間を切っている。呼び出し側では文字列
// リテラルを直書きせず、必ず本定数を参照すること。
const SourceAttrKey = "annict_source"

// ReverseProxySourceはRails版がエラーを返した際にリバースプロキシミドルウェアが
// SourceAttrKeyに設定する値。Rails側の障害 (HTTP 502など) はRailsの
// Sentryプロジェクトで扱うべきなので、beforeSendで本タグの付いた
// イベントを破棄する。
const ReverseProxySource = "reverse_proxy"

// NewBaseHandlerはアプリケーションの基底slogハンドラーを返す。標準エラー出力へ
// LevelInfoで書き出すテキストハンドラーで、ログの出力形式とレベルの唯一の
// 情報源となる。
//
// デフォルトロガーはこの基底をNewSlogHandlerでラップし、ErrorとFatalの
// レコードをSentryにもファンアウトさせる。Sentryに流してはならない
// バックグラウンドロガーはNewBaseHandlerを直接使う (例: 自己回復する接続の
// 瞬断をErrorレベルで出力するRiverの内部ロガー)。コンストラクタを共有する
// ことで両者の形式と詳細度が一致し、将来レベルを変更してもドリフトしない。
func NewBaseHandler() slog.Handler {
	return slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
}

// NewSlogHandlerはbaseハンドラーをアプリケーション固有のSentryイベントハンドラーと
// 合成して返す。slog.LevelErrorとLevelFatalのレコードをSentryにイベント
// として送信し、それ以外のレベルはbaseにだけ流す。Sentry Go 0.48でslog
// 連携からイベント作成機能が削除されたため、イベント専用の小さなハンドラーを
// アプリケーション側で持ち、Sentry Logs APIにはopt inしない。
func NewSlogHandler(base slog.Handler) slog.Handler {
	return newMultiHandler(base, &sentryEventHandler{})
}

const defaultMaxErrorDepth = 100

type sentryEventAttr struct {
	groups []string
	attr   slog.Attr
}

// ErrorとFatalのslogレコードだけをSentryイベントへ変換する。
// sentry-go/slog 0.48でイベント作成機能が削除される前にAnnictが依存していた
// 契約だけを意図的に実装する。
type sentryEventHandler struct {
	attrs  []sentryEventAttr
	groups []string
}

func (h *sentryEventHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level == slog.LevelError || level == sentryslog.LevelFatal
}

func (h *sentryEventHandler) Handle(ctx context.Context, record slog.Record) error {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	event := sentry.NewEvent()
	event.Timestamp = record.Time.UTC()
	event.Level = sentry.LevelError
	if record.Level == sentryslog.LevelFatal {
		event.Level = sentry.LevelFatal
	}
	event.Message = record.Message
	event.Logger = "slog"

	var originalErr error
	for _, a := range h.attrs {
		if err := addSlogAttrToEvent(event, strings.Join(a.groups, "."), a.attr); err != nil && originalErr == nil {
			originalErr = err
		}
	}
	record.Attrs(func(a slog.Attr) bool {
		if err := addSlogAttrToEvent(event, strings.Join(h.groups, "."), a); err != nil && originalErr == nil {
			originalErr = err
		}
		return true
	})

	errorDepth := defaultMaxErrorDepth
	if client := hub.Client(); client != nil {
		errorDepth = client.Options().MaxErrorDepth
	}
	event.SetException(originalErr, errorDepth)
	hub.CaptureEventWithHint(event, &sentry.EventHint{Context: ctx})
	return nil
}

func (h *sentryEventHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	out := &sentryEventHandler{
		attrs:  make([]sentryEventAttr, 0, len(h.attrs)+len(attrs)),
		groups: append([]string(nil), h.groups...),
	}
	out.attrs = append(out.attrs, h.attrs...)
	for _, a := range attrs {
		out.attrs = append(out.attrs, sentryEventAttr{
			groups: append([]string(nil), h.groups...),
			attr:   a,
		})
	}
	return out
}

func (h *sentryEventHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	return &sentryEventHandler{
		attrs:  append([]sentryEventAttr(nil), h.attrs...),
		groups: append(append([]string(nil), h.groups...), name),
	}
}

// 1つのslog属性をイベントへ追加する。error属性はevent.Exceptionに
// 設定できるよう、タグには追加せず呼び出し元へ返す。
func addSlogAttrToEvent(event *sentry.Event, group string, attr slog.Attr) error {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return nil
	}

	key := attr.Key
	if group != "" {
		if key != "" {
			key = group + "." + key
		} else {
			key = group
		}
	}

	if attr.Value.Kind() == slog.KindGroup {
		var firstErr error
		for _, child := range attr.Value.Group() {
			if err := addSlogAttrToEvent(event, key, child); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		return firstErr
	}

	if (attr.Key == "error" || attr.Key == "err") && attr.Value.Kind() == slog.KindAny {
		if err, ok := attr.Value.Any().(error); ok {
			return err
		}
	}
	if key != "" {
		event.Tags[key] = fmt.Sprint(attr.Value.Any())
	}
	return nil
}

// 1レコードを複数のslog.Handlerにファンアウトする。baseのテキストハンドラーを維持し、
// 同じレコードをSentry用ハンドラーにも渡すために使う。
// samber/slog-multiのような外部依存を避けるため、ファンアウトに必要最小限の
// 実装を内製している。
type multiHandler struct {
	handlers []slog.Handler
}

func newMultiHandler(handlers ...slog.Handler) *multiHandler {
	return &multiHandler{handlers: handlers}
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, record slog.Record) error {
	var firstErr error
	for _, h := range m.handlers {
		if !h.Enabled(ctx, record.Level) {
			continue
		}
		// ハンドラーごとにCloneすることで、Record.AddAttrsのように属性
		// 配列を直接書き換える実装が後続ハンドラーに影響しないようにする。
		if err := h.Handle(ctx, record.Clone()); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	// slog-multiの挙動に合わせて最初のエラーだけ返す。Handleが返せる
	// エラーは1件のため、2件目以降は意図的に捨てている。
	return firstErr
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	out := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		out[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: out}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	out := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		out[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: out}
}
