package usecase

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/password_reset"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

// TestUpdatePasswordResetUsecase_Execute は正常系のテストです
func TestUpdatePasswordResetUsecase_Execute(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// パスワードをハッシュ化
	hashedPassword, err := auth.HashPassword("oldpassword123")
	if err != nil {
		t.Fatalf("パスワードのハッシュ化に失敗: %v", err)
	}

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("update_password_test").
		WithEmail("update@example.com").
		WithEncryptedPassword(hashedPassword).
		Build()

	// トークンを生成してデータベースに保存
	token, err := password_reset.GenerateToken()
	if err != nil {
		t.Fatalf("トークン生成に失敗: %v", err)
	}
	tokenDigest := password_reset.HashToken(token)

	queries := query.New(db).WithTx(tx)
	_, err = queries.CreatePasswordResetToken(context.Background(), query.CreatePasswordResetTokenParams{
		UserID:      userID,
		TokenDigest: tokenDigest,
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	})
	if err != nil {
		t.Fatalf("トークンの保存に失敗: %v", err)
	}

	// トランザクションをコミット（UseCaseが新しいトランザクションを開始するため）
	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	// テスト終了時にデータを削除
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM sessions WHERE user_id = $1", userID)
		_, _ = db.Exec("DELETE FROM password_reset_tokens WHERE user_id = $1", userID)
		_, _ = db.Exec("DELETE FROM users WHERE id = $1", userID)
	})

	// UseCase を作成
	queriesWithoutTx := query.New(db)
	sessionRepo := repository.NewSessionRepository(queriesWithoutTx)
	v := validator.NewPasswordUpdateValidator()
	uc := NewUpdatePasswordResetUsecase(db, repository.NewPasswordResetTokenRepository(queriesWithoutTx), repository.NewUserRepository(queriesWithoutTx), sessionRepo, v)

	// パスワードを更新
	ctx := context.Background()
	newPassword := "newpassword456"
	result, err := uc.Execute(ctx, UpdatePasswordResetInput{
		Token:                token,
		Password:             newPassword,
		PasswordConfirmation: newPassword,
	})
	if err != nil {
		t.Fatalf("パスワード更新に失敗: %v", err)
	}

	// 結果を検証
	if result.UserID != userID {
		t.Errorf("ユーザーIDが一致しません: got %d, want %d", result.UserID, userID)
	}
	if result.SessionID == "" {
		t.Error("セッションIDが空です")
	}

	// パスワードが更新されているか確認
	user, err := queriesWithoutTx.GetUserByID(ctx, userID)
	if err != nil {
		t.Fatalf("ユーザー情報の取得に失敗: %v", err)
	}

	if user.EncryptedPassword == "" {
		t.Error("encrypted_passwordが空です")
	}

	// トークンが使用済みになっているか確認
	var usedAt sql.NullTime
	err = db.QueryRow("SELECT used_at FROM password_reset_tokens WHERE token_digest = $1", tokenDigest).Scan(&usedAt)
	if err != nil {
		t.Fatalf("トークンの取得に失敗: %v", err)
	}
	if !usedAt.Valid {
		t.Error("トークンが使用済みになっていません")
	}
}

// TestUpdatePasswordResetUsecase_Execute_ValidationError はバリデーションエラーのテストです
func TestUpdatePasswordResetUsecase_Execute_ValidationError(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// トランザクションをコミット
	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	// UseCase を作成
	queries := query.New(db)
	sessionRepo := repository.NewSessionRepository(queries)
	v := validator.NewPasswordUpdateValidator()
	uc := NewUpdatePasswordResetUsecase(db, repository.NewPasswordResetTokenRepository(queries), repository.NewUserRepository(queries), sessionRepo, v)

	// パスワードが一致しない入力
	ctx := context.Background()
	_, err := uc.Execute(ctx, UpdatePasswordResetInput{
		Token:                "some_token",
		Password:             "password123",
		PasswordConfirmation: "different_password",
	})
	ve := model.AsValidationError(err)

	// バリデーションエラーが返されることを確認
	if ve == nil {
		t.Fatalf("バリデーションエラーが期待されましたが、返されませんでした: %v", err)
	}
}

