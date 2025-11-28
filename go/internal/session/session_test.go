package session

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	_ "github.com/lib/pq"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/repository"
)

var (
	testDB     *sql.DB
	testDBOnce sync.Once
)

// setupTestDB テスト用のDBとトランザクションをセットアップ
func setupTestDB(t *testing.T) (*sql.DB, *sql.Tx, *repository.SessionRepository) {
	t.Helper()

	// テスト用データベース接続の初期化
	testDBOnce.Do(func() {
		// テスト用データベースの接続情報（デフォルト値を使用）
		dsn := getEnv("DATABASE_URL", "postgres://postgres@postgresql:5432/annict_test?sslmode=disable")

		// データベース接続の確立
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			panic(fmt.Sprintf("テスト用データベースへの接続に失敗しました: %v", err))
		}

		// 接続プールの設定
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)

		// 接続確認
		if err := db.Ping(); err != nil {
			panic(fmt.Sprintf("テスト用データベースへのping失敗: %v", err))
		}

		testDB = db
	})

	// トランザクション開始
	tx, err := testDB.Begin()
	if err != nil {
		t.Fatalf("トランザクション開始エラー: %v", err)
	}

	// テスト終了時にロールバック
	t.Cleanup(func() {
		_ = tx.Rollback()
	})

	// sqlcクエリを作成
	queries := query.New(testDB).WithTx(tx)

	// SessionRepositoryを作成
	sessionRepo := repository.NewSessionRepository(queries)

	return testDB, tx, sessionRepo
}

// getEnv は環境変数を取得し、存在しない場合はデフォルト値を返します
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestSetValue_GetValue セッション値の保存と取得のテスト
func TestSetValue_GetValue(t *testing.T) {
	// テストDBとトランザクションをセットアップ
	_, _, sessionRepo := setupTestDB(t)

	// テスト用のConfig
	cfg := &config.Config{
		CookieDomain:    ".test.example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	// セッションマネージャーを作成
	manager := NewManager(sessionRepo, cfg)

	ctx := context.Background()

	// テストケース
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "メールアドレスを保存",
			key:   "sign_in_email",
			value: "test@example.com",
		},
		{
			name:  "ユーザーIDを保存",
			key:   "sign_in_user_id",
			value: "12345",
		},
		{
			name:  "任意の文字列を保存",
			key:   "custom_key",
			value: "custom_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 各テストケースで新しいHTTPリクエストとレスポンスを作成
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			// 値を保存
			if err := manager.SetValue(ctx, w, req, tt.key, tt.value); err != nil {
				t.Fatalf("SetValue() error = %v", err)
			}

			// Cookieを取得して新しいリクエストに設定
			cookies := w.Result().Cookies()
			if len(cookies) == 0 {
				t.Fatal("セッションクッキーが設定されていません")
			}

			req2 := httptest.NewRequest("GET", "/test", nil)
			for _, cookie := range cookies {
				req2.AddCookie(cookie)
			}

			// 値を取得
			got, err := manager.GetValue(ctx, req2, tt.key)
			if err != nil {
				t.Fatalf("GetValue() error = %v", err)
			}

			if got != tt.value {
				t.Errorf("GetValue() = %v, want %v", got, tt.value)
			}
		})
	}
}

// TestDeleteValue セッション値の削除のテスト
func TestDeleteValue(t *testing.T) {
	// テストDBとトランザクションをセットアップ
	_, _, sessionRepo := setupTestDB(t)

	// テスト用のConfig
	cfg := &config.Config{
		CookieDomain:    ".test.example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	// セッションマネージャーを作成
	manager := NewManager(sessionRepo, cfg)

	ctx := context.Background()

	// HTTPリクエストとレスポンスを作成
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// 値を保存
	key := "test_key"
	value := "test_value"
	if err := manager.SetValue(ctx, w, req, key, value); err != nil {
		t.Fatalf("SetValue() error = %v", err)
	}

	// Cookieを取得して新しいリクエストに設定
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("セッションクッキーが設定されていません")
	}

	req2 := httptest.NewRequest("GET", "/test", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}

	// 値が存在することを確認
	got, err := manager.GetValue(ctx, req2, key)
	if err != nil {
		t.Fatalf("GetValue() error = %v", err)
	}
	if got != value {
		t.Errorf("GetValue() = %v, want %v", got, value)
	}

	// 値を削除
	if err := manager.DeleteValue(ctx, req2, key); err != nil {
		t.Fatalf("DeleteValue() error = %v", err)
	}

	// 値が削除されたことを確認
	got, err = manager.GetValue(ctx, req2, key)
	if err != nil {
		t.Fatalf("GetValue() error = %v", err)
	}
	if got != "" {
		t.Errorf("GetValue() = %v, want empty string", got)
	}
}

