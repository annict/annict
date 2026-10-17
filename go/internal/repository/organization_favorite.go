package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// OrganizationFavoriteRepositoryは団体のお気に入り関連のデータアクセスを担当します
type OrganizationFavoriteRepository struct {
	queries *query.Queries
}

// NewOrganizationFavoriteRepositoryはOrganizationFavoriteRepositoryを作成します
func NewOrganizationFavoriteRepository(queries *query.Queries) *OrganizationFavoriteRepository {
	return &OrganizationFavoriteRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *OrganizationFavoriteRepository) WithTx(tx *sql.Tx) *OrganizationFavoriteRepository {
	return &OrganizationFavoriteRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーの団体のお気に入りを最大limit件削除し、削除した件数を返します。
// 団体のお気に入り数を削除した件数だけ減算します。
func (r *OrganizationFavoriteRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteOrganizationFavoritesByUserID(ctx, query.DeleteOrganizationFavoritesByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
