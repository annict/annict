package query_test

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/testutil"
)

// TestCreateSignInCode はSignInCodeの作成をテスト
func TestCreateSignInCode(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("test_create_user").
		WithEmail("test_create@example.com").
		Build()

	// SignInCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	params := query.CreateSignInCodeParams{
		UserID:     userID,
		CodeDigest: "test_digest_123",
		ExpiresAt:  expiresAt,
	}

	code, err := queries.CreateSignInCode(context.Background(), params)
	if err != nil {
		t.Fatalf("Failed to create sign in code: %v", err)
	}

	// 基本的なアサーション
	if code.UserID != userID {
		t.Errorf("Expected user ID %d, got %d", userID, code.UserID)
	}
	if code.CodeDigest != "test_digest_123" {
		t.Errorf("Expected code digest 'test_digest_123', got %s", code.CodeDigest)
	}
	if code.Attempts != 0 {
		t.Errorf("Expected attempts 0, got %d", code.Attempts)
	}
	if code.UsedAt.Valid {
		t.Error("Expected used_at to be NULL")
	}
}

// TestGetValidSignInCode は有効なSignInCodeの取得をテスト
func TestGetValidSignInCode(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("test_get_valid_user").
		WithEmail("test_get_valid@example.com").
		Build()

	// 有効なSignInCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	params := query.CreateSignInCodeParams{
		UserID:     userID,
		CodeDigest: "valid_code_digest",
		ExpiresAt:  expiresAt,
	}
	createdCode, err := queries.CreateSignInCode(context.Background(), params)
	if err != nil {
		t.Fatalf("Failed to create sign in code: %v", err)
	}

	// 有効なコードを取得
	code, err := queries.GetValidSignInCode(context.Background(), userID)
	if err != nil {
		t.Fatalf("Failed to get valid sign in code: %v", err)
	}

	if code.ID != createdCode.ID {
		t.Errorf("Expected code ID %d, got %d", createdCode.ID, code.ID)
	}
	if code.CodeDigest != "valid_code_digest" {
		t.Errorf("Expected code digest 'valid_code_digest', got %s", code.CodeDigest)
	}
}

// TestGetValidSignInCode_Expired は期限切れのコードが取得されないことをテスト
func TestGetValidSignInCode_Expired(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("test_expired_user").
		WithEmail("test_expired@example.com").
		Build()

	// 期限切れのSignInCodeを作成
	expiresAt := time.Now().Add(-1 * time.Minute) // 1分前に期限切れ
	params := query.CreateSignInCodeParams{
		UserID:     userID,
		CodeDigest: "expired_code_digest",
		ExpiresAt:  expiresAt,
	}
	_, err := queries.CreateSignInCode(context.Background(), params)
	if err != nil {
		t.Fatalf("Failed to create sign in code: %v", err)
	}

	// 有効なコードの取得を試みる（失敗するべき）
	_, err = queries.GetValidSignInCode(context.Background(), userID)
	if err == nil {
		t.Error("Expected error for expired code, but got nil")
	}
}

// TestGetValidSignInCode_Used は使用済みのコードが取得されないことをテスト
func TestGetValidSignInCode_Used(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("test_used_user").
		WithEmail("test_used@example.com").
		Build()

	// SignInCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	params := query.CreateSignInCodeParams{
		UserID:     userID,
		CodeDigest: "used_code_digest",
		ExpiresAt:  expiresAt,
	}
	code, err := queries.CreateSignInCode(context.Background(), params)
	if err != nil {
		t.Fatalf("Failed to create sign in code: %v", err)
	}

	// コードを使用済みにする
	err = queries.MarkSignInCodeAsUsed(context.Background(), code.ID)
	if err != nil {
		t.Fatalf("Failed to mark code as used: %v", err)
	}

	// 有効なコードの取得を試みる（失敗するべき）
	_, err = queries.GetValidSignInCode(context.Background(), userID)
	if err == nil {
		t.Error("Expected error for used code, but got nil")
	}
}

