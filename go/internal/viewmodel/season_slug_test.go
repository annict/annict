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

	// seasonsは作品フォームとリリース時期フィルタが共有する単一ソース。enum <->
	// スラッグ <-> i18nの対応と昇順を固定する (RailsのSeason::NAME_HASH: winter=1,
	// spring=2, summer=3, autumn=4をミラー)。
	want := []seasonDef{
		{value: 1, slug: "winter", key: "season_winter"},
		{value: 2, slug: "spring", key: "season_spring"},
		{value: 3, slug: "summer", key: "season_summer"},
		{value: 4, slug: "autumn", key: "season_autumn"},
	}
	if !reflect.DeepEqual(seasons, want) {
		t.Errorf("seasons = %+v、期待値 = %+v", seasons, want)
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
			t.Errorf("seasonLabelKey(%d) = %q、期待値 = %q", tt.value, got, tt.want)
		}
	}
}

func TestSeasonMaxYear(t *testing.T) {
	t.Parallel()

	before := time.Now().Year()
	got := seasonMaxYear()
	after := time.Now().Year()

	// seasonMaxYearは内部でtime.Now() を読むため、年またぎでのflakeを避けて
	// 前後どちらの年の読み取りも許容する。
	if got != before+seasonMaxYearOffset && got != after+seasonMaxYearOffset {
		t.Errorf("seasonMaxYear() = %d、期待値 = %d", got, before+seasonMaxYearOffset)
	}
}

func TestParseSeasonSlugs(t *testing.T) {
	t.Parallel()

	t.Run("有効なスラッグを並列の (年, 季節) ペアに変換する", func(t *testing.T) {
		t.Parallel()

		years, names := ParseSeasonSlugs([]string{"2024-spring", "2023-winter", "2022-autumn"})
		if want := []int32{2024, 2023, 2022}; !reflect.DeepEqual(years, want) {
			t.Errorf("years = %v、期待値 = %v", years, want)
		}
		// 季節のenum値はspring=2, winter=1, autumn=4。
		if want := []int32{2, 1, 4}; !reflect.DeepEqual(names, want) {
			t.Errorf("names = %v、期待値 = %v", names, want)
		}
	})

	t.Run("不正・範囲外のスラッグはスキップする", func(t *testing.T) {
		t.Parallel()

		years, names := ParseSeasonSlugs([]string{
			"2024-spring",  // 有効
			"invalid",      // "-" が無い
			"2024-unknown", // 未知の季節名
			"9999-spring",  // 年が範囲外 (> now+5)
			"1889-winter",  // 年が下限未満 (< 1890)
			"abc-spring",   // 年が数値でない
		})
		if want := []int32{2024}; !reflect.DeepEqual(years, want) {
			t.Errorf("years = %v、期待値 = %v", years, want)
		}
		if want := []int32{2}; !reflect.DeepEqual(names, want) {
			t.Errorf("names = %v、期待値 = %v", names, want)
		}
	})

	t.Run("空入力では空スライスを返す", func(t *testing.T) {
		t.Parallel()

		years, names := ParseSeasonSlugs(nil)
		if len(years) != 0 || len(names) != 0 {
			t.Errorf("years/names = %v/%v、期待値 = 空", years, names)
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

	t.Run("選択済みスラッグにSelectedとjaラベルが付く", func(t *testing.T) {
		t.Parallel()

		opt, ok := byslug["2024-spring"]
		if !ok {
			t.Fatal("2024-springの選択肢が無い")
		}
		if !opt.Selected {
			t.Error("2024-springが選択済みになっていない")
		}
		if opt.Label != "2024年春" {
			t.Errorf("Label = %q、期待値 = %q", opt.Label, "2024年春")
		}
	})

	t.Run("未選択スラッグはSelectedがfalse", func(t *testing.T) {
		t.Parallel()

		opt, ok := byslug["2024-summer"]
		if !ok {
			t.Fatal("2024-summerの選択肢が無い")
		}
		if opt.Selected {
			t.Error("2024-summerが選択済みになっている")
		}
	})

	t.Run("年は降順・年内は季節のenum降順で並ぶ", func(t *testing.T) {
		t.Parallel()

		idx := make(map[string]int, len(options))
		for i, opt := range options {
			idx[opt.Slug] = i
		}

		// 年内の季節はautumn→summer→spring→winter (enum値の降順)、新しい年が古い年より
		// 前に来るため2024-winterは2023-autumnより前。各スラッグは次のものより厳密に前に並ぶ。
		ordered := []string{"2024-autumn", "2024-summer", "2024-spring", "2024-winter", "2023-autumn"}
		for i := 1; i < len(ordered); i++ {
			if idx[ordered[i-1]] >= idx[ordered[i]] {
				t.Errorf("%sは%sより前に並ぶべき", ordered[i-1], ordered[i])
			}
		}
	})
}
