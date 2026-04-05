package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/validator"
)

// AuthenticateByPasswordUsecase はパスワードによる認証を担当するユースケースです
type AuthenticateByPasswordUsecase struct {
	userRepo        *repository.UserRepository
	createSessionUC *CreateSessionUsecase
	v               *validator.CreateSignInPasswordValidator
}

// NewAuthenticateByPasswordUsecase は新しいAuthenticateByPasswordUsecaseを作成します
func NewAuthenticateByPasswordUsecase(
	userRepo *repository.UserRepository,
	createSessionUC *CreateSessionUsecase,
	v *validator.CreateSignInPasswordValidator,
) *AuthenticateByPasswordUsecase {
	return &AuthenticateByPasswordUsecase{
		userRepo:        userRepo,
		createSessionUC: createSessionUC,
		v:               v,
	}
}

// AuthenticateByPasswordInput はユースケースの入力パラメータです
type AuthenticateByPasswordInput struct {
	Email    string
	Password string
}

// AuthenticateByPasswordResult はユースケースの結果を表します
type AuthenticateByPasswordResult struct {
	FormErrors *session.FormErrors // バリデーションエラー（nilなら成功）
	PublicID   string              // セッションのPublicID（成功時）
	UserID     int64               // ユーザーID（成功時）
	Username   string              // ユーザー名（成功時）
}

// Execute はパスワード認証を行い、セッションを作成します
func (uc *AuthenticateByPasswordUsecase) Execute(ctx context.Context, input AuthenticateByPasswordInput) (*AuthenticateByPasswordResult, error) {
	// 1. バリデーション
	valResult := uc.v.Validate(ctx, validator.CreateSignInPasswordValidatorInput{
		Password: input.Password,
	})
	if valResult.FormErrors != nil && valResult.FormErrors.HasErrors() {
		return &AuthenticateByPasswordResult{FormErrors: valResult.FormErrors}, nil
	}

	// 2. ユーザー検索
	user, err := uc.userRepo.GetByEmailOrUsername(ctx, input.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.InfoContext(ctx, "ユーザーが見つかりません", "email", input.Email)
			formErrors := &session.FormErrors{}
			formErrors.AddGlobalError(i18n.T(ctx, "sign_in_error_invalid_credentials"))
			return &AuthenticateByPasswordResult{FormErrors: formErrors}, nil
		}
		return nil, fmt.Errorf("ユーザーの検索に失敗: %w", err)
	}

	// 3. NOT NULL制約があるフィールドの検証（fail-fast）
	if err := validator.ValidateNotNullTime(user.UpdatedAt, "updated_at", user.ID); err != nil {
		slog.ErrorContext(ctx, "データベース制約違反を検出しました（NOT NULL制約）",
			"table", "users",
			"field", "updated_at",
			"user_id", user.ID,
			"error", err,
		)
		return nil, fmt.Errorf("データベース制約違反: %w", err)
	}

	// 4. パスワード検証
	if err := auth.CheckPassword(user.EncryptedPassword, input.Password); err != nil {
		slog.InfoContext(ctx, "パスワードが一致しません", "email", input.Email)
		formErrors := &session.FormErrors{}
		formErrors.AddGlobalError(i18n.T(ctx, "sign_in_error_invalid_credentials"))
		return &AuthenticateByPasswordResult{FormErrors: formErrors}, nil
	}

	// 5. セッション作成
	sessionResult, err := uc.createSessionUC.Execute(ctx, nil, user.ID, user.EncryptedPassword)
	if err != nil {
		return nil, fmt.Errorf("セッションの作成に失敗: %w", err)
	}

	slog.InfoContext(ctx, "ログイン成功", "user_id", user.ID, "username", user.Username)

	return &AuthenticateByPasswordResult{
		PublicID: sessionResult.PublicID,
		UserID:   user.ID,
		Username: user.Username,
	}, nil
}
