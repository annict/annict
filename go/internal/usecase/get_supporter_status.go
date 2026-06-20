package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetSupporterStatusUsecase はユーザーのサポーターステータスを取得するユースケースです
type GetSupporterStatusUsecase struct {
	stripeSubscriberRepo  *repository.StripeSubscriberRepository
	gumroadSubscriberRepo *repository.GumroadSubscriberRepository
}

// NewGetSupporterStatusUsecase は新しいGetSupporterStatusUsecaseを作成します
func NewGetSupporterStatusUsecase(
	stripeSubscriberRepo *repository.StripeSubscriberRepository,
	gumroadSubscriberRepo *repository.GumroadSubscriberRepository,
) *GetSupporterStatusUsecase {
	return &GetSupporterStatusUsecase{
		stripeSubscriberRepo:  stripeSubscriberRepo,
		gumroadSubscriberRepo: gumroadSubscriberRepo,
	}
}

// GetSupporterStatusInput はユースケースの入力です
type GetSupporterStatusInput struct {
	User *model.User
}

// GetSupporterStatusOutput はユースケースの出力です
type GetSupporterStatusOutput struct {
	IsStripeActive    bool
	IsGumroadActive   bool
	StripeSubscriber  *model.StripeSubscriber
	GumroadSubscriber *model.GumroadSubscriber
}

// Execute はユーザーのサポーターステータスを取得します
func (uc *GetSupporterStatusUsecase) Execute(ctx context.Context, input GetSupporterStatusInput) (*GetSupporterStatusOutput, error) {
	user := input.User
	if user == nil {
		return nil, fmt.Errorf("ユーザーがnilです")
	}

	output := &GetSupporterStatusOutput{}

	// Stripeサブスクリプションをチェック
	if user.StripeSubscriberID != nil && uc.stripeSubscriberRepo != nil {
		stripeSubscriber, err := uc.stripeSubscriberRepo.GetByID(ctx, *user.StripeSubscriberID)
		if err == nil && stripeSubscriber != nil && uc.stripeSubscriberRepo.IsActive(stripeSubscriber) {
			output.IsStripeActive = true
			output.StripeSubscriber = stripeSubscriber
		}
	}

	// Gumroadサブスクリプションをチェック
	if user.GumroadSubscriberID != nil && uc.gumroadSubscriberRepo != nil {
		gumroadSubscriber, err := uc.gumroadSubscriberRepo.GetByID(ctx, *user.GumroadSubscriberID)
		if err == nil && uc.gumroadSubscriberRepo.IsActive(&gumroadSubscriber) {
			output.IsGumroadActive = true
			output.GumroadSubscriber = &gumroadSubscriber
		}
	}

	return output, nil
}
