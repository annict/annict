package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// CharacterFavoriteRepositoryはキャラクターのお気に入り関連のデータアクセスを担当します
type CharacterFavoriteRepository struct {
	queries *query.Queries
}

// NewCharacterFavoriteRepositoryはCharacterFavoriteRepositoryを作成します
func NewCharacterFavoriteRepository(queries *query.Queries) *CharacterFavoriteRepository {
	return &CharacterFavoriteRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *CharacterFavoriteRepository) WithTx(tx *sql.Tx) *CharacterFavoriteRepository {
	return &CharacterFavoriteRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByUserIDはユーザーのキャラクターのお気に入りを最大limit件削除し、削除した件数を返します。
// キャラクターのお気に入り数を削除した件数だけ減算します。
func (r *CharacterFavoriteRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteCharacterFavoritesByUserID(ctx, query.DeleteCharacterFavoritesByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
