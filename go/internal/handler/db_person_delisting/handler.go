// Package db_person_delistingはAnnict DB管理画面で人物の掲載を停止するHTTPハンドラーを定義する。
// 人物の掲載停止の解除はその掲載停止の削除 (DELETE /db/people/:id/delisting) として本パッケージが
// 担う。
package db_person_delisting

import (
	"net/http"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/redirect"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// HandlerはAnnict DB管理画面の人物の掲載停止HTTPハンドラーが共有する依存をまとめる。
type Handler struct {
	cfg                         *config.Config
	sessionManager              *session.Manager
	flashMgr                    *session.FlashManager
	getDBPersonDelistingNewUC   *usecase.GetDBPersonDelistingNewUsecase
	getDBPersonUndelistingNewUC *usecase.GetDBPersonUndelistingNewUsecase
	delistPersonUC              *usecase.DelistPersonUsecase
	undelistPersonUC            *usecase.UndelistPersonUsecase
}

func NewHandler(
	cfg *config.Config,
	sessionManager *session.Manager,
	flashMgr *session.FlashManager,
	getDBPersonDelistingNewUC *usecase.GetDBPersonDelistingNewUsecase,
	getDBPersonUndelistingNewUC *usecase.GetDBPersonUndelistingNewUsecase,
	delistPersonUC *usecase.DelistPersonUsecase,
	undelistPersonUC *usecase.UndelistPersonUsecase,
) *Handler {
	return &Handler{
		cfg:                         cfg,
		sessionManager:              sessionManager,
		flashMgr:                    flashMgr,
		getDBPersonDelistingNewUC:   getDBPersonDelistingNewUC,
		getDBPersonUndelistingNewUC: getDBPersonUndelistingNewUC,
		delistPersonUC:              delistPersonUC,
		undelistPersonUC:            undelistPersonUC,
	}
}

// dbPersonListPathはリクエストがreturn_toを伴わないときに人物の掲載停止・解除が着地する先。
// Annict DB管理画面の人物一覧。
const dbPersonListPath = "/db/people"

// returnPathは本パッケージの画面が読み手を戻す一覧。確認画面は読み手が来た一覧を持ち回る
// ため、確認をやめた場合も変更を完了した場合も、読み手が求めていない一覧に着地させない。
// 値が無い場合やAnnict DBのパスでない場合は人物一覧にフォールバックする。
func returnPath(r *http.Request) string {
	return redirect.GetSafeDBReturnURL(r.FormValue("return_to"), dbPersonListPath)
}
