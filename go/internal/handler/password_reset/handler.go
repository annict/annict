// Package password_resetはパスワードリセット機能を提供します
package password_reset

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/ratelimit"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/turnstile"
	"github.com/annict/annict/go/internal/usecase"
)

// Handlerはパスワードリセット申請機能のハンドラーです
type Handler struct {
	cfg                *config.Config
	sessionMgr         *session.Manager
	limiter            *ratelimit.Limiter
	turnstileClient    turnstile.Verifier
	createTokenUseCase *usecase.CreatePasswordResetTokenUsecase
}

// NewHandlerは新しいHandlerを作成します
func NewHandler(cfg *config.Config, sessionMgr *session.Manager, limiter *ratelimit.Limiter, turnstileClient turnstile.Verifier, createTokenUseCase *usecase.CreatePasswordResetTokenUsecase) *Handler {
	return &Handler{
		cfg:                cfg,
		sessionMgr:         sessionMgr,
		limiter:            limiter,
		turnstileClient:    turnstileClient,
		createTokenUseCase: createTokenUseCase,
	}
}
