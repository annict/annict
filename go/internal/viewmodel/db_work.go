package viewmodel

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// DBWorkListItem はDB管理画面の作品一覧用の表示データです
type DBWorkListItem struct {
	ID            int64
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
		ID:            item.ID,
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
