package validator

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// 許可されたメディア種別の値
var allowedMediaValues = map[string]bool{
	"0": true, // other
	"1": true, // tv
	"2": true, // ova
	"3": true, // movie
	"4": true, // web
}

// DbWorkCreateValidator は作品作成フォームのバリデーションを行う
type DbWorkCreateValidator struct{}

// NewDbWorkCreateValidator は DbWorkCreateValidator を生成する
func NewDbWorkCreateValidator() *DbWorkCreateValidator {
	return &DbWorkCreateValidator{}
}

// DbWorkCreateValidatorInput はバリデーションの入力パラメータ
type DbWorkCreateValidatorInput struct {
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

// Validate はバリデーションを行う
func (v *DbWorkCreateValidator) Validate(ctx context.Context, input DbWorkCreateValidatorInput) error {
	ve := model.NewValidationError()

	// タイトル: 必須
	if strings.TrimSpace(input.Title) == "" {
		ve.AddField("title", i18n.T(ctx, "db_works_error_title_required"))
	}

	// メディア: 必須 + 許可された値
	if strings.TrimSpace(input.Media) == "" {
		ve.AddField("media", i18n.T(ctx, "db_works_error_media_required"))
	} else if !allowedMediaValues[input.Media] {
		ve.AddField("media", i18n.T(ctx, "db_works_error_media_invalid"))
	}

	// URL形式チェック（空の場合はスキップ）
	validateOptionalURL(ctx, ve, "official_site_url", input.OfficialSiteURL)
	validateOptionalURL(ctx, ve, "official_site_url_en", input.OfficialSiteURLEn)
	validateOptionalURL(ctx, ve, "wikipedia_url", input.WikipediaURL)
	validateOptionalURL(ctx, ve, "wikipedia_url_en", input.WikipediaURLEn)

	// sc_tid: 整数（空の場合はスキップ）
	if input.ScTid != "" {
		if _, err := strconv.Atoi(input.ScTid); err != nil {
			ve.AddField("sc_tid", i18n.T(ctx, "db_works_error_sc_tid_invalid"))
		}
	}

	// mal_anime_id: 整数（空の場合はスキップ）
	if input.MalAnimeID != "" {
		if _, err := strconv.Atoi(input.MalAnimeID); err != nil {
			ve.AddField("mal_anime_id", i18n.T(ctx, "db_works_error_mal_anime_id_invalid"))
		}
	}

	// あらすじと出典のペアチェック
	validatePresencePair(ctx, ve, "synopsis_source", input.Synopsis, input.SynopsisSource, "db_works_error_synopsis_source_required")
	validatePresencePair(ctx, ve, "synopsis_source_en", input.SynopsisEn, input.SynopsisSourceEn, "db_works_error_synopsis_source_en_required")

	if ve.HasErrors() {
		return ve
	}
	return nil
}

// validateOptionalURL はURLが空でない場合にURL形式をチェックする
func validateOptionalURL(ctx context.Context, ve *model.ValidationError, field, value string) {
	if value == "" {
		return
	}
	u, err := url.ParseRequestURI(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		ve.AddField(field, i18n.T(ctx, "db_works_error_url_invalid"))
	}
}

// validatePresencePair はペアの片方がある場合に、もう片方も必須にする
func validatePresencePair(ctx context.Context, ve *model.ValidationError, sourceField, content, source, errKey string) {
	if strings.TrimSpace(content) != "" && strings.TrimSpace(source) == "" {
		ve.AddField(sourceField, i18n.T(ctx, errKey))
	}
}
