package db_work

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_works"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// Create processes the work creation request in the Annict DB admin UI (POST /db/works).
//
// [Ja] Annict DB 管理画面の作品作成リクエスト (POST /db/works) を処理する。
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	input := usecase.CreateWorkInput{
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

	output, err := h.createWorkUC.Execute(ctx, input)
	if err != nil {
		if ve := model.AsValidationError(err); ve != nil {
			h.renderNewWithErrors(w, r, input, ve)
			return
		}
		slog.ErrorContext(ctx, "作品の作成に失敗しました", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.flashMgr.SetSuccess(w, i18n.T(ctx, "flash_db_work_created"))

	// TODO: redirect to the edit page once it is implemented; for now redirect back to the index.
	//
	// [Ja] TODO: 編集ページ実装後は編集ページへリダイレクトする。現状は一覧ページへリダイレクトする。
	http.Redirect(w, r, fmt.Sprintf("/db/works?highlight=%d", output.WorkID), http.StatusSeeOther)
}

// renderNewWithErrors re-renders the new-work form with validation errors and the previously submitted values.
//
// [Ja] バリデーションエラーと送信済みの入力値を保持したまま新規作成フォームを再描画する。
func (h *Handler) renderNewWithErrors(w http.ResponseWriter, r *http.Request, input usecase.CreateWorkInput, formErrors *model.ValidationError) {
	ctx := r.Context()

	optionsResult, err := h.getDbWorkFormOptionsUC.Execute(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "NumberFormatの取得エラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	formOptions := viewmodel.NewDBWorkFormOptions(ctx, optionsResult.NumberFormats)
	csrfToken := middleware.GetCSRFToken(r, h.sessionManager)

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "db_works_new_title")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusUnprocessableEntity)
	component := layouts.Db(
		meta,
		h.cfg.GetAssetVersion(),
		db_works.New(db_works.NewPageData{
			CSRFToken:   csrfToken,
			FormOptions: formOptions,
			FormErrors:  formErrors,
			FormInput:   viewmodel.NewDBWorkFormInput(input),
		}),
	)
	if err := component.Render(ctx, w); err != nil {
		slog.ErrorContext(ctx, "テンプレートのレンダリングエラー", "error", err)
	}
}
