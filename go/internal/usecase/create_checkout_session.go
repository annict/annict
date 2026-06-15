package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	annictstripe "github.com/annict/annict/go/internal/stripe"
	"github.com/annict/annict/go/internal/validator"
)

// CheckoutSessionCreator abstracts creating a Stripe Checkout session. It is
// defined on the caller (UseCase) side so the UseCase depends on a small
// interface rather than the concrete *stripe.Client, and tests can inject a fake.
//
// [Ja] CheckoutSessionCreator は Stripe Checkout セッション作成を抽象化する。
// 呼び出し側 (UseCase) で定義することで、UseCase は具象 *stripe.Client ではなく
// 小さな interface に依存し、テストでは fake を注入できる。
type CheckoutSessionCreator interface {
	CreateCheckoutSession(ctx context.Context, params annictstripe.CheckoutSessionParams) (string, error)
}

// CreateCheckoutSessionUsecase はStripe Checkoutセッション作成のユースケースです
type CreateCheckoutSessionUsecase struct {
	cfg                  *config.Config
	stripeSubscriberRepo *repository.StripeSubscriberRepository
	stripeCfg            *annictstripe.Config
	checkoutCreator      CheckoutSessionCreator
	validator            *validator.SupportersCheckoutCreateValidator
}

// NewCreateCheckoutSessionUsecase は新しいCreateCheckoutSessionUsecaseを作成します
func NewCreateCheckoutSessionUsecase(
	cfg *config.Config,
	stripeSubscriberRepo *repository.StripeSubscriberRepository,
	stripeCfg *annictstripe.Config,
	checkoutCreator CheckoutSessionCreator,
	validator *validator.SupportersCheckoutCreateValidator,
) *CreateCheckoutSessionUsecase {
	return &CreateCheckoutSessionUsecase{
		cfg:                  cfg,
		stripeSubscriberRepo: stripeSubscriberRepo,
		stripeCfg:            stripeCfg,
		checkoutCreator:      checkoutCreator,
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
	if user.StripeSubscriberID != nil && uc.stripeSubscriberRepo != nil {
		stripeSubscriber, err := uc.stripeSubscriberRepo.GetByID(ctx, *user.StripeSubscriberID)
		// A not-found subscriber (nil) is normal here (the referenced row may have
		// been removed); only a real retrieval error must abort the checkout. The
		// previous `if err == nil` swallowed every error and let checkout proceed.
		//
		// [Ja] 未存在 (nil) はここでは正常 (参照先の行が消えている可能性がある)。本物の
		// 取得エラーのみ checkout を中断させる。以前の `if err == nil` は全エラーを握り潰し、
		// checkout を続行させていた。
		if err != nil {
			return nil, fmt.Errorf("StripeSubscriber取得に失敗: %w", err)
		}
		if stripeSubscriber != nil && uc.stripeSubscriberRepo.IsActive(stripeSubscriber) {
			return nil, model.NewAppError(
				model.AppErrCodeConflict,
				i18n.T(ctx, "supporters_checkout_already_active"),
				nil,
			)
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
	if uc.checkoutCreator == nil {
		return nil, fmt.Errorf("Stripeクライアントが設定されていません")
	}

	// 5. Stripe Checkoutセッションの作成
	successURL := uc.cfg.AppURL() + "/supporters?success=true"
	cancelURL := uc.cfg.AppURL() + "/supporters?canceled=true"

	// Default to English; only "ja" is rendered in Japanese.
	// [Ja] デフォルトは英語。"ja" のときのみ日本語で表示する。
	locale := "en"
	if input.Locale == "ja" {
		locale = "ja"
	}

	checkoutURL, err := uc.checkoutCreator.CreateCheckoutSession(ctx, annictstripe.CheckoutSessionParams{
		PriceID:    priceID,
		SuccessURL: successURL,
		CancelURL:  cancelURL,
		UserID:     user.ID.String(),
		Locale:     locale,
	})
	if err != nil {
		return nil, fmt.Errorf("stripe Checkoutセッションの作成に失敗しました: %w", err)
	}

	return &CreateCheckoutSessionOutput{CheckoutURL: checkoutURL}, nil
}
