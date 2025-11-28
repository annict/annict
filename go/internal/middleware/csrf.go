// Package middleware はHTTPミドルウェアを提供します
package middleware

import (
	"net/http"

	"github.com/annict/annict/internal/session"
)

// CSRFMiddleware はCSRF保護ミドルウェア
type CSRFMiddleware struct {
	sessionManager *session.Manager
}

// NewCSRFMiddleware は新しいCSRFミドルウェアを作成
func NewCSRFMiddleware(sessionManager *session.Manager) *CSRFMiddleware {
	return &CSRFMiddleware{
		sessionManager: sessionManager,
	}
}

// Middleware はCSRFトークン検証ミドルウェアを返す
func (m *CSRFMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// GETリクエストはCSRFチェックをスキップ
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// セッションIDを取得
		sessionID, err := m.sessionManager.GetSessionID(r)
		if err != nil || sessionID == "" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// セッションデータを取得
		sessionData, err := m.sessionManager.GetSession(ctx, sessionID)
		if err != nil || sessionData == nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// フォームからCSRFトークンを取得（フォームパラメータまたはヘッダー）
		formToken := r.FormValue("csrf_token")
		if formToken == "" {
			formToken = r.Header.Get("X-CSRF-Token")
		}

		// トークンが一致しない場合は403エラー
		if formToken != sessionData.CSRFToken {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetCSRFToken はリクエストからCSRFトークンを取得
// テンプレートでトークンを表示する際に使用
func GetCSRFToken(r *http.Request, sessionManager *session.Manager) string {
	ctx := r.Context()

	// セッションIDを取得
	sessionID, err := sessionManager.GetSessionID(r)
	if err != nil || sessionID == "" {
		return ""
	}

	// セッションデータを取得
	sessionData, err := sessionManager.GetSession(ctx, sessionID)
	if err != nil || sessionData == nil {
		return ""
	}

	return sessionData.CSRFToken
}

// GetOrCreateCSRFToken はCSRFトークンを取得し、セッションが存在しない場合は新規作成
// ログインページなど、セッションがまだ存在しない可能性があるページで使用
func GetOrCreateCSRFToken(w http.ResponseWriter, r *http.Request, sessionManager *session.Manager) string {
	ctx := r.Context()

	// session.ManagerのEnsureCSRFToken()を使用してCSRFトークンを取得または生成
	token, err := sessionManager.EnsureCSRFToken(ctx, w, r)
	if err != nil {
		return ""
	}

	return token
}
