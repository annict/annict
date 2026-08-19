package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// WorkRepository handles data access for the works table and related joins.
//
// [Ja] WorkRepository は works テーブルおよび関連 JOIN へのデータアクセスを担う。
type WorkRepository struct {
	queries *query.Queries
}

func NewWorkRepository(queries *query.Queries) *WorkRepository {
	return &WorkRepository{queries: queries}
}

func (r *WorkRepository) GetByID(ctx context.Context, id model.WorkID) (*model.Work, error) {
	row, err := r.queries.GetWorkByID(ctx, int64(id))
	if err != nil {
		return nil, err
	}
	return workFromGetByIDRow(row), nil
}

func workFromGetByIDRow(row query.GetWorkByIDRow) *model.Work {
	work := &model.Work{
		ID:                  model.WorkID(row.ID),
		Title:               row.Title,
		TitleEn:             row.TitleEn,
		RecommendedImageURL: row.RecommendedImageUrl,
		WatchersCount:       row.WatchersCount,
	}
	if row.TitleKana != "" {
		titleKana := row.TitleKana
		work.TitleKana = &titleKana
	}
	applyNullableWorkFields(work, row.SeasonYear, row.SeasonName, row.CreatedAt)
	return work
}

// GetForArchiveByID loads the columns the Annict DB admin archive-confirmation screen
// needs: the title to display and the work-state source (unpublished_at / deleted_at)
// so the caller can derive the current status and reject a work that is not archivable.
// It returns (nil, nil) when no work matches the id.
//
// [Ja] GetForArchiveByID は Annict DB 管理画面の非公開確認画面が必要とするカラムを
// 読み込む: 表示するタイトルと、呼び出し側が現在の状態を導出しアーカイブ不可の作品を
// 弾けるようにするための作品状態の source (unpublished_at / deleted_at)。該当する work が
// 無い場合は (nil, nil) を返す。
func (r *WorkRepository) GetForArchiveByID(ctx context.Context, id model.WorkID) (*model.Work, error) {
	row, err := r.queries.GetWorkForArchiveByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	work := &model.Work{
		ID:    model.WorkID(row.ID),
		Title: row.Title,
	}
	if row.UnpublishedAt.Valid {
		unpublishedAt := row.UnpublishedAt.Time
		work.UnpublishedAt = &unpublishedAt
	}
	if row.DeletedAt.Valid {
		deletedAt := row.DeletedAt.Time
		work.DeletedAt = &deletedAt
	}
	return work, nil
}

// GetForEditByID loads every works column the Annict DB admin edit form needs to
// pre-populate its fields, and returns (nil, nil) when no work matches the id.
//
// [Ja] GetForEditByID は Annict DB 管理画面の編集フォームが各フィールドを初期表示
// するために必要な works の全カラムを読み込む。該当する work が無い場合は
// (nil, nil) を返す。
func (r *WorkRepository) GetForEditByID(ctx context.Context, id model.WorkID) (*model.Work, error) {
	row, err := r.queries.GetWorkForEditByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return workFromGetForEditByIDRow(row), nil
}

func workFromGetForEditByIDRow(row query.GetWorkForEditByIDRow) *model.Work {
	work := &model.Work{
		ID:                    model.WorkID(row.ID),
		Title:                 row.Title,
		TitleAlter:            row.TitleAlter,
		TitleEn:               row.TitleEn,
		TitleAlterEn:          row.TitleAlterEn,
		Media:                 row.Media,
		OfficialSiteURL:       row.OfficialSiteUrl,
		OfficialSiteURLEn:     row.OfficialSiteUrlEn,
		WikipediaURL:          row.WikipediaUrl,
		WikipediaURLEn:        row.WikipediaUrlEn,
		Synopsis:              row.Synopsis,
		SynopsisSource:        row.SynopsisSource,
		SynopsisEn:            row.SynopsisEn,
		SynopsisSourceEn:      row.SynopsisSourceEn,
		StartEpisodeRawNumber: row.StartEpisodeRawNumber,
		NoEpisodes:            row.NoEpisodes,
	}
	if row.TitleKana != "" {
		titleKana := row.TitleKana
		work.TitleKana = &titleKana
	}
	applyNullableWorkFields(work, row.SeasonYear, row.SeasonName, sql.NullTime{})
	if row.StartedOn.Valid {
		startedOn := row.StartedOn.Time
		work.StartedOn = &startedOn
	}
	if row.EndedOn.Valid {
		endedOn := row.EndedOn.Time
		work.EndedOn = &endedOn
	}
	if row.TwitterUsername.Valid {
		twitterUsername := row.TwitterUsername.String
		work.TwitterUsername = &twitterUsername
	}
	if row.TwitterHashtag.Valid {
		twitterHashtag := row.TwitterHashtag.String
		work.TwitterHashtag = &twitterHashtag
	}
	if row.ScTid.Valid {
		scTid := row.ScTid.Int32
		work.ScTid = &scTid
	}
	if row.MalAnimeID.Valid {
		malAnimeID := row.MalAnimeID.Int32
		work.MalAnimeID = &malAnimeID
	}
	if row.ManualEpisodesCount.Valid {
		manualEpisodesCount := row.ManualEpisodesCount.Int32
		work.ManualEpisodesCount = &manualEpisodesCount
	}
	if row.NumberFormatID.Valid {
		numberFormatID := model.NumberFormatID(row.NumberFormatID.Int64)
		work.NumberFormatID = &numberFormatID
	}
	if row.UpdatedAt.Valid {
		updatedAt := row.UpdatedAt.Time
		work.UpdatedAt = &updatedAt
	}
	return work
}

