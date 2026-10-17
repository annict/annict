package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// WorkCommentRepositoryは作品へのコメント関連のデータアクセスを担当します
type WorkCommentRepository struct {
	queries *query.Queries
}

// NewWorkCommentRepositoryはWorkCommentRepositoryを作成します
func NewWorkCommentRepository(queries *query.Queries) *WorkCommentRepository {
	return &WorkCommentRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *WorkCommentRepository) WithTx(tx *sql.Tx) *WorkCommentRepository {
	return &WorkCommentRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーが書いた作品へのコメントを最大limit件削除し、削除した件数を返します。
func (r *WorkCommentRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteWorkCommentsByUserID(ctx, query.DeleteWorkCommentsByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
