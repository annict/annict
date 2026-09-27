// Package sign_outはログアウト機能を提供します
package sign_out

import (
	"github.com/annict/annict/go/internal/session"
)

// Handlerはログアウト関連のHTTPハンドラーです
type Handler struct {
	sessionMgr *session.Manager
}

// NewHandlerは新しいHandlerを作成します
func NewHandler(sessionMgr *session.Manager) *Handler {
	return &Handler{
		sessionMgr: sessionMgr,
	}
}
