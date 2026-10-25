package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// PersonFavoriteRepositoryは人物のお気に入り関連のデータアクセスを担当します
type PersonFavoriteRepository struct {
	queries *query.Queries
}

// NewPersonFavoriteRepositoryはPersonFavoriteRepositoryを作成します
func NewPersonFavoriteRepository(queries *query.Queries) *PersonFavoriteRepository {
	return &PersonFavoriteRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *PersonFavoriteRepository) WithTx(tx *sql.Tx) *PersonFavoriteRepository {
	return &PersonFavoriteRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーの人物のお気に入りを最大limit件削除し、削除した件数を返します。
// 人物のお気に入り数を削除した件数だけ減算します。
func (r *PersonFavoriteRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeletePersonFavoritesByUserID(ctx, query.DeletePersonFavoritesByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
