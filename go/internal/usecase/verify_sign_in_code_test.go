package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/annict/annict/internal/auth"
	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/testutil"
)

func TestVerifySignInCodeUsecase_Execute_Success(t *testing.T) {
	t.Parallel()

	// テスト用DBをセットアップ
	db, _ := testutil.SetupTestDB(t)
	queries := query.New(db)

	// テストユーザーを作成
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin transaction failed: %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	userID := testutil.NewUserBuilder(t, setupTx).
		WithUsername("verify_code_success_user").
		WithEmail("verify_code_success@example.com").
		Build()

	if err := setupTx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// コードを生成して保存
	code := "123456"
	codeDigest, err := auth.HashCode(code)
	if err != nil {
		t.Fatalf("HashCode failed: %v", err)
	}

	params := query.CreateSignInCodeParams{
		UserID:     userID,
		CodeDigest: codeDigest,
		ExpiresAt:  time.Now().Add(15 * time.Minute),
	}

	ctx := context.Background()
	if _, err := queries.CreateSignInCode(ctx, params); err != nil {
		t.Fatalf("CreateSignInCode failed: %v", err)
	}

	// ユースケースを作成
	uc := NewVerifySignInCodeUsecase(db, queries)

	// Execute を実行（正しいコード）
	err = uc.Execute(ctx, userID, code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// コードが使用済みになっていることを確認
	savedCode, err := queries.GetValidSignInCode(ctx, userID)
	if err == nil {
		t.Errorf("GetValidSignInCode should fail after code is used, but got: %+v", savedCode)
	}
}

func TestVerifySignInCodeUsecase_Execute_CodeNotFound(t *testing.T) {
	t.Parallel()

	// テスト用DBをセットアップ
	db, _ := testutil.SetupTestDB(t)
	queries := query.New(db)

	// テストユーザーを作成
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin transaction failed: %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	userID := testutil.NewUserBuilder(t, setupTx).
		WithUsername("verify_code_not_found_user").
		WithEmail("verify_code_not_found@example.com").
		Build()

	if err := setupTx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// ユースケースを作成
	uc := NewVerifySignInCodeUsecase(db, queries)

	ctx := context.Background()

	// Execute を実行（コードが存在しない）
	err = uc.Execute(ctx, userID, "123456")
	if err == nil {
		t.Fatal("Execute should fail when code not found")
	}

	if !errors.Is(err, ErrCodeNotFound) {
		t.Errorf("Expected ErrCodeNotFound, got: %v", err)
	}
}

func TestVerifySignInCodeUsecase_Execute_CodeExpired(t *testing.T) {
	t.Parallel()

	// テスト用DBをセットアップ
	db, _ := testutil.SetupTestDB(t)
	queries := query.New(db)

	// テストユーザーを作成
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin transaction failed: %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	userID := testutil.NewUserBuilder(t, setupTx).
		WithUsername("verify_code_expired_user").
		WithEmail("verify_code_expired@example.com").
		Build()

	if err := setupTx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// 有効期限切れのコードを作成
	code := "123456"
	codeDigest, err := auth.HashCode(code)
	if err != nil {
		t.Fatalf("HashCode failed: %v", err)
	}

	params := query.CreateSignInCodeParams{
		UserID:     userID,
		CodeDigest: codeDigest,
		ExpiresAt:  time.Now().Add(-1 * time.Minute), // 1分前に期限切れ
	}

	ctx := context.Background()
	if _, err := queries.CreateSignInCode(ctx, params); err != nil {
		t.Fatalf("CreateSignInCode failed: %v", err)
	}

	// ユースケースを作成
	uc := NewVerifySignInCodeUsecase(db, queries)

	// Execute を実行（有効期限切れ）
	err = uc.Execute(ctx, userID, code)
	if err == nil {
		t.Fatal("Execute should fail when code is expired")
	}

	if !errors.Is(err, ErrCodeNotFound) {
		t.Errorf("Expected ErrCodeNotFound, got: %v", err)
	}
}

func TestVerifySignInCodeUsecase_Execute_InvalidCode(t *testing.T) {
	t.Parallel()

	// テスト用DBをセットアップ
	db, _ := testutil.SetupTestDB(t)
	queries := query.New(db)

	// テストユーザーを作成
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin transaction failed: %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	userID := testutil.NewUserBuilder(t, setupTx).
		WithUsername("verify_code_invalid_user").
		WithEmail("verify_code_invalid@example.com").
		Build()

	if err := setupTx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// コードを生成して保存
	code := "123456"
	codeDigest, err := auth.HashCode(code)
	if err != nil {
		t.Fatalf("HashCode failed: %v", err)
	}

	params := query.CreateSignInCodeParams{
		UserID:     userID,
		CodeDigest: codeDigest,
		ExpiresAt:  time.Now().Add(15 * time.Minute),
	}

	ctx := context.Background()
	savedCode, err := queries.CreateSignInCode(ctx, params)
	if err != nil {
		t.Fatalf("CreateSignInCode failed: %v", err)
	}

	// ユースケースを作成
	uc := NewVerifySignInCodeUsecase(db, queries)

	// Execute を実行（間違ったコード）
	err = uc.Execute(ctx, userID, "999999")
	if err == nil {
		t.Fatal("Execute should fail when code is invalid")
	}

	if !errors.Is(err, ErrCodeInvalid) {
		t.Errorf("Expected ErrCodeInvalid, got: %v", err)
	}

	// 試行回数がインクリメントされていることを確認
	updatedCode, err := queries.GetValidSignInCode(ctx, userID)
	if err != nil {
		t.Fatalf("GetValidSignInCode failed: %v", err)
	}

	if updatedCode.Attempts != savedCode.Attempts+1 {
		t.Errorf("Attempts: got %d, want %d", updatedCode.Attempts, savedCode.Attempts+1)
	}
}

func TestVerifySignInCodeUsecase_Execute_AttemptsExceeded(t *testing.T) {
	t.Parallel()

	// テスト用DBをセットアップ
	db, _ := testutil.SetupTestDB(t)
	queries := query.New(db)

	// テストユーザーを作成
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin transaction failed: %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	userID := testutil.NewUserBuilder(t, setupTx).
		WithUsername("verify_code_attempts_user").
		WithEmail("verify_code_attempts@example.com").
		Build()

	if err := setupTx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// コードを生成して保存
	code := "123456"
	codeDigest, err := auth.HashCode(code)
	if err != nil {
		t.Fatalf("HashCode failed: %v", err)
	}

	params := query.CreateSignInCodeParams{
		UserID:     userID,
		CodeDigest: codeDigest,
		ExpiresAt:  time.Now().Add(15 * time.Minute),
	}

	ctx := context.Background()
	savedCode, err := queries.CreateSignInCode(ctx, params)
	if err != nil {
		t.Fatalf("CreateSignInCode failed: %v", err)
	}

	// ユースケースを作成
	uc := NewVerifySignInCodeUsecase(db, queries)

	// 5回間違ったコードを入力
	for i := 0; i < 5; i++ {
		err := uc.Execute(ctx, userID, "999999")
		if err == nil {
			t.Fatalf("Execute #%d should fail", i+1)
		}

		if !errors.Is(err, ErrCodeInvalid) && !errors.Is(err, ErrCodeAttemptsExceeded) {
			t.Errorf("Execute #%d: unexpected error: %v", i+1, err)
		}
	}

	// 6回目は試行回数超過エラーになる
	err = uc.Execute(ctx, userID, code)
	if err == nil {
		t.Fatal("Execute should fail when attempts exceeded")
	}

	if !errors.Is(err, ErrCodeAttemptsExceeded) {
		t.Errorf("Expected ErrCodeAttemptsExceeded, got: %v", err)
	}

	// コードが無効化されていることを確認
	_, err = queries.GetValidSignInCode(ctx, userID)
	if err == nil {
		t.Error("GetValidSignInCode should fail after code is invalidated")
	}

	// 試行回数が5回で記録されていることを確認（used_atが設定されているため直接は取れない）
	// sign_in_codesテーブルから直接取得して確認
	rows, err := db.Query("SELECT attempts FROM sign_in_codes WHERE id = $1", savedCode.ID)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		t.Fatal("No rows returned")
	}

	var attempts int32
	if err := rows.Scan(&attempts); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if attempts != 5 {
		t.Errorf("Attempts: got %d, want 5", attempts)
	}
}

