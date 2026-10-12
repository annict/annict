package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// DbCommentRepositoryはAnnict DBのコメント関連のデータアクセスを担当します
type DbCommentRepository struct {
	queries *query.Queries
}

// NewDbCommentRepositoryはDbCommentRepositoryを作成します
func NewDbCommentRepository(queries *query.Queries) *DbCommentRepository {
	return &DbCommentRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *DbCommentRepository) WithTx(tx *sql.Tx) *DbCommentRepository {
	return &DbCommentRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーが書いたAnnict DBのコメントを最大limit件削除し、削除した件数を返します。
func (r *DbCommentRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteDbCommentsByUserID(ctx, query.DeleteDbCommentsByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
