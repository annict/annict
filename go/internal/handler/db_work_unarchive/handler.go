// Package db_work_unarchive provides the HTTP handler for the publish (un-archive)
// confirmation screen of a work in the Annict DB admin UI. The re-publish itself is served by
// db_work_archive, since publishing a work is deleting its archive.
//
// [Ja] db_work_unarchive パッケージは Annict DB 管理画面で作品を公開 (アーカイブ解除) する
// 確認画面の HTTP ハンドラーを定義する。作品の公開はそのアーカイブの削除であるため、公開の
// 実行自体は db_work_archive が担う。
package db_work_unarchive

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// Handler bundles the dependencies the work publish-confirmation HTTP handler needs in the
// Annict DB admin UI.
//
// [Ja] Handler は Annict DB 管理画面の作品公開確認 HTTP ハンドラーが必要とする依存をまとめる。
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
