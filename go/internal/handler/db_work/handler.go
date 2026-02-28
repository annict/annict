// Package db_work はDB管理画面の作品関連機能を提供します
package db_work

import (
	"database/sql"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// Handler はDB管理画面の作品関連のHTTPハンドラーです
type Handler struct {
	cfg              *config.Config
	db               *sql.DB
	workRepo         *repository.WorkRepository
	numberFormatRepo *repository.NumberFormatRepository
	sessionManager   *session.Manager
	createWorkUC     *usecase.CreateWorkUsecase
}

// NewHandler は新しいHandlerを作成します
func NewHandler(
	cfg *config.Config,
	db *sql.DB,
	workRepo *repository.WorkRepository,
	numberFormatRepo *repository.NumberFormatRepository,
	sessionManager *session.Manager,
) *Handler {
	return &Handler{
		cfg:              cfg,
		db:               db,
		workRepo:         workRepo,
		numberFormatRepo: numberFormatRepo,
		sessionManager:   sessionManager,
		createWorkUC:     usecase.NewCreateWorkUsecase(db, workRepo),
	}
}
