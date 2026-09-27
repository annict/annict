// Package supporters_portalはサポーターCustomer Portal関連のハンドラーを提供します
package supporters_portal

import (
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// HandlerはサポーターCustomer Portal関連のHTTPハンドラーです
type Handler struct {
	flashMgr              *session.FlashManager
	createPortalSessionUC *usecase.CreatePortalSessionUsecase
}

// NewHandlerは新しいHandlerを作成します
func NewHandler(
	flashMgr *session.FlashManager,
	createPortalSessionUC *usecase.CreatePortalSessionUsecase,
) *Handler {
	return &Handler{
		flashMgr:              flashMgr,
		createPortalSessionUC: createPortalSessionUC,
	}
}
