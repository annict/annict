// Package db_work_deletionはAnnict DB管理画面で作品を削除する確認画面のHTTP
// ハンドラーを定義する。削除の実行自体は、確認フォームの送信先である作品のエンドポイントを
// 持つdb_workが担う。
package db_work_deletion

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// HandlerはAnnict DB管理画面の作品削除確認HTTPハンドラーが必要とする依存をまとめる。
type Handler struct {
	cfg                    *config.Config
	sessionManager         *session.Manager
	getDBWorkDeletionNewUC *usecase.GetDBWorkDeletionNewUsecase
}

func NewHandler(
	cfg *config.Config,
	sessionManager *session.Manager,
	getDBWorkDeletionNewUC *usecase.GetDBWorkDeletionNewUsecase,
) *Handler {
	return &Handler{
		cfg:                    cfg,
		sessionManager:         sessionManager,
		getDBWorkDeletionNewUC: getDBWorkDeletionNewUC,
	}
}
