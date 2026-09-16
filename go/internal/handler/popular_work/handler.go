// Package popular_workは人気作品表示機能を提供します
package popular_work

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/image"
	"github.com/annict/annict/go/internal/usecase"
)

// Handlerは人気作品関連のHTTPハンドラーです
type Handler struct {
	cfg               *config.Config
	getPopularWorksUC *usecase.GetPopularWorksUsecase
	imageHelper       *image.Helper
}

// NewHandlerは新しいHandlerを作成します
func NewHandler(cfg *config.Config, getPopularWorksUC *usecase.GetPopularWorksUsecase, imageHelper *image.Helper) *Handler {
	return &Handler{
		cfg:               cfg,
		getPopularWorksUC: getPopularWorksUC,
		imageHelper:       imageHelper,
	}
}
