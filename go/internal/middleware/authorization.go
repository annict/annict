package middleware

import (
	"net/http"

	"github.com/annict/annict/go/internal/query"
)

// ユーザーの権限を表す定数
// Rails版の User#role enum と対応: user: 0, admin: 1, editor: 2
const (
	RoleUser   int32 = 0
	RoleAdmin  int32 = 1
	RoleEditor int32 = 2
)

// IsAdmin はユーザーが管理者かどうかを判定します
func IsAdmin(user *query.GetUserByIDRow) bool {
	return user != nil && user.Role == RoleAdmin
}

// IsEditor はユーザーが編集者かどうかを判定します
func IsEditor(user *query.GetUserByIDRow) bool {
	return user != nil && user.Role == RoleEditor
}

// IsCommitter はユーザーが管理者または編集者かどうかを判定します
// Rails版の User#committer? に対応
func IsCommitter(user *query.GetUserByIDRow) bool {
	return IsAdmin(user) || IsEditor(user)
}

// RequireCommitter は管理者または編集者のみアクセスを許可するミドルウェアです
// 未認証の場合はログインページにリダイレクトし、権限不足の場合は403を返します
func RequireCommitter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			redirectToSignIn(w, r)
			return
		}

		if !IsCommitter(user) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAdmin は管理者のみアクセスを許可するミドルウェアです
// 未認証の場合はログインページにリダイレクトし、権限不足の場合は403を返します
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			redirectToSignIn(w, r)
			return
		}

		if !IsAdmin(user) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