// DBEpisodeListWork is the parent work of an Annict DB episode list page: the work itself,
// plus the two values the page's auto-generation notice reports. Those values aggregate the
// work's episodes and slots rather than naming columns of works, so they ride alongside the
// work instead of being folded into model.Work.
//
// [Ja] DBEpisodeListWork は Annict DB のエピソード一覧ページの親作品を表す。作品そのものと、
// ページの自動生成の案内が報告する 2 つの値を持つ。これらの値は works のカラムではなく
// 作品のエピソード・スロットの集計であるため、model.Work に畳み込まず作品と並べて持つ。
type DBEpisodeListWork struct {
	Work *model.Work
	// PublishedEpisodeCount counts the work's episodes that are neither unpublished nor
	// deleted (the Rails only_kept scope). It differs from the list's own total, which
	// keeps unpublished episodes because the list shows them.
	//
	// [Ja] PublishedEpisodeCount は作品のエピソードのうち、非公開でも削除済みでもないもの
	// (Rails の only_kept スコープ) の件数。一覧自体の総件数とは異なる (一覧は非公開の
	// エピソードも表示するため総件数に含める)。
	PublishedEpisodeCount int64
	// MaxGeneratableEpisodeNumber is the highest number among the work's kept slots: the
	// episode number the Syobocal auto-generation would reach. It is 0 while the work has
	// no such slot.
	//
	// [Ja] MaxGeneratableEpisodeNumber は作品が持つ有効なスロットの最大 number で、しょぼい
	// カレンダー由来の自動生成が到達する話数を表す。そのようなスロットが無い作品では 0。
	MaxGeneratableEpisodeNumber int64
}

// GetForEpisodeListByID loads what the Annict DB episode list needs from the parent work:
// the title for the page heading, no_episodes for the shared work subnav, and the derived
// values its auto-generation notice reports. Works whose deleted_at is set are excluded
// by the query, mirroring the Rails Work.without_deleted.find the episode list uses, so (nil, nil)
// means the id names no listable work.
//
// [Ja] GetForEpisodeListByID は Annict DB のエピソード一覧が親作品から必要とするもの
// (ページ見出しに使う title、共有の作品サブナビが使う no_episodes、自動生成の案内が報告する
// 集計値) を読み込む。deleted_at が入った作品はクエリ側で除外する (エピソード一覧が使う Rails
// の Work.without_deleted.find と同じ)。そのため (nil, nil) は一覧を出せる作品がその id に
// 無いことを表す。
func (r *WorkRepository) GetForEpisodeListByID(ctx context.Context, id model.WorkID) (*DBEpisodeListWork, error) {
	row, err := r.queries.GetWorkForEpisodeListByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	work := &model.Work{
		ID:         model.WorkID(row.ID),
		Title:      row.Title,
		NoEpisodes: row.NoEpisodes,
	}
	if row.ManualEpisodesCount.Valid {
		manualEpisodesCount := row.ManualEpisodesCount.Int32
		work.ManualEpisodesCount = &manualEpisodesCount
	}

	return &DBEpisodeListWork{
		Work:                        work,
		PublishedEpisodeCount:       row.PublishedEpisodeCount,
		MaxGeneratableEpisodeNumber: row.MaxGeneratableEpisodeNumber,
	}, nil
}

// DBEpisodeFormWork carries the parent work and the Rails-compatible manual-creation state
// rendered by the Annict DB episode form.
//
// [Ja] DBEpisodeFormWork は Annict DB エピソードフォームが描画する親作品と、Rails 互換の
// 手動作成状態を保持する。
type DBEpisodeFormWork struct {
	Work                *model.Work
	ManualCreationState model.ManualEpisodeCreationState
}

// GetForEpisodeFormByID loads what the Annict DB episode form needs from the parent work:
// the title for the page heading, no_episodes for the shared work subnav and the reasons
// manual creation may be restricted. Deleted works are excluded by the query, so (nil, nil)
// means the id names no work the form can be shown for.
//
// [Ja] GetForEpisodeFormByID は Annict DB のエピソードフォームが親作品から必要とするもの
// (ページ見出しに使う title、共有の作品サブナビが使う no_episodes、手動作成を制限する理由)
// を読み込む。削除済みの作品はクエリ側で除外するため、(nil, nil) はフォームを出せる作品が
// その id に無いことを表す。
func (r *WorkRepository) GetForEpisodeFormByID(ctx context.Context, id model.WorkID) (*DBEpisodeFormWork, error) {
	row, err := r.queries.GetWorkForEpisodeFormByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &DBEpisodeFormWork{
		Work: &model.Work{
			ID:         model.WorkID(row.ID),
			Title:      row.Title,
			NoEpisodes: row.NoEpisodes,
		},
		ManualCreationState: model.ManualEpisodeCreationState{
			EpisodesFilled: row.EpisodesFilled.Valid && row.EpisodesFilled.Bool,
			SlotsExist:     row.SlotsExist,
		},
	}, nil
}

