package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetTwoFactorAuthStatusUsecaseは、ユーザーが二要素認証を有効にしているかと、残りのリカバリーコードの数を取得する。
type GetTwoFactorAuthStatusUsecase struct {
	userTwoFactorAuthRepo         *repository.UserTwoFactorAuthRepository
	userTwoFactorRecoveryCodeRepo *repository.UserTwoFactorRecoveryCodeRepository
}

// NewGetTwoFactorAuthStatusUsecaseはGetTwoFactorAuthStatusUsecaseを生成する。
func NewGetTwoFactorAuthStatusUsecase(
	userTwoFactorAuthRepo *repository.UserTwoFactorAuthRepository,
	userTwoFactorRecoveryCodeRepo *repository.UserTwoFactorRecoveryCodeRepository,
) *GetTwoFactorAuthStatusUsecase {
	return &GetTwoFactorAuthStatusUsecase{
		userTwoFactorAuthRepo:         userTwoFactorAuthRepo,
		userTwoFactorRecoveryCodeRepo: userTwoFactorRecoveryCodeRepo,
	}
}

// GetTwoFactorAuthStatusInputはGetTwoFactorAuthStatusUsecase.Executeの入力。
type GetTwoFactorAuthStatusInput struct {
	UserID model.UserID
}

// GetTwoFactorAuthStatusOutputはGetTwoFactorAuthStatusUsecase.Executeの結果。
type GetTwoFactorAuthStatusOutput struct {
	// Enabledは二要素認証を有効にしているか。登録の途中はfalse。
	Enabled bool
	// UnusedRecoveryCodeCountはまだ使っていないリカバリーコードの数。無効のときは0。
	UnusedRecoveryCodeCount int
}

// Executeは二要素認証の状態を返す。
// 登録の途中の設定は、ログインでコードを求めないため無効として扱う。
func (uc *GetTwoFactorAuthStatusUsecase) Execute(ctx context.Context, input GetTwoFactorAuthStatusInput) (*GetTwoFactorAuthStatusOutput, error) {
	twoFactorAuth, err := uc.userTwoFactorAuthRepo.FindByUserID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("二要素認証の設定の取得に失敗: %w", err)
	}
	if twoFactorAuth == nil || !twoFactorAuth.IsEnabled() {
		return &GetTwoFactorAuthStatusOutput{}, nil
	}

	count, err := uc.userTwoFactorRecoveryCodeRepo.CountUnused(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("リカバリーコードの数の取得に失敗: %w", err)
	}

	return &GetTwoFactorAuthStatusOutput{Enabled: true, UnusedRecoveryCodeCount: count}, nil
}
