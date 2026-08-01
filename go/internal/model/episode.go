package model

import "time"

// EpisodeStatus represents the lifecycle state of an episode (published / archived /
// deleted), derived from unpublished_at / deleted_at. It also types the dormant
// episodes.status column, whose enum carries the same three values.
//
// [Ja] EpisodeStatus は unpublished_at / deleted_at から導出したエピソードのライフサイクル
// 状態 (published / archived / deleted) を表す。同じ 3 値を持つ休眠カラム
// episodes.status の型も兼ねる。
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
// AnimeID and ParentAnimeID are populated only by the anime-sync loader
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
// AnimeID と ParentAnimeID に値が入るのは anime 同期ローダー
// (ListForAnimeSyncByIDs) のみ。フェーズ 2 のリコンシリエーションで animes /
// anime_classifications に写像する episodes カラムを射影したもの。AnimeID は
// episodes.anime_id のマッピングカラム (nil は未同期 = anime 未作成)。ParentAnimeID は
// episodes.work_id 経由で解決した親作品の anime_id で、nil は親作品が未同期であることを
// 表し、その場合は後続の実行までエピソードをリコンサイルできない。
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

	// UnpublishedAt / DeletedAt are the source-of-truth state columns for an episode
	// (Unpublishable / SoftDeletable). The phase 2 reconciliation derives anime.status
	// from them through DerivedStatus. They supersede the dormant episodes.status
	// column, which is never written by production code. The anime-sync loader
	// (ListForAnimeSyncByIDs) and the Annict DB list loader (ListForDB) populate them;
	// a loader that selects neither column leaves both nil, so DerivedStatus reports
	// published for such an episode.
	//
	// [Ja] UnpublishedAt / DeletedAt はエピソードの状態を表す正本カラム (Unpublishable /
	// SoftDeletable)。フェーズ 2 のリコンシリエーションは DerivedStatus を通じてこれらから
	// anime.status を導出する。本番コードから書き込まれない休眠カラム episodes.status に
	// 取って代わる。値が入るのは anime 同期ローダー (ListForAnimeSyncByIDs) と Annict DB
	// 一覧のローダー (ListForDB)。どちらのカラムも選択しないローダーで取得したエピソードでは
	// 両方が nil のまま残るため、DerivedStatus は published を返す。
	UnpublishedAt *time.Time
	DeletedAt     *time.Time

	// EpisodeRecordsCount is the episodes.episode_records_count counter cache. Only
	// the Annict DB list loader (ListForDB) populates it; other loaders leave it at
	// zero.
	//
	// [Ja] EpisodeRecordsCount は episodes.episode_records_count のカウンターキャッシュ。
	// 値が入るのは Annict DB 一覧のローダー (ListForDB) のみで、他のロード経路では 0 の
	// まま残る。
	EpisodeRecordsCount int32
}

// DerivedStatus returns the episode's lifecycle status derived from its Unpublishable /
// SoftDeletable timestamps, the source of truth for episode state. deleted_at wins over
// unpublished_at (a deleted episode is deleted regardless of publish state), matching the
// Rails visibility scope only_kept = without_deleted.published (both must be NULL to be
// published). This is the single place that encodes the timestamp-to-status priority;
// the dormant episodes.status column is intentionally not read, so callers do not have
// to reimplement the ordering.
//
// [Ja] DerivedStatus は episode の状態の正本である Unpublishable / SoftDeletable
// タイムスタンプからエピソードのライフサイクル状態を導出する。deleted_at が unpublished_at
// より優先される (削除済みのエピソードは公開状態に関わらず deleted)。これは Rails の可視性
// scope only_kept = without_deleted.published (公開は両方が NULL のとき) に揃う。
// timestamps から status への優先順位を定めるのはこの 1 箇所で、休眠している episodes.status
// カラムは意図的に読まない。これにより、呼び出し側が優先順位を再実装せずに済む。
func (e *Episode) DerivedStatus() EpisodeStatus {
	switch {
	case e.DeletedAt != nil:
		return EpisodeStatusDeleted
	case e.UnpublishedAt != nil:
		return EpisodeStatusArchived
	default:
		return EpisodeStatusPublished
	}
}
