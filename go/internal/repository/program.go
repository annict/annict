package repository

import (
	"context"

	"github.com/annict/annict/go/internal/query"
)

// ProgramRepositoryは番組情報 (programs) のデータアクセスを担当する。
type ProgramRepository struct {
	queries *query.Queries
}

// NewProgramRepositoryはProgramRepositoryを作成する。
func NewProgramRepository(queries *query.Queries) *ProgramRepository {
	return &ProgramRepository{queries: queries}
}

// ReconcileAnimeIDsAfterは、IDがafterIDより大きい番組情報の1ページについて、anime_idを
// 作品のanime_idと照合して直す。作品を指していない行と、作品のanime_idがまだ無い行はNULLにする。
func (r *ProgramRepository) ReconcileAnimeIDsAfter(ctx context.Context, afterID int64, batchSize int) (AnimeIDReconcilePage, error) {
	row, err := r.queries.ReconcileProgramAnimeIDsAfter(ctx, query.ReconcileProgramAnimeIDsAfterParams{
		AfterID: afterID,
		// batchSizeは小さく上限のあるページサイズ (既定1000) でint32上限には達しない。
		BatchSize: int32(batchSize), // #nosec G115
	})
	if err != nil {
		return AnimeIDReconcilePage{}, err
	}
	return newAnimeIDReconcilePage(row.LastID, row.UpdatedCount), nil
}
