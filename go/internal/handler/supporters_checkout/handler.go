// Package supporters_checkout はサポーター登録Checkout関連のハンドラーを提供します
package supporters_checkout

import (
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// Handler はサポーター登録Checkout関連のHTTPハンドラーです
type Handler struct {
	sessionManager          *session.Manager
	createCheckoutSessionUC *usecase.CreateCheckoutSessionUsecase
}

// NewHandler は新しいHandlerを作成します
func NewHandler(
	sessionManager *session.Manager,
	createCheckoutSessionUC *usecase.CreateCheckoutSessionUsecase,
) *Handler {
	return &Handler{
		sessionManager:          sessionManager,
		createCheckoutSessionUC: createCheckoutSessionUC,
	}
}
