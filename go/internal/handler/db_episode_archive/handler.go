// Package db_episode_archiveはAnnict DB管理画面でエピソードを非公開 (アーカイブ) に
// し、また再公開するHTTPハンドラーを定義する。エンドポイントは作品側が作品のidに紐づくのと
// 同じく、エピソード自身のid (/db/episodes/{id}/archive) に紐づく。
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

// HandlerはAnnict DB管理画面のエピソード非公開HTTPハンドラーが共有する依存をまとめる。
type Handler struct {
	cfg                      *config.Config
	sessionManager           *session.Manager
	flashMgr                 *session.FlashManager
	getDBEpisodeArchiveNewUC *usecase.GetDBEpisodeArchiveNewUsecase
	archiveEpisodeUC         *usecase.ArchiveEpisodeUsecase
	unarchiveEpisodeUC       *usecase.UnarchiveEpisodeUsecase
}

func NewHandler(
	cfg *config.Config,
	sessionManager *session.Manager,
	flashMgr *session.FlashManager,
	getDBEpisodeArchiveNewUC *usecase.GetDBEpisodeArchiveNewUsecase,
	archiveEpisodeUC *usecase.ArchiveEpisodeUsecase,
	unarchiveEpisodeUC *usecase.UnarchiveEpisodeUsecase,
) *Handler {
	return &Handler{
		cfg:                      cfg,
		sessionManager:           sessionManager,
		flashMgr:                 flashMgr,
		getDBEpisodeArchiveNewUC: getDBEpisodeArchiveNewUC,
		archiveEpisodeUC:         archiveEpisodeUC,
		unarchiveEpisodeUC:       unarchiveEpisodeUC,
	}
}

// parseEpisodeIDParamはリクエストが対象とするエピソードを {id} のルートパラメータから
// 読み取り、数値でない場合はfalseを返す。すべてのエンドポイントがここを通るため、不正なidの
// 扱いが確認ページと両方の送信で揃う。
func parseEpisodeIDParam(r *http.Request) (model.EpisodeID, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return 0, false
	}

	return model.EpisodeID(id), true
}

// episodesPathはある作品のエピソード一覧のパスを生成する。確認ページのキャンセル先で
// あり、非公開や再公開が成功したときの着地点でもある。Railsのafter_destroyed_path /
// after_created_path (いずれもdb_episode_list_path) と同じで、編集者が次に確認するのは他の行と
// 並んだ変更後の行であるため。
func episodesPath(workID model.WorkID) string {
	return "/db/works/" + strconv.FormatInt(int64(workID), 10) + "/episodes"
}
