package middleware

import (
	"net/http"

	"github.com/annict/annict/go/internal/model"
)

// ロール定数はmodel.Userに定義されているものを再エクスポート
const (
	RoleUser   = model.RoleUser
	RoleAdmin  = model.RoleAdmin
	RoleEditor = model.RoleEditor
)

// IsAdmin はユーザーが管理者かどうかを判定します
func IsAdmin(user *model.User) bool {
	return user != nil && user.IsAdmin()
}

// IsEditor はユーザーが編集者かどうかを判定します
func IsEditor(user *model.User) bool {
	return user != nil && user.IsEditor()
}

// IsCommitter はユーザーが管理者または編集者かどうかを判定します
// Rails版の User#committer? に対応
func IsCommitter(user *model.User) bool {
	return user != nil && user.IsCommitter()
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
