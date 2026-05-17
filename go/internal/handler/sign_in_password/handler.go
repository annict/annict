// Package sign_in_password はパスワードログイン機能を提供します
package sign_in_password

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// Handler サインイン関連のHTTPハンドラーです
type Handler struct {
	cfg                      *config.Config
	sessionMgr               *session.Manager
	flashMgr                 *session.FlashManager
	authenticateByPasswordUC *usecase.AuthenticateByPasswordUsecase
}

// NewHandler 新しいHandlerを作成します
func NewHandler(cfg *config.Config, sessionMgr *session.Manager, flashMgr *session.FlashManager, authenticateByPasswordUC *usecase.AuthenticateByPasswordUsecase) *Handler {
	return &Handler{
		cfg:                      cfg,
		sessionMgr:               sessionMgr,
		flashMgr:                 flashMgr,
		authenticateByPasswordUC: authenticateByPasswordUC,
	}
}
