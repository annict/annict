// Package health はヘルスチェックのハンドラーを提供します
package health

import (
	"github.com/annict/annict/internal/config"
	"github.com/annict/annict/internal/repository"
)

// Handler はヘルスチェック関連のHTTPハンドラーです
type Handler struct {
	cfg      *config.Config
	workRepo *repository.WorkRepository
}

// NewHandler は新しいHandlerを作成します
func NewHandler(cfg *config.Config, workRepo *repository.WorkRepository) *Handler {
	return &Handler{
		cfg:      cfg,
		workRepo: workRepo,
	}
}
