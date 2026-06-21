package sentry

import (
	"context"
	"log/slog"
	"os"

	sentryslog "github.com/getsentry/sentry-go/slog"
)

// SourceAttrKey is the slog attribute key used to tag the origin of a log
// event. Code that wants beforeSend to drop the resulting Sentry event
// (typically because the failure belongs to another Sentry project) sets this
// attribute with a recognized source value such as ReverseProxySource.
//
// The key is namespaced with the "annict_" prefix so it cannot collide with a
// generic "source" attribute that another developer might add for unrelated
// logging purposes. Always reference this constant instead of writing the
// string literal at call sites.
//
// [Ja] ログイベントの発生源をタグ付けする slog 属性のキー名。本キーに既知の
// 値 (例: ReverseProxySource) を載せたエラーログは、beforeSend 側でその値を
// 検出して Sentry 送信を抑止する。
//
// 別の開発者が無関係な目的で汎用的な "source" 属性を追加した場合に衝突しない
// よう、"annict_" プレフィックスで名前空間を切っている。呼び出し側では文字列
// リテラルを直書きせず、必ず本定数を参照すること。
const SourceAttrKey = "annict_source"

// ReverseProxySource is the SourceAttrKey value set by the reverse-proxy
// middleware when Rails returns an error response. beforeSend drops events
// carrying this tag because Rails-side failures (HTTP 502 etc.) belong to the
// Rails Sentry project, not the Go one.
//
// [Ja] Rails 版がエラーを返した際にリバースプロキシミドルウェアが
// SourceAttrKey に設定する値。Rails 側の障害 (HTTP 502 など) は Rails の
// Sentry プロジェクトで扱うべきなので、beforeSend で本タグの付いた
// イベントを破棄する。
const ReverseProxySource = "reverse_proxy"

// NewBaseHandler returns the application's base slog handler: a text handler
// that writes to stderr at LevelInfo. It is the single source of truth for the
// log output format and level.
//
// The default logger wraps this base with NewSlogHandler so that Error and
// Fatal records also fan out to Sentry. Background loggers that must NOT reach
// Sentry use NewBaseHandler directly (e.g. River's internal logger, which logs
// self-healing connection blips at Error level). Sharing one constructor keeps
// their format and verbosity identical and prevents the two from drifting apart
// if the level is ever changed.
//
// [Ja] アプリケーションの基底 slog ハンドラーを返す。標準エラー出力へ
// LevelInfo で書き出すテキストハンドラーで、ログの出力形式とレベルの唯一の
// 情報源となる。
//
// デフォルトロガーはこの基底を NewSlogHandler でラップし、Error と Fatal の
// レコードを Sentry にもファンアウトさせる。Sentry に流してはならない
// バックグラウンドロガーは NewBaseHandler を直接使う (例: 自己回復する接続の
// 瞬断を Error レベルで出力する River の内部ロガー)。コンストラクタを共有する
// ことで両者の形式と詳細度が一致し、将来レベルを変更してもドリフトしない。
func NewBaseHandler() slog.Handler {
	return slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
}

// NewSlogHandler wraps base in a fan-out handler that also forwards
// slog.LevelError (and LevelFatal) records to Sentry as events. Other levels
// reach only the base handler. The Sentry Logs API is intentionally disabled
// so this integration captures errors only.
//
// [Ja] base ハンドラーを sentryslog ハンドラーと合成して返す。slog.LevelError
// と LevelFatal のレコードを Sentry にイベントとして送信し、それ以外のレベルは
// base にだけ流す。Sentry Logs API への送信は明示的に無効化している (本連携は
// エラー検出のみを担うため)。
func NewSlogHandler(base slog.Handler) slog.Handler {
	sentryHandler := sentryslog.Option{
		// Capture exactly Error and Fatal as Sentry events.
		//
		// [Ja] Error と Fatal だけを Sentry イベント化する。
		EventLevel: []slog.Level{slog.LevelError, sentryslog.LevelFatal},
		// Disable the Sentry Logs API by passing an empty (non-nil) slice.
		// A nil slice falls back to the package default of all levels.
		//
		// [Ja] 空 (非 nil) のスライスを渡すことで Sentry Logs API への送信を
		// 無効化する。nil にするとパッケージ既定値 (全レベル送信) に
		// フォールバックしてしまう。
		LogLevel: []slog.Level{},
	}.NewSentryHandler(context.Background())
	return newMultiHandler(base, sentryHandler)
}

// multiHandler fans out one slog.Record to multiple slog.Handlers. It keeps
// the base text handler in place while adding Sentry capture on top. The
// implementation is intentionally minimal so we do not pull in
// samber/slog-multi just for fan-out.
//
// [Ja] 1 レコードを複数の slog.Handler にファンアウトする。base のテキスト
// ハンドラーをそのままに、Sentry 用ハンドラーを並列で動かすために使う。
// samber/slog-multi のような外部依存を避けるため、ファンアウトに必要最小限の
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
		// Clone the record per handler so a handler that mutates the record's
		// attribute back-array (e.g. via Record.AddAttrs) cannot affect later
		// handlers in the fan-out.
		//
		// [Ja] ハンドラーごとに Clone することで、Record.AddAttrs のように属性
		// 配列を直接書き換える実装が後続ハンドラーに影響しないようにする。
		if err := h.Handle(ctx, record.Clone()); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	// Return only the first non-nil error to mirror slog-multi semantics; later
	// handlers' errors are intentionally discarded because slog itself only
	// surfaces a single error per Handle call.
	//
	// [Ja] slog-multi の挙動に合わせて最初のエラーだけ返す。Handle が返せる
	// エラーは 1 件のため、2 件目以降は意図的に捨てている。
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
