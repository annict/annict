// Package db_work はDB管理画面の作品関連機能を提供します
package db_work

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
)

// Handler はDB管理画面の作品関連のHTTPハンドラーです
type Handler struct {
	cfg            *config.Config
	workRepo       *repository.WorkRepository
	sessionManager *session.Manager
}

// NewHandler は新しいHandlerを作成します
func NewHandler(cfg *config.Config, workRepo *repository.WorkRepository, sessionManager *session.Manager) *Handler {
	return &Handler{
		cfg:            cfg,
		workRepo:       workRepo,
		sessionManager: sessionManager,
	}
}