// TestGetValue_NonExistentKey 存在しないキーを取得するテスト
func TestGetValue_NonExistentKey(t *testing.T) {
	// テストDBとトランザクションをセットアップ
	_, _, sessionRepo := setupTestDB(t)

	// テスト用のConfig
	cfg := &config.Config{
		CookieDomain:    ".test.example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	// セッションマネージャーを作成
	manager := NewManager(sessionRepo, cfg)

	ctx := context.Background()

	// HTTPリクエストとレスポンスを作成
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// 値を保存
	if err := manager.SetValue(ctx, w, req, "existing_key", "value"); err != nil {
		t.Fatalf("SetValue() error = %v", err)
	}

	// Cookieを取得して新しいリクエストに設定
	cookies := w.Result().Cookies()
	req2 := httptest.NewRequest("GET", "/test", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}

	// 存在しないキーを取得
	got, err := manager.GetValue(ctx, req2, "nonexistent_key")
	if err != nil {
		t.Fatalf("GetValue() error = %v", err)
	}
	if got != "" {
		t.Errorf("GetValue() = %v, want empty string", got)
	}
}

// TestSetValue_MultipleValues 複数の値を保存・取得するテスト
func TestSetValue_MultipleValues(t *testing.T) {
	// テストDBとトランザクションをセットアップ
	_, _, sessionRepo := setupTestDB(t)

	// テスト用のConfig
	cfg := &config.Config{
		CookieDomain:    ".test.example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	// セッションマネージャーを作成
	manager := NewManager(sessionRepo, cfg)

	ctx := context.Background()

	// HTTPリクエストとレスポンスを作成
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// 複数の値を保存（順番を保証するためにスライスを使用）
	values := []struct {
		key   string
		value string
	}{
		{"key1", "value1"},
		{"key2", "value2"},
		{"key3", "value3"},
	}

	// 最初の値を保存
	if err := manager.SetValue(ctx, w, req, values[0].key, values[0].value); err != nil {
		t.Fatalf("SetValue() error = %v", err)
	}

	// Cookieを取得
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("セッションクッキーが設定されていません")
	}

	// 2つ目以降の値を保存（Cookieを引き継ぐ）
	for i := 1; i < len(values); i++ {
		reqNext := httptest.NewRequest("GET", "/test", nil)
		for _, cookie := range cookies {
			reqNext.AddCookie(cookie)
		}
		wNext := httptest.NewRecorder()

		if err := manager.SetValue(ctx, wNext, reqNext, values[i].key, values[i].value); err != nil {
			t.Fatalf("SetValue() error = %v", err)
		}

		// Cookieを更新（念のため、新しいCookieがあれば使用）
		newCookies := wNext.Result().Cookies()
		if len(newCookies) > 0 {
			cookies = newCookies
		}
	}

	// すべての値を取得して確認
	req2 := httptest.NewRequest("GET", "/test", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}

	for _, kv := range values {
		got, err := manager.GetValue(ctx, req2, kv.key)
		if err != nil {
			t.Fatalf("GetValue() error = %v", err)
		}
		if got != kv.value {
			t.Errorf("GetValue(%v) = %v, want %v", kv.key, got, kv.value)
		}
	}
}

// TestDeleteValue_NonExistentSession 存在しないセッションに対する削除のテスト
func TestDeleteValue_NonExistentSession(t *testing.T) {
	// テストDBとトランザクションをセットアップ
	_, _, sessionRepo := setupTestDB(t)

	// テスト用のConfig
	cfg := &config.Config{
		CookieDomain:    ".test.example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	// セッションマネージャーを作成
	manager := NewManager(sessionRepo, cfg)

	ctx := context.Background()

	// セッションがないリクエスト
	req := httptest.NewRequest("GET", "/test", nil)

	// 削除はエラーなく成功するべき
	if err := manager.DeleteValue(ctx, req, "any_key"); err != nil {
		t.Fatalf("DeleteValue() error = %v, want nil", err)
	}
}

// TestCreateSession CreateSessionメソッドのテスト
func TestCreateSession(t *testing.T) {
	// テストDBとトランザクションをセットアップ
	_, tx, sessionRepo := setupTestDB(t)

	// テスト用のConfig
	cfg := &config.Config{
		CookieDomain:    ".test.example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	// セッションマネージャーを作成
	manager := NewManager(sessionRepo, cfg)

	ctx := context.Background()

	// テスト用のユーザーを作成
	var userID int64
	err := tx.QueryRow(`
		INSERT INTO users (email, username, role, encrypted_password, time_zone, locale, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id
	`, "test@example.com", "testuser", 0, "", "UTC", "ja").Scan(&userID)
	if err != nil {
		t.Fatalf("ユーザー作成エラー: %v", err)
	}

	// ResponseRecorderとRequestを作成
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test", nil)

	// CreateSessionを実行
	if err := manager.CreateSession(ctx, w, r, userID); err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	// Cookieが設定されていることを確認
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("セッションクッキーが設定されていません")
	}

	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == SessionKey {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("セッションクッキーが見つかりません")
	}

	// セッションIDを取得
	sessionID := sessionCookie.Value
	if sessionID == "" {
		t.Fatal("セッションIDが空です")
	}

	// セッションデータをDBから取得
	session, err := sessionRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		t.Fatalf("セッション取得エラー: %v", err)
	}

	// セッションデータをパース
	var sessionData map[string]any
	if err := json.Unmarshal(session.Data, &sessionData); err != nil {
		t.Fatalf("セッションデータのパースエラー: %v", err)
	}

	// CSRFトークンが保存されていることを確認
	csrfToken, ok := sessionData["_csrf_token"].(string)
	if !ok {
		t.Fatal("CSRFトークンがセッションに保存されていません")
	}
	if csrfToken == "" {
		t.Fatal("CSRFトークンが空です")
	}

	// ユーザーIDが保存されていることを確認
	wardenKey, ok := sessionData["warden.user.user.key"]
	if !ok {
		t.Fatal("warden.user.user.keyがセッションに保存されていません")
	}

	// wardenKeyの構造を確認: [[user_id], "authenticatable_salt"]
	wardenArray, ok := wardenKey.([]any)
	if !ok || len(wardenArray) < 1 {
		t.Fatal("warden.user.user.keyの形式が不正です")
	}

	userArray, ok := wardenArray[0].([]any)
	if !ok || len(userArray) < 1 {
		t.Fatal("ユーザー配列の形式が不正です")
	}

	storedUserID, ok := userArray[0].(float64)
	if !ok {
		t.Fatal("ユーザーIDの形式が不正です")
	}

	if int64(storedUserID) != userID {
		t.Errorf("ユーザーIDが一致しません: got %d, want %d", int64(storedUserID), userID)
	}

	// Cookieの属性を確認
	// 注: http.Cookieは先頭のドットを削除することがあるため、期待値から先頭のドットを削除して比較
	expectedDomain := cfg.CookieDomain
	if len(expectedDomain) > 0 && expectedDomain[0] == '.' {
		expectedDomain = expectedDomain[1:]
	}
	if sessionCookie.Domain != expectedDomain {
		t.Errorf("Cookie domain = %s, want %s", sessionCookie.Domain, expectedDomain)
	}
	if !sessionCookie.HttpOnly {
		t.Error("CookieがHttpOnlyではありません")
	}
	if sessionCookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("Cookie SameSite = %v, want %v", sessionCookie.SameSite, http.SameSiteLaxMode)
	}
	if sessionCookie.MaxAge != 30*24*60*60 {
		t.Errorf("Cookie MaxAge = %d, want %d", sessionCookie.MaxAge, 30*24*60*60)
	}

	// テスト終了時にユーザーを削除（ロールバックで自動削除されるが明示的に記載）
	_, _ = tx.Exec("DELETE FROM users WHERE id = $1", userID)
}

