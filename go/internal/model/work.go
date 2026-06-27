package model

import "time"

// WorkStatus represents the lifecycle state of a work and mirrors the work_status PostgreSQL enum.
//
// [Ja] WorkStatus は作品のライフサイクル状態を表し、PostgreSQL の work_status enum と対応する。
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
	// Populated only by loaders that select the works.status column (e.g. ListForDB).
	//
	// [Ja] works.status カラムを select するロード経路 (例: ListForDB) でのみ値が入る。
	Status    WorkStatus
	CreatedAt time.Time

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
	ArchiveMessage        *string
	NoEpisodes            bool
	ManualEpisodesCount   *int32
	StartEpisodeRawNumber float64
	NumberFormatID        *NumberFormatID
	AnimeID               *AnimeID

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
