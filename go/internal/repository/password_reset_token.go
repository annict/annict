// Package repository はデータアクセス層を提供します
package repository

import (
	"context"

	"github.com/annict/annict/internal/query"
)

// PasswordResetTokenRepository はPasswordResetToken関連のデータアクセスを担当します
type PasswordResetTokenRepository struct {
	queries *query.Queries
}

// NewPasswordResetTokenRepository はPasswordResetTokenRepositoryを作成します
func NewPasswordResetTokenRepository(queries *query.Queries) *PasswordResetTokenRepository {
	return &PasswordResetTokenRepository{queries: queries}
}

// GetByDigest はトークンダイジェストでパスワードリセットトークンを検索します
func (r *PasswordResetTokenRepository) GetByDigest(ctx context.Context, tokenDigest string) (query.PasswordResetToken, error) {
	return r.queries.GetPasswordResetTokenByDigest(ctx, tokenDigest)
}
