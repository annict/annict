package sign_in_password

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/testutil"
	"github.com/annict/annict/internal/usecase"
)

// setupSessionWithEmail はセッションにメールアドレスを設定し、セッションCookieを返します
func setupSessionWithEmail(t *testing.T, sessionMgr *session.Manager, email string) *http.Cookie {
	t.Helper()

	// ダミーリクエストとレスポンスを作成
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	ctx := req.Context()

	// セッションにメールアドレスを設定
	if err := sessionMgr.SetValue(ctx, rr, req, "sign_in_email", email); err != nil {
		t.Fatalf("セッションへのメールアドレス設定に失敗: %v", err)
	}

	// 作成されたセッションCookieを取得
	cookies := rr.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == session.SessionKey {
			return cookie
		}
	}

	t.Fatal("セッションCookieが作成されませんでした")
	return nil
}

// TestNew GET /sign_in/passwordのテスト
func TestNew(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// テストユーザーを作成
	testutil.NewUserBuilder(t, tx).
		WithUsername("signin_new_user").
		WithEmail("signin_new@example.com").
		Build()

	// 設定とセッションマネージャーを作成
	cfg := &config.Config{
		CookieDomain:  ".examle.com",
		SessionSecure: "false", // テスト環境ではfalse
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)

	// UserRepositoryとCreateSessionUsecaseを作成
	userRepo := repository.NewUserRepository(queries)
	createSessionUC := usecase.NewCreateSessionUsecase(queries)

	handler := NewHandler(cfg, userRepo, sessionMgr, createSessionUC)

	// セッションにメールアドレスを設定
	sessionCookie := setupSessionWithEmail(t, sessionMgr, "signin_new@example.com")

	// リクエストを作成
	req := httptest.NewRequest("GET", "/sign_in/password", nil)
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()

	// I18nミドルウェアを適用（テストでもlocaleを設定）
	testutil.ApplyI18nMiddleware(t, handler.New)(rr, req)

	// ステータスコードを確認
	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコードが正しくない: got %v want %v", rr.Code, http.StatusOK)
	}

	// Content-Typeを確認
	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Content-Typeが正しくない: got %v", contentType)
	}
}

// TestNew_WithBackParam backパラメータが渡されたときにhiddenフィールドに含まれることをテスト
func TestNew_WithBackParam(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// テストユーザーを作成
	testutil.NewUserBuilder(t, tx).
		WithUsername("signin_back_param_user").
		WithEmail("signin_back_param@example.com").
		Build()

	cfg := &config.Config{
		CookieDomain:  ".examle.com",
		SessionSecure: "false",
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)

	userRepo := repository.NewUserRepository(queries)
	createSessionUC := usecase.NewCreateSessionUsecase(queries)

	handler := NewHandler(cfg, userRepo, sessionMgr, createSessionUC)

	// セッションにメールアドレスを設定
	sessionCookie := setupSessionWithEmail(t, sessionMgr, "signin_back_param@example.com")

	// backパラメータ付きでリクエストを作成
	req := httptest.NewRequest("GET", "/sign_in/password?back=%2Foauth%2Fauthorize%3Fclient_id%3Dtest", nil)
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()

	testutil.ApplyI18nMiddleware(t, handler.New)(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコードが正しくない: got %v want %v", rr.Code, http.StatusOK)
	}

	// レスポンスボディにbackパラメータのhiddenフィールドが含まれていることを確認
	body := rr.Body.String()
	if !strings.Contains(body, `name="back"`) {
		t.Error("backパラメータのhiddenフィールドが見つかりません")
	}
	if !strings.Contains(body, `/oauth/authorize?client_id=test`) {
		t.Error("backパラメータの値がhiddenフィールドに含まれていません")
	}
}

// TestNew_WithoutSessionEmail セッションにメールアドレスがない場合は/sign_inにリダイレクト
func TestNew_WithoutSessionEmail(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	cfg := &config.Config{
		CookieDomain:  ".examle.com",
		SessionSecure: "false",
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)

	userRepo := repository.NewUserRepository(queries)
	createSessionUC := usecase.NewCreateSessionUsecase(queries)

	handler := NewHandler(cfg, userRepo, sessionMgr, createSessionUC)

	// セッションなしでリクエストを作成
	req := httptest.NewRequest("GET", "/sign_in/password", nil)
	rr := httptest.NewRecorder()

	handler.New(rr, req)

	// /sign_inにリダイレクトされることを確認
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ステータスコードが正しくない: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	location := rr.Header().Get("Location")
	if location != "/sign_in" {
		t.Errorf("リダイレクト先が正しくない: got %v want %v", location, "/sign_in")
	}
}
