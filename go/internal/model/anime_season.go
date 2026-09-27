package model

import "time"

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
