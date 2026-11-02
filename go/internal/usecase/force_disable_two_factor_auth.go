package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// ForceDisableTwoFactorAuthUsecaseは、本人確認をせずに二要素認証を無効にし、設定とリカバリーコードを削除する。
//
// 認証アプリとリカバリーコードの両方を失ったユーザーを、管理者が救済するためのもの。
// 本人確認はこのUseCaseの外で管理者が済ませる前提で、Web画面からは呼ばない。
type ForceDisableTwoFactorAuthUsecase struct {
	db                            *sql.DB
	userTwoFactorAuthRepo         *repository.UserTwoFactorAuthRepository
	userTwoFactorRecoveryCodeRepo *repository.UserTwoFactorRecoveryCodeRepository
	userRepo                      *repository.UserRepository
	dispatcher                    *dispatcher.Dispatcher
}

// NewForceDisableTwoFactorAuthUsecaseはForceDisableTwoFactorAuthUsecaseを生成する。
func NewForceDisableTwoFactorAuthUsecase(
	db *sql.DB,
	userTwoFactorAuthRepo *repository.UserTwoFactorAuthRepository,
	userTwoFactorRecoveryCodeRepo *repository.UserTwoFactorRecoveryCodeRepository,
	userRepo *repository.UserRepository,
	dispatcher *dispatcher.Dispatcher,
) *ForceDisableTwoFactorAuthUsecase {
	return &ForceDisableTwoFactorAuthUsecase{
		db:                            db,
		userTwoFactorAuthRepo:         userTwoFactorAuthRepo,
		userTwoFactorRecoveryCodeRepo: userTwoFactorRecoveryCodeRepo,
		userRepo:                      userRepo,
		dispatcher:                    dispatcher,
	}
}

// ForceDisableTwoFactorAuthInputはForceDisableTwoFactorAuthUsecase.Executeの入力。
type ForceDisableTwoFactorAuthInput struct {
	UserID model.UserID
}

// ForceDisableTwoFactorAuthOutputはForceDisableTwoFactorAuthUsecase.Executeの結果。
type ForceDisableTwoFactorAuthOutput struct {
	// Disabledは有効だった二要素認証を無効にしたか。有効にしていなかったときはfalse
	Disabled bool
}

// Executeは二要素認証の設定とリカバリーコードを1つのトランザクションで削除する。
// 登録の途中の設定も削除するが、ログインでコードを求めていなかったため、無効にしたとは数えない。
func (uc *ForceDisableTwoFactorAuthUsecase) Execute(ctx context.Context, input ForceDisableTwoFactorAuthInput) (*ForceDisableTwoFactorAuthOutput, error) {
	deleted, err := uc.delete(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	disabled := deleted != nil && deleted.IsEnabled()
	// 登録の途中の設定を消しただけのときは、ログインでコードを求めていなかったため知らせない。
	if disabled {
		enqueueTwoFactorAuthNotification(ctx, uc.dispatcher, uc.userRepo, input.UserID, model.TwoFactorAuthNotificationEventDisabledByAdmin)
	}

	return &ForceDisableTwoFactorAuthOutput{Disabled: disabled}, nil
}

// deleteは二要素認証の設定とリカバリーコードを削除し、実際に削除した設定を返す。
// 先に状態を読まないのは、ロック待ちの間に有効になった設定も、その削除時点の状態で判定するため。
func (uc *ForceDisableTwoFactorAuthUsecase) delete(ctx context.Context, userID model.UserID) (*model.UserTwoFactorAuth, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクションの開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	deleted, err := uc.userTwoFactorAuthRepo.WithTx(tx).DeleteByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("二要素認証の設定の削除に失敗: %w", err)
	}
	if err := uc.userTwoFactorRecoveryCodeRepo.WithTx(tx).DeleteByUserID(ctx, userID); err != nil {
		return nil, fmt.Errorf("リカバリーコードの削除に失敗: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	return deleted, nil
}
