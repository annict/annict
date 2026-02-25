// Package repository はデータアクセス層を提供します
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/annict/annict/go/internal/query"
)

// PasswordResetToken はパスワードリセットトークンの型エイリアスです
type PasswordResetToken = query.PasswordResetToken

// PasswordResetTokenRepository はPasswordResetToken関連のデータアクセスを担当します
type PasswordResetTokenRepository struct {
	queries *query.Queries
}

// NewPasswordResetTokenRepository はPasswordResetTokenRepositoryを作成します
func NewPasswordResetTokenRepository(queries *query.Queries) *PasswordResetTokenRepository {
	return &PasswordResetTokenRepository{queries: queries}
}

// WithTx はトランザクションを使用する新しいRepositoryを返します
func (r *PasswordResetTokenRepository) WithTx(tx *sql.Tx) *PasswordResetTokenRepository {
	return &PasswordResetTokenRepository{queries: r.queries.WithTx(tx)}
}

// GetByDigest はトークンダイジェストでパスワードリセットトークンを検索します
func (r *PasswordResetTokenRepository) GetByDigest(ctx context.Context, tokenDigest string) (query.PasswordResetToken, error) {
	return r.queries.GetPasswordResetTokenByDigest(ctx, tokenDigest)
}

// Create はパスワードリセットトークンを作成します
func (r *PasswordResetTokenRepository) Create(ctx context.Context, userID int64, tokenDigest string, expiresAt time.Time) (query.PasswordResetToken, error) {
	return r.queries.CreatePasswordResetToken(ctx, query.CreatePasswordResetTokenParams{
		UserID:      userID,
		TokenDigest: tokenDigest,
		ExpiresAt:   expiresAt,
	})
}

// DeleteUnusedByUserID は指定ユーザーの未使用トークンを削除します
func (r *PasswordResetTokenRepository) DeleteUnusedByUserID(ctx context.Context, userID int64) error {
	return r.queries.DeleteUnusedPasswordResetTokensByUserID(ctx, userID)
}

// MarkAsUsed はトークンを使用済みにマークします
func (r *PasswordResetTokenRepository) MarkAsUsed(ctx context.Context, id int64) error {
	return r.queries.MarkPasswordResetTokenAsUsed(ctx, id)
}