// TestSetValue_NewSession_CSRFToken SetValueで新規セッション作成時にCSRFトークンが自動生成されることをテスト
func TestSetValue_NewSession_CSRFToken(t *testing.T) {
	// テストDBとトランザクションをセットアップ
	_, _, sessionRepo := setupTestDB(t)

	// テスト用のConfig
	cfg := &config.Config{
		CookieDomain:    ".test.example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	// セッションマネージャーを作成
	manager := NewManager(sessionRepo, cfg)

	ctx := context.Background()

	// ResponseRecorderとRequestを作成（セッションクッキーなし）
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// SetValueを実行（新規セッション作成）
	if err := manager.SetValue(ctx, w, req, "test_key", "test_value"); err != nil {
		t.Fatalf("SetValue() error = %v", err)
	}

	// Cookieが設定されていることを確認
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("セッションクッキーが設定されていません")
	}

	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == SessionKey {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("セッションクッキーが見つかりません")
	}

	// セッションIDを取得
	sessionID := sessionCookie.Value
	if sessionID == "" {
		t.Fatal("セッションIDが空です")
	}

	// セッションデータをDBから取得
	session, err := sessionRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		t.Fatalf("セッション取得エラー: %v", err)
	}

	// セッションデータをパース
	var sessionData map[string]any
	if err := json.Unmarshal(session.Data, &sessionData); err != nil {
		t.Fatalf("セッションデータのパースエラー: %v", err)
	}

	// CSRFトークンが自動生成されていることを確認
	csrfToken, ok := sessionData["_csrf_token"].(string)
	if !ok {
		t.Fatal("CSRFトークンがセッションに保存されていません")
	}
	if csrfToken == "" {
		t.Fatal("CSRFトークンが空です")
	}

	// 設定した値が保存されていることを確認
	testValue, ok := sessionData["test_key"].(string)
	if !ok {
		t.Fatal("test_keyがセッションに保存されていません")
	}
	if testValue != "test_value" {
		t.Errorf("test_value = %s, want test_value", testValue)
	}
}

