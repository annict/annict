// Package db_work_archive provides HTTP handlers for archiving (unpublishing) a work in
// the Annict DB admin UI.
//
// [Ja] db_work_archive パッケージは Annict DB 管理画面で作品を非公開 (アーカイブ) にする
// HTTP ハンドラーを定義する。
package db_work_archive

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// Handler bundles the dependencies shared by the work archive HTTP handlers in the Annict
// DB admin UI.
//
// [Ja] Handler は Annict DB 管理画面の作品非公開 HTTP ハンドラーが共有する依存をまとめる。
type Handler struct {
	cfg                   *config.Config
	sessionManager        *session.Manager
	flashMgr              *session.FlashManager
	getDBWorkArchiveNewUC *usecase.GetDBWorkArchiveNewUsecase
	archiveWorkUC         *usecase.ArchiveWorkUsecase
	unarchiveWorkUC       *usecase.UnarchiveWorkUsecase
}

func NewHandler(
	cfg *config.Config,
	sessionManager *session.Manager,
	flashMgr *session.FlashManager,
	getDBWorkArchiveNewUC *usecase.GetDBWorkArchiveNewUsecase,
	archiveWorkUC *usecase.ArchiveWorkUsecase,
	unarchiveWorkUC *usecase.UnarchiveWorkUsecase,
) *Handler {
	return &Handler{
		cfg:                   cfg,
		sessionManager:        sessionManager,
		flashMgr:              flashMgr,
		getDBWorkArchiveNewUC: getDBWorkArchiveNewUC,
		archiveWorkUC:         archiveWorkUC,
		unarchiveWorkUC:       unarchiveWorkUC,
	}
}
