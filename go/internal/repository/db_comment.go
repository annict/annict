package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// DBCommentRepositoryはAnnict DBのコメント関連のデータアクセスを担当します
type DBCommentRepository struct {
	queries *query.Queries
}

// NewDBCommentRepositoryはDBCommentRepositoryを作成します
func NewDBCommentRepository(queries *query.Queries) *DBCommentRepository {
	return &DBCommentRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *DBCommentRepository) WithTx(tx *sql.Tx) *DBCommentRepository {
	return &DBCommentRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーが書いたAnnict DBのコメントを最大limit件削除し、削除した件数を返します。
func (r *DBCommentRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteDbCommentsByUserID(ctx, query.DeleteDbCommentsByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
