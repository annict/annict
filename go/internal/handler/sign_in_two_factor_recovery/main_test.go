package sign_in_two_factor_recovery

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/auth"
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

// testRecoveryCodesは、テストで発行したことにするリカバリーコード。
var testRecoveryCodes = []string{"abcd-2345", "efgh-6789"}

// testEnvは、共有テストDBに対して組み立てたHandlerと、準備に使うものをまとめる。
// リカバリーコードでのログインはトランザクションを開くため、テストは準備した行をコミットしておく。
type testEnv struct {
	db         *sql.DB
	handler    *Handler
	sessionMgr *session.Manager
	limiter    *ratelimit.Limiter
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	db := testutil.GetTestDB()
	queries := query.New(db)
	cfg := &config.Config{Env: "test", Domain: "test.annict.com", CookieDomain: ".test.annict.com", SessionSecure: "false", SessionHTTPOnly: "true"}
	sessionMgr := session.NewManager(repository.NewSessionRepository(queries), cfg)
	limiter := ratelimit.NewLimiter(repository.NewRateLimitRepository(queries))
	userRepo := repository.NewUserRepository(queries)
	authRepo := repository.NewUserTwoFactorAuthRepository(queries)

	handler := NewHandler(
		cfg,
		sessionMgr,
		session.NewFlashManager(cfg.CookieDomain, false),
		limiter,
		usecase.NewGetSignInTwoFactorStatusUsecase(userRepo, authRepo),
		usecase.NewCreateSignInTwoFactorRecoveryUsecase(
			db,
			testutil.NewTestTwoFactorKey(t),
			validator.NewSignInTwoFactorRecoveryCreateValidator(),
			userRepo,
			authRepo,
			repository.NewUserTwoFactorRecoveryCodeRepository(queries),
			usecase.NewCreateSessionUsecase(repository.NewSessionRepository(queries)),
		),
	)
	return &testEnv{db: db, handler: handler, sessionMgr: sessionMgr, limiter: limiter}
}

// createUserは、二要素認証を有効にしてtestRecoveryCodesを発行したユーザーをコミットして作る。
// enableがfalseのときは、二要素認証の設定を作らない。
func (e *testEnv) createUser(t *testing.T, enable bool) model.UserID {
	t.Helper()

	tx, err := e.db.Begin()
	if err != nil {
		t.Fatalf("準備用トランザクションの開始に失敗: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	userID := testutil.NewUserBuilder(t, tx).Build()
	if enable {
		testutil.EnableTwoFactorAuthForTest(t, tx, userID)
		key := testutil.NewTestTwoFactorKey(t)
		digests := make([]string, len(testRecoveryCodes))
		for i, code := range testRecoveryCodes {
			digests[i] = key.RecoveryCodeDigest(auth.NormalizeRecoveryCode(code))
		}
		if err := repository.NewUserTwoFactorRecoveryCodeRepository(query.New(e.db).WithTx(tx)).CreateAll(context.Background(), userID, digests); err != nil {
			t.Fatalf("CreateAll()のエラー = %v", err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("準備用トランザクションのコミットに失敗: %v", err)
	}
	t.Cleanup(func() {
		_, _ = e.db.Exec("DELETE FROM user_two_factor_recovery_codes WHERE user_id = $1", int64(userID))
		_, _ = e.db.Exec("DELETE FROM user_two_factor_auths WHERE user_id = $1", int64(userID))
	})
	return userID
}

// startSignInTwoFactorは、第1の要素を確かめたあとと同じくログイン途中の状態を作り、そのセッションのCookieを返す。
func (e *testEnv) startSignInTwoFactor(t *testing.T, userID model.UserID) *http.Cookie {
	t.Helper()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sign_in/password", nil)
	if err := e.sessionMgr.SetValue(req.Context(), rr, req, "sign_in_email", "user@example.com"); err != nil {
		t.Fatalf("SetValue()のエラー = %v", err)
	}
	cookie := sessionCookie(t, rr)
	req.AddCookie(cookie)
	if err := e.sessionMgr.StartSignInTwoFactor(req.Context(), httptest.NewRecorder(), req, userID, time.Now()); err != nil {
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
	r.Get("/sign_in/two_factor/recovery", e.handler.New)
	r.Post("/sign_in/two_factor/recovery", e.handler.Create)
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

// assertNoNewSessionCookieは、応答がログイン前と違うセッションのCookieを設定していないことを確かめる。
func assertNoNewSessionCookie(t *testing.T, rr *httptest.ResponseRecorder, before *http.Cookie) {
	t.Helper()

	for _, c := range rr.Result().Cookies() {
		if c.Name == session.SessionKey && c.Value != before.Value {
			t.Error("ログインさせない応答で新しいセッションのCookieが設定された")
		}
	}
}

// assertPendingClearedは、リクエストのCookieが指すセッションからログイン途中の状態が消えていることを確かめる。
func (e *testEnv) assertPendingCleared(t *testing.T, cookie *http.Cookie) {
	t.Helper()

	pendingReq := newRequest(http.MethodGet, "/", nil, cookie)
	if _, ok, err := e.sessionMgr.SignInTwoFactorUserID(pendingReq.Context(), pendingReq, time.Now()); err != nil || ok {
		t.Errorf("SignInTwoFactorUserID() = (ok:%t, %v)、ログイン途中の状態が消えていることを期待", ok, err)
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
