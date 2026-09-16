package middleware

import (
	"net/http"
)

// RequestBodyLimitMiddlewareはリクエストボディのサイズを制限するミドルウェア
type RequestBodyLimitMiddleware struct {
	maxBytes int64
}

// NewRequestBodyLimitMiddlewareは新しいRequestBodyLimitMiddlewareを作成
func NewRequestBodyLimitMiddleware(maxBytes int64) *RequestBodyLimitMiddleware {
	return &RequestBodyLimitMiddleware{
		maxBytes: maxBytes,
	}
}

// MiddlewareはHTTPミドルウェアを返す
// リクエストボディのサイズが上限を超えた場合は413 Payload Too Largeを返す
func (m *RequestBodyLimitMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// GETやHEADなどボディのないリクエストはそのまま通す
		if r.Body == nil || r.ContentLength == 0 {
			next.ServeHTTP(w, r)
			return
		}

		// Content-Lengthヘッダーが上限を超えている場合は早期にエラーを返す
		if r.ContentLength > m.maxBytes {
			// 本処理はリバースプロキシより前で走るため、その内側に登録した
			// SecurityHeadersミドルウェアは本レスポンスを見ない。
			setSecurityHeaders(w)
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}

		// ボディを読み込む際にサイズ制限をかける
		r.Body = http.MaxBytesReader(w, r.Body, m.maxBytes)

		next.ServeHTTP(w, r)
	})
}
