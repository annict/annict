package repository

import (
	"context"

	"github.com/annict/annict/go/internal/query"
)

// VodTitleRepositoryはVODタイトル (vod_titles) のデータアクセスを担当する。
type VodTitleRepository struct {
	queries *query.Queries
}

// NewVodTitleRepositoryはVodTitleRepositoryを作成する。
func NewVodTitleRepository(queries *query.Queries) *VodTitleRepository {
	return &VodTitleRepository{queries: queries}
}

// ReconcileAnimeIDsAfterは、IDがafterIDより大きいVODタイトルの1ページについて、anime_idを
// 作品のanime_idと照合して直す。作品を指していない行と、作品のanime_idがまだ無い行はNULLにする。
func (r *VodTitleRepository) ReconcileAnimeIDsAfter(ctx context.Context, afterID int64, batchSize int) (AnimeIDReconcilePage, error) {
	row, err := r.queries.ReconcileVodTitleAnimeIDsAfter(ctx, query.ReconcileVodTitleAnimeIDsAfterParams{
		AfterID: afterID,
		// batchSizeは小さく上限のあるページサイズ (既定1000) でint32上限には達しない。
		BatchSize: int32(batchSize), // #nosec G115
	})
	if err != nil {
		return AnimeIDReconcilePage{}, err
	}
	return newAnimeIDReconcilePage(row.LastID, row.UpdatedCount), nil
}
