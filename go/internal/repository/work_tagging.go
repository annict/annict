package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// WorkTaggingRepositoryは作品のタグ付け関連のデータアクセスを担当します
type WorkTaggingRepository struct {
	queries *query.Queries
}

// NewWorkTaggingRepositoryはWorkTaggingRepositoryを作成します
func NewWorkTaggingRepository(queries *query.Queries) *WorkTaggingRepository {
	return &WorkTaggingRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *WorkTaggingRepository) WithTx(tx *sql.Tx) *WorkTaggingRepository {
	return &WorkTaggingRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーが作品に付けたタグを最大limit件削除し、削除した件数を返します。
// タグの作品数を削除した件数だけ減算します。
func (r *WorkTaggingRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteWorkTaggingsByUserID(ctx, query.DeleteWorkTaggingsByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
