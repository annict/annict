package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/go-chi/chi/v5"
)

// overridableMethodsは _methodパラメータによってPOSTから書き換えられるメソッドの
// 一覧。ReverseProxyMiddlewareも、オーバーライド経由でしか到達できないルートを認識するために
// これを読む。両ミドルウェアが「オーバーライドが生みうるメソッド」の認識を共有するため。
var overridableMethods = []string{
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
}

// MethodOverrideはHTMLフォームから送信された_methodパラメータを読み取り、
// HTTPメソッドを上書きします (Rails方式)
//
// 使用例:
//
//	<form method="POST" action="/password">
//	  <input type="hidden" name="_method" value="PUT">
//	</form>
//
// これにより、HTMLフォーム (GETとPOSTのみサポート) とREST API (PUT/PATCH/DELETE) で
// 同じルーティングを使用できます。
//
// GETを上書きできるとリンク1つで状態を変える要求を作れてしまうため、上書きするのはPOSTの要求だけにする。
func MethodOverride(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			if err := r.ParseForm(); err == nil {
				method := strings.ToUpper(r.PostFormValue("_method"))
				if slices.Contains(overridableMethods, method) {
					r.Method = method
					// 外側のルーターからMountで載せられていると、chiはルーティングに使うメソッドを
					// 外側のルーティングの時点で決めて内側でも使い回すため、r.Methodだけを変えても
					// 上書きがルーティングに効かない。
					if rctx := chi.RouteContext(r.Context()); rctx != nil {
						rctx.RouteMethod = method
					}
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