// TestIncrementSignInCodeAttempts は試行回数のインクリメントをテスト
func TestIncrementSignInCodeAttempts(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("test_attempts_user").
		WithEmail("test_attempts@example.com").
		Build()

	// SignInCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	params := query.CreateSignInCodeParams{
		UserID:     userID,
		CodeDigest: "test_attempts",
		ExpiresAt:  expiresAt,
	}
	code, err := queries.CreateSignInCode(context.Background(), params)
	if err != nil {
		t.Fatalf("Failed to create sign in code: %v", err)
	}

	// 試行回数をインクリメント
	err = queries.IncrementSignInCodeAttempts(context.Background(), code.ID)
	if err != nil {
		t.Fatalf("Failed to increment attempts: %v", err)
	}

	// コードを再取得して確認
	updatedCode, err := queries.GetValidSignInCode(context.Background(), userID)
	if err != nil {
		t.Fatalf("Failed to get code after increment: %v", err)
	}

	if updatedCode.Attempts != 1 {
		t.Errorf("Expected attempts 1, got %d", updatedCode.Attempts)
	}

	// さらにインクリメント
	err = queries.IncrementSignInCodeAttempts(context.Background(), code.ID)
	if err != nil {
		t.Fatalf("Failed to increment attempts again: %v", err)
	}

	updatedCode, err = queries.GetValidSignInCode(context.Background(), userID)
	if err != nil {
		t.Fatalf("Failed to get code after second increment: %v", err)
	}

	if updatedCode.Attempts != 2 {
		t.Errorf("Expected attempts 2, got %d", updatedCode.Attempts)
	}
}

// TestMarkSignInCodeAsUsed はコードを使用済みにするテスト
func TestMarkSignInCodeAsUsed(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("test_mark_used_user").
		WithEmail("test_mark_used@example.com").
		Build()

	// SignInCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	params := query.CreateSignInCodeParams{
		UserID:     userID,
		CodeDigest: "test_used",
		ExpiresAt:  expiresAt,
	}
	code, err := queries.CreateSignInCode(context.Background(), params)
	if err != nil {
		t.Fatalf("Failed to create sign in code: %v", err)
	}

	// コードを使用済みにする
	err = queries.MarkSignInCodeAsUsed(context.Background(), code.ID)
	if err != nil {
		t.Fatalf("Failed to mark code as used: %v", err)
	}

	// 有効なコードの取得を試みる（失敗するべき）
	_, err = queries.GetValidSignInCode(context.Background(), userID)
	if err == nil {
		t.Error("Expected error after marking code as used, but got nil")
	}
}

// TestDeleteExpiredSignInCodes は期限切れコードの削除をテスト
func TestDeleteExpiredSignInCodes(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("test_delete_expired_user").
		WithEmail("test_delete_expired@example.com").
		Build()

	// 期限切れのSignInCodeを作成
	expiredTime := time.Now().Add(-2 * time.Hour)
	params1 := query.CreateSignInCodeParams{
		UserID:     userID,
		CodeDigest: "expired_code_1",
		ExpiresAt:  expiredTime,
	}
	_, err := queries.CreateSignInCode(context.Background(), params1)
	if err != nil {
		t.Fatalf("Failed to create expired code: %v", err)
	}

	// 有効なSignInCodeを作成
	validTime := time.Now().Add(15 * time.Minute)
	params2 := query.CreateSignInCodeParams{
		UserID:     userID,
		CodeDigest: "valid_code",
		ExpiresAt:  validTime,
	}
	_, err = queries.CreateSignInCode(context.Background(), params2)
	if err != nil {
		t.Fatalf("Failed to create valid code: %v", err)
	}

	// 期限切れコードを削除
	cutoffTime := time.Now().Add(-1 * time.Hour)
	err = queries.DeleteExpiredSignInCodes(context.Background(), cutoffTime)
	if err != nil {
		t.Fatalf("Failed to delete expired codes: %v", err)
	}

	// 有効なコードが残っていることを確認
	_, err = queries.GetValidSignInCode(context.Background(), userID)
	if err != nil {
		t.Error("Expected valid code to still exist after deletion of expired codes")
	}
}

// TestInvalidateUserSignInCodes はユーザーの全コード無効化をテスト
func TestInvalidateUserSignInCodes(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	// テストユーザーを作成
	userID := testutil.NewUserBuilder(t, tx).
		WithUsername("test_invalidate_user").
		WithEmail("test_invalidate@example.com").
		Build()

	// 複数のSignInCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	for i := 0; i < 3; i++ {
		params := query.CreateSignInCodeParams{
			UserID:     userID,
			CodeDigest: "code_" + string(rune('0'+i)),
			ExpiresAt:  expiresAt,
		}
		_, err := queries.CreateSignInCode(context.Background(), params)
		if err != nil {
			t.Fatalf("Failed to create sign in code %d: %v", i, err)
		}
	}

	// ユーザーのすべてのコードを無効化
	err := queries.InvalidateUserSignInCodes(context.Background(), userID)
	if err != nil {
		t.Fatalf("Failed to invalidate user codes: %v", err)
	}

	// 有効なコードの取得を試みる（失敗するべき）
	_, err = queries.GetValidSignInCode(context.Background(), userID)
	if err == nil {
		t.Error("Expected error after invalidating all user codes, but got nil")
	}
}
