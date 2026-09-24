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

// PortalSessionCreatorはStripe Billing Portalセッション作成を抽象化する。
// 呼び出し側 (UseCase) で定義することで、UseCaseは具象 *stripe.Clientではなく
// 小さなinterfaceに依存し、テストではfakeを注入できる。
type PortalSessionCreator interface {
	CreatePortalSession(ctx context.Context, params annictstripe.PortalSessionParams) (string, error)
}

// CreatePortalSessionUsecaseはStripe Customer Portalセッション作成のユースケースです
type CreatePortalSessionUsecase struct {
	cfg                  *config.Config
	stripeSubscriberRepo *repository.StripeSubscriberRepository
	portalCreator        PortalSessionCreator
}

// NewCreatePortalSessionUsecaseは新しいCreatePortalSessionUsecaseを作成します
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

// CreatePortalSessionInputはユースケースの入力です
type CreatePortalSessionInput struct {
	User   *model.User
	Locale string
}

// CreatePortalSessionOutputはユースケースの出力です
type CreatePortalSessionOutput struct {
	PortalURL string
}

// NotStripeSubscriberErrorはStripeサポーターではない場合のエラーです
type NotStripeSubscriberError struct{}

func (e *NotStripeSubscriberError) Error() string {
	return "Stripeサポーターではありません"
}

// IsNotStripeSubscriberErrorはNotStripeSubscriberErrorかどうかを判定します
func IsNotStripeSubscriberError(err error) bool {
	var e *NotStripeSubscriberError
	return errors.As(err, &e)
}

// ExecuteはStripe Customer Portalセッションを作成します
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
	// ユーザーが既に存在しないサブスクライバーを参照している状態であり、nilを
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

	// デフォルトは英語。"ja" のときのみ日本語で表示する。
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
