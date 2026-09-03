package model

import (
	"database/sql"
	"time"
)

// AnimeStatus is the content-identity lifecycle of an anime and mirrors the
// anime_status PostgreSQL enum. All states are soft; rows are never physically
// deleted.
//
// [Ja] AnimeStatus はアニメのコンテンツ同一性のライフサイクルで、PostgreSQL の
// anime_status enum と対応する。いずれの状態もソフトで、行は物理削除しない。
type AnimeStatus string

const (
	AnimeStatusPublished AnimeStatus = "published"
	AnimeStatusArchived  AnimeStatus = "archived"
	AnimeStatusMerged    AnimeStatus = "merged"
	AnimeStatusDeleted   AnimeStatus = "deleted"
)

// String returns the textual representation of the status.
//
// [Ja] ステータスの文字列表現を返す。
func (s AnimeStatus) String() string { return string(s) }

// AnimeMedia is the delivery medium of an anime and mirrors the anime_media
// PostgreSQL enum. The zero value (empty string) represents NULL (medium not
// set).
//
// [Ja] AnimeMedia はアニメの配信媒体で、PostgreSQL の anime_media enum と対応する。
// ゼロ値 (空文字列) は NULL (媒体未設定) を表す。
type AnimeMedia string

const (
	AnimeMediaTV    AnimeMedia = "tv"
	AnimeMediaOVA   AnimeMedia = "ova"
	AnimeMediaMovie AnimeMedia = "movie"
	AnimeMediaONA   AnimeMedia = "ona"
	AnimeMediaOther AnimeMedia = "other"
)

// String returns the textual representation of the medium.
//
// [Ja] 媒体の文字列表現を返す。
func (m AnimeMedia) String() string { return string(m) }

// ReleaseStatus is the broadcast/release lifecycle of an anime and mirrors the
// release_status PostgreSQL enum. The zero value (empty string) represents NULL
// (status not set).
//
// [Ja] ReleaseStatus はアニメの放送/公開ライフサイクルで、PostgreSQL の
// release_status enum と対応する。ゼロ値 (空文字列) は NULL (未設定) を表す。
type ReleaseStatus string

const (
	ReleaseStatusNotYetReleased ReleaseStatus = "not_yet_released"
	ReleaseStatusReleased       ReleaseStatus = "released"
	ReleaseStatusCancelled      ReleaseStatus = "cancelled"
)

// String returns the textual representation of the release status.
//
// [Ja] 公開ステータスの文字列表現を返す。
func (s ReleaseStatus) String() string { return string(s) }

// Anime is the domain entity for layer 1 (content identity). It holds the
// re-classification-invariant content attributes; the catalog classification
// (work / episode) lives in AnimeClassification.
//
// [Ja] Anime は第 1 層 (コンテンツ同一性) のドメインエンティティ。再分類で
// 変わらない内容属性を持ち、カタログ上の分類 (work / episode) は
// AnimeClassification が持つ。
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
