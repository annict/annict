package middleware

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/ratelimit"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/testutil"
)

// newTestRateLimitMiddlewareはトランザクション内で数える限定器を使い、上限をlimitにした
// ミドルウェアを返す
func newTestRateLimitMiddleware(t *testing.T, limit int) *RateLimitMiddleware {
	t.Helper()

	db, tx := testutil.SetupTx(t)
	return newRateLimitMiddlewareWithTx(db, tx, limit)
}

func newRateLimitMiddlewareWithTx(db *sql.DB, tx *sql.Tx, limit int) *RateLimitMiddleware {
	queries := query.New(db)
	limiter := ratelimit.NewLimiter(repository.NewRateLimitRepository(queries)).WithTx(tx)
	cfg := &config.Config{
		Domain:                "annict.test",
		WebRateLimitPerMinute: limit,
	}
	sessions := session.NewManager(repository.NewSessionRepository(queries.WithTx(tx)), cfg)
	return NewRateLimitMiddleware(cfg, limiter, sessions)
}

// newRateLimitRequestは接続元IPをclientIPにしたリクエストを返す
func newRateLimitRequest(method string, target string, clientIP string) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	req.Host = "annict.test"
	req.Header.Set("CF-Connecting-IP", clientIP)
	req.Header.Set("Accept-Language", "ja")
	return req
}

func TestRateLimitMiddleware_ExceedsLimit(t *testing.T) {
	t.Parallel()

	handler := newTestRateLimitMiddleware(t, 2).Middleware(testHandler())
	clientIP := testutil.UniqueRateLimitPrefix(t)

	for i := 0; i < 2; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, newRateLimitRequest(http.MethodGet, "/works/1", clientIP))
		if rr.Code != http.StatusOK {
			t.Fatalf("%d回目のステータスコード = %d、期待値 = %d", i+1, rr.Code, http.StatusOK)
		}
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, newRateLimitRequest(http.MethodGet, "/works/1", clientIP))

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("上限を超えたリクエストのステータスコード = %d、期待値 = %d", rr.Code, http.StatusTooManyRequests)
	}

	retryAfter, err := strconv.Atoi(rr.Header().Get("Retry-After"))
	if err != nil {
		t.Fatalf("Retry-Afterが整数ではない: %q", rr.Header().Get("Retry-After"))
	}
	if retryAfter < 1 || retryAfter > 60 {
		t.Errorf("Retry-After = %d、期待値 = 1から60の秒数", retryAfter)
	}

	if got := rr.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q、期待値 = nosniff", got)
	}
	if got := rr.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q、期待値 = text/html; charset=utf-8", got)
	}
	if body := rr.Body.String(); !strings.Contains(body, "アクセスが集中しています") {
		t.Error("レスポンスボディに429ページのタイトルが含まれていない")
	}
	if strings.Contains(rr.Body.String(), "OK") {
		t.Error("上限を超えたリクエストが後続のハンドラーに渡された")
	}
}

