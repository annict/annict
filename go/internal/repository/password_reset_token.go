// Package repository はデータアクセス層を提供します
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/password_reset"
	"github.com/annict/annict/go/internal/query"
)

// ErrInvalidPasswordResetToken はトークンが無効（存在しない、使用済み、期限切れ）であることを示すエラーです
var ErrInvalidPasswordResetToken = errors.New("invalid password reset token")

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
func (r *PasswordResetTokenRepository) GetByDigest(ctx context.Context, tokenDigest string) (*model.PasswordResetToken, error) {
	row, err := r.queries.GetPasswordResetTokenByDigest(ctx, tokenDigest)
	if err != nil {
		return nil, err
	}
	return passwordResetTokenFromRow(row), nil
}

// GetByUserID は指定ユーザーのパスワードリセットトークンをすべて取得します（使用済み・期限切れも含む）
func (r *PasswordResetTokenRepository) GetByUserID(ctx context.Context, userID model.UserID) ([]*model.PasswordResetToken, error) {
	rows, err := r.queries.GetPasswordResetTokensByUserID(ctx, int64(userID))
	if err != nil {
		return nil, err
	}

	tokens := make([]*model.PasswordResetToken, len(rows))
	for i, row := range rows {
		tokens[i] = passwordResetTokenFromRow(row)
	}
	return tokens, nil
}

// Create はパスワードリセットトークンを作成します
func (r *PasswordResetTokenRepository) Create(ctx context.Context, userID model.UserID, tokenDigest string, expiresAt time.Time) (*model.PasswordResetToken, error) {
	row, err := r.queries.CreatePasswordResetToken(ctx, query.CreatePasswordResetTokenParams{
		UserID:      int64(userID),
		TokenDigest: tokenDigest,
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		return nil, err
	}
	return passwordResetTokenFromRow(row), nil
}

// DeleteUnusedByUserID は指定ユーザーの未使用トークンを削除します
func (r *PasswordResetTokenRepository) DeleteUnusedByUserID(ctx context.Context, userID model.UserID) error {
	return r.queries.DeleteUnusedPasswordResetTokensByUserID(ctx, int64(userID))
}

// MarkAsUsed はトークンを使用済みにマークします
func (r *PasswordResetTokenRepository) MarkAsUsed(ctx context.Context, id model.PasswordResetTokenID) error {
	return r.queries.MarkPasswordResetTokenAsUsed(ctx, int64(id))
}

// GetValidByToken はトークン文字列からハッシュを計算し、有効なトークンを取得します。
// トークンが存在しない、使用済み、期限切れの場合は ErrInvalidPasswordResetToken を返します。
func (r *PasswordResetTokenRepository) GetValidByToken(ctx context.Context, token string) (*model.PasswordResetToken, error) {
	tokenDigest := password_reset.HashToken(token)
	row, err := r.queries.GetPasswordResetTokenByDigest(ctx, tokenDigest)
	if err == sql.ErrNoRows {
		slog.WarnContext(ctx, "無効なパスワードリセットトークン",
			"token_digest", tokenDigest,
			"reason", "not_found",
		)
		return nil, ErrInvalidPasswordResetToken
	} else if err != nil {
		return nil, fmt.Errorf("パスワードリセットトークンの取得に失敗: %w", err)
	}

	if row.UsedAt.Valid {
		slog.WarnContext(ctx, "無効なパスワードリセットトークン",
			"token_digest", tokenDigest,
			"reason", "used",
		)
		return nil, ErrInvalidPasswordResetToken
	}

	if time.Now().After(row.ExpiresAt) {
		slog.WarnContext(ctx, "無効なパスワードリセットトークン",
			"token_digest", tokenDigest,
			"reason", "expired",
		)
		return nil, ErrInvalidPasswordResetToken
	}

	return passwordResetTokenFromRow(row), nil
}

// DeleteExpired は指定日時より前に期限切れまたは使用済みになったトークンを削除します
func (r *PasswordResetTokenRepository) DeleteExpired(ctx context.Context, cutoff time.Time) error {
	return r.queries.DeleteExpiredPasswordResetTokens(ctx, cutoff)
}

func passwordResetTokenFromRow(row query.PasswordResetToken) *model.PasswordResetToken {
	return &model.PasswordResetToken{
		ID:          model.PasswordResetTokenID(row.ID),
		UserID:      model.UserID(row.UserID),
		TokenDigest: row.TokenDigest,
		ExpiresAt:   row.ExpiresAt,
		UsedAt:      row.UsedAt,
		CreatedAt:   row.CreatedAt,
	}
}
