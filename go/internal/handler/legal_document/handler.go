// Package legal_documentは法的文書 (利用規約・プライバシーポリシーなど) のページのHTTPハンドラーを定義する。
package legal_document

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/image"
)

// Handlerは法的文書のページのHTTPハンドラーが必要とする依存をまとめる。
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
