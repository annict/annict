package model

import "time"

// AnimeEventKindはAnimeEventの種別を表し、PostgreSQLのanime_event_kind enumと
// 対応する。worksがsourceとするのはbroadcastイベントのみ (started_on / ended_on由来) で、
// revival_screening / otherは編集者がanimeに直接足すイベント向けに予約する。
type AnimeEventKind string

const (
	AnimeEventKindBroadcast        AnimeEventKind = "broadcast"
	AnimeEventKindRevivalScreening AnimeEventKind = "revival_screening"
	AnimeEventKindOther            AnimeEventKind = "other"
)

// Stringはイベント種別の文字列表現を返す。
func (k AnimeEventKind) String() string { return string(k) }

// AnimeEventはanime_eventsのドメインエンティティ。animeのカレンダーイベント
// (放送期間など) を (anime_id, kind) をキーに持つ。第1層のanime同一性をキーにするため、
// 再分類をまたいでもイベントが紐づき続ける。StartedOnは開始日 (NOT NULL)、EndedOnは任意の
// 終了日で、終了が未定・不明のときはnil。Title / TitleEn / Description / DescriptionEnは
// 任意のラベルで、無い場合はnil (worksはsourceしないため、同期した行ではnilのまま)。
type AnimeEvent struct {
	ID            AnimeEventID
	AnimeID       AnimeID
	Kind          AnimeEventKind
	StartedOn     time.Time
	EndedOn       *time.Time
	Title         *string
	TitleEn       *string
	Description   *string
	DescriptionEn *string
	SortNumber    int32
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
