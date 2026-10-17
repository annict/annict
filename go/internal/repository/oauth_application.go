package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// OAuthApplicationRepositoryはOAuthアプリケーション関連のデータアクセスを担当します
type OAuthApplicationRepository struct {
	queries *query.Queries
}

// NewOAuthApplicationRepositoryはOAuthApplicationRepositoryを作成します
func NewOAuthApplicationRepository(queries *query.Queries) *OAuthApplicationRepository {
	return &OAuthApplicationRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *OAuthApplicationRepository) WithTx(tx *sql.Tx) *OAuthApplicationRepository {
	return &OAuthApplicationRepository{queries: r.queries.WithTx(tx)}
}

// DeleteDeletedByOwnerIDはユーザーが持つ論理削除済みのアプリケーションを最大limit件削除し、削除した件数を返します。
func (r *OAuthApplicationRepository) DeleteDeletedByOwnerID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteDeletedOAuthApplicationsByOwnerID(ctx, query.DeleteDeletedOAuthApplicationsByOwnerIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}

// ExistsKeptByOwnerIDはユーザーが持つ論理削除されていないアプリケーションがあるかを返します。
func (r *OAuthApplicationRepository) ExistsKeptByOwnerID(ctx context.Context, userID model.UserID) (bool, error) {
	return r.queries.ExistsKeptOAuthApplicationByOwnerID(ctx, int64(userID))
}
