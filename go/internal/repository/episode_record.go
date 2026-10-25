package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// EpisodeRecordRepositoryはエピソードの記録関連のデータアクセスを担当します
type EpisodeRecordRepository struct {
	queries *query.Queries
}

// NewEpisodeRecordRepositoryはEpisodeRecordRepositoryを作成します
func NewEpisodeRecordRepository(queries *query.Queries) *EpisodeRecordRepository {
	return &EpisodeRecordRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *EpisodeRecordRepository) WithTx(tx *sql.Tx) *EpisodeRecordRepository {
	return &EpisodeRecordRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーのエピソードの記録を最大limit件削除し、削除した件数を返します。
// 記録に付いたいいねとその通知も同じ文の中で削除し、論理削除されていない記録の分だけ
// エピソードの記録数と本文のある記録数を減算します。
func (r *EpisodeRecordRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteEpisodeRecordsByUserID(ctx, query.DeleteEpisodeRecordsByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}

// DetachDeletedOAuthApplicationsByOwnerIDは、ユーザーが持つ論理削除済みのアプリケーションから
// 投稿されたエピソードの記録を最大limit件、アプリケーションとの紐づきを外し、外した件数を返します。
func (r *EpisodeRecordRepository) DetachDeletedOAuthApplicationsByOwnerID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DetachDeletedOAuthApplicationsFromEpisodeRecords(ctx, query.DetachDeletedOAuthApplicationsFromEpisodeRecordsParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
