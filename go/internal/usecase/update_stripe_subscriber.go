// Package usecase はビジネスロジック層のユースケースを提供します
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

// ErrStripeSubscriberNotFound is returned by the update / delete usecases when no
// StripeSubscriber matches the given Stripe subscription ID. The webhook
// orchestrator detects it with errors.Is and records the event as skipped instead
// of failed, so a webhook for an unknown subscription does not generate Sentry
// noise or endless retries.
//
// [Ja] 指定された Stripe サブスクリプション ID に対応する StripeSubscriber が存在しない
// ときに update / delete ユースケースが返す。Webhook オーケストレーターは errors.Is で
// これを検出し、イベントを failed ではなく skipped として記録する。これにより未知の
// サブスクリプションに対する Webhook が Sentry ノイズや無限リトライを生まない。
var ErrStripeSubscriberNotFound = errors.New("対応するStripeSubscriberが見つかりません")

// UpdateStripeSubscriberUsecase はサブスクリプション更新イベント処理のユースケース
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
	StripeSubscriber model.StripeSubscriber
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
	// Not-found is reported as a sentinel so the webhook layer can skip it without
	// depending on sql.ErrNoRows.
	//
	// [Ja] 未存在は sentinel で返し、Webhook 層が sql.ErrNoRows に依存せずスキップできるようにする。
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
	// The record was just updated above, so a nil here means an unexpected internal
	// inconsistency rather than a normal not-found.
	//
	// [Ja] 直前に更新したレコードのため、ここでの nil は通常の未存在ではなく想定外の
	// 内部不整合を意味する。
	if updated == nil {
		return nil, fmt.Errorf("更新後のStripeSubscriberが見つかりません")
	}

	return &UpdateStripeSubscriberResult{
		StripeSubscriber: *updated,
	}, nil
}
