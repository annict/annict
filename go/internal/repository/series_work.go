package repository

import (
	"context"

	"github.com/annict/annict/go/internal/query"
)

// SeriesWorkRepositoryはシリーズの作品 (series_works) のデータアクセスを担当する。
type SeriesWorkRepository struct {
	queries *query.Queries
}

// NewSeriesWorkRepositoryはSeriesWorkRepositoryを作成する。
func NewSeriesWorkRepository(queries *query.Queries) *SeriesWorkRepository {
	return &SeriesWorkRepository{queries: queries}
}

// ReconcileAnimeIDsAfterは、IDがafterIDより大きいシリーズの作品の1ページについて、anime_idを
// 作品のanime_idと照合して直す。作品を指していない行と、作品のanime_idがまだ無い行はNULLにする。
func (r *SeriesWorkRepository) ReconcileAnimeIDsAfter(ctx context.Context, afterID int64, batchSize int) (AnimeIDReconcilePage, error) {
	row, err := r.queries.ReconcileSeriesWorkAnimeIDsAfter(ctx, query.ReconcileSeriesWorkAnimeIDsAfterParams{
		AfterID: afterID,
		// batchSizeは小さく上限のあるページサイズ (既定1000) でint32上限には達しない。
		BatchSize: int32(batchSize), // #nosec G115
	})
	if err != nil {
		return AnimeIDReconcilePage{}, err
	}
	return newAnimeIDReconcilePage(row.LastID, row.UpdatedCount), nil
}
