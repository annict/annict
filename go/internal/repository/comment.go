package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// CommentRepositoryはコメント関連のデータアクセスを担当します
type CommentRepository struct {
	queries *query.Queries
}

// NewCommentRepositoryはCommentRepositoryを作成します
func NewCommentRepository(queries *query.Queries) *CommentRepository {
	return &CommentRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *CommentRepository) WithTx(tx *sql.Tx) *CommentRepository {
	return &CommentRepository{queries: r.queries.WithTx(tx)}
}

// DeleteOnEpisodeRecordsByUserIDはユーザーのエピソードの記録に付いたコメントを、書いたのが誰かを問わず
// 最大limit件削除し、削除した件数を返します。コメントの通知と、コメントに付いたいいね・その通知も
// 同じ文の中で削除し、他人宛ての未読の通知の分だけ宛先の未読数を減算します。
func (r *CommentRepository) DeleteOnEpisodeRecordsByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteCommentsOnEpisodeRecordsByUserID(ctx, query.DeleteCommentsOnEpisodeRecordsByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}

// DeleteByUserIDはユーザーが書いたコメントを最大limit件削除し、削除した件数を返します。
// コメントに付いたいいねも同じ文の中で削除し、コメント先の記録のコメント数を減算します。
// コメントの通知とコメントへのいいねの通知は先に削除しておく必要があります。
func (r *CommentRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteCommentsByUserID(ctx, query.DeleteCommentsByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
