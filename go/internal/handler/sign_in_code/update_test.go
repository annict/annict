package sign_in_code

import (
	"context"
	"fmt"
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
)

// TestUpdate_Success 6桁コード再送信成功のテスト
func TestUpdate_Success(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// テストユーザーを作成（encrypted_passwordが空 = パスワードなし）
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("resend_code_user").
		WithEmail("resend_code@example.com").
		WithEncryptedPassword(""). // パスワードなし
		Build()

	// Configを作成
	cfg := &config.Config{
		Env:              "test",
		Port:             "3000",
		Domain:           "localhost",
		CookieDomain:     "localhost",
		SessionSecure:    "false",
		DisableRateLimit: true, // Rate Limitを無効化
	}

	// セッションマネージャーを作成
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)
	userRepo := repository.NewUserRepository(queries)

	// ハンドラーを作成
	sendSignInCodeUC := usecase.NewSendSignInCodeUsecase(db, queries, nil)
	verifySignInCodeUC := usecase.NewVerifySignInCodeUsecase(db, queries)
	createSessionUC := usecase.NewCreateSessionUsecase(queries)
	handler := NewHandler(cfg, sessionMgr, userRepo, db, nil, sendSignInCodeUC, verifySignInCodeUC, createSessionUC)

	// リクエストを作成
	form := url.Values{}
	req := httptest.NewRequest(http.MethodPatch, "/sign_in/code", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// セッションにメールアドレスとユーザーIDを設定
	ctx := context.Background()

	// 最初の値を保存
	rr1 := httptest.NewRecorder()
	if err := sessionMgr.SetValue(ctx, rr1, req, "sign_in_email", "resend_code@example.com"); err != nil {
		t.Fatalf("セッション値の設定エラー (sign_in_email): %v", err)
	}

	// Cookieを取得
	cookies := rr1.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("セッションクッキーが設定されていません")
	}

	// 2つ目の値を保存（Cookieを引き継ぐ）
	req2 := httptest.NewRequest(http.MethodPatch, "/sign_in/code", strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}
	rr2 := httptest.NewRecorder()
	if err := sessionMgr.SetValue(ctx, rr2, req2, "sign_in_user_id", fmt.Sprintf("%d", userID)); err != nil {
		t.Fatalf("セッション値の設定エラー (sign_in_user_id): %v", err)
	}

	// 最終的なCookieを取得
	finalCookies := rr2.Result().Cookies()
	if len(finalCookies) == 0 {
		finalCookies = cookies
	}

	// 実際のリクエストにCookieを設定
	req = httptest.NewRequest(http.MethodPatch, "/sign_in/code", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range finalCookies {
		req.AddCookie(cookie)
	}

	// レスポンスレコーダーを作成
	rr := httptest.NewRecorder()

	// ハンドラーを実行
	handler.Update(rr, req)

	// ステータスコードをチェック
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ステータスコードが正しくありません: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	// リダイレクト先をチェック
	location := rr.Header().Get("Location")
	if location != "/sign_in/code" {
		t.Errorf("リダイレクト先が正しくありません: got %v want %v", location, "/sign_in/code")
	}
}

// TestUpdate_AlreadyLoggedIn すでにログイン済みの場合のテスト
func TestUpdate_AlreadyLoggedIn(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("already_logged_in_user").
		WithEmail("already_logged_in@example.com").
		Build()

	// Configを作成
	cfg := &config.Config{
		Env:              "test",
		Port:             "3000",
		Domain:           "localhost",
		CookieDomain:     "localhost",
		SessionSecure:    "false",
		DisableRateLimit: true,
	}

	// セッションマネージャーを作成
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)
	userRepo := repository.NewUserRepository(queries)

	// ハンドラーを作成
	sendSignInCodeUC := usecase.NewSendSignInCodeUsecase(db, queries, nil)
	verifySignInCodeUC := usecase.NewVerifySignInCodeUsecase(db, queries)
	createSessionUC := usecase.NewCreateSessionUsecase(queries)
	handler := NewHandler(cfg, sessionMgr, userRepo, db, nil, sendSignInCodeUC, verifySignInCodeUC, createSessionUC)

	// リクエストを作成
	form := url.Values{}
	req := httptest.NewRequest(http.MethodPatch, "/sign_in/code", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// ログインセッションを作成
	ctx := context.Background()
	user, err := queries.GetUserByID(ctx, userID)
	if err != nil {
		t.Fatalf("ユーザー取得エラー: %v", err)
	}

	sessionResult, err := createSessionUC.Execute(ctx, tx, userID, user.EncryptedPassword, "")
	if err != nil {
		t.Fatalf("セッション作成エラー: %v", err)
	}

	// Cookieを設定
	req.AddCookie(&http.Cookie{
		Name:  session.SessionKey,
		Value: sessionResult.PublicID,
		Path:  "/",
	})

	// レスポンスレコーダーを作成
	rr := httptest.NewRecorder()

	// ハンドラーを実行
	handler.Update(rr, req)

	// ステータスコードをチェック
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ステータスコードが正しくありません: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	// リダイレクト先をチェック（すでにログイン済みの場合は / にリダイレクト）
	location := rr.Header().Get("Location")
	if location != "/" {
		t.Errorf("リダイレクト先が正しくありません: got %v want %v", location, "/")
	}
}

