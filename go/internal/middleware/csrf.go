// Package middlewareはHTTPミドルウェアを提供します
package middleware

import (
	"net/http"
	"strings"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/session"
)

// CSRFMiddlewareはCSRF保護ミドルウェア
type CSRFMiddleware struct {
	sessionManager *session.Manager
	skipPaths      []string
}

// NewCSRFMiddlewareは新しいCSRFミドルウェアを作成
func NewCSRFMiddleware(sessionManager *session.Manager) *CSRFMiddleware {
	return &CSRFMiddleware{
		sessionManager: sessionManager,
		skipPaths: []string{
			"/webhooks/stripe", // Stripeは独自の署名検証を使用
			"/sign_out",        // Rails版からのリクエスト対応のためCSRFを適用しない
		},
	}
}

// MiddlewareはCSRFトークン検証ミドルウェアを返す
func (m *CSRFMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// GETリクエストはCSRFチェックをスキップ
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// スキップパスに一致する場合はCSRFチェックをスキップ
		for _, path := range m.skipPaths {
			if strings.HasPrefix(r.URL.Path, path) {
				next.ServeHTTP(w, r)
				return
			}
		}

		// 以下の3つの判定 (セッションID無し / セッションデータ無し / トークン不一致) は
		// いずれも同じ応答を返す。区別すると送信側にセッションの有無を伝えることになり、
		// 読み手に求める次の行動もどの場合も同じ (ページを開き直して再送信する) であるため。

		// セッションIDを取得
		sessionID, err := m.sessionManager.GetSessionID(r)
		if err != nil || sessionID == "" {
			httperror.InvalidCSRFToken(w, r)
			return
		}

		// セッションデータを取得
		sessionData, err := m.sessionManager.GetSession(ctx, sessionID)
		if err != nil || sessionData == nil {
			httperror.InvalidCSRFToken(w, r)
			return
		}

		// フォームからCSRFトークンを取得 (フォームパラメータまたはヘッダー)
		formToken := r.FormValue("csrf_token")
		if formToken == "" {
			formToken = r.Header.Get("X-CSRF-Token")
		}

		if formToken != sessionData.CSRFToken {
			httperror.InvalidCSRFToken(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetCSRFTokenはリクエストからCSRFトークンを取得
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

// GetOrCreateCSRFTokenはCSRFトークンを取得し、セッションが存在しない場合は新規作成
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
