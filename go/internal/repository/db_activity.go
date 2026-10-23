package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// DbActivityRepositoryはAnnict DBの編集履歴関連のデータアクセスを担当します
type DbActivityRepository struct {
	queries *query.Queries
}

// NewDbActivityRepositoryはDbActivityRepositoryを作成します
func NewDbActivityRepository(queries *query.Queries) *DbActivityRepository {
	return &DbActivityRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *DbActivityRepository) WithTx(tx *sql.Tx) *DbActivityRepository {
	return &DbActivityRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーが行ったAnnict DBの編集履歴を最大limit件削除し、削除した件数を返します。
func (r *DbActivityRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteDbActivitiesByUserID(ctx, query.DeleteDbActivitiesByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
