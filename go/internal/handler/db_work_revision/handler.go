// Package db_work_revisionはAnnict DB管理画面の作品の履歴のHTTPハンドラーを定義する。
package db_work_revision

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/image"
	"github.com/annict/annict/go/internal/usecase"
)

// HandlerはAnnict DB管理画面の作品の履歴のHTTPハンドラーが共有する依存をまとめる。
type Handler struct {
	cfg                  *config.Config
	imageHelper          *image.Helper
	getDBWorkRevisionsUC *usecase.GetDBWorkRevisionsUsecase
}

func NewHandler(cfg *config.Config, imageHelper *image.Helper, getDBWorkRevisionsUC *usecase.GetDBWorkRevisionsUsecase) *Handler {
	return &Handler{
		cfg:                  cfg,
		imageHelper:          imageHelper,
		getDBWorkRevisionsUC: getDBWorkRevisionsUC,
	}
}
