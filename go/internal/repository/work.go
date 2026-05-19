package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// WorkRepository handles data access for the works table and related joins.
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

	if row.ImageData.Valid {
		work.ImageData = row.ImageData.String
	}
	applyNullableWorkFields(work, row.SeasonYear, row.SeasonName, row.CreatedAt)
	return work
}

// applyNullableWorkFields maps sqlc's nullable columns onto *model.Work.
// SeasonYear / SeasonName / CreatedAt show up on multiple row types, so the
// conversion is centralised here to avoid drift between callers.
//
// [Ja] sqlc 生成型の nullable カラムを *model.Work にマッピングするヘルパー。
// SeasonYear / SeasonName / CreatedAt は複数の row 型で共通するため、
// 呼び出し元ごとに揺れないよう変換ロジックを 1 箇所に集約している。
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

func (r *WorkRepository) ListForDB(ctx context.Context, params DBWorkListParams) ([]model.DBWorkListItem, error) {
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

	items := make([]model.DBWorkListItem, len(rows))
	for i, row := range rows {
		items[i] = model.DBWorkListItem{
			ID:            model.WorkID(row.ID),
			Title:         row.Title,
			WatchersCount: row.WatchersCount,
			Status:        string(row.Status),
			HasImage:      row.HasImage,
		}
		if row.SeasonYear.Valid {
			items[i].SeasonYear = &row.SeasonYear.Int32
		}
		if row.SeasonName.Valid {
			items[i].SeasonName = &row.SeasonName.Int32
		}
	}
	return items, nil
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

func nullInt32FromPtr(v *int32) sql.NullInt32 {
	if v == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: *v, Valid: true}
}
