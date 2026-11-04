package middleware

import "net/http"

// APIResponseHeaderはREST API V2のレスポンスに共通のヘッダーを付けるミドルウェア。
// メンテナンス中の503やパニック時の500にも付けるため、それらのミドルウェアより外側に置く。
// ハンドラーは必要に応じて値を上書きできる (ヘッダーはレスポンスを書き始めるまで変更できる)。
//
//   - Cache-Control: トークンで認証したレスポンスをキャッシュに残さない
//   - X-Content-Type-Options: 宣言したContent-Type以外の形式としてブラウザに解釈させない
//   - X-Robots-Tag: APIのレスポンスを検索エンジンにインデックスさせない
//   - Content-Security-Policy: ブラウザで開いたレスポンスを他のサイトに埋め込ませない。
//     SecurityHeadersが付ける `X-Frame-Options: SAMEORIGIN` より、`frame-ancestors` の指定をブラウザは優先する
func APIResponseHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if IsAPIPath(r.URL.Path) {
			h := w.Header()
			h.Set("Cache-Control", "no-store")
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Robots-Tag", "noindex")
			h.Set("Content-Security-Policy", "frame-ancestors 'none'")
		}
		next.ServeHTTP(w, r)
	})
}
