// Package sign_inはログイン機能を提供します
package sign_in

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/turnstile"
	"github.com/annict/annict/go/internal/usecase"
)

// Handlerサインイン関連のHTTPハンドラーです
type Handler struct {
	cfg              *config.Config
	sessionMgr       *session.Manager
	flashMgr         *session.FlashManager
	sendSignInCodeUC *usecase.SendSignInCodeUsecase
	turnstileClient  *turnstile.Client
}

// NewHandler新しいHandlerを作成します
func NewHandler(
	cfg *config.Config,
	sessionMgr *session.Manager,
	flashMgr *session.FlashManager,
	sendSignInCodeUC *usecase.SendSignInCodeUsecase,
	turnstileClient *turnstile.Client,
) *Handler {
	return &Handler{
		cfg:              cfg,
		sessionMgr:       sessionMgr,
		flashMgr:         flashMgr,
		sendSignInCodeUC: sendSignInCodeUC,
		turnstileClient:  turnstileClient,
	}
}
