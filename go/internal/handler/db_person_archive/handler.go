// Package db_person_archiveはAnnict DB管理画面で人物をアーカイブするHTTPハンドラーを定義する。
// 人物のアーカイブ解除はそのアーカイブの削除 (DELETE /db/people/:id/archive) として本パッケージが
// 担う。
package db_person_archive

import (
	"net/http"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/redirect"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// HandlerはAnnict DB管理画面の人物のアーカイブHTTPハンドラーが共有する依存をまとめる。
type Handler struct {
	cfg                     *config.Config
	sessionManager          *session.Manager
	flashMgr                *session.FlashManager
	getDBPersonArchiveNewUC *usecase.GetDBPersonArchiveNewUsecase
	archivePersonUC         *usecase.ArchivePersonUsecase
	unarchivePersonUC       *usecase.UnarchivePersonUsecase
}

func NewHandler(
	cfg *config.Config,
	sessionManager *session.Manager,
	flashMgr *session.FlashManager,
	getDBPersonArchiveNewUC *usecase.GetDBPersonArchiveNewUsecase,
	archivePersonUC *usecase.ArchivePersonUsecase,
	unarchivePersonUC *usecase.UnarchivePersonUsecase,
) *Handler {
	return &Handler{
		cfg:                     cfg,
		sessionManager:          sessionManager,
		flashMgr:                flashMgr,
		getDBPersonArchiveNewUC: getDBPersonArchiveNewUC,
		archivePersonUC:         archivePersonUC,
		unarchivePersonUC:       unarchivePersonUC,
	}
}

// dbPersonListPathはリクエストがreturn_toを伴わないときに人物の状態変更が着地する先。
// Annict DB管理画面の人物一覧。
const dbPersonListPath = "/db/people"

// returnPathは本パッケージの画面が読み手を戻す一覧。確認画面は読み手が来た一覧を持ち回る
// ため、確認をやめた場合も変更を完了した場合も、読み手が求めていない一覧に着地させない。
// return_toはリクエスト全体から読むので、確認フォームのフィールドとリンクのクエリ文字列の双方を
// 扱える。値が無い場合やAnnict DBのパスでない場合は人物一覧にフォールバックする。人物一覧の
// ボタンはreturn_toを送らないので人物一覧に着地する。
func returnPath(r *http.Request) string {
	return redirect.GetSafeDBReturnURL(r.FormValue("return_to"), dbPersonListPath)
}
