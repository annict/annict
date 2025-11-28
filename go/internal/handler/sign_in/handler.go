// Package sign_in はログイン機能を提供します
package sign_in

import (
	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/turnstile"
	"github.com/annict/annict/internal/usecase"
)

// Handler サインイン関連のHTTPハンドラーです
type Handler struct {
	cfg              *config.Config
	sessionMgr       *session.Manager
	userRepo         *repository.UserRepository
	sendSignInCodeUC *usecase.SendSignInCodeUsecase
	turnstileClient  *turnstile.Client
}

// NewHandler 新しいHandlerを作成します
func NewHandler(
	cfg *config.Config,
	sessionMgr *session.Manager,
	userRepo *repository.UserRepository,
	sendSignInCodeUC *usecase.SendSignInCodeUsecase,
	turnstileClient *turnstile.Client,
) *Handler {
	return &Handler{
		cfg:              cfg,
		sessionMgr:       sessionMgr,
		userRepo:         userRepo,
		sendSignInCodeUC: sendSignInCodeUC,
		turnstileClient:  turnstileClient,
	}
}
