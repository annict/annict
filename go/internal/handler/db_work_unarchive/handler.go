// Package db_work_unarchiveはAnnict DB管理画面で作品を公開 (アーカイブ解除) する
// 確認画面のHTTPハンドラーを定義する。作品の公開はそのアーカイブの削除であるため、公開の
// 実行自体はdb_work_archiveが担う。
package db_work_unarchive

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// HandlerはAnnict DB管理画面の作品公開確認HTTPハンドラーが必要とする依存をまとめる。
type Handler struct {
	cfg                     *config.Config
	sessionManager          *session.Manager
	getDBWorkUnarchiveNewUC *usecase.GetDBWorkUnarchiveNewUsecase
}

func NewHandler(
	cfg *config.Config,
	sessionManager *session.Manager,
	getDBWorkUnarchiveNewUC *usecase.GetDBWorkUnarchiveNewUsecase,
) *Handler {
	return &Handler{
		cfg:                     cfg,
		sessionManager:          sessionManager,
		getDBWorkUnarchiveNewUC: getDBWorkUnarchiveNewUC,
	}
}