// DBEpisodeCreateWork is the parent work of an Annict DB episode bulk create: the work
// itself (its id and the anime it is mapped to) plus the anchors the new rows are numbered
// from. The anchors aggregate the work's existing episodes rather than naming columns of
// works, so they ride alongside the work instead of being folded into model.Work.
//
// [Ja] DBEpisodeCreateWork は Annict DB のエピソード一括作成の親作品を表す。作品そのもの
// (id とマッピング先の anime) と、新規行の採番の起点になる値を持つ。起点の値は works の
// カラムではなく作品の既存エピソードの集計であるため、model.Work に畳み込まず作品と並べて
// 持つ。
type DBEpisodeCreateWork struct {
	Work                *model.Work
	ManualCreationState model.ManualEpisodeCreationState
	// EpisodeCount counts every episode of the work, including the unpublished and the
	// deleted ones, matching the Rails form the numbering is taken from.
	//
	// [Ja] EpisodeCount は作品のエピソードを、非公開のものも削除済みのものも含めて数える
	// (採番の元にした Rails のフォームと同じ)。
	EpisodeCount int64
	// LatestEpisode is the episode holding the greatest sort_number, which the first
	// created row names as its preceding episode. It is nil while the work has no episode.
	//
	// [Ja] LatestEpisode は sort_number が最大のエピソードで、最初に作る行が直前の
	// エピソードとして名指しする。作品がエピソードを持たないあいだは nil。
	LatestEpisode *DBEpisodeSortAnchor
}

// ExistsForEpisodeCreateByID reports whether the requested, non-deleted work exists. The
// create use case uses this inexpensive check before parsing the submitted rows, then locks
// and rechecks the same work inside its write transaction.
//
// [Ja] ExistsForEpisodeCreateByID は指定された未削除作品が存在するかを返す。作成ユースケースは
// 入力行をパースする前にこの軽量な確認を行い、書き込みトランザクション内で同じ作品をロック
// して再確認する。
func (r *WorkRepository) ExistsForEpisodeCreateByID(ctx context.Context, id model.WorkID) (bool, error) {
	return r.queries.ExistsWorkForEpisodeCreateByID(ctx, int64(id))
}

