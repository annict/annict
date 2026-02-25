package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// DeleteStripeSubscriberUsecase はサブスクリプション削除イベント処理のユースケース
type DeleteStripeSubscriberUsecase struct {
	db                   *sql.DB
	stripeSubscriberRepo *repository.StripeSubscriberRepository
	userRepo             *repository.UserRepository
}

// NewDeleteStripeSubscriberUsecase はDeleteStripeSubscriberUsecaseを作成します
func NewDeleteStripeSubscriberUsecase(
	db *sql.DB,
	stripeSubscriberRepo *repository.StripeSubscriberRepository,
	userRepo *repository.UserRepository,
) *DeleteStripeSubscriberUsecase {
	return &DeleteStripeSubscriberUsecase{
		db:                   db,
		stripeSubscriberRepo: stripeSubscriberRepo,
		userRepo:             userRepo,
	}
}

// DeleteStripeSubscriberInput はcustomer.subscription.deletedイベントの入力データ
type DeleteStripeSubscriberInput struct {
	StripeSubscriptionID string    // StripeのサブスクリプションID (sub_xxx)
	StripeCanceledAt     time.Time // キャンセル日時
}

// DeleteStripeSubscriberResult はcustomer.subscription.deletedイベント処理の結果
type DeleteStripeSubscriberResult struct {
	StripeSubscriber repository.StripeSubscriber
	UserID           *int64 // 紐付け解除されたユーザーID（存在する場合）
}

// Execute はcustomer.subscription.deletedイベントを処理します
//
// 処理フロー:
// 1. StripeサブスクリプションIDで既存レコードを検索
// 2. ステータスをcanceledに更新
// 3. Userとの紐付けを解除
func (uc *DeleteStripeSubscriberUsecase) Execute(
	ctx context.Context,
	input DeleteStripeSubscriberInput,
) (*DeleteStripeSubscriberResult, error) {
	// 既存のStripeSubscriberを取得
	subscriber, err := uc.stripeSubscriberRepo.GetByStripeSubscriptionID(ctx, input.StripeSubscriptionID)
	if err != nil {
		return nil, fmt.Errorf("StripeSubscriber取得に失敗: %w", err)
	}

	// トランザクション開始
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// トランザクションを使用するRepositoryを取得
	stripeSubscriberRepoTx := uc.stripeSubscriberRepo.WithTx(tx)
	userRepoTx := uc.userRepo.WithTx(tx)

	// ステータスをcanceledに更新
	err = stripeSubscriberRepoTx.Update(ctx, repository.UpdateStripeSubscriberParams{
		ID:                       subscriber.ID,
		StripePriceID:            subscriber.StripePriceID,
		StripeStatus:             string(model.StripeSubscriptionStatusCanceled),
		StripeCurrentPeriodStart: subscriber.StripeCurrentPeriodStart,
		StripeCurrentPeriodEnd:   subscriber.StripeCurrentPeriodEnd,
		StripeCancelAt:           subscriber.StripeCancelAt,
		StripeCanceledAt: sql.NullTime{
			Time:  input.StripeCanceledAt,
			Valid: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("StripeSubscriberステータス更新に失敗: %w", err)
	}

	// Userとの紐付けを解除
	// StripeSubscriberIDが一致するユーザーを探してnilに設定
	userID, err := userRepoTx.FindUserIDByStripeSubscriberID(ctx, subscriber.ID)
	if err != nil {
		return nil, fmt.Errorf("ユーザー検索に失敗: %w", err)
	}

	if userID != nil {
		// 紐付けを解除
		err = userRepoTx.UpdateStripeSubscriberID(ctx, *userID, nil)
		if err != nil {
			return nil, fmt.Errorf("ユーザー紐付け解除に失敗: %w", err)
		}
	}

	// トランザクションコミット
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションコミットに失敗: %w", err)
	}

	// 更新後のレコードを取得
	updated, err := uc.stripeSubscriberRepo.GetByID(ctx, subscriber.ID)
	if err != nil {
		return nil, fmt.Errorf("更新後のStripeSubscriber取得に失敗: %w", err)
	}

	return &DeleteStripeSubscriberResult{
		StripeSubscriber: updated,
		UserID:           userID,
	}, nil
}
