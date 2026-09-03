package model

import "time"

// EpisodeStatus represents the lifecycle state of an episode (published / archived /
// deleted), derived from unpublished_at / deleted_at.
//
// [Ja] EpisodeStatus は unpublished_at / deleted_at から導出したエピソードのライフサイクル
// 状態 (published / archived / deleted) を表す。
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
	ID            EpisodeID
	WorkID        WorkID
	Title         *string
	TitleRo       string
	TitleEn       string
	Number        *string
	SortNumber    int32
	RawNumber     *float64
	AnimeID       *AnimeID
	ParentAnimeID *AnimeID

	// UnpublishedAt / DeletedAt are the source-of-truth state columns for an episode
	// (Unpublishable / SoftDeletable). The phase 2 reconciliation derives anime.status
	// from them through DerivedStatus. The anime-sync loader (ListForAnimeSyncByIDs)
	// and the Annict DB list loader (ListForDB) populate them; a loader that selects
	// neither column leaves both nil, so DerivedStatus reports published for such an
	// episode.
	//
	// [Ja] UnpublishedAt / DeletedAt はエピソードの状態を表す正本カラム (Unpublishable /
	// SoftDeletable)。フェーズ 2 のリコンシリエーションは DerivedStatus を通じてこれらから
	// anime.status を導出する。値が入るのは anime 同期ローダー (ListForAnimeSyncByIDs) と
	// Annict DB 一覧のローダー (ListForDB)。どちらのカラムも選択しないローダーで取得した
	// エピソードでは両方が nil のまま残るため、DerivedStatus は published を返す。
	UnpublishedAt *time.Time
	DeletedAt     *time.Time

	// UpdatedAt is when the row was last written. The Annict DB edit form loader
	// (GetForEditByID) populates it and the form carries it back as the version its
	// submit was made against, so an update can reject a submit that would silently
	// overwrite someone else's change. Loaders that do not select the column leave it
	// nil, as do rows whose persisted updated_at is NULL.
	//
	// [Ja] UpdatedAt は行が最後に書かれた時刻。値を入れるのは Annict DB 編集フォームの
	// ローダー (GetForEditByID) で、フォームは送信が前提とする版としてこれを持ち帰る。
	// 他者の変更を黙って上書きする送信を、更新側で却下できるようにするため。カラムを
	// 選択しないローダーと、保存済みの updated_at が NULL の行では nil のまま残る。
	UpdatedAt *time.Time

	// EpisodeRecordsCount is the episodes.episode_records_count counter cache. Only
	// the Annict DB list loader (ListForDB) populates it; other loaders leave it at
	// zero.
	//
	// [Ja] EpisodeRecordsCount は episodes.episode_records_count のカウンターキャッシュ。
	// 値が入るのは Annict DB 一覧のローダー (ListForDB) のみで、他のロード経路では 0 の
	// まま残る。
	EpisodeRecordsCount int32

	// PrevNumber and PrevRawNumber are the display and numeric numbers of the episode
	// that comes just before this one in sort_number order. The Annict DB list loader
	// (ListForDB) derives them from the neighbouring row; both stay nil for the work's
	// first episode and for loaders that do not derive them.
	//
	// [Ja] PrevNumber と PrevRawNumber は、sort_number 順でこのエピソードの直前に来る
	// エピソードの表示用話数と数値話数。Annict DB 一覧のローダー (ListForDB) が隣接行から
	// 導出する。作品の最初のエピソードと、導出しないローダーではいずれも nil のまま残る。
	PrevNumber    *string
	PrevRawNumber *float64
}

// DerivedStatus returns the episode's lifecycle status derived from its Unpublishable /
// SoftDeletable timestamps, the source of truth for episode state. deleted_at wins over
// unpublished_at (a deleted episode is deleted regardless of publish state), matching the
// Rails visibility scope only_kept = without_deleted.published (both must be NULL to be
// published). This is the single place that encodes the timestamp-to-status priority, so
// callers do not have to reimplement the ordering.
//
// [Ja] DerivedStatus は episode の状態の正本である Unpublishable / SoftDeletable
// タイムスタンプからエピソードのライフサイクル状態を導出する。deleted_at が unpublished_at
// より優先される (削除済みのエピソードは公開状態に関わらず deleted)。これは Rails の可視性
// scope only_kept = without_deleted.published (公開は両方が NULL のとき) に揃う。
// timestamps から status への優先順位を定めるのはこの 1 箇所で、これにより呼び出し側が
// 優先順位を再実装せずに済む。
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

// ManualEpisodeCreationRestriction names the reason an editor may not create a work's episodes
// by hand.
//
// [Ja] ManualEpisodeCreationRestriction は編集者が作品のエピソードを手動作成できない理由を
// 表す。
type ManualEpisodeCreationRestriction string

const (
	ManualEpisodeCreationAllowed        ManualEpisodeCreationRestriction = ""
	ManualEpisodeCreationEpisodesFilled ManualEpisodeCreationRestriction = "episodes_filled"
	ManualEpisodeCreationSlotsExist     ManualEpisodeCreationRestriction = "slots_exist"
)

// ManualEpisodeCreationState holds the conditions under which Rails stops a non-admin from
// creating episodes by hand: the work already has its expected number of episodes
// (Work#episodes_filled?), or it owns a slot with a start time and therefore generates its
// episodes automatically (Work#slots_exists?).
//
// [Ja] ManualEpisodeCreationState は Rails が管理者以外の手動エピソード作成を止める条件を
// 保持する。作品が予定話数までエピソードを持っている (Work#episodes_filled?) か、開始時刻を
// 持つ放送枠があってエピソードが自動生成される (Work#slots_exists?) かのいずれか。
type ManualEpisodeCreationState struct {
	EpisodesFilled bool
	SlotsExist     bool
}

// Restriction reports which reason applies, and ManualEpisodeCreationAllowed when none does.
// A work that satisfies both reasons reports the filled count, which is the order the Rails
// form states them in. Deciding that order here keeps the rejected submit and the page's
// warning naming the same reason.
//
// [Ja] Restriction はどの理由が当てはまるかを返し、いずれも当てはまらない場合は
// ManualEpisodeCreationAllowed を返す。両方に当てはまる作品は予定話数到達を報告する
// (Rails のフォームが理由を述べる順序と同じ)。順序をここで決めることで、送信の却下と
// ページの警告が同じ理由を名指しする。
func (s ManualEpisodeCreationState) Restriction() ManualEpisodeCreationRestriction {
	switch {
	case s.EpisodesFilled:
		return ManualEpisodeCreationEpisodesFilled
	case s.SlotsExist:
		return ManualEpisodeCreationSlotsExist
	default:
		return ManualEpisodeCreationAllowed
	}
}

// Allowed reports whether ordinary committers may create episodes manually.
//
// [Ja] Allowed は通常のコミッターがエピソードを手動作成できるかを返す。
func (s ManualEpisodeCreationState) Allowed() bool {
	return s.Restriction() == ManualEpisodeCreationAllowed
}
