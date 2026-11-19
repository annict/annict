package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// DBActivityRepositoryはAnnict DBの編集履歴関連のデータアクセスを担当します
type DBActivityRepository struct {
	queries *query.Queries
}

// NewDBActivityRepositoryはDBActivityRepositoryを作成します
func NewDBActivityRepository(queries *query.Queries) *DBActivityRepository {
	return &DBActivityRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *DBActivityRepository) WithTx(tx *sql.Tx) *DBActivityRepository {
	return &DBActivityRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーが行ったAnnict DBの編集履歴を最大limit件削除し、削除した件数を返します。
func (r *DBActivityRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteDbActivitiesByUserID(ctx, query.DeleteDbActivitiesByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
