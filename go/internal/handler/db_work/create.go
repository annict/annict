package db_work

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/db_works"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// Create POST /db/works - DB管理画面の作品作成処理
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	input := CreateValidatorInput{
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

	// バリデーション実行
	validator := NewCreateValidator()
	result := validator.Validate(ctx, input)
	if result.FormErrors != nil && result.FormErrors.HasErrors() {
		h.renderNewWithErrors(w, r, input, result.FormErrors)
		return
	}

	// フォーム値をユースケース入力に変換
	ucInput, err := buildCreateWorkInput(input)
	if err != nil {
		slog.ErrorContext(ctx, "入力値の変換に失敗しました", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// ユースケース実行
	ucResult, err := h.createWorkUC.Execute(ctx, ucInput)
	if err != nil {
		slog.ErrorContext(ctx, "作品の作成に失敗しました", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 成功フラッシュメッセージを設定
	_ = h.sessionManager.SetFlash(ctx, w, r, session.FlashSuccess, i18n.T(ctx, "db_works_created"))

	// 作品一覧ページにリダイレクト（将来的に編集ページにリダイレクトを変更予定）
	http.Redirect(w, r, fmt.Sprintf("/db/works?highlight=%d", ucResult.WorkID), http.StatusSeeOther)
}

// renderNewWithErrors はバリデーションエラー時にフォームを再表示します
func (h *Handler) renderNewWithErrors(w http.ResponseWriter, r *http.Request, input CreateValidatorInput, formErrors *session.FormErrors) {
	ctx := r.Context()

	numberFormats, err := h.numberFormatRepo.ListAll(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "NumberFormatの取得エラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	formOptions := viewmodel.NewDBWorkFormOptions(ctx, numberFormats)
	csrfToken := middleware.GetCSRFToken(r, h.sessionManager)

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "db_works_new_title")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusUnprocessableEntity)
	component := layouts.Db(
		ctx,
		meta,
		nil,
		h.cfg.GetAssetVersion(),
		db_works.New(db_works.NewPageData{
			CSRFToken:   csrfToken,
			FormOptions: formOptions,
			FormErrors:  formErrors,
			FormValues: &db_works.FormValues{
				Title:                 input.Title,
				TitleKana:             input.TitleKana,
				TitleAlter:            input.TitleAlter,
				TitleEn:               input.TitleEn,
				TitleAlterEn:          input.TitleAlterEn,
				Media:                 input.Media,
				SeasonYear:            input.SeasonYear,
				SeasonName:            input.SeasonName,
				StartedOn:             input.StartedOn,
				EndedOn:               input.EndedOn,
				OfficialSiteURL:       input.OfficialSiteURL,
				OfficialSiteURLEn:     input.OfficialSiteURLEn,
				WikipediaURL:          input.WikipediaURL,
				WikipediaURLEn:        input.WikipediaURLEn,
				TwitterUsername:       input.TwitterUsername,
				TwitterHashtag:        input.TwitterHashtag,
				ScTid:                 input.ScTid,
				MalAnimeID:            input.MalAnimeID,
				Synopsis:              input.Synopsis,
				SynopsisSource:        input.SynopsisSource,
				SynopsisEn:            input.SynopsisEn,
				SynopsisSourceEn:      input.SynopsisSourceEn,
				ManualEpisodesCount:   input.ManualEpisodesCount,
				StartEpisodeRawNumber: input.StartEpisodeRawNumber,
				NumberFormatID:        input.NumberFormatID,
				NoEpisodes:            input.NoEpisodes,
			},
		}),
	)
	if err := component.Render(ctx, w); err != nil {
		slog.ErrorContext(ctx, "テンプレートのレンダリングエラー", "error", err)
	}
}

// buildCreateWorkInput はフォーム入力値をユースケース入力に変換します
func buildCreateWorkInput(input CreateValidatorInput) (usecase.CreateWorkInput, error) {
	media, err := strconv.ParseInt(input.Media, 10, 32)
	if err != nil {
		return usecase.CreateWorkInput{}, fmt.Errorf("メディア値の変換に失敗: %w", err)
	}

	ucInput := usecase.CreateWorkInput{
		Title:                 strings.TrimSpace(input.Title),
		TitleKana:             strings.TrimSpace(input.TitleKana),
		TitleAlter:            strings.TrimSpace(input.TitleAlter),
		TitleEn:               strings.TrimSpace(input.TitleEn),
		TitleAlterEn:          strings.TrimSpace(input.TitleAlterEn),
		Media:                 int32(media),
		OfficialSiteURL:       strings.TrimSpace(input.OfficialSiteURL),
		OfficialSiteURLEn:     strings.TrimSpace(input.OfficialSiteURLEn),
		WikipediaURL:          strings.TrimSpace(input.WikipediaURL),
		WikipediaURLEn:        strings.TrimSpace(input.WikipediaURLEn),
		Synopsis:              strings.TrimSpace(input.Synopsis),
		SynopsisSource:        strings.TrimSpace(input.SynopsisSource),
		SynopsisEn:            strings.TrimSpace(input.SynopsisEn),
		SynopsisSourceEn:      strings.TrimSpace(input.SynopsisSourceEn),
		NoEpisodes:            input.NoEpisodes == "1",
		StartEpisodeRawNumber: 1.0,
	}

	// season_year
	if input.SeasonYear != "" {
		v, err := strconv.ParseInt(input.SeasonYear, 10, 32)
		if err == nil {
			ucInput.SeasonYear = sql.NullInt32{Int32: int32(v), Valid: true}
		}
	}

	// season_name
	if input.SeasonName != "" {
		v, err := strconv.ParseInt(input.SeasonName, 10, 32)
		if err == nil {
			ucInput.SeasonName = sql.NullInt32{Int32: int32(v), Valid: true}
		}
	}

	// started_on
	if input.StartedOn != "" {
		t, err := time.Parse("2006-01-02", input.StartedOn)
		if err == nil {
			ucInput.StartedOn = sql.NullTime{Time: t, Valid: true}
		}
	}

	// ended_on
	if input.EndedOn != "" {
		t, err := time.Parse("2006-01-02", input.EndedOn)
		if err == nil {
			ucInput.EndedOn = sql.NullTime{Time: t, Valid: true}
		}
	}

	// twitter_username
	if input.TwitterUsername != "" {
		ucInput.TwitterUsername = sql.NullString{String: strings.TrimSpace(input.TwitterUsername), Valid: true}
	}

	// twitter_hashtag
	if input.TwitterHashtag != "" {
		ucInput.TwitterHashtag = sql.NullString{String: strings.TrimSpace(input.TwitterHashtag), Valid: true}
	}

	// sc_tid
	if input.ScTid != "" {
		v, err := strconv.ParseInt(input.ScTid, 10, 32)
		if err == nil {
			ucInput.ScTid = sql.NullInt32{Int32: int32(v), Valid: true}
		}
	}

	// mal_anime_id
	if input.MalAnimeID != "" {
		v, err := strconv.ParseInt(input.MalAnimeID, 10, 32)
		if err == nil {
			ucInput.MalAnimeID = sql.NullInt32{Int32: int32(v), Valid: true}
		}
	}

	// manual_episodes_count
	if input.ManualEpisodesCount != "" {
		v, err := strconv.ParseInt(input.ManualEpisodesCount, 10, 32)
		if err == nil {
			ucInput.ManualEpisodesCount = sql.NullInt32{Int32: int32(v), Valid: true}
		}
	}

	// start_episode_raw_number
	if input.StartEpisodeRawNumber != "" {
		v, err := strconv.ParseFloat(input.StartEpisodeRawNumber, 64)
		if err == nil {
			ucInput.StartEpisodeRawNumber = v
		}
	}

	// number_format_id
	if input.NumberFormatID != "" {
		v, err := strconv.ParseInt(input.NumberFormatID, 10, 64)
		if err == nil {
			ucInput.NumberFormatID = sql.NullInt64{Int64: v, Valid: true}
		}
	}

	return ucInput, nil
}
