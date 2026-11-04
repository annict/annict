package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// AnimeRedirectRepositoryは、重複を統合したanimeから統合先を引くanime_redirectsテーブルへのデータアクセスを担う
type AnimeRedirectRepository struct {
	queries *query.Queries
}

// NewAnimeRedirectRepositoryはAnimeRedirectRepositoryを生成する
func NewAnimeRedirectRepository(queries *query.Queries) *AnimeRedirectRepository {
	return &AnimeRedirectRepository{queries: queries}
}

// FindCanonicalAnimeIDは、統合したanimeの統合先のIDを返す。統合の記録が無い場合は (nil, nil) を返す。
// 統合を重ねても1回で統合先に届くよう、記録は常に最終の統合先を指している
func (r *AnimeRedirectRepository) FindCanonicalAnimeID(ctx context.Context, oldAnimeID model.AnimeID) (*model.AnimeID, error) {
	row, err := r.queries.GetAnimeRedirectByOldAnimeID(ctx, int64(oldAnimeID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	canonicalAnimeID := model.AnimeID(row.CanonicalAnimeID)
	return &canonicalAnimeID, nil
}

// ListCanonicalAnimeIDsは、統合したanimeのIDから統合先のIDへのmapを返す。
// 統合の記録が無いanimeは、mapに含めない。空入力ではクエリしない
func (r *AnimeRedirectRepository) ListCanonicalAnimeIDs(ctx context.Context, oldAnimeIDs []model.AnimeID) (map[model.AnimeID]model.AnimeID, error) {
	canonicalIDs := make(map[model.AnimeID]model.AnimeID, len(oldAnimeIDs))
	if len(oldAnimeIDs) == 0 {
		return canonicalIDs, nil
	}

	rawIDs := make([]int64, len(oldAnimeIDs))
	for i, id := range oldAnimeIDs {
		rawIDs[i] = int64(id)
	}
	rows, err := r.queries.ListAnimeRedirectsByOldAnimeIDs(ctx, rawIDs)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		canonicalIDs[model.AnimeID(row.OldAnimeID)] = model.AnimeID(row.CanonicalAnimeID)
	}
	return canonicalIDs, nil
}
