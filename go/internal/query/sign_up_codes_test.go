package query_test

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/testutil"
)

// TestCreateSignUpCode はSignUpCodeの作成をテスト
func TestCreateSignUpCode(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	// SignUpCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	params := query.CreateSignUpCodeParams{
		Email:      "test_create@example.com",
		CodeDigest: "test_digest_123",
		ExpiresAt:  expiresAt,
	}

	code, err := queries.CreateSignUpCode(context.Background(), params)
	if err != nil {
		t.Fatalf("Failed to create sign up code: %v", err)
	}

	// 基本的なアサーション
	if code.Email != "test_create@example.com" {
		t.Errorf("Expected email 'test_create@example.com', got %s", code.Email)
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

// TestGetValidSignUpCode は有効なSignUpCodeの取得をテスト
func TestGetValidSignUpCode(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	email := "test_get_valid@example.com"

	// 有効なSignUpCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	params := query.CreateSignUpCodeParams{
		Email:      email,
		CodeDigest: "valid_code_digest",
		ExpiresAt:  expiresAt,
	}
	createdCode, err := queries.CreateSignUpCode(context.Background(), params)
	if err != nil {
		t.Fatalf("Failed to create sign up code: %v", err)
	}

	// 有効なコードを取得
	code, err := queries.GetValidSignUpCode(context.Background(), email)
	if err != nil {
		t.Fatalf("Failed to get valid sign up code: %v", err)
	}

	if code.ID != createdCode.ID {
		t.Errorf("Expected code ID %d, got %d", createdCode.ID, code.ID)
	}
	if code.CodeDigest != "valid_code_digest" {
		t.Errorf("Expected code digest 'valid_code_digest', got %s", code.CodeDigest)
	}
}

// TestGetValidSignUpCode_Expired は期限切れのコードが取得されないことをテスト
func TestGetValidSignUpCode_Expired(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	email := "test_expired@example.com"

	// 期限切れのSignUpCodeを作成
	expiresAt := time.Now().Add(-1 * time.Minute) // 1分前に期限切れ
	params := query.CreateSignUpCodeParams{
		Email:      email,
		CodeDigest: "expired_code_digest",
		ExpiresAt:  expiresAt,
	}
	_, err := queries.CreateSignUpCode(context.Background(), params)
	if err != nil {
		t.Fatalf("Failed to create sign up code: %v", err)
	}

	// 有効なコードの取得を試みる（失敗するべき）
	_, err = queries.GetValidSignUpCode(context.Background(), email)
	if err == nil {
		t.Error("Expected error for expired code, but got nil")
	}
}

// TestGetValidSignUpCode_Used は使用済みのコードが取得されないことをテスト
func TestGetValidSignUpCode_Used(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	email := "test_used@example.com"

	// SignUpCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	params := query.CreateSignUpCodeParams{
		Email:      email,
		CodeDigest: "used_code_digest",
		ExpiresAt:  expiresAt,
	}
	code, err := queries.CreateSignUpCode(context.Background(), params)
	if err != nil {
		t.Fatalf("Failed to create sign up code: %v", err)
	}

	// コードを使用済みにする
	err = queries.MarkSignUpCodeAsUsed(context.Background(), code.ID)
	if err != nil {
		t.Fatalf("Failed to mark code as used: %v", err)
	}

	// 有効なコードの取得を試みる（失敗するべき）
	_, err = queries.GetValidSignUpCode(context.Background(), email)
	if err == nil {
		t.Error("Expected error for used code, but got nil")
	}
}

// TestGetValidSignUpCode_ExceedsAttempts は試行回数制限を超えたコードが取得されないことをテスト
func TestGetValidSignUpCode_ExceedsAttempts(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	email := "test_attempts@example.com"

	// SignUpCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	params := query.CreateSignUpCodeParams{
		Email:      email,
		CodeDigest: "attempts_code_digest",
		ExpiresAt:  expiresAt,
	}
	code, err := queries.CreateSignUpCode(context.Background(), params)
	if err != nil {
		t.Fatalf("Failed to create sign up code: %v", err)
	}

	// 試行回数を5回インクリメント
	for i := 0; i < 5; i++ {
		err = queries.IncrementSignUpCodeAttempts(context.Background(), code.ID)
		if err != nil {
			t.Fatalf("Failed to increment attempts: %v", err)
		}
	}

	// 有効なコードの取得を試みる（失敗するべき）
	_, err = queries.GetValidSignUpCode(context.Background(), email)
	if err == nil {
		t.Error("Expected error for code with too many attempts, but got nil")
	}
}

