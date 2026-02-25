// Package repository はデータアクセス層を提供します
package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/query"
)

// ProfileRepository はProfile関連のデータアクセスを担当します
type ProfileRepository struct {
	queries *query.Queries
}

// NewProfileRepository はProfileRepositoryを作成します
func NewProfileRepository(queries *query.Queries) *ProfileRepository {
	return &ProfileRepository{queries: queries}
}

// WithTx はトランザクションを使用する新しいRepositoryを返します
func (r *ProfileRepository) WithTx(tx *sql.Tx) *ProfileRepository {
	return &ProfileRepository{queries: r.queries.WithTx(tx)}
}

// Create はプロフィールを作成します
func (r *ProfileRepository) Create(ctx context.Context, userID int64, name string) error {
	_, err := r.queries.CreateProfile(ctx, query.CreateProfileParams{
		UserID: userID,
		Name:   name,
	})
	return err
}
