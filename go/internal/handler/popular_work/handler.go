// Package popular_work は人気作品表示機能を提供します
package popular_work

import (
	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/image"
	"github.com/annict/annict/internal/repository"
	"github.com/annict/annict/internal/session"
)

// Handler は人気作品関連のHTTPハンドラーです
type Handler struct {
	cfg            *config.Config
	workRepo       *repository.WorkRepository
	imageHelper    *image.Helper
	sessionManager *session.Manager
}

// NewHandler は新しいHandlerを作成します
func NewHandler(cfg *config.Config, workRepo *repository.WorkRepository, imageHelper *image.Helper, sessionManager *session.Manager) *Handler {
	return &Handler{
		cfg:            cfg,
		workRepo:       workRepo,
		imageHelper:    imageHelper,
		sessionManager: sessionManager,
	}
}
