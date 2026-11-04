package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/apierror"
)

func TestIsAPIPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path string
		want bool
	}{
		{path: "/api/v2", want: true},
		{path: "/api/v2/", want: true},
		{path: "/api/v2/anime/1", want: true},
		{path: "/api/v20", want: false},
		{path: "/api/internal/works", want: false},
		{path: "/works", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()

			if got := IsAPIPath(tt.path); got != tt.want {
				t.Errorf("IsAPIPath(%q) = %v、期待値 = %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestAPIResponseHeader(t *testing.T) {
	t.Parallel()

	apiHeaders := map[string]string{
		"Cache-Control":           "no-store",
		"X-Content-Type-Options":  "nosniff",
		"X-Robots-Tag":            "noindex",
		"Content-Security-Policy": "frame-ancestors 'none'",
	}

	t.Run("APIのパスには共通のヘッダーを付ける", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		APIResponseHeader(testHandler()).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v2/openapi.yaml", nil))

		for name, want := range apiHeaders {
			if got := rr.Header().Get(name); got != want {
				t.Errorf("%s = %q、期待値 = %q", name, got, want)
			}
		}
	})

	t.Run("APIのパスでなければ付けない", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		APIResponseHeader(testHandler()).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/works", nil))

		for name := range apiHeaders {
			if got := rr.Header().Get(name); got != "" {
				t.Errorf("%s = %q、期待値 = 付かない", name, got)
			}
		}
	})
}

func TestAPICORS(t *testing.T) {
	t.Parallel()

	t.Run("APIのリクエストに全オリジンの許可と公開するヘッダーを付ける", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/api/v2/openapi.yaml", nil)
		req.Header.Set("Origin", "https://client.example.com")
		rr := httptest.NewRecorder()
		APICORS(testHandler()).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK || rr.Body.String() != "OK" {
			t.Errorf("ステータスコード = %d、本文 = %q、期待値 = 次のハンドラーの応答", rr.Code, rr.Body.String())
		}
		if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
			t.Errorf("Access-Control-Allow-Origin = %q、期待値 = *", got)
		}
		want := "Link, Retry-After, RateLimit-Limit, RateLimit-Remaining, RateLimit-Reset"
		if got := rr.Header().Get("Access-Control-Expose-Headers"); got != want {
			t.Errorf("Access-Control-Expose-Headers = %q、期待値 = %q", got, want)
		}
	})

	t.Run("プリフライトに204で応え、次のハンドラーを呼ばない", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodOptions, "/api/v2/anime/1", nil)
		req.Header.Set("Origin", "https://client.example.com")
		req.Header.Set("Access-Control-Request-Method", http.MethodGet)
		req.Header.Set("Access-Control-Request-Headers", "authorization")
		rr := httptest.NewRecorder()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("プリフライトで次のハンドラーが呼ばれた")
		})
		APICORS(next).ServeHTTP(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusNoContent)
		}
		wantHeaders := map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET",
			"Access-Control-Allow-Headers": "Authorization",
			"Access-Control-Max-Age":       "7200",
		}
		for name, want := range wantHeaders {
			if got := rr.Header().Get(name); got != want {
				t.Errorf("%s = %q、期待値 = %q", name, got, want)
			}
		}
	})

	t.Run("OriginなどのないOPTIONSはプリフライトとして扱わない", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		APICORS(testHandler()).ServeHTTP(rr, httptest.NewRequest(http.MethodOptions, "/api/v2/anime/1", nil))

		if rr.Code != http.StatusOK {
			t.Errorf("ステータスコード = %d、期待値 = 次のハンドラーの応答", rr.Code)
		}
	})

	t.Run("APIのパスでなければ付けない", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodOptions, "/works", nil)
		req.Header.Set("Origin", "https://client.example.com")
		req.Header.Set("Access-Control-Request-Method", http.MethodGet)
		rr := httptest.NewRecorder()
		APICORS(testHandler()).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("ステータスコード = %d、期待値 = 次のハンドラーの応答", rr.Code)
		}
		if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("Access-Control-Allow-Origin = %q、期待値 = 付かない", got)
		}
	})
}

func TestRecoverer(t *testing.T) {
	t.Parallel()

	panicking := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("テストのパニック")
	})
	m := NewRecoverer(apierror.NewWriter("https://example.com"))

	t.Run("APIのパスにはProblem Detailsの500を返す", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		m.Middleware(panicking).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v2/openapi.yaml", nil))

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusInternalServerError)
		}
		if got := rr.Header().Get("Content-Type"); got != apierror.ContentType {
			t.Errorf("Content-Type = %q、期待値 = %q", got, apierror.ContentType)
		}
		if !strings.Contains(rr.Body.String(), `"status":500`) {
			t.Errorf("本文 = %q、Problem Detailsの500を期待", rr.Body.String())
		}
	})

	t.Run("APIのパスでなければchiのRecovererに任せる", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()
		m.Middleware(panicking).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/works", nil))

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusInternalServerError)
		}
		if got := rr.Header().Get("Content-Type"); got == apierror.ContentType {
			t.Error("APIのパスでないのにProblem Detailsを返した")
		}
	})

	t.Run("応答の中断のパニックは投げ直す", func(t *testing.T) {
		t.Parallel()

		aborting := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic(http.ErrAbortHandler)
		})
		defer func() {
			if rvr := recover(); rvr != http.ErrAbortHandler {
				t.Errorf("recover() = %v、期待値 = http.ErrAbortHandler", rvr)
			}
		}()
		m.Middleware(aborting).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/v2/openapi.yaml", nil))
	})
}
