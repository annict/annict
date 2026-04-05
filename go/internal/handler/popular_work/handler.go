// Package popular_work は人気作品表示機能を提供します
package popular_work

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/image"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// Handler は人気作品関連のHTTPハンドラーです
type Handler struct {
	cfg               *config.Config
	getPopularWorksUC *usecase.GetPopularWorksUsecase
	imageHelper       *image.Helper
	sessionManager    *session.Manager
}

// NewHandler は新しいHandlerを作成します
func NewHandler(cfg *config.Config, getPopularWorksUC *usecase.GetPopularWorksUsecase, imageHelper *image.Helper, sessionManager *session.Manager) *Handler {
	return &Handler{
		cfg:               cfg,
		getPopularWorksUC: getPopularWorksUC,
		imageHelper:       imageHelper,
		sessionManager:    sessionManager,
	}
}
