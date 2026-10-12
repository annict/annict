package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// ForumPostParticipantRepositoryはフォーラムのスレッドの参加者関連のデータアクセスを担当します
type ForumPostParticipantRepository struct {
	queries *query.Queries
}

// NewForumPostParticipantRepositoryはForumPostParticipantRepositoryを作成します
func NewForumPostParticipantRepository(queries *query.Queries) *ForumPostParticipantRepository {
	return &ForumPostParticipantRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *ForumPostParticipantRepository) WithTx(tx *sql.Tx) *ForumPostParticipantRepository {
	return &ForumPostParticipantRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーのスレッドへの参加と、ユーザーが立てたスレッドの参加者を最大limit件削除し、削除した件数を返します。
func (r *ForumPostParticipantRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteForumPostParticipantsByUserID(ctx, query.DeleteForumPostParticipantsByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
