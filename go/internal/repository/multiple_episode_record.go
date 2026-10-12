package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// MultipleEpisodeRecordRepositoryは複数エピソードの記録関連のデータアクセスを担当します
type MultipleEpisodeRecordRepository struct {
	queries *query.Queries
}

// NewMultipleEpisodeRecordRepositoryはMultipleEpisodeRecordRepositoryを作成します
func NewMultipleEpisodeRecordRepository(queries *query.Queries) *MultipleEpisodeRecordRepository {
	return &MultipleEpisodeRecordRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *MultipleEpisodeRecordRepository) WithTx(tx *sql.Tx) *MultipleEpisodeRecordRepository {
	return &MultipleEpisodeRecordRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーの複数エピソードの記録を最大limit件削除し、削除した件数を返します。
// 記録に付いたいいねとその通知も同じ文の中で削除します。
func (r *MultipleEpisodeRecordRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteMultipleEpisodeRecordsByUserID(ctx, query.DeleteMultipleEpisodeRecordsByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
