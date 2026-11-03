package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// UserTwoFactorRecoveryCodeRepositoryはuser_two_factor_recovery_codesテーブルへのデータアクセスを担う。
//
// コードはダイジェストのまま受け渡し、このリポジトリは平文のコードを扱わない。
// 呼び出し側が必要とするのは未使用の数と使えたかどうかだけのため、行をモデルとして返すメソッドは持たない。
type UserTwoFactorRecoveryCodeRepository struct {
	queries *query.Queries
}

// NewUserTwoFactorRecoveryCodeRepositoryはUserTwoFactorRecoveryCodeRepositoryを生成する。
func NewUserTwoFactorRecoveryCodeRepository(queries *query.Queries) *UserTwoFactorRecoveryCodeRepository {
	return &UserTwoFactorRecoveryCodeRepository{queries: queries}
}

// WithTxはクエリをtx内で実行する新しいUserTwoFactorRecoveryCodeRepositoryを返す。
func (r *UserTwoFactorRecoveryCodeRepository) WithTx(tx *sql.Tx) *UserTwoFactorRecoveryCodeRepository {
	return &UserTwoFactorRecoveryCodeRepository{queries: r.queries.WithTx(tx)}
}

// CreateAllはユーザーのリカバリーコードをダイジェストの数だけ作る。
// 1つの文で作るため、一部だけが保存されることは無い。
func (r *UserTwoFactorRecoveryCodeRepository) CreateAll(ctx context.Context, userID model.UserID, codeDigests []string) error {
	if len(codeDigests) == 0 {
		return nil
	}

	return r.queries.CreateUserTwoFactorRecoveryCodes(ctx, query.CreateUserTwoFactorRecoveryCodesParams{
		UserID:      int64(userID),
		CodeDigests: codeDigests,
	})
}

// CountUnusedはユーザーのまだ使っていないリカバリーコードの数を返す。
func (r *UserTwoFactorRecoveryCodeRepository) CountUnused(ctx context.Context, userID model.UserID) (int, error) {
	count, err := r.queries.CountUnusedUserTwoFactorRecoveryCodes(ctx, int64(userID))
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// Useはダイジェストが一致する未使用のコードを使用済みにする。
// 一致するコードが無い・使用済みのときは更新せずfalseを返す。
//
// 照合と使用を分けずに条件付きのUPDATEで判定するのは、同じコードを同時に送られても一方しか通さないため。
func (r *UserTwoFactorRecoveryCodeRepository) Use(ctx context.Context, userID model.UserID, codeDigest string) (bool, error) {
	affected, err := r.queries.UseUserTwoFactorRecoveryCode(ctx, query.UseUserTwoFactorRecoveryCodeParams{
		UserID:     int64(userID),
		CodeDigest: codeDigest,
	})
	if err != nil {
		return false, err
	}

	return affected > 0, nil
}

// DeleteByUserIDはユーザーのリカバリーコードをすべて削除する。
func (r *UserTwoFactorRecoveryCodeRepository) DeleteByUserID(ctx context.Context, userID model.UserID) error {
	return r.queries.DeleteUserTwoFactorRecoveryCodesByUserID(ctx, int64(userID))
}