// TestIncrementSignUpCodeAttempts は試行回数のインクリメントをテスト
func TestIncrementSignUpCodeAttempts(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	email := "test_increment@example.com"

	// SignUpCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	params := query.CreateSignUpCodeParams{
		Email:      email,
		CodeDigest: "test_attempts",
		ExpiresAt:  expiresAt,
	}
	code, err := queries.CreateSignUpCode(context.Background(), params)
	if err != nil {
		t.Fatalf("Failed to create sign up code: %v", err)
	}

	// 試行回数をインクリメント
	err = queries.IncrementSignUpCodeAttempts(context.Background(), code.ID)
	if err != nil {
		t.Fatalf("Failed to increment attempts: %v", err)
	}

	// コードを再取得して確認
	updatedCode, err := queries.GetValidSignUpCode(context.Background(), email)
	if err != nil {
		t.Fatalf("Failed to get code after increment: %v", err)
	}

	if updatedCode.Attempts != 1 {
		t.Errorf("Expected attempts 1, got %d", updatedCode.Attempts)
	}

	// さらにインクリメント
	err = queries.IncrementSignUpCodeAttempts(context.Background(), code.ID)
	if err != nil {
		t.Fatalf("Failed to increment attempts again: %v", err)
	}

	updatedCode, err = queries.GetValidSignUpCode(context.Background(), email)
	if err != nil {
		t.Fatalf("Failed to get code after second increment: %v", err)
	}

	if updatedCode.Attempts != 2 {
		t.Errorf("Expected attempts 2, got %d", updatedCode.Attempts)
	}
}

// TestMarkSignUpCodeAsUsed はコードを使用済みにするテスト
func TestMarkSignUpCodeAsUsed(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	email := "test_mark_used@example.com"

	// SignUpCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	params := query.CreateSignUpCodeParams{
		Email:      email,
		CodeDigest: "test_used",
		ExpiresAt:  expiresAt,
	}
	code, err := queries.CreateSignUpCode(context.Background(), params)
	if err != nil {
		t.Fatalf("Failed to create sign up code: %v", err)
	}

	// コードを使用済みにする
	err = queries.MarkSignUpCodeAsUsed(context.Background(), code.ID)
	if err != nil {
		t.Fatalf("Failed to mark code as used: %v", err)
	}

	// 有効なコードの取得を試みる（失敗するべき）
	_, err = queries.GetValidSignUpCode(context.Background(), email)
	if err == nil {
		t.Error("Expected error after marking code as used, but got nil")
	}
}

// TestInvalidateSignUpCodesByEmail はメールアドレスの全コード無効化をテスト
func TestInvalidateSignUpCodesByEmail(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	email := "test_invalidate@example.com"

	// 複数のSignUpCodeを作成
	expiresAt := time.Now().Add(15 * time.Minute)
	for i := 0; i < 3; i++ {
		params := query.CreateSignUpCodeParams{
			Email:      email,
			CodeDigest: "code_" + string(rune('0'+i)),
			ExpiresAt:  expiresAt,
		}
		_, err := queries.CreateSignUpCode(context.Background(), params)
		if err != nil {
			t.Fatalf("Failed to create sign up code %d: %v", i, err)
		}
	}

	// メールアドレスのすべてのコードを無効化
	err := queries.InvalidateSignUpCodesByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("Failed to invalidate codes by email: %v", err)
	}

	// 有効なコードの取得を試みる（失敗するべき）
	_, err = queries.GetValidSignUpCode(context.Background(), email)
	if err == nil {
		t.Error("Expected error after invalidating all codes by email, but got nil")
	}
}
