// Package db_person_unarchiveはAnnict DB管理画面で人物のアーカイブを解除する確認画面の
// HTTPハンドラーを定義する。人物のアーカイブ解除はそのアーカイブの削除であるため、解除の実行
// 自体はdb_person_archiveが担う。
package db_person_unarchive

import (
	"net/http"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/redirect"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// HandlerはAnnict DB管理画面の人物の確認HTTPハンドラーが必要とする依存をまとめる。
type Handler struct {
	cfg                       *config.Config
	sessionManager            *session.Manager
	getDBPersonUnarchiveNewUC *usecase.GetDBPersonUnarchiveNewUsecase
}

func NewHandler(
	cfg *config.Config,
	sessionManager *session.Manager,
	getDBPersonUnarchiveNewUC *usecase.GetDBPersonUnarchiveNewUsecase,
) *Handler {
	return &Handler{
		cfg:                       cfg,
		sessionManager:            sessionManager,
		getDBPersonUnarchiveNewUC: getDBPersonUnarchiveNewUC,
	}
}

// dbPersonListPathはリンクがreturn_toを伴わないときに確認画面が読み手を戻す先。Annict DB
// 管理画面の人物一覧。
const dbPersonListPath = "/db/people"

// returnPathは確認画面がキャンセルリンクと送信フォームの双方で読み手を戻す一覧。読み手が
// 来た一覧を持ち回ることで、操作をやめた場合も完了した場合も、読み手が求めていない一覧に着地
// させない。値が無い場合やAnnict DBのパスでない場合は人物一覧にフォールバックする。
func returnPath(r *http.Request) string {
	return redirect.GetSafeDBReturnURL(r.FormValue("return_to"), dbPersonListPath)
}
