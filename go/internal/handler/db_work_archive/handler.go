// Package db_work_archiveはAnnict DB管理画面で作品を非公開 (アーカイブ) にする
// HTTPハンドラーを定義する。
package db_work_archive

import (
	"net/http"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/redirect"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// HandlerはAnnict DB管理画面の作品非公開HTTPハンドラーが共有する依存をまとめる。
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

// dbWorkListPathはリクエストがreturn_toを伴わないときに作品の状態変更が着地する先。
// Annict DB管理画面の作品一覧。
const dbWorkListPath = "/db/works"

// returnPathは本パッケージの画面が読み手を戻す一覧。確認画面は読み手が来た一覧を持ち回る
// ため、確認をやめた場合も変更を完了した場合も、読み手が求めていない一覧に着地させない。
// return_toはリクエスト全体から読むので、確認フォームのフィールドとリンクのクエリ文字列の双方を
// 扱える。値が無い場合やAnnict DBのパスでない場合は作品一覧にフォールバックする。作品一覧の
// ボタンはreturn_toを送らないので従来どおりそこに着地する。
func returnPath(r *http.Request) string {
	return redirect.GetSafeDBReturnURL(r.FormValue("return_to"), dbWorkListPath)
}
