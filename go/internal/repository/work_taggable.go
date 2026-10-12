package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// WorkTaggableRepositoryはユーザーが使ったタグ関連のデータアクセスを担当します
type WorkTaggableRepository struct {
	queries *query.Queries
}

// NewWorkTaggableRepositoryはWorkTaggableRepositoryを作成します
func NewWorkTaggableRepository(queries *query.Queries) *WorkTaggableRepository {
	return &WorkTaggableRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *WorkTaggableRepository) WithTx(tx *sql.Tx) *WorkTaggableRepository {
	return &WorkTaggableRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーが使ったタグの記録を最大limit件削除し、削除した件数を返します。
func (r *WorkTaggableRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteWorkTaggablesByUserID(ctx, query.DeleteWorkTaggablesByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