// LockForEpisodeCreateByID locks the requested work for the current transaction. The lock
// serializes bulk creates for one work; false means it is missing or was deleted.
//
// [Ja] LockForEpisodeCreateByID は現在のトランザクションで対象作品をロックする。このロックで
// 同じ作品への一括作成を直列化する。false は作品が存在しないか削除済みであることを表す。
func (r *WorkRepository) LockForEpisodeCreateByID(ctx context.Context, id model.WorkID) (bool, error) {
	_, err := r.queries.LockWorkForEpisodeCreateByID(ctx, int64(id))
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// DBEpisodeSortAnchor is an episode reduced to what the bulk create's numbering needs: its
// id (to be stored as the next episode's prev_episode_id) and its sort_number (to tell
// whether a newly created episode takes its place as the greatest).
//
// [Ja] DBEpisodeSortAnchor は一括作成の採番が必要とする分だけに絞ったエピソード。id (次の
// エピソードの prev_episode_id として保存する) と sort_number (新規作成したエピソードが最大
// の座を引き継ぐかの判定に使う) を持つ。
type DBEpisodeSortAnchor struct {
	ID         model.EpisodeID
	SortNumber int32
}

// GetForEpisodeCreateByID loads the parent work of an episode bulk create together with the
// numbering anchors. Deleted works are excluded by the query, mirroring the Rails
// Work.without_deleted.find the create action uses, so (nil, nil) means the id names no work
// episodes can be created under.
//
// [Ja] GetForEpisodeCreateByID はエピソード一括作成の親作品を、採番の起点と併せて読み込む。
// 削除済みの作品はクエリ側で除外する (create アクションが使う Rails の
// Work.without_deleted.find と同じ)。そのため (nil, nil) はエピソードを作成できる作品がその
// id に無いことを表す。
func (r *WorkRepository) GetForEpisodeCreateByID(ctx context.Context, id model.WorkID) (*DBEpisodeCreateWork, error) {
	row, err := r.queries.GetWorkForEpisodeCreateByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	work := &model.Work{ID: model.WorkID(row.ID)}
	if row.AnimeID.Valid {
		animeID := model.AnimeID(row.AnimeID.Int64)
		work.AnimeID = &animeID
	}

	result := &DBEpisodeCreateWork{
		Work:         work,
		EpisodeCount: row.EpisodeCount,
		ManualCreationState: model.ManualEpisodeCreationState{
			EpisodesFilled: row.EpisodesFilled.Valid && row.EpisodesFilled.Bool,
			SlotsExist:     row.SlotsExist,
		},
	}
	if row.LatestEpisodeID != 0 {
		result.LatestEpisode = &DBEpisodeSortAnchor{
			ID:         model.EpisodeID(row.LatestEpisodeID),
			SortNumber: row.LatestSortNumber,
		}
	}

	return result, nil
}

// IncrementEpisodesCount applies the Rails counter-cache and touch side effects after the
// transaction has created its published episodes. The caller already holds the work lock.
//
// [Ja] IncrementEpisodesCount は公開エピソードを作成した後、Rails のカウンターキャッシュと
// touch の副作用を適用する。呼び出し元は既に作品をロックしている。
func (r *WorkRepository) IncrementEpisodesCount(ctx context.Context, workID model.WorkID, createdCount int32) error {
	affected, err := r.queries.IncrementWorkEpisodesCount(ctx, query.IncrementWorkEpisodesCountParams{
		CreatedCount: createdCount,
		WorkID:       int64(workID),
	})
	if err != nil {
		return err
	}
	if affected != 1 {
		return errors.New("作品のエピソード件数を更新できませんでした")
	}
	return nil
}

// GetPopular returns popular works. Each *model.Work in the returned slice is
// freshly allocated on every call, so callers (typically UseCase code) are
// free to attach related entities such as Casts / Staffs to the returned
// pointers after the fact. Revisit this contract if the repository ever
// starts caching or pooling these structs.
//
// [Ja] 人気作品を返す。戻り値の各 *model.Work は呼び出しごとに新規生成されるため、
// 呼び出し側 (主に UseCase) が Casts / Staffs などの関連エンティティを後付けで
// 代入する用法を許容している。Repository でキャッシュやプール再利用を導入する
// 場合はこの前提を見直すこと。
func (r *WorkRepository) GetPopular(ctx context.Context) ([]*model.Work, error) {
	rows, err := r.queries.GetPopularWorks(ctx)
	if err != nil {
		return nil, err
	}

	works := make([]*model.Work, len(rows))
	for i, row := range rows {
		works[i] = workFromPopularRow(row)
	}
	return works, nil
}

func workFromPopularRow(row query.GetPopularWorksRow) *model.Work {
	work := &model.Work{
		ID:                  model.WorkID(row.ID),
		Title:               row.Title,
		TitleEn:             row.TitleEn,
		RecommendedImageURL: row.RecommendedImageUrl,
		WatchersCount:       row.WatchersCount,
	}
	applyImageData(work, row.ImageData)
	applyNullableWorkFields(work, row.SeasonYear, row.SeasonName, row.CreatedAt)
	return work
}

// applyNullableWorkFields maps sqlc's nullable columns onto *model.Work.
// SeasonYear / SeasonName / CreatedAt show up on multiple row types, so the
// conversion is centralised here to avoid drift between callers. Callers that
// do not load CreatedAt may pass sql.NullTime{} to skip it.
//
// [Ja] sqlc 生成型の nullable カラムを *model.Work にマッピングするヘルパー。
// SeasonYear / SeasonName / CreatedAt は複数の row 型で共通するため、
// 呼び出し元ごとに揺れないよう変換ロジックを 1 箇所に集約している。
// CreatedAt をロードしない呼び出し元は sql.NullTime{} を渡すことでスキップできる。
func applyNullableWorkFields(work *model.Work, seasonYear, seasonName sql.NullInt32, createdAt sql.NullTime) {
	if seasonYear.Valid {
		v := seasonYear.Int32
		work.SeasonYear = &v
	}
	if seasonName.Valid {
		v := seasonName.Int32
		work.SeasonName = &v
	}
	if createdAt.Valid {
		work.CreatedAt = createdAt.Time
	}
}

// applyImageData maps the work_images.image_data column onto *model.Work.
// A LEFT JOIN with no matching work_images row yields Valid=false, in which
// case ImageData stays as the empty string.
//
// [Ja] work_images.image_data カラムを *model.Work にマッピングするヘルパー。
// LEFT JOIN で work_images 行が一致しない場合は Valid=false となり、
// ImageData は空文字列のままになる。
func applyImageData(work *model.Work, imageData sql.NullString) {
	if imageData.Valid {
		work.ImageData = imageData.String
	}
}

func (r *WorkRepository) WithTx(tx *sql.Tx) *WorkRepository {
	return &WorkRepository{queries: r.queries.WithTx(tx)}
}

type DBWorkListParams struct {
	FilterNoEpisodes bool
	FilterNoImage    bool
	FilterNoSeason   bool
	FilterNoSlots    bool
	SeasonYear       *int32
	SeasonName       *int32
	// SeasonYears / SeasonNames are parallel arrays describing the (year, name)
	// pairs to match for the release-season multi-select filter. An empty slice
	// disables the filter. Both slices must have the same length; element i of
	// SeasonYears pairs with element i of SeasonNames.
	//
	// [Ja] SeasonYears / SeasonNames はリリース時期の複数選択フィルタで照合する
	// (年, 季節) ペアを表す並列配列。空スライスならフィルタは無効。両スライスは同じ
	// 長さで、SeasonYears の i 番目が SeasonNames の i 番目と対になる。
	SeasonYears []int32
	SeasonNames []int32
	Page        int32
	PerPage     int32
}

func (r *WorkRepository) ListForDB(ctx context.Context, params DBWorkListParams) ([]*model.Work, error) {
	// Widen before multiplying: callers accept any page number that fits in an int32, and at
	// 100 rows per page the int32 product wraps negative inside that range, which PostgreSQL
	// rejects as an OFFSET.
	//
	// [Ja] 乗算の前に幅を広げる。呼び出し側は int32 に収まるページ番号をすべて受け付けるが、
	// 1 ページ 100 件では int32 同士の積がその範囲内で負に折り返し、PostgreSQL がその OFFSET
	// を拒否するため。
	offset := int64(params.Page-1) * int64(params.PerPage)

	rows, err := r.queries.ListDBWorks(ctx, query.ListDBWorksParams{
		FilterNoEpisodes: sql.NullBool{Bool: params.FilterNoEpisodes, Valid: params.FilterNoEpisodes},
		FilterNoImage:    sql.NullBool{Bool: params.FilterNoImage, Valid: params.FilterNoImage},
		FilterNoSeason:   sql.NullBool{Bool: params.FilterNoSeason, Valid: params.FilterNoSeason},
		FilterNoSlots:    sql.NullBool{Bool: params.FilterNoSlots, Valid: params.FilterNoSlots},
		SeasonYear:       nullInt32FromPtr(params.SeasonYear),
		SeasonName:       nullInt32FromPtr(params.SeasonName),
		SeasonYears:      params.SeasonYears,
		SeasonNames:      params.SeasonNames,
		PerPage:          params.PerPage,
		PageOffset:       offset,
	})
	if err != nil {
		return nil, err
	}

	works := make([]*model.Work, len(rows))
	for i, row := range rows {
		work := &model.Work{
			ID:            model.WorkID(row.ID),
			Title:         row.Title,
			TitleEn:       row.TitleEn,
			Media:         row.Media,
			WatchersCount: row.WatchersCount,
		}
		if row.TitleKana != "" {
			titleKana := row.TitleKana
			work.TitleKana = &titleKana
		}
		if row.ScTid.Valid {
			scTid := row.ScTid.Int32
			work.ScTid = &scTid
		}
		if row.MalAnimeID.Valid {
			malAnimeID := row.MalAnimeID.Int32
			work.MalAnimeID = &malAnimeID
		}
		if row.UnpublishedAt.Valid {
			unpublishedAt := row.UnpublishedAt.Time
			work.UnpublishedAt = &unpublishedAt
		}
		if row.DeletedAt.Valid {
			deletedAt := row.DeletedAt.Time
			work.DeletedAt = &deletedAt
		}
		applyImageData(work, row.ImageData)
		applyNullableWorkFields(work, row.SeasonYear, row.SeasonName, sql.NullTime{})
		works[i] = work
	}
	return works, nil
}

func (r *WorkRepository) CountForDB(ctx context.Context, params DBWorkListParams) (int64, error) {
	return r.queries.CountDBWorks(ctx, query.CountDBWorksParams{
		FilterNoEpisodes: sql.NullBool{Bool: params.FilterNoEpisodes, Valid: params.FilterNoEpisodes},
		FilterNoImage:    sql.NullBool{Bool: params.FilterNoImage, Valid: params.FilterNoImage},
		FilterNoSeason:   sql.NullBool{Bool: params.FilterNoSeason, Valid: params.FilterNoSeason},
		FilterNoSlots:    sql.NullBool{Bool: params.FilterNoSlots, Valid: params.FilterNoSlots},
		SeasonYear:       nullInt32FromPtr(params.SeasonYear),
		SeasonName:       nullInt32FromPtr(params.SeasonName),
		SeasonYears:      params.SeasonYears,
		SeasonNames:      params.SeasonNames,
	})
}

// ExistsKeptByTitle reports whether a kept work already uses title. "Kept" is the Rails
// Work uniqueness scope (only_kept): deleted_at and unpublished_at both NULL, so an
// archived or deleted work never blocks a title. excludeID names the work being edited so
// an update that leaves the title untouched does not collide with itself; pass nil when
// creating.
//
// [Ja] ExistsKeptByTitle は title を使っている生存中の work があるかを返す。「生存中」は
// Rails の Work の一意性スコープ (only_kept) と同じく deleted_at と unpublished_at が
// ともに NULL の行を指し、非公開・削除済みの作品はタイトルを塞がない。excludeID は編集中の
// work を指し、タイトルを変えない更新が自分自身と衝突しないようにする。作成時は nil を渡す。
func (r *WorkRepository) ExistsKeptByTitle(ctx context.Context, title string, excludeID *model.WorkID) (bool, error) {
	params := query.ExistsKeptWorkByTitleParams{Title: title}
	if excludeID != nil {
		params.ExcludeID = sql.NullInt64{Int64: int64(*excludeID), Valid: true}
	}
	return r.queries.ExistsKeptWorkByTitle(ctx, params)
}

type CreateWorkParams struct {
	Title                 string
	TitleKana             string
	TitleAlter            string
	TitleEn               string
	TitleAlterEn          string
	Media                 int32
	SeasonYear            sql.NullInt32
	SeasonName            sql.NullInt32
	StartedOn             sql.NullTime
	EndedOn               sql.NullTime
	OfficialSiteURL       string
	OfficialSiteURLEn     string
	WikipediaURL          string
	WikipediaURLEn        string
	TwitterUsername       sql.NullString
	TwitterHashtag        sql.NullString
	ScTid                 sql.NullInt32
	MalAnimeID            sql.NullInt32
	Synopsis              string
	SynopsisSource        string
	SynopsisEn            string
	SynopsisSourceEn      string
	ManualEpisodesCount   sql.NullInt32
	StartEpisodeRawNumber float64
	NumberFormatID        sql.NullInt64
	NoEpisodes            bool
}

func (r *WorkRepository) Create(ctx context.Context, params CreateWorkParams) (model.WorkID, error) {
	id, err := r.queries.CreateWork(ctx, query.CreateWorkParams{
		Title:                 params.Title,
		TitleKana:             params.TitleKana,
		TitleAlter:            params.TitleAlter,
		TitleEn:               params.TitleEn,
		TitleAlterEn:          params.TitleAlterEn,
		Media:                 params.Media,
		SeasonYear:            params.SeasonYear,
		SeasonName:            params.SeasonName,
		StartedOn:             params.StartedOn,
		EndedOn:               params.EndedOn,
		OfficialSiteUrl:       params.OfficialSiteURL,
		OfficialSiteUrlEn:     params.OfficialSiteURLEn,
		WikipediaUrl:          params.WikipediaURL,
		WikipediaUrlEn:        params.WikipediaURLEn,
		TwitterUsername:       params.TwitterUsername,
		TwitterHashtag:        params.TwitterHashtag,
		ScTid:                 params.ScTid,
		MalAnimeID:            params.MalAnimeID,
		Synopsis:              params.Synopsis,
		SynopsisSource:        params.SynopsisSource,
		SynopsisEn:            params.SynopsisEn,
		SynopsisSourceEn:      params.SynopsisSourceEn,
		ManualEpisodesCount:   params.ManualEpisodesCount,
		StartEpisodeRawNumber: params.StartEpisodeRawNumber,
		NumberFormatID:        params.NumberFormatID,
		NoEpisodes:            params.NoEpisodes,
	})
	if err != nil {
		return 0, err
	}
	return model.WorkID(id), nil
}

// UpdateWorkParams holds the editable works columns for the Annict DB admin work
// edit form, identified by ID. It embeds CreateWorkParams (exactly the columns the
// create form writes) and adds the target ID, so the two write paths share one field
// set. status / anime_id and the derived counters are intentionally left untouched
// (status changes belong to the archive/delete flow, and anime_id is a mapping column
// the sync owns).
//
// [Ja] UpdateWorkParams は Annict DB 管理画面の作品編集フォームで編集可能な works
// カラムを ID で特定して保持する。CreateWorkParams (作成フォームが書くカラムそのもの) を
// 埋め込み、対象 ID を足すことで、作成と更新の両書き込み経路が 1 つのフィールド集合を
// 共有する。status / anime_id と派生カウンターは意図的に触れない (status 変更はアーカイブ/
// 削除フローの管轄で、anime_id は同期が持つマッピングカラム)。
type UpdateWorkParams struct {
	ID model.WorkID
	// Version is the updated_at the submit was made against. nil means the row carried no
	// updated_at when the form was opened, which the shared nullable column allows and which
	// the update matches as a version of its own.
	//
	// [Ja] Version は送信が前提とする updated_at。nil は、フォームを開いた時点で行が
	// updated_at を持っていなかったことを表す。共有カラムが NULL 許容であるため、これも 1 つの
	// 版として照合する。
	Version *time.Time
	CreateWorkParams
}

// Update overwrites the editable columns of the work with the given ID and bumps updated_at,
// reporting false when no row matched: either the work is gone or its updated_at has moved on
// since the edit form was opened. The caller turns that into a conflict rather than retrying, so
// a submit made against a stale read never overwrites the write that happened in between.
//
// [Ja] Update は指定 ID の work の編集可能カラムを上書きして updated_at を更新し、どの行も
// 一致しなかった場合に false を返す (work が失われたか、編集フォームを開いてから updated_at が
// 進んだか)。呼び出し側はこれを再試行せず競合として扱うため、古い読み取りに対する送信が、その間に
// 入った書き込みを上書きすることはない。
func (r *WorkRepository) Update(ctx context.Context, params UpdateWorkParams) (bool, error) {
	rows, err := r.queries.UpdateWork(ctx, query.UpdateWorkParams{
		ID:                    int64(params.ID),
		Version:               nullTimeFromPtr(params.Version),
		Title:                 params.Title,
		TitleKana:             params.TitleKana,
		TitleAlter:            params.TitleAlter,
		TitleEn:               params.TitleEn,
		TitleAlterEn:          params.TitleAlterEn,
		Media:                 params.Media,
		SeasonYear:            params.SeasonYear,
		SeasonName:            params.SeasonName,
		StartedOn:             params.StartedOn,
		EndedOn:               params.EndedOn,
		OfficialSiteUrl:       params.OfficialSiteURL,
		OfficialSiteUrlEn:     params.OfficialSiteURLEn,
		WikipediaUrl:          params.WikipediaURL,
		WikipediaUrlEn:        params.WikipediaURLEn,
		TwitterUsername:       params.TwitterUsername,
		TwitterHashtag:        params.TwitterHashtag,
		ScTid:                 params.ScTid,
		MalAnimeID:            params.MalAnimeID,
		Synopsis:              params.Synopsis,
		SynopsisSource:        params.SynopsisSource,
		SynopsisEn:            params.SynopsisEn,
		SynopsisSourceEn:      params.SynopsisSourceEn,
		ManualEpisodesCount:   params.ManualEpisodesCount,
		StartEpisodeRawNumber: params.StartEpisodeRawNumber,
		NumberFormatID:        params.NumberFormatID,
		NoEpisodes:            params.NoEpisodes,
	})
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}

// ListForAnimeSyncByIDs loads the works with the given IDs, projecting the columns
// the phase 2 reconciliation maps onto animes / anime_classifications (including the
// works.anime_id mapping column). Rows are ordered by id; missing IDs are silently
// skipped. An empty input returns an empty slice without querying.
//
// [Ja] ListForAnimeSyncByIDs は指定 ID の works を、フェーズ 2 のリコンシリエーションが
// animes / anime_classifications に写像するカラム (works.anime_id のマッピングカラムを
// 含む) を射影してロードする。行は id 昇順で、存在しない ID は黙って除外される。
// 空入力ではクエリせず空スライスを返す。
func (r *WorkRepository) ListForAnimeSyncByIDs(ctx context.Context, workIDs []model.WorkID) ([]*model.Work, error) {
	if len(workIDs) == 0 {
		return []*model.Work{}, nil
	}

	ids := make([]int64, len(workIDs))
	for i, id := range workIDs {
		ids[i] = int64(id)
	}

	rows, err := r.queries.ListWorksForAnimeSyncByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	works := make([]*model.Work, len(rows))
	for i, row := range rows {
		works[i] = workFromAnimeSyncRow(row)
	}
	return works, nil
}

// ListForSatelliteSyncByIDs loads the works with the given IDs, projecting the
// columns the phase 2 reconciliation maps onto the satellite tables (external IDs /
// links / official accounts / hashtags / seasons / events) plus the works.anime_id
// mapping column. Rows are ordered by id; missing IDs are silently skipped. An empty
// input returns an empty slice without querying.
//
// [Ja] ListForSatelliteSyncByIDs は指定 ID の works を、フェーズ 2 のリコンシリエーションが
// 別表 (外部 ID / リンク / 公式アカウント / ハッシュタグ / 季節 / イベント) に写像する
// カラムと works.anime_id のマッピングカラムを射影してロードする。行は id 昇順で、存在
// しない ID は黙って除外される。空入力ではクエリせず空スライスを返す。
func (r *WorkRepository) ListForSatelliteSyncByIDs(ctx context.Context, workIDs []model.WorkID) ([]*model.Work, error) {
	if len(workIDs) == 0 {
		return []*model.Work{}, nil
	}

	ids := make([]int64, len(workIDs))
	for i, id := range workIDs {
		ids[i] = int64(id)
	}

	rows, err := r.queries.ListWorksForSatelliteSyncByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	works := make([]*model.Work, len(rows))
	for i, row := range rows {
		works[i] = workFromSatelliteSyncRow(row)
	}
	return works, nil
}

// ListIDsAfter returns up to batchSize work IDs whose id is greater than afterID,
// in ascending id order. It is the keyset-pagination primitive the phase 2 batch
// job (task 2-4) uses to walk the whole works table page by page: pass afterID=0
// for the first page, then the last returned id as the cursor for the next page,
// until an empty page signals the end. Keyset (id > cursor) is used over LIMIT/
// OFFSET because the batch scans a large table and OFFSET re-reads all skipped rows
// on every page, degrading to O(n^2) over a full scan.
//
// [Ja] ListIDsAfter は afterID より大きい work ID を昇順で最大 batchSize 件返す。
// フェーズ 2 のバッチジョブ (タスク 2-4) が works テーブル全体をページ単位で走査する
// ための keyset ページネーションの基本操作で、最初のページは afterID=0 を渡し、以降は
// 直前に返った末尾の id をカーソルにして次ページを引き、空ページで終端を知る。
// LIMIT/OFFSET ではなく keyset (id > カーソル) を使うのは、バッチが大テーブルを走査する
// ため。OFFSET はページごとにスキップ分を読み直し、全件走査では O(n^2) に劣化する。
func (r *WorkRepository) ListIDsAfter(ctx context.Context, afterID model.WorkID, batchSize int) ([]model.WorkID, error) {
	rows, err := r.queries.ListWorkIDsAfter(ctx, query.ListWorkIDsAfterParams{
		AfterID: int64(afterID),
		// batchSize is a small bounded page size (default 1000), never near int32 max.
		//
		// [Ja] batchSize は小さく上限のあるページサイズ (既定 1000) で int32 上限には達しない。
		BatchSize: int32(batchSize), // #nosec G115
	})
	if err != nil {
		return nil, err
	}

	ids := make([]model.WorkID, len(rows))
	for i, id := range rows {
		ids[i] = model.WorkID(id)
	}
	return ids, nil
}

// UpdateAnimeID writes back the works.anime_id mapping column, marking the work as
// synced to the given anime. updated_at is intentionally left untouched so the
// bookkeeping write is not mistaken for a content change on the source-of-truth row.
//
// [Ja] UpdateAnimeID は works.anime_id マッピングカラムを書き戻し、作品を指定アニメへ
// 同期済みとして印付ける。updated_at は意図的に触れず、正本側の行への記帳書き込みが
// 内容変更と取り違えられないようにする。
func (r *WorkRepository) UpdateAnimeID(ctx context.Context, workID model.WorkID, animeID model.AnimeID) error {
	return r.queries.UpdateWorkAnimeID(ctx, query.UpdateWorkAnimeIDParams{
		ID:      int64(workID),
		AnimeID: sql.NullInt64{Int64: int64(animeID), Valid: true},
	})
}

// UpdateUnpublishedAt writes the works.unpublished_at state column and bumps updated_at.
// Passing a non-nil time archives the work (Unpublishable#unpublish); passing nil clears
// it, re-publishing the work (Unpublishable#publish). Unlike UpdateAnimeID this bumps
// updated_at because a publish-state change is a genuine content change, not bookkeeping.
//
// [Ja] UpdateUnpublishedAt は works.unpublished_at 状態カラムを書き込み、updated_at を
// 更新する。非 nil の時刻を渡すと作品を非公開にし (Unpublishable#unpublish)、nil を渡すと
// クリアして再公開する (Unpublishable#publish)。UpdateAnimeID と異なり updated_at を更新
// するのは、公開状態の変更が記帳ではなく実質的な内容変更であるため。
func (r *WorkRepository) UpdateUnpublishedAt(ctx context.Context, id model.WorkID, unpublishedAt *time.Time) error {
	var nullTime sql.NullTime
	if unpublishedAt != nil {
		nullTime = sql.NullTime{Time: *unpublishedAt, Valid: true}
	}
	return r.queries.UpdateWorkUnpublishedAt(ctx, query.UpdateWorkUnpublishedAtParams{
		ID:            int64(id),
		UnpublishedAt: nullTime,
	})
}

// UpdateDeletedAt writes the works.deleted_at state column and bumps updated_at. Passing a
// non-nil time soft-deletes the work (SoftDeletable#destroy); passing nil restores it. Like
// UpdateUnpublishedAt this bumps updated_at because a delete-state change is a genuine
// content change, not bookkeeping.
//
// [Ja] UpdateDeletedAt は works.deleted_at 状態カラムを書き込み、updated_at を更新する。
// 非 nil の時刻を渡すと作品をソフトデリートし (SoftDeletable#destroy)、nil を渡すと復元する。
// UpdateUnpublishedAt と同じく updated_at を更新するのは、削除状態の変更が記帳ではなく実質的な
// 内容変更であるため。
func (r *WorkRepository) UpdateDeletedAt(ctx context.Context, id model.WorkID, deletedAt *time.Time) error {
	var nullTime sql.NullTime
	if deletedAt != nil {
		nullTime = sql.NullTime{Time: *deletedAt, Valid: true}
	}
	return r.queries.UpdateWorkDeletedAt(ctx, query.UpdateWorkDeletedAtParams{
		ID:        int64(id),
		DeletedAt: nullTime,
	})
}

// workFromAnimeSyncRow converts an anime-sync query row into *model.Work. The works
// text columns are NOT NULL DEFAULT ”, so the empty string is preserved here and
// mapped to NULL later (in the sync usecase) where animes uses NULL for "absent".
//
// [Ja] workFromAnimeSyncRow は anime 同期の query 行を *model.Work に変換する。
// works のテキストカラムは NOT NULL DEFAULT ” なのでここでは空文字列のまま保持し、
// animes が「未設定」を NULL で表す都合に合わせて後段 (同期 UseCase) で NULL に写像する。
func workFromAnimeSyncRow(row query.ListWorksForAnimeSyncByIDsRow) *model.Work {
	work := &model.Work{
		ID:                    model.WorkID(row.ID),
		Title:                 row.Title,
		TitleRo:               row.TitleRo,
		TitleEn:               row.TitleEn,
		TitleAlter:            row.TitleAlter,
		TitleAlterEn:          row.TitleAlterEn,
		Media:                 row.Media,
		Synopsis:              row.Synopsis,
		SynopsisEn:            row.SynopsisEn,
		SynopsisSource:        row.SynopsisSource,
		SynopsisSourceEn:      row.SynopsisSourceEn,
		NoEpisodes:            row.NoEpisodes,
		StartEpisodeRawNumber: row.StartEpisodeRawNumber,
	}
	if row.TitleKana != "" {
		titleKana := row.TitleKana
		work.TitleKana = &titleKana
	}
	if row.UnpublishedAt.Valid {
		unpublishedAt := row.UnpublishedAt.Time
		work.UnpublishedAt = &unpublishedAt
	}
	if row.DeletedAt.Valid {
		deletedAt := row.DeletedAt.Time
		work.DeletedAt = &deletedAt
	}
	if row.ManualEpisodesCount.Valid {
		manualEpisodesCount := row.ManualEpisodesCount.Int32
		work.ManualEpisodesCount = &manualEpisodesCount
	}
	if row.NumberFormatID.Valid {
		numberFormatID := model.NumberFormatID(row.NumberFormatID.Int64)
		work.NumberFormatID = &numberFormatID
	}
	if row.AnimeID.Valid {
		animeID := model.AnimeID(row.AnimeID.Int64)
		work.AnimeID = &animeID
	}
	return work
}

// workFromSatelliteSyncRow converts a satellite-sync query row into *model.Work. Only
// the columns the satellite tables are sourced from (plus id / anime_id) are projected;
// the rest of *model.Work stays at its zero value. The NOT NULL DEFAULT ” url columns
// keep the empty string here and are mapped to "no row" later (in the reconcile helper),
// mirroring how workFromAnimeSyncRow defers the empty-to-NULL mapping to the sync usecase.
//
// [Ja] workFromSatelliteSyncRow は別表同期の query 行を *model.Work に変換する。別表が
// source とするカラム (と id / anime_id) だけを射影し、残りの *model.Work フィールドは
// ゼロ値のまま。NOT NULL DEFAULT ” の url 列はここでは空文字列のまま保持し、後段
// (リコンサイルヘルパー) で「行なし」に写像する。workFromAnimeSyncRow が空→NULL の写像を
// 同期 UseCase に委ねるのと同じ扱い。
func workFromSatelliteSyncRow(row query.ListWorksForSatelliteSyncByIDsRow) *model.Work {
	work := &model.Work{
		ID:                model.WorkID(row.ID),
		OfficialSiteURL:   row.OfficialSiteUrl,
		OfficialSiteURLEn: row.OfficialSiteUrlEn,
		WikipediaURL:      row.WikipediaUrl,
		WikipediaURLEn:    row.WikipediaUrlEn,
	}
	if row.AnimeID.Valid {
		animeID := model.AnimeID(row.AnimeID.Int64)
		work.AnimeID = &animeID
	}
	if row.ScTid.Valid {
		scTid := row.ScTid.Int32
		work.ScTid = &scTid
	}
	if row.MalAnimeID.Valid {
		malAnimeID := row.MalAnimeID.Int32
		work.MalAnimeID = &malAnimeID
	}
	if row.TwitterUsername.Valid {
		twitterUsername := row.TwitterUsername.String
		work.TwitterUsername = &twitterUsername
	}
	if row.TwitterHashtag.Valid {
		twitterHashtag := row.TwitterHashtag.String
		work.TwitterHashtag = &twitterHashtag
	}
	if row.SeasonYear.Valid {
		seasonYear := row.SeasonYear.Int32
		work.SeasonYear = &seasonYear
	}
	if row.SeasonName.Valid {
		seasonName := row.SeasonName.Int32
		work.SeasonName = &seasonName
	}
	if row.StartedOn.Valid {
		startedOn := row.StartedOn.Time
		work.StartedOn = &startedOn
	}
	if row.EndedOn.Valid {
		endedOn := row.EndedOn.Time
		work.EndedOn = &endedOn
	}
	return work
}

func nullInt32FromPtr(v *int32) sql.NullInt32 {
	if v == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: *v, Valid: true}
}
