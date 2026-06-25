package model

// EpisodeStatus represents the lifecycle state of an episode and mirrors the episode_status PostgreSQL enum.
//
// [Ja] EpisodeStatus はエピソードのライフサイクル状態を表し、PostgreSQL の episode_status enum と対応する。
type EpisodeStatus string

const (
	EpisodeStatusPublished EpisodeStatus = "published"
	EpisodeStatusArchived  EpisodeStatus = "archived"
	EpisodeStatusDeleted   EpisodeStatus = "deleted"
)

// String returns the textual representation of the status.
//
// [Ja] ステータスの文字列表現を返す。
func (s EpisodeStatus) String() string { return string(s) }

// Episode is the domain entity for an episode, kept page-independent and
// generic. It belongs to the Domain layer and must not depend on the
// Presentation layer.
//
// The fields below are currently populated only by the anime-sync loader
// (ListForAnimeSyncByIDs), which projects the episodes columns mapped onto
// animes / anime_classifications during the phase 2 reconciliation. AnimeID is
// the episodes.anime_id mapping column (nil = not yet synced to an anime).
// ParentAnimeID is the parent work's anime_id resolved through episodes.work_id;
// nil means the parent work is not yet synced, so the episode cannot be
// reconciled until a later run.
//
// [Ja] Episode はエピソードのドメインエンティティ (ページに依存しない汎用的な構造)。
// Domain 層に属し、Presentation 層に依存しない。
//
// 以下のフィールドは現状 anime 同期ローダー (ListForAnimeSyncByIDs) でのみ値が入る。
// フェーズ 2 のリコンシリエーションで animes / anime_classifications に写像する
// episodes カラムを射影したもの。AnimeID は episodes.anime_id のマッピングカラム
// (nil は未同期 = anime 未作成)。ParentAnimeID は episodes.work_id 経由で解決した
// 親作品の anime_id で、nil は親作品が未同期であることを表し、その場合は後続の実行まで
// エピソードをリコンサイルできない。
type Episode struct {
	ID             EpisodeID
	WorkID         WorkID
	Title          *string
	TitleRo        string
	TitleEn        string
	Number         *string
	SortNumber     int32
	RawNumber      *float64
	Status         EpisodeStatus
	ArchiveMessage *string
	AnimeID        *AnimeID
	ParentAnimeID  *AnimeID
}
