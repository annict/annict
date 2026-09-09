package middleware

import (
	"net/http"
	"slices"
	"strings"
)

// overridableMethods lists the methods a POST can be rewritten to through the _method
// parameter. ReverseProxyMiddleware reads it too, to recognise a route that is only
// reachable through the override, so both middlewares agree on what an override can produce.
//
// [Ja] overridableMethods は _method パラメータによって POST から書き換えられるメソッドの
// 一覧。ReverseProxyMiddleware も、オーバーライド経由でしか到達できないルートを認識するために
// これを読む。両ミドルウェアが「オーバーライドが生みうるメソッド」の認識を共有するため。
var overridableMethods = []string{
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
}

// MethodOverride はHTMLフォームから送信された_methodパラメータを読み取り、
// HTTPメソッドを上書きします（Rails方式）
//
// 使用例:
//
//	<form method="POST" action="/password">
//	  <input type="hidden" name="_method" value="PUT">
//	</form>
//
// これにより、HTMLフォーム（GETとPOSTのみサポート）とREST API（PUT/PATCH/DELETE）で
// 同じルーティングを使用できます。
func MethodOverride(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// POSTリクエストのみ処理
		if r.Method == http.MethodPost {
			// フォームデータから_methodパラメータを取得
			if err := r.ParseForm(); err == nil {
				method := r.PostFormValue("_method")
				if method != "" {
					// サポートされているメソッドのみ許可
					method = strings.ToUpper(method)
					if slices.Contains(overridableMethods, method) {
						r.Method = method
					}
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
