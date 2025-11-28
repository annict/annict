package password

import (
	"database/sql"
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

func TestUpdate_Success(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// テストユーザーを作成
	oldPassword := "OldPassword123!"
	encryptedPassword, err := auth.HashPassword(oldPassword)
	if err != nil {
		t.Fatalf("パスワードのハッシュ化に失敗: %v", err)
	}

	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("update_test_user").
		WithEmail("update_test@example.com").
		WithEncryptedPassword(encryptedPassword).
		Build()

	// トランザクションをコミット
	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	// テスト終了時にクリーンアップ
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM sessions WHERE user_id = $1", userID)
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

	// パスワード更新リクエストを送信
	newPassword := "NewPassword456!"
	form := url.Values{}
	form.Add("token", plainToken)
	form.Add("password", newPassword)
	form.Add("password_confirmation", newPassword)

	req := httptest.NewRequest("PATCH", "/password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	// I18nミドルウェアを適用
	testutil.ApplyI18nMiddleware(t, handler.Update)(rr, req)

	// リダイレクトされることを確認（ステータス303）
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("パスワード更新が失敗しました: status=%d, body=%s", rr.Code, rr.Body.String())
	}

	// ログインページにリダイレクトされることを確認
	location := rr.Header().Get("Location")
	if location != "/sign_in" {
		t.Errorf("リダイレクト先が正しくありません: got=%s, want=/sign_in", location)
	}

	// トークンが使用済みになっていることを確認
	var usedAt sql.NullTime
	err = db.QueryRow(`
		SELECT used_at FROM password_reset_tokens
		WHERE user_id = $1 AND token_digest = $2
	`, userID, tokenDigest).Scan(&usedAt)
	if err != nil {
		t.Fatalf("トークンの確認に失敗: %v", err)
	}

	if !usedAt.Valid {
		t.Error("トークンが使用済みになっていません")
	}

	// 新しいパスワードでログインできることを確認
	ctx := req.Context()
	user, err := queries.GetUserByID(ctx, userID)
	if err != nil {
		t.Fatalf("ユーザーの取得に失敗: %v", err)
	}

	err = auth.CheckPassword(user.EncryptedPassword, newPassword)
	if err != nil {
		t.Error("新しいパスワードでログインできません")
	}

	// 古いパスワードではログインできないことを確認
	err = auth.CheckPassword(user.EncryptedPassword, oldPassword)
	if err == nil {
		t.Error("古いパスワードでログインできてしまいます")
	}
}

func TestUpdate_PasswordMismatch(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("update_mismatch_test_user").
		WithEmail("update_mismatch_test@example.com").
		Build()

	// トランザクションをコミット
	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	// テスト終了時にクリーンアップ
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM sessions WHERE user_id = $1", userID)
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

	// パスワードが一致しないリクエスト
	form := url.Values{}
	form.Add("token", plainToken)
	form.Add("password", "NewPassword456!")
	form.Add("password_confirmation", "DifferentPassword789!")

	req := httptest.NewRequest("PATCH", "/password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	// I18nミドルウェアを適用
	testutil.ApplyI18nMiddleware(t, handler.Update)(rr, req)

	// リダイレクトされることを確認（バリデーションエラー）
	if rr.Code != http.StatusSeeOther {
		t.Errorf("ステータスコードが正しくありません: got=%d, want=%d", rr.Code, http.StatusSeeOther)
	}

	// editページにリダイレクトされることを確認
	location := rr.Header().Get("Location")
	expectedRedirect := "/password/edit?token=" + plainToken
	if location != expectedRedirect {
		t.Errorf("リダイレクト先が正しくありません: got=%s, want=%s", location, expectedRedirect)
	}
}

func TestUpdate_InvalidToken(t *testing.T) {
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
	form := url.Values{}
	form.Add("token", "invalid_token")
	form.Add("password", "NewPassword456!")
	form.Add("password_confirmation", "NewPassword456!")

	req := httptest.NewRequest("PATCH", "/password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	// I18nミドルウェアを適用
	testutil.ApplyI18nMiddleware(t, handler.Update)(rr, req)

	// BadRequestが返されることを確認
	if rr.Code != http.StatusBadRequest {
		t.Errorf("ステータスコードが正しくありません: got=%d, want=%d", rr.Code, http.StatusBadRequest)
	}
}
