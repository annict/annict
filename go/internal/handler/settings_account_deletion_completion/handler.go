// Package settings_account_deletion_completionは退会の完了ページのHTTPハンドラーを定義する。
package settings_account_deletion_completion

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/image"
)

// Handlerは退会の完了ページのHTTPハンドラーが必要とする依存をまとめる。
type Handler struct {
	cfg         *config.Config
	imageHelper *image.Helper
}

func NewHandler(cfg *config.Config, imageHelper *image.Helper) *Handler {
	return &Handler{
		cfg:         cfg,
		imageHelper: imageHelper,
	}
}
