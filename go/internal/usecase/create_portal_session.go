package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/stripe/stripe-go/v84"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// CreatePortalSessionUsecase はStripe Customer Portalセッション作成のユースケースです
type CreatePortalSessionUsecase struct {
	cfg                  *config.Config
	stripeSubscriberRepo *repository.StripeSubscriberRepository
	stripeClient         *stripe.Client
}

// NewCreatePortalSessionUsecase は新しいCreatePortalSessionUsecaseを作成します
func NewCreatePortalSessionUsecase(
	cfg *config.Config,
	stripeSubscriberRepo *repository.StripeSubscriberRepository,
	stripeClient *stripe.Client,
) *CreatePortalSessionUsecase {
	return &CreatePortalSessionUsecase{
		cfg:                  cfg,
		stripeSubscriberRepo: stripeSubscriberRepo,
		stripeClient:         stripeClient,
	}
}

// CreatePortalSessionInput はユースケースの入力です
type CreatePortalSessionInput struct {
	User   *model.User
	Locale string
}

// CreatePortalSessionOutput はユースケースの出力です
type CreatePortalSessionOutput struct {
	PortalURL string
}

// NotStripeSubscriberError はStripeサポーターではない場合のエラーです
type NotStripeSubscriberError struct{}

func (e *NotStripeSubscriberError) Error() string {
	return "Stripeサポーターではありません"
}

// IsNotStripeSubscriberError はNotStripeSubscriberErrorかどうかを判定します
func IsNotStripeSubscriberError(err error) bool {
	var e *NotStripeSubscriberError
	return errors.As(err, &e)
}

// Execute はStripe Customer Portalセッションを作成します
func (uc *CreatePortalSessionUsecase) Execute(ctx context.Context, input CreatePortalSessionInput) (*CreatePortalSessionOutput, error) {
	user := input.User

	// 1. Stripeサポーターのチェック
	if !user.StripeSubscriberID.Valid {
		return nil, &NotStripeSubscriberError{}
	}

	stripeSubscriber, err := uc.stripeSubscriberRepo.GetByID(ctx, user.StripeSubscriberID.Int64)
	if err != nil {
		return nil, fmt.Errorf("stripeサブスクライバーの取得に失敗しました: %w", err)
	}

	// 2. アクティブなサブスクリプションのチェック
	if !uc.stripeSubscriberRepo.IsActive(&stripeSubscriber) {
		return nil, &NotStripeSubscriberError{}
	}

	// 3. Stripeクライアントのチェック
	if uc.stripeClient == nil {
		return nil, fmt.Errorf("Stripeクライアントが設定されていません")
	}

	// 4. Stripe Customer Portalセッションの作成
	returnURL := uc.cfg.AppURL() + "/supporters"

	params := &stripe.BillingPortalSessionCreateParams{
		Customer:  stripe.String(stripeSubscriber.StripeCustomerID),
		ReturnURL: stripe.String(returnURL),
	}

	if input.Locale == "ja" {
		params.Locale = stripe.String("ja")
	} else {
		params.Locale = stripe.String("en")
	}

	portalSession, err := uc.stripeClient.V1BillingPortalSessions.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("stripe Customer Portalセッションの作成に失敗しました: %w", err)
	}

	return &CreatePortalSessionOutput{PortalURL: portalSession.URL}, nil
}