// TestRateLimitMiddleware_LoggedInLocaleは429でもユーザーの設定言語を優先することを確認する。
func TestRateLimitMiddleware_LoggedInLocale(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	userID := testutil.NewUserBuilder(t, tx).WithLocale("en").Build()
	sessionID := testutil.NewSessionBuilder(t, tx).WithUserID(userID).Build()
	handler := newRateLimitMiddlewareWithTx(db, tx, 1).Middleware(testHandler())
	clientIP := testutil.UniqueRateLimitPrefix(t)

	for i := 0; i < 2; i++ {
		req := newRateLimitRequest(http.MethodGet, "/works/1", clientIP)
		req.AddCookie(&http.Cookie{Name: session.SessionKey, Value: sessionID})
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if i == 0 {
			if rr.Code != http.StatusOK {
				t.Fatalf("最初のリクエストのステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
			}
			continue
		}
		if rr.Code != http.StatusTooManyRequests {
			t.Fatalf("超過時のステータスコード = %d、期待値 = %d", rr.Code, http.StatusTooManyRequests)
		}
		if !strings.Contains(rr.Body.String(), `<html lang="en"`) {
			t.Error("429ページのhtml要素の言語がユーザーの設定言語ではない")
		}
		if !strings.Contains(rr.Body.String(), "Too many requests") {
			t.Error("429ページの文言がユーザーの設定言語ではない")
		}
	}
}

// TestRateLimitMiddleware_SessionErrorは、429でユーザーを取得できなかったときに
// ブラウザの言語で429ページを返すことを確認する
func TestRateLimitMiddleware_SessionError(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	_, sessionTx := testutil.SetupTx(t)
	// 終了したトランザクションでセッションを引かせ、ユーザーの取得を失敗させる
	if err := sessionTx.Rollback(); err != nil {
		t.Fatalf("トランザクションのロールバックに失敗: %v", err)
	}
	mw := newRateLimitMiddlewareWithTx(db, tx, 1)
	mw.sessions = session.NewManager(repository.NewSessionRepository(query.New(db).WithTx(sessionTx)), mw.cfg)
	handler := mw.Middleware(testHandler())
	clientIP := testutil.UniqueRateLimitPrefix(t)

	codes := make([]int, 0, 2)
	var rr *httptest.ResponseRecorder
	for i := 0; i < 2; i++ {
		req := newRateLimitRequest(http.MethodGet, "/works/1", clientIP)
		req.AddCookie(&http.Cookie{Name: session.SessionKey, Value: "dummy-session-id"})
		rr = httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		codes = append(codes, rr.Code)
	}

	if codes[0] != http.StatusOK || codes[1] != http.StatusTooManyRequests {
		t.Fatalf("ステータスコード = %v、期待値 = [200 429]", codes)
	}
	if rr.Header().Get("Retry-After") == "" {
		t.Error("429にRetry-Afterが付いていない")
	}
	body := rr.Body.String()
	if !strings.Contains(body, `<html lang="ja"`) || !strings.Contains(body, "アクセスが集中しています") {
		t.Error("429ページがブラウザの言語 (日本語) で描画されていない")
	}
	if strings.Contains(body, "OK") {
		t.Error("上限を超えたリクエストが後続のハンドラーに渡された")
	}
}

// TestRateLimitMiddleware_CountsPerClientIPは接続元IPごとに別々に数えることを確認する
func TestRateLimitMiddleware_CountsPerClientIP(t *testing.T) {
	t.Parallel()

	handler := newTestRateLimitMiddleware(t, 1).Middleware(testHandler())
	prefix := testutil.UniqueRateLimitPrefix(t)

	for _, clientIP := range []string{prefix + "_a", prefix + "_b"} {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, newRateLimitRequest(http.MethodGet, "/", clientIP))
		if rr.Code != http.StatusOK {
			t.Errorf("接続元IP %qの最初のリクエストのステータスコード = %d、期待値 = %d", clientIP, rr.Code, http.StatusOK)
		}
	}
}

// TestRateLimitMiddleware_ExemptRequestsは対象外のリクエストを数えないことを確認する。
// 上限1で対象外のリクエストを2回ずつ送り、そのあとの通常のリクエストも通ることで、
// 対象外のリクエストが件数に加わっていないことを確かめる。
func TestRateLimitMiddleware_ExemptRequests(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		host   string
		path   string
	}{
		{name: "Go版の静的ファイル", method: http.MethodGet, host: "annict.test", path: "/static/app.css"},
		{name: "Rails版のアセット", method: http.MethodGet, host: "annict.test", path: "/assets/application-abc123.js"},
		{name: "Rails版のpublic配下の画像", method: http.MethodGet, host: "annict.test", path: "/images/no-image.jpg"},
		{name: "Rails版のpublic配下のJavaScript", method: http.MethodGet, host: "annict.test", path: "/js/fontawesome.min.js"},
		{name: "favicon", method: http.MethodGet, host: "annict.test", path: "/favicon.ico"},
		{name: "robots.txt", method: http.MethodGet, host: "annict.test", path: "/robots.txt"},
		{name: "ヘルスチェック", method: http.MethodGet, host: "annict.test", path: "/health"},
		{name: "StripeのWebhook", method: http.MethodPost, host: "annict.test", path: "/webhooks/stripe"},
		{name: "APIホスト", method: http.MethodGet, host: "api.annict.test", path: "/v1/works"},
		{name: "ポート付きのAPIホスト", method: http.MethodPost, host: "api.annict.test:8080", path: "/graphql"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := newTestRateLimitMiddleware(t, 1).Middleware(testHandler())
			clientIP := testutil.UniqueRateLimitPrefix(t)

			for i := 0; i < 2; i++ {
				req := newRateLimitRequest(tt.method, tt.path, clientIP)
				req.Host = tt.host
				rr := httptest.NewRecorder()
				handler.ServeHTTP(rr, req)
				if rr.Code != http.StatusOK {
					t.Fatalf("%d回目のステータスコード = %d、期待値 = %d", i+1, rr.Code, http.StatusOK)
				}
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, newRateLimitRequest(http.MethodGet, "/", clientIP))
			if rr.Code != http.StatusOK {
				t.Errorf("対象外のリクエストの後の通常のリクエストのステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
			}
		})
	}
}

