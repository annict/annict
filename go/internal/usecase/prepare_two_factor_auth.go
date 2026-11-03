package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// PrepareTwoFactorAuthUsecaseは、認証アプリへ登録する新しい秘密鍵を作り、登録の途中の設定として保存する。
//
// 登録の画面を開くたびに秘密鍵を作り直す。前に表示した秘密鍵が画面の外に残っていても、有効にする前に使えなくなる。
type PrepareTwoFactorAuthUsecase struct {
	twoFactorKey          *auth.TwoFactorKey
	userTwoFactorAuthRepo *repository.UserTwoFactorAuthRepository
}

// NewPrepareTwoFactorAuthUsecaseはPrepareTwoFactorAuthUsecaseを生成する。
func NewPrepareTwoFactorAuthUsecase(twoFactorKey *auth.TwoFactorKey, userTwoFactorAuthRepo *repository.UserTwoFactorAuthRepository) *PrepareTwoFactorAuthUsecase {
	return &PrepareTwoFactorAuthUsecase{twoFactorKey: twoFactorKey, userTwoFactorAuthRepo: userTwoFactorAuthRepo}
}

// PrepareTwoFactorAuthInputはPrepareTwoFactorAuthUsecase.Executeの入力。
type PrepareTwoFactorAuthInput struct {
	User *model.User
}

// PrepareTwoFactorAuthOutputはPrepareTwoFactorAuthUsecase.Executeの結果。
type PrepareTwoFactorAuthOutput struct {
	// Setupは画面に出す秘密鍵。既に有効にしているときはnil。
	Setup *TwoFactorAuthSetup
}

// Executeは新しい秘密鍵を暗号化して保存し、画面に出す形で返す。
// 既に有効にしているときは、登録済みの認証アプリのコードが通らなくならないよう、秘密鍵を差し替えない。
func (uc *PrepareTwoFactorAuthUsecase) Execute(ctx context.Context, input PrepareTwoFactorAuthInput) (*PrepareTwoFactorAuthOutput, error) {
	secret, err := auth.GenerateTOTPSecret()
	if err != nil {
		return nil, fmt.Errorf("TOTPの秘密鍵の生成に失敗: %w", err)
	}
	setup, err := newTwoFactorAuthSetup(secret, input.User)
	if err != nil {
		return nil, err
	}
	ciphertext, err := uc.twoFactorKey.EncryptTOTPSecret(secret, twoFactorSecretAssociatedData(input.User.ID))
	if err != nil {
		return nil, fmt.Errorf("TOTPの秘密鍵の暗号化に失敗: %w", err)
	}

	pending, err := uc.userTwoFactorAuthRepo.UpsertPending(ctx, input.User.ID, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("登録の途中の二要素認証の設定の保存に失敗: %w", err)
	}
	if pending == nil {
		return &PrepareTwoFactorAuthOutput{}, nil
	}

	return &PrepareTwoFactorAuthOutput{Setup: setup}, nil
}
