package viewmodel

import (
	"context"
	"fmt"
	"time"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// DBWorkListItem はDB管理画面の作品一覧用の表示データです
type DBWorkListItem struct {
	ID            WorkID
	Title         string
	Season        string // フォーマット済みのシーズン表示文字列
	WatchersCount int32
	Status        string
	HasImage      bool
}

// NewDBWorkListItems は model.DBWorkListItem のスライスを viewmodel.DBWorkListItem のスライスに変換します
func NewDBWorkListItems(ctx context.Context, items []model.DBWorkListItem) []DBWorkListItem {
	result := make([]DBWorkListItem, len(items))
	for i, item := range items {
		result[i] = NewDBWorkListItem(ctx, item)
	}
	return result
}

// NewDBWorkListItem は model.DBWorkListItem を viewmodel.DBWorkListItem に変換します
func NewDBWorkListItem(ctx context.Context, item model.DBWorkListItem) DBWorkListItem {
	return DBWorkListItem{
		ID:            WorkID(item.ID),
		Title:         item.Title,
		Season:        formatSeason(ctx, item.SeasonYear, item.SeasonName),
		WatchersCount: item.WatchersCount,
		Status:        item.Status,
		HasImage:      item.HasImage,
	}
}

// formatSeason はシーズン情報をフォーマットします
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

// SelectOption はセレクトボックスの選択肢です
type SelectOption struct {
	Value string
	Label string
}

// DBWorkFormOptions は作品フォームのセレクトボックス用データです
type DBWorkFormOptions struct {
	MediaOptions        []SelectOption
	SeasonYearOptions   []SelectOption
	SeasonNameOptions   []SelectOption
	NumberFormatOptions []SelectOption
}

// NewDBWorkFormOptions は作品フォームのセレクトボックス用データを作成します
func NewDBWorkFormOptions(ctx context.Context, numberFormats []model.NumberFormat) DBWorkFormOptions {
	return DBWorkFormOptions{
		MediaOptions:        buildMediaOptions(ctx),
		SeasonYearOptions:   buildSeasonYearOptions(),
		SeasonNameOptions:   buildSeasonNameOptions(ctx),
		NumberFormatOptions: buildNumberFormatOptions(numberFormats),
	}
}

// buildMediaOptions はメディア種別の選択肢を作成します
func buildMediaOptions(ctx context.Context) []SelectOption {
	return []SelectOption{
		{Value: "1", Label: i18n.T(ctx, "media_tv")},
		{Value: "2", Label: i18n.T(ctx, "media_ova")},
		{Value: "3", Label: i18n.T(ctx, "media_movie")},
		{Value: "4", Label: i18n.T(ctx, "media_web")},
		{Value: "0", Label: i18n.T(ctx, "media_other")},
	}
}

// buildSeasonYearOptions はシーズン年の選択肢を作成します
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

// buildSeasonNameOptions はシーズン名の選択肢を作成します
func buildSeasonNameOptions(ctx context.Context) []SelectOption {
	return []SelectOption{
		{Value: "1", Label: i18n.T(ctx, "season_winter")},
		{Value: "2", Label: i18n.T(ctx, "season_spring")},
		{Value: "3", Label: i18n.T(ctx, "season_summer")},
		{Value: "4", Label: i18n.T(ctx, "season_autumn")},
	}
}

// buildNumberFormatOptions はNumberFormatの選択肢を作成します
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