// TestSetValue_ExistingSession_CSRFToken SetValueで既存セッション更新時にCSRFトークンが保持されることをテスト
func TestSetValue_ExistingSession_CSRFToken(t *testing.T) {
	// テストDBとトランザクションをセットアップ
	_, _, sessionRepo := setupTestDB(t)

	// テスト用のConfig
	cfg := &config.Config{
		CookieDomain:    ".test.example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	// セッションマネージャーを作成
	manager := NewManager(sessionRepo, cfg)

	ctx := context.Background()

	// ResponseRecorderとRequestを作成（セッションクッキーなし）
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/test", nil)

	// 最初のSetValueで新規セッション作成
	if err := manager.SetValue(ctx, w1, req1, "key1", "value1"); err != nil {
		t.Fatalf("最初のSetValue() error = %v", err)
	}

	// セッションIDを取得
	cookies := w1.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == SessionKey {
			sessionCookie = cookie
			break
		}
	}
	sessionID := sessionCookie.Value

	// 最初のCSRFトークンを取得
	session1, err := sessionRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		t.Fatalf("セッション取得エラー: %v", err)
	}
	var sessionData1 map[string]any
	if err := json.Unmarshal(session1.Data, &sessionData1); err != nil {
		t.Fatalf("セッションデータのパースエラー: %v", err)
	}
	originalCSRFToken, _ := sessionData1["_csrf_token"].(string)

	// 2回目のSetValueで既存セッション更新（セッションクッキーを含める）
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.AddCookie(sessionCookie)

	if err := manager.SetValue(ctx, w2, req2, "key2", "value2"); err != nil {
		t.Fatalf("2回目のSetValue() error = %v", err)
	}

	// セッションデータを再取得
	session2, err := sessionRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		t.Fatalf("セッション取得エラー: %v", err)
	}
	var sessionData2 map[string]any
	if err := json.Unmarshal(session2.Data, &sessionData2); err != nil {
		t.Fatalf("セッションデータのパースエラー: %v", err)
	}

	// CSRFトークンが保持されていることを確認
	csrfToken, ok := sessionData2["_csrf_token"].(string)
	if !ok {
		t.Fatal("CSRFトークンがセッションに保存されていません")
	}
	if csrfToken != originalCSRFToken {
		t.Errorf("CSRFトークンが変更されました: got %s, want %s", csrfToken, originalCSRFToken)
	}

	// 両方の値が保存されていることを確認
	value1, ok := sessionData2["key1"].(string)
	if !ok || value1 != "value1" {
		t.Error("key1が保持されていません")
	}
	value2, ok := sessionData2["key2"].(string)
	if !ok || value2 != "value2" {
		t.Error("key2が保存されていません")
	}
}

