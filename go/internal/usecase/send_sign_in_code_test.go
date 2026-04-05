package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

func TestSendSignInCodeUsecase_Execute(t *testing.T) {
	t.Parallel()

	// テスト用DBをセットアップ
	db := testutil.GetTestDB(t)
	queries := query.New(db)

	// 別のトランザクションでテストユーザーを作成してコミット
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin transaction failed: %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	testEmail := "sign_in_code_test_1@example.com"
	testutil.NewUserBuilder(t, setupTx).
		WithUsername("sign_in_code_test_user_1").
		WithEmail(testEmail).
		WithEncryptedPassword(""). // パスワードなし（コードログイン）
		Build()

	if err := setupTx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// ユースケースを作成（Dispatcherはnil）
	v := validator.NewCreateSignInValidator()
	uc := NewSendSignInCodeUsecase(db, repository.NewSignInCodeRepository(queries), repository.NewUserRepository(queries), nil, v)

	ctx := context.Background()

	// Execute を実行
	result, err := uc.Execute(ctx, SendSignInCodeInput{Email: testEmail})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// バリデーションエラーがないことを確認
	if result.FormErrors != nil && result.FormErrors.HasErrors() {
		t.Fatalf("unexpected form errors: %+v", result.FormErrors)
	}

	// パスワードなしユーザーなのでHasPasswordはfalse
	if result.HasPassword {
		t.Error("HasPassword should be false for user without password")
	}

	// 結果の検証
	if result.Email != testEmail {
		t.Errorf("Email: got %q, want %q", result.Email, testEmail)
	}

	// コードが6桁の数字であることを確認
	if len(result.Code) != 6 {
		t.Errorf("Code length: got %d, want 6", len(result.Code))
	}
	for _, c := range result.Code {
		if c < '0' || c > '9' {
			t.Errorf("Code contains non-digit character: %c", c)
		}
	}

	// データベースに保存されたコードを取得
	savedCode, err := queries.GetValidSignInCode(ctx, result.UserID)
	if err != nil {
		t.Fatalf("GetValidSignInCode failed: %v", err)
	}

	// コードダイジェストが bcrypt でハッシュ化されていることを確認
	if !auth.VerifyCode(result.Code, savedCode.CodeDigest) {
		t.Error("Code verification failed")
	}

	// 有効期限が 15 分後に設定されていることを確認（許容誤差: 10秒）
	expectedExpiry := time.Now().Add(15 * time.Minute)
	timeDiff := savedCode.ExpiresAt.Sub(expectedExpiry)
	if timeDiff < -10*time.Second || timeDiff > 10*time.Second {
		t.Errorf("ExpiresAt: got %v, want around %v (diff: %v)", savedCode.ExpiresAt, expectedExpiry, timeDiff)
	}

	// 試行回数が0であることを確認
	if savedCode.Attempts != 0 {
		t.Errorf("Attempts: got %d, want 0", savedCode.Attempts)
	}

	// 使用済みフラグがnilであることを確認
	if savedCode.UsedAt.Valid {
		t.Error("UsedAt should be NULL")
	}
}

func TestSendSignInCodeUsecase_Execute_InvalidatesOldCodes(t *testing.T) {
	t.Parallel()

	// テスト用DBをセットアップ
	db := testutil.GetTestDB(t)
	queries := query.New(db)

	// 別のトランザクションでテストユーザーを作成してコミット
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin transaction failed: %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	testEmail := "sign_in_code_test_2@example.com"
	testutil.NewUserBuilder(t, setupTx).
		WithUsername("sign_in_code_test_user_2").
		WithEmail(testEmail).
		WithEncryptedPassword(""). // パスワードなし
		Build()

	if err := setupTx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// ユースケースを作成
	v := validator.NewCreateSignInValidator()
	uc := NewSendSignInCodeUsecase(db, repository.NewSignInCodeRepository(queries), repository.NewUserRepository(queries), nil, v)

	ctx := context.Background()

	// 1回目のコード生成
	_, err = uc.Execute(ctx, SendSignInCodeInput{Email: testEmail})
	if err != nil {
		t.Fatalf("First Execute failed: %v", err)
	}

	// 2回目のコード生成（古いコードは無効化されるはず）
	result2, err := uc.Execute(ctx, SendSignInCodeInput{Email: testEmail})
	if err != nil {
		t.Fatalf("Second Execute failed: %v", err)
	}

	// 最新のコードのみが有効であることを確認
	savedCode, err := queries.GetValidSignInCode(ctx, result2.UserID)
	if err != nil {
		t.Fatalf("GetValidSignInCode failed: %v", err)
	}

	// 2回目のコードが保存されていることを確認
	if !auth.VerifyCode(result2.Code, savedCode.CodeDigest) {
		t.Error("Second code verification failed")
	}
}

