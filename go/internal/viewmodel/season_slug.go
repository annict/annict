package viewmodel

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/annict/annict/go/internal/i18n"
)

// seasonStartYearはリリース時期のUI (フォームの年selectと一覧フィルタ) が提供する
// 最古の年で、RailsのSeason::YEAR_LISTの下限に合わせている。
const seasonStartYear = 1890

// seasonMaxYearOffsetはリリース時期のUIが現在の年から何年先まで提供するかで、
// RailsのSeason::YEAR_LISTの上限 (現在の年 + 5) に合わせている。
const seasonMaxYearOffset = 5

// seasonDefはworks.season_nameのenum値を、リリース時期のスラッグトークンと
// i18nラベルキーに対応づける。
type seasonDef struct {
	value int32
	slug  string
	key   string
}

// seasonsは季節のenum <-> スラッグ <-> i18nの対応の単一ソースで、作品フォーム
// (db_work.go) と下のリリース時期フィルタが共有する。RailsのSeason::NAME_HASH
// (winter=1, spring=2, summer=3, autumn=4) をミラーし、enum値の昇順で保持する。降順の
// 表示が必要な参照側は逆順で走査する。
var seasons = []seasonDef{
	{value: 1, slug: "winter", key: "season_winter"},
	{value: 2, slug: "spring", key: "season_spring"},
	{value: 3, slug: "summer", key: "season_summer"},
	{value: 4, slug: "autumn", key: "season_autumn"},
}

// seasonMaxYearはリリース時期のUIが提供する最新の年 (現在の年 + 5) を返す。
func seasonMaxYear() int {
	return time.Now().Year() + seasonMaxYearOffset
}

// seasonLabelKeyはworks.season_nameのenum値に対応するi18nラベルキーを返す。
// 未知の値のときは "" を返す。
func seasonLabelKey(value int32) string {
	for _, s := range seasons {
		if s.value == value {
			return s.key
		}
	}
	return ""
}

// SeasonFilterOptionはリリース時期の複数選択フィルタの1オプション。
type SeasonFilterOption struct {
	Slug     string
	Label    string
	Selected bool
}

// ParseSeasonSlugsはリリース時期のスラッグ ("2024-spring") を、DB一覧フィルタが
// 照合する並列の (年, 季節) enumペアに変換する。不正・範囲外のスラッグはセーフティネット
// としてスキップする (スラッグはサーバー生成の <option> 由来のため実際には発生しない)。
// 戻り値の2スライスは常に同じ長さ。
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
			// yは上で [seasonStartYear, maxYear] に制限済みのためint32に収まる。
			return int32(y), s.value, true // #nosec G109 G115
		}
	}
	return 0, 0, false
}

// NewSeasonFilterOptionsはリリース時期の複数選択オプションを降順 (新しい年・季節が先)
// で構築し、RailsのSeason.list(sort: :desc) に合わせる。selectedSlugsは事前選択済みの
// オプションを印付け、フォームが利用者の現在の選択を再描画できるようにする。
func NewSeasonFilterOptions(ctx context.Context, selectedSlugs []string) []SeasonFilterOption {
	selected := make(map[string]bool, len(selectedSlugs))
	for _, slug := range selectedSlugs {
		selected[slug] = true
	}

	maxYear := seasonMaxYear()
	options := make([]SeasonFilterOption, 0, (maxYear-seasonStartYear+1)*len(seasons))
	for year := maxYear; year >= seasonStartYear; year-- {
		// seasonsを逆順に走査し、年内で新しい季節を先頭にする
		// (autumn -> summer -> spring -> winter)。Railsの降順に合わせる。
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
