package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// OAuthAccessGrantRepositoryはOAuthの認可関連のデータアクセスを担当します
type OAuthAccessGrantRepository struct {
	queries *query.Queries
}

// NewOAuthAccessGrantRepositoryはOAuthAccessGrantRepositoryを作成します
func NewOAuthAccessGrantRepository(queries *query.Queries) *OAuthAccessGrantRepository {
	return &OAuthAccessGrantRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *OAuthAccessGrantRepository) WithTx(tx *sql.Tx) *OAuthAccessGrantRepository {
	return &OAuthAccessGrantRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーが許可した認可と、ユーザーが持つ論理削除済みのアプリケーションに残る認可を最大limit件削除し、削除した件数を返します。
func (r *OAuthAccessGrantRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteOAuthAccessGrantsByUserID(ctx, query.DeleteOAuthAccessGrantsByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
