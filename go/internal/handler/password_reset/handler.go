// Package password_reset はパスワードリセット機能を提供します
package password_reset

import (
	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/ratelimit"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/turnstile"
	"github.com/annict/annict/internal/usecase"
)

// Handler はパスワードリセット申請機能のハンドラーです
type Handler struct {
	cfg                *config.Config
	userRepo           *repository.UserRepository
	sessionManager     *session.Manager
	limiter            *ratelimit.Limiter
	turnstileClient    turnstile.Verifier
	createTokenUseCase *usecase.CreatePasswordResetTokenUsecase
}

// NewHandler は新しいHandlerを作成します
func NewHandler(cfg *config.Config, userRepo *repository.UserRepository, sessionManager *session.Manager, limiter *ratelimit.Limiter, turnstileClient turnstile.Verifier, createTokenUseCase *usecase.CreatePasswordResetTokenUsecase) *Handler {
	return &Handler{
		cfg:                cfg,
		userRepo:           userRepo,
		sessionManager:     sessionManager,
		limiter:            limiter,
		turnstileClient:    turnstileClient,
		createTokenUseCase: createTokenUseCase,
	}
}