func TestSendSignInCodeUsecase_Execute_UserNotFound(t *testing.T) {
	t.Parallel()

	// テスト用DBとトランザクションをセットアップ
	db := testutil.GetTestDB(t)
	queries := query.New(db)

	// ユースケースを作成
	v := validator.NewCreateSignInValidator()
	uc := NewSendSignInCodeUsecase(db, repository.NewSignInCodeRepository(queries), repository.NewUserRepository(queries), nil, v)

	ctx := context.Background()

	// 存在しないメールアドレスでExecuteを実行
	result, err := uc.Execute(ctx, SendSignInCodeInput{Email: "nonexistent@example.com"})
	if err != nil {
		t.Fatalf("Execute should not return system error: %v", err)
	}

	// FormErrorsが返されることを確認
	if result.FormErrors == nil || !result.FormErrors.HasErrors() {
		t.Error("expected form errors for non-existent user")
	}
}

func TestSendSignInCodeUsecase_Execute_ValidationError(t *testing.T) {
	t.Parallel()

	// テスト用DBをセットアップ
	db := testutil.GetTestDB(t)
	queries := query.New(db)

	// ユースケースを作成
	v := validator.NewCreateSignInValidator()
	uc := NewSendSignInCodeUsecase(db, repository.NewSignInCodeRepository(queries), repository.NewUserRepository(queries), nil, v)

	ctx := context.Background()

	// 空のメールアドレスでExecuteを実行
	result, err := uc.Execute(ctx, SendSignInCodeInput{Email: ""})
	if err != nil {
		t.Fatalf("Execute should not return system error: %v", err)
	}

	// FormErrorsが返されることを確認
	if result.FormErrors == nil || !result.FormErrors.HasErrors() {
		t.Error("expected form errors for empty email")
	}

	if !result.FormErrors.HasFieldError("email") {
		t.Error("expected email field error")
	}
}

func TestSendSignInCodeUsecase_Execute_UserWithPassword(t *testing.T) {
	t.Parallel()

	// テスト用DBをセットアップ
	db := testutil.GetTestDB(t)
	queries := query.New(db)

	// パスワードありのユーザーを作成
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin transaction failed: %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	testEmail := "sign_in_password_user@example.com"
	testutil.NewUserBuilder(t, setupTx).
		WithUsername("sign_in_password_user").
		WithEmail(testEmail).
		WithEncryptedPassword("encrypted_password_hash").
		Build()

	if err := setupTx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// ユースケースを作成
	v := validator.NewCreateSignInValidator()
	uc := NewSendSignInCodeUsecase(db, repository.NewSignInCodeRepository(queries), repository.NewUserRepository(queries), nil, v)

	ctx := context.Background()

	// パスワードありユーザーのメールアドレスでExecuteを実行
	result, err := uc.Execute(ctx, SendSignInCodeInput{Email: testEmail})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// バリデーションエラーがないことを確認
	if result.FormErrors != nil && result.FormErrors.HasErrors() {
		t.Fatalf("unexpected form errors: %+v", result.FormErrors)
	}

	// HasPasswordがtrueであることを確認
	if !result.HasPassword {
		t.Error("HasPassword should be true for user with password")
	}

	// コードは送信されていないことを確認
	if result.Code != "" {
		t.Errorf("Code should be empty for user with password, got %q", result.Code)
	}

	// メールアドレスが正しいことを確認
	if result.Email != testEmail {
		t.Errorf("Email: got %q, want %q", result.Email, testEmail)
	}
}
