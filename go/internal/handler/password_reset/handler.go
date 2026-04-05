// Package password_reset はパスワードリセット機能を提供します
package password_reset

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/ratelimit"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/turnstile"
	"github.com/annict/annict/go/internal/usecase"
)

// Handler はパスワードリセット申請機能のハンドラーです
type Handler struct {
	cfg                *config.Config
	sessionManager     *session.Manager
	limiter            *ratelimit.Limiter
	turnstileClient    turnstile.Verifier
	createTokenUseCase *usecase.CreatePasswordResetTokenUsecase
}

// NewHandler は新しいHandlerを作成します
func NewHandler(cfg *config.Config, sessionManager *session.Manager, limiter *ratelimit.Limiter, turnstileClient turnstile.Verifier, createTokenUseCase *usecase.CreatePasswordResetTokenUsecase) *Handler {
	return &Handler{
		cfg:                cfg,
		sessionManager:     sessionManager,
		limiter:            limiter,
		turnstileClient:    turnstileClient,
		createTokenUseCase: createTokenUseCase,
	}
}