// TestGetSession_WithCSRFToken GetSessionメソッドでCSRFトークンを正しく取得できることをテスト
func TestGetSession_WithCSRFToken(t *testing.T) {
	// テストDBとトランザクションをセットアップ
	_, tx, sessionRepo := setupTestDB(t)

	// テスト用のConfig
	cfg := &config.Config{
		CookieDomain:    ".test.example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	// セッションマネージャーを作成
	manager := NewManager(sessionRepo, cfg)

	ctx := context.Background()

	// テスト用のユーザーを作成
	var userID int64
	err := tx.QueryRow(`
		INSERT INTO users (email, username, role, encrypted_password, time_zone, locale, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id
	`, "test@example.com", "testuser", 0, "", "UTC", "ja").Scan(&userID)
	if err != nil {
		t.Fatalf("ユーザー作成エラー: %v", err)
	}

	// セッションを作成
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test", nil)
	if err := manager.CreateSession(ctx, w, r, userID); err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	// セッションIDを取得
	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == SessionKey {
			sessionCookie = cookie
			break
		}
	}
	sessionID := sessionCookie.Value

	// GetSessionでセッションデータを取得
	sessionData, err := manager.GetSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}

	// CSRFトークンが取得できることを確認
	if sessionData.CSRFToken == "" {
		t.Fatal("CSRFトークンが空です")
	}

	// ユーザーIDが取得できることを確認
	if sessionData.UserID == nil {
		t.Fatal("ユーザーIDがnilです")
	}
	if *sessionData.UserID != userID {
		t.Errorf("ユーザーID = %d, want %d", *sessionData.UserID, userID)
	}

	// テスト終了時にユーザーを削除（ロールバックで自動削除されるが明示的に記載）
	_, _ = tx.Exec("DELETE FROM users WHERE id = $1", userID)
}