// TestUpdatePasswordResetUsecase_Execute_WithInvalidToken はfail-fastケース：無効なトークンでのパスワード更新をテストします
func TestUpdatePasswordResetUsecase_Execute_WithInvalidToken(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// トランザクションをコミット
	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	// UseCase を作成
	queries := query.New(db)
	sessionRepo := repository.NewSessionRepository(queries)
	v := validator.NewPasswordUpdateValidator()
	uc := NewUpdatePasswordResetUsecase(db, repository.NewPasswordResetTokenRepository(queries), repository.NewUserRepository(queries), sessionRepo, v)

	// 存在しないトークンでパスワード更新を試みる
	ctx := context.Background()
	invalidToken := "invalid_token_12345678901234567890123456789012"
	newPassword := "newpassword456"
	_, err := uc.Execute(ctx, UpdatePasswordResetInput{
		Token:                invalidToken,
		Password:             newPassword,
		PasswordConfirmation: newPassword,
	})

	// エラーが発生することを確認
	if err == nil {
		t.Error("無効なトークンでパスワード更新が成功してしまいました")
	}

	// エラーの種類を確認
	if !errors.Is(err, repository.ErrInvalidPasswordResetToken) {
		t.Errorf("期待されるエラーと一致しません: got %v, want %v", err, repository.ErrInvalidPasswordResetToken)
	}
}

// TestUpdatePasswordResetUsecase_Execute_WithUsedToken はfail-fastケース：使用済みトークンでのパスワード更新をテストします
func TestUpdatePasswordResetUsecase_Execute_WithUsedToken(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// パスワードをハッシュ化
	hashedPassword, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("パスワードのハッシュ化に失敗: %v", err)
	}

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("used_token_test").
		WithEmail("used@example.com").
		WithEncryptedPassword(hashedPassword).
		Build()

	// トークンを生成してデータベースに保存（使用済みマークを付ける）
	token, err := password_reset.GenerateToken()
	if err != nil {
		t.Fatalf("トークン生成に失敗: %v", err)
	}
	tokenDigest := password_reset.HashToken(token)

	queries := query.New(db).WithTx(tx)
	resetToken, err := queries.CreatePasswordResetToken(context.Background(), query.CreatePasswordResetTokenParams{
		UserID:      userID,
		TokenDigest: tokenDigest,
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	})
	if err != nil {
		t.Fatalf("トークンの保存に失敗: %v", err)
	}

	// トークンを使用済みにマーク
	if err := queries.MarkPasswordResetTokenAsUsed(context.Background(), resetToken.ID); err != nil {
		t.Fatalf("トークンの使用済みマークに失敗: %v", err)
	}

	// トランザクションをコミット
	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	// テスト終了時にデータを削除
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM password_reset_tokens WHERE user_id = $1", userID)
		_, _ = db.Exec("DELETE FROM users WHERE id = $1", userID)
	})

	// UseCase を作成
	queriesWithoutTx := query.New(db)
	sessionRepo := repository.NewSessionRepository(queriesWithoutTx)
	v := validator.NewPasswordUpdateValidator()
	uc := NewUpdatePasswordResetUsecase(db, repository.NewPasswordResetTokenRepository(queriesWithoutTx), repository.NewUserRepository(queriesWithoutTx), sessionRepo, v)

	// 使用済みトークンでパスワード更新を試みる
	ctx := context.Background()
	newPassword := "newpassword456"
	_, err = uc.Execute(ctx, UpdatePasswordResetInput{
		Token:                token,
		Password:             newPassword,
		PasswordConfirmation: newPassword,
	})

	// エラーが発生することを確認
	if err == nil {
		t.Error("使用済みトークンでパスワード更新が成功してしまいました")
	}

	// エラーの種類を確認
	if !errors.Is(err, repository.ErrInvalidPasswordResetToken) {
		t.Errorf("期待されるエラーと一致しません: got %v, want %v", err, repository.ErrInvalidPasswordResetToken)
	}
}

