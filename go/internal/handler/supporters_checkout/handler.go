// Package supporters_checkoutはサポーター登録Checkout関連のハンドラーを提供します
package supporters_checkout

import (
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// Handlerはサポーター登録Checkout関連のHTTPハンドラーです
type Handler struct {
	flashMgr                *session.FlashManager
	createCheckoutSessionUC *usecase.CreateCheckoutSessionUsecase
}

// NewHandlerは新しいHandlerを作成します
func NewHandler(
	flashMgr *session.FlashManager,
	createCheckoutSessionUC *usecase.CreateCheckoutSessionUsecase,
) *Handler {
	return &Handler{
		flashMgr:                flashMgr,
		createCheckoutSessionUC: createCheckoutSessionUC,
	}
}
