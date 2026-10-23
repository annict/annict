package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// ActivityGroupRepositoryはアクティビティグループ関連のデータアクセスを担当します
type ActivityGroupRepository struct {
	queries *query.Queries
}

// NewActivityGroupRepositoryはActivityGroupRepositoryを作成します
func NewActivityGroupRepository(queries *query.Queries) *ActivityGroupRepository {
	return &ActivityGroupRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *ActivityGroupRepository) WithTx(tx *sql.Tx) *ActivityGroupRepository {
	return &ActivityGroupRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーのアクティビティグループを最大limit件削除し、削除した件数を返します。
// グループに属するアクティビティは先に削除しておく必要があります。
func (r *ActivityGroupRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteActivityGroupsByUserID(ctx, query.DeleteActivityGroupsByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
