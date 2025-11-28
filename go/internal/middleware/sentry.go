// Package middleware はHTTPミドルウェアを提供します
package middleware

import (
	"net/http"
	"strconv"

	"github.com/getsentry/sentry-go"
)

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
					ID:       strconv.FormatInt(user.ID, 10),
					Username: user.Username,
				})
			}
		}

		next.ServeHTTP(w, r)
	})
}