func TestVerifySignInCodeUsecase_Execute_CodeUsedOnce(t *testing.T) {
	t.Parallel()

	// テスト用DBをセットアップ
	db, _ := testutil.SetupTestDB(t)
	queries := query.New(db)

	// テストユーザーを作成
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin transaction failed: %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	userID := testutil.NewUserBuilder(t, setupTx).
		WithUsername("verify_code_used_once_user").
		WithEmail("verify_code_used_once@example.com").
		Build()

	if err := setupTx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// コードを生成して保存
	code := "123456"
	codeDigest, err := auth.HashCode(code)
	if err != nil {
		t.Fatalf("HashCode failed: %v", err)
	}

	params := query.CreateSignInCodeParams{
		UserID:     userID,
		CodeDigest: codeDigest,
		ExpiresAt:  time.Now().Add(15 * time.Minute),
	}

	ctx := context.Background()
	if _, err := queries.CreateSignInCode(ctx, params); err != nil {
		t.Fatalf("CreateSignInCode failed: %v", err)
	}

	// ユースケースを作成
	uc := NewVerifySignInCodeUsecase(db, queries)

	// 1回目：正しいコードで成功
	err = uc.Execute(ctx, userID, code)
	if err != nil {
		t.Fatalf("First Execute failed: %v", err)
	}

	// 2回目：同じコードは使用済みなのでエラー
	err = uc.Execute(ctx, userID, code)
	if err == nil {
		t.Fatal("Second Execute should fail when code is already used")
	}

	if !errors.Is(err, ErrCodeNotFound) {
		t.Errorf("Expected ErrCodeNotFound, got: %v", err)
	}
}
