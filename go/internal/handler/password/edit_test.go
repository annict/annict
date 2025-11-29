package password

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/password_reset"
	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/testutil"
	"github.com/annict/annict/internal/usecase"
)

func TestEdit_ValidToken(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("edit_test_user").
		WithEmail("edit_test@example.com").
		Build()

	// トランザクションをコミット
	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	// テスト終了時にクリーンアップ
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM users WHERE id = $1", userID)
		_, _ = db.Exec("DELETE FROM password_reset_tokens WHERE user_id = $1", userID)
	})

	queries := query.New(db)

	// テスト用の設定を作成
	cfg := &config.Config{
		Domain:           "example.com",
		CookieDomain:     ".example.com",
		SessionSecure:    "false",
		SessionHTTPOnly:  "true",
		DisableRateLimit: false,
		AssetVersion:     "test",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	passwordResetTokenRepo := repository.NewPasswordResetTokenRepository(queries)
	updatePasswordUseCase := usecase.NewUpdatePasswordResetUsecase(db, queries)

	handler := NewHandler(cfg, db, passwordResetTokenRepo, sessionManager, nil, updatePasswordUseCase)

	// トークンを手動で作成
	plainToken, tokenDigest, err := createTestToken()
	if err != nil {
		t.Fatalf("テスト用トークンの作成に失敗: %v", err)
	}

	// トークンをDBに保存
	_, err = db.Exec(`
		INSERT INTO password_reset_tokens (user_id, token_digest, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
	`, userID, tokenDigest, time.Now().Add(1*time.Hour))
	if err != nil {
		t.Fatalf("トークンの保存に失敗: %v", err)
	}

	// リクエストを作成
	req := httptest.NewRequest("GET", "/password/edit?token="+plainToken, nil)
	rr := httptest.NewRecorder()

	// I18nミドルウェアを適用
	testutil.ApplyI18nMiddleware(t, handler.Edit)(rr, req)

	// ステータスコードを確認
	if rr.Code != http.StatusOK {
		t.Errorf("ステータスコードが正しくありません: got=%d, want=%d", rr.Code, http.StatusOK)
	}

	// Content-Typeを確認
	if contentType := rr.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Errorf("Content-Typeが正しくありません: got=%s, want=text/html; charset=utf-8", contentType)
	}
}

func TestEdit_InvalidToken(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// トランザクションをコミット
	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	queries := query.New(db)

	// テスト用の設定を作成
	cfg := &config.Config{
		Domain:           "example.com",
		CookieDomain:     ".example.com",
		SessionSecure:    "false",
		SessionHTTPOnly:  "true",
		DisableRateLimit: false,
		AssetVersion:     "test",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	passwordResetTokenRepo := repository.NewPasswordResetTokenRepository(queries)
	updatePasswordUseCase := usecase.NewUpdatePasswordResetUsecase(db, queries)

	handler := NewHandler(cfg, db, passwordResetTokenRepo, sessionManager, nil, updatePasswordUseCase)

	// 無効なトークンでリクエスト
	req := httptest.NewRequest("GET", "/password/edit?token=invalid_token", nil)
	rr := httptest.NewRecorder()

	// I18nミドルウェアを適用
	testutil.ApplyI18nMiddleware(t, handler.Edit)(rr, req)

	// ステータスコードを確認（BadRequest）
	if rr.Code != http.StatusBadRequest {
		t.Errorf("ステータスコードが正しくありません: got=%d, want=%d", rr.Code, http.StatusBadRequest)
	}
}

func TestEdit_ExpiredToken(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("edit_expired_test_user").
		WithEmail("edit_expired_test@example.com").
		Build()

	// トランザクションをコミット
	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	// テスト終了時にクリーンアップ
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM users WHERE id = $1", userID)
		_, _ = db.Exec("DELETE FROM password_reset_tokens WHERE user_id = $1", userID)
	})

	queries := query.New(db)

	// テスト用の設定を作成
	cfg := &config.Config{
		Domain:           "example.com",
		CookieDomain:     ".example.com",
		SessionSecure:    "false",
		SessionHTTPOnly:  "true",
		DisableRateLimit: false,
		AssetVersion:     "test",
	}

	sessionRepo := repository.NewSessionRepository(queries)
	sessionManager := session.NewManager(sessionRepo, cfg)
	passwordResetTokenRepo := repository.NewPasswordResetTokenRepository(queries)
	updatePasswordUseCase := usecase.NewUpdatePasswordResetUsecase(db, queries)

	handler := NewHandler(cfg, db, passwordResetTokenRepo, sessionManager, nil, updatePasswordUseCase)

	// トークンを手動で作成
	plainToken, tokenDigest, err := createTestToken()
	if err != nil {
		t.Fatalf("テスト用トークンの作成に失敗: %v", err)
	}

	// 有効期限切れのトークンをDBに保存
	_, err = db.Exec(`
		INSERT INTO password_reset_tokens (user_id, token_digest, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
	`, userID, tokenDigest, time.Now().Add(-1*time.Hour)) // 1時間前に期限切れ
	if err != nil {
		t.Fatalf("トークンの保存に失敗: %v", err)
	}

	// リクエストを作成
	req := httptest.NewRequest("GET", "/password/edit?token="+plainToken, nil)
	rr := httptest.NewRecorder()

	// I18nミドルウェアを適用
	testutil.ApplyI18nMiddleware(t, handler.Edit)(rr, req)

	// ステータスコードを確認（BadRequest）
	if rr.Code != http.StatusBadRequest {
		t.Errorf("ステータスコードが正しくありません: got=%d, want=%d", rr.Code, http.StatusBadRequest)
	}
}

// createTestToken はテスト用のトークンを生成します
func createTestToken() (plainToken string, tokenDigest string, err error) {
	plainToken, err = password_reset.GenerateToken()
	if err != nil {
		return "", "", err
	}

	tokenDigest = password_reset.HashToken(plainToken)
	return plainToken, tokenDigest, nil
}
