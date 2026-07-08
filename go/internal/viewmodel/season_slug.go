package viewmodel

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/annict/annict/go/internal/i18n"
)

// seasonStartYear is the oldest year the release-season UI (the form's year select and
// the list filter) offers, matching the lower bound of the Rails Season::YEAR_LIST.
//
// [Ja] seasonStartYear はリリース時期の UI (フォームの年 select と一覧フィルタ) が提供する
// 最古の年で、Rails の Season::YEAR_LIST の下限に合わせている。
const seasonStartYear = 1890

// seasonMaxYearOffset is how many years past the current year the release-season UI
// offers, matching the upper bound of the Rails Season::YEAR_LIST (current year + 5).
//
// [Ja] seasonMaxYearOffset はリリース時期の UI が現在の年から何年先まで提供するかで、
// Rails の Season::YEAR_LIST の上限 (現在の年 + 5) に合わせている。
const seasonMaxYearOffset = 5

// seasonDef pairs a works.season_name enum value with its release-season slug token and
// i18n label key.
//
// [Ja] seasonDef は works.season_name の enum 値を、リリース時期のスラッグトークンと
// i18n ラベルキーに対応づける。
type seasonDef struct {
	value int32
	slug  string
	key   string
}

// seasons is the single source for the season enum <-> slug <-> i18n mapping, shared by
// the work form (db_work.go) and the release-season filter below. It mirrors Rails'
// Season::NAME_HASH (winter=1, spring=2, summer=3, autumn=4) and is stored in ascending
// enum order; callers that need descending display order iterate it in reverse.
//
// [Ja] seasons は季節の enum <-> スラッグ <-> i18n の対応の単一ソースで、作品フォーム
// (db_work.go) と下のリリース時期フィルタが共有する。Rails の Season::NAME_HASH
// (winter=1, spring=2, summer=3, autumn=4) をミラーし、enum 値の昇順で保持する。降順の
// 表示が必要な参照側は逆順で走査する。
var seasons = []seasonDef{
	{value: 1, slug: "winter", key: "season_winter"},
	{value: 2, slug: "spring", key: "season_spring"},
	{value: 3, slug: "summer", key: "season_summer"},
	{value: 4, slug: "autumn", key: "season_autumn"},
}

// seasonMaxYear returns the newest year the release-season UI offers (current year + 5).
//
// [Ja] seasonMaxYear はリリース時期の UI が提供する最新の年 (現在の年 + 5) を返す。
func seasonMaxYear() int {
	return time.Now().Year() + seasonMaxYearOffset
}

// seasonLabelKey returns the i18n label key for a works.season_name enum value, or ""
// when the value is unknown.
//
// [Ja] seasonLabelKey は works.season_name の enum 値に対応する i18n ラベルキーを返す。
// 未知の値のときは "" を返す。
func seasonLabelKey(value int32) string {
	for _, s := range seasons {
		if s.value == value {
			return s.key
		}
	}
	return ""
}

// SeasonFilterOption is one option of the release-season multi-select filter.
//
// [Ja] SeasonFilterOption はリリース時期の複数選択フィルタの 1 オプション。
type SeasonFilterOption struct {
	Slug     string
	Label    string
	Selected bool
}

// ParseSeasonSlugs converts release-season slugs ("2024-spring") into the parallel
// (year, name) enum pairs the DB list filter matches on. Malformed or out-of-range
// slugs are skipped as a safety net; in practice they never occur because slugs come
// from the server-rendered <option> list. The two returned slices always have equal
// length.
//
// [Ja] ParseSeasonSlugs はリリース時期のスラッグ ("2024-spring") を、DB 一覧フィルタが
// 照合する並列の (年, 季節) enum ペアに変換する。不正・範囲外のスラッグはセーフティネット
// としてスキップする (スラッグはサーバー生成の <option> 由来のため実際には発生しない)。
// 戻り値の 2 スライスは常に同じ長さ。
func ParseSeasonSlugs(slugs []string) (years []int32, names []int32) {
	maxYear := seasonMaxYear()
	for _, slug := range slugs {
		year, name, ok := parseSeasonSlug(slug, maxYear)
		if !ok {
			continue
		}
		years = append(years, year)
		names = append(names, name)
	}
	return years, names
}

func parseSeasonSlug(slug string, maxYear int) (year int32, name int32, ok bool) {
	yearStr, nameStr, found := strings.Cut(slug, "-")
	if !found {
		return 0, 0, false
	}
	y, err := strconv.Atoi(yearStr)
	if err != nil || y < seasonStartYear || y > maxYear {
		return 0, 0, false
	}
	for _, s := range seasons {
		if s.slug == nameStr {
			// y is bounded to [seasonStartYear, maxYear] above, so it fits in int32.
			//
			// [Ja] y は上で [seasonStartYear, maxYear] に制限済みのため int32 に収まる。
			return int32(y), s.value, true // #nosec G109 G115
		}
	}
	return 0, 0, false
}

// NewSeasonFilterOptions builds the release-season multi-select options in descending
// order (newest year and season first), matching Rails' Season.list(sort: :desc).
// selectedSlugs marks which options are pre-selected so the form re-renders the user's
// current selection.
//
// [Ja] NewSeasonFilterOptions はリリース時期の複数選択オプションを降順 (新しい年・季節が先)
// で構築し、Rails の Season.list(sort: :desc) に合わせる。selectedSlugs は事前選択済みの
// オプションを印付け、フォームが利用者の現在の選択を再描画できるようにする。
func NewSeasonFilterOptions(ctx context.Context, selectedSlugs []string) []SeasonFilterOption {
	selected := make(map[string]bool, len(selectedSlugs))
	for _, slug := range selectedSlugs {
		selected[slug] = true
	}

	maxYear := seasonMaxYear()
	options := make([]SeasonFilterOption, 0, (maxYear-seasonStartYear+1)*len(seasons))
	for year := maxYear; year >= seasonStartYear; year-- {
		// Iterate seasons in reverse so the newest season within a year comes first
		// (autumn -> summer -> spring -> winter), matching Rails' descending order.
		//
		// [Ja] seasons を逆順に走査し、年内で新しい季節を先頭にする
		// (autumn -> summer -> spring -> winter)。Rails の降順に合わせる。
		for i := len(seasons) - 1; i >= 0; i-- {
			s := seasons[i]
			slug := fmt.Sprintf("%d-%s", year, s.slug)
			options = append(options, SeasonFilterOption{
				Slug: slug,
				Label: i18n.T(ctx, "year_season", map[string]any{
					"Year":   year,
					"Season": i18n.T(ctx, s.key),
				}),
				Selected: selected[slug],
			})
		}
	}
	return options
}
