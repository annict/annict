package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/validator"
)

// AuthenticateByPasswordUsecaseはパスワードによる認証を担当するユースケースです
type AuthenticateByPasswordUsecase struct {
	createSessionUC       *CreateSessionUsecase
	validator             *validator.SignInPasswordCreateValidator
	userTwoFactorAuthRepo *repository.UserTwoFactorAuthRepository
}

// NewAuthenticateByPasswordUsecaseは新しいAuthenticateByPasswordUsecaseを作成します
func NewAuthenticateByPasswordUsecase(
	createSessionUC *CreateSessionUsecase,
	validator *validator.SignInPasswordCreateValidator,
	userTwoFactorAuthRepo *repository.UserTwoFactorAuthRepository,
) *AuthenticateByPasswordUsecase {
	return &AuthenticateByPasswordUsecase{
		createSessionUC:       createSessionUC,
		validator:             validator,
		userTwoFactorAuthRepo: userTwoFactorAuthRepo,
	}
}

// AuthenticateByPasswordInputはユースケースの入力パラメータです
type AuthenticateByPasswordInput struct {
	Email    string
	Password string
}

// AuthenticateByPasswordOutputはユースケースの結果を表します
type AuthenticateByPasswordOutput struct {
	PublicID string       // セッションのPublicID。TwoFactorAuthRequiredのときは空
	UserID   model.UserID // ユーザーID
	Username string       // ユーザー名
	// TwoFactorAuthRequiredは、セッションを発行する前に認証アプリのコードを求めるかどうか。
	// trueのときはセッションを発行していない。
	TwoFactorAuthRequired bool
}

// Executeはパスワード認証を行い、セッションを作成します。
// 二要素認証を有効にしているユーザーには、セッションを作らずTwoFactorAuthRequiredを返します。
func (uc *AuthenticateByPasswordUsecase) Execute(ctx context.Context, input AuthenticateByPasswordInput) (*AuthenticateByPasswordOutput, error) {
	// 1. バリデーション (形式チェック + 存在確認 + パスワード照合)
	valOutput, err := uc.validator.Validate(ctx, validator.SignInPasswordCreateValidatorInput{
		EmailOrUsername: input.Email,
		Password:        input.Password,
	})
	if err != nil {
		return nil, err
	}
	user := valOutput.User

	// 2. NOT NULL制約があるフィールドの検証 (fail-fast)
	if err := validator.ValidateNotNullTime(user.UpdatedAt, "updated_at", user.ID); err != nil {
		slog.ErrorContext(ctx, "データベース制約違反を検出しました (NOT NULL制約)",
			"table", "users",
			"field", "updated_at",
			"user_id", user.ID,
			"error", err,
		)
		return nil, fmt.Errorf("データベース制約違反: %w", err)
	}

	// 3. 二要素認証を有効にしているときは、認証アプリのコードを確かめるまでセッションを作らない
	twoFactorAuthRequired, err := isTwoFactorAuthEnabled(ctx, uc.userTwoFactorAuthRepo, model.UserID(user.ID))
	if err != nil {
		return nil, err
	}
	if twoFactorAuthRequired {
		slog.InfoContext(ctx, "パスワードを確かめたため、認証アプリのコードを求めます", "user_id", user.ID)
		return &AuthenticateByPasswordOutput{
			UserID:                model.UserID(user.ID),
			Username:              user.Username,
			TwoFactorAuthRequired: true,
		}, nil
	}

	// 4. セッション作成
	sessionResult, err := uc.createSessionUC.Execute(ctx, nil, model.UserID(user.ID), user.EncryptedPassword)
	if err != nil {
		return nil, fmt.Errorf("セッションの作成に失敗: %w", err)
	}

	slog.InfoContext(ctx, "ログイン成功", "user_id", user.ID, "username", user.Username)

	return &AuthenticateByPasswordOutput{
		PublicID: sessionResult.PublicID,
		UserID:   model.UserID(user.ID),
		Username: user.Username,
	}, nil
}
