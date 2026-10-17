package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// ActivityRepositoryはアクティビティ関連のデータアクセスを担当します
type ActivityRepository struct {
	queries *query.Queries
}

// NewActivityRepositoryはActivityRepositoryを作成します
func NewActivityRepository(queries *query.Queries) *ActivityRepository {
	return &ActivityRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *ActivityRepository) WithTx(tx *sql.Tx) *ActivityRepository {
	return &ActivityRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーのアクティビティを最大limit件削除し、削除した件数を返します。
func (r *ActivityRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteActivitiesByUserID(ctx, query.DeleteActivitiesByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
