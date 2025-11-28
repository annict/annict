package sign_in_password

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/testutil"
	"github.com/annict/annict/internal/usecase"
	"golang.org/x/crypto/bcrypt"
)

// TestCreate_Success ログイン成功のテスト
func TestCreate_Success(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// テストユーザーを作成（bcryptでハッシュ化したパスワード）
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("パスワードハッシュ生成エラー: %v", err)
	}

	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("signin_success_user").
		WithEmail("signin_success@example.com").
		WithEncryptedPassword(string(hashedPassword)).
		Build()

	// 設定とセッションマネージャーを作成
	cfg := &config.Config{
		CookieDomain:    ".examle.com",
		SessionSecure:   "false", // テスト環境ではfalse
		SessionHTTPOnly: "true",
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)

	// UserRepositoryとCreateSessionUsecaseを作成
	userRepo := repository.NewUserRepository(queries)
	createSessionUC := usecase.NewCreateSessionUsecase(queries)

	handler := NewHandler(cfg, userRepo, sessionMgr, createSessionUC)

	// セッションにメールアドレスを設定
	sessionCookie := setupSessionWithEmail(t, sessionMgr, "signin_success@example.com")

	// フォームデータを作成（email_usernameフィールドはもう必要なし）
	form := url.Values{}
	form.Set("password", "password123")

	// リクエストを作成
	req := httptest.NewRequest("POST", "/sign_in/password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()

	// ハンドラーを実行
	handler.Create(rr, req)

	// ステータスコードを確認（リダイレクト）
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ステータスコードが正しくない: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	// リダイレクト先を確認
	location := rr.Header().Get("Location")
	if location != "/" {
		t.Errorf("リダイレクト先が正しくない: got %v want %v", location, "/")
	}

	// セッションCookieが設定されているか確認
	cookies := rr.Result().Cookies()
	var newSessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == session.SessionKey {
			newSessionCookie = cookie
			break
		}
	}

	if newSessionCookie == nil {
		t.Error("セッションCookieが設定されていません")
	} else {
		if newSessionCookie.Value == "" {
			t.Error("セッションCookieの値が空です")
		}
		// net/httpは先頭の"."を除去するため、".examle.com" -> "examle.com"になる
		expectedDomain := cfg.CookieDomain
		if len(expectedDomain) > 0 && expectedDomain[0] == '.' {
			expectedDomain = expectedDomain[1:]
		}
		if newSessionCookie.Domain != expectedDomain {
			t.Errorf("セッションCookieのドメインが正しくない: got %v want %v",
				newSessionCookie.Domain, expectedDomain)
		}
		if !newSessionCookie.HttpOnly {
			t.Error("セッションCookieがHttpOnlyではありません")
		}
		// テスト環境ではSecureはfalse（本番環境ではtrue）
		if newSessionCookie.Secure {
			t.Error("セッションCookieがSecureになっていますが、テスト環境ではfalseであるべきです")
		}
	}

	_ = userID // 未使用変数の警告を回避
}

// TestCreate_InvalidCredentials 認証失敗のテスト
func TestCreate_InvalidCredentials(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// テストユーザーを作成
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("パスワードハッシュ生成エラー: %v", err)
	}

	testutil.NewUserBuilder(t, tx).
		WithUsername("signin_invalid_user").
		WithEmail("signin_invalid@example.com").
		WithEncryptedPassword(string(hashedPassword)).
		Build()

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
	sessionCookie := setupSessionWithEmail(t, sessionMgr, "signin_invalid@example.com")

	// 間違ったパスワードでログイン試行
	form := url.Values{}
	form.Set("password", "wrongpassword")

	req := httptest.NewRequest("POST", "/sign_in/password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	// ステータスコードを確認（リダイレクト）
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ステータスコードが正しくない: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	// /sign_in/passwordにリダイレクトされているか確認（Flashメッセージ使用）
	location := rr.Header().Get("Location")
	if location != "/sign_in/password" {
		t.Errorf("リダイレクト先が正しくない: got %v want %v", location, "/sign_in/password")
	}
}

