package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// UserlandProjectMemberRepositoryはUserlandプロジェクトの参加者関連のデータアクセスを担当します
type UserlandProjectMemberRepository struct {
	queries *query.Queries
}

// NewUserlandProjectMemberRepositoryはUserlandProjectMemberRepositoryを作成します
func NewUserlandProjectMemberRepository(queries *query.Queries) *UserlandProjectMemberRepository {
	return &UserlandProjectMemberRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *UserlandProjectMemberRepository) WithTx(tx *sql.Tx) *UserlandProjectMemberRepository {
	return &UserlandProjectMemberRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーのUserlandプロジェクトへの参加を最大limit件削除し、削除した件数を返します。
func (r *UserlandProjectMemberRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteUserlandProjectMembersByUserID(ctx, query.DeleteUserlandProjectMembersByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
