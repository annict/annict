// Package supporters_portal はサポーターCustomer Portal関連のハンドラーを提供します
package supporters_portal

import (
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// Handler はサポーターCustomer Portal関連のHTTPハンドラーです
type Handler struct {
	sessionManager        *session.Manager
	createPortalSessionUC *usecase.CreatePortalSessionUsecase
}

// NewHandler は新しいHandlerを作成します
func NewHandler(
	sessionManager *session.Manager,
	createPortalSessionUC *usecase.CreatePortalSessionUsecase,
) *Handler {
	return &Handler{
		sessionManager:        sessionManager,
		createPortalSessionUC: createPortalSessionUC,
	}
}
