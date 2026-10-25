package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// FollowRepositoryはフォロー関連のデータアクセスを担当します
type FollowRepository struct {
	queries *query.Queries
}

// NewFollowRepositoryはFollowRepositoryを作成します
func NewFollowRepository(queries *query.Queries) *FollowRepository {
	return &FollowRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *FollowRepository) WithTx(tx *sql.Tx) *FollowRepository {
	return &FollowRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーがしたフォローを最大limit件削除し、削除した件数を返します。
// フォロー先のフォロワー数を削除した件数だけ減算します。
func (r *FollowRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteFollowsByUserID(ctx, query.DeleteFollowsByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}

// DeleteByFollowingIDはユーザーへのフォローを最大limit件削除し、削除した件数を返します。
// フォローしていたユーザーのフォロー数を削除した件数だけ減算します。
func (r *FollowRepository) DeleteByFollowingID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteFollowsByFollowingID(ctx, query.DeleteFollowsByFollowingIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
