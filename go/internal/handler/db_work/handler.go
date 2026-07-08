// Package db_work provides HTTP handlers for work-related features in the Annict DB admin UI.
//
// [Ja] Annict DB 管理画面の作品関連機能を提供する HTTP ハンドラーを定義する。
package db_work

import (
	"net/http"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/image"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// Handler bundles the dependencies shared by work-related HTTP handlers in the Annict DB admin UI.
//
// [Ja] Annict DB 管理画面の作品関連 HTTP ハンドラーが共有する依存をまとめる。
type Handler struct {
	cfg                    *config.Config
	sessionManager         *session.Manager
	flashMgr               *session.FlashManager
	imageHelper            *image.Helper
	getDBWorksUC           *usecase.GetDBWorksUsecase
	getDBWorkFormOptionsUC *usecase.GetDBWorkFormOptionsUsecase
	getDBWorkEditUC        *usecase.GetDBWorkEditUsecase
	createWorkUC           *usecase.CreateWorkUsecase
	updateWorkUC           *usecase.UpdateWorkUsecase
}

func NewHandler(
	cfg *config.Config,
	sessionManager *session.Manager,
	flashMgr *session.FlashManager,
	imageHelper *image.Helper,
	getDBWorksUC *usecase.GetDBWorksUsecase,
	getDBWorkFormOptionsUC *usecase.GetDBWorkFormOptionsUsecase,
	getDBWorkEditUC *usecase.GetDBWorkEditUsecase,
	createWorkUC *usecase.CreateWorkUsecase,
	updateWorkUC *usecase.UpdateWorkUsecase,
) *Handler {
	return &Handler{
		cfg:                    cfg,
		sessionManager:         sessionManager,
		flashMgr:               flashMgr,
		imageHelper:            imageHelper,
		getDBWorksUC:           getDBWorksUC,
		getDBWorkFormOptionsUC: getDBWorkFormOptionsUC,
		getDBWorkEditUC:        getDBWorkEditUC,
		createWorkUC:           createWorkUC,
		updateWorkUC:           updateWorkUC,
	}
}

// parseWorkForm reads the Annict DB work form fields from the request into the shared
// WorkFormInput, so the create and update handlers do not duplicate the field mapping.
//
// [Ja] parseWorkForm はリクエストから Annict DB 作品フォームのフィールドを共有の
// WorkFormInput に読み取る。作成・更新ハンドラーがフィールドの写像を重複させないようにする。
func parseWorkForm(r *http.Request) usecase.WorkFormInput {
	return usecase.WorkFormInput{
		Title:                 r.FormValue("title"),
		TitleKana:             r.FormValue("title_kana"),
		TitleAlter:            r.FormValue("title_alter"),
		TitleEn:               r.FormValue("title_en"),
		TitleAlterEn:          r.FormValue("title_alter_en"),
		Media:                 r.FormValue("media"),
		SeasonYear:            r.FormValue("season_year"),
		SeasonName:            r.FormValue("season_name"),
		StartedOn:             r.FormValue("started_on"),
		EndedOn:               r.FormValue("ended_on"),
		OfficialSiteURL:       r.FormValue("official_site_url"),
		OfficialSiteURLEn:     r.FormValue("official_site_url_en"),
		WikipediaURL:          r.FormValue("wikipedia_url"),
		WikipediaURLEn:        r.FormValue("wikipedia_url_en"),
		TwitterUsername:       r.FormValue("twitter_username"),
		TwitterHashtag:        r.FormValue("twitter_hashtag"),
		ScTid:                 r.FormValue("sc_tid"),
		MalAnimeID:            r.FormValue("mal_anime_id"),
		Synopsis:              r.FormValue("synopsis"),
		SynopsisSource:        r.FormValue("synopsis_source"),
		SynopsisEn:            r.FormValue("synopsis_en"),
		SynopsisSourceEn:      r.FormValue("synopsis_source_en"),
		ManualEpisodesCount:   r.FormValue("manual_episodes_count"),
		StartEpisodeRawNumber: r.FormValue("start_episode_raw_number"),
		NumberFormatID:        r.FormValue("number_format_id"),
		NoEpisodes:            r.FormValue("no_episodes"),
	}
}
