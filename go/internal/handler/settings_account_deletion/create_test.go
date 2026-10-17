package settings_account_deletion

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/i18n"
	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/testutil"
)

func newCreateRequest(t *testing.T, user *model.User, username string) *http.Request {
	t.Helper()

	form := url.Values{"username": {username}}
	req := httptest.NewRequest(http.MethodPost, "/settings/account/deletion", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := i18n.SetLocale(req.Context(), "ja")
	ctx = context.WithValue(ctx, authMiddleware.UserContextKey, user)
	return req.WithContext(ctx)
}

func serveCreate(t *testing.T, handler *Handler, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()

	r := chi.NewRouter()
	r.Post("/settings/account/deletion", handler.Create)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

// TestCreateは、ユーザー名が一致すれば退会を受け付け、現在のセッションを破棄して
// 完了ページへリダイレクトすることを検証する。
func TestCreate(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("準備用トランザクションの開始に失敗: %v", err)
	}
	userID := testutil.NewUserBuilder(t, setupTx).WithUsername("deletion_create_user").WithEmail("deletion_create_user@example.com").Build()
	if err := setupTx.Commit(); err != nil {
		t.Fatalf("準備用トランザクションのコミットに失敗: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM river_job WHERE kind = 'delete_user' AND (args->>'user_id')::bigint = $1", int64(userID))
		_, _ = db.Exec("DELETE FROM river_job WHERE kind = 'send_account_deletion_email' AND args->>'email' = $1", "deletion_create_user@example.com")
		_, _ = db.Exec("DELETE FROM sessions WHERE user_id = $1", int64(userID))
		_, _ = db.Exec("DELETE FROM users WHERE id = $1", int64(userID))
	})

	handler, sessionManager := newTestHandlerWithQueries(t, db, query.New(db))
	sessionCookie := createSessionForTest(t, sessionManager, userID)

	req := newCreateRequest(t, &model.User{ID: userID, Username: "deletion_create_user"}, "deletion_create_user")
	req.AddCookie(sessionCookie)
	rr := serveCreate(t, handler, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusSeeOther)
	}
	if location := rr.Header().Get("Location"); location != "/settings/account/deletion/completed" {
		t.Errorf("Location = %q、期待値 = %q", location, "/settings/account/deletion/completed")
	}

	var deletedCookie *http.Cookie
	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == session.SessionKey {
			deletedCookie = cookie
		}
	}
	if deletedCookie == nil || deletedCookie.MaxAge != -1 {
		t.Errorf("セッションCookie = %v、MaxAge = -1で削除することを期待", deletedCookie)
	}
	sessionData, err := sessionManager.GetSession(context.Background(), sessionCookie.Value)
	if err != nil {
		t.Fatalf("セッションの取得に失敗: %v", err)
	}
	if sessionData != nil {
		t.Error("セッションの行が削除されていない")
	}

	var deletedAt sql.NullTime
	if err := db.QueryRow("SELECT deleted_at FROM users WHERE id = $1", int64(userID)).Scan(&deletedAt); err != nil {
		t.Fatalf("ユーザーの取得に失敗: %v", err)
	}
	if !deletedAt.Valid {
		t.Error("ユーザーのdeleted_atが立っていない")
	}
}

// TestCreate_UsernameMismatchは、ユーザー名が一致しなければ確認ページを422で再表示し、
// エラーを入力欄に関連付けて示すことを検証する。
func TestCreate_UsernameMismatch(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("準備用トランザクションの開始に失敗: %v", err)
	}
	userID := testutil.NewUserBuilder(t, tx).Build()
	if err := tx.Commit(); err != nil {
		t.Fatalf("準備用トランザクションのコミットに失敗: %v", err)
	}
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM users WHERE id = $1", int64(userID)) })
	handler, _ := newTestHandlerWithQueries(t, db, query.New(db))

	rr := serveCreate(t, handler, newCreateRequest(t, &model.User{ID: userID, Username: "mismatch_user"}, "other_user"))

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusUnprocessableEntity)
	}
	body := rr.Body.String()
	for _, expected := range []string{
		`action="/settings/account/deletion"`,
		`aria-invalid="true"`,
		`value="other_user"`,
		`aria-describedby="username-error-1"`,
		`<p id="username-error-1" role="alert">ユーザー名が一致しません</p>`,
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}
}

// TestCreate_KeptOAuthApplicationsは、OAuthアプリケーションが残っていれば受け付けず、
// フォームの代わりにアプリケーションの管理ページへの案内を422で示すことを検証する。
func TestCreate_KeptOAuthApplications(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("準備用トランザクションの開始に失敗: %v", err)
	}
	userID := testutil.NewUserBuilder(t, tx).Build()
	testutil.CreateTestOAuthApplication(t, tx, userID, false)
	if err := tx.Commit(); err != nil {
		t.Fatalf("準備用トランザクションのコミットに失敗: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM oauth_applications WHERE owner_id = $1", int64(userID))
		_, _ = db.Exec("DELETE FROM users WHERE id = $1", int64(userID))
	})
	handler, _ := newTestHandlerWithQueries(t, db, query.New(db))

	rr := serveCreate(t, handler, newCreateRequest(t, &model.User{ID: userID, Username: "oauth_owner"}, "oauth_owner"))

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusUnprocessableEntity)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `href="/oauth/applications"`) {
		t.Errorf("レスポンスに含まれていない文字列 = %q", `href="/oauth/applications"`)
	}
	if strings.Contains(body, `action="/settings/account/deletion"`) {
		t.Error("OAuthアプリケーションが残っているのにフォームが表示されている")
	}
}

// TestCreate_RequiresAuthは、未ログインのリクエストを受け付けずログイン画面へ送ることを検証する。
func TestCreate_RequiresAuth(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler, sessionManager := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.With(authMiddleware.NewAuthMiddleware(sessionManager).RequireAuth).Post("/settings/account/deletion", handler.Create)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/settings/account/deletion", nil))

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusSeeOther)
	}
	if location := rr.Header().Get("Location"); !strings.HasPrefix(location, "/sign_in") {
		t.Errorf("Location = %q、ログイン画面へのリダイレクトを期待", location)
	}
}

// createSessionForTestはユーザーのログインセッションを作成し、セッションCookieを返す。
func createSessionForTest(t *testing.T, sessionManager *session.Manager, userID model.UserID) *http.Cookie {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	if err := sessionManager.CreateSession(req.Context(), rr, req, userID); err != nil {
		t.Fatalf("セッションの作成に失敗: %v", err)
	}
	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == session.SessionKey {
			return cookie
		}
	}
	t.Fatal("セッションCookieが作成されなかった")
	return nil
}
