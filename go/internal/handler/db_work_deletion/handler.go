// Package db_work_deletion provides the HTTP handler for the delete confirmation screen of a
// work in the Annict DB admin UI. The delete itself is served by db_work, the package that
// owns the work endpoint the confirmation form submits to.
//
// [Ja] db_work_deletion パッケージは Annict DB 管理画面で作品を削除する確認画面の HTTP
// ハンドラーを定義する。削除の実行自体は、確認フォームの送信先である作品のエンドポイントを
// 持つ db_work が担う。
package db_work_deletion

import (
	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// Handler bundles the dependencies the work delete-confirmation HTTP handler needs in the
// Annict DB admin UI.
//
// [Ja] Handler は Annict DB 管理画面の作品削除確認 HTTP ハンドラーが必要とする依存をまとめる。
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