// TestSetSessionCookie_SecureAttribute はsetSessionCookieのSecure属性の判定をテストする
func TestSetSessionCookie_SecureAttribute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		sessionSecure   string
		xForwardedProto string
		expectedSecure  bool
	}{
		{
			name:            "SessionSecure=true、X-Forwarded-Protoなし",
			sessionSecure:   "true",
			xForwardedProto: "",
			expectedSecure:  true,
		},
		{
			name:            "SessionSecure=false、X-Forwarded-Protoなし",
			sessionSecure:   "false",
			xForwardedProto: "",
			expectedSecure:  false,
		},
		{
			name:            "SessionSecure=false、X-Forwarded-Proto=https",
			sessionSecure:   "false",
			xForwardedProto: "https",
			expectedSecure:  true,
		},
		{
			name:            "SessionSecure=false、X-Forwarded-Proto=http",
			sessionSecure:   "false",
			xForwardedProto: "http",
			expectedSecure:  false,
		},
		{
			name:            "SessionSecure=true、X-Forwarded-Proto=http",
			sessionSecure:   "true",
			xForwardedProto: "http",
			expectedSecure:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.Config{
				CookieDomain:    ".example.com",
				SessionSecure:   tt.sessionSecure,
				SessionHTTPOnly: "true",
			}

			manager := &Manager{
				cfg: cfg,
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			if tt.xForwardedProto != "" {
				r.Header.Set("X-Forwarded-Proto", tt.xForwardedProto)
			}

			publicID := "test-session-id-12345"
			manager.setSessionCookie(w, r, publicID)

			cookies := w.Result().Cookies()
			if len(cookies) == 0 {
				t.Fatal("Cookieが設定されていません")
			}

			var sessionCookie *http.Cookie
			for _, cookie := range cookies {
				if cookie.Name == SessionKey {
					sessionCookie = cookie
					break
				}
			}

			if sessionCookie == nil {
				t.Fatal("セッションCookieが見つかりません")
			}

			if sessionCookie.Secure != tt.expectedSecure {
				t.Errorf("Secure = %v, want %v", sessionCookie.Secure, tt.expectedSecure)
			}
		})
	}
}

// TestSetSessionCookie_CookieAttributes はsetSessionCookieのCookie属性をテストする
func TestSetSessionCookie_CookieAttributes(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		CookieDomain:    ".test.example.com",
		SessionSecure:   "true",
		SessionHTTPOnly: "true",
	}

	manager := &Manager{
		cfg: cfg,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	publicID := "test-session-public-id-67890"

	manager.setSessionCookie(w, r, publicID)

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("Cookieが設定されていません")
	}

	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == SessionKey {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("セッションCookieが見つかりません")
	}

	// Cookie名の確認
	if sessionCookie.Name != SessionKey {
		t.Errorf("Cookie名 = %v, want %v", sessionCookie.Name, SessionKey)
	}

	// 値の確認
	if sessionCookie.Value != publicID {
		t.Errorf("Cookie値 = %v, want %v", sessionCookie.Value, publicID)
	}

	// Pathの確認
	if sessionCookie.Path != "/" {
		t.Errorf("Path = %v, want /", sessionCookie.Path)
	}

	// Domainの確認（http.SetCookieは先頭のドットを除去する）
	expectedDomain := "test.example.com"
	if sessionCookie.Domain != expectedDomain {
		t.Errorf("Domain = %v, want %v", sessionCookie.Domain, expectedDomain)
	}

	// HttpOnlyの確認
	if !sessionCookie.HttpOnly {
		t.Error("HttpOnly = false, want true")
	}

	// SameSiteの確認
	if sessionCookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite = %v, want %v", sessionCookie.SameSite, http.SameSiteLaxMode)
	}

	// MaxAgeの確認（30日 = 30 * 24 * 60 * 60 = 2592000秒）
	expectedMaxAge := 30 * 24 * 60 * 60
	if sessionCookie.MaxAge != expectedMaxAge {
		t.Errorf("MaxAge = %v, want %v", sessionCookie.MaxAge, expectedMaxAge)
	}
}

// TestSetSessionCookie_HttpOnlyFalse はHttpOnly=falseの場合をテストする
func TestSetSessionCookie_HttpOnlyFalse(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		CookieDomain:    ".example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "false",
	}

	manager := &Manager{
		cfg: cfg,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	publicID := "test-session-id"

	manager.setSessionCookie(w, r, publicID)

	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == SessionKey {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("セッションCookieが見つかりません")
	}

	if sessionCookie.HttpOnly {
		t.Error("HttpOnly = true, want false")
	}
}

