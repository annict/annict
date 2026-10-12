package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// WorkRecordRepositoryは作品の記録関連のデータアクセスを担当します
type WorkRecordRepository struct {
	queries *query.Queries
}

// NewWorkRecordRepositoryはWorkRecordRepositoryを作成します
func NewWorkRecordRepository(queries *query.Queries) *WorkRecordRepository {
	return &WorkRecordRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *WorkRecordRepository) WithTx(tx *sql.Tx) *WorkRecordRepository {
	return &WorkRecordRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーの作品の記録を最大limit件削除し、削除した件数を返します。
// 記録に付いたいいねとその通知も同じ文の中で削除し、論理削除されていない記録の分だけ
// 作品の記録数と本文のある記録数を減算します。
func (r *WorkRecordRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteWorkRecordsByUserID(ctx, query.DeleteWorkRecordsByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}

// DetachDeletedOAuthApplicationsByOwnerIDは、ユーザーが持つ論理削除済みのアプリケーションから
// 投稿された作品の記録を最大limit件、アプリケーションとの紐づきを外し、外した件数を返します。
func (r *WorkRecordRepository) DetachDeletedOAuthApplicationsByOwnerID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DetachDeletedOAuthApplicationsFromWorkRecords(ctx, query.DetachDeletedOAuthApplicationsFromWorkRecordsParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