// TestUpdatePasswordResetUsecase_Execute_WithExpiredToken はfail-fastケース：期限切れトークンでのパスワード更新をテストします
func TestUpdatePasswordResetUsecase_Execute_WithExpiredToken(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// パスワードをハッシュ化
	hashedPassword, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("パスワードのハッシュ化に失敗: %v", err)
	}

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("expired_token_test").
		WithEmail("expired@example.com").
		WithEncryptedPassword(hashedPassword).
		Build()

	// トークンを生成してデータベースに保存（過去の有効期限を設定）
	token, err := password_reset.GenerateToken()
	if err != nil {
		t.Fatalf("トークン生成に失敗: %v", err)
	}
	tokenDigest := password_reset.HashToken(token)

	queries := query.New(db).WithTx(tx)
	_, err = queries.CreatePasswordResetToken(context.Background(), query.CreatePasswordResetTokenParams{
		UserID:      userID,
		TokenDigest: tokenDigest,
		ExpiresAt:   time.Now().Add(-1 * time.Hour), // 1時間前に期限切れ
	})
	if err != nil {
		t.Fatalf("トークンの保存に失敗: %v", err)
	}

	// トランザクションをコミット
	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	// テスト終了時にデータを削除
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM password_reset_tokens WHERE user_id = $1", userID)
		_, _ = db.Exec("DELETE FROM users WHERE id = $1", userID)
	})

	// UseCase を作成
	queriesWithoutTx := query.New(db)
	sessionRepo := repository.NewSessionRepository(queriesWithoutTx)
	v := validator.NewPasswordUpdateValidator()
	uc := NewUpdatePasswordResetUsecase(db, repository.NewPasswordResetTokenRepository(queriesWithoutTx), repository.NewUserRepository(queriesWithoutTx), sessionRepo, v)

	// 期限切れトークンでパスワード更新を試みる
	ctx := context.Background()
	newPassword := "newpassword456"
	_, err = uc.Execute(ctx, UpdatePasswordResetInput{
		Token:                token,
		Password:             newPassword,
		PasswordConfirmation: newPassword,
	})

	// エラーが発生することを確認
	if err == nil {
		t.Error("期限切れトークンでパスワード更新が成功してしまいました")
	}

	// エラーの種類を確認
	if !errors.Is(err, repository.ErrInvalidPasswordResetToken) {
		t.Errorf("期待されるエラーと一致しません: got %v, want %v", err, repository.ErrInvalidPasswordResetToken)
	}
}

// TestUpdatePasswordResetUsecase_Execute_WithNonExistentUser はfail-fastケース：存在しないユーザーのトークンでのパスワード更新をテストします
func TestUpdatePasswordResetUsecase_Execute_WithNonExistentUser(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// トークンを生成してデータベースに保存（存在しないユーザーIDを使用）
	token, err := password_reset.GenerateToken()
	if err != nil {
		t.Fatalf("トークン生成に失敗: %v", err)
	}
	tokenDigest := password_reset.HashToken(token)

	queries := query.New(db).WithTx(tx)

	// パスワードをハッシュ化
	hashedPassword, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("パスワードのハッシュ化に失敗: %v", err)
	}

	// 一時的なユーザーを作成
	tempUserID := testutil.NewUserBuilder(t, tx).
		WithUsername("temp_user_for_deletion").
		WithEmail("temp@example.com").
		WithEncryptedPassword(hashedPassword).
		Build()

	// トークンを作成
	_, err = queries.CreatePasswordResetToken(context.Background(), query.CreatePasswordResetTokenParams{
		UserID:      tempUserID,
		TokenDigest: tokenDigest,
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	})
	if err != nil {
		t.Fatalf("トークンの保存に失敗: %v", err)
	}

	// トランザクションをコミット
	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	// ユーザーを削除する前に関連レコードを削除
	_, err = db.Exec("DELETE FROM email_notifications WHERE user_id = $1", tempUserID)
	if err != nil {
		t.Fatalf("メール通知設定の削除に失敗: %v", err)
	}
	_, err = db.Exec("DELETE FROM settings WHERE user_id = $1", tempUserID)
	if err != nil {
		t.Fatalf("設定の削除に失敗: %v", err)
	}
	_, err = db.Exec("DELETE FROM profiles WHERE user_id = $1", tempUserID)
	if err != nil {
		t.Fatalf("プロフィールの削除に失敗: %v", err)
	}

	// ユーザーを削除（外部キー制約でトークンも削除される）
	_, err = db.Exec("DELETE FROM users WHERE id = $1", tempUserID)
	if err != nil {
		t.Fatalf("ユーザーの削除に失敗: %v", err)
	}

	// テスト終了時にデータを削除
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM password_reset_tokens WHERE user_id = $1", tempUserID)
		_, _ = db.Exec("DELETE FROM email_notifications WHERE user_id = $1", tempUserID)
		_, _ = db.Exec("DELETE FROM settings WHERE user_id = $1", tempUserID)
		_, _ = db.Exec("DELETE FROM profiles WHERE user_id = $1", tempUserID)
	})

	// UseCase を作成
	queriesWithoutTx := query.New(db)
	sessionRepo := repository.NewSessionRepository(queriesWithoutTx)
	v := validator.NewPasswordUpdateValidator()
	uc := NewUpdatePasswordResetUsecase(db, repository.NewPasswordResetTokenRepository(queriesWithoutTx), repository.NewUserRepository(queriesWithoutTx), sessionRepo, v)

	// 削除されたユーザーのトークンでパスワード更新を試みる
	ctx := context.Background()
	newPassword := "newpassword456"
	_, err = uc.Execute(ctx, UpdatePasswordResetInput{
		Token:                token,
		Password:             newPassword,
		PasswordConfirmation: newPassword,
	})

	// エラーが発生することを確認（トークンが削除されているため）
	if err == nil {
		t.Error("削除されたユーザーのトークンでパスワード更新が成功してしまいました")
	}
}

