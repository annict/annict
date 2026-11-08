package repository

import (
	"context"

	"github.com/annict/annict/go/internal/query"
)

// WorkImageRepositoryは作品画像 (work_images) のデータアクセスを担当する。
type WorkImageRepository struct {
	queries *query.Queries
}

// NewWorkImageRepositoryはWorkImageRepositoryを作成する。
func NewWorkImageRepository(queries *query.Queries) *WorkImageRepository {
	return &WorkImageRepository{queries: queries}
}

// ReconcileAnimeIDsAfterは、IDがafterIDより大きい作品画像の1ページについて、anime_idを
// 作品のanime_idと照合して直す。作品を指していない行と、作品のanime_idがまだ無い行はNULLにする。
func (r *WorkImageRepository) ReconcileAnimeIDsAfter(ctx context.Context, afterID int64, batchSize int) (AnimeIDReconcilePage, error) {
	row, err := r.queries.ReconcileWorkImageAnimeIDsAfter(ctx, query.ReconcileWorkImageAnimeIDsAfterParams{
		AfterID: afterID,
		// batchSizeは小さく上限のあるページサイズ (既定1000) でint32上限には達しない。
		BatchSize: int32(batchSize), // #nosec G115
	})
	if err != nil {
		return AnimeIDReconcilePage{}, err
	}
	return newAnimeIDReconcilePage(row.LastID, row.UpdatedCount), nil
}
