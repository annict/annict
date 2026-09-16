// Package homeはホームページのハンドラーを提供します
package home

import (
	"github.com/annict/annict/go/internal/config"
)

// Handlerはホームページ関連のHTTPハンドラーです
type Handler struct {
	cfg *config.Config
}

// NewHandlerは新しいHandlerを作成します
func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		cfg: cfg,
	}
}