// TestUpdatePasswordResetUsecase_Execute_WithNullUserData はfail-fastケース：ユーザーデータにNULL値がある場合のテストです
func TestUpdatePasswordResetUsecase_Execute_WithNullUserData(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// パスワードをハッシュ化
	hashedPassword, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("パスワードのハッシュ化に失敗: %v", err)
	}

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("null_data_test").
		WithEmail("null@example.com").
		WithEncryptedPassword(hashedPassword).
		Build()

	// トークンを生成してデータベースに保存
	token, err := password_reset.GenerateToken()
	if err != nil {
		t.Fatalf("トークン生成に失敗: %v", err)
	}
	tokenDigest := password_reset.HashToken(token)

	queries := query.New(db).WithTx(tx)
	_, err = queries.CreatePasswordResetToken(context.Background(), query.CreatePasswordResetTokenParams{
		UserID:      userID,
		TokenDigest: tokenDigest,
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	})
	if err != nil {
		t.Fatalf("トークンの保存に失敗: %v", err)
	}

	// トランザクションをコミット
	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	// ユーザーのcreated_atをNULLに設定（データベース制約違反のシミュレーション）
	_, err = db.Exec("ALTER TABLE users ALTER COLUMN created_at DROP NOT NULL")
	if err != nil {
		t.Fatalf("NOT NULL制約の削除に失敗: %v", err)
	}
	_, err = db.Exec("UPDATE users SET created_at = NULL WHERE id = $1", userID)
	if err != nil {
		t.Fatalf("created_atをNULLに設定できませんでした: %v", err)
	}

	// テスト終了時にデータを削除し、制約を復元
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM sessions WHERE user_id = $1", userID)
		_, _ = db.Exec("DELETE FROM password_reset_tokens WHERE user_id = $1", userID)
		_, _ = db.Exec("DELETE FROM users WHERE id = $1", userID)
		_, _ = db.Exec("ALTER TABLE users ALTER COLUMN created_at SET NOT NULL")
	})

	// UseCase を作成
	queriesWithoutTx := query.New(db)
	sessionRepo := repository.NewSessionRepository(queriesWithoutTx)
	v := validator.NewPasswordUpdateValidator()
	uc := NewUpdatePasswordResetUsecase(db, repository.NewPasswordResetTokenRepository(queriesWithoutTx), repository.NewUserRepository(queriesWithoutTx), sessionRepo, v)

	// パスワード更新を試みる
	ctx := context.Background()
	newPassword := "newpassword456"
	_, err = uc.Execute(ctx, UpdatePasswordResetInput{
		Token:                token,
		Password:             newPassword,
		PasswordConfirmation: newPassword,
	})

	// データベース制約違反のエラーが発生することを確認
	if err == nil {
		t.Error("NULL値を持つユーザーデータに対してパスワード更新が成功してしまいました")
	}

	// エラーメッセージに「データベース制約違反」が含まれることを確認
	if err != nil && !strings.Contains(err.Error(), "データベース制約違反") && !strings.Contains(err.Error(), "created_at") {
		t.Errorf("期待されるエラーメッセージが含まれていません: %v", err)
	}
}

