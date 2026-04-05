package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/annict/annict/go/internal/query"
)

// SignInCode はサインインコードの型エイリアスです
type SignInCode = query.SignInCode

// SignInCodeRepository はサインインコード関連のデータアクセスを担当します
type SignInCodeRepository struct {
	queries *query.Queries
}

// NewSignInCodeRepository はSignInCodeRepositoryを作成します
func NewSignInCodeRepository(queries *query.Queries) *SignInCodeRepository {
	return &SignInCodeRepository{queries: queries}
}

// WithTx はトランザクションを使用する新しいRepositoryを返します
func (r *SignInCodeRepository) WithTx(tx *sql.Tx) *SignInCodeRepository {
	return &SignInCodeRepository{queries: r.queries.WithTx(tx)}
}

// SignInCodeCreateParams はサインインコード作成のパラメータ
type SignInCodeCreateParams struct {
	UserID     int64
	CodeDigest string
	ExpiresAt  time.Time
}

// Create は新しいサインインコードを作成します
func (r *SignInCodeRepository) Create(ctx context.Context, params SignInCodeCreateParams) error {
	_, err := r.queries.CreateSignInCode(ctx, query.CreateSignInCodeParams{
		UserID:     params.UserID,
		CodeDigest: params.CodeDigest,
		ExpiresAt:  params.ExpiresAt,
	})
	return err
}

// InvalidateByUserID はユーザーの既存の未使用コードを無効化します
func (r *SignInCodeRepository) InvalidateByUserID(ctx context.Context, userID int64) error {
	return r.queries.InvalidateUserSignInCodes(ctx, userID)
}

// GetValidByUserID はユーザーの有効なサインインコードを取得します
func (r *SignInCodeRepository) GetValidByUserID(ctx context.Context, userID int64) (query.SignInCode, error) {
	return r.queries.GetValidSignInCode(ctx, userID)
}

// MarkAsUsed はサインインコードを使用済みにマークします
func (r *SignInCodeRepository) MarkAsUsed(ctx context.Context, id int64) error {
	return r.queries.MarkSignInCodeAsUsed(ctx, id)
}

// IncrementAttempts はサインインコードの試行回数をインクリメントします
func (r *SignInCodeRepository) IncrementAttempts(ctx context.Context, id int64) error {
	return r.queries.IncrementSignInCodeAttempts(ctx, id)
}

// DeleteExpired は指定日時より前に期限切れまたは使用済みになったコードを削除します
func (r *SignInCodeRepository) DeleteExpired(ctx context.Context, cutoff time.Time) error {
	return r.queries.DeleteExpiredSignInCodes(ctx, cutoff)
}
