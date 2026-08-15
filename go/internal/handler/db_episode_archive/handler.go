// Package db_episode_archive provides HTTP handlers for archiving (unpublishing) an episode in
// the Annict DB admin UI. The endpoints are keyed by the episode's own id
// (/db/episodes/{id}/archive), as the work counterpart is keyed by the work's.
//
// [Ja] db_episode_archive パッケージは Annict DB 管理画面でエピソードを非公開 (アーカイブ) に
// する HTTP ハンドラーを定義する。エンドポイントは作品側が作品の id に紐づくのと同じく、
// エピソード自身の id (/db/episodes/{id}/archive) に紐づく。
package db_episode_archive

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// Handler bundles the dependencies shared by the episode archive HTTP handlers in the Annict
// DB admin UI.
//
// [Ja] Handler は Annict DB 管理画面のエピソード非公開 HTTP ハンドラーが共有する依存をまとめる。
type Handler struct {
	cfg                      *config.Config
	sessionManager           *session.Manager
	flashMgr                 *session.FlashManager
	getDBEpisodeArchiveNewUC *usecase.GetDBEpisodeArchiveNewUsecase
	archiveEpisodeUC         *usecase.ArchiveEpisodeUsecase
}

func NewHandler(
	cfg *config.Config,
	sessionManager *session.Manager,
	flashMgr *session.FlashManager,
	getDBEpisodeArchiveNewUC *usecase.GetDBEpisodeArchiveNewUsecase,
	archiveEpisodeUC *usecase.ArchiveEpisodeUsecase,
) *Handler {
	return &Handler{
		cfg:                      cfg,
		sessionManager:           sessionManager,
		flashMgr:                 flashMgr,
		getDBEpisodeArchiveNewUC: getDBEpisodeArchiveNewUC,
		archiveEpisodeUC:         archiveEpisodeUC,
	}
}

// parseEpisodeIDParam reads the episode the request addresses from the {id} route parameter,
// reporting false when it is not a number. Both endpoints go through it, so a malformed id is
// turned away the same way on the confirmation page and on the submit.
//
// [Ja] parseEpisodeIDParam はリクエストが対象とするエピソードを {id} のルートパラメータから
// 読み取り、数値でない場合は false を返す。両エンドポイントがここを通るため、不正な id の扱いが
// 確認ページと送信で揃う。
func parseEpisodeIDParam(r *http.Request) (model.EpisodeID, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return 0, false
	}

	return model.EpisodeID(id), true
}

// episodesPath builds the path of a work's episode list, which the confirmation page cancels
// back to and a successful archive lands on. It matches the Rails after_destroyed_path
// (db_episode_list_path): the archived row among the others is what the editor checks next.
//
// [Ja] episodesPath はある作品のエピソード一覧のパスを生成する。確認ページのキャンセル先で
// あり、非公開が成功したときの着地点でもある。Rails の after_destroyed_path
// (db_episode_list_path) と同じで、編集者が次に確認するのは他の行と並んだ非公開後の行である
// ため。
func episodesPath(workID model.WorkID) string {
	return "/db/works/" + strconv.FormatInt(int64(workID), 10) + "/episodes"
}