// TestRateLimitMiddleware_CountsRailsPagesAndManifestは、静的ファイルと名前の似たパスや
// Rails版へ中継するページを数えることを確認する
func TestRateLimitMiddleware_CountsRailsPagesAndManifest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
	}{
		{name: "Web App Manifest", path: "/manifest.json"},
		{name: "Rails版のページ", path: "/works/1/records"},
		{name: "接頭辞が静的ファイルに似たパス", path: "/statics"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := newTestRateLimitMiddleware(t, 1).Middleware(testHandler())
			clientIP := testutil.UniqueRateLimitPrefix(t)

			codes := make([]int, 0, 2)
			for i := 0; i < 2; i++ {
				rr := httptest.NewRecorder()
				handler.ServeHTTP(rr, newRateLimitRequest(http.MethodGet, tt.path, clientIP))
				codes = append(codes, rr.Code)
			}

			if codes[0] != http.StatusOK || codes[1] != http.StatusTooManyRequests {
				t.Errorf("ステータスコード = %v、期待値 = [200 429]", codes)
			}
		})
	}
}

func TestRateLimitMiddleware_Disabled(t *testing.T) {
	t.Parallel()

	mw := newTestRateLimitMiddleware(t, 1)
	mw.cfg.DisableRateLimit = true
	handler := mw.Middleware(testHandler())
	clientIP := testutil.UniqueRateLimitPrefix(t)

	for i := 0; i < 3; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, newRateLimitRequest(http.MethodGet, "/", clientIP))
		if rr.Code != http.StatusOK {
			t.Errorf("%d回目のステータスコード = %d、期待値 = %d", i+1, rr.Code, http.StatusOK)
		}
	}
}

// TestRateLimitMiddleware_LimiterErrorはカウンターの読み書きに失敗したときに制限を掛けずに
// 通すことを確認する
func TestRateLimitMiddleware_LimiterError(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	// 終了したトランザクションで数えさせ、カウンターの読み書きを失敗させる
	if err := tx.Rollback(); err != nil {
		t.Fatalf("トランザクションのロールバックに失敗: %v", err)
	}
	handler := newRateLimitMiddlewareWithTx(db, tx, 1).Middleware(testHandler())
	clientIP := testutil.UniqueRateLimitPrefix(t)

	for i := 0; i < 2; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, newRateLimitRequest(http.MethodGet, "/", clientIP))
		if rr.Code != http.StatusOK {
			t.Errorf("%d回目のステータスコード = %d、期待値 = %d", i+1, rr.Code, http.StatusOK)
		}
	}
}

func TestRateLimitMiddleware_RetryAfterSeconds(t *testing.T) {
	t.Parallel()

	resetAt := time.Date(2026, 10, 14, 12, 35, 0, 0, time.UTC)
	tests := []struct {
		name string
		now  time.Time
		want int
	}{
		{name: "窓の始め", now: time.Date(2026, 10, 14, 12, 34, 0, 0, time.UTC), want: 60},
		{name: "端数は切り上げる", now: time.Date(2026, 10, 14, 12, 34, 41, 500_000_000, time.UTC), want: 19},
		{name: "窓の終わりぎわでも1秒", now: time.Date(2026, 10, 14, 12, 34, 59, 999_000_000, time.UTC), want: 1},
		{name: "窓を過ぎていても1秒", now: time.Date(2026, 10, 14, 12, 35, 1, 0, time.UTC), want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mw := &RateLimitMiddleware{now: func() time.Time { return tt.now }}
			if got := mw.retryAfterSeconds(resetAt); got != tt.want {
				t.Errorf("retryAfterSeconds = %d、期待値 = %d", got, tt.want)
			}
		})
	}
}
