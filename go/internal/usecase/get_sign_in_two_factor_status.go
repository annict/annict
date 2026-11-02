package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetSignInTwoFactorStatusUsecaseは、ログイン途中のユーザーが二要素認証を続けられるかを取得する。
// 画面を開く段階ではコードを照合せず、タイムステップも消費しない。
type GetSignInTwoFactorStatusUsecase struct {
	userRepo              *repository.UserRepository
	userTwoFactorAuthRepo *repository.UserTwoFactorAuthRepository
}

// NewGetSignInTwoFactorStatusUsecaseはGetSignInTwoFactorStatusUsecaseを生成する。
func NewGetSignInTwoFactorStatusUsecase(
	userRepo *repository.UserRepository,
	userTwoFactorAuthRepo *repository.UserTwoFactorAuthRepository,
) *GetSignInTwoFactorStatusUsecase {
	return &GetSignInTwoFactorStatusUsecase{userRepo: userRepo, userTwoFactorAuthRepo: userTwoFactorAuthRepo}
}

// GetSignInTwoFactorStatusInputはGetSignInTwoFactorStatusUsecase.Executeの入力。
type GetSignInTwoFactorStatusInput struct {
	// UserIDは第1の要素を確かめたユーザーのID。ログイン途中の状態から取得する
	UserID model.UserID
}

// GetSignInTwoFactorStatusOutputはGetSignInTwoFactorStatusUsecase.Executeの結果。
type GetSignInTwoFactorStatusOutput struct {
	// Availableはユーザーが有効で、二要素認証も引き続き有効かどうか
	Available bool
}

// Executeは、ユーザーの退会と二要素認証の無効化・登録の途中を利用不可として返す。
func (uc *GetSignInTwoFactorStatusUsecase) Execute(ctx context.Context, input GetSignInTwoFactorStatusInput) (*GetSignInTwoFactorStatusOutput, error) {
	if _, err := uc.userRepo.GetActiveByID(ctx, input.UserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &GetSignInTwoFactorStatusOutput{}, nil
		}
		return nil, fmt.Errorf("ログイン途中のユーザーの取得に失敗: %w", err)
	}

	enabled, err := isTwoFactorAuthEnabled(ctx, uc.userTwoFactorAuthRepo, input.UserID)
	if err != nil {
		return nil, err
	}
	return &GetSignInTwoFactorStatusOutput{Available: enabled}, nil
}
