// Package supporters はサポーターページ関連のハンドラーを提供します
package supporters

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
)

// Handler はサポーター関連のHTTPハンドラーです
type Handler struct {
	cfg                   *config.Config
	sessionManager        *session.Manager
	stripeSubscriberRepo  *repository.StripeSubscriberRepository
	gumroadSubscriberRepo *repository.GumroadSubscriberRepository
}

// NewHandler は新しいHandlerを作成します
func NewHandler(
	cfg *config.Config,
	sessionManager *session.Manager,
	stripeSubscriberRepo *repository.StripeSubscriberRepository,
	gumroadSubscriberRepo *repository.GumroadSubscriberRepository,
) *Handler {
	return &Handler{
		cfg:                   cfg,
		sessionManager:        sessionManager,
		stripeSubscriberRepo:  stripeSubscriberRepo,
		gumroadSubscriberRepo: gumroadSubscriberRepo,
	}
}
