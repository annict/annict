package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// SeasonNameはAnimeSeasonの季節を表し、PostgreSQLのseason_name enumと対応する。
// 旧works.season_nameのinteger (1=winter, 2=spring, 3=summer, 4=autumn) がこれらの値に
// 写像され、autumnはfallに寄せる。anime_seasonsがmodel層での最初の利用者のためここで
// 定義する。後続のテーブルが必要とした時点で共有の場所へ引き上げてよい。
type SeasonName string

const (
	SeasonNameWinter SeasonName = "winter"
	SeasonNameSpring SeasonName = "spring"
	SeasonNameSummer SeasonName = "summer"
	SeasonNameFall   SeasonName = "fall"
)

// Stringは季節名の文字列表現を返す。
func (n SeasonName) String() string { return string(n) }

// AnimeSeasonはanime_seasonsのドメインエンティティ。animeが掲載される季節を
// (anime_id, year, name) をキーに持つ。第1層のanime同一性をキーにするため、再分類を
// またいでも季節が紐づき続ける。Nameは季節名が未定 (年のみ判明) のときnilで、
// (anime_id, year, name) のUNIQUEインデックスはNULLのnameを「区別しない」(NULLS NOT
// DISTINCT) ため、animeは年ごとにそうした行を高々1つ持つ。IsPrimaryはworksがsource
// する主季節を示し、worksは自身がsourceする単一の季節についてtrueに設定する。編集者が
// 直接足す副次シーズン (後続フェーズ) はis_primary=falseで、同期が自身の行と編集者追加の
// 行を見分ける手がかりになる (部分UNIQUEインデックスもanimeごとにis_primary行を高々
// 1つに保つ)。
type AnimeSeason struct {
	ID        AnimeSeasonID
	AnimeID   AnimeID
	Year      int32
	Name      *SeasonName
	IsPrimary bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SortKeyは、季節の順に並べるためのキーを返す。年 * 10に、季節 (冬1・春2・夏3・秋4・不明5) を足す。
// REST API V2の作品の一覧のクエリ (ListAPIWorkIDsOrderBySeason) とインデックスも同じ式で並べるため、
// 式を変えるときはクエリとインデックスもあわせて変える
func (s *AnimeSeason) SortKey() int32 {
	return s.Year*10 + seasonNameSortOrder(s.Name)
}

func seasonNameSortOrder(name *SeasonName) int32 {
	if name == nil {
		return 5
	}
	switch *name {
	case SeasonNameWinter:
		return 1
	case SeasonNameSpring:
		return 2
	case SeasonNameSummer:
		return 3
	case SeasonNameFall:
		return 4
	default:
		return 5
	}
}

// SeasonBucketは、作品を絞り込む季節 (年と季節) を表す。Nameがnilのものは、年だけが分かっていて季節の分からない作品を指す
type SeasonBucket struct {
	Year int32
	Name *SeasonName
}

// seasonBucketUnknownは、季節の分からない作品を指すSeasonBucketの文字列表現の季節の部分
const seasonBucketUnknown = "unknown"

// Stringは `2026-spring`・`2026-unknown` の形の文字列を返す
func (b SeasonBucket) String() string {
	name := seasonBucketUnknown
	if b.Name != nil {
		name = string(*b.Name)
	}
	return strconv.FormatInt(int64(b.Year), 10) + "-" + name
}

// ParseSeasonBucketは `2026-spring`・`2026-unknown` の形の文字列を読む
func ParseSeasonBucket(s string) (SeasonBucket, error) {
	yearText, nameText, ok := strings.Cut(s, "-")
	if !ok {
		return SeasonBucket{}, fmt.Errorf("季節 %q は `{年}-{季節}` の形ではありません", s)
	}
	year, err := strconv.ParseInt(yearText, 10, 32)
	if err != nil {
		return SeasonBucket{}, fmt.Errorf("季節 %q の年を読めません: %w", s, err)
	}

	bucket := SeasonBucket{Year: int32(year)}
	switch name := SeasonName(nameText); name {
	case SeasonNameWinter, SeasonNameSpring, SeasonNameSummer, SeasonNameFall:
		bucket.Name = &name
	case seasonBucketUnknown:
	default:
		return SeasonBucket{}, fmt.Errorf("季節 %q の季節を読めません", s)
	}
	return bucket, nil
}
