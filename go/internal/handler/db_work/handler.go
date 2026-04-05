// Package db_work はDB管理画面の作品関連機能を提供します
package db_work

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// Handler はDB管理画面の作品関連のHTTPハンドラーです
type Handler struct {
	cfg                    *config.Config
	sessionManager         *session.Manager
	listDbWorksUC          *usecase.ListDbWorksUsecase
	getDbWorkFormOptionsUC *usecase.GetDbWorkFormOptionsUsecase
	createWorkUC           *usecase.CreateWorkUsecase
}

// NewHandler は新しいHandlerを作成します
func NewHandler(
	cfg *config.Config,
	sessionManager *session.Manager,
	listDbWorksUC *usecase.ListDbWorksUsecase,
	getDbWorkFormOptionsUC *usecase.GetDbWorkFormOptionsUsecase,
	createWorkUC *usecase.CreateWorkUsecase,
) *Handler {
	return &Handler{
		cfg:                    cfg,
		sessionManager:         sessionManager,
		listDbWorksUC:          listDbWorksUC,
		getDbWorkFormOptionsUC: getDbWorkFormOptionsUC,
		createWorkUC:           createWorkUC,
	}
}
