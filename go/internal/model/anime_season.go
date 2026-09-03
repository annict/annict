package model

import "time"

// SeasonName identifies the season of an AnimeSeason and mirrors the season_name
// PostgreSQL enum. The legacy works.season_name integer (1=winter, 2=spring,
// 3=summer, 4=autumn) maps onto these values, with autumn folded to fall. It is
// defined here because anime_seasons is its first consumer in the model layer; a
// later table that needs it can lift this type to a shared location.
//
// [Ja] SeasonName は AnimeSeason の季節を表し、PostgreSQL の season_name enum と対応する。
// 旧 works.season_name の integer (1=winter, 2=spring, 3=summer, 4=autumn) がこれらの値に
// 写像され、autumn は fall に寄せる。anime_seasons が model 層での最初の利用者のためここで
// 定義する。後続のテーブルが必要とした時点で共有の場所へ引き上げてよい。
type SeasonName string

const (
	SeasonNameWinter SeasonName = "winter"
	SeasonNameSpring SeasonName = "spring"
	SeasonNameSummer SeasonName = "summer"
	SeasonNameFall   SeasonName = "fall"
)

// String returns the textual representation of the season name.
//
// [Ja] 季節名の文字列表現を返す。
func (n SeasonName) String() string { return string(n) }

// AnimeSeason is the domain entity for anime_seasons: it holds the seasons an anime is
// listed in, keyed by (anime_id, year, name). It is keyed by the layer-1 anime
// identity, so a season stays attached across re-classification. Name is nil when the
// season name is undetermined (only the year is known); the (anime_id, year, name)
// UNIQUE index treats NULL names as not-distinct so an anime holds at most one such
// row per year. IsPrimary marks the work-sourced primary season; works set it true for
// the single season they source, while a secondary season an editor adds directly (a
// later phase) is is_primary=false, which is how the sync tells its own row apart from
// editor-added ones (a partial UNIQUE index also keeps at most one is_primary row per
// anime).
//
// [Ja] AnimeSeason は anime_seasons のドメインエンティティ。anime が掲載される季節を
// (anime_id, year, name) をキーに持つ。第 1 層の anime 同一性をキーにするため、再分類を
// またいでも季節が紐づき続ける。Name は季節名が未定 (年のみ判明) のとき nil で、
// (anime_id, year, name) の UNIQUE インデックスは NULL の name を「区別しない」(NULLS NOT
// DISTINCT) ため、anime は年ごとにそうした行を高々 1 つ持つ。IsPrimary は works が source
// する主季節を示し、works は自身が source する単一の季節について true に設定する。編集者が
// 直接足す副次シーズン (後続フェーズ) は is_primary=false で、同期が自身の行と編集者追加の
// 行を見分ける手がかりになる (部分 UNIQUE インデックスも anime ごとに is_primary 行を高々
// 1 つに保つ)。
type AnimeSeason struct {
	ID        AnimeSeasonID
	AnimeID   AnimeID
	Year      int32
	Name      *SeasonName
	IsPrimary bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
