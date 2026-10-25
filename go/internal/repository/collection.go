package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// CollectionRepositoryはコレクション関連のデータアクセスを担当します
type CollectionRepository struct {
	queries *query.Queries
}

// NewCollectionRepositoryはCollectionRepositoryを作成します
func NewCollectionRepository(queries *query.Queries) *CollectionRepository {
	return &CollectionRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *CollectionRepository) WithTx(tx *sql.Tx) *CollectionRepository {
	return &CollectionRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーのコレクションを、付いたいいねと一緒に最大limit件削除し、削除した件数を返します。
func (r *CollectionRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteCollectionsByUserID(ctx, query.DeleteCollectionsByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
