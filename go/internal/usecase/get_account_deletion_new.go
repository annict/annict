package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetAccountDeletionNewUsecaseは退会の確認ページに必要なデータ (サポーター契約の状態と、
// 退会を妨げるOAuthアプリケーションの有無) を取得するユースケース。
type GetAccountDeletionNewUsecase struct {
	stripeSubscriberRepo  *repository.StripeSubscriberRepository
	gumroadSubscriberRepo *repository.GumroadSubscriberRepository
	oauthApplicationRepo  *repository.OAuthApplicationRepository
}

func NewGetAccountDeletionNewUsecase(
	stripeSubscriberRepo *repository.StripeSubscriberRepository,
	gumroadSubscriberRepo *repository.GumroadSubscriberRepository,
	oauthApplicationRepo *repository.OAuthApplicationRepository,
) *GetAccountDeletionNewUsecase {
	return &GetAccountDeletionNewUsecase{
		stripeSubscriberRepo:  stripeSubscriberRepo,
		gumroadSubscriberRepo: gumroadSubscriberRepo,
		oauthApplicationRepo:  oauthApplicationRepo,
	}
}

type GetAccountDeletionNewInput struct {
	User *model.User
}

type GetAccountDeletionNewOutput struct {
	// IsStripeSupporterは、退会すると削除ジョブが即時解約するStripeのサブスクリプションがあるかを表す。
	IsStripeSupporter bool
	// IsGumroadSupporterは、Gumroad側での解約が必要なサブスクリプションがあるかを表す。
	IsGumroadSupporter bool
	// HasKeptOAuthApplicationsは、退会を妨げる論理削除されていないOAuthアプリケーションがあるかを表す。
	HasKeptOAuthApplications bool
}

func (uc *GetAccountDeletionNewUsecase) Execute(ctx context.Context, input GetAccountDeletionNewInput) (*GetAccountDeletionNewOutput, error) {
	user := input.User
	if user == nil {
		return nil, fmt.Errorf("ユーザーがnilです")
	}

	isStripeSupporter, err := uc.isStripeSupporter(ctx, user)
	if err != nil {
		return nil, err
	}

	isGumroadSupporter, err := uc.isGumroadSupporter(ctx, user)
	if err != nil {
		return nil, err
	}

	hasKeptOAuthApplications, err := uc.oauthApplicationRepo.ExistsKeptByOwnerID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("OAuthアプリケーションの有無の取得に失敗: %w", err)
	}

	return &GetAccountDeletionNewOutput{
		IsStripeSupporter:        isStripeSupporter,
		IsGumroadSupporter:       isGumroadSupporter,
		HasKeptOAuthApplications: hasKeptOAuthApplications,
	}, nil
}

func (uc *GetAccountDeletionNewUsecase) isStripeSupporter(ctx context.Context, user *model.User) (bool, error) {
	isStripeSupporter, err := uc.stripeSubscriberRepo.HasActiveSubscription(ctx, user.StripeSubscriberID)
	if err != nil {
		return false, fmt.Errorf("StripeSubscriberの取得に失敗: %w", err)
	}
	return isStripeSupporter, nil
}

func (uc *GetAccountDeletionNewUsecase) isGumroadSupporter(ctx context.Context, user *model.User) (bool, error) {
	if user.GumroadSubscriberID == nil {
		return false, nil
	}

	subscriber, err := uc.gumroadSubscriberRepo.GetByID(ctx, *user.GumroadSubscriberID)
	if err != nil {
		return false, fmt.Errorf("GumroadSubscriberの取得に失敗: %w", err)
	}

	return uc.gumroadSubscriberRepo.IsActive(&subscriber), nil
}
