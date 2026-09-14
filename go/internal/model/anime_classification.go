package model

import (
	"database/sql"
	"time"
)

// AnimeClassificationKindは分類行をworkかepisodeに区別し、PostgreSQLの
// anime_classification_kind enumと対応する。
type AnimeClassificationKind string

const (
	AnimeClassificationKindWork    AnimeClassificationKind = "work"
	AnimeClassificationKindEpisode AnimeClassificationKind = "episode"
)

// Stringは種別の文字列表現を返す。
func (k AnimeClassificationKind) String() string { return string(k) }

// AnimeClassificationは第2層 (カタログ分類) のドメインエンティティ。
// UNIQUE (anime_id) によりアニメと1:1で結びつくため、同一のコンテンツを
// 同一性を変えずにworkとepisodeの間で再分類できる。
//
// Number / EpisodeStartNumberはNUMERICカラムに対応し、浮動小数の丸めなしに
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
