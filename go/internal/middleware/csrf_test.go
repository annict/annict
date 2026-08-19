package middleware_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/testutil"
)

// generatePrivateSessionID はpublic IDからprivate IDを生成
// Rails/Rackの実装と互換性のある形式: SHA256(publicID)
func generatePrivateSessionID(publicID string) string {
	hash := sha256.Sum256([]byte(publicID))
	return hex.EncodeToString(hash[:])
}

// assertInvalidCSRFTokenPage asserts that a rejected form submission is served as the shared
// error page rather than as the one-line plain text http.Error used to return. The copy is the
// CSRF-specific one, not the permission wording the role middleware serves for its own 403.
//
// [Ja] assertInvalidCSRFTokenPage は、拒否されたフォーム送信が、以前 http.Error が返していた
// 1 行のプレーンテキストではなく共通のエラーページとして配信されることを検証する。文言は
// CSRF 専用のもので、ロールのミドルウェアが自身の 403 で出す権限不足の文言ではない。
func assertInvalidCSRFTokenPage(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()

	if contentType := rr.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/html; charset=utf-8", contentType)
	}

	body := rr.Body.String()
	for _, expected := range []string{
		"<title>フォームを送信できませんでした | Annict</title>",
		"ページを開き直してから、もう一度送信してください。",
		`href="/"`,
		"ホームに戻る",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("CSRF 検証失敗のレスポンスに %q が含まれていません", expected)
		}
	}
	forbiddenMessage := i18n.T(i18n.SetLocale(context.Background(), "ja"), "error_forbidden_message")
	if strings.Contains(body, forbiddenMessage) {
		t.Error("CSRF 検証失敗のレスポンスに権限不足の文言が使われています")
	}
}

