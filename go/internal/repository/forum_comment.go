package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// ForumCommentRepositoryはフォーラムのコメント関連のデータアクセスを担当します
type ForumCommentRepository struct {
	queries *query.Queries
}

// NewForumCommentRepositoryはForumCommentRepositoryを作成します
func NewForumCommentRepository(queries *query.Queries) *ForumCommentRepository {
	return &ForumCommentRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *ForumCommentRepository) WithTx(tx *sql.Tx) *ForumCommentRepository {
	return &ForumCommentRepository{queries: r.queries.WithTx(tx)}
}

// DeleteOnForumPostsByUserIDはユーザーが立てたスレッドに付いたコメントを最大limit件削除し、削除した件数を返します。
func (r *ForumCommentRepository) DeleteOnForumPostsByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteForumCommentsOnForumPostsByUserID(ctx, query.DeleteForumCommentsOnForumPostsByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}

// DetachFromUserはユーザーが他人のスレッドに書いたコメントを最大limit件、投稿者との紐づきを外し、
// 外した件数を返します。
func (r *ForumCommentRepository) DetachFromUser(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DetachForumCommentsFromUser(ctx, query.DetachForumCommentsFromUserParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
