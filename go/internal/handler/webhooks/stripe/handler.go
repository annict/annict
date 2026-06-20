// Package stripe はStripe Webhook関連のハンドラーを提供します
package stripe

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/usecase"
)

// Handler はStripe Webhook関連のHTTPハンドラーです
type Handler struct {
	cfg                    *config.Config
	processStripeWebhookUC *usecase.ProcessStripeWebhookUsecase
}

// NewHandler は新しいHandlerを作成します
func NewHandler(
	cfg *config.Config,
	processStripeWebhookUC *usecase.ProcessStripeWebhookUsecase,
) *Handler {
	return &Handler{
		cfg:                    cfg,
		processStripeWebhookUC: processStripeWebhookUC,
	}
}
