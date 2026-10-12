package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// CollectionItemRepositoryはコレクション項目関連のデータアクセスを担当します
type CollectionItemRepository struct {
	queries *query.Queries
}

// NewCollectionItemRepositoryはCollectionItemRepositoryを作成します
func NewCollectionItemRepository(queries *query.Queries) *CollectionItemRepository {
	return &CollectionItemRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *CollectionItemRepository) WithTx(tx *sql.Tx) *CollectionItemRepository {
	return &CollectionItemRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーのコレクション項目と、ユーザーのコレクションに入っている項目を最大limit件削除し、削除した件数を返します。
func (r *CollectionItemRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteCollectionItemsByUserID(ctx, query.DeleteCollectionItemsByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