// TestSetSessionCookieByPublicID_SetsCookie はSetSessionCookieByPublicIDがCookieを正しく設定することをテストする
func TestSetSessionCookieByPublicID_SetsCookie(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		CookieDomain:    ".test.example.com",
		SessionSecure:   "true",
		SessionHTTPOnly: "true",
	}

	manager := &Manager{
		cfg: cfg,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	publicID := "test-public-id-from-usecase"

	manager.SetSessionCookieByPublicID(w, r, publicID)

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("Cookieが設定されていません")
	}

	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == SessionKey {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("セッションCookieが見つかりません")
	}

	// 値が正しく設定されていることを確認
	if sessionCookie.Value != publicID {
		t.Errorf("Cookie値 = %v, want %v", sessionCookie.Value, publicID)
	}

	// Cookie属性が正しく設定されていることを確認
	expectedDomain := "test.example.com"
	if sessionCookie.Domain != expectedDomain {
		t.Errorf("Domain = %v, want %v", sessionCookie.Domain, expectedDomain)
	}

	if !sessionCookie.Secure {
		t.Error("Secure = false, want true")
	}

	if !sessionCookie.HttpOnly {
		t.Error("HttpOnly = false, want true")
	}

	if sessionCookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite = %v, want %v", sessionCookie.SameSite, http.SameSiteLaxMode)
	}

	expectedMaxAge := 30 * 24 * 60 * 60
	if sessionCookie.MaxAge != expectedMaxAge {
		t.Errorf("MaxAge = %v, want %v", sessionCookie.MaxAge, expectedMaxAge)
	}
}

// TestSetSessionCookieByPublicID_WithXForwardedProto はX-Forwarded-Protoヘッダーが考慮されることをテストする
func TestSetSessionCookieByPublicID_WithXForwardedProto(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		CookieDomain:    ".example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	manager := &Manager{
		cfg: cfg,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set("X-Forwarded-Proto", "https")
	publicID := "test-public-id-https"

	manager.SetSessionCookieByPublicID(w, r, publicID)

	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == SessionKey {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("セッションCookieが見つかりません")
	}

	// X-Forwarded-Proto=httpsの場合、Secure=trueになる
	if !sessionCookie.Secure {
		t.Error("Secure = false, want true (X-Forwarded-Proto=httpsなので)")
	}
}

// TestEnsureCSRFToken_NewSession セッションがない場合に新規作成してCSRFトークンを返すテスト
func TestEnsureCSRFToken_NewSession(t *testing.T) {
	// テストDBとトランザクションをセットアップ
	_, _, sessionRepo := setupTestDB(t)

	// テスト用のConfig
	cfg := &config.Config{
		CookieDomain:    ".test.example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	// セッションマネージャーを作成
	manager := NewManager(sessionRepo, cfg)

	ctx := context.Background()

	// セッションがないリクエストを作成
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/sign_in", nil)

	// EnsureCSRFTokenを実行
	csrfToken, err := manager.EnsureCSRFToken(ctx, w, r)
	if err != nil {
		t.Fatalf("EnsureCSRFToken() error = %v", err)
	}

	// CSRFトークンが返されることを確認
	if csrfToken == "" {
		t.Fatal("CSRFトークンが空です")
	}

	// Cookieが設定されていることを確認
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("セッションCookieが設定されていません")
	}

	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == SessionKey {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("セッションCookieが見つかりません")
	}

	// セッションデータをDBから取得して確認
	sessionID := sessionCookie.Value
	session, err := sessionRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		t.Fatalf("セッション取得エラー: %v", err)
	}

	var sessionData map[string]any
	if err := json.Unmarshal(session.Data, &sessionData); err != nil {
		t.Fatalf("セッションデータのパースエラー: %v", err)
	}

	// CSRFトークンがセッションに保存されていることを確認
	storedToken, ok := sessionData["_csrf_token"].(string)
	if !ok {
		t.Fatal("CSRFトークンがセッションに保存されていません")
	}
	if storedToken != csrfToken {
		t.Errorf("保存されたCSRFトークン = %v, 返されたトークン = %v", storedToken, csrfToken)
	}

	// _csrf_initializedダミーキーが存在しないことを確認
	if _, exists := sessionData["_csrf_initialized"]; exists {
		t.Error("_csrf_initializedダミーキーが存在します（廃止されるべき）")
	}
}

