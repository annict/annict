// Package db_work provides HTTP handlers for work-related features in the Annict DB admin UI.
//
// [Ja] Annict DB 管理画面の作品関連機能を提供する HTTP ハンドラーを定義する。
package db_work

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// Handler bundles the dependencies shared by work-related HTTP handlers in the Annict DB admin UI.
// [Ja] Annict DB 管理画面の作品関連 HTTP ハンドラーが共有する依存をまとめる。
type Handler struct {
	cfg                    *config.Config
	sessionManager         *session.Manager
	flashMgr               *session.FlashManager
	listDbWorksUC          *usecase.ListDbWorksUsecase
	getDbWorkFormOptionsUC *usecase.GetDbWorkFormOptionsUsecase
	createWorkUC           *usecase.CreateWorkUsecase
}

func NewHandler(
	cfg *config.Config,
	sessionManager *session.Manager,
	flashMgr *session.FlashManager,
	listDbWorksUC *usecase.ListDbWorksUsecase,
	getDbWorkFormOptionsUC *usecase.GetDbWorkFormOptionsUsecase,
	createWorkUC *usecase.CreateWorkUsecase,
) *Handler {
	return &Handler{
		cfg:                    cfg,
		sessionManager:         sessionManager,
		flashMgr:               flashMgr,
		listDbWorksUC:          listDbWorksUC,
		getDbWorkFormOptionsUC: getDbWorkFormOptionsUC,
		createWorkUC:           createWorkUC,
	}
}
