package model

import "time"

// AnimeExternalServiceはAnimeExternalIDの行が指す外部アニメDBを表し、
// PostgreSQLのanime_external_service enumと対応する (現状はSyobocalのtitle DBと
// MyAnimeList)。
type AnimeExternalService string

const (
	AnimeExternalServiceSyobocal AnimeExternalService = "syobocal"
	AnimeExternalServiceMal      AnimeExternalService = "mal"
)

// Stringはサービスの文字列表現を返す。
func (s AnimeExternalService) String() string { return string(s) }

// AnimeExternalIDはanime_external_idsのドメインエンティティ。animeを外部
// データベース上の同一作品のIDへ対応づける。第1層のanime同一性をキーにするため、
// 再分類をまたいでも対応が安定する。外部サービスのIDはサービス横断で統一するため
// ExternalIDに文字列で持つ (integerのSyobocal / MyAnimeListのIDは文字列化する)。
type AnimeExternalID struct {
	ID         AnimeExternalIDID
	AnimeID    AnimeID
	Service    AnimeExternalService
	ExternalID string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
