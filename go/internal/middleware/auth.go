// Package middleware はHTTPミドルウェアを提供します
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
)

type contextKey string

const (
	// UserContextKey はコンテキストからユーザーを取得するためのキー
	UserContextKey contextKey = "user"
)

// AuthMiddleware は認証を行うミドルウェア
type AuthMiddleware struct {
	sessionManager *session.Manager
	sessionRepo    *repository.SessionRepository
}

// NewAuthMiddleware は新しいAuthMiddlewareを作成
func NewAuthMiddleware(sessionManager *session.Manager, sessionRepo *repository.SessionRepository) *AuthMiddleware {
	return &AuthMiddleware{
		sessionManager: sessionManager,
		sessionRepo:    sessionRepo,
	}
}

// Middleware はHTTPミドルウェアを返す
func (a *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// セッションからユーザー情報を取得
		user, err := a.sessionManager.GetCurrentUser(r.Context(), r)
		if err != nil {
			// エラーが発生してもリクエスト処理は続行
			// ログ出力などが必要な場合はここに追加
			next.ServeHTTP(w, r)
			return
		}

		// ユーザー情報がある場合はコンテキストに設定
		if user != nil {
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAuth は認証が必要なエンドポイント用のミドルウェア
func (a *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := GetUserFromContext(ctx)
		if user == nil {
			// 未認証の場合はログインページにリダイレクト（元のURLを back パラメータに付与）
			redirectToSignIn(w, r)
			return
		}

		// セッションIDを取得
		sessionID, err := a.sessionManager.GetSessionID(r)
		if err != nil || sessionID == "" {
			// セッションIDが取得できない場合はログインページにリダイレクト
			redirectToSignIn(w, r)
			return
		}

		// セッションのupdated_atを更新
		// エラーが発生してもログに記録するだけで処理を継続
		if err := a.sessionRepo.TouchSession(ctx, sessionID); err != nil {
			slog.WarnContext(ctx, "セッション更新エラー", "error", err)
		}

		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext はコンテキストからユーザー情報を取得
func GetUserFromContext(ctx context.Context) *query.GetUserByIDRow {
	if user, ok := ctx.Value(UserContextKey).(*query.GetUserByIDRow); ok {
		return user
	}
	return nil
}

// redirectToSignIn は未認証ユーザーをログインページにリダイレクトする
// 元のURLを back パラメータとして付与することで、ログイン後に元のページに戻れるようにする
func redirectToSignIn(w http.ResponseWriter, r *http.Request) {
	backURL := r.URL.RequestURI()
	redirectURL := "/sign_in?back=" + url.QueryEscape(backURL)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}
