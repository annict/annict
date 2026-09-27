package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// SignInCodeRepositoryはサインインコード関連のデータアクセスを担当します
type SignInCodeRepository struct {
	queries *query.Queries
}

// NewSignInCodeRepositoryはSignInCodeRepositoryを作成します
func NewSignInCodeRepository(queries *query.Queries) *SignInCodeRepository {
	return &SignInCodeRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *SignInCodeRepository) WithTx(tx *sql.Tx) *SignInCodeRepository {
	return &SignInCodeRepository{queries: r.queries.WithTx(tx)}
}

// SignInCodeCreateParamsはサインインコード作成のパラメータ
type SignInCodeCreateParams struct {
	UserID     model.UserID
	CodeDigest string
	ExpiresAt  time.Time
}

// Createは新しいサインインコードを作成します
func (r *SignInCodeRepository) Create(ctx context.Context, params SignInCodeCreateParams) (*model.SignInCode, error) {
	row, err := r.queries.CreateSignInCode(ctx, query.CreateSignInCodeParams{
		UserID:     int64(params.UserID),
		CodeDigest: params.CodeDigest,
		ExpiresAt:  params.ExpiresAt,
	})
	if err != nil {
		return nil, err
	}
	return signInCodeFromRow(row), nil
}

// InvalidateByUserIDはユーザーの既存の未使用コードを無効化します
func (r *SignInCodeRepository) InvalidateByUserID(ctx context.Context, userID model.UserID) error {
	return r.queries.InvalidateUserSignInCodes(ctx, int64(userID))
}

// GetValidByUserIDはユーザーの有効なサインインコードを取得します
func (r *SignInCodeRepository) GetValidByUserID(ctx context.Context, userID model.UserID) (*model.SignInCode, error) {
	row, err := r.queries.GetValidSignInCode(ctx, int64(userID))
	if err != nil {
		return nil, err
	}
	return signInCodeFromRow(row), nil
}

// MarkAsUsedはサインインコードを使用済みにマークします
func (r *SignInCodeRepository) MarkAsUsed(ctx context.Context, id model.SignInCodeID) error {
	return r.queries.MarkSignInCodeAsUsed(ctx, int64(id))
}

// IncrementAttemptsはサインインコードの試行回数をインクリメントします
func (r *SignInCodeRepository) IncrementAttempts(ctx context.Context, id model.SignInCodeID) error {
	return r.queries.IncrementSignInCodeAttempts(ctx, int64(id))
}

// DeleteExpiredは指定日時より前に期限切れまたは使用済みになったコードを削除します
func (r *SignInCodeRepository) DeleteExpired(ctx context.Context, cutoff time.Time) error {
	return r.queries.DeleteExpiredSignInCodes(ctx, cutoff)
}

func signInCodeFromRow(row query.SignInCode) *model.SignInCode {
	return &model.SignInCode{
		ID:         model.SignInCodeID(row.ID),
		UserID:     model.UserID(row.UserID),
		CodeDigest: row.CodeDigest,
		Attempts:   row.Attempts,
		ExpiresAt:  row.ExpiresAt,
		UsedAt:     row.UsedAt,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}
