package viewmodel

import (
	"context"
	"fmt"
	"time"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

// WorkStatus wraps model.WorkStatus for use in the Presentation layer, since templates may not depend on the model package.
//
// [Ja] WorkStatus は Presentation 層から扱える形に model.WorkStatus をラップした型 (templates は model に直接依存できないため)。
type WorkStatus model.WorkStatus

const (
	WorkStatusPublished WorkStatus = WorkStatus(model.WorkStatusPublished)
	WorkStatusArchived  WorkStatus = WorkStatus(model.WorkStatusArchived)
	WorkStatusDeleted   WorkStatus = WorkStatus(model.WorkStatusDeleted)
)

// String returns the textual representation of the status.
//
// [Ja] ステータスの文字列表現を返す。
func (s WorkStatus) String() string { return string(s) }

// DBWorkListItem is the per-row display data for the work list on the Annict DB admin screen.
//
// [Ja] DBWorkListItem は Annict DB 管理画面の作品一覧で 1 行ごとに表示する整形済みデータ。
type DBWorkListItem struct {
	ID    WorkID
	Title string
	// Pre-formatted season display string.
	//
	// [Ja] フォーマット済みのシーズン表示文字列。
	Season        string
	WatchersCount int32
	Status        WorkStatus
	HasImage      bool
}

func NewDBWorkListItems(ctx context.Context, works []*model.Work) []DBWorkListItem {
	result := make([]DBWorkListItem, len(works))
	for i, work := range works {
		result[i] = NewDBWorkListItem(ctx, work)
	}
	return result
}

func NewDBWorkListItem(ctx context.Context, work *model.Work) DBWorkListItem {
	return DBWorkListItem{
		ID:            WorkID(work.ID),
		Title:         work.Title,
		Season:        formatSeason(ctx, work.SeasonYear, work.SeasonName),
		WatchersCount: work.WatchersCount,
		Status:        WorkStatus(work.Status),
		HasImage:      work.ImageData != "",
	}
}

func formatSeason(ctx context.Context, year *int32, name *int32) string {
	if year == nil || name == nil {
		return ""
	}

	seasonKey := ""
	switch *name {
	case 1:
		seasonKey = "season_winter"
	case 2:
		seasonKey = "season_spring"
	case 3:
		seasonKey = "season_summer"
	case 4:
		seasonKey = "season_autumn"
	}

	if seasonKey == "" {
		return fmt.Sprintf("%d", *year)
	}

	return fmt.Sprintf("%d %s", *year, i18n.T(ctx, seasonKey))
}

type SelectOption struct {
	Value string
	Label string
}

type DBWorkFormOptions struct {
	MediaOptions        []SelectOption
	SeasonYearOptions   []SelectOption
	SeasonNameOptions   []SelectOption
	NumberFormatOptions []SelectOption
}

func NewDBWorkFormOptions(ctx context.Context, numberFormats []model.NumberFormat) DBWorkFormOptions {
	return DBWorkFormOptions{
		MediaOptions:        buildMediaOptions(ctx),
		SeasonYearOptions:   buildSeasonYearOptions(),
		SeasonNameOptions:   buildSeasonNameOptions(ctx),
		NumberFormatOptions: buildNumberFormatOptions(numberFormats),
	}
}

func buildMediaOptions(ctx context.Context) []SelectOption {
	return []SelectOption{
		{Value: "1", Label: i18n.T(ctx, "media_tv")},
		{Value: "2", Label: i18n.T(ctx, "media_ova")},
		{Value: "3", Label: i18n.T(ctx, "media_movie")},
		{Value: "4", Label: i18n.T(ctx, "media_web")},
		{Value: "0", Label: i18n.T(ctx, "media_other")},
	}
}

func buildSeasonYearOptions() []SelectOption {
	currentYear := time.Now().Year() + 5
	startYear := 1890
	options := make([]SelectOption, 0, currentYear-startYear+1)
	for y := currentYear; y >= startYear; y-- {
		options = append(options, SelectOption{
			Value: fmt.Sprintf("%d", y),
			Label: fmt.Sprintf("%d", y),
		})
	}
	return options
}

func buildSeasonNameOptions(ctx context.Context) []SelectOption {
	return []SelectOption{
		{Value: "1", Label: i18n.T(ctx, "season_winter")},
		{Value: "2", Label: i18n.T(ctx, "season_spring")},
		{Value: "3", Label: i18n.T(ctx, "season_summer")},
		{Value: "4", Label: i18n.T(ctx, "season_autumn")},
	}
}

func buildNumberFormatOptions(formats []model.NumberFormat) []SelectOption {
	options := make([]SelectOption, len(formats))
	for i, f := range formats {
		options[i] = SelectOption{
			Value: fmt.Sprintf("%d", f.ID),
			Label: f.Name,
		}
	}
	return options
}

// DBWorkFormInput holds the submitted form values so the work form can be re-rendered with the user's input after a validation error.
//
// [Ja] DBWorkFormInput はバリデーションエラー時に作品フォームを再描画するために、送信された入力値を保持する。
type DBWorkFormInput struct {
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

func NewDBWorkFormInput(input usecase.CreateWorkInput) *DBWorkFormInput {
	return &DBWorkFormInput{
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
	}
}

// Val returns the form value for the given field, or "" when the receiver is nil.
//
// [Ja] Val は指定フィールドのフォーム値を返す。レシーバが nil のときは "" を返す。
func (d *DBWorkFormInput) Val(field string) string {
	if d == nil {
		return ""
	}
	switch field {
	case "title":
		return d.Title
	case "title_kana":
		return d.TitleKana
	case "title_alter":
		return d.TitleAlter
	case "title_en":
		return d.TitleEn
	case "title_alter_en":
		return d.TitleAlterEn
	case "media":
		return d.Media
	case "season_year":
		return d.SeasonYear
	case "season_name":
		return d.SeasonName
	case "started_on":
		return d.StartedOn
	case "ended_on":
		return d.EndedOn
	case "official_site_url":
		return d.OfficialSiteURL
	case "official_site_url_en":
		return d.OfficialSiteURLEn
	case "wikipedia_url":
		return d.WikipediaURL
	case "wikipedia_url_en":
		return d.WikipediaURLEn
	case "twitter_username":
		return d.TwitterUsername
	case "twitter_hashtag":
		return d.TwitterHashtag
	case "sc_tid":
		return d.ScTid
	case "mal_anime_id":
		return d.MalAnimeID
	case "synopsis":
		return d.Synopsis
	case "synopsis_source":
		return d.SynopsisSource
	case "synopsis_en":
		return d.SynopsisEn
	case "synopsis_source_en":
		return d.SynopsisSourceEn
	case "manual_episodes_count":
		return d.ManualEpisodesCount
	case "start_episode_raw_number":
		return d.StartEpisodeRawNumber
	case "number_format_id":
		return d.NumberFormatID
	case "no_episodes":
		return d.NoEpisodes
	default:
		return ""
	}
}
