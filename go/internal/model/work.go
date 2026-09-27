package model

import "time"

// WorkStatusはunpublished_at / deleted_atから導出した作品のライフサイクル状態
// (published / archived / deleted) を表す。
type WorkStatus string

const (
	WorkStatusPublished WorkStatus = "published"
	WorkStatusArchived  WorkStatus = "archived"
	WorkStatusDeleted   WorkStatus = "deleted"
)

// Stringはステータスの文字列表現を返す。
func (s WorkStatus) String() string { return string(s) }

// Workは作品のドメインエンティティ (ページに依存しない汎用的な構造)。
// Domain / Infrastructure層に属し、Presentation層に依存しない。
type Work struct {
	ID                  WorkID
	Title               string
	TitleEn             string
	TitleKana           *string
	RecommendedImageURL string
	// work_imagesテーブルのimage_dataカラム (JSON)。LEFT JOINで行が無い場合は空文字列。
	ImageData     string
	WatchersCount int32
	SeasonYear    *int32
	// シーズン番号 (1=冬、2=春、3=夏、4=秋)
	SeasonName *int32
	CreatedAt  time.Time

	// 以下のフィールドはanime同期ローダー (ListForAnimeSyncByIDs) でのみ値が入る。
	// フェーズ2のリコンシリエーションでanimes / anime_classificationsに写像する
	// worksカラムを射影したもので、他のロード経路ではゼロ値のまま。
	// AnimeIDはworks.anime_idのマッピングカラムで、nilは未同期 (anime未作成) を表す。
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

	// UnpublishedAt / DeletedAtは作品の状態を表す正本カラム (Unpublishable /
	// SoftDeletable)。フェーズ2のリコンシリエーションはこれらからanime.statusを
	// 導出する: deleted_at有 -> deleted、なければunpublished_at有 -> archived、
	// どちらも無ければpublished。
	UnpublishedAt *time.Time
	DeletedAt     *time.Time

	// UpdatedAtは行が最後に書かれた時刻。値を入れるのはAnnict DB編集フォームの
	// ローダー (GetForEditByID) で、フォームは送信が前提とする版としてこれを持ち帰る。
	// 他者の変更を黙って上書きする送信を、更新側で却下できるようにするため。カラムを選択
	// しないローダーと、保存済みのupdated_atがNULLの行ではnilのまま残る。
	UpdatedAt *time.Time

	// 以下のフィールドは別表同期ローダー (ListForSatelliteSyncByIDs) でのみ値が入る。
	// フェーズ2のリコンシリエーションで別表 (anime_external_ids / anime_links /
	// anime_official_accounts / anime_hashtags / anime_seasons / anime_events) に写像する
	// worksカラムを射影したもの。上のAnimeID / SeasonYear / SeasonNameも本ローダーで
	// 再利用する。他のロード経路ではゼロ値のまま。NULL許容のテキスト列 (twitter_*) と
	// integer列 (sc_tid / mal_anime_id) は「未設定」を空文字列・0と区別するためポインタで
	// 持ち、NOT NULL DEFAULT '' のurl列は空文字列のまま保持して後段 (リコンサイルヘルパー)
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

	// 関連エンティティ。明示的にロードした場合のみセットされ、通常はnil。
	Casts  []*Cast
	Staffs []*Staff
}

// DerivedStatusはworkの状態の正本であるUnpublishable / SoftDeletableタイムスタンプ
// から作品のライフサイクル状態を導出する。deleted_atがunpublished_atより優先される
// (削除済みの作品は公開状態に関わらずdeleted)。これはRailsの可視性scope
// only_kept = without_deleted.published (公開は両方がNULLのとき) に揃う。timestampsから
// statusへの優先順位を定めるのはこの1箇所。呼び出し側は結果を各自のenum (表示用の
// viewmodel.PublishingStatus、anime同期用のmodel.AnimeStatus) に写像するため、一覧画面と
// リコンシリエーションの間で優先順位がずれない。
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