func TestCSRFMiddleware_GET(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	cfg := &config.Config{
		CookieDomain:    ".example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	csrfMiddleware := middleware.NewCSRFMiddleware(sessionManager)

	// テスト用ハンドラー
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// GETリクエストはCSRFチェックをスキップ
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	csrfMiddleware.Middleware(testHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GETリクエストでCSRFチェックがスキップされませんでした: got %v want %v", rr.Code, http.StatusOK)
	}

	if rr.Body.String() != "OK" {
		t.Errorf("GETリクエストでレスポンスが正しくありません: got %v want %v", rr.Body.String(), "OK")
	}
}

func TestCSRFMiddleware_POST_ValidToken(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	cfg := &config.Config{
		CookieDomain:    ".example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	csrfMiddleware := middleware.NewCSRFMiddleware(sessionManager)

	// テスト用セッションを作成（CSRFトークンを含む）
	csrfToken := "test_csrf_token_12345"
	publicSessionID := "test_session_valid"

	// セッションをDBに直接保存（CSRFトークンを含む）
	sessionData := map[string]any{
		"_csrf_token":          csrfToken,
		"warden.user.user.key": []any{[]any{float64(1)}, "salt"},
	}
	sessionDataJSON, _ := json.Marshal(sessionData)

	// private IDを生成（Rails/Rack互換）
	// "2::" + SHA256(publicID)
	hash := "2::" + generatePrivateSessionID(publicSessionID)

	_, err := tx.Exec(`
		INSERT INTO sessions (session_id, data, created_at, updated_at)
		VALUES ($1, $2::jsonb, NOW(), NOW())
	`, hash, string(sessionDataJSON))
	if err != nil {
		t.Fatalf("セッションデータの作成に失敗しました: %v", err)
	}

	sessionID := publicSessionID

	// テスト用ハンドラー
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// POSTリクエストに正しいCSRFトークンを含める
	form := url.Values{}
	form.Set("csrf_token", csrfToken)

	req := httptest.NewRequest("POST", "/test", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{
		Name:  session.SessionKey,
		Value: sessionID,
	})

	rr := httptest.NewRecorder()

	csrfMiddleware.Middleware(testHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("正しいCSRFトークンで403エラーが返されました: got %v want %v", rr.Code, http.StatusOK)
	}

	if rr.Body.String() != "OK" {
		t.Errorf("レスポンスが正しくありません: got %v want %v", rr.Body.String(), "OK")
	}
}

func TestCSRFMiddleware_POST_InvalidToken(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	cfg := &config.Config{
		CookieDomain:    ".example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	csrfMiddleware := middleware.NewCSRFMiddleware(sessionManager)

	// テスト用セッションを作成（CSRFトークンを含む）
	csrfToken := "test_csrf_token_12345"
	publicSessionID := "test_session_invalid"

	// セッションをDBに直接保存（CSRFトークンを含む）
	sessionData := map[string]any{
		"_csrf_token":          csrfToken,
		"warden.user.user.key": []any{[]any{float64(1)}, "salt"},
	}
	sessionDataJSON, _ := json.Marshal(sessionData)

	// private IDを生成（Rails/Rack互換）
	hash := "2::" + generatePrivateSessionID(publicSessionID)

	_, err := tx.Exec(`
		INSERT INTO sessions (session_id, data, created_at, updated_at)
		VALUES ($1, $2::jsonb, NOW(), NOW())
	`, hash, string(sessionDataJSON))
	if err != nil {
		t.Fatalf("セッションデータの作成に失敗しました: %v", err)
	}

	sessionID := publicSessionID

	// テスト用ハンドラー
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// POSTリクエストに不正なCSRFトークンを含める
	form := url.Values{}
	form.Set("csrf_token", "invalid_token")

	req := httptest.NewRequest("POST", "/test", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{
		Name:  session.SessionKey,
		Value: sessionID,
	})

	rr := httptest.NewRecorder()

	csrfMiddleware.Middleware(testHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("不正なCSRFトークンで403エラーが返されませんでした: got %v want %v", rr.Code, http.StatusForbidden)
	}
	assertInvalidCSRFTokenPage(t, rr)
}

func TestCSRFMiddleware_POST_NoSession(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	cfg := &config.Config{
		CookieDomain:    ".example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	csrfMiddleware := middleware.NewCSRFMiddleware(sessionManager)

	// テスト用ハンドラー
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// POSTリクエストにセッションCookieを含めない
	form := url.Values{}
	form.Set("csrf_token", "test_token")

	req := httptest.NewRequest("POST", "/test", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	csrfMiddleware.Middleware(testHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("セッションなしで403エラーが返されませんでした: got %v want %v", rr.Code, http.StatusForbidden)
	}
	assertInvalidCSRFTokenPage(t, rr)
}

func TestCSRFMiddleware_POST_HeaderToken(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	cfg := &config.Config{
		CookieDomain:    ".example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	csrfMiddleware := middleware.NewCSRFMiddleware(sessionManager)

	// テスト用セッションを作成（CSRFトークンを含む）
	csrfToken := "test_csrf_token_header"
	publicSessionID := "test_session_header"

	// セッションをDBに直接保存（CSRFトークンを含む）
	sessionData := map[string]any{
		"_csrf_token":          csrfToken,
		"warden.user.user.key": []any{[]any{float64(1)}, "salt"},
	}
	sessionDataJSON, _ := json.Marshal(sessionData)

	// private IDを生成（Rails/Rack互換）
	hash := "2::" + generatePrivateSessionID(publicSessionID)

	_, err := tx.Exec(`
		INSERT INTO sessions (session_id, data, created_at, updated_at)
		VALUES ($1, $2::jsonb, NOW(), NOW())
	`, hash, string(sessionDataJSON))
	if err != nil {
		t.Fatalf("セッションデータの作成に失敗しました: %v", err)
	}

	sessionID := publicSessionID

	// テスト用ハンドラー
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// POSTリクエストにヘッダーでCSRFトークンを含める
	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("X-CSRF-Token", csrfToken)
	req.AddCookie(&http.Cookie{
		Name:  session.SessionKey,
		Value: sessionID,
	})

	rr := httptest.NewRecorder()

	csrfMiddleware.Middleware(testHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ヘッダーのCSRFトークンで403エラーが返されました: got %v want %v", rr.Code, http.StatusOK)
	}

	if rr.Body.String() != "OK" {
		t.Errorf("レスポンスが正しくありません: got %v want %v", rr.Body.String(), "OK")
	}
}

func TestGetCSRFToken(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	cfg := &config.Config{
		CookieDomain:    ".example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)

	// テスト用セッションを作成（CSRFトークンを含む）
	csrfToken := "test_csrf_token_get"
	publicSessionID := "test_session_get"

	// セッションをDBに直接保存（CSRFトークンを含む）
	sessionData := map[string]any{
		"_csrf_token":          csrfToken,
		"warden.user.user.key": []any{[]any{float64(1)}, "salt"},
	}
	sessionDataJSON, _ := json.Marshal(sessionData)

	// private IDを生成（Rails/Rack互換）
	hash := "2::" + generatePrivateSessionID(publicSessionID)

	_, err := tx.Exec(`
		INSERT INTO sessions (session_id, data, created_at, updated_at)
		VALUES ($1, $2::jsonb, NOW(), NOW())
	`, hash, string(sessionDataJSON))
	if err != nil {
		t.Fatalf("セッションデータの作成に失敗しました: %v", err)
	}

	sessionID := publicSessionID

	// テスト用リクエスト
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  session.SessionKey,
		Value: sessionID,
	})

	// GetCSRFToken関数をテスト
	token := middleware.GetCSRFToken(req, sessionManager)

	if token != csrfToken {
		t.Errorf("CSRFトークンが正しく取得できませんでした: got %v want %v", token, csrfToken)
	}
}

func TestGetCSRFToken_NoSession(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	cfg := &config.Config{
		CookieDomain:    ".example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)

	// テスト用リクエスト（セッションCookieなし）
	req := httptest.NewRequest("GET", "/test", nil)

	// GetCSRFToken関数をテスト
	token := middleware.GetCSRFToken(req, sessionManager)

	if token != "" {
		t.Errorf("セッションなしで空文字列が返されませんでした: got %v want %v", token, "")
	}
}

func TestGetOrCreateCSRFToken_NoSession(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	cfg := &config.Config{
		CookieDomain:    ".example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)

	// テスト用リクエスト（セッションCookieなし）
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	// GetOrCreateCSRFToken関数をテスト
	token := middleware.GetOrCreateCSRFToken(rr, req, sessionManager)

	// CSRFトークンが生成されていることを確認
	if token == "" {
		t.Error("CSRFトークンが生成されませんでした")
	}

	// トークンの長さを確認（Base64エンコード後：約44文字）
	if len(token) < 40 || len(token) > 50 {
		t.Errorf("CSRFトークンの長さが不正です: got %v", len(token))
	}

	// セッションCookieが設定されていることを確認
	cookies := rr.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == session.SessionKey {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("セッションCookieが設定されていません")
	}

	// 新しく作成されたセッションIDでCSRFトークンを取得できることを確認
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.AddCookie(sessionCookie)

	token2 := middleware.GetCSRFToken(req2, sessionManager)
	if token2 != token {
		t.Errorf("セッションから取得したCSRFトークンが一致しません: got %v want %v", token2, token)
	}

	// _csrf_initializedダミーキーが存在しないことを確認
	sessionID := sessionCookie.Value
	sessionRecord, err := sessionRepo.GetSessionByID(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("セッション取得エラー: %v", err)
	}

	var sessionData map[string]any
	if err := json.Unmarshal(sessionRecord.Data, &sessionData); err != nil {
		t.Fatalf("セッションデータのパースエラー: %v", err)
	}

	if _, exists := sessionData["_csrf_initialized"]; exists {
		t.Error("_csrf_initializedダミーキーが存在します（廃止されるべき）")
	}

	t.Logf("新規セッションが作成され、CSRFトークンが生成されました: %s（長さ: %d文字）", token, len(token))
}

func TestGetOrCreateCSRFToken_ExistingSession(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	cfg := &config.Config{
		CookieDomain:    ".example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)

	// テスト用セッションを作成（CSRFトークンを含む）
	csrfToken := "existing_csrf_token_12345678"
	publicSessionID := "test_session_existing"

	// セッションをDBに直接保存（CSRFトークンを含む）
	sessionData := map[string]any{
		"_csrf_token":          csrfToken,
		"warden.user.user.key": []any{[]any{float64(1)}, "salt"},
	}
	sessionDataJSON, _ := json.Marshal(sessionData)

	// private IDを生成（Rails/Rack互換）
	hash := "2::" + generatePrivateSessionID(publicSessionID)

	_, err := tx.Exec(`
		INSERT INTO sessions (session_id, data, created_at, updated_at)
		VALUES ($1, $2::jsonb, NOW(), NOW())
	`, hash, string(sessionDataJSON))
	if err != nil {
		t.Fatalf("セッションデータの作成に失敗しました: %v", err)
	}

	// テスト用リクエスト（既存セッションCookieあり）
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  session.SessionKey,
		Value: publicSessionID,
	})
	rr := httptest.NewRecorder()

	// GetOrCreateCSRFToken関数をテスト
	token := middleware.GetOrCreateCSRFToken(rr, req, sessionManager)

	// 既存のCSRFトークンが返されることを確認
	if token != csrfToken {
		t.Errorf("既存のCSRFトークンが返されませんでした: got %v want %v", token, csrfToken)
	}

	// 新しいセッションCookieが設定されていないことを確認（既存セッションを使用）
	cookies := rr.Result().Cookies()
	if len(cookies) > 0 {
		t.Logf("警告: 新しいCookieが設定されました（既存セッションの場合は設定されないはず）: %v", cookies)
	}

	t.Logf("既存セッションのCSRFトークンが正しく返されました: %s", token)
}

// TestCSRFIntegration_SessionCreationAndValidation は統合テスト:
// セッション作成時にCSRFトークンが保存され、CSRF検証が正常に動作することをテスト
func TestCSRFIntegration_SessionCreationAndValidation(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	cfg := &config.Config{
		CookieDomain:    ".example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	csrfMiddleware := middleware.NewCSRFMiddleware(sessionManager)

	// 1. テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("testuser").
		WithEmail("test@example.com").
		Build()

	// 2. セッションを作成（CreateSessionメソッドを使用）
	rrSession := httptest.NewRecorder()
	reqSession := httptest.NewRequest("POST", "/test", nil)
	ctx := context.Background()
	err := sessionManager.CreateSession(ctx, rrSession, reqSession, userID)
	if err != nil {
		t.Fatalf("セッション作成エラー: %v", err)
	}

	// 3. セッションCookieを取得
	cookies := rrSession.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("セッションCookieが設定されていません")
	}

	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == session.SessionKey {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatalf("セッションCookieが取得できませんでした")
	}

	// 4. セッションデータを取得してCSRFトークンが保存されていることを確認
	sessionData, err := sessionManager.GetSession(ctx, sessionCookie.Value)
	if err != nil {
		t.Fatalf("セッション取得エラー: %v", err)
	}

	if sessionData == nil {
		t.Fatalf("セッションデータが取得できませんでした")
	}

	if sessionData.CSRFToken == "" {
		t.Errorf("CSRFトークンがセッションに保存されていません")
	}

	// CSRFトークンの長さを確認（Base64エンコード後は約44文字）
	if len(sessionData.CSRFToken) < 40 {
		t.Errorf("CSRFトークンの長さが不正です: got %d, want >= 40", len(sessionData.CSRFToken))
	}

	t.Logf("CSRFトークンが正しく生成されました: %s（長さ: %d文字）", sessionData.CSRFToken, len(sessionData.CSRFToken))

	// 5. 正しいCSRFトークンを使用してPOSTリクエストを送信
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	form := url.Values{}
	form.Set("csrf_token", sessionData.CSRFToken)
	form.Set("test_field", "test_value")

	req := httptest.NewRequest("POST", "/test", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)

	rr := httptest.NewRecorder()

	csrfMiddleware.Middleware(testHandler).ServeHTTP(rr, req)

	// ステータスコードを確認（200 OKが期待される）
	if rr.Code != http.StatusOK {
		t.Errorf("正しいCSRFトークンでリクエストが失敗しました: got %d, want %d", rr.Code, http.StatusOK)
	}

	if rr.Body.String() != "OK" {
		t.Errorf("レスポンスボディが正しくありません: got %q, want %q", rr.Body.String(), "OK")
	}

	t.Logf("正しいCSRFトークンでリクエストが成功しました（ステータス: %d）", rr.Code)

	// 6. 不正なCSRFトークンを使用してPOSTリクエストを送信
	formInvalid := url.Values{}
	formInvalid.Set("csrf_token", "invalid_token_12345")
	formInvalid.Set("test_field", "test_value")

	reqInvalid := httptest.NewRequest("POST", "/test", strings.NewReader(formInvalid.Encode()))
	reqInvalid.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqInvalid.AddCookie(sessionCookie)

	rrInvalid := httptest.NewRecorder()

	csrfMiddleware.Middleware(testHandler).ServeHTTP(rrInvalid, reqInvalid)

	// ステータスコードを確認（403 Forbiddenが期待される）
	if rrInvalid.Code != http.StatusForbidden {
		t.Errorf("不正なCSRFトークンで403エラーが返されませんでした: got %d, want %d", rrInvalid.Code, http.StatusForbidden)
	}

	t.Logf("不正なCSRFトークンで403エラーが正しく返されました（ステータス: %d）", rrInvalid.Code)
}

// TestCSRFIntegration_EmptyTokenReturns403 は統合テスト:
// CSRFトークンが空の場合に403エラーが返されることをテスト
func TestCSRFIntegration_EmptyTokenReturns403(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	cfg := &config.Config{
		CookieDomain:    ".example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	csrfMiddleware := middleware.NewCSRFMiddleware(sessionManager)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("testuser").
		WithEmail("test@example.com").
		Build()

	// セッションを作成
	rrSession := httptest.NewRecorder()
	reqSession := httptest.NewRequest("POST", "/test", nil)
	err := sessionManager.CreateSession(context.Background(), rrSession, reqSession, userID)
	if err != nil {
		t.Fatalf("セッション作成エラー: %v", err)
	}

	// セッションCookieを取得
	cookies := rrSession.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == session.SessionKey {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatalf("セッションCookieが取得できませんでした")
	}

	// テストハンドラー
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// 空のCSRFトークンを使用してPOSTリクエストを送信
	form := url.Values{}
	form.Set("csrf_token", "")
	form.Set("test_field", "test_value")

	req := httptest.NewRequest("POST", "/test", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	req = req.WithContext(context.Background())

	rr := httptest.NewRecorder()

	csrfMiddleware.Middleware(testHandler).ServeHTTP(rr, req)

	// ステータスコードを確認（403 Forbiddenが期待される）
	if rr.Code != http.StatusForbidden {
		t.Errorf("空のCSRFトークンで403エラーが返されませんでした: got %d, want %d", rr.Code, http.StatusForbidden)
	}
	assertInvalidCSRFTokenPage(t, rr)

	t.Logf("空のCSRFトークンで403エラーが正しく返されました（ステータス: %d）", rr.Code)
}

// TestCSRFIntegration_NoSessionCookieReturns403 は統合テスト:
// セッションCookieがない場合に403エラーが返されることをテスト
func TestCSRFIntegration_NoSessionCookieReturns403(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	cfg := &config.Config{
		CookieDomain:    ".example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	csrfMiddleware := middleware.NewCSRFMiddleware(sessionManager)

	// テストハンドラー
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// セッションCookieなしでPOSTリクエストを送信
	form := url.Values{}
	form.Set("csrf_token", "some_token")
	form.Set("test_field", "test_value")

	req := httptest.NewRequest("POST", "/test", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.Background())

	rr := httptest.NewRecorder()

	csrfMiddleware.Middleware(testHandler).ServeHTTP(rr, req)

	// ステータスコードを確認（403 Forbiddenが期待される）
	if rr.Code != http.StatusForbidden {
		t.Errorf("セッションCookieがない場合に403エラーが返されませんでした: got %d, want %d", rr.Code, http.StatusForbidden)
	}
	assertInvalidCSRFTokenPage(t, rr)

	t.Logf("セッションCookieがない場合に403エラーが正しく返されました（ステータス: %d）", rr.Code)
}

// TestCSRFMiddleware_POST_HTMXFailureBranchesAreIndistinguishable drives all three rejection branches
// (no session ID, no session data, token mismatch) and fixes that their responses are identical.
// Answering differently would let a caller learn from the response whether a session cookie is
// still live, and the reader has the same thing to do in every case.
//
// [Ja] TestCSRFMiddleware_POST_HTMXFailureBranchesAreIndistinguishable は拒否の 3 分岐
// (セッション ID 無し / セッションデータ無し / トークン不一致) をすべて通し、応答が同一である
// ことを固定する。分岐ごとに応答を変えると、呼び出し側がセッション Cookie がまだ生きているかを
// 応答から知れてしまい、読み手がすべきことはどの場合も同じであるため。
func TestCSRFMiddleware_POST_HTMXFailureBranchesAreIndistinguishable(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	cfg := &config.Config{
		CookieDomain:    ".example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	csrfMiddleware := middleware.NewCSRFMiddleware(sessionManager)

	csrfToken := "test_csrf_token_branches"
	publicSessionID := "test_session_branches"

	sessionData := map[string]any{
		"_csrf_token":          csrfToken,
		"warden.user.user.key": []any{[]any{float64(1)}, "salt"},
	}
	sessionDataJSON, _ := json.Marshal(sessionData)

	hash := "2::" + generatePrivateSessionID(publicSessionID)

	_, err := tx.Exec(`
		INSERT INTO sessions (session_id, data, created_at, updated_at)
		VALUES ($1, $2::jsonb, NOW(), NOW())
	`, hash, string(sessionDataJSON))
	if err != nil {
		t.Fatalf("セッションデータの作成に失敗しました: %v", err)
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	newRequest := func(sessionID, formToken string) *http.Request {
		form := url.Values{}
		form.Set("csrf_token", formToken)

		req := httptest.NewRequest("POST", "/test", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("HX-Request", "true")
		if sessionID != "" {
			req.AddCookie(&http.Cookie{
				Name:  session.SessionKey,
				Value: sessionID,
			})
		}
		return req
	}

	tests := []struct {
		name string
		req  *http.Request
	}{
		{
			name: "セッションIDなし",
			req:  newRequest("", "some_token"),
		},
		{
			name: "セッションデータなし",
			req:  newRequest("test_session_branches_unknown", "some_token"),
		},
		{
			name: "トークン不一致",
			req:  newRequest(publicSessionID, "invalid_token"),
		},
	}

	var wantBody string
	for i, tt := range tests {
		rr := httptest.NewRecorder()

		csrfMiddleware.Middleware(testHandler).ServeHTTP(rr, tt.req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("%s: status = %d, want %d", tt.name, rr.Code, http.StatusForbidden)
		}
		if got := rr.Header().Get("HX-Redirect"); got != httperror.InvalidCSRFTokenPath {
			t.Errorf("%s: HX-Redirect = %q, want %q", tt.name, got, httperror.InvalidCSRFTokenPath)
		}
		assertInvalidCSRFTokenPage(t, rr)

		if i == 0 {
			wantBody = rr.Body.String()
			continue
		}
		if rr.Body.String() != wantBody {
			t.Errorf("%s: 応答が他の分岐と異なります（分岐を区別してはいけません）", tt.name)
		}
	}
}
