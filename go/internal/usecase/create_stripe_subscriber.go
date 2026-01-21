// Package usecase はビジネスロジック層のユースケースを提供します
package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/subscription"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	annictstripe "github.com/annict/annict/go/internal/stripe"
)

// CreateStripeSubscriberUsecase はcheckout.session.completedイベント処理のユースケース
type CreateStripeSubscriberUsecase struct {
	db                   *sql.DB
	stripeSubscriberRepo *repository.StripeSubscriberRepository
	userRepo             *repository.UserRepository
	stripeClient         *stripe.Client
}

// NewCreateStripeSubscriberUsecase はCreateStripeSubscriberUsecaseを作成します
func NewCreateStripeSubscriberUsecase(
	db *sql.DB,
	stripeSubscriberRepo *repository.StripeSubscriberRepository,
	userRepo *repository.UserRepository,
	stripeClient *stripe.Client,
) *CreateStripeSubscriberUsecase {
	return &CreateStripeSubscriberUsecase{
		db:                   db,
		stripeSubscriberRepo: stripeSubscriberRepo,
		userRepo:             userRepo,
		stripeClient:         stripeClient,
	}
}

// CreateStripeSubscriberInput はcheckout.session.completedイベントの入力データ
type CreateStripeSubscriberInput struct {
	StripeCustomerID     string // Stripeの顧客ID (cus_xxx)
	StripeSubscriptionID string // StripeのサブスクリプションID (sub_xxx)
	UserID               int64  // AnnictのユーザーID（metadataから取得）
}

// CreateStripeSubscriberResult はcheckout.session.completedイベント処理の結果
type CreateStripeSubscriberResult struct {
	StripeSubscriber query.StripeSubscriber
}

// Execute はcheckout.session.completedイベントを処理します
//
// 処理フロー:
// 1. Stripe APIからサブスクリプション詳細を取得
// 2. StripeSubscriberレコードを作成
// 3. Userとの紐付け
func (uc *CreateStripeSubscriberUsecase) Execute(
	ctx context.Context,
	input CreateStripeSubscriberInput,
) (*CreateStripeSubscriberResult, error) {
	// Stripe APIからサブスクリプション詳細を取得
	sub, err := subscription.Get(input.StripeSubscriptionID, nil)
	if err != nil {
		return nil, fmt.Errorf("サブスクリプション取得に失敗: %w", err)
	}

	// ステータスを検証
	status := model.StripeSubscriptionStatus(sub.Status)
	if !status.IsValid() {
		return nil, &InvalidSubscriptionStatusError{Status: string(sub.Status)}
	}

	// 価格IDと請求期間を取得（最初のアイテムから）
	if len(sub.Items.Data) == 0 {
		return nil, fmt.Errorf("サブスクリプションにアイテムが含まれていません")
	}
	item := sub.Items.Data[0]
	priceID := item.Price.ID

	// トランザクション開始
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// StripeSubscriberレコードを作成
	stripeSubscriber, err := uc.stripeSubscriberRepo.Create(ctx, query.CreateStripeSubscriberParams{
		StripeCustomerID:         input.StripeCustomerID,
		StripeSubscriptionID:     input.StripeSubscriptionID,
		StripePriceID:            priceID,
		StripeStatus:             string(sub.Status),
		StripeCurrentPeriodStart: time.Unix(item.CurrentPeriodStart, 0),
		StripeCurrentPeriodEnd:   time.Unix(item.CurrentPeriodEnd, 0),
		StripeCancelAt:           annictstripe.NullTimeFromUnix(sub.CancelAt),
		StripeCanceledAt:         annictstripe.NullTimeFromUnix(sub.CanceledAt),
	})
	if err != nil {
		return nil, fmt.Errorf("StripeSubscriber作成に失敗: %w", err)
	}

	// ユーザーとの紐付け
	err = uc.userRepo.UpdateStripeSubscriberID(ctx, input.UserID, &stripeSubscriber.ID)
	if err != nil {
		return nil, fmt.Errorf("ユーザー紐付けに失敗: %w", err)
	}

	// トランザクションコミット
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションコミットに失敗: %w", err)
	}

	return &CreateStripeSubscriberResult{
		StripeSubscriber: stripeSubscriber,
	}, nil
}

// ParseUserIDFromMetadata はCheckoutセッションのmetadataからユーザーIDを取得します
func ParseUserIDFromMetadata(metadata map[string]string) (int64, error) {
	userIDStr, ok := metadata["user_id"]
	if !ok {
		return 0, &MetadataUserIDMissingError{}
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return 0, &MetadataUserIDInvalidError{Value: userIDStr}
	}

	return userID, nil
}

// InvalidSubscriptionStatusError は無効なサブスクリプションステータスを示すエラー
type InvalidSubscriptionStatusError struct {
	Status string
}

func (e *InvalidSubscriptionStatusError) Error() string {
	return fmt.Sprintf("invalid subscription status: %s", e.Status)
}

// IsInvalidSubscriptionStatusError はエラーがInvalidSubscriptionStatusErrorかどうかを判定します
func IsInvalidSubscriptionStatusError(err error) bool {
	var e *InvalidSubscriptionStatusError
	return errors.As(err, &e)
}

// MetadataUserIDMissingError はmetadataにuser_idが含まれていないことを示すエラー
type MetadataUserIDMissingError struct{}

func (e *MetadataUserIDMissingError) Error() string {
	return "user_id is missing from metadata"
}

// IsMetadataUserIDMissingError はエラーがMetadataUserIDMissingErrorかどうかを判定します
func IsMetadataUserIDMissingError(err error) bool {
	var e *MetadataUserIDMissingError
	return errors.As(err, &e)
}

// MetadataUserIDInvalidError はmetadataのuser_idが無効であることを示すエラー
type MetadataUserIDInvalidError struct {
	Value string
}

func (e *MetadataUserIDInvalidError) Error() string {
	return fmt.Sprintf("invalid user_id in metadata: %s", e.Value)
}

// IsMetadataUserIDInvalidError はエラーがMetadataUserIDInvalidErrorかどうかを判定します
func IsMetadataUserIDInvalidError(err error) bool {
	var e *MetadataUserIDInvalidError
	return errors.As(err, &e)
}
