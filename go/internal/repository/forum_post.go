package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// ForumPostRepositoryはフォーラムのスレッド関連のデータアクセスを担当します
type ForumPostRepository struct {
	queries *query.Queries
}

// NewForumPostRepositoryはForumPostRepositoryを作成します
func NewForumPostRepository(queries *query.Queries) *ForumPostRepository {
	return &ForumPostRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *ForumPostRepository) WithTx(tx *sql.Tx) *ForumPostRepository {
	return &ForumPostRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーが立てたスレッドを最大limit件削除し、削除した件数を返します。
func (r *ForumPostRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteForumPostsByUserID(ctx, query.DeleteForumPostsByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
