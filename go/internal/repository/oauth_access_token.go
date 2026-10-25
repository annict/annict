package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// OAuthAccessTokenRepositoryはOAuthのアクセストークン関連のデータアクセスを担当します
type OAuthAccessTokenRepository struct {
	queries *query.Queries
}

// NewOAuthAccessTokenRepositoryはOAuthAccessTokenRepositoryを作成します
func NewOAuthAccessTokenRepository(queries *query.Queries) *OAuthAccessTokenRepository {
	return &OAuthAccessTokenRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *OAuthAccessTokenRepository) WithTx(tx *sql.Tx) *OAuthAccessTokenRepository {
	return &OAuthAccessTokenRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーに発行したアクセストークンと、ユーザーが持つ論理削除済みのアプリケーションに残るアクセストークンを最大limit件削除し、削除した件数を返します。
func (r *OAuthAccessTokenRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteOAuthAccessTokensByUserID(ctx, query.DeleteOAuthAccessTokensByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
