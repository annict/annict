package viewmodel

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/i18n"
)

func TestSeasonsSingleSource(t *testing.T) {
	t.Parallel()

	// seasons is the single source shared by the work form and the release-season
	// filter; pin its enum <-> slug <-> i18n mapping and ascending order (mirroring
	// Rails' Season::NAME_HASH: winter=1, spring=2, summer=3, autumn=4).
	//
	// [Ja] seasons は作品フォームとリリース時期フィルタが共有する単一ソース。enum <->
	// スラッグ <-> i18n の対応と昇順を固定する (Rails の Season::NAME_HASH: winter=1,
	// spring=2, summer=3, autumn=4 をミラー)。
	want := []seasonDef{
		{value: 1, slug: "winter", key: "season_winter"},
		{value: 2, slug: "spring", key: "season_spring"},
		{value: 3, slug: "summer", key: "season_summer"},
		{value: 4, slug: "autumn", key: "season_autumn"},
	}
	if !reflect.DeepEqual(seasons, want) {
		t.Errorf("seasons = %+v, want %+v", seasons, want)
	}
}

func TestSeasonLabelKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value int32
		want  string
	}{
		{value: 1, want: "season_winter"},
		{value: 2, want: "season_spring"},
		{value: 3, want: "season_summer"},
		{value: 4, want: "season_autumn"},
		{value: 0, want: ""},
		{value: 5, want: ""},
	}
	for _, tt := range tests {
		if got := seasonLabelKey(tt.value); got != tt.want {
			t.Errorf("seasonLabelKey(%d) = %q, want %q", tt.value, got, tt.want)
		}
	}
}

func TestSeasonMaxYear(t *testing.T) {
	t.Parallel()

	before := time.Now().Year()
	got := seasonMaxYear()
	after := time.Now().Year()

	// seasonMaxYear reads time.Now() internally; accept either surrounding year read so
	// the test never flakes at the New Year boundary.
	//
	// [Ja] seasonMaxYear は内部で time.Now() を読むため、年またぎでの flake を避けて
	// 前後どちらの年の読み取りも許容する。
	if got != before+seasonMaxYearOffset && got != after+seasonMaxYearOffset {
		t.Errorf("seasonMaxYear() = %d, want %d", got, before+seasonMaxYearOffset)
	}
}

func TestParseSeasonSlugs(t *testing.T) {
	t.Parallel()

	t.Run("有効なスラッグを並列の (年, 季節) ペアに変換する", func(t *testing.T) {
		t.Parallel()

		years, names := ParseSeasonSlugs([]string{"2024-spring", "2023-winter", "2022-autumn"})
		if want := []int32{2024, 2023, 2022}; !reflect.DeepEqual(years, want) {
			t.Errorf("years = %v, want %v", years, want)
		}
		// The name enum values are spring=2, winter=1, autumn=4.
		//
		// [Ja] 季節の enum 値は spring=2, winter=1, autumn=4。
		if want := []int32{2, 1, 4}; !reflect.DeepEqual(names, want) {
			t.Errorf("names = %v, want %v", names, want)
		}
	})

	t.Run("不正・範囲外のスラッグはスキップする", func(t *testing.T) {
		t.Parallel()

		years, names := ParseSeasonSlugs([]string{
			"2024-spring",  // valid
			"invalid",      // no "-"
			"2024-unknown", // unknown season name
			"9999-spring",  // year out of range (> now+5)
			"1889-winter",  // year below the lower bound (< 1890)
			"abc-spring",   // non-numeric year
		})
		if want := []int32{2024}; !reflect.DeepEqual(years, want) {
			t.Errorf("years = %v, want %v", years, want)
		}
		if want := []int32{2}; !reflect.DeepEqual(names, want) {
			t.Errorf("names = %v, want %v", names, want)
		}
	})

	t.Run("空入力では空スライスを返す", func(t *testing.T) {
		t.Parallel()

		years, names := ParseSeasonSlugs(nil)
		if len(years) != 0 || len(names) != 0 {
			t.Errorf("years/names = %v/%v, want empty", years, names)
		}
	})
}

func TestNewSeasonFilterOptions(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	options := NewSeasonFilterOptions(ctx, []string{"2024-spring"})

	byslug := make(map[string]SeasonFilterOption, len(options))
	for _, opt := range options {
		byslug[opt.Slug] = opt
	}

	t.Run("選択済みスラッグに Selected と ja ラベルが付く", func(t *testing.T) {
		t.Parallel()

		opt, ok := byslug["2024-spring"]
		if !ok {
			t.Fatal("2024-spring option should exist")
		}
		if !opt.Selected {
			t.Error("2024-spring should be marked selected")
		}
		if opt.Label != "2024年春" {
			t.Errorf("Label = %q, want %q", opt.Label, "2024年春")
		}
	})

	t.Run("未選択スラッグは Selected が false", func(t *testing.T) {
		t.Parallel()

		opt, ok := byslug["2024-summer"]
		if !ok {
			t.Fatal("2024-summer option should exist")
		}
		if opt.Selected {
			t.Error("2024-summer should not be selected")
		}
	})

	t.Run("年は降順・年内は季節の enum 降順で並ぶ", func(t *testing.T) {
		t.Parallel()

		idx := make(map[string]int, len(options))
		for i, opt := range options {
			idx[opt.Slug] = i
		}

		// Within a year the seasons run autumn→summer→spring→winter (descending enum
		// value); a newer year precedes an older one, so 2024-winter comes before
		// 2023-autumn. Each slug must appear strictly before the next.
		//
		// [Ja] 年内の季節は autumn→summer→spring→winter (enum 値の降順)、新しい年が古い年より
		// 前に来るため 2024-winter は 2023-autumn より前。各スラッグは次のものより厳密に前に並ぶ。
		ordered := []string{"2024-autumn", "2024-summer", "2024-spring", "2024-winter", "2023-autumn"}
		for i := 1; i < len(ordered); i++ {
			if idx[ordered[i-1]] >= idx[ordered[i]] {
				t.Errorf("%s は %s より前に並ぶべき", ordered[i-1], ordered[i])
			}
		}
	})
}
