package model

import "time"

// WorkStatus represents the lifecycle state of a work (published / archived / deleted)
// derived from unpublished_at / deleted_at.
//
// [Ja] WorkStatus は unpublished_at / deleted_at から導出した作品のライフサイクル状態
// (published / archived / deleted) を表す。
type WorkStatus string

const (
	WorkStatusPublished WorkStatus = "published"
	WorkStatusArchived  WorkStatus = "archived"
	WorkStatusDeleted   WorkStatus = "deleted"
)

// String returns the textual representation of the status.
//
// [Ja] ステータスの文字列表現を返す。
func (s WorkStatus) String() string { return string(s) }

// Work is the domain entity for an anime work, kept page-independent and generic.
// It belongs to the Domain / Infrastructure layer and must not depend on the Presentation layer.
//
// [Ja] Work は作品のドメインエンティティ (ページに依存しない汎用的な構造)。
// Domain / Infrastructure 層に属し、Presentation 層に依存しない。
type Work struct {
	ID                  WorkID
	Title               string
	TitleEn             string
	TitleKana           *string
	RecommendedImageURL string
	// JSON payload from the work_images.image_data column; empty when no work_images row is joined.
	//
	// [Ja] work_images テーブルの image_data カラム (JSON)。LEFT JOIN で行が無い場合は空文字列。
	ImageData     string
	WatchersCount int32
	SeasonYear    *int32
	// Season number: 1=winter, 2=spring, 3=summer, 4=autumn.
	//
	// [Ja] シーズン番号 (1=冬、2=春、3=夏、4=秋)
	SeasonName *int32
	CreatedAt  time.Time

	// Fields below are populated only by the anime-sync loader (ListForAnimeSyncByIDs),
	// which projects the works columns mapped onto animes / anime_classifications during
	// the phase 2 reconciliation. Other loaders leave them at their zero value.
	// AnimeID is the works.anime_id mapping column: nil means the row is not yet
	// synced to an anime.
	//
	// [Ja] 以下のフィールドは anime 同期ローダー (ListForAnimeSyncByIDs) でのみ値が入る。
	// フェーズ 2 のリコンシリエーションで animes / anime_classifications に写像する
	// works カラムを射影したもので、他のロード経路ではゼロ値のまま。
	// AnimeID は works.anime_id のマッピングカラムで、nil は未同期 (anime 未作成) を表す。
	TitleRo               string
	TitleAlter            string
	TitleAlterEn          string
	Media                 int32
	Synopsis              string
	SynopsisEn            string
	SynopsisSource        string
	SynopsisSourceEn      string
	NoEpisodes            bool
	ManualEpisodesCount   *int32
	StartEpisodeRawNumber float64
	NumberFormatID        *NumberFormatID
	AnimeID               *AnimeID

	// UnpublishedAt / DeletedAt are the source-of-truth state columns for a work
	// (Unpublishable / SoftDeletable). The phase 2 reconciliation derives
	// anime.status from them: deleted_at set -> deleted, else unpublished_at set
	// -> archived, else published. They supersede the dormant works.status column,
	// which is never written by production code.
	//
	// [Ja] UnpublishedAt / DeletedAt は作品の状態を表す正本カラム (Unpublishable /
	// SoftDeletable)。フェーズ 2 のリコンシリエーションはこれらから anime.status を
	// 導出する: deleted_at 有 -> deleted、なければ unpublished_at 有 -> archived、
	// どちらも無ければ published。本番コードから書き込まれない休眠カラム works.status に
	// 取って代わる。
	UnpublishedAt *time.Time
	DeletedAt     *time.Time

	// Fields below are populated only by the satellite-sync loader
	// (ListForSatelliteSyncByIDs), which projects the works columns mapped onto the
	// satellite tables (anime_external_ids / anime_links / anime_official_accounts /
	// anime_hashtags / anime_seasons / anime_events) during the phase 2 reconciliation.
	// AnimeID, SeasonYear and SeasonName above are reused by this loader too. Other
	// loaders leave these at their zero value. NULL-able text columns (twitter_*) and
	// integer columns (sc_tid / mal_anime_id) use pointers so "absent" is distinct
	// from the empty string / zero, while the NOT NULL DEFAULT '' url columns keep the
	// empty string and are mapped to "no row" later (in the reconcile helper).
	//
	// [Ja] 以下のフィールドは別表同期ローダー (ListForSatelliteSyncByIDs) でのみ値が入る。
	// フェーズ 2 のリコンシリエーションで別表 (anime_external_ids / anime_links /
	// anime_official_accounts / anime_hashtags / anime_seasons / anime_events) に写像する
	// works カラムを射影したもの。上の AnimeID / SeasonYear / SeasonName も本ローダーで
	// 再利用する。他のロード経路ではゼロ値のまま。NULL 許容のテキスト列 (twitter_*) と
	// integer 列 (sc_tid / mal_anime_id) は「未設定」を空文字列・0 と区別するためポインタで
	// 持ち、NOT NULL DEFAULT '' の url 列は空文字列のまま保持して後段 (リコンサイルヘルパー)
	// で「行なし」に写像する。
	ScTid             *int32
	MalAnimeID        *int32
	OfficialSiteURL   string
	OfficialSiteURLEn string
	WikipediaURL      string
	WikipediaURLEn    string
	TwitterUsername   *string
	TwitterHashtag    *string
	StartedOn         *time.Time
	EndedOn           *time.Time

	// Related entities. Set only when the caller has explicitly loaded them; nil by default.
	//
	// [Ja] 関連エンティティ。明示的にロードした場合のみセットされ、通常は nil。
	Casts  []*Cast
	Staffs []*Staff
}

// DerivedStatus returns the work's lifecycle status derived from its Unpublishable /
// SoftDeletable timestamps, the source of truth for work state. deleted_at wins over
// unpublished_at (a deleted work is deleted regardless of publish state), matching the
// Rails visibility scope only_kept = without_deleted.published (both must be NULL to be
// published). This is the single place that encodes the timestamp-to-status priority;
// the dormant works.status column is intentionally not read. Callers map the result onto
// their own enum (viewmodel.WorkStatus for display, model.AnimeStatus for the anime sync),
// so the priority never drifts between the list screen and the reconciliation.
//
// [Ja] DerivedStatus は work の状態の正本である Unpublishable / SoftDeletable タイムスタンプ
// から作品のライフサイクル状態を導出する。deleted_at が unpublished_at より優先される
// (削除済みの作品は公開状態に関わらず deleted)。これは Rails の可視性 scope
// only_kept = without_deleted.published (公開は両方が NULL のとき) に揃う。timestamps から
// status への優先順位を定めるのはこの 1 箇所で、休眠している works.status カラムは意図的に
// 読まない。呼び出し側は結果を各自の enum (表示用の viewmodel.WorkStatus、anime 同期用の
// model.AnimeStatus) に写像するため、一覧画面とリコンシリエーションの間で優先順位がずれない。
func (w *Work) DerivedStatus() WorkStatus {
	switch {
	case w.DeletedAt != nil:
		return WorkStatusDeleted
	case w.UnpublishedAt != nil:
		return WorkStatusArchived
	default:
		return WorkStatusPublished
	}
}
