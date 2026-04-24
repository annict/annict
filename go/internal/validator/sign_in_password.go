package validator

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// SignInPasswordCreateValidator はパスワードログインのバリデーションを行う
type SignInPasswordCreateValidator struct {
	userRepo *repository.UserRepository
}

// NewSignInPasswordCreateValidator は SignInPasswordCreateValidator を生成する
func NewSignInPasswordCreateValidator(userRepo *repository.UserRepository) *SignInPasswordCreateValidator {
	return &SignInPasswordCreateValidator{
		userRepo: userRepo,
	}
}

// SignInPasswordCreateValidatorInput はバリデーションの入力パラメータ
type SignInPasswordCreateValidatorInput struct {
	EmailOrUsername string
	Password        string
}

// SignInPasswordCreateValidateOutput はバリデーション成功時の出力
type SignInPasswordCreateValidateOutput struct {
	User repository.GetUserByEmailOrUsernameRow
}

// Validate はバリデーションを行い、成功時は認証済みユーザー情報を返す
func (v *SignInPasswordCreateValidator) Validate(ctx context.Context, input SignInPasswordCreateValidatorInput) (*SignInPasswordCreateValidateOutput, error) {
	// 1. 形式バリデーション
	ve := model.NewValidationError()

	if strings.TrimSpace(input.EmailOrUsername) == "" {
		ve.AddField("email_or_username", i18n.T(ctx, "sign_in_error_email_required"))
	}

	if strings.TrimSpace(input.Password) == "" {
		ve.AddField("password", i18n.T(ctx, "sign_in_error_password_required"))
	}

	if ve.HasErrors() {
		return nil, ve
	}

	// 2. 状態バリデーション（DB検証）
	user, err := v.userRepo.GetByEmailOrUsername(ctx, input.EmailOrUsername)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ve.AddGlobal(i18n.T(ctx, "sign_in_error_invalid_credentials"))
			return nil, ve
		}
		return nil, err
	}

	// パスワード検証
	if err := auth.CheckPassword(user.EncryptedPassword, input.Password); err != nil {
		ve.AddGlobal(i18n.T(ctx, "sign_in_error_invalid_credentials"))
		return nil, ve
	}

	return &SignInPasswordCreateValidateOutput{User: user}, nil
}
