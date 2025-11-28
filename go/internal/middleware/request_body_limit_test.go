package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestBodyLimitMiddleware_NoBody(t *testing.T) {
	t.Parallel()

	mw := NewRequestBodyLimitMiddleware(1024) // 1KB
	handler := mw.Middleware(testHandler())

	// GETリクエスト（ボディなし）
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ボディなしリクエスト: ステータスコード = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestRequestBodyLimitMiddleware_SmallBody(t *testing.T) {
	t.Parallel()

	mw := NewRequestBodyLimitMiddleware(1024) // 1KB

	// 小さいボディを読み込んで成功を確認するハンドラー
	handler := mw.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))

	// 制限内のサイズのボディ（100バイト）
	body := strings.Repeat("a", 100)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("小さいボディ: ステータスコード = %d, want %d", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != body {
		t.Errorf("小さいボディ: レスポンスボディが一致しません")
	}
}

func TestRequestBodyLimitMiddleware_ExactLimit(t *testing.T) {
	t.Parallel()

	limit := int64(1024)
	mw := NewRequestBodyLimitMiddleware(limit) // 1KB

	// ボディを読み込んで成功を確認するハンドラー
	handler := mw.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	// ちょうど制限サイズのボディ
	body := strings.Repeat("a", int(limit))
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ちょうど制限サイズ: ステータスコード = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestRequestBodyLimitMiddleware_ExceedsLimit_ContentLength(t *testing.T) {
	t.Parallel()

	mw := NewRequestBodyLimitMiddleware(1024) // 1KB
	handler := mw.Middleware(testHandler())

	// Content-Lengthが制限を超えている場合
	body := strings.Repeat("a", 2048) // 2KB
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Content-Length超過: ステータスコード = %d, want %d", rr.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestRequestBodyLimitMiddleware_ExceedsLimit_ReadBody(t *testing.T) {
	t.Parallel()

	limit := int64(1024)
	mw := NewRequestBodyLimitMiddleware(limit) // 1KB

	// ボディを読み込もうとしてエラーになることを確認するハンドラー
	var readErr error
	handler := mw.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, readErr = io.ReadAll(r.Body)
		if readErr != nil {
			// http.MaxBytesReaderはエラーを返すが、
			// レスポンスは自動的に413が設定される
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	// Content-Lengthが設定されていない場合（chunked encoding などの想定）
	// 実際に読み込んでエラーになるケースをテスト
	body := bytes.NewReader(make([]byte, int(limit)+100))
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.ContentLength = -1 // Content-Lengthを未設定に
	req.Header.Set("Content-Type", "application/octet-stream")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// http.MaxBytesReaderが上限を超えた場合はエラーを返す
	if readErr == nil {
		t.Error("ボディ読み込み時にエラーが発生するべきでした")
	}
}

func TestRequestBodyLimitMiddleware_ZeroContentLength(t *testing.T) {
	t.Parallel()

	mw := NewRequestBodyLimitMiddleware(1024) // 1KB
	handler := mw.Middleware(testHandler())

	// Content-Length = 0 のリクエスト
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("空ボディ: ステータスコード = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestRequestBodyLimitMiddleware_10MB(t *testing.T) {
	t.Parallel()

	// 本番環境で使用する10MBの制限をテスト
	limit := int64(10 * 1024 * 1024) // 10MB
	mw := NewRequestBodyLimitMiddleware(limit)
	handler := mw.Middleware(testHandler())

	testCases := []struct {
		name     string
		size     int64
		wantCode int
	}{
		{"1MB（制限内）", 1 * 1024 * 1024, http.StatusOK},
		{"5MB（制限内）", 5 * 1024 * 1024, http.StatusOK},
		{"10MB（ちょうど制限）", 10 * 1024 * 1024, http.StatusOK},
		{"11MB（制限超過）", 11 * 1024 * 1024, http.StatusRequestEntityTooLarge},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body := bytes.NewReader(make([]byte, tc.size))
			req := httptest.NewRequest(http.MethodPost, "/upload", body)
			req.Header.Set("Content-Type", "application/octet-stream")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantCode {
				t.Errorf("%s: ステータスコード = %d, want %d", tc.name, rr.Code, tc.wantCode)
			}
		})
	}
}

func TestRequestBodyLimitMiddleware_DifferentMethods(t *testing.T) {
	t.Parallel()

	mw := NewRequestBodyLimitMiddleware(1024) // 1KB
	handler := mw.Middleware(testHandler())

	methods := []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodDelete,
		http.MethodOptions,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("%s: ステータスコード = %d, want %d", method, rr.Code, http.StatusOK)
			}
		})
	}
}
