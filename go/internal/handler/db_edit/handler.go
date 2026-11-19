// Package db_editはAnnict DB全体の編集履歴のHTTPハンドラーを定義する。
package db_edit

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/image"
	"github.com/annict/annict/go/internal/usecase"
)

// HandlerはAnnict DB全体の編集履歴のHTTPハンドラーが共有する依存をまとめる。
type Handler struct {
	cfg          *config.Config
	imageHelper  *image.Helper
	getDBEditsUC *usecase.GetDBEditsUsecase
}

func NewHandler(cfg *config.Config, imageHelper *image.Helper, getDBEditsUC *usecase.GetDBEditsUsecase) *Handler {
	return &Handler{
		cfg:          cfg,
		imageHelper:  imageHelper,
		getDBEditsUC: getDBEditsUC,
	}
}
