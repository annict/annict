package middleware

import (
	"net/http"
	"strings"
)

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
					switch method {
					case http.MethodPut, http.MethodPatch, http.MethodDelete:
						r.Method = method
					}
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
