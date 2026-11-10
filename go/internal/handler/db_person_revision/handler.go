// Package db_person_revisionはAnnict DB管理画面の人物の履歴のHTTPハンドラーを定義する。
package db_person_revision

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/image"
	"github.com/annict/annict/go/internal/usecase"
)

// HandlerはAnnict DB管理画面の人物の履歴のHTTPハンドラーが共有する依存をまとめる。
type Handler struct {
	cfg                    *config.Config
	imageHelper            *image.Helper
	getDBPersonRevisionsUC *usecase.GetDBPersonRevisionsUsecase
}

func NewHandler(cfg *config.Config, imageHelper *image.Helper, getDBPersonRevisionsUC *usecase.GetDBPersonRevisionsUsecase) *Handler {
	return &Handler{
		cfg:                    cfg,
		imageHelper:            imageHelper,
		getDBPersonRevisionsUC: getDBPersonRevisionsUC,
	}
}
