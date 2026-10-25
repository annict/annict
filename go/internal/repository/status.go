package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// StatusRepositoryはステータス関連のデータアクセスを担当します
type StatusRepository struct {
	queries *query.Queries
}

// NewStatusRepositoryはStatusRepositoryを作成します
func NewStatusRepository(queries *query.Queries) *StatusRepository {
	return &StatusRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *StatusRepository) WithTx(tx *sql.Tx) *StatusRepository {
	return &StatusRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーのステータスを最大limit件削除し、削除した件数を返します。
// ステータスに付いたいいねと、そのいいねでユーザーに届いた通知も同じ文の中で削除します。
// ステータスを参照するアクティビティとライブラリエントリは先に削除しておく必要があります。
func (r *StatusRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteStatusesByUserID(ctx, query.DeleteStatusesByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}

// DetachDeletedOAuthApplicationsByOwnerIDは、ユーザーが持つ論理削除済みのアプリケーションから
// 投稿されたステータスを最大limit件、アプリケーションとの紐づきを外し、外した件数を返します。
func (r *StatusRepository) DetachDeletedOAuthApplicationsByOwnerID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DetachDeletedOAuthApplicationsFromStatuses(ctx, query.DetachDeletedOAuthApplicationsFromStatusesParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
