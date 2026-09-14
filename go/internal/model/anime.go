package model

import (
	"database/sql"
	"time"
)

// AnimeStatusはアニメのコンテンツ同一性のライフサイクルで、PostgreSQLの
// anime_status enumと対応する。いずれの状態もソフトで、行は物理削除しない。
type AnimeStatus string

const (
	AnimeStatusPublished AnimeStatus = "published"
	AnimeStatusArchived  AnimeStatus = "archived"
	AnimeStatusMerged    AnimeStatus = "merged"
	AnimeStatusDeleted   AnimeStatus = "deleted"
)

// Stringはステータスの文字列表現を返す。
func (s AnimeStatus) String() string { return string(s) }

// AnimeMediaはアニメの配信媒体で、PostgreSQLのanime_media enumと対応する。
// ゼロ値 (空文字列) はNULL (媒体未設定) を表す。
type AnimeMedia string

const (
	AnimeMediaTV    AnimeMedia = "tv"
	AnimeMediaOVA   AnimeMedia = "ova"
	AnimeMediaMovie AnimeMedia = "movie"
	AnimeMediaONA   AnimeMedia = "ona"
	AnimeMediaOther AnimeMedia = "other"
)

// Stringは媒体の文字列表現を返す。
func (m AnimeMedia) String() string { return string(m) }

// ReleaseStatusはアニメの放送/公開ライフサイクルで、PostgreSQLの
// release_status enumと対応する。ゼロ値 (空文字列) はNULL (未設定) を表す。
type ReleaseStatus string

const (
	ReleaseStatusNotYetReleased ReleaseStatus = "not_yet_released"
	ReleaseStatusReleased       ReleaseStatus = "released"
	ReleaseStatusCancelled      ReleaseStatus = "cancelled"
)

// Stringは公開ステータスの文字列表現を返す。
func (s ReleaseStatus) String() string { return string(s) }

// Animeは第1層 (コンテンツ同一性) のドメインエンティティ。再分類で
// 変わらない内容属性を持ち、カタログ上の分類 (work / episode) は
// AnimeClassificationが持つ。
type Anime struct {
	ID               AnimeID
	Title            sql.NullString
	TitleKana        sql.NullString
	TitleRo          sql.NullString
	TitleEn          sql.NullString
	TitleAlter       sql.NullString
	TitleAlterRo     sql.NullString
	TitleAlterEn     sql.NullString
	TitleAlterOther  sql.NullString
	Media            AnimeMedia
	ReleaseStatus    ReleaseStatus
	Synopsis         sql.NullString
	SynopsisEn       sql.NullString
	SynopsisSource   sql.NullString
	SynopsisSourceEn sql.NullString
	Status           AnimeStatus
	ArchiveMessage   sql.NullString
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
