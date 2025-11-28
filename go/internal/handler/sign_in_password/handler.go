// Package sign_in_password はパスワードログイン機能を提供します
package sign_in_password

import (
	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
	"github.com/annict/annict/internal/usecase"
)

// Handler サインイン関連のHTTPハンドラーです
type Handler struct {
	cfg             *config.Config
	userRepo        *repository.UserRepository
	sessionMgr      *session.Manager
	createSessionUC *usecase.CreateSessionUsecase
}

// NewHandler 新しいHandlerを作成します
func NewHandler(cfg *config.Config, userRepo *repository.UserRepository, sessionMgr *session.Manager, createSessionUC *usecase.CreateSessionUsecase) *Handler {
	return &Handler{
		cfg:             cfg,
		userRepo:        userRepo,
		sessionMgr:      sessionMgr,
		createSessionUC: createSessionUC,
	}
}
