package usecase

import (
	"context"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

func TestAuthenticateByPasswordUsecase_Execute_Success(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB(t)
	queries := query.New(db)

	// テストユーザーを作成してコミット
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin transaction failed: %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("パスワードハッシュ生成エラー: %v", err)
	}

	testutil.NewUserBuilder(t, setupTx).
		WithUsername("auth_pw_success").
		WithEmail("auth_pw_success@example.com").
		WithEncryptedPassword(string(hashedPassword)).
		Build()

	if err := setupTx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// ユースケースを作成
	userRepo := repository.NewUserRepository(queries)
	sessionRepo := repository.NewSessionRepository(queries)
	createSessionUC := NewCreateSessionUsecase(sessionRepo)
	v := validator.NewCreateSignInPasswordValidator()
	uc := NewAuthenticateByPasswordUsecase(userRepo, createSessionUC, v)

	ctx := context.Background()

	result, err := uc.Execute(ctx, AuthenticateByPasswordInput{
		Email:    "auth_pw_success@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.FormErrors != nil && result.FormErrors.HasErrors() {
		t.Fatalf("unexpected form errors: %+v", result.FormErrors)
	}

	if result.PublicID == "" {
		t.Error("PublicID should not be empty")
	}

	if result.Username != "auth_pw_success" {
		t.Errorf("Username: got %q, want %q", result.Username, "auth_pw_success")
	}
}

func TestAuthenticateByPasswordUsecase_Execute_InvalidPassword(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB(t)
	queries := query.New(db)

	// テストユーザーを作成してコミット
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin transaction failed: %v", err)
	}
	defer func() { _ = setupTx.Rollback() }()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("パスワードハッシュ生成エラー: %v", err)
	}

	testutil.NewUserBuilder(t, setupTx).
		WithUsername("auth_pw_invalid").
		WithEmail("auth_pw_invalid@example.com").
		WithEncryptedPassword(string(hashedPassword)).
		Build()

	if err := setupTx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	userRepo := repository.NewUserRepository(queries)
	sessionRepo := repository.NewSessionRepository(queries)
	createSessionUC := NewCreateSessionUsecase(sessionRepo)
	v := validator.NewCreateSignInPasswordValidator()
	uc := NewAuthenticateByPasswordUsecase(userRepo, createSessionUC, v)

	ctx := context.Background()

	result, err := uc.Execute(ctx, AuthenticateByPasswordInput{
		Email:    "auth_pw_invalid@example.com",
		Password: "wrongpassword",
	})
	if err != nil {
		t.Fatalf("Execute should not return system error: %v", err)
	}

	if result.FormErrors == nil || !result.FormErrors.HasErrors() {
		t.Error("expected form errors for invalid password")
	}

	if result.PublicID != "" {
		t.Error("PublicID should be empty on failure")
	}
}

func TestAuthenticateByPasswordUsecase_Execute_UserNotFound(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB(t)
	queries := query.New(db)

	userRepo := repository.NewUserRepository(queries)
	sessionRepo := repository.NewSessionRepository(queries)
	createSessionUC := NewCreateSessionUsecase(sessionRepo)
	v := validator.NewCreateSignInPasswordValidator()
	uc := NewAuthenticateByPasswordUsecase(userRepo, createSessionUC, v)

	ctx := context.Background()

	result, err := uc.Execute(ctx, AuthenticateByPasswordInput{
		Email:    "nonexistent@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Execute should not return system error: %v", err)
	}

	if result.FormErrors == nil || !result.FormErrors.HasErrors() {
		t.Error("expected form errors for non-existent user")
	}
}

func TestAuthenticateByPasswordUsecase_Execute_ValidationError(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB(t)
	queries := query.New(db)

	userRepo := repository.NewUserRepository(queries)
	sessionRepo := repository.NewSessionRepository(queries)
	createSessionUC := NewCreateSessionUsecase(sessionRepo)
	v := validator.NewCreateSignInPasswordValidator()
	uc := NewAuthenticateByPasswordUsecase(userRepo, createSessionUC, v)

	ctx := context.Background()

	// 空のパスワードでバリデーションエラー
	result, err := uc.Execute(ctx, AuthenticateByPasswordInput{
		Email:    "test@example.com",
		Password: "",
	})
	if err != nil {
		t.Fatalf("Execute should not return system error: %v", err)
	}

	if result.FormErrors == nil || !result.FormErrors.HasErrors() {
		t.Error("expected form errors for empty password")
	}

	if !result.FormErrors.HasFieldError("password") {
		t.Error("expected password field error")
	}
}

func TestAuthenticateByPasswordUsecase_Execute_WhitespacePassword(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB(t)
	queries := query.New(db)

	userRepo := repository.NewUserRepository(queries)
	sessionRepo := repository.NewSessionRepository(queries)
	createSessionUC := NewCreateSessionUsecase(sessionRepo)
	v := validator.NewCreateSignInPasswordValidator()
	uc := NewAuthenticateByPasswordUsecase(userRepo, createSessionUC, v)

	ctx := context.Background()

	// スペースのみのパスワードでバリデーションエラー
	result, err := uc.Execute(ctx, AuthenticateByPasswordInput{
		Email:    "test@example.com",
		Password: "   ",
	})
	if err != nil {
		t.Fatalf("Execute should not return system error: %v", err)
	}

	if result.FormErrors == nil || !result.FormErrors.HasErrors() {
		t.Error("expected form errors for whitespace-only password")
	}

	if !result.FormErrors.HasFieldError("password") {
		t.Error("expected password field error")
	}
}
