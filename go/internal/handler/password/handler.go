// Package password はパスワード変更機能を提供します
package password

import (
	"database/sql"

	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/ratelimit"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/usecase"
)

// Handler はパスワード編集・更新のHTTPハンドラーです
type Handler struct {
	cfg                          *config.Config
	passwordResetTokenRepository *repository.PasswordResetTokenRepository
	sessionManager               *session.Manager
	limiter                      *ratelimit.Limiter
	updatePasswordUseCase        *usecase.UpdatePasswordResetUsecase
}

// NewHandler は新しいHandlerを作成します
func NewHandler(cfg *config.Config, db *sql.DB, passwordResetTokenRepository *repository.PasswordResetTokenRepository, sessionManager *session.Manager, limiter *ratelimit.Limiter, updatePasswordUseCase *usecase.UpdatePasswordResetUsecase) *Handler {
	return &Handler{
		cfg:                          cfg,
		passwordResetTokenRepository: passwordResetTokenRepository,
		sessionManager:               sessionManager,
		limiter:                      limiter,
		updatePasswordUseCase:        updatePasswordUseCase,
	}
}
