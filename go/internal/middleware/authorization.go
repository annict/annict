package middleware

import (
	"net/http"

	"github.com/annict/annict/go/internal/model"
)

// Re-export role constants from model.User so that callers can reference roles without importing the model package.
//
// [Ja] ロール定数を model.User から再エクスポートする。呼び出し側がロール値のためだけに model パッケージを import せずに済むようにする目的。
const (
	RoleUser   = model.RoleUser
	RoleAdmin  = model.RoleAdmin
	RoleEditor = model.RoleEditor
)

func IsAdmin(user *model.User) bool {
	return user != nil && user.IsAdmin()
}

func IsEditor(user *model.User) bool {
	return user != nil && user.IsEditor()
}

// IsCommitter reports whether the user is either an admin or an editor. Corresponds to the Rails-side User#committer?.
//
// [Ja] ユーザーが管理者または編集者かどうかを返す。Rails 版の User#committer? に対応する。
func IsCommitter(user *model.User) bool {
	return user != nil && user.IsCommitter()
}

// RequireCommitter only allows requests from admins or editors. Unauthenticated users are redirected to the sign-in page,
// and users without sufficient permission receive a 403 response.
//
// [Ja] 管理者または編集者のみアクセスを許可するミドルウェア。未認証の場合はログインページにリダイレクトし、
// 権限不足の場合は 403 を返す。
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

// RequireAdmin only allows requests from admins. Unauthenticated users are redirected to the sign-in page,
// and users without admin permission receive a 403 response.
//
// [Ja] 管理者のみアクセスを許可するミドルウェア。未認証の場合はログインページにリダイレクトし、
// 権限不足の場合は 403 を返す。
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
