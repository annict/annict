// Package supporters_checkout はサポーター登録Checkout関連のハンドラーを提供します
package supporters_checkout

import (
	"github.com/stripe/stripe-go/v84"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	annictStripe "github.com/annict/annict/go/internal/stripe"
)

// Handler はサポーター登録Checkout関連のHTTPハンドラーです
type Handler struct {
	cfg                  *config.Config
	sessionManager       *session.Manager
	stripeSubscriberRepo *repository.StripeSubscriberRepository
	stripeCfg            *annictStripe.Config
	stripeClient         *stripe.Client
}

// NewHandler は新しいHandlerを作成します
func NewHandler(
	cfg *config.Config,
	sessionManager *session.Manager,
	stripeSubscriberRepo *repository.StripeSubscriberRepository,
	stripeCfg *annictStripe.Config,
	stripeClient *stripe.Client,
) *Handler {
	return &Handler{
		cfg:                  cfg,
		sessionManager:       sessionManager,
		stripeSubscriberRepo: stripeSubscriberRepo,
		stripeCfg:            stripeCfg,
		stripeClient:         stripeClient,
	}
}
