package settings_account_deletion

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/i18n"
	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/validator"
)

func newTestHandler(t *testing.T, db *sql.DB, tx *sql.Tx) (*Handler, *session.Manager) {
	t.Helper()

	return newTestHandlerWithQueries(t, db, query.New(db).WithTx(tx))
}

// newTestHandlerWithQueriesはqueriesを使うリポジトリでHandlerを組み立てる。受付のUseCaseは
// dbから自前のトランザクションを開くため、受付を最後まで通すテストはコミットした行とトランザクションを
// 使わないqueriesで組み立てる。
func newTestHandlerWithQueries(t *testing.T, db *sql.DB, queries *query.Queries) (*Handler, *session.Manager) {
	t.Helper()

	cfg := &config.Config{Env: "test", Domain: "test.annict.com"}
	sessionManager := session.NewManager(repository.NewSessionRepository(queries), cfg)
	oauthApplicationRepo := repository.NewOAuthApplicationRepository(queries)
	stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
	getAccountDeletionNewUC := usecase.NewGetAccountDeletionNewUsecase(
		stripeSubscriberRepo,
		repository.NewGumroadSubscriberRepository(queries),
		oauthApplicationRepo,
	)
	txInserter, err := river.NewClient(riverdatabasesql.New(db), &river.Config{})
	if err != nil {
		t.Fatalf("Riverクライアントの作成に失敗: %v", err)
	}
	createAccountDeletionUC := usecase.NewCreateAccountDeletionUsecase(
		db,
		repository.NewUserRepository(queries),
		stripeSubscriberRepo,
		dispatcher.NewDispatcher(nil, txInserter),
		validator.NewSettingsAccountDeletionCreateValidator(oauthApplicationRepo),
	)

	return NewHandler(cfg, sessionManager, nil, getAccountDeletionNewUC, createAccountDeletionUC), sessionManager
}

func newRequest(t *testing.T, user *model.User, locale string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/settings/account/deletion/new", nil)
	ctx := i18n.SetLocale(req.Context(), locale)
	ctx = context.WithValue(ctx, authMiddleware.UserContextKey, user)
	return req.WithContext(ctx)
}

func serveNew(t *testing.T, handler *Handler, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()

	r := chi.NewRouter()
	r.Get("/settings/account/deletion/new", handler.New)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

// TestNewは確認ページが、削除されるもの・残るものとユーザー名の入力欄を持つフォームを描画することを検証する。
func TestNew(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	userID := testutil.NewUserBuilder(t, tx).Build()
	handler, _ := newTestHandler(t, db, tx)

	rr := serveNew(t, handler, newRequest(t, &model.User{ID: userID, Username: "deletion_new_user"}, "ja"))

	if rr.Code != http.StatusOK {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q、期待値 = text/html; charset=utf-8", ct)
	}

	body := rr.Body.String()
	for _, expected := range []string{
		"<title>アカウントの削除 | Annict</title>",
		`<meta property="og:url" content="https://test.annict.com/settings/account/deletion/new">`,
		"削除されるもの",
		"退会後も残るもの",
		`action="/settings/account/deletion"`,
		`method="POST"`,
		`name="csrf_token"`,
		`<label class="label" for="username">`,
		"確認のため、ユーザー名「deletion_new_user」を入力してください",
		`id="username"`,
		`name="username"`,
		"退会する",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}
	for _, unexpected := range []string{
		"サポーター契約は即時に解約されます",
		"Gumroadでの解約が必要です",
		`href="/oauth/applications"`,
	} {
		if strings.Contains(body, unexpected) {
			t.Errorf("レスポンスに含まれるべきでない文字列 = %q", unexpected)
		}
	}
}

// TestNew_Englishは英語のロケールで文言が英語になることを検証する。
func TestNew_English(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	userID := testutil.NewUserBuilder(t, tx).Build()
	handler, _ := newTestHandler(t, db, tx)

	rr := serveNew(t, handler, newRequest(t, &model.User{ID: userID, Username: "english_user"}, "en"))

	if rr.Code != http.StatusOK {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}
	body := rr.Body.String()
	for _, expected := range []string{
		"<title>Delete Account | Annict</title>",
		"What will be deleted",
		"To confirm, type your username &#34;english_user&#34;",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}
}

// TestNew_Supportersはサポーター契約に応じた注意を表示することを検証する。
// Stripeは退会で即時解約され、Gumroadは利用者自身での解約が必要になるため、それぞれ別の注意を出す。
func TestNew_Supporters(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	userID := testutil.NewUserBuilder(t, tx).Build()
	stripeSubscriberID := testutil.NewStripeSubscriberBuilder(t, tx).WithStripeStatus("active").Build()
	gumroadSubscriberID := testutil.NewGumroadSubscriberBuilder(t, tx).Build()
	handler, _ := newTestHandler(t, db, tx)

	user := &model.User{
		ID:                  userID,
		Username:            "supporter_user",
		StripeSubscriberID:  &stripeSubscriberID,
		GumroadSubscriberID: &gumroadSubscriberID,
	}
	rr := serveNew(t, handler, newRequest(t, user, "ja"))

	if rr.Code != http.StatusOK {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}
	body := rr.Body.String()
	for _, expected := range []string{
		"サポーター契約は即時に解約されます",
		"残りの期間の料金は返金されません",
		"Gumroadでの解約が必要です",
		`action="/settings/account/deletion"`,
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}
}

// TestNew_KeptOAuthApplicationsは、論理削除されていないOAuthアプリケーションがあると
// 退会を受け付けないため、フォームを出さずにアプリケーションの管理ページへ案内することを検証する。
func TestNew_KeptOAuthApplications(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	userID := testutil.NewUserBuilder(t, tx).Build()
	testutil.CreateTestOAuthApplication(t, tx, userID, false)
	handler, _ := newTestHandler(t, db, tx)

	rr := serveNew(t, handler, newRequest(t, &model.User{ID: userID, Username: "oauth_owner"}, "ja"))

	if rr.Code != http.StatusOK {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
	}
	body := rr.Body.String()
	for _, expected := range []string{
		"作成したアプリケーションが残っているため、退会できません",
		`href="/oauth/applications"`,
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}
	for _, unexpected := range []string{
		`action="/settings/account/deletion"`,
		`id="username"`,
	} {
		if strings.Contains(body, unexpected) {
			t.Errorf("レスポンスに含まれるべきでない文字列 = %q", unexpected)
		}
	}
}

// TestNew_RequiresAuthは、未ログインのリクエストを確認ページに通さずログイン画面へ送ることを検証する。
func TestNew_RequiresAuth(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	handler, sessionManager := newTestHandler(t, db, tx)

	r := chi.NewRouter()
	r.With(authMiddleware.NewAuthMiddleware(sessionManager).RequireAuth).Get("/settings/account/deletion/new", handler.New)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/settings/account/deletion/new", nil))

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusSeeOther)
	}
	if location := rr.Header().Get("Location"); location != "/sign_in?back=%2Fsettings%2Faccount%2Fdeletion%2Fnew" {
		t.Errorf("Location = %q、期待値 = %q", location, "/sign_in?back=%2Fsettings%2Faccount%2Fdeletion%2Fnew")
	}
}
