package viewmodel

import (
	"strings"

	"github.com/annict/annict/go/internal/config"
)

// Season は季節情報を表します
type Season struct {
	Value string // "2026-winter" 形式
}

// Path は季節のパスを返します（例: "/works/2026-winter"）
func (s Season) Path() string {
	if s.Value == "" {
		return "/works"
	}
	return "/works/" + s.Value
}

// Icon は季節に応じたアイコン名を返します
func (s Season) Icon() string {
	if s.Value == "" {
		return "info"
	}

	// 季節を抽出（例: "2026-winter" → "winter"）
	parts := strings.Split(s.Value, "-")
	if len(parts) < 2 {
		return "info"
	}
	seasonName := parts[len(parts)-1]

	switch seasonName {
	case "winter":
		return "snowflake-regular"
	case "spring":
		return "flower-regular"
	case "summer":
		return "island-regular"
	case "autumn":
		return "acorn-regular"
	default:
		return "info"
	}
}

// Seasons は前・現在・次の季節情報をまとめた構造体です
type Seasons struct {
	Previous Season
	Current  Season
	Next     Season
}

// NewSeasons は config から Seasons を生成します
func NewSeasons(cfg *config.Config) Seasons {
	return Seasons{
		Previous: Season{Value: cfg.SeasonPrevious},
		Current:  Season{Value: cfg.SeasonCurrent},
		Next:     Season{Value: cfg.SeasonNext},
	}
}
