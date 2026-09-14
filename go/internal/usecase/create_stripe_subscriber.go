// Package usecaseはビジネスロジック層のユースケースを提供します
package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	annictstripe "github.com/annict/annict/go/internal/stripe"
)

// SubscriptionRetrieverはStripeからのサブスクリプション取得を抽象化する。
// 呼び出し側 (UseCase) で定義することで、UseCaseは具象 *stripe.Clientではなく
// 小さなinterfaceに依存し、テストではfakeを注入できる。
type SubscriptionRetriever interface {
	RetrieveSubscription(ctx context.Context, subscriptionID string) (*annictstripe.Subscription, error)
}

// CreateStripeSubscriberUsecaseはcheckout.session.completedイベント処理のユースケース
type CreateStripeSubscriberUsecase struct {
	db                    *sql.DB
	stripeSubscriberRepo  *repository.StripeSubscriberRepository
	userRepo              *repository.UserRepository
	subscriptionRetriever SubscriptionRetriever
}

// NewCreateStripeSubscriberUsecaseはCreateStripeSubscriberUsecaseを作成します
func NewCreateStripeSubscriberUsecase(
	db *sql.DB,
	stripeSubscriberRepo *repository.StripeSubscriberRepository,
	userRepo *repository.UserRepository,
	subscriptionRetriever SubscriptionRetriever,
) *CreateStripeSubscriberUsecase {
	return &CreateStripeSubscriberUsecase{
		db:                    db,
		stripeSubscriberRepo:  stripeSubscriberRepo,
		userRepo:              userRepo,
		subscriptionRetriever: subscriptionRetriever,
	}
}

// CreateStripeSubscriberInputはcheckout.session.completedイベントの入力データ
type CreateStripeSubscriberInput struct {
	StripeCustomerID     string       // Stripeの顧客ID (cus_xxx)
	StripeSubscriptionID string       // StripeのサブスクリプションID (sub_xxx)
	UserID               model.UserID // AnnictのユーザーID (metadataから取得)
}

// CreateStripeSubscriberResultはcheckout.session.completedイベント処理の結果
type CreateStripeSubscriberResult struct {
	StripeSubscriber model.StripeSubscriber
}

// Executeはcheckout.session.completedイベントを処理します
//
// 処理フロー:
// 1. Stripe APIからサブスクリプション詳細を取得
// 2. StripeSubscriberレコードを作成
// 3. Userとの紐付け
func (uc *CreateStripeSubscriberUsecase) Execute(
	ctx context.Context,
	input CreateStripeSubscriberInput,
) (*CreateStripeSubscriberResult, error) {
	// Stripeクライアントがnilの場合はエラー
	if uc.subscriptionRetriever == nil {
		return nil, fmt.Errorf("Stripeクライアントが設定されていません")
	}

	// Stripe APIからサブスクリプション詳細を取得
	sub, err := uc.subscriptionRetriever.RetrieveSubscription(ctx, input.StripeSubscriptionID)
	if err != nil {
		return nil, fmt.Errorf("サブスクリプション取得に失敗: %w", err)
	}

	// ステータスを検証
	status := model.StripeSubscriptionStatus(sub.Status)
	if !status.IsValid() {
		return nil, &InvalidSubscriptionStatusError{Status: sub.Status}
	}

	// 価格IDと請求期間を取得 (最初のアイテムから)
	if len(sub.Items) == 0 {
		return nil, fmt.Errorf("サブスクリプションにアイテムが含まれていません")
	}
	item := sub.Items[0]
	priceID := item.PriceID

	// トランザクション開始
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// トランザクションを使用するRepositoryを取得
	stripeSubscriberRepoTx := uc.stripeSubscriberRepo.WithTx(tx)
	userRepoTx := uc.userRepo.WithTx(tx)

	// StripeSubscriberレコードを作成
	stripeSubscriber, err := stripeSubscriberRepoTx.Create(ctx, repository.CreateStripeSubscriberParams{
		StripeCustomerID:         input.StripeCustomerID,
		StripeSubscriptionID:     input.StripeSubscriptionID,
		StripePriceID:            priceID,
		StripeStatus:             sub.Status,
		StripeCurrentPeriodStart: item.CurrentPeriodStart,
		StripeCurrentPeriodEnd:   item.CurrentPeriodEnd,
		StripeCancelAt:           sub.CancelAt,
		StripeCanceledAt:         sub.CanceledAt,
	})
	if err != nil {
		return nil, fmt.Errorf("StripeSubscriber作成に失敗: %w", err)
	}

	// ユーザーとの紐付け
	err = userRepoTx.UpdateStripeSubscriberID(ctx, input.UserID, &stripeSubscriber.ID)
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

// ParseUserIDFromMetadataはCheckoutセッションのmetadataからユーザーIDを取得します
func ParseUserIDFromMetadata(metadata map[string]string) (model.UserID, error) {
	userIDStr, ok := metadata["user_id"]
	if !ok {
		return 0, &MetadataUserIDMissingError{}
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return 0, &MetadataUserIDInvalidError{Value: userIDStr}
	}

	return model.UserID(userID), nil
}

// InvalidSubscriptionStatusErrorは無効なサブスクリプションステータスを示すエラー
type InvalidSubscriptionStatusError struct {
	Status string
}

func (e *InvalidSubscriptionStatusError) Error() string {
	return fmt.Sprintf("invalid subscription status: %s", e.Status)
}

// IsInvalidSubscriptionStatusErrorはエラーがInvalidSubscriptionStatusErrorかどうかを判定します
func IsInvalidSubscriptionStatusError(err error) bool {
	var e *InvalidSubscriptionStatusError
	return errors.As(err, &e)
}

// MetadataUserIDMissingErrorはmetadataにuser_idが含まれていないことを示すエラー
type MetadataUserIDMissingError struct{}

func (e *MetadataUserIDMissingError) Error() string {
	return "user_id is missing from metadata"
}

// IsMetadataUserIDMissingErrorはエラーがMetadataUserIDMissingErrorかどうかを判定します
func IsMetadataUserIDMissingError(err error) bool {
	var e *MetadataUserIDMissingError
	return errors.As(err, &e)
}

// MetadataUserIDInvalidErrorはmetadataのuser_idが無効であることを示すエラー
type MetadataUserIDInvalidError struct {
	Value string
}

func (e *MetadataUserIDInvalidError) Error() string {
	return fmt.Sprintf("invalid user_id in metadata: %s", e.Value)
}

// IsMetadataUserIDInvalidErrorはエラーがMetadataUserIDInvalidErrorかどうかを判定します
func IsMetadataUserIDInvalidError(err error) bool {
	var e *MetadataUserIDInvalidError
	return errors.As(err, &e)
}
