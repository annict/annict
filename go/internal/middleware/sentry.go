// Package middlewareはHTTPミドルウェアを提供します
package middleware

import (
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
)

// SentryTransactionはchiのマッチしたルートパターン (例: "GET /@{username}/ics") をSentryの
// transaction名として使うミドルウェア。URLパラメータ値の高カーディナリティで
// Sentryのトランザクション一覧が爆発するのを防ぐ。ミドルウェアは2つの仕事
// をする:
//
//  1. 入口でリクエストスコープにEventProcessorを仕込む。これはerror
//     イベントのevent.Transactionをchi.RouteContext().RoutePattern() から
//     キャプチャ時に遅延埋めする。sentryhttpのrecoverWithSentry経由の
//     panicイベントも、ハンドラー内の明示的なhub.CaptureExceptionも
//     どちらでも動く。chiはルートマッチ時 (= ハンドラー実行前) に
//     RoutePatternを確定させるため、キャプチャ時点での読み出しは安全。
//
//  2. 出口のdeferで進行中のトランザクションのName / Sourceを上書きする。
//     sentryhttpのtransaction.Finish() が送るトランザクションイベント
//     (パフォーマンストレース) に、生URLではなくルートパターンが乗る。
//
// このミドルウェアはsentryhttpの **あとに登録** すること (= sentryhttpの
// 内側)。LIFOのdefer順序により以下が保証される:
//   - 正常応答時: 本deferがspanのName / Sourceを書き換えてから
//     sentryhttpのdeferがtransaction.Finish() を呼ぶ。
//   - panic時: スタック巻き戻し中に本deferが先に走るため、sentryhttpの
//     recoverWithSentryがpanicイベントを捕捉する時点で、EventProcessorが
//     event.Transactionを埋めるために必要なルートパターンに到達できる。
func SentryTransaction(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
			hub.Scope().AddEventProcessor(sentryTransactionEventProcessor(r))
		}
		defer applySentryTransactionName(r)
		next.ServeHTTP(w, r)
	})
}

// errorイベントのevent.TransactionをchiのRoutePatternから埋める
// EventProcessorを返す。transaction種別のイベントはspan.Name
// (applySentryTransactionNameで設定) からTransactionを得るため、ここでは
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

// 進行中のSentryトランザクションのName / Sourceをchiのマッチした
// ルートパターンで上書きする。sentry-go v0.46系では *ScopeにSetTransaction
// が無いため、spanの更新でtransaction名を差し替える。
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

// chiのマッチしたルートパターンを返す。chiのルーティングを通っていない
// リクエストでは "" を返す。
func matchedRoutePattern(r *http.Request) string {
	rctx := chi.RouteContext(r.Context())
	if rctx == nil {
		return ""
	}
	return rctx.RoutePattern()
}

// SentryUserContextMiddlewareは認証済みユーザーのコンテキストをSentryに設定するミドルウェア
type SentryUserContextMiddleware struct{}

// NewSentryUserContextMiddlewareは新しいSentryUserContextMiddlewareを作成
func NewSentryUserContextMiddleware() *SentryUserContextMiddleware {
	return &SentryUserContextMiddleware{}
}

// MiddlewareはHTTPミドルウェアを返す
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
