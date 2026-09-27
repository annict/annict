// Package healthはヘルスチェックのハンドラーを提供します
package health

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/usecase"
)

// Handlerはヘルスチェック関連のHTTPハンドラーです
type Handler struct {
	cfg           *config.Config
	checkHealthUC *usecase.CheckHealthUsecase
}

// NewHandlerは新しいHandlerを作成します
func NewHandler(cfg *config.Config, checkHealthUC *usecase.CheckHealthUsecase) *Handler {
	return &Handler{
		cfg:           cfg,
		checkHealthUC: checkHealthUC,
	}
}
