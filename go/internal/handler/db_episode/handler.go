// Package db_episodeはAnnict DB管理画面のエピソード関連HTTPハンドラーを定義する。
// エピソードのエンドポイントは2つのURL接頭辞に分かれ、ある作品のエピソードをコレクション
// として扱うものは作品の下 (/db/works/{work_id}/episodes) に、単一のエピソードを扱うものは
// エピソード自身のid (/db/episodes/{id}) に紐づく。いずれもエピソードというリソースに属する
// ため、RailsのDb::EpisodesControllerと同じく1つのパッケージにまとめる。
package db_episode

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// HandlerはAnnict DB管理画面のエピソード関連HTTPハンドラーが共有する依存を
// まとめる。
type Handler struct {
	cfg                *config.Config
	sessionManager     *session.Manager
	flashMgr           *session.FlashManager
	getDBEpisodesUC    *usecase.GetDBEpisodesUsecase
	getDBEpisodeNewUC  *usecase.GetDBEpisodeNewUsecase
	createEpisodesUC   *usecase.CreateEpisodesUsecase
	getDBEpisodeEditUC *usecase.GetDBEpisodeEditUsecase
	updateEpisodeUC    *usecase.UpdateEpisodeUsecase
	deleteEpisodeUC    *usecase.DeleteEpisodeUsecase
}

func NewHandler(
	cfg *config.Config,
	sessionManager *session.Manager,
	flashMgr *session.FlashManager,
	getDBEpisodesUC *usecase.GetDBEpisodesUsecase,
	getDBEpisodeNewUC *usecase.GetDBEpisodeNewUsecase,
	createEpisodesUC *usecase.CreateEpisodesUsecase,
	getDBEpisodeEditUC *usecase.GetDBEpisodeEditUsecase,
	updateEpisodeUC *usecase.UpdateEpisodeUsecase,
	deleteEpisodeUC *usecase.DeleteEpisodeUsecase,
) *Handler {
	return &Handler{
		cfg:                cfg,
		sessionManager:     sessionManager,
		flashMgr:           flashMgr,
		getDBEpisodesUC:    getDBEpisodesUC,
		getDBEpisodeNewUC:  getDBEpisodeNewUC,
		createEpisodesUC:   createEpisodesUC,
		getDBEpisodeEditUC: getDBEpisodeEditUC,
		updateEpisodeUC:    updateEpisodeUC,
		deleteEpisodeUC:    deleteEpisodeUC,
	}
}

// parseWorkIDParamはリクエストが対象とする作品を {work_id} のルートパラメータから
// 読み取り、数値でない場合はfalseを返す。作品の下にネストしたエンドポイントはいずれもここを
// 通るため、不正なidの扱いがすべてで揃う。
func parseWorkIDParam(r *http.Request) (model.WorkID, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "work_id"), 10, 64)
	if err != nil {
		return 0, false
	}

	return model.WorkID(id), true
}

// parseEpisodeIDParamはリクエストが対象とするエピソードを {id} のルートパラメータから
// 読み取り、数値でない場合はfalseを返す。エピソード基点のエンドポイントはいずれもここを
// 通るため、不正なidの扱いがすべてで揃う。
func parseEpisodeIDParam(r *http.Request) (model.EpisodeID, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return 0, false
	}

	return model.EpisodeID(id), true
}
