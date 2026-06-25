package repository

import (
	"context"
	"database/sql"

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
	SeasonYear       *int32
	SeasonName       *int32
	Page             int32
	PerPage          int32
}

func (r *WorkRepository) ListForDB(ctx context.Context, params DBWorkListParams) ([]*model.Work, error) {
	offset := (params.Page - 1) * params.PerPage

	rows, err := r.queries.ListDBWorks(ctx, query.ListDBWorksParams{
		FilterNoEpisodes: sql.NullBool{Bool: params.FilterNoEpisodes, Valid: params.FilterNoEpisodes},
		FilterNoImage:    sql.NullBool{Bool: params.FilterNoImage, Valid: params.FilterNoImage},
		FilterNoSeason:   sql.NullBool{Bool: params.FilterNoSeason, Valid: params.FilterNoSeason},
		SeasonYear:       nullInt32FromPtr(params.SeasonYear),
		SeasonName:       nullInt32FromPtr(params.SeasonName),
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
			WatchersCount: row.WatchersCount,
			Status:        model.WorkStatus(row.Status),
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
		SeasonYear:       nullInt32FromPtr(params.SeasonYear),
		SeasonName:       nullInt32FromPtr(params.SeasonName),
	})
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
		Status:                model.WorkStatus(row.Status),
		NoEpisodes:            row.NoEpisodes,
		StartEpisodeRawNumber: row.StartEpisodeRawNumber,
	}
	if row.TitleKana != "" {
		titleKana := row.TitleKana
		work.TitleKana = &titleKana
	}
	if row.ArchiveMessage.Valid {
		archiveMessage := row.ArchiveMessage.String
		work.ArchiveMessage = &archiveMessage
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

func nullInt32FromPtr(v *int32) sql.NullInt32 {
	if v == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: *v, Valid: true}
}
