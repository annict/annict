package usecase

import (
	"context"
	"fmt"
	"strconv"

	"github.com/stripe/stripe-go/v84"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	annictStripe "github.com/annict/annict/go/internal/stripe"
	"github.com/annict/annict/go/internal/validator"
)

// CreateCheckoutSessionUsecase はStripe Checkoutセッション作成のユースケースです
type CreateCheckoutSessionUsecase struct {
	cfg                  *config.Config
	stripeSubscriberRepo *repository.StripeSubscriberRepository
	stripeCfg            *annictStripe.Config
	stripeClient         *stripe.Client
	validator            *validator.SupportersCheckoutCreateValidator
}

// NewCreateCheckoutSessionUsecase は新しいCreateCheckoutSessionUsecaseを作成します
func NewCreateCheckoutSessionUsecase(
	cfg *config.Config,
	stripeSubscriberRepo *repository.StripeSubscriberRepository,
	stripeCfg *annictStripe.Config,
	stripeClient *stripe.Client,
	validator *validator.SupportersCheckoutCreateValidator,
) *CreateCheckoutSessionUsecase {
	return &CreateCheckoutSessionUsecase{
		cfg:                  cfg,
		stripeSubscriberRepo: stripeSubscriberRepo,
		stripeCfg:            stripeCfg,
		stripeClient:         stripeClient,
		validator:            validator,
	}
}

// CreateCheckoutSessionInput はユースケースの入力です
type CreateCheckoutSessionInput struct {
	User   *model.User
	Plan   string
	Locale string
}

// CreateCheckoutSessionOutput はユースケースの出力です
type CreateCheckoutSessionOutput struct {
	CheckoutURL string
}

// Execute はStripe Checkoutセッションを作成します
func (uc *CreateCheckoutSessionUsecase) Execute(ctx context.Context, input CreateCheckoutSessionInput) (*CreateCheckoutSessionOutput, error) {
	// 1. バリデーション
	if err := uc.validator.Validate(ctx, validator.SupportersCheckoutCreateValidatorInput{
		Plan: input.Plan,
	}); err != nil {
		return nil, err
	}

	// 2. 重複サブスクリプションチェック
	user := input.User
	if user.StripeSubscriberID.Valid && uc.stripeSubscriberRepo != nil {
		stripeSubscriber, err := uc.stripeSubscriberRepo.GetByID(ctx, user.StripeSubscriberID.Int64)
		if err == nil {
			if uc.stripeSubscriberRepo.IsActive(&stripeSubscriber) {
				return nil, model.NewAppError(
					model.AppErrCodeConflict,
					i18n.T(ctx, "supporters_checkout_already_active"),
					nil,
				)
			}
		}
	}

	// 3. 価格IDの決定
	var priceID string
	switch input.Plan {
	case "monthly":
		priceID = uc.stripeCfg.PriceMonthlyID
	case "yearly":
		priceID = uc.stripeCfg.PriceYearlyID
	}

	if priceID == "" {
		return nil, fmt.Errorf("Stripe価格IDが設定されていません: plan=%s", input.Plan)
	}

	// 4. Stripeクライアントのチェック
	if uc.stripeClient == nil {
		return nil, fmt.Errorf("Stripeクライアントが設定されていません")
	}

	// 5. Stripe Checkoutセッションの作成
	successURL := uc.cfg.AppURL() + "/supporters?success=true"
	cancelURL := uc.cfg.AppURL() + "/supporters?canceled=true"

	params := &stripe.CheckoutSessionCreateParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionCreateLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Metadata: map[string]string{
			"user_id": strconv.FormatInt(user.ID, 10),
		},
	}

	// ロケールの設定
	if input.Locale == "ja" {
		params.Locale = stripe.String("ja")
	} else {
		params.Locale = stripe.String("en")
	}

	checkoutSession, err := uc.stripeClient.V1CheckoutSessions.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("stripe Checkoutセッションの作成に失敗しました: %w", err)
	}

	return &CreateCheckoutSessionOutput{CheckoutURL: checkoutSession.URL}, nil
}
