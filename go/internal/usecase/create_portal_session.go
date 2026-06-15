package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	annictstripe "github.com/annict/annict/go/internal/stripe"
)

// PortalSessionCreator abstracts creating a Stripe Billing Portal session. It is
// defined on the caller (UseCase) side so the UseCase depends on a small
// interface rather than the concrete *stripe.Client, and tests can inject a fake.
//
// [Ja] PortalSessionCreator は Stripe Billing Portal セッション作成を抽象化する。
// 呼び出し側 (UseCase) で定義することで、UseCase は具象 *stripe.Client ではなく
// 小さな interface に依存し、テストでは fake を注入できる。
type PortalSessionCreator interface {
	CreatePortalSession(ctx context.Context, params annictstripe.PortalSessionParams) (string, error)
}

// CreatePortalSessionUsecase はStripe Customer Portalセッション作成のユースケースです
type CreatePortalSessionUsecase struct {
	cfg                  *config.Config
	stripeSubscriberRepo *repository.StripeSubscriberRepository
	portalCreator        PortalSessionCreator
}

// NewCreatePortalSessionUsecase は新しいCreatePortalSessionUsecaseを作成します
func NewCreatePortalSessionUsecase(
	cfg *config.Config,
	stripeSubscriberRepo *repository.StripeSubscriberRepository,
	portalCreator PortalSessionCreator,
) *CreatePortalSessionUsecase {
	return &CreatePortalSessionUsecase{
		cfg:                  cfg,
		stripeSubscriberRepo: stripeSubscriberRepo,
		portalCreator:        portalCreator,
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
	if user.StripeSubscriberID == nil {
		return nil, &NotStripeSubscriberError{}
	}

	stripeSubscriber, err := uc.stripeSubscriberRepo.GetByID(ctx, *user.StripeSubscriberID)
	if err != nil {
		return nil, fmt.Errorf("stripeサブスクライバーの取得に失敗しました: %w", err)
	}
	// The user references a subscriber that no longer exists; treat it as a
	// non-supporter rather than dereferencing a nil value.
	//
	// [Ja] ユーザーが既に存在しないサブスクライバーを参照している状態であり、nil を
	// 参照外しせず非サポーターとして扱う。
	if stripeSubscriber == nil {
		return nil, &NotStripeSubscriberError{}
	}

	// 2. アクティブなサブスクリプションのチェック
	if !uc.stripeSubscriberRepo.IsActive(stripeSubscriber) {
		return nil, &NotStripeSubscriberError{}
	}

	// 3. Stripeクライアントのチェック
	if uc.portalCreator == nil {
		return nil, fmt.Errorf("Stripeクライアントが設定されていません")
	}

	// 4. Stripe Customer Portalセッションの作成
	returnURL := uc.cfg.AppURL() + "/supporters"

	// Default to English; only "ja" is rendered in Japanese.
	// [Ja] デフォルトは英語。"ja" のときのみ日本語で表示する。
	locale := "en"
	if input.Locale == "ja" {
		locale = "ja"
	}

	portalURL, err := uc.portalCreator.CreatePortalSession(ctx, annictstripe.PortalSessionParams{
		CustomerID: stripeSubscriber.StripeCustomerID,
		ReturnURL:  returnURL,
		Locale:     locale,
	})
	if err != nil {
		return nil, fmt.Errorf("stripe Customer Portalセッションの作成に失敗しました: %w", err)
	}

	return &CreatePortalSessionOutput{PortalURL: portalURL}, nil
}
