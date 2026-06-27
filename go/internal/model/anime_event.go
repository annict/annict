package model

import "time"

// AnimeEventKind identifies the kind of an AnimeEvent and mirrors the anime_event_kind
// PostgreSQL enum. Works are the source of the broadcast event only (from started_on /
// ended_on); revival_screening and other are reserved for events an editor adds to an
// anime directly.
//
// [Ja] AnimeEventKind は AnimeEvent の種別を表し、PostgreSQL の anime_event_kind enum と
// 対応する。works が source とするのは broadcast イベントのみ (started_on / ended_on 由来) で、
// revival_screening / other は編集者が anime に直接足すイベント向けに予約する。
type AnimeEventKind string

const (
	AnimeEventKindBroadcast        AnimeEventKind = "broadcast"
	AnimeEventKindRevivalScreening AnimeEventKind = "revival_screening"
	AnimeEventKindOther            AnimeEventKind = "other"
)

// String returns the textual representation of the event kind.
//
// [Ja] イベント種別の文字列表現を返す。
func (k AnimeEventKind) String() string { return string(k) }

// AnimeEvent is the domain entity for anime_events: it holds the calendar events of an
// anime (its broadcast period, ...) keyed by (anime_id, kind). It is keyed by the
// layer-1 anime identity, so an event stays attached across re-classification. StartedOn
// is the start date (NOT NULL); EndedOn is the optional end date and is nil when the
// event is open-ended or its end is unknown. Title / TitleEn / Description /
// DescriptionEn are the optional labels and are nil when absent (works do not source
// them, so synced rows leave them nil).
//
// [Ja] AnimeEvent は anime_events のドメインエンティティ。anime のカレンダーイベント
// (放送期間など) を (anime_id, kind) をキーに持つ。第 1 層の anime 同一性をキーにするため、
// 再分類をまたいでもイベントが紐づき続ける。StartedOn は開始日 (NOT NULL)、EndedOn は任意の
// 終了日で、終了が未定・不明のときは nil。Title / TitleEn / Description / DescriptionEn は
// 任意のラベルで、無い場合は nil (works は source しないため、同期した行では nil のまま)。
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
