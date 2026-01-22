// Package supporters_portal はサポーターCustomer Portal関連のハンドラーを提供します
package supporters_portal

import (
	"github.com/stripe/stripe-go/v84"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
)

// Handler はサポーターCustomer Portal関連のHTTPハンドラーです
type Handler struct {
	cfg                  *config.Config
	sessionManager       *session.Manager
	stripeSubscriberRepo *repository.StripeSubscriberRepository
	stripeClient         *stripe.Client
}

// NewHandler は新しいHandlerを作成します
func NewHandler(
	cfg *config.Config,
	sessionManager *session.Manager,
	stripeSubscriberRepo *repository.StripeSubscriberRepository,
	stripeClient *stripe.Client,
) *Handler {
	return &Handler{
		cfg:                  cfg,
		sessionManager:       sessionManager,
		stripeSubscriberRepo: stripeSubscriberRepo,
		stripeClient:         stripeClient,
	}
}
