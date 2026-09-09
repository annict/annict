// Package db_work_archive provides HTTP handlers for archiving (unpublishing) a work in
// the Annict DB admin UI.
//
// [Ja] db_work_archive パッケージは Annict DB 管理画面で作品を非公開 (アーカイブ) にする
// HTTP ハンドラーを定義する。
package db_work_archive

import (
	"net/http"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/redirect"
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

// dbWorkListPath is where a work state change lands when the request names no return_to: the
// work list of the Annict DB admin UI.
//
// [Ja] dbWorkListPath はリクエストが return_to を伴わないときに作品の状態変更が着地する先。
// Annict DB 管理画面の作品一覧。
const dbWorkListPath = "/db/works"

// returnPath is the listing this package's screens return the reader to. A confirmation screen
// carries the listing the reader came from, so neither leaving the confirmation nor completing
// the change drops them on a list they did not ask for. return_to is read from the request as a
// whole, which covers both the confirmation form field and a link query string, and an absent or
// non-Annict-DB value falls back to the work list. The work list's own buttons submit no
// return_to and keep landing there.
//
// [Ja] returnPath は本パッケージの画面が読み手を戻す一覧。確認画面は読み手が来た一覧を持ち回る
// ため、確認をやめた場合も変更を完了した場合も、読み手が求めていない一覧に着地させない。
// return_to はリクエスト全体から読むので、確認フォームのフィールドとリンクのクエリ文字列の双方を
// 扱える。値が無い場合や Annict DB のパスでない場合は作品一覧にフォールバックする。作品一覧の
// ボタンは return_to を送らないので従来どおりそこに着地する。
func returnPath(r *http.Request) string {
	return redirect.GetSafeDBReturnURL(r.FormValue("return_to"), dbWorkListPath)
}
