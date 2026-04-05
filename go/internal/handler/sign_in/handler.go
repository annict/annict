// Package sign_in はログイン機能を提供します
package sign_in

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/turnstile"
	"github.com/annict/annict/go/internal/usecase"
)

// Handler サインイン関連のHTTPハンドラーです
type Handler struct {
	cfg              *config.Config
	sessionMgr       *session.Manager
	sendSignInCodeUC *usecase.SendSignInCodeUsecase
	turnstileClient  *turnstile.Client
}

// NewHandler 新しいHandlerを作成します
func NewHandler(
	cfg *config.Config,
	sessionMgr *session.Manager,
	sendSignInCodeUC *usecase.SendSignInCodeUsecase,
	turnstileClient *turnstile.Client,
) *Handler {
	return &Handler{
		cfg:              cfg,
		sessionMgr:       sessionMgr,
		sendSignInCodeUC: sendSignInCodeUC,
		turnstileClient:  turnstileClient,
	}
}
