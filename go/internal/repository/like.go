package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// LikeRepositoryはいいね関連のデータアクセスを担当します
type LikeRepository struct {
	queries *query.Queries
}

// NewLikeRepositoryはLikeRepositoryを作成します
func NewLikeRepository(queries *query.Queries) *LikeRepository {
	return &LikeRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *LikeRepository) WithTx(tx *sql.Tx) *LikeRepository {
	return &LikeRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーが付けたいいねを最大limit件削除し、削除した件数を返します。
// いいね先のいいね数を削除した件数だけ減算します。いいねの通知は先に削除しておく必要があります。
func (r *LikeRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteLikesByUserID(ctx, query.DeleteLikesByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
