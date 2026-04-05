// Package repository はデータアクセス層を提供します
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/annict/annict/go/internal/password_reset"
	"github.com/annict/annict/go/internal/query"
)

// ErrInvalidPasswordResetToken はトークンが無効（存在しない、使用済み、期限切れ）であることを示すエラーです
var ErrInvalidPasswordResetToken = errors.New("invalid password reset token")

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

// GetValidByToken はトークン文字列からハッシュを計算し、有効なトークンを取得します。
// トークンが存在しない、使用済み、期限切れの場合は ErrInvalidPasswordResetToken を返します。
func (r *PasswordResetTokenRepository) GetValidByToken(ctx context.Context, token string) (query.PasswordResetToken, error) {
	tokenDigest := password_reset.HashToken(token)
	resetToken, err := r.queries.GetPasswordResetTokenByDigest(ctx, tokenDigest)
	if err == sql.ErrNoRows {
		slog.WarnContext(ctx, "無効なパスワードリセットトークン",
			"token_digest", tokenDigest,
			"reason", "not_found",
		)
		return query.PasswordResetToken{}, ErrInvalidPasswordResetToken
	} else if err != nil {
		return query.PasswordResetToken{}, fmt.Errorf("パスワードリセットトークンの取得に失敗: %w", err)
	}

	if resetToken.UsedAt.Valid {
		slog.WarnContext(ctx, "無効なパスワードリセットトークン",
			"token_digest", tokenDigest,
			"reason", "used",
		)
		return query.PasswordResetToken{}, ErrInvalidPasswordResetToken
	}

	if time.Now().After(resetToken.ExpiresAt) {
		slog.WarnContext(ctx, "無効なパスワードリセットトークン",
			"token_digest", tokenDigest,
			"reason", "expired",
		)
		return query.PasswordResetToken{}, ErrInvalidPasswordResetToken
	}

	return resetToken, nil
}

// DeleteExpired は指定日時より前に期限切れまたは使用済みになったトークンを削除します
func (r *PasswordResetTokenRepository) DeleteExpired(ctx context.Context, cutoff time.Time) error {
	return r.queries.DeleteExpiredPasswordResetTokens(ctx, cutoff)
}