// TestEnsureCSRFToken_ExistingSession 既存セッションがある場合は既存のCSRFトークンを返すテスト
func TestEnsureCSRFToken_ExistingSession(t *testing.T) {
	// テストDBとトランザクションをセットアップ
	_, tx, sessionRepo := setupTestDB(t)

	// テスト用のConfig
	cfg := &config.Config{
		CookieDomain:    ".test.example.com",
		SessionSecure:   "false",
		SessionHTTPOnly: "true",
	}

	// セッションマネージャーを作成
	manager := NewManager(sessionRepo, cfg)

	ctx := context.Background()

	// テスト用のユーザーを作成
	var userID int64
	err := tx.QueryRow(`
		INSERT INTO users (email, username, role, encrypted_password, time_zone, locale, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id
	`, "test@example.com", "testuser", 0, "", "UTC", "ja").Scan(&userID)
	if err != nil {
		t.Fatalf("ユーザー作成エラー: %v", err)
	}

	// 既存のセッションを作成
	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest("POST", "/test", nil)
	if err := manager.CreateSession(ctx, w1, r1, userID); err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	// セッションCookieを取得
	cookies := w1.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == SessionKey {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("セッションCookieが見つかりません")
	}

	// 既存セッションのCSRFトークンを取得
	sessionData, err := manager.GetSession(ctx, sessionCookie.Value)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	originalCSRFToken := sessionData.CSRFToken

	// 既存セッションCookieを含むリクエストでEnsureCSRFTokenを実行
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("GET", "/sign_in", nil)
	r2.AddCookie(sessionCookie)

	csrfToken, err := manager.EnsureCSRFToken(ctx, w2, r2)
	if err != nil {
		t.Fatalf("EnsureCSRFToken() error = %v", err)
	}

	// 既存のCSRFトークンが返されることを確認
	if csrfToken != originalCSRFToken {
		t.Errorf("CSRFトークンが変更されました: got %v, want %v", csrfToken, originalCSRFToken)
	}

	// 新しいCookieが設定されていないことを確認（既存セッションを使用）
	newCookies := w2.Result().Cookies()
	if len(newCookies) > 0 {
		for _, c := range newCookies {
			if c.Name == SessionKey {
				t.Error("既存セッションがあるのに新しいCookieが設定されました")
			}
		}
	}
}

// TestEnsureCSRFToken_CookieAttributes 新規セッション作成時のCookie属性テスト
func TestEnsureCSRFToken_CookieAttributes(t *testing.T) {
	// テストDBとトランザクションをセットアップ
	_, _, sessionRepo := setupTestDB(t)

	// テスト用のConfig
	cfg := &config.Config{
		CookieDomain:    ".test.example.com",
		SessionSecure:   "true",
		SessionHTTPOnly: "true",
	}

	// セッションマネージャーを作成
	manager := NewManager(sessionRepo, cfg)

	ctx := context.Background()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/sign_in", nil)

	_, err := manager.EnsureCSRFToken(ctx, w, r)
	if err != nil {
		t.Fatalf("EnsureCSRFToken() error = %v", err)
	}

	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == SessionKey {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("セッションCookieが見つかりません")
	}

	// Cookie属性の確認
	expectedDomain := "test.example.com"
	if sessionCookie.Domain != expectedDomain {
		t.Errorf("Domain = %v, want %v", sessionCookie.Domain, expectedDomain)
	}

	if !sessionCookie.Secure {
		t.Error("Secure = false, want true")
	}

	if !sessionCookie.HttpOnly {
		t.Error("HttpOnly = false, want true")
	}

	if sessionCookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite = %v, want %v", sessionCookie.SameSite, http.SameSiteLaxMode)
	}

	expectedMaxAge := 30 * 24 * 60 * 60
	if sessionCookie.MaxAge != expectedMaxAge {
		t.Errorf("MaxAge = %v, want %v", sessionCookie.MaxAge, expectedMaxAge)
	}
}
