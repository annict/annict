package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

func TestVerifySignInCodeUsecase_Execute_Success(t *testing.T) {
	t.Parallel()

	// テスト用DBをセットアップ
	db := testutil.GetTestDB()
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

	signInCodeRepo := repository.NewSignInCodeRepository(queries)
	ctx := context.Background()
	if _, err := signInCodeRepo.Create(ctx, repository.SignInCodeCreateParams{
		UserID:     userID,
		CodeDigest: codeDigest,
		ExpiresAt:  time.Now().Add(15 * time.Minute),
	}); err != nil {
		t.Fatalf("repo.Create failed: %v", err)
	}

	// ユースケースを作成
	userRepo := repository.NewUserRepository(queries)
	v := validator.NewSignInCodeCreateValidator()
	uc := NewVerifySignInCodeUsecase(db, signInCodeRepo, userRepo, v)

	// Execute を実行（正しいコード）
	result, err := uc.Execute(ctx, VerifySignInCodeInput{UserID: userID, Code: code})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result == nil {
		t.Fatal("result should not be nil")
	}

	if result.Username != "verify_code_success_user" {
		t.Errorf("Username: got %q, want %q", result.Username, "verify_code_success_user")
	}

	// コードが使用済みになっていることを確認
	savedCode, err := signInCodeRepo.GetValidByUserID(ctx, userID)
	if err == nil {
		t.Errorf("GetValidByUserID should fail after code is used, but got: %+v", savedCode)
	}
}

func TestVerifySignInCodeUsecase_Execute_ValidationError(t *testing.T) {
	t.Parallel()

	// テスト用DBをセットアップ
	db := testutil.GetTestDB()
	queries := query.New(db)

	// テストユーザーを作成
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin transaction failed: %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	userID := testutil.NewUserBuilder(t, setupTx).
		WithUsername("verify_code_val_error_user").
		WithEmail("verify_code_val_error@example.com").
		Build()

	if err := setupTx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// ユースケースを作成
	signInCodeRepo := repository.NewSignInCodeRepository(queries)
	userRepo := repository.NewUserRepository(queries)
	v := validator.NewSignInCodeCreateValidator()
	uc := NewVerifySignInCodeUsecase(db, signInCodeRepo, userRepo, v)

	ctx := context.Background()

	tests := []struct {
		name string
		code string
	}{
		{"空のコード", ""},
		{"5桁の数字", "12345"},
		{"7桁の数字", "1234567"},
		{"英字を含む", "12345a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.Execute(ctx, VerifySignInCodeInput{UserID: userID, Code: tt.code})
			ve := model.AsValidationError(err)
			if ve == nil {
				t.Fatalf("expected validation error but got: %v", err)
			}
		})
	}
}

func TestVerifySignInCodeUsecase_Execute_CodeNotFound(t *testing.T) {
	t.Parallel()

	// テスト用DBをセットアップ
	db := testutil.GetTestDB()
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
	signInCodeRepo := repository.NewSignInCodeRepository(queries)
	userRepo := repository.NewUserRepository(queries)
	v := validator.NewSignInCodeCreateValidator()
	uc := NewVerifySignInCodeUsecase(db, signInCodeRepo, userRepo, v)

	ctx := context.Background()

	// Execute を実行（コードが存在しない）
	_, err = uc.Execute(ctx, VerifySignInCodeInput{UserID: userID, Code: "123456"})
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
	db := testutil.GetTestDB()
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

	signInCodeRepo := repository.NewSignInCodeRepository(queries)
	ctx := context.Background()
	if _, err := signInCodeRepo.Create(ctx, repository.SignInCodeCreateParams{
		UserID:     userID,
		CodeDigest: codeDigest,
		ExpiresAt:  time.Now().Add(-1 * time.Minute), // 1分前に期限切れ
	}); err != nil {
		t.Fatalf("repo.Create failed: %v", err)
	}

	// ユースケースを作成
	userRepo := repository.NewUserRepository(queries)
	v := validator.NewSignInCodeCreateValidator()
	uc := NewVerifySignInCodeUsecase(db, signInCodeRepo, userRepo, v)

	// Execute を実行（有効期限切れ）
	_, err = uc.Execute(ctx, VerifySignInCodeInput{UserID: userID, Code: code})
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
	db := testutil.GetTestDB()
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

	signInCodeRepo := repository.NewSignInCodeRepository(queries)
	ctx := context.Background()
	savedCode, err := signInCodeRepo.Create(ctx, repository.SignInCodeCreateParams{
		UserID:     userID,
		CodeDigest: codeDigest,
		ExpiresAt:  time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		t.Fatalf("repo.Create failed: %v", err)
	}

	// ユースケースを作成
	userRepo := repository.NewUserRepository(queries)
	v := validator.NewSignInCodeCreateValidator()
	uc := NewVerifySignInCodeUsecase(db, signInCodeRepo, userRepo, v)

	// Execute を実行（間違ったコード）
	_, err = uc.Execute(ctx, VerifySignInCodeInput{UserID: userID, Code: "999999"})
	if err == nil {
		t.Fatal("Execute should fail when code is invalid")
	}

	if !errors.Is(err, ErrCodeInvalid) {
		t.Errorf("Expected ErrCodeInvalid, got: %v", err)
	}

	// 試行回数がインクリメントされていることを確認
	updatedCode, err := signInCodeRepo.GetValidByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("GetValidByUserID failed: %v", err)
	}

	if updatedCode.Attempts != savedCode.Attempts+1 {
		t.Errorf("Attempts: got %d, want %d", updatedCode.Attempts, savedCode.Attempts+1)
	}
}

