package db_work

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/session"
)

// 許可されたメディア種別の値
var allowedMediaValues = map[string]bool{
	"0": true, // other
	"1": true, // tv
	"2": true, // ova
	"3": true, // movie
	"4": true, // web
}

// CreateValidator は作品作成フォームのバリデーションを行う
type CreateValidator struct{}

// NewCreateValidator は CreateValidator を生成する
func NewCreateValidator() *CreateValidator {
	return &CreateValidator{}
}

// CreateValidatorInput はバリデーションの入力パラメータ
type CreateValidatorInput struct {
	Title                 string
	TitleKana             string
	TitleAlter            string
	TitleEn               string
	TitleAlterEn          string
	Media                 string
	SeasonYear            string
	SeasonName            string
	StartedOn             string
	EndedOn               string
	OfficialSiteURL       string
	OfficialSiteURLEn     string
	WikipediaURL          string
	WikipediaURLEn        string
	TwitterUsername       string
	TwitterHashtag        string
	ScTid                 string
	MalAnimeID            string
	Synopsis              string
	SynopsisSource        string
	SynopsisEn            string
	SynopsisSourceEn      string
	ManualEpisodesCount   string
	StartEpisodeRawNumber string
	NumberFormatID        string
	NoEpisodes            string
}

// CreateValidatorResult はバリデーションの結果
type CreateValidatorResult struct {
	FormErrors *session.FormErrors
}

// Validate はバリデーションを行う
func (v *CreateValidator) Validate(ctx context.Context, input CreateValidatorInput) *CreateValidatorResult {
	formErrors := &session.FormErrors{}

	// タイトル: 必須
	if strings.TrimSpace(input.Title) == "" {
		formErrors.AddFieldError("title", i18n.T(ctx, "db_works_error_title_required"))
	}

	// メディア: 必須 + 許可された値
	if strings.TrimSpace(input.Media) == "" {
		formErrors.AddFieldError("media", i18n.T(ctx, "db_works_error_media_required"))
	} else if !allowedMediaValues[input.Media] {
		formErrors.AddFieldError("media", i18n.T(ctx, "db_works_error_media_invalid"))
	}

	// URL形式チェック（空の場合はスキップ）
	validateOptionalURL(ctx, formErrors, "official_site_url", input.OfficialSiteURL)
	validateOptionalURL(ctx, formErrors, "official_site_url_en", input.OfficialSiteURLEn)
	validateOptionalURL(ctx, formErrors, "wikipedia_url", input.WikipediaURL)
	validateOptionalURL(ctx, formErrors, "wikipedia_url_en", input.WikipediaURLEn)

	// sc_tid: 整数（空の場合はスキップ）
	if input.ScTid != "" {
		if _, err := strconv.Atoi(input.ScTid); err != nil {
			formErrors.AddFieldError("sc_tid", i18n.T(ctx, "db_works_error_sc_tid_invalid"))
		}
	}

	// mal_anime_id: 整数（空の場合はスキップ）
	if input.MalAnimeID != "" {
		if _, err := strconv.Atoi(input.MalAnimeID); err != nil {
			formErrors.AddFieldError("mal_anime_id", i18n.T(ctx, "db_works_error_mal_anime_id_invalid"))
		}
	}

	// あらすじと出典のペアチェック
	validatePresencePair(ctx, formErrors, "synopsis_source", input.Synopsis, input.SynopsisSource, "db_works_error_synopsis_source_required")
	validatePresencePair(ctx, formErrors, "synopsis_source_en", input.SynopsisEn, input.SynopsisSourceEn, "db_works_error_synopsis_source_en_required")

	if formErrors.HasErrors() {
		return &CreateValidatorResult{FormErrors: formErrors}
	}

	return &CreateValidatorResult{}
}

// validateOptionalURL はURLが空でない場合にURL形式をチェックする
func validateOptionalURL(ctx context.Context, formErrors *session.FormErrors, field, value string) {
	if value == "" {
		return
	}
	u, err := url.ParseRequestURI(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		formErrors.AddFieldError(field, i18n.T(ctx, "db_works_error_url_invalid"))
	}
}

// validatePresencePair はペアの片方がある場合に、もう片方も必須にする
func validatePresencePair(ctx context.Context, formErrors *session.FormErrors, sourceField, content, source, errKey string) {
	if strings.TrimSpace(content) != "" && strings.TrimSpace(source) == "" {
		formErrors.AddFieldError(sourceField, i18n.T(ctx, errKey))
	}
}
