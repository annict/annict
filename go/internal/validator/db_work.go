package validator

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// allowedMediaValues lists the media type codes accepted by the create-work form.
// The mapping mirrors the Rails enum on the works.media column
// (0=other, 1=tv, 2=ova, 3=movie, 4=web).
//
// [Ja] allowedMediaValues は作品作成フォームで許可されるメディア種別コードの一覧。
// Rails 版の works.media enum と対応している (0=その他, 1=テレビ, 2=OVA, 3=映画, 4=Web)。
var allowedMediaValues = map[string]bool{
	"0": true,
	"1": true,
	"2": true,
	"3": true,
	"4": true,
}

// DbWorkCreateValidator validates the create-work form on the Annict DB admin screen.
//
// [Ja] DbWorkCreateValidator は Annict DB 管理画面の作品作成フォームを検証する。
type DbWorkCreateValidator struct{}

func NewDbWorkCreateValidator() *DbWorkCreateValidator {
	return &DbWorkCreateValidator{}
}

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

func (v *DbWorkCreateValidator) Validate(ctx context.Context, input DbWorkCreateValidatorInput) error {
	ve := model.NewValidationError()

	if strings.TrimSpace(input.Title) == "" {
		ve.AddField("title", i18n.T(ctx, "validation_required"))
	}

	if strings.TrimSpace(input.Media) == "" {
		ve.AddField("media", i18n.T(ctx, "validation_required"))
	} else if !allowedMediaValues[input.Media] {
		ve.AddField("media", i18n.T(ctx, "validation_media_invalid"))
	}

	validateOptionalURL(ctx, ve, "official_site_url", input.OfficialSiteURL)
	validateOptionalURL(ctx, ve, "official_site_url_en", input.OfficialSiteURLEn)
	validateOptionalURL(ctx, ve, "wikipedia_url", input.WikipediaURL)
	validateOptionalURL(ctx, ve, "wikipedia_url_en", input.WikipediaURLEn)

	if input.ScTid != "" {
		if _, err := strconv.Atoi(input.ScTid); err != nil {
			ve.AddField("sc_tid", i18n.T(ctx, "validation_integer_invalid"))
		}
	}

	if input.MalAnimeID != "" {
		if _, err := strconv.Atoi(input.MalAnimeID); err != nil {
			ve.AddField("mal_anime_id", i18n.T(ctx, "validation_integer_invalid"))
		}
	}

	validatePresencePair(ctx, ve, "synopsis_source", input.Synopsis, input.SynopsisSource, "validation_synopsis_source_required")
	validatePresencePair(ctx, ve, "synopsis_source_en", input.SynopsisEn, input.SynopsisSourceEn, "validation_synopsis_source_en_required")

	if ve.HasErrors() {
		return ve
	}
	return nil
}

func validateOptionalURL(ctx context.Context, ve *model.ValidationError, field, value string) {
	if value == "" {
		return
	}
	u, err := url.ParseRequestURI(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		ve.AddField(field, i18n.T(ctx, "validation_url_invalid"))
	}
}

// validatePresencePair requires the source field whenever the content field is filled in.
// Used for paired inputs like a synopsis and its citation, where filling one half
// without the other would leave a half-completed record.
//
// [Ja] validatePresencePair は対になる 2 フィールドのうち、content が入力されているときに
// source も必須にする。あらすじと出典のように対で意味を持つ入力で使い、片方だけ埋まった
// 中途半端なレコードを防ぐ。
func validatePresencePair(ctx context.Context, ve *model.ValidationError, sourceField, content, source, errKey string) {
	if strings.TrimSpace(content) != "" && strings.TrimSpace(source) == "" {
		ve.AddField(sourceField, i18n.T(ctx, errKey))
	}
}