// TestUpdate_NoEmailInSession セッションにメールアドレスがない場合のテスト
func TestUpdate_NoEmailInSession(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// Configを作成
	cfg := &config.Config{
		Env:              "test",
		Port:             "3000",
		Domain:           "localhost",
		CookieDomain:     "localhost",
		SessionSecure:    "false",
		DisableRateLimit: true,
	}

	// セッションマネージャーを作成
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)
	userRepo := repository.NewUserRepository(queries)

	// ハンドラーを作成
	sendSignInCodeUC := usecase.NewSendSignInCodeUsecase(db, queries, nil)
	verifySignInCodeUC := usecase.NewVerifySignInCodeUsecase(db, queries)
	createSessionUC := usecase.NewCreateSessionUsecase(queries)
	handler := NewHandler(cfg, sessionMgr, userRepo, db, nil, sendSignInCodeUC, verifySignInCodeUC, createSessionUC)

	// リクエストを作成
	form := url.Values{}
	req := httptest.NewRequest(http.MethodPatch, "/sign_in/code", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// レスポンスレコーダーを作成
	rr := httptest.NewRecorder()

	// ハンドラーを実行
	handler.Update(rr, req)

	// ステータスコードをチェック
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ステータスコードが正しくありません: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	// リダイレクト先をチェック（メールアドレスがない場合は /sign_in にリダイレクト）
	location := rr.Header().Get("Location")
	if location != "/sign_in" {
		t.Errorf("リダイレクト先が正しくありません: got %v want %v", location, "/sign_in")
	}
}

// TestUpdate_NoUserIDInSession セッションにユーザーIDがない場合のテスト
func TestUpdate_NoUserIDInSession(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// Configを作成
	cfg := &config.Config{
		Env:              "test",
		Port:             "3000",
		Domain:           "localhost",
		CookieDomain:     "localhost",
		SessionSecure:    "false",
		DisableRateLimit: true,
	}

	// セッションマネージャーを作成
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)
	userRepo := repository.NewUserRepository(queries)

	// ハンドラーを作成
	sendSignInCodeUC := usecase.NewSendSignInCodeUsecase(db, queries, nil)
	verifySignInCodeUC := usecase.NewVerifySignInCodeUsecase(db, queries)
	createSessionUC := usecase.NewCreateSessionUsecase(queries)
	handler := NewHandler(cfg, sessionMgr, userRepo, db, nil, sendSignInCodeUC, verifySignInCodeUC, createSessionUC)

	// リクエストを作成
	form := url.Values{}
	req := httptest.NewRequest(http.MethodPatch, "/sign_in/code", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// セッションにメールアドレスのみを設定（ユーザーIDは設定しない）
	ctx := context.Background()
	rr := httptest.NewRecorder()
	if err := sessionMgr.SetValue(ctx, rr, req, "sign_in_email", "test@example.com"); err != nil {
		t.Fatalf("セッション値の設定エラー: %v", err)
	}

	// Cookieを設定
	for _, cookie := range rr.Result().Cookies() {
		req.AddCookie(cookie)
	}

	// レスポンスレコーダーを作成
	rr = httptest.NewRecorder()

	// ハンドラーを実行
	handler.Update(rr, req)

	// ステータスコードをチェック
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ステータスコードが正しくありません: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	// リダイレクト先をチェック（ユーザーIDがない場合は /sign_in にリダイレクト）
	location := rr.Header().Get("Location")
	if location != "/sign_in" {
		t.Errorf("リダイレクト先が正しくありません: got %v want %v", location, "/sign_in")
	}
}
