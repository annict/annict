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

// seasonNoneSlugはリリース時期フィルタで「季節未登録」を表すスラッグトークン
// ("2026-none")。共有URLに載るため改名しない。季節の部分にハイフンを含めると
// RailsのSeason.find_by_slugの分割 (split("-")) と食い違うため、1語にしている。
const seasonNoneSlug = "none"

// seasonNoneValueは「季節未登録」を (年, 季節) の組の季節として表す番兵値。
// works.season_nameのenumは1〜4で0は使われないため、一覧のクエリは0を
// season_nameがNULLの作品に一致させる。seasonsには含めず、作品フォームの季節の
// selectやseasonLabelKeyには現れない。
const seasonNoneValue int32 = 0

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
// 照合する並列の (年, 季節) enumペアに変換する。季節未登録のスラッグ ("2024-none")
// は季節をseasonNoneValueにする。共有URLや手入力で届く不正・範囲外のスラッグは
// スキップする。戻り値の2スライスは常に同じ長さ。
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
	// yは上で [seasonStartYear, maxYear] に制限済みのためint32に収まる。
	if nameStr == seasonNoneSlug {
		return int32(y), seasonNoneValue, true // #nosec G109 G115
	}
	for _, s := range seasons {
		if s.slug == nameStr {
			return int32(y), s.value, true // #nosec G109 G115
		}
	}
	return 0, 0, false
}

// NewSeasonFilterOptionsはリリース時期の複数選択オプションを降順 (新しい年・季節が先)
// で構築し、RailsのSeason.list(sort: :desc) に合わせる。各年の季節の後ろには季節
// 未登録の選択肢を置く (Railsのinclude_all: trueが年単位の選択肢を年内の末尾に置くのに
// 合わせる)。selectedSlugsは事前選択済みの
// オプションを印付け、フォームが利用者の現在の選択を再描画できるようにする。
func NewSeasonFilterOptions(ctx context.Context, selectedSlugs []string) []SeasonFilterOption {
	selected := make(map[string]bool, len(selectedSlugs))
	for _, slug := range selectedSlugs {
		selected[slug] = true
	}

	maxYear := seasonMaxYear()
	options := make([]SeasonFilterOption, 0, (maxYear-seasonStartYear+1)*(len(seasons)+1))
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
		noneSlug := fmt.Sprintf("%d-%s", year, seasonNoneSlug)
		options = append(options, SeasonFilterOption{
			Slug:     noneSlug,
			Label:    i18n.T(ctx, "year_no_season", map[string]any{"Year": year}),
			Selected: selected[noneSlug],
		})
	}
	return options
}
