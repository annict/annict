package usecase

import (
	"context"
	"errors"

	"github.com/annict/annict/go/internal/repository"
)

// GetPasswordResetTokenUsecase はパスワードリセットトークンの有効性を検証するユースケースです
type GetPasswordResetTokenUsecase struct {
	passwordResetTokenRepo *repository.PasswordResetTokenRepository
}

// NewGetPasswordResetTokenUsecase は新しいGetPasswordResetTokenUsecaseを作成します
func NewGetPasswordResetTokenUsecase(passwordResetTokenRepo *repository.PasswordResetTokenRepository) *GetPasswordResetTokenUsecase {
	return &GetPasswordResetTokenUsecase{
		passwordResetTokenRepo: passwordResetTokenRepo,
	}
}

// GetPasswordResetTokenInput はユースケースの入力パラメータです
type GetPasswordResetTokenInput struct {
	Token string
}

// GetPasswordResetTokenResult はトークン検証の結果を表します
type GetPasswordResetTokenResult struct {
	Valid bool
}

// Execute はトークンの有効性を検証します
func (uc *GetPasswordResetTokenUsecase) Execute(ctx context.Context, input GetPasswordResetTokenInput) (*GetPasswordResetTokenResult, error) {
	if input.Token == "" {
		return &GetPasswordResetTokenResult{Valid: false}, nil
	}

	_, err := uc.passwordResetTokenRepo.GetValidByToken(ctx, input.Token)
	if errors.Is(err, repository.ErrInvalidPasswordResetToken) {
		return &GetPasswordResetTokenResult{Valid: false}, nil
	} else if err != nil {
		return nil, err
	}

	return &GetPasswordResetTokenResult{Valid: true}, nil
}