func TestVerifySignInCodeUsecase_Execute_AttemptsExceeded(t *testing.T) {
	t.Parallel()

	// テスト用DBをセットアップ
	db := testutil.GetTestDB()
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

	signInCodeRepo := repository.NewSignInCodeRepository(queries)
	ctx := context.Background()
	savedCode, err := signInCodeRepo.Create(ctx, repository.SignInCodeCreateParams{
		UserID:     userID,
		CodeDigest: codeDigest,
		ExpiresAt:  time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		t.Fatalf("repo.Create failed: %v", err)
	}

	// ユースケースを作成
	userRepo := repository.NewUserRepository(queries)
	v := validator.NewSignInCodeCreateValidator()
	uc := NewVerifySignInCodeUsecase(db, signInCodeRepo, userRepo, v)

	// 5回間違ったコードを入力
	for i := 0; i < 5; i++ {
		_, err := uc.Execute(ctx, VerifySignInCodeInput{UserID: userID, Code: "999999"})
		if err == nil {
			t.Fatalf("Execute #%d should fail", i+1)
		}

		if !errors.Is(err, ErrCodeInvalid) && !errors.Is(err, ErrCodeAttemptsExceeded) {
			t.Errorf("Execute #%d: unexpected error: %v", i+1, err)
		}
	}

	// 6回目は試行回数超過エラーになる
	_, err = uc.Execute(ctx, VerifySignInCodeInput{UserID: userID, Code: code})
	if err == nil {
		t.Fatal("Execute should fail when attempts exceeded")
	}

	if !errors.Is(err, ErrCodeAttemptsExceeded) {
		t.Errorf("Expected ErrCodeAttemptsExceeded, got: %v", err)
	}

	// コードが無効化されていることを確認
	_, err = signInCodeRepo.GetValidByUserID(ctx, userID)
	if err == nil {
		t.Error("GetValidByUserID should fail after code is invalidated")
	}

	// 試行回数が5回で記録されていることを確認（used_atが設定されているため直接は取れない）
	// sign_in_codesテーブルから直接取得して確認
	rows, err := db.Query("SELECT attempts FROM sign_in_codes WHERE id = $1", int64(savedCode.ID))
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
	db := testutil.GetTestDB()
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

	signInCodeRepo := repository.NewSignInCodeRepository(queries)
	ctx := context.Background()
	if _, err := signInCodeRepo.Create(ctx, repository.SignInCodeCreateParams{
		UserID:     userID,
		CodeDigest: codeDigest,
		ExpiresAt:  time.Now().Add(15 * time.Minute),
	}); err != nil {
		t.Fatalf("repo.Create failed: %v", err)
	}

	// ユースケースを作成
	userRepo := repository.NewUserRepository(queries)
	v := validator.NewSignInCodeCreateValidator()
	uc := NewVerifySignInCodeUsecase(db, signInCodeRepo, userRepo, v)

	// 1回目：正しいコードで成功
	result, err := uc.Execute(ctx, VerifySignInCodeInput{UserID: userID, Code: code})
	if err != nil {
		t.Fatalf("First Execute failed: %v", err)
	}

	if result == nil {
		t.Fatal("result should not be nil")
	}

	// 2回目：同じコードは使用済みなのでエラー
	_, err = uc.Execute(ctx, VerifySignInCodeInput{UserID: userID, Code: code})
	if err == nil {
		t.Fatal("Second Execute should fail when code is already used")
	}

	if !errors.Is(err, ErrCodeNotFound) {
		t.Errorf("Expected ErrCodeNotFound, got: %v", err)
	}
}
