// Package db_episode_revisionはAnnict DB管理画面のエピソードの履歴のHTTPハンドラーを定義する。
package db_episode_revision

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/image"
	"github.com/annict/annict/go/internal/usecase"
)

// HandlerはAnnict DB管理画面のエピソードの履歴のHTTPハンドラーが共有する依存をまとめる。
type Handler struct {
	cfg                     *config.Config
	imageHelper             *image.Helper
	getDBEpisodeRevisionsUC *usecase.GetDBEpisodeRevisionsUsecase
}

func NewHandler(cfg *config.Config, imageHelper *image.Helper, getDBEpisodeRevisionsUC *usecase.GetDBEpisodeRevisionsUsecase) *Handler {
	return &Handler{
		cfg:                     cfg,
		imageHelper:             imageHelper,
		getDBEpisodeRevisionsUC: getDBEpisodeRevisionsUC,
	}
}