// TestCreate_WithBackParam ログイン成功時にbackパラメータでリダイレクトされることをテスト
func TestCreate_WithBackParam(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("パスワードハッシュ生成エラー: %v", err)
	}

	testutil.NewUserBuilder(t, tx).
		WithUsername("signin_back_user").
		WithEmail("signin_back@example.com").
		WithEncryptedPassword(string(hashedPassword)).
		Build()

	cfg := &config.Config{
		CookieDomain:    ".examle.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)

	userRepo := repository.NewUserRepository(queries)
	createSessionUC := usecase.NewCreateSessionUsecase(queries)

	handler := NewHandler(cfg, userRepo, sessionMgr, createSessionUC)

	// セッションにメールアドレスを設定
	sessionCookie := setupSessionWithEmail(t, sessionMgr, "signin_back@example.com")

	// backパラメータ付きでログイン
	form := url.Values{}
	form.Set("password", "password123")
	form.Set("back", "/oauth/authorize?client_id=test")

	req := httptest.NewRequest("POST", "/sign_in/password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("ステータスコードが正しくない: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	// backパラメータで指定したURLにリダイレクトされることを確認
	location := rr.Header().Get("Location")
	if location != "/oauth/authorize?client_id=test" {
		t.Errorf("リダイレクト先が正しくない: got %v want %v", location, "/oauth/authorize?client_id=test")
	}
}

// TestCreate_WithInvalidBackParam 無効なbackパラメータの場合はデフォルトリダイレクトになることをテスト
func TestCreate_WithInvalidBackParam(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("パスワードハッシュ生成エラー: %v", err)
	}

	testutil.NewUserBuilder(t, tx).
		WithUsername("signin_invalid_back_user").
		WithEmail("signin_invalid_back@example.com").
		WithEncryptedPassword(string(hashedPassword)).
		Build()

	cfg := &config.Config{
		CookieDomain:    ".examle.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)

	userRepo := repository.NewUserRepository(queries)
	createSessionUC := usecase.NewCreateSessionUsecase(queries)

	handler := NewHandler(cfg, userRepo, sessionMgr, createSessionUC)

	// セッションにメールアドレスを設定
	sessionCookie := setupSessionWithEmail(t, sessionMgr, "signin_invalid_back@example.com")

	// 無効なbackパラメータ（絶対URL）でログイン
	form := url.Values{}
	form.Set("password", "password123")
	form.Set("back", "https://evil.com/phishing")

	req := httptest.NewRequest("POST", "/sign_in/password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("ステータスコードが正しくない: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	// 無効なbackパラメータの場合はデフォルト（"/"）にリダイレクトされることを確認
	location := rr.Header().Get("Location")
	if location != "/" {
		t.Errorf("リダイレクト先が正しくない: got %v want %v（無効なbackパラメータはデフォルトにリダイレクトされるべき）", location, "/")
	}
}

// TestCreate_WithProtocolRelativeBackParam プロトコル相対URLのbackパラメータは無効になることをテスト
func TestCreate_WithProtocolRelativeBackParam(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("パスワードハッシュ生成エラー: %v", err)
	}

	testutil.NewUserBuilder(t, tx).
		WithUsername("signin_proto_rel_user").
		WithEmail("signin_proto_rel@example.com").
		WithEncryptedPassword(string(hashedPassword)).
		Build()

	cfg := &config.Config{
		CookieDomain:    ".examle.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)

	userRepo := repository.NewUserRepository(queries)
	createSessionUC := usecase.NewCreateSessionUsecase(queries)

	handler := NewHandler(cfg, userRepo, sessionMgr, createSessionUC)

	// セッションにメールアドレスを設定
	sessionCookie := setupSessionWithEmail(t, sessionMgr, "signin_proto_rel@example.com")

	// プロトコル相対URLのbackパラメータでログイン
	form := url.Values{}
	form.Set("password", "password123")
	form.Set("back", "//evil.com/phishing")

	req := httptest.NewRequest("POST", "/sign_in/password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("ステータスコードが正しくない: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	// プロトコル相対URLの場合もデフォルト（"/"）にリダイレクトされることを確認
	location := rr.Header().Get("Location")
	if location != "/" {
		t.Errorf("リダイレクト先が正しくない: got %v want %v（プロトコル相対URLはデフォルトにリダイレクトされるべき）", location, "/")
	}
}

// TestCreate_GlobalError グローバルエラー時のform_errorsパーシャルが正しく読み込まれているか確認するテスト
// このテストは認証失敗時のグローバルエラーメッセージが正しく表示されることを確認する
func TestCreate_GlobalError(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// テストユーザーを作成（パスワードが間違っている場合のテスト）
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("パスワードハッシュ生成エラー: %v", err)
	}

	testutil.NewUserBuilder(t, tx).
		WithUsername("signin_global_error_user").
		WithEmail("signin_global_error@example.com").
		WithEncryptedPassword(string(hashedPassword)).
		Build()

	cfg := &config.Config{
		CookieDomain:  ".examle.com",
		SessionSecure: "false",
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)

	// UserRepositoryとCreateSessionUsecaseを作成
	userRepo := repository.NewUserRepository(queries)
	createSessionUC := usecase.NewCreateSessionUsecase(queries)

	handler := NewHandler(cfg, userRepo, sessionMgr, createSessionUC)

	// セッションにメールアドレスを設定
	sessionCookie := setupSessionWithEmail(t, sessionMgr, "signin_global_error@example.com")

	// Step 1: 間違ったパスワードでPOST（グローバルエラーを発生させる）
	form := url.Values{}
	form.Set("password", "wrongpassword")

	reqPost := httptest.NewRequest("POST", "/sign_in/password", strings.NewReader(form.Encode()))
	reqPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqPost.AddCookie(sessionCookie)
	rrPost := httptest.NewRecorder()

	// I18nミドルウェアを適用
	testutil.ApplyI18nMiddleware(t, handler.Create)(rrPost, reqPost)

	// ステータスコードを確認（認証失敗の場合はリダイレクト）
	if rrPost.Code != http.StatusSeeOther {
		t.Errorf("ステータスコードが正しくない: got %v want %v", rrPost.Code, http.StatusSeeOther)
	}

	// リダイレクト先を確認
	location := rrPost.Header().Get("Location")
	if location != "/sign_in/password" {
		t.Errorf("リダイレクト先が正しくない: got %v want %v", location, "/sign_in/password")
	}

	// Step 2: リダイレクト後のGETリクエストでグローバルエラーが表示されることを確認
	// 既存のセッションが更新されるだけなので、元のセッションCookieを使用
	reqGet := httptest.NewRequest("GET", "/sign_in/password", nil)
	reqGet.AddCookie(sessionCookie)
	rrGet := httptest.NewRecorder()

	// I18nミドルウェアを適用
	testutil.ApplyI18nMiddleware(t, handler.New)(rrGet, reqGet)

	// ステータスコードを確認
	if rrGet.Code != http.StatusOK {
		t.Errorf("ステータスコードが正しくない: got %v want %v", rrGet.Code, http.StatusOK)
	}

	// レスポンスボディを確認（form_errorsパーシャルが正しくレンダリングされているか）
	body := rrGet.Body.String()

	// form_errorsパーシャルが正しく機能していることを確認
	// （グローバルエラーメッセージのスタイルクラスが含まれている）
	if !strings.Contains(body, "alert-destructive") {
		t.Error("エラーメッセージのスタイルクラスが見つかりません（form_errorsパーシャルが読み込まれていない可能性があります）")
	}

	// Content-Typeを確認
	contentType := rrGet.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Content-Typeが正しくない: got %v", contentType)
	}
}

// TestCreate_WithoutSessionEmail セッションにメールアドレスがない場合は/sign_inにリダイレクト
func TestCreate_WithoutSessionEmail(t *testing.T) {
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
	form := url.Values{}
	form.Set("password", "password123")

	req := httptest.NewRequest("POST", "/sign_in/password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	// /sign_inにリダイレクトされることを確認
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ステータスコードが正しくない: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	location := rr.Header().Get("Location")
	if location != "/sign_in" {
		t.Errorf("リダイレクト先が正しくない: got %v want %v", location, "/sign_in")
	}
}
