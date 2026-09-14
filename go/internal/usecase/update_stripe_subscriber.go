// Package usecaseはビジネスロジック層のユースケースを提供します
package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// ErrStripeSubscriberNotFoundは、指定されたStripeサブスクリプションIDに対応する
// StripeSubscriberが存在しないときにupdate / deleteユースケースが返す。Webhookオーケストレーターはerrors.Isで
// これを検出し、イベントをfailedではなくskippedとして記録する。これにより未知の
// サブスクリプションに対するWebhookがSentryノイズや無限リトライを生まない。
var ErrStripeSubscriberNotFound = errors.New("対応するStripeSubscriberが見つかりません")

// UpdateStripeSubscriberUsecaseはサブスクリプション更新イベント処理のユースケース
type UpdateStripeSubscriberUsecase struct {
	db                   *sql.DB
	stripeSubscriberRepo *repository.StripeSubscriberRepository
	userRepo             *repository.UserRepository
}

// NewUpdateStripeSubscriberUsecaseはUpdateStripeSubscriberUsecaseを作成します
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

// UpdateStripeSubscriberInputはcustomer.subscription.updatedイベントの入力データ
type UpdateStripeSubscriberInput struct {
	StripeSubscriptionID     string    // StripeのサブスクリプションID (sub_xxx)
	StripePriceID            string    // Stripeの価格ID (price_xxx)
	StripeStatus             string    // サブスクリプション状態 (active, canceled, etc.)
	StripeCurrentPeriodStart time.Time // 現在の請求期間開始
	StripeCurrentPeriodEnd   time.Time // 現在の請求期間終了
	StripeCancelAt           sql.NullTime
	StripeCanceledAt         sql.NullTime
}

// UpdateStripeSubscriberResultはcustomer.subscription.updatedイベント処理の結果
type UpdateStripeSubscriberResult struct {
	StripeSubscriber model.StripeSubscriber
}

// Executeはcustomer.subscription.updatedイベントを処理します
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
	// 未存在はsentinelで返し、Webhook層がsql.ErrNoRowsに依存せずスキップできるようにする。
	if subscriber == nil {
		return nil, ErrStripeSubscriberNotFound
	}

	// サブスクリプション情報を更新
	err = uc.stripeSubscriberRepo.Update(ctx, repository.UpdateStripeSubscriberParams{
		ID:                       int64(subscriber.ID),
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
	// 直前に更新したレコードのため、ここでのnilは通常の未存在ではなく想定外の
	// 内部不整合を意味する。
	if updated == nil {
		return nil, fmt.Errorf("更新後のStripeSubscriberが見つかりません")
	}

	return &UpdateStripeSubscriberResult{
		StripeSubscriber: *updated,
	}, nil
}
