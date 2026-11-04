package repository

import (
	"context"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// AnimeLegacyIDRepositoryは、animeに写した旧作品・旧エピソードのIDを、旧テーブル (works・episodes) から読む。
// 旧テーブルを読むのはこのリポジトリに閉じ、テーブルを再編してもこのリポジトリの実装だけを差し替える
type AnimeLegacyIDRepository struct {
	queries *query.Queries
}

// NewAnimeLegacyIDRepositoryはAnimeLegacyIDRepositoryを生成する
func NewAnimeLegacyIDRepository(queries *query.Queries) *AnimeLegacyIDRepository {
	return &AnimeLegacyIDRepository{queries: queries}
}

// ListByAnimeIDsは、指定したanimeの旧IDの組を返す。
// 旧作品・旧エピソードのどちらからも写していないanimeは、結果に含めない。空入力ではクエリしない
func (r *AnimeLegacyIDRepository) ListByAnimeIDs(ctx context.Context, animeIDs []model.AnimeID) ([]*model.AnimeLegacyID, error) {
	if len(animeIDs) == 0 {
		return []*model.AnimeLegacyID{}, nil
	}

	rawIDs := make([]int64, len(animeIDs))
	for i, id := range animeIDs {
		rawIDs[i] = int64(id)
	}

	workRows, err := r.queries.ListWorkIDsByAnimeIDs(ctx, rawIDs)
	if err != nil {
		return nil, err
	}
	episodeRows, err := r.queries.ListEpisodeIDsByAnimeIDs(ctx, rawIDs)
	if err != nil {
		return nil, err
	}

	legacyIDs := make([]*model.AnimeLegacyID, 0, len(workRows)+len(episodeRows))
	for _, row := range workRows {
		workID := model.WorkID(row.ID)
		legacyIDs = append(legacyIDs, &model.AnimeLegacyID{AnimeID: model.AnimeID(row.AnimeID), WorkID: &workID})
	}
	for _, row := range episodeRows {
		episodeID := model.EpisodeID(row.ID)
		legacyIDs = append(legacyIDs, &model.AnimeLegacyID{AnimeID: model.AnimeID(row.AnimeID), EpisodeID: &episodeID})
	}
	return legacyIDs, nil
}

// ListByWorkIDsは、指定した旧作品の旧IDの組を返す。
// animeに写していない旧作品と、存在しない旧作品は、結果に含めない。空入力ではクエリしない
func (r *AnimeLegacyIDRepository) ListByWorkIDs(ctx context.Context, workIDs []model.WorkID) ([]*model.AnimeLegacyID, error) {
	if len(workIDs) == 0 {
		return []*model.AnimeLegacyID{}, nil
	}

	rawIDs := make([]int64, len(workIDs))
	for i, id := range workIDs {
		rawIDs[i] = int64(id)
	}
	rows, err := r.queries.ListAnimeIDsByWorkIDs(ctx, rawIDs)
	if err != nil {
		return nil, err
	}

	legacyIDs := make([]*model.AnimeLegacyID, len(rows))
	for i, row := range rows {
		workID := model.WorkID(row.ID)
		legacyIDs[i] = &model.AnimeLegacyID{AnimeID: model.AnimeID(row.AnimeID), WorkID: &workID}
	}
	return legacyIDs, nil
}

// ListByEpisodeIDsは、指定した旧エピソードの旧IDの組を返す。
// animeに写していない旧エピソードと、存在しない旧エピソードは、結果に含めない。空入力ではクエリしない
func (r *AnimeLegacyIDRepository) ListByEpisodeIDs(ctx context.Context, episodeIDs []model.EpisodeID) ([]*model.AnimeLegacyID, error) {
	if len(episodeIDs) == 0 {
		return []*model.AnimeLegacyID{}, nil
	}

	rawIDs := make([]int64, len(episodeIDs))
	for i, id := range episodeIDs {
		rawIDs[i] = int64(id)
	}
	rows, err := r.queries.ListAnimeIDsByEpisodeIDs(ctx, rawIDs)
	if err != nil {
		return nil, err
	}

	legacyIDs := make([]*model.AnimeLegacyID, len(rows))
	for i, row := range rows {
		episodeID := model.EpisodeID(row.ID)
		legacyIDs[i] = &model.AnimeLegacyID{AnimeID: model.AnimeID(row.AnimeID), EpisodeID: &episodeID}
	}
	return legacyIDs, nil
}
