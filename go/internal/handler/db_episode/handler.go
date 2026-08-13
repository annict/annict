// Package db_episode provides HTTP handlers for episodes in the Annict DB admin UI. The
// episode endpoints live under two URL prefixes: the ones addressing a work's episodes as a
// collection are nested under the work (/db/works/{work_id}/episodes), while the ones
// addressing a single episode are keyed by its own id (/db/episodes/{id}). Both belong to the
// episode resource, so one package holds them, as the Rails Db::EpisodesController does.
//
// [Ja] Package db_episode は Annict DB 管理画面のエピソード関連 HTTP ハンドラーを定義する。
// エピソードのエンドポイントは 2 つの URL 接頭辞に分かれ、ある作品のエピソードをコレクション
// として扱うものは作品の下 (/db/works/{work_id}/episodes) に、単一のエピソードを扱うものは
// エピソード自身の id (/db/episodes/{id}) に紐づく。いずれもエピソードというリソースに属する
// ため、Rails の Db::EpisodesController と同じく 1 つのパッケージにまとめる。
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

// Handler bundles the dependencies shared by the episode HTTP handlers in the Annict DB
// admin UI.
//
// [Ja] Handler は Annict DB 管理画面のエピソード関連 HTTP ハンドラーが共有する依存を
// まとめる。
type Handler struct {
	cfg                *config.Config
	sessionManager     *session.Manager
	flashMgr           *session.FlashManager
	getDBEpisodesUC    *usecase.GetDBEpisodesUsecase
	getDBEpisodeNewUC  *usecase.GetDBEpisodeNewUsecase
	createEpisodesUC   *usecase.CreateEpisodesUsecase
	getDBEpisodeEditUC *usecase.GetDBEpisodeEditUsecase
}

func NewHandler(
	cfg *config.Config,
	sessionManager *session.Manager,
	flashMgr *session.FlashManager,
	getDBEpisodesUC *usecase.GetDBEpisodesUsecase,
	getDBEpisodeNewUC *usecase.GetDBEpisodeNewUsecase,
	createEpisodesUC *usecase.CreateEpisodesUsecase,
	getDBEpisodeEditUC *usecase.GetDBEpisodeEditUsecase,
) *Handler {
	return &Handler{
		cfg:                cfg,
		sessionManager:     sessionManager,
		flashMgr:           flashMgr,
		getDBEpisodesUC:    getDBEpisodesUC,
		getDBEpisodeNewUC:  getDBEpisodeNewUC,
		createEpisodesUC:   createEpisodesUC,
		getDBEpisodeEditUC: getDBEpisodeEditUC,
	}
}

// parseWorkIDParam reads the work the request addresses from the {work_id} route parameter,
// reporting false when it is not a number. Every endpoint nested under a work goes through
// it, so a malformed id is turned away the same way on all of them.
//
// [Ja] parseWorkIDParam はリクエストが対象とする作品を {work_id} のルートパラメータから
// 読み取り、数値でない場合は false を返す。作品の下にネストしたエンドポイントはいずれもここを
// 通るため、不正な id の扱いがすべてで揃う。
func parseWorkIDParam(r *http.Request) (model.WorkID, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "work_id"), 10, 64)
	if err != nil {
		return 0, false
	}

	return model.WorkID(id), true
}

// parseEpisodeIDParam reads the episode the request addresses from the {id} route parameter,
// reporting false when it is not a number. Every endpoint keyed by an episode goes through
// it, so a malformed id is turned away the same way on all of them.
//
// [Ja] parseEpisodeIDParam はリクエストが対象とするエピソードを {id} のルートパラメータから
// 読み取り、数値でない場合は false を返す。エピソード基点のエンドポイントはいずれもここを
// 通るため、不正な id の扱いがすべてで揃う。
func parseEpisodeIDParam(r *http.Request) (model.EpisodeID, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return 0, false
	}

	return model.EpisodeID(id), true
}
