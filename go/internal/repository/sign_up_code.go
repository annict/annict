package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/annict/annict/go/internal/query"
)

// SignUpCode はサインアップコードの型エイリアスです
type SignUpCode = query.SignUpCode

// SignUpCodeRepository はサインアップコード関連のデータアクセスを担当します
type SignUpCodeRepository struct {
	queries *query.Queries
}

// NewSignUpCodeRepository はSignUpCodeRepositoryを作成します
func NewSignUpCodeRepository(queries *query.Queries) *SignUpCodeRepository {
	return &SignUpCodeRepository{queries: queries}
}

// WithTx はトランザクションを使用する新しいRepositoryを返します
func (r *SignUpCodeRepository) WithTx(tx *sql.Tx) *SignUpCodeRepository {
	return &SignUpCodeRepository{queries: r.queries.WithTx(tx)}
}

// SignUpCodeCreateParams はサインアップコード作成のパラメータ
type SignUpCodeCreateParams struct {
	Email      string
	CodeDigest string
	ExpiresAt  time.Time
}

// Create は新しいサインアップコードを作成します
func (r *SignUpCodeRepository) Create(ctx context.Context, params SignUpCodeCreateParams) error {
	_, err := r.queries.CreateSignUpCode(ctx, query.CreateSignUpCodeParams{
		Email:      params.Email,
		CodeDigest: params.CodeDigest,
		ExpiresAt:  params.ExpiresAt,
	})
	return err
}

// InvalidateByEmail はメールアドレスに関連する既存の未使用コードを無効化します
func (r *SignUpCodeRepository) InvalidateByEmail(ctx context.Context, email string) error {
	return r.queries.InvalidateSignUpCodesByEmail(ctx, email)
}

// GetValidByEmail はメールアドレスの有効なサインアップコードを取得します
func (r *SignUpCodeRepository) GetValidByEmail(ctx context.Context, email string) (query.SignUpCode, error) {
	return r.queries.GetValidSignUpCode(ctx, email)
}

// MarkAsUsed はサインアップコードを使用済みにマークします
func (r *SignUpCodeRepository) MarkAsUsed(ctx context.Context, id int64) error {
	return r.queries.MarkSignUpCodeAsUsed(ctx, id)
}

// IncrementAttempts はサインアップコードの試行回数をインクリメントします
func (r *SignUpCodeRepository) IncrementAttempts(ctx context.Context, id int64) error {
	return r.queries.IncrementSignUpCodeAttempts(ctx, id)
}
