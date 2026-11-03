package settings_two_factor_auth

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/i18n"
	authMiddleware "github.com/annict/annict/go/internal/middleware"
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

// newTestHandlerは共有テストDBに対してHandlerを組み立てる。
// 有効化のUseCaseはdbから自前のトランザクションを開くため、テストは準備した行をコミットしておく。
func newTestHandler(t *testing.T, db *sql.DB) *Handler {
	t.Helper()

	queries := query.New(db)
	cfg := &config.Config{Env: "test", Domain: "test.annict.com"}
	key := testutil.NewTestTwoFactorKey(t)
	authRepo := repository.NewUserTwoFactorAuthRepository(queries)
	recoveryCodeRepo := repository.NewUserTwoFactorRecoveryCodeRepository(queries)

	return NewHandler(
		cfg,
		session.NewManager(repository.NewSessionRepository(queries), cfg),
		session.NewFlashManager(cfg.CookieDomain, false),
		nil,
		ratelimit.NewLimiter(repository.NewRateLimitRepository(queries)),
		usecase.NewGetTwoFactorAuthStatusUsecase(authRepo, recoveryCodeRepo),
		usecase.NewPrepareTwoFactorAuthUsecase(key, authRepo),
		usecase.NewGetPendingTwoFactorAuthUsecase(key, authRepo),
		usecase.NewEnableTwoFactorAuthUsecase(db, key, validator.NewSettingsTwoFactorAuthCreateValidator(), authRepo, recoveryCodeRepo, repository.NewUserRepository(queries), nil),
		usecase.NewDisableTwoFactorAuthUsecase(db, key, validator.NewSettingsTwoFactorAuthDeleteValidator(), repository.NewUserRepository(queries), authRepo, recoveryCodeRepo, nil),
	)
}

// createCommittedUserはユーザーを作成してコミットし、テストの終了時に二要素認証の行とユーザーを消す。
func createCommittedUser(t *testing.T, db *sql.DB) *model.User {
	t.Helper()

	return createCommittedUserWithEncryptedPassword(t, db, "encrypted_test_password")
}

// createCommittedUserWithPasswordは、passwordでログインできるユーザーを作成してコミットする。
func createCommittedUserWithPassword(t *testing.T, db *sql.DB, password string) *model.User {
	t.Helper()

	hashed, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword()のエラー = %v", err)
	}
	return createCommittedUserWithEncryptedPassword(t, db, hashed)
}

// createCommittedUserWithEncryptedPasswordは、パスワードのハッシュを指定したユーザーを作成してコミットし、
// テストの終了時に二要素認証の行とユーザーを消す。
func createCommittedUserWithEncryptedPassword(t *testing.T, db *sql.DB, encryptedPassword string) *model.User {
	t.Helper()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("準備用トランザクションの開始に失敗: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	userID := testutil.NewUserBuilder(t, tx).WithEncryptedPassword(encryptedPassword).Build()
	var username string
	if err := tx.QueryRow("SELECT username FROM users WHERE id = $1", int64(userID)).Scan(&username); err != nil {
		t.Fatalf("ユーザーの取得に失敗: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("準備用トランザクションのコミットに失敗: %v", err)
	}

	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM user_two_factor_recovery_codes WHERE user_id = $1", int64(userID))
		_, _ = db.Exec("DELETE FROM user_two_factor_auths WHERE user_id = $1", int64(userID))
		_, _ = db.Exec("DELETE FROM users WHERE id = $1", int64(userID))
	})
	return &model.User{ID: userID, Username: username}
}

// enableForTestはユーザーの二要素認証を有効にした状態にし、未使用のリカバリーコードをcount個作る。
func enableForTest(t *testing.T, db *sql.DB, userID model.UserID, count int) {
	t.Helper()

	if _, err := db.Exec("INSERT INTO user_two_factor_auths (user_id, secret_ciphertext, enabled_at) VALUES ($1, $2, NOW())", int64(userID), []byte("enabled-secret")); err != nil {
		t.Fatalf("二要素認証の設定の作成に失敗: %v", err)
	}
	for i := range count {
		if _, err := db.Exec("INSERT INTO user_two_factor_recovery_codes (user_id, code_digest) VALUES ($1, $2)", int64(userID), "digest-"+string(rune('a'+i))); err != nil {
			t.Fatalf("リカバリーコードの作成に失敗: %v", err)
		}
	}
}

