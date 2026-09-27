package usecase

import (
	"context"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

func TestAuthenticateByPasswordUsecase_Execute_Success(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	queries := query.New(db)

	// テストユーザーを作成してコミット
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("トランザクションのBeginのエラー = %v", err)
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
		t.Fatalf("Commit()のエラー = %v", err)
	}

	// ユースケースを作成
	userRepo := repository.NewUserRepository(queries)
	sessionRepo := repository.NewSessionRepository(queries)
	createSessionUC := NewCreateSessionUsecase(sessionRepo)
	v := validator.NewSignInPasswordCreateValidator(userRepo)
	uc := NewAuthenticateByPasswordUsecase(createSessionUC, v)

	ctx := context.Background()

	result, err := uc.Execute(ctx, AuthenticateByPasswordInput{
		Email:    "auth_pw_success@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Executeのエラー = %v", err)
	}

	if result.PublicID == "" {
		t.Error("PublicIDが空だった")
	}

	if result.Username != "auth_pw_success" {
		t.Errorf("Username = %q、期待値 = %q", result.Username, "auth_pw_success")
	}
}

func TestAuthenticateByPasswordUsecase_Execute_InvalidPassword(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	queries := query.New(db)

	// テストユーザーを作成してコミット
	setupTx, err := db.Begin()
	if err != nil {
		t.Fatalf("トランザクションのBeginのエラー = %v", err)
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
		t.Fatalf("Commit()のエラー = %v", err)
	}

	userRepo := repository.NewUserRepository(queries)
	sessionRepo := repository.NewSessionRepository(queries)
	createSessionUC := NewCreateSessionUsecase(sessionRepo)
	v := validator.NewSignInPasswordCreateValidator(userRepo)
	uc := NewAuthenticateByPasswordUsecase(createSessionUC, v)

	ctx := context.Background()

	result, err := uc.Execute(ctx, AuthenticateByPasswordInput{
		Email:    "auth_pw_invalid@example.com",
		Password: "wrongpassword",
	})
	ve := model.AsValidationError(err)
	if ve == nil {
		t.Fatalf("passwordが不正なときのエラー = %v、期待値 = バリデーションエラー", err)
	}

	if result != nil {
		t.Error("失敗時のresultがnilでなかった")
	}
}

func TestAuthenticateByPasswordUsecase_Execute_UserNotFound(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	queries := query.New(db)

	userRepo := repository.NewUserRepository(queries)
	sessionRepo := repository.NewSessionRepository(queries)
	createSessionUC := NewCreateSessionUsecase(sessionRepo)
	v := validator.NewSignInPasswordCreateValidator(userRepo)
	uc := NewAuthenticateByPasswordUsecase(createSessionUC, v)

	ctx := context.Background()

	_, err := uc.Execute(ctx, AuthenticateByPasswordInput{
		Email:    "nonexistent@example.com",
		Password: "password123",
	})
	ve := model.AsValidationError(err)
	if ve == nil {
		t.Fatalf("存在しないユーザーでのエラー = %v、期待値 = バリデーションエラー", err)
	}
}

func TestAuthenticateByPasswordUsecase_Execute_ValidationError(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	queries := query.New(db)

	userRepo := repository.NewUserRepository(queries)
	sessionRepo := repository.NewSessionRepository(queries)
	createSessionUC := NewCreateSessionUsecase(sessionRepo)
	v := validator.NewSignInPasswordCreateValidator(userRepo)
	uc := NewAuthenticateByPasswordUsecase(createSessionUC, v)

	ctx := context.Background()

	// 空のパスワードでバリデーションエラー
	_, err := uc.Execute(ctx, AuthenticateByPasswordInput{
		Email:    "test@example.com",
		Password: "",
	})
	ve := model.AsValidationError(err)
	if ve == nil {
		t.Fatalf("passwordが空のときのエラー = %v、期待値 = バリデーションエラー", err)
	}

	if !ve.HasFieldError("password") {
		t.Error("passwordフィールドのエラーを期待したが、無かった")
	}
}

func TestAuthenticateByPasswordUsecase_Execute_WhitespacePassword(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	queries := query.New(db)

	userRepo := repository.NewUserRepository(queries)
	sessionRepo := repository.NewSessionRepository(queries)
	createSessionUC := NewCreateSessionUsecase(sessionRepo)
	v := validator.NewSignInPasswordCreateValidator(userRepo)
	uc := NewAuthenticateByPasswordUsecase(createSessionUC, v)

	ctx := context.Background()

	// スペースのみのパスワードでバリデーションエラー
	_, err := uc.Execute(ctx, AuthenticateByPasswordInput{
		Email:    "test@example.com",
		Password: "   ",
	})
	ve := model.AsValidationError(err)
	if ve == nil {
		t.Fatalf("passwordが空白のみのときのエラー = %v、期待値 = バリデーションエラー", err)
	}

	if !ve.HasFieldError("password") {
		t.Error("passwordフィールドのエラーを期待したが、無かった")
	}
}
