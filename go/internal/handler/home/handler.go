// Package home はホームページのハンドラーを提供します
package home

import (
	"github.com/annict/annict/internal/config"
)

// Handler はホームページ関連のHTTPハンドラーです
type Handler struct {
	cfg *config.Config
}

// NewHandler は新しいHandlerを作成します
func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		cfg: cfg,
	}
}
