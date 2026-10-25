package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// MuteUserRepositoryはミュート関連のデータアクセスを担当します
type MuteUserRepository struct {
	queries *query.Queries
}

// NewMuteUserRepositoryはMuteUserRepositoryを作成します
func NewMuteUserRepository(queries *query.Queries) *MuteUserRepository {
	return &MuteUserRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *MuteUserRepository) WithTx(tx *sql.Tx) *MuteUserRepository {
	return &MuteUserRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーがしたミュートを最大limit件削除し、削除した件数を返します。
func (r *MuteUserRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteMuteUsersByUserID(ctx, query.DeleteMuteUsersByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}

// DeleteByMutedUserIDは他のユーザーがユーザーをミュートした行を最大limit件削除し、削除した件数を返します。
func (r *MuteUserRepository) DeleteByMutedUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteMuteUsersByMutedUserID(ctx, query.DeleteMuteUsersByMutedUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
