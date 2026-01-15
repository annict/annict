package viewmodel

import (
	"testing"

	"github.com/annict/annict/go/internal/config"
)

func TestSeason_Path(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "空の値の場合は/worksを返す",
			value: "",
			want:  "/works",
		},
		{
			name:  "値がある場合は/works/{value}を返す",
			value: "2026-winter",
			want:  "/works/2026-winter",
		},
		{
			name:  "spring",
			value: "2026-spring",
			want:  "/works/2026-spring",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := Season{Value: tt.value}
			if got := s.Path(); got != tt.want {
				t.Errorf("Season.Path() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeason_Icon(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "空の値の場合はinfoを返す",
			value: "",
			want:  "info",
		},
		{
			name:  "ハイフンなしの不正なフォーマットの場合はinfoを返す",
			value: "invalid",
			want:  "info",
		},
		{
			name:  "winterの場合はsnowflake-regularを返す",
			value: "2026-winter",
			want:  "snowflake-regular",
		},
		{
			name:  "springの場合はflower-regularを返す",
			value: "2026-spring",
			want:  "flower-regular",
		},
		{
			name:  "summerの場合はisland-regularを返す",
			value: "2026-summer",
			want:  "island-regular",
		},
		{
			name:  "autumnの場合はacorn-regularを返す",
			value: "2026-autumn",
			want:  "acorn-regular",
		},
		{
			name:  "不明な季節の場合はinfoを返す",
			value: "2026-unknown",
			want:  "info",
		},
		{
			name:  "複数のハイフンがある場合は最後の部分を季節として扱う",
			value: "2026-2027-winter",
			want:  "snowflake-regular",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := Season{Value: tt.value}
			if got := s.Icon(); got != tt.want {
				t.Errorf("Season.Icon() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSeasons(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		SeasonPrevious: "2025-autumn",
		SeasonCurrent:  "2026-winter",
		SeasonNext:     "2026-spring",
	}

	seasons := NewSeasons(cfg)

	if seasons.Previous.Value != "2025-autumn" {
		t.Errorf("Previous.Value = %v, want %v", seasons.Previous.Value, "2025-autumn")
	}
	if seasons.Current.Value != "2026-winter" {
		t.Errorf("Current.Value = %v, want %v", seasons.Current.Value, "2026-winter")
	}
	if seasons.Next.Value != "2026-spring" {
		t.Errorf("Next.Value = %v, want %v", seasons.Next.Value, "2026-spring")
	}
}

func TestNewSeasons_EmptyConfig(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}

	seasons := NewSeasons(cfg)

	if seasons.Previous.Value != "" {
		t.Errorf("Previous.Value = %v, want empty string", seasons.Previous.Value)
	}
	if seasons.Current.Value != "" {
		t.Errorf("Current.Value = %v, want empty string", seasons.Current.Value)
	}
	if seasons.Next.Value != "" {
		t.Errorf("Next.Value = %v, want empty string", seasons.Next.Value)
	}

	// 空の場合はデフォルトのパスとアイコンを返す
	if seasons.Previous.Path() != "/works" {
		t.Errorf("Previous.Path() = %v, want /works", seasons.Previous.Path())
	}
	if seasons.Previous.Icon() != "info" {
		t.Errorf("Previous.Icon() = %v, want info", seasons.Previous.Icon())
	}
}
