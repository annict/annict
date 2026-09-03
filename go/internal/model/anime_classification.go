package model

import (
	"database/sql"
	"time"
)

// AnimeClassificationKind discriminates a classification row as a work or an
// episode and mirrors the anime_classification_kind PostgreSQL enum.
//
// [Ja] AnimeClassificationKind は分類行を work か episode に区別し、PostgreSQL の
// anime_classification_kind enum と対応する。
type AnimeClassificationKind string

const (
	AnimeClassificationKindWork    AnimeClassificationKind = "work"
	AnimeClassificationKindEpisode AnimeClassificationKind = "episode"
)

// String returns the textual representation of the kind.
//
// [Ja] 種別の文字列表現を返す。
func (k AnimeClassificationKind) String() string { return string(k) }

// AnimeClassification is the domain entity for layer 2 (catalog
// classification). A UNIQUE (anime_id) constraint ties it 1:1 to an anime, so
// the same content can be reclassified between work and episode without
// changing its identity.
//
// Number / EpisodeStartNumber back NUMERIC columns and are carried as strings
// to preserve the exact decimal (e.g. "3.5") without float rounding.
//
// [Ja] AnimeClassification は第 2 層 (カタログ分類) のドメインエンティティ。
// UNIQUE (anime_id) によりアニメと 1:1 で結びつくため、同一のコンテンツを
// 同一性を変えずに work と episode の間で再分類できる。
//
// Number / EpisodeStartNumber は NUMERIC カラムに対応し、浮動小数の丸めなしに
// 正確な小数 (例: "3.5") を保つため文字列で持つ。
type AnimeClassification struct {
	ID                    AnimeClassificationID
	AnimeID               AnimeID
	Kind                  AnimeClassificationKind
	ParentAnimeID         *AnimeID
	Number                sql.NullString
	NumberText            sql.NullString
	SortNumber            sql.NullInt32
	Standalone            bool
	NumberFormatID        *NumberFormatID
	EpisodeStartNumber    sql.NullString
	ExpectedEpisodesCount sql.NullInt32
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
