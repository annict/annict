package middleware

import (
	"net/http"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/model"
)

// ロール定数をmodel.Userから再エクスポートする。呼び出し側がロール値のためだけにmodelパッケージをimportせずに済むようにする目的。
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

// IsCommitterはユーザーが管理者または編集者かどうかを返す。Rails版のUser#committer? に対応する。
func IsCommitter(user *model.User) bool {
	return user != nil && user.IsCommitter()
}

// RequireCommitterは管理者または編集者のみアクセスを許可するミドルウェア。未認証の場合はログインページにリダイレクトし、
// 権限不足の場合は403を返す。
func RequireCommitter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			redirectToSignIn(w, r)
			return
		}

		if !IsCommitter(user) {
			httperror.Forbidden(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAdminは管理者のみアクセスを許可するミドルウェア。未認証の場合はログインページにリダイレクトし、
// 権限不足の場合は403を返す。
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			redirectToSignIn(w, r)
			return
		}

		if !IsAdmin(user) {
			httperror.Forbidden(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
