// Package stripe はStripe Webhook関連のハンドラーを提供します
package stripe

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/usecase"
)

// Handler はStripe Webhook関連のHTTPハンドラーです
type Handler struct {
	cfg                      *config.Config
	stripeWebhookEventRepo   *repository.StripeWebhookEventRepository
	stripeSubscriberRepo     *repository.StripeSubscriberRepository
	userRepo                 *repository.UserRepository
	createStripeSubscriberUC *usecase.CreateStripeSubscriberUsecase
	updateStripeSubscriberUC *usecase.UpdateStripeSubscriberUsecase
	deleteStripeSubscriberUC *usecase.DeleteStripeSubscriberUsecase
}

// NewHandler は新しいHandlerを作成します
func NewHandler(
	cfg *config.Config,
	stripeWebhookEventRepo *repository.StripeWebhookEventRepository,
	stripeSubscriberRepo *repository.StripeSubscriberRepository,
	userRepo *repository.UserRepository,
	createStripeSubscriberUC *usecase.CreateStripeSubscriberUsecase,
	updateStripeSubscriberUC *usecase.UpdateStripeSubscriberUsecase,
	deleteStripeSubscriberUC *usecase.DeleteStripeSubscriberUsecase,
) *Handler {
	return &Handler{
		cfg:                      cfg,
		stripeWebhookEventRepo:   stripeWebhookEventRepo,
		stripeSubscriberRepo:     stripeSubscriberRepo,
		userRepo:                 userRepo,
		createStripeSubscriberUC: createStripeSubscriberUC,
		updateStripeSubscriberUC: updateStripeSubscriberUC,
		deleteStripeSubscriberUC: deleteStripeSubscriberUC,
	}
}
