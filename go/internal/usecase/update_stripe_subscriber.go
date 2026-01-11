// Package usecase はビジネスロジック層のユースケースを提供します
package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
)

// UpdateStripeSubscriberUsecase はサブスクリプション更新/削除イベント処理のユースケース
type UpdateStripeSubscriberUsecase struct {
	db                   *sql.DB
	stripeSubscriberRepo *repository.StripeSubscriberRepository
	userRepo             *repository.UserRepository
}

// NewUpdateStripeSubscriberUsecase はUpdateStripeSubscriberUsecaseを作成します
func NewUpdateStripeSubscriberUsecase(
	db *sql.DB,
	stripeSubscriberRepo *repository.StripeSubscriberRepository,
	userRepo *repository.UserRepository,
) *UpdateStripeSubscriberUsecase {
	return &UpdateStripeSubscriberUsecase{
		db:                   db,
		stripeSubscriberRepo: stripeSubscriberRepo,
		userRepo:             userRepo,
	}
}

// UpdateStripeSubscriberInput はcustomer.subscription.updatedイベントの入力データ
type UpdateStripeSubscriberInput struct {
	StripeSubscriptionID     string    // StripeのサブスクリプションID (sub_xxx)
	StripePriceID            string    // Stripeの価格ID (price_xxx)
	StripeStatus             string    // サブスクリプション状態 (active, canceled, etc.)
	StripeCurrentPeriodStart time.Time // 現在の請求期間開始
	StripeCurrentPeriodEnd   time.Time // 現在の請求期間終了
	StripeCancelAt           sql.NullTime
	StripeCanceledAt         sql.NullTime
}

// UpdateStripeSubscriberResult はcustomer.subscription.updatedイベント処理の結果
type UpdateStripeSubscriberResult struct {
	StripeSubscriber query.StripeSubscriber
}

// Execute はcustomer.subscription.updatedイベントを処理します
//
// 処理フロー:
// 1. StripeサブスクリプションIDで既存レコードを検索
// 2. サブスクリプション情報を更新
func (uc *UpdateStripeSubscriberUsecase) Execute(
	ctx context.Context,
	input UpdateStripeSubscriberInput,
) (*UpdateStripeSubscriberResult, error) {
	// ステータスを検証
	status := model.StripeSubscriptionStatus(input.StripeStatus)
	if !status.IsValid() {
		return nil, &InvalidSubscriptionStatusError{Status: input.StripeStatus}
	}

	// 既存のStripeSubscriberを取得
	subscriber, err := uc.stripeSubscriberRepo.GetByStripeSubscriptionID(ctx, input.StripeSubscriptionID)
	if err != nil {
		return nil, fmt.Errorf("StripeSubscriber取得に失敗: %w", err)
	}

	// サブスクリプション情報を更新
	err = uc.stripeSubscriberRepo.Update(ctx, query.UpdateStripeSubscriberParams{
		ID:                       subscriber.ID,
		StripePriceID:            input.StripePriceID,
		StripeStatus:             input.StripeStatus,
		StripeCurrentPeriodStart: input.StripeCurrentPeriodStart,
		StripeCurrentPeriodEnd:   input.StripeCurrentPeriodEnd,
		StripeCancelAt:           input.StripeCancelAt,
		StripeCanceledAt:         input.StripeCanceledAt,
	})
	if err != nil {
		return nil, fmt.Errorf("StripeSubscriber更新に失敗: %w", err)
	}

	// 更新後のレコードを取得
	updated, err := uc.stripeSubscriberRepo.GetByID(ctx, subscriber.ID)
	if err != nil {
		return nil, fmt.Errorf("更新後のStripeSubscriber取得に失敗: %w", err)
	}

	return &UpdateStripeSubscriberResult{
		StripeSubscriber: updated,
	}, nil
}

// DeleteStripeSubscriberInput はcustomer.subscription.deletedイベントの入力データ
type DeleteStripeSubscriberInput struct {
	StripeSubscriptionID string    // StripeのサブスクリプションID (sub_xxx)
	StripeCanceledAt     time.Time // キャンセル日時
}

// DeleteStripeSubscriberResult はcustomer.subscription.deletedイベント処理の結果
type DeleteStripeSubscriberResult struct {
	StripeSubscriber query.StripeSubscriber
	UserID           *int64 // 紐付け解除されたユーザーID（存在する場合）
}

// ExecuteDelete はcustomer.subscription.deletedイベントを処理します
//
// 処理フロー:
// 1. StripeサブスクリプションIDで既存レコードを検索
// 2. ステータスをcanceledに更新
// 3. Userとの紐付けを解除
func (uc *UpdateStripeSubscriberUsecase) ExecuteDelete(
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

	// ステータスをcanceledに更新
	err = uc.stripeSubscriberRepo.Update(ctx, query.UpdateStripeSubscriberParams{
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
	userID, err := uc.userRepo.FindUserIDByStripeSubscriberID(ctx, subscriber.ID)
	if err != nil {
		return nil, fmt.Errorf("ユーザー検索に失敗: %w", err)
	}

	if userID != nil {
		// 紐付けを解除
		err = uc.userRepo.UpdateStripeSubscriberID(ctx, *userID, nil)
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