// newRequestはログイン中のユーザーとロケールを載せたリクエストを作る。
func newRequest(t *testing.T, method, target string, form url.Values, user *model.User, locale string) *http.Request {
	t.Helper()

	var req *http.Request
	if form != nil {
		req = httptest.NewRequest(method, target, strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req = httptest.NewRequest(method, target, nil)
	}
	ctx := i18n.SetLocale(req.Context(), locale)
	ctx = context.WithValue(ctx, authMiddleware.UserContextKey, user)
	return req.WithContext(ctx)
}

// serveはserve.goと同じパスでハンドラーを登録したルーターにリクエストを渡す。
// HTMLのフォームはDELETEをPOSTの_methodで送るため、serve.goと同じくMethodOverrideを通す。
func serve(handler *Handler, req *http.Request) *httptest.ResponseRecorder {
	r := chi.NewRouter()
	r.Use(authMiddleware.MethodOverride)
	r.Get("/settings/two_factor_auth", handler.Show)
	r.Get("/settings/two_factor_auth/new", handler.New)
	r.Post("/settings/two_factor_auth", handler.Create)
	r.Delete("/settings/two_factor_auth", handler.Delete)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

// pendingSecretForTestは登録画面を開いて作られた、登録の途中の秘密鍵を復号して返す。
func pendingSecretForTest(t *testing.T, handler *Handler, user *model.User) string {
	t.Helper()

	output, err := handler.getPendingTwoFactorAuthUC.Execute(context.Background(), usecase.GetPendingTwoFactorAuthInput{User: user})
	if err != nil || output.Setup == nil {
		t.Fatalf("GetPendingTwoFactorAuthUsecase.Execute() = (%+v, %v)、秘密鍵を期待", output, err)
	}
	return output.Setup.Secret
}

// serveUnauthenticatedは、serve.goと同じくRequireAuthを付けたルーターに、未ログインのリクエストを渡す。
func serveUnauthenticated(handler *Handler, method, target string) *httptest.ResponseRecorder {
	requireAuth := authMiddleware.NewAuthMiddleware(handler.sessionManager).RequireAuth
	r := chi.NewRouter()
	r.With(requireAuth).Get("/settings/two_factor_auth", handler.Show)
	r.With(requireAuth).Get("/settings/two_factor_auth/new", handler.New)
	r.With(requireAuth).Post("/settings/two_factor_auth", handler.Create)
	r.With(requireAuth).Delete("/settings/two_factor_auth", handler.Delete)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest(method, target, nil))
	return rr
}

// assertRedirectToSignInは、ログイン画面へ送り、開こうとしたパスをbackに付けていることを確かめる。
func assertRedirectToSignIn(t *testing.T, rr *httptest.ResponseRecorder, path string) {
	t.Helper()

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusSeeOther)
	}
	want := "/sign_in?back=" + url.QueryEscape(path)
	if location := rr.Header().Get("Location"); location != want {
		t.Errorf("Location = %q、期待値 = %q", location, want)
	}
}

// assertNoStoreは応答にキャッシュさせない指示が付いていることを確かめる。
func assertNoStore(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()

	if got := rr.Header().Get("Cache-Control"); got != "private, no-store" {
		t.Errorf("Cache-Control = %q、期待値 = %q", got, "private, no-store")
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

// assertFocusedErrorSummaryは、エラーの要約が失敗理由を含み、フォーカスの移動先が1つだけあることを確かめる。
func assertFocusedErrorSummary(t *testing.T, body string, expected ...string) {
	t.Helper()

	summary := regexp.MustCompile(`(?s)<div[^>]*role="alert"[^>]*>.*?</div>`).FindString(body)
	assertContains(t, summary, append(expected, `tabindex="-1"`, "autofocus")...)
	if count := strings.Count(body, "autofocus"); count != 1 {
		t.Errorf("フォーカスの移動先の数 = %d、期待値 = 1", count)
	}
}
