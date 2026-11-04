package middleware

import (
	"net/http"
	"strconv"
)

// apiCORSMaxAgeはプリフライトの結果をブラウザがキャッシュしてよい秒数
const apiCORSMaxAge = 7200

const (
	// apiCORSAllowMethodsはプリフライトで許可するメソッド。V2は読み取りのoperationだけを持つ
	apiCORSAllowMethods = "GET"
	// apiCORSAllowHeadersはプリフライトで許可するリクエストヘッダー。
	// トークンを送る `Authorization` はCORSで常に許可が要るヘッダーのため明示する
	apiCORSAllowHeaders = "Authorization"
	// apiCORSExposeHeadersはブラウザのスクリプトに読ませるレスポンスヘッダー。
	// ページ送りとレート制限の情報は、既定では読めないヘッダーで返すため並べる
	apiCORSExposeHeaders = "Link, Retry-After, RateLimit-Limit, RateLimit-Remaining, RateLimit-Reset"
)

// APICORSはREST API V2にCORSのヘッダーを付け、プリフライトに応えるミドルウェア。
// V2はCookieで認証せず `Authorization` ヘッダーのトークンだけで認証するため、
// ほかのサイトのブラウザアプリも利用者自身のトークンで呼べるよう、全オリジンに許可する。
// 429・413・503・500もブラウザのスクリプトが読めるよう、レート制限と早期応答より外側に置く。
// プリフライトはこの位置で応えるため、レート制限で数えず、メンテナンス中や上限超過でも本要求を止めない
func APICORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsAPIPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Expose-Headers", apiCORSExposeHeaders)

		if isCORSPreflight(r) {
			h.Set("Access-Control-Allow-Methods", apiCORSAllowMethods)
			h.Set("Access-Control-Allow-Headers", apiCORSAllowHeaders)
			h.Set("Access-Control-Max-Age", strconv.Itoa(apiCORSMaxAge))
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isCORSPreflightはrがCORSのプリフライトかを返す。
// `Origin` と `Access-Control-Request-Method` の無いOPTIONSはプリフライトではないため、ルーターに任せる
func isCORSPreflight(r *http.Request) bool {
	return r.Method == http.MethodOptions &&
		r.Header.Get("Origin") != "" &&
		r.Header.Get("Access-Control-Request-Method") != ""
}
