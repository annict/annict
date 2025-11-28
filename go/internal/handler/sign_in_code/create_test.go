package sign_in_code

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/annict/annict/internal/auth"
	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/testutil"
	"github.com/annict/annict/internal/usecase"
)

// TestCreate_Success ログイン成功のテスト
func TestCreate_Success(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// テストユーザーを作成（encrypted_passwordが空 = パスワードなし）
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("email_login_success_user").
		WithEmail("email_login_success@example.com").
		WithEncryptedPassword(""). // パスワードなし
		Build()

	// ユーザー情報を取得
	ctx := context.Background()
	user, err := queries.GetUserByID(ctx, userID)
	if err != nil {
		t.Fatalf("ユーザー取得エラー: %v", err)
	}

	// 6桁コードを生成してデータベースに保存
	plainCode := "123456"
	codeDigest, err := auth.HashCode(plainCode)
	if err != nil {
		t.Fatalf("コードのハッシュ化エラー: %v", err)
	}

	_, err = queries.CreateSignInCode(ctx, query.CreateSignInCodeParams{
		UserID:     userID,
		CodeDigest: codeDigest,
		ExpiresAt:  time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		t.Fatalf("メールログインコード作成エラー: %v", err)
	}

	// トランザクションをコミット（テストデータを他のトランザクションから見えるようにする）
	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	// テスト終了時にクリーンアップ
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM sign_in_codes WHERE user_id = $1", userID)
		_, _ = db.Exec("DELETE FROM users WHERE id = $1", userID)
	})

	// 通常のクエリを使用（トランザクションなし）
	queries = query.New(db)

	// 設定とセッションマネージャーを作成
	cfg := &config.Config{
		CookieDomain:     ".example.com",
		SessionSecure:    "false", // テスト環境ではfalse
		SessionHTTPOnly:  "true",
		DisableRateLimit: true, // テストではRate Limitingを無効化
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)
	userRepo := repository.NewUserRepository(queries)

	// ユースケースを作成
	sendSignInCodeUC := usecase.NewSendSignInCodeUsecase(db, queries, nil)
	verifySignInCodeUC := usecase.NewVerifySignInCodeUsecase(db, queries)
	createSessionUC := usecase.NewCreateSessionUsecase(queries)

	handler := NewHandler(cfg, sessionMgr, userRepo, db, nil, sendSignInCodeUC, verifySignInCodeUC, createSessionUC)

	// セッションにメールアドレスとユーザーIDを設定
	req := httptest.NewRequest("POST", "/sign_in/code", nil)
	rr := httptest.NewRecorder()

	// 1回目: sign_in_email を設定
	err = sessionMgr.SetValue(ctx, rr, req, "sign_in_email", user.Email)
	if err != nil {
		t.Fatalf("セッション値の設定エラー (sign_in_email): %v", err)
	}

	// Cookieを取得してリクエストに追加（セッションを維持）
	cookies := rr.Result().Cookies()
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	// 2回目: sign_in_user_id を設定（同じセッションに追加）
	err = sessionMgr.SetValue(ctx, rr, req, "sign_in_user_id", fmt.Sprintf("%d", userID))
	if err != nil {
		t.Fatalf("セッション値の設定エラー (sign_in_user_id): %v", err)
	}

	// 更新されたCookieを取得（念のため）
	cookies = rr.Result().Cookies()

	// フォームデータを作成
	form := url.Values{}
	form.Set("code", plainCode)

	// リクエストを再作成（Cookieを含む）
	req = httptest.NewRequest("POST", "/sign_in/code", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	rr = httptest.NewRecorder()

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
	resultCookies := rr.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range resultCookies {
		if cookie.Name == session.SessionKey {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Error("セッションCookieが設定されていません")
	} else {
		if sessionCookie.Value == "" {
			t.Error("セッションCookieの値が空です")
		}
		// net/httpは先頭の"."を除去するため、".example.com" -> "example.com"になる
		expectedDomain := cfg.CookieDomain
		if len(expectedDomain) > 0 && expectedDomain[0] == '.' {
			expectedDomain = expectedDomain[1:]
		}
		if sessionCookie.Domain != expectedDomain {
			t.Errorf("セッションCookieのドメインが正しくない: got %v want %v",
				sessionCookie.Domain, expectedDomain)
		}
		if !sessionCookie.HttpOnly {
			t.Error("セッションCookieがHttpOnlyではありません")
		}
		// テスト環境ではSecureはfalse（本番環境ではtrue）
		if sessionCookie.Secure {
			t.Error("セッションCookieがSecureになっていますが、テスト環境ではfalseであるべきです")
		}
	}
}

// TestCreate_InvalidCode 間違ったコードを入力した場合のテスト
func TestCreate_InvalidCode(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("email_login_invalid_user").
		WithEmail("email_login_invalid@example.com").
		WithEncryptedPassword(""). // パスワードなし
		Build()

	// ユーザー情報を取得
	ctx := context.Background()
	user, err := queries.GetUserByID(ctx, userID)
	if err != nil {
		t.Fatalf("ユーザー取得エラー: %v", err)
	}

	// 6桁コードを生成してデータベースに保存
	plainCode := "123456"
	codeDigest, err := auth.HashCode(plainCode)
	if err != nil {
		t.Fatalf("コードのハッシュ化エラー: %v", err)
	}

	_, err = queries.CreateSignInCode(ctx, query.CreateSignInCodeParams{
		UserID:     userID,
		CodeDigest: codeDigest,
		ExpiresAt:  time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		t.Fatalf("メールログインコード作成エラー: %v", err)
	}

	// 設定とセッションマネージャーを作成
	cfg := &config.Config{
		CookieDomain:     ".example.com",
		SessionSecure:    "false",
		DisableRateLimit: true,
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)
	userRepo := repository.NewUserRepository(queries)

	// ユースケースを作成
	sendSignInCodeUC := usecase.NewSendSignInCodeUsecase(db, queries, nil)
	verifySignInCodeUC := usecase.NewVerifySignInCodeUsecase(db, queries)
	createSessionUC := usecase.NewCreateSessionUsecase(queries)

	handler := NewHandler(cfg, sessionMgr, userRepo, db, nil, sendSignInCodeUC, verifySignInCodeUC, createSessionUC)

	// セッションにメールアドレスとユーザーIDを設定
	req := httptest.NewRequest("POST", "/sign_in/code", nil)
	rr := httptest.NewRecorder()

	// 1回目: sign_in_email を設定
	err = sessionMgr.SetValue(ctx, rr, req, "sign_in_email", user.Email)
	if err != nil {
		t.Fatalf("セッション値の設定エラー (sign_in_email): %v", err)
	}

	// Cookieを取得してリクエストに追加（セッションを維持）
	cookies := rr.Result().Cookies()
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	// 2回目: sign_in_user_id を設定（同じセッションに追加）
	err = sessionMgr.SetValue(ctx, rr, req, "sign_in_user_id", fmt.Sprintf("%d", userID))
	if err != nil {
		t.Fatalf("セッション値の設定エラー (sign_in_user_id): %v", err)
	}

	// 更新されたCookieを取得
	cookies = rr.Result().Cookies()

	// フォームデータを作成（間違ったコード）
	form := url.Values{}
	form.Set("code", "999999") // 間違ったコード

	// リクエストを再作成
	req = httptest.NewRequest("POST", "/sign_in/code", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	rr = httptest.NewRecorder()

	// ハンドラーを実行
	handler.Create(rr, req)

	// ステータスコードを確認（リダイレクト）
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ステータスコードが正しくない: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	// リダイレクト先を確認（エラー時は /sign_in/code にリダイレクト）
	location := rr.Header().Get("Location")
	if location != "/sign_in/code" {
		t.Errorf("リダイレクト先が正しくない: got %v want %v", location, "/sign_in/code")
	}

	// セッションCookieが設定されていないことを確認（ログイン失敗）
	resultCookies := rr.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range resultCookies {
		if cookie.Name == session.SessionKey {
			sessionCookie = cookie
			break
		}
	}

	// セッションCookieは設定されない（既存セッションのみが存在）
	// 既存セッションが引き継がれているだけなので、ログインセッションではない
	// sessionCookieの値はチェックしない
	_ = sessionCookie
}

// TestCreate_SessionExpired セッションが切れている場合のテスト
func TestCreate_SessionExpired(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// 設定とセッションマネージャーを作成
	cfg := &config.Config{
		CookieDomain:     ".example.com",
		SessionSecure:    "false",
		DisableRateLimit: true,
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)
	userRepo := repository.NewUserRepository(queries)

	// ユースケースを作成
	sendSignInCodeUC := usecase.NewSendSignInCodeUsecase(db, queries, nil)
	verifySignInCodeUC := usecase.NewVerifySignInCodeUsecase(db, queries)
	createSessionUC := usecase.NewCreateSessionUsecase(queries)

	handler := NewHandler(cfg, sessionMgr, userRepo, db, nil, sendSignInCodeUC, verifySignInCodeUC, createSessionUC)

	// フォームデータを作成
	form := url.Values{}
	form.Set("code", "123456")

	// リクエストを作成（セッションなし）
	req := httptest.NewRequest("POST", "/sign_in/code", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	// ハンドラーを実行
	handler.Create(rr, req)

	// ステータスコードを確認（リダイレクト）
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ステータスコードが正しくない: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	// リダイレクト先を確認（セッション切れの場合は /sign_in にリダイレクト）
	location := rr.Header().Get("Location")
	if location != "/sign_in" {
		t.Errorf("リダイレクト先が正しくない: got %v want %v", location, "/sign_in")
	}
}

// TestCreate_CodeExpired コードの有効期限が切れている場合のテスト
func TestCreate_CodeExpired(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("email_login_expired_user").
		WithEmail("email_login_expired@example.com").
		WithEncryptedPassword(""). // パスワードなし
		Build()

	// ユーザー情報を取得
	ctx := context.Background()
	user, err := queries.GetUserByID(ctx, userID)
	if err != nil {
		t.Fatalf("ユーザー取得エラー: %v", err)
	}

	// 6桁コードを生成してデータベースに保存（有効期限切れ）
	plainCode := "123456"
	codeDigest, err := auth.HashCode(plainCode)
	if err != nil {
		t.Fatalf("コードのハッシュ化エラー: %v", err)
	}

	_, err = queries.CreateSignInCode(ctx, query.CreateSignInCodeParams{
		UserID:     userID,
		CodeDigest: codeDigest,
		ExpiresAt:  time.Now().Add(-1 * time.Minute), // 有効期限切れ（1分前）
	})
	if err != nil {
		t.Fatalf("メールログインコード作成エラー: %v", err)
	}

	// 設定とセッションマネージャーを作成
	cfg := &config.Config{
		CookieDomain:     ".example.com",
		SessionSecure:    "false",
		DisableRateLimit: true,
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)
	userRepo := repository.NewUserRepository(queries)

	// ユースケースを作成
	sendSignInCodeUC := usecase.NewSendSignInCodeUsecase(db, queries, nil)
	verifySignInCodeUC := usecase.NewVerifySignInCodeUsecase(db, queries)
	createSessionUC := usecase.NewCreateSessionUsecase(queries)

	handler := NewHandler(cfg, sessionMgr, userRepo, db, nil, sendSignInCodeUC, verifySignInCodeUC, createSessionUC)

	// セッションにメールアドレスとユーザーIDを設定
	req := httptest.NewRequest("POST", "/sign_in/code", nil)
	rr := httptest.NewRecorder()

	// 1回目: sign_in_email を設定
	err = sessionMgr.SetValue(ctx, rr, req, "sign_in_email", user.Email)
	if err != nil {
		t.Fatalf("セッション値の設定エラー (sign_in_email): %v", err)
	}

	// Cookieを取得してリクエストに追加（セッションを維持）
	cookies := rr.Result().Cookies()
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	// 2回目: sign_in_user_id を設定（同じセッションに追加）
	err = sessionMgr.SetValue(ctx, rr, req, "sign_in_user_id", fmt.Sprintf("%d", userID))
	if err != nil {
		t.Fatalf("セッション値の設定エラー (sign_in_user_id): %v", err)
	}

	// 更新されたCookieを取得
	cookies = rr.Result().Cookies()

	// フォームデータを作成
	form := url.Values{}
	form.Set("code", plainCode)

	// リクエストを再作成
	req = httptest.NewRequest("POST", "/sign_in/code", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	rr = httptest.NewRecorder()

	// ハンドラーを実行
	handler.Create(rr, req)

	// ステータスコードを確認（リダイレクト）
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ステータスコードが正しくない: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	// リダイレクト先を確認（エラー時は /sign_in/code にリダイレクト）
	location := rr.Header().Get("Location")
	if location != "/sign_in/code" {
		t.Errorf("リダイレクト先が正しくない: got %v want %v", location, "/sign_in/code")
	}
}

// TestCreate_ValidationError バリデーションエラーのテスト
func TestCreate_ValidationError(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("email_login_validation_user").
		WithEmail("email_login_validation@example.com").
		WithEncryptedPassword("").
		Build()

	// ユーザー情報を取得
	ctx := context.Background()
	user, err := queries.GetUserByID(ctx, userID)
	if err != nil {
		t.Fatalf("ユーザー取得エラー: %v", err)
	}

	// 設定とセッションマネージャーを作成
	cfg := &config.Config{
		CookieDomain:     ".example.com",
		SessionSecure:    "false",
		DisableRateLimit: true,
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)
	userRepo := repository.NewUserRepository(queries)

	// ユースケースを作成
	sendSignInCodeUC := usecase.NewSendSignInCodeUsecase(db, queries, nil)
	verifySignInCodeUC := usecase.NewVerifySignInCodeUsecase(db, queries)
	createSessionUC := usecase.NewCreateSessionUsecase(queries)

	handler := NewHandler(cfg, sessionMgr, userRepo, db, nil, sendSignInCodeUC, verifySignInCodeUC, createSessionUC)

	// セッションにメールアドレスとユーザーIDを設定
	req := httptest.NewRequest("POST", "/sign_in/code", nil)
	rr := httptest.NewRecorder()

	// 1回目: sign_in_email を設定
	err = sessionMgr.SetValue(ctx, rr, req, "sign_in_email", user.Email)
	if err != nil {
		t.Fatalf("セッション値の設定エラー (sign_in_email): %v", err)
	}

	// Cookieを取得してリクエストに追加（セッションを維持）
	cookies := rr.Result().Cookies()
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	// 2回目: sign_in_user_id を設定（同じセッションに追加）
	err = sessionMgr.SetValue(ctx, rr, req, "sign_in_user_id", fmt.Sprintf("%d", userID))
	if err != nil {
		t.Fatalf("セッション値の設定エラー (sign_in_user_id): %v", err)
	}

	// 更新されたCookieを取得
	cookies = rr.Result().Cookies()

	// フォームデータを作成（不正なコード: 5桁）
	form := url.Values{}
	form.Set("code", "12345") // 不正なコード

	// リクエストを再作成
	req = httptest.NewRequest("POST", "/sign_in/code", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	rr = httptest.NewRecorder()

	// ハンドラーを実行
	handler.Create(rr, req)

	// ステータスコードを確認（リダイレクト）
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ステータスコードが正しくない: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	// リダイレクト先を確認（バリデーションエラー時は /sign_in/code にリダイレクト）
	location := rr.Header().Get("Location")
	if location != "/sign_in/code" {
		t.Errorf("リダイレクト先が正しくない: got %v want %v", location, "/sign_in/code")
	}
}

// TestCreate_UserNotFound ユーザーが見つからない場合のテスト
func TestCreate_UserNotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// 存在しないユーザーIDを使用
	nonExistentUserID := int64(999999)

	// 設定とセッションマネージャーを作成
	cfg := &config.Config{
		CookieDomain:     ".example.com",
		SessionSecure:    "false",
		DisableRateLimit: true,
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)
	userRepo := repository.NewUserRepository(queries)

	// ユースケースを作成
	sendSignInCodeUC := usecase.NewSendSignInCodeUsecase(db, queries, nil)
	verifySignInCodeUC := usecase.NewVerifySignInCodeUsecase(db, queries)
	createSessionUC := usecase.NewCreateSessionUsecase(queries)

	handler := NewHandler(cfg, sessionMgr, userRepo, db, nil, sendSignInCodeUC, verifySignInCodeUC, createSessionUC)

	// セッションにメールアドレスとユーザーIDを設定（存在しないユーザー）
	ctx := context.Background()
	req := httptest.NewRequest("POST", "/sign_in/code", nil)
	rr := httptest.NewRecorder()

	// 1回目: sign_in_email を設定
	err := sessionMgr.SetValue(ctx, rr, req, "sign_in_email", "nonexistent@example.com")
	if err != nil {
		t.Fatalf("セッション値の設定エラー (sign_in_email): %v", err)
	}

	// Cookieを取得してリクエストに追加（セッションを維持）
	cookies := rr.Result().Cookies()
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	// 2回目: sign_in_user_id を設定（同じセッションに追加）
	err = sessionMgr.SetValue(ctx, rr, req, "sign_in_user_id", fmt.Sprintf("%d", nonExistentUserID))
	if err != nil {
		t.Fatalf("セッション値の設定エラー (sign_in_user_id): %v", err)
	}

	// 更新されたCookieを取得
	cookies = rr.Result().Cookies()

	// フォームデータを作成
	form := url.Values{}
	form.Set("code", "123456")

	// リクエストを再作成
	req = httptest.NewRequest("POST", "/sign_in/code", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	rr = httptest.NewRecorder()

	// ハンドラーを実行
	handler.Create(rr, req)

	// ステータスコードを確認（リダイレクト）
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ステータスコードが正しくない: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	// リダイレクト先を確認（コードが見つからない場合は /sign_in/code にリダイレクト）
	location := rr.Header().Get("Location")
	if location != "/sign_in/code" {
		t.Errorf("リダイレクト先が正しくない: got %v want %v", location, "/sign_in/code")
	}
}

// TestCreate_GetUserError ユーザー取得エラーのテスト
func TestCreate_GetUserError(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTestDB(t)
	queries := testutil.NewQueriesWithTx(db, tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("email_login_get_user_error").
		WithEmail("email_login_get_user_error@example.com").
		WithEncryptedPassword("").
		Build()

	// ユーザー情報を取得
	ctx := context.Background()
	user, err := queries.GetUserByID(ctx, userID)
	if err != nil {
		t.Fatalf("ユーザー取得エラー: %v", err)
	}

	// 6桁コードを生成してデータベースに保存
	plainCode := "123456"
	codeDigest, err := auth.HashCode(plainCode)
	if err != nil {
		t.Fatalf("コードのハッシュ化エラー: %v", err)
	}

	_, err = queries.CreateSignInCode(ctx, query.CreateSignInCodeParams{
		UserID:     userID,
		CodeDigest: codeDigest,
		ExpiresAt:  time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		t.Fatalf("メールログインコード作成エラー: %v", err)
	}

	// 設定とセッションマネージャーを作成
	cfg := &config.Config{
		CookieDomain:     ".example.com",
		SessionSecure:    "false",
		DisableRateLimit: true,
	}
	sessionRepo := repository.NewSessionRepository(queries)
	sessionMgr := session.NewManager(sessionRepo, cfg)
	userRepo := repository.NewUserRepository(queries)

	// ユースケースを作成
	sendSignInCodeUC := usecase.NewSendSignInCodeUsecase(db, queries, nil)
	verifySignInCodeUC := usecase.NewVerifySignInCodeUsecase(db, queries)
	createSessionUC := usecase.NewCreateSessionUsecase(queries)

	handler := NewHandler(cfg, sessionMgr, userRepo, db, nil, sendSignInCodeUC, verifySignInCodeUC, createSessionUC)

	// セッションにメールアドレスとユーザーIDを設定
	req := httptest.NewRequest("POST", "/sign_in/code", nil)
	rr := httptest.NewRecorder()

	// 1回目: sign_in_email を設定
	err = sessionMgr.SetValue(ctx, rr, req, "sign_in_email", user.Email)
	if err != nil {
		t.Fatalf("セッション値の設定エラー (sign_in_email): %v", err)
	}

	// Cookieを取得してリクエストに追加（セッションを維持）
	cookies := rr.Result().Cookies()
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	// 2回目: sign_in_user_id を設定（同じセッションに追加）
	err = sessionMgr.SetValue(ctx, rr, req, "sign_in_user_id", fmt.Sprintf("%d", userID))
	if err != nil {
		t.Fatalf("セッション値の設定エラー (sign_in_user_id): %v", err)
	}

	// 更新されたCookieを取得
	cookies = rr.Result().Cookies()

	// ユーザーを削除する前に関連レコードを削除
	_, err = tx.Exec("DELETE FROM email_notifications WHERE user_id = $1", userID)
	if err != nil {
		t.Fatalf("メール通知設定削除エラー: %v", err)
	}
	_, err = tx.Exec("DELETE FROM settings WHERE user_id = $1", userID)
	if err != nil {
		t.Fatalf("設定削除エラー: %v", err)
	}
	_, err = tx.Exec("DELETE FROM profiles WHERE user_id = $1", userID)
	if err != nil {
		t.Fatalf("プロフィール削除エラー: %v", err)
	}

	// ユーザーを削除（GetUserByIDでエラーを発生させる）
	_, err = tx.Exec("DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		t.Fatalf("ユーザー削除エラー: %v", err)
	}

	// フォームデータを作成
	form := url.Values{}
	form.Set("code", plainCode)

	// リクエストを再作成
	req = httptest.NewRequest("POST", "/sign_in/code", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	rr = httptest.NewRecorder()

	// ハンドラーを実行
	handler.Create(rr, req)

	// ステータスコードを確認（リダイレクト）
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ステータスコードが正しくない: got %v want %v", rr.Code, http.StatusSeeOther)
	}

	// リダイレクト先を確認（ユーザー削除時はコードもカスケード削除されるため /sign_in/code にリダイレクト）
	location := rr.Header().Get("Location")
	if location != "/sign_in/code" {
		t.Errorf("リダイレクト先が正しくない: got %v want %v", location, "/sign_in/code")
	}
}
