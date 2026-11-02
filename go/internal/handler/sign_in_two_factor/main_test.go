package sign_in_two_factor

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/ratelimit"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/validator"
)

func TestMain(m *testing.M) {
	os.Exit(testutil.SetupTestMain(m))
}

// testEnvは、テスト用トランザクションの中で組み立てたHandlerと、準備に使うものをまとめる。
// 照合とセッションの作成はトランザクションを開かないため、準備した行をコミットせずに使える。
type testEnv struct {
	handler    *Handler
	sessionMgr *session.Manager
	tx         *sql.Tx
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	cfg := &config.Config{Env: "test", Domain: "test.annict.com", CookieDomain: ".test.annict.com", SessionSecure: "false", SessionHTTPOnly: "true"}
	sessionMgr := session.NewManager(repository.NewSessionRepository(queries), cfg)

	handler := NewHandler(
		cfg,
		sessionMgr,
		session.NewFlashManager(cfg.CookieDomain, false),
		ratelimit.NewLimiter(repository.NewRateLimitRepository(queries)),
		usecase.NewGetSignInTwoFactorStatusUsecase(repository.NewUserRepository(queries), repository.NewUserTwoFactorAuthRepository(queries)),
		usecase.NewCreateSignInTwoFactorUsecase(
			testutil.NewTestTwoFactorKey(t),
			validator.NewSignInTwoFactorCreateValidator(),
			repository.NewUserRepository(queries),
			repository.NewUserTwoFactorAuthRepository(queries),
		),
		usecase.NewCreateSessionUsecase(repository.NewSessionRepository(queries)),
	)
	return &testEnv{handler: handler, sessionMgr: sessionMgr, tx: tx}
}

// startSignInTwoFactorは、第1の要素を確かめたあとと同じくログイン途中の状態を作り、そのセッションのCookieを返す。
func (e *testEnv) startSignInTwoFactor(t *testing.T, userID model.UserID, now time.Time) *http.Cookie {
	t.Helper()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sign_in/password", nil)
	if err := e.sessionMgr.SetValue(req.Context(), rr, req, "sign_in_email", "user@example.com"); err != nil {
		t.Fatalf("SetValue()のエラー = %v", err)
	}
	cookie := sessionCookie(t, rr)
	req.AddCookie(cookie)
	if err := e.sessionMgr.StartSignInTwoFactor(req.Context(), httptest.NewRecorder(), req, userID, now); err != nil {
		t.Fatalf("StartSignInTwoFactor()のエラー = %v", err)
	}
	return cookie
}

// sessionCookieは応答が設定したセッションのCookieを返す。
func sessionCookie(t *testing.T, rr *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()

	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == session.SessionKey {
			return cookie
		}
	}
	t.Fatal("セッションのCookieが設定されていない")
	return nil
}

// newRequestは、セッションのCookieとロケールを載せたリクエストを作る。cookieがnilならCookieを載せない。
func newRequest(method, target string, form url.Values, cookie *http.Cookie) *http.Request {
	var req *http.Request
	if form != nil {
		req = httptest.NewRequest(method, target, strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req = httptest.NewRequest(method, target, nil)
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	return req.WithContext(i18n.SetLocale(req.Context(), "ja"))
}

// serveはserve.goと同じパスでハンドラーを登録したルーターにリクエストを渡す。
func (e *testEnv) serve(req *http.Request) *httptest.ResponseRecorder {
	r := chi.NewRouter()
	r.Get("/sign_in/two_factor", e.handler.New)
	r.Post("/sign_in/two_factor", e.handler.Create)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

// assertNoStoreは応答にキャッシュさせない指示が付いていることを確かめる。
func assertNoStore(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()

	if got := rr.Header().Get("Cache-Control"); got != "private, no-store" {
		t.Errorf("Cache-Control = %q、期待値 = %q", got, "private, no-store")
	}
}

// assertRedirectは、指定した場所へ303で送ることを確かめる。
func assertRedirect(t *testing.T, rr *httptest.ResponseRecorder, want string) {
	t.Helper()

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusSeeOther)
	}
	if location := rr.Header().Get("Location"); location != want {
		t.Errorf("Location = %q、期待値 = %q", location, want)
	}
}

// assertRestartFlashは、応答がログインのやり直しを伝えるエラーのフラッシュメッセージを設定したことを確かめる。
func (e *testEnv) assertRestartFlash(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()

	flashReq := newRequest(http.MethodGet, "/sign_in", nil, nil)
	for _, responseCookie := range rr.Result().Cookies() {
		flashReq.AddCookie(responseCookie)
	}
	flash := e.handler.flashMgr.GetFlash(httptest.NewRecorder(), flashReq)
	if flash == nil || flash.Type != session.FlashError || flash.Message != i18n.T(flashReq.Context(), "flash_sign_in_two_factor_restart") {
		t.Errorf("フラッシュメッセージ = %+v、ログインのやり直しを伝えるエラーを期待", flash)
	}
}

// assertContainsは本文にすべての文字列が含まれることを確かめる。
func assertContains(t *testing.T, body string, expected ...string) {
	t.Helper()

	for _, s := range expected {
		if !strings.Contains(body, s) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", s)
		}
	}
}