// TestUpdatePasswordResetUsecase_Execute_TransactionRollback はfail-fastケース：トランザクションロールバックのテストです
func TestUpdatePasswordResetUsecase_Execute_TransactionRollback(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)

	// パスワードをハッシュ化
	hashedPassword, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("パスワードのハッシュ化に失敗: %v", err)
	}

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("rollback_test").
		WithEmail("rollback@example.com").
		WithEncryptedPassword(hashedPassword).
		Build()

	// トークンを生成してデータベースに保存
	token, err := password_reset.GenerateToken()
	if err != nil {
		t.Fatalf("トークン生成に失敗: %v", err)
	}
	tokenDigest := password_reset.HashToken(token)

	queries := query.New(db).WithTx(tx)
	_, err = queries.CreatePasswordResetToken(context.Background(), query.CreatePasswordResetTokenParams{
		UserID:      userID,
		TokenDigest: tokenDigest,
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	})
	if err != nil {
		t.Fatalf("トークンの保存に失敗: %v", err)
	}

	// 元のパスワードハッシュを取得
	user, err := queries.GetUserByID(context.Background(), userID)
	if err != nil {
		t.Fatalf("ユーザー情報の取得に失敗: %v", err)
	}
	originalPasswordHash := user.EncryptedPassword

	// トランザクションをコミット
	if err := tx.Commit(); err != nil {
		t.Fatalf("トランザクションのコミットに失敗: %v", err)
	}

	// テスト終了時にデータを削除
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM sessions WHERE user_id = $1", userID)
		_, _ = db.Exec("DELETE FROM password_reset_tokens WHERE user_id = $1", userID)
		_, _ = db.Exec("DELETE FROM users WHERE id = $1", userID)
	})

	// UseCase を作成
	queriesWithoutTx := query.New(db)
	sessionRepo := repository.NewSessionRepository(queriesWithoutTx)
	v := validator.NewPasswordUpdateValidator()
	uc := NewUpdatePasswordResetUsecase(db, repository.NewPasswordResetTokenRepository(queriesWithoutTx), repository.NewUserRepository(queriesWithoutTx), sessionRepo, v)

	// 無効なパスワード（空文字列）でパスワード更新を試みる
	ctx := context.Background()
	invalidPassword := "" // 空のパスワード
	_, err = uc.Execute(ctx, UpdatePasswordResetInput{
		Token:                token,
		Password:             invalidPassword,
		PasswordConfirmation: invalidPassword,
	})
	ve := model.AsValidationError(err)

	// バリデーションエラーが返されることを確認（空パスワードは形式バリデーションで弾かれる）
	if ve == nil {
		t.Fatalf("バリデーションエラーが期待されましたが、返されませんでした: %v", err)
	}

	// パスワードが更新されていないことを確認
	user2, err := queriesWithoutTx.GetUserByID(ctx, userID)
	if err != nil {
		t.Fatalf("ユーザー情報の取得に失敗: %v", err)
	}
	if user2.EncryptedPassword != originalPasswordHash {
		t.Error("バリデーションエラーにもかかわらずパスワードが変更されています")
	}

	// トークンが使用済みになっていないことを確認
	resetToken, err := queriesWithoutTx.GetPasswordResetTokenByDigest(ctx, tokenDigest)
	if err != nil {
		t.Fatalf("トークンの取得に失敗: %v", err)
	}
	if resetToken.UsedAt.Valid {
		t.Error("バリデーションエラーにもかかわらずトークンが使用済みになっています")
	}
}
