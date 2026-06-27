package model

import "time"

// AnimeExternalService identifies which external anime database an
// AnimeExternalID row maps to and mirrors the anime_external_service PostgreSQL
// enum (the Syobocal title DB and MyAnimeList for now).
//
// [Ja] AnimeExternalService は AnimeExternalID の行が指す外部アニメ DB を表し、
// PostgreSQL の anime_external_service enum と対応する (現状は Syobocal の title DB と
// MyAnimeList)。
type AnimeExternalService string

const (
	AnimeExternalServiceSyobocal AnimeExternalService = "syobocal"
	AnimeExternalServiceMal      AnimeExternalService = "mal"
)

// String returns the textual representation of the service.
//
// [Ja] サービスの文字列表現を返す。
func (s AnimeExternalService) String() string { return string(s) }

// AnimeExternalID is the domain entity for anime_external_ids: it maps an anime
// to the same work's ID in an external database. It is keyed by the layer-1
// anime identity, so the mapping stays stable across re-classification. The
// external service's ID is stored as text in ExternalID (the integer Syobocal /
// MyAnimeList IDs are stringified) to stay uniform across services.
//
// [Ja] AnimeExternalID は anime_external_ids のドメインエンティティ。anime を外部
// データベース上の同一作品の ID へ対応づける。第 1 層の anime 同一性をキーにするため、
// 再分類をまたいでも対応が安定する。外部サービスの ID はサービス横断で統一するため
// ExternalID に文字列で持つ (integer の Syobocal / MyAnimeList の ID は文字列化する)。
type AnimeExternalID struct {
	ID         AnimeExternalIDID
	AnimeID    AnimeID
	Service    AnimeExternalService
	ExternalID string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
