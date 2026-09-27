package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// SignUpCodeRepositoryはサインアップコード関連のデータアクセスを担当します
type SignUpCodeRepository struct {
	queries *query.Queries
}

// NewSignUpCodeRepositoryはSignUpCodeRepositoryを作成します
func NewSignUpCodeRepository(queries *query.Queries) *SignUpCodeRepository {
	return &SignUpCodeRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *SignUpCodeRepository) WithTx(tx *sql.Tx) *SignUpCodeRepository {
	return &SignUpCodeRepository{queries: r.queries.WithTx(tx)}
}

// SignUpCodeCreateParamsはサインアップコード作成のパラメータ
type SignUpCodeCreateParams struct {
	Email      string
	CodeDigest string
	ExpiresAt  time.Time
}

// Createは新しいサインアップコードを作成します
func (r *SignUpCodeRepository) Create(ctx context.Context, params SignUpCodeCreateParams) (*model.SignUpCode, error) {
	row, err := r.queries.CreateSignUpCode(ctx, query.CreateSignUpCodeParams{
		Email:      params.Email,
		CodeDigest: params.CodeDigest,
		ExpiresAt:  params.ExpiresAt,
	})
	if err != nil {
		return nil, err
	}
	return signUpCodeFromRow(row), nil
}

// InvalidateByEmailはメールアドレスに関連する既存の未使用コードを無効化します
func (r *SignUpCodeRepository) InvalidateByEmail(ctx context.Context, email string) error {
	return r.queries.InvalidateSignUpCodesByEmail(ctx, email)
}

// GetValidByEmailはメールアドレスの有効なサインアップコードを取得します
func (r *SignUpCodeRepository) GetValidByEmail(ctx context.Context, email string) (*model.SignUpCode, error) {
	row, err := r.queries.GetValidSignUpCode(ctx, email)
	if err != nil {
		return nil, err
	}
	return signUpCodeFromRow(row), nil
}

// MarkAsUsedはサインアップコードを使用済みにマークします
func (r *SignUpCodeRepository) MarkAsUsed(ctx context.Context, id model.SignUpCodeID) error {
	return r.queries.MarkSignUpCodeAsUsed(ctx, int64(id))
}

// IncrementAttemptsはサインアップコードの試行回数をインクリメントします
func (r *SignUpCodeRepository) IncrementAttempts(ctx context.Context, id model.SignUpCodeID) error {
	return r.queries.IncrementSignUpCodeAttempts(ctx, int64(id))
}

func signUpCodeFromRow(row query.SignUpCode) *model.SignUpCode {
	return &model.SignUpCode{
		ID:         model.SignUpCodeID(row.ID),
		Email:      row.Email,
		CodeDigest: row.CodeDigest,
		Attempts:   row.Attempts,
		UsedAt:     row.UsedAt,
		ExpiresAt:  row.ExpiresAt,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}
