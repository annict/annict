package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/apierror"
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/ratelimit"
)

// testAPILimiterは、受け取った入力を残し、countの値で判定するレート制限
type testAPILimiter struct {
	count   int
	resetAt time.Time
	err     error
	inputs  []ratelimit.CheckInput
}

func (l *testAPILimiter) Check(_ context.Context, input ratelimit.CheckInput) (*ratelimit.CheckResult, error) {
	l.inputs = append(l.inputs, input)
	if l.err != nil {
		return nil, l.err
	}
	return &ratelimit.CheckResult{
		Allowed:   l.count <= input.Limit,
		Count:     l.count,
		Remaining: max(input.Limit-l.count, 0),
		ResetAt:   l.resetAt,
	}, nil
}

// newTestAPIRateLimitHandlerは、CountMiddlewareで数えてからEnforceMiddlewareで判定し、nextを呼ぶハンドラーを返す
func newTestAPIRateLimitHandler(cfg *config.Config, limiter *testAPILimiter, now time.Time, next http.Handler) http.Handler {
	l := NewAPIRateLimit(cfg, limiter, apierror.NewWriter("https://example.com"))
	l.now = func() time.Time { return now }
	return l.CountMiddleware(l.EnforceMiddleware(next))
}

func TestAPIRateLimit_Keys(t *testing.T) {
	t.Parallel()

	digest := sha256.Sum256([]byte("secret-token"))

	tests := []struct {
		name          string
		authorization string
		wantKey       string
		wantLimit     int
	}{
		{
			name:          "トークンを送ったリクエストは、トークンのSHA-256と接続元IPの組で数える",
			authorization: "Bearer secret-token",
			wantKey:       "api:token:" + hex.EncodeToString(digest[:]) + ":203.0.113.1",
			wantLimit:     600,
		},
		{name: "トークンを送らないリクエストは接続元IPで数える", wantKey: "api:ip:203.0.113.1", wantLimit: 60},
		{name: "Bearerでない認証はトークンの無いリクエストとして数える", authorization: "Basic dXNlcjpwYXNz", wantKey: "api:ip:203.0.113.1", wantLimit: 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			limiter := &testAPILimiter{count: 1, resetAt: time.Now().Add(time.Minute)}
			req := httptest.NewRequest(http.MethodGet, "/api/v2/anime/1", nil)
			req.Header.Set("CF-Connecting-IP", "203.0.113.1")
			if tt.authorization != "" {
				req.Header.Set("Authorization", tt.authorization)
			}
			newTestAPIRateLimitHandler(&config.Config{}, limiter, time.Now(), testHandler()).ServeHTTP(httptest.NewRecorder(), req)

			if len(limiter.inputs) != 1 {
				t.Fatalf("数えた回数 = %d、期待値 = 1", len(limiter.inputs))
			}
			input := limiter.inputs[0]
			if input.Key != tt.wantKey {
				t.Errorf("Key = %q、期待値 = %q", input.Key, tt.wantKey)
			}
			if input.Limit != tt.wantLimit {
				t.Errorf("Limit = %d、期待値 = %d", input.Limit, tt.wantLimit)
			}
			if input.Window != time.Minute {
				t.Errorf("Window = %v、期待値 = 1m", input.Window)
			}
		})
	}
}

func TestAPIRateLimit_Headers(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 3, 12, 34, 50, 500*int(time.Millisecond), time.UTC)
	resetAt := time.Date(2026, 11, 3, 12, 35, 0, 0, time.UTC)

	t.Run("上限までは通し、残量を付ける", func(t *testing.T) {
		t.Parallel()

		limiter := &testAPILimiter{count: 60, resetAt: resetAt}
		rr := httptest.NewRecorder()
		newTestAPIRateLimitHandler(&config.Config{}, limiter, now, testHandler()).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v2/anime/1", nil))

		if rr.Code != http.StatusOK {
			t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
		}
		wantHeaders := map[string]string{
			"RateLimit-Limit":     "60",
			"RateLimit-Remaining": "0",
			// 窓の終わりまでの9.5秒は切り上げる
			"RateLimit-Reset": "10",
		}
		for name, want := range wantHeaders {
			if got := rr.Header().Get(name); got != want {
				t.Errorf("%s = %q、期待値 = %q", name, got, want)
			}
		}
	})

	t.Run("上限を超えたら429を返す", func(t *testing.T) {
		t.Parallel()

		limiter := &testAPILimiter{count: 61, resetAt: resetAt}
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("上限を超えたリクエストで次のハンドラーが呼ばれた")
		})
		rr := httptest.NewRecorder()
		newTestAPIRateLimitHandler(&config.Config{}, limiter, now, next).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v2/anime/1", nil))

		if rr.Code != http.StatusTooManyRequests {
			t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusTooManyRequests)
		}
		if got := rr.Header().Get("Retry-After"); got != "10" {
			t.Errorf("Retry-After = %q、期待値 = 10", got)
		}
		if got := rr.Header().Get("Content-Type"); got != apierror.ContentType {
			t.Errorf("Content-Type = %q、期待値 = %q", got, apierror.ContentType)
		}
	})
}

func TestAPIRateLimit_Bypass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cfg         *config.Config
		limiter     *testAPILimiter
		wantChecked bool
	}{
		{name: "カウンターの失敗ではAPIを止めない", cfg: &config.Config{}, limiter: &testAPILimiter{err: errors.New("DBに接続できません")}, wantChecked: true},
		{name: "レート制限を無効にした環境では数えない", cfg: &config.Config{DisableRateLimit: true}, limiter: &testAPILimiter{count: 100}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()
			newTestAPIRateLimitHandler(tt.cfg, tt.limiter, time.Now(), testHandler()).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v2/anime/1", nil))

			if rr.Code != http.StatusOK {
				t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
			}
			if got := rr.Header().Get("RateLimit-Limit"); got != "" {
				t.Errorf("RateLimit-Limit = %q、期待値 = 付かない", got)
			}
			if checked := len(tt.limiter.inputs) > 0; checked != tt.wantChecked {
				t.Errorf("数えたか = %v、期待値 = %v", checked, tt.wantChecked)
			}
		})
	}
}

func TestAPIRateLimit_PanicAfterCounting(t *testing.T) {
	t.Parallel()

	limiter := &testAPILimiter{count: 1, resetAt: time.Now().Add(time.Minute)}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("APIの処理中のパニック")
	})
	rr := httptest.NewRecorder()
	newTestAPIRateLimitHandler(&config.Config{}, limiter, time.Now(), next).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v2/anime/1", nil))

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusInternalServerError)
	}
	if got := rr.Header().Get("Content-Type"); got != apierror.ContentType {
		t.Errorf("Content-Type = %q、期待値 = %q", got, apierror.ContentType)
	}
	if got := rr.Header().Get("RateLimit-Limit"); got != "60" {
		t.Errorf("RateLimit-Limit = %q、期待値 = 60", got)
	}
	if len(limiter.inputs) != 1 {
		t.Errorf("計数の回数 = %d、期待値 = 1", len(limiter.inputs))
	}
}
