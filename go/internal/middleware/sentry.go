// Package middleware はHTTPミドルウェアを提供します
package middleware

import (
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
)

// SentryTransaction wires chi's matched route pattern into Sentry as the
// transaction name (e.g. "GET /@{username}/ics"), so the high-cardinality URL
// parameter values do not blow up the transaction list in Sentry. The
// middleware does two things:
//
//  1. On entry it attaches an EventProcessor to the request scope that lazily
//     fills in event.Transaction for error events using chi's RoutePattern at
//     capture time. This covers both panic events emitted by sentryhttp's
//     recoverWithSentry and explicit hub.CaptureException calls inside the
//     handler. chi populates RoutePattern as part of route dispatch (i.e.
//     before the matched handler runs), so reading it at capture time is safe.
//
//  2. On exit (defer) it overwrites the in-flight transaction's Name and
//     Source so the transaction event (performance trace) emitted by
//     sentryhttp's transaction.Finish() carries the route pattern instead of
//     the raw URL.
//
// Register this middleware AFTER sentryhttp so it lives inside the sentryhttp
// wrapper. LIFO defer order then guarantees:
//   - on a normal response: this defer rewrites span Name / Source before
//     sentryhttp's defer calls transaction.Finish().
//   - on a panic: this defer runs first as the stack unwinds, so by the time
//     sentryhttp's recoverWithSentry captures the panic event, the route
//     pattern is already reachable for the EventProcessor that fills in
//     event.Transaction.
//
// [Ja] chi のマッチしたルートパターン (例: "GET /@{username}/ics") を Sentry の
// transaction 名として使うミドルウェア。URL パラメータ値の高カーディナリティで
// Sentry のトランザクション一覧が爆発するのを防ぐ。ミドルウェアは 2 つの仕事
// をする:
//
//  1. 入口でリクエストスコープに EventProcessor を仕込む。これは error
//     イベントの event.Transaction を chi.RouteContext().RoutePattern() から
//     キャプチャ時に遅延埋めする。sentryhttp の recoverWithSentry 経由の
//     panic イベントも、ハンドラー内の明示的な hub.CaptureException も
//     どちらでも動く。chi はルートマッチ時 (= ハンドラー実行前) に
//     RoutePattern を確定させるため、キャプチャ時点での読み出しは安全。
//
//  2. 出口の defer で進行中のトランザクションの Name / Source を上書きする。
//     sentryhttp の transaction.Finish() が送るトランザクションイベント
//     (パフォーマンストレース) に、生 URL ではなくルートパターンが乗る。
//
// このミドルウェアは sentryhttp の **あとに登録** すること (= sentryhttp の
// 内側)。LIFO の defer 順序により以下が保証される:
//   - 正常応答時: 本 defer が span の Name / Source を書き換えてから
//     sentryhttp の defer が transaction.Finish() を呼ぶ。
//   - panic 時: スタック巻き戻し中に本 defer が先に走るため、sentryhttp の
//     recoverWithSentry が panic イベントを捕捉する時点で、EventProcessor が
//     event.Transaction を埋めるために必要なルートパターンに到達できる。
func SentryTransaction(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
			hub.Scope().AddEventProcessor(sentryTransactionEventProcessor(r))
		}
		defer applySentryTransactionName(r)
		next.ServeHTTP(w, r)
	})
}

// sentryTransactionEventProcessor returns an EventProcessor that fills in
// event.Transaction from chi's RoutePattern for error events. Transaction-type
// events get their Transaction field from span.Name (set in
// applySentryTransactionName) and are left untouched here.
//
// [Ja] error イベントの event.Transaction を chi の RoutePattern から埋める
// EventProcessor を返す。transaction 種別のイベントは span.Name
// (applySentryTransactionName で設定) から Transaction を得るため、ここでは
// 触らない。
func sentryTransactionEventProcessor(r *http.Request) sentry.EventProcessor {
	return func(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
		if event == nil || event.Type == "transaction" || event.Transaction != "" {
			return event
		}
		if pattern := matchedRoutePattern(r); pattern != "" {
			event.Transaction = r.Method + " " + pattern
		}
		return event
	}
}

// applySentryTransactionName overwrites the in-flight Sentry transaction's
// Name / Source with chi's matched route pattern. Updating the span is the
// only way to rename the transaction here because *Scope has no
// SetTransaction in sentry-go v0.46.
//
// [Ja] 進行中の Sentry トランザクションの Name / Source を chi のマッチした
// ルートパターンで上書きする。sentry-go v0.46 系では *Scope に SetTransaction
// が無いため、span の更新で transaction 名を差し替える。
func applySentryTransactionName(r *http.Request) {
	pattern := matchedRoutePattern(r)
	if pattern == "" {
		return
	}
	name := r.Method + " " + pattern
	if transaction := sentry.TransactionFromContext(r.Context()); transaction != nil {
		transaction.Name = name
		transaction.Source = sentry.SourceRoute
	}
}

// matchedRoutePattern returns chi's matched route pattern, or "" when the
// request did not go through chi routing.
//
// [Ja] chi のマッチしたルートパターンを返す。chi のルーティングを通っていない
// リクエストでは "" を返す。
func matchedRoutePattern(r *http.Request) string {
	rctx := chi.RouteContext(r.Context())
	if rctx == nil {
		return ""
	}
	return rctx.RoutePattern()
}

// SentryUserContextMiddleware は認証済みユーザーのコンテキストをSentryに設定するミドルウェア
type SentryUserContextMiddleware struct{}

// NewSentryUserContextMiddleware は新しいSentryUserContextMiddlewareを作成
func NewSentryUserContextMiddleware() *SentryUserContextMiddleware {
	return &SentryUserContextMiddleware{}
}

// Middleware はHTTPミドルウェアを返す
// 認証ミドルウェアの後に配置することで、ユーザー情報をSentryに設定できる
func (s *SentryUserContextMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// コンテキストからユーザー情報を取得
		user := GetUserFromContext(ctx)
		if user != nil {
			// SentryのHubをコンテキストから取得
			if hub := sentry.GetHubFromContext(ctx); hub != nil {
				hub.Scope().SetUser(sentry.User{
					ID:       user.ID.String(),
					Username: user.Username,
				})
			}
		}

		next.ServeHTTP(w, r)
	})
}
