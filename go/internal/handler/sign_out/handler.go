// Package sign_out はログアウト機能を提供します
package sign_out

import (
	"github.com/annict/annict/go/internal/session"
)

// Handler はログアウト関連のHTTPハンドラーです
type Handler struct {
	sessionMgr *session.Manager
}

// NewHandler は新しいHandlerを作成します
func NewHandler(sessionMgr *session.Manager) *Handler {
	return &Handler{
		sessionMgr: sessionMgr,
	}
}
