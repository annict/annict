package repository

import (
	"context"

	"github.com/annict/annict/go/internal/query"
)

// TrailerRepositoryはPV (trailers) のデータアクセスを担当する。
type TrailerRepository struct {
	queries *query.Queries
}

// NewTrailerRepositoryはTrailerRepositoryを作成する。
func NewTrailerRepository(queries *query.Queries) *TrailerRepository {
	return &TrailerRepository{queries: queries}
}

// ReconcileAnimeIDsAfterは、IDがafterIDより大きいPVの1ページについて、anime_idを
// 作品のanime_idと照合して直す。作品を指していない行と、作品のanime_idがまだ無い行はNULLにする。
func (r *TrailerRepository) ReconcileAnimeIDsAfter(ctx context.Context, afterID int64, batchSize int) (AnimeIDReconcilePage, error) {
	row, err := r.queries.ReconcileTrailerAnimeIDsAfter(ctx, query.ReconcileTrailerAnimeIDsAfterParams{
		AfterID: afterID,
		// batchSizeは小さく上限のあるページサイズ (既定1000) でint32上限には達しない。
		BatchSize: int32(batchSize), // #nosec G115
	})
	if err != nil {
		return AnimeIDReconcilePage{}, err
	}
	return newAnimeIDReconcilePage(row.LastID, row.UpdatedCount), nil
}
