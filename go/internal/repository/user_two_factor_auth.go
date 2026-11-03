package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// UserTwoFactorAuthRepositoryはuser_two_factor_authsテーブルへのデータアクセスを担う。
//
// 秘密鍵は暗号化した値のまま受け渡し、このリポジトリは平文の秘密鍵を扱わない。
type UserTwoFactorAuthRepository struct {
	queries *query.Queries
}

// NewUserTwoFactorAuthRepositoryはUserTwoFactorAuthRepositoryを生成する。
func NewUserTwoFactorAuthRepository(queries *query.Queries) *UserTwoFactorAuthRepository {
	return &UserTwoFactorAuthRepository{queries: queries}
}

// WithTxはクエリをtx内で実行する新しいUserTwoFactorAuthRepositoryを返す。
func (r *UserTwoFactorAuthRepository) WithTx(tx *sql.Tx) *UserTwoFactorAuthRepository {
	return &UserTwoFactorAuthRepository{queries: r.queries.WithTx(tx)}
}

// UpsertPendingは認証アプリへの登録の途中の設定を作るか、その秘密鍵を差し替える。
// 既に有効にしているときは書き換えず (nil, nil) を返す。
//
// 登録の画面を開き直すたびに新しい秘密鍵にするため、途中の設定は上書きしてよい。
// 有効な設定を上書きすると、登録済みの認証アプリのコードが通らなくなるため守る。
func (r *UserTwoFactorAuthRepository) UpsertPending(ctx context.Context, userID model.UserID, secretCiphertext []byte) (*model.UserTwoFactorAuth, error) {
	row, err := r.queries.UpsertPendingUserTwoFactorAuth(ctx, query.UpsertPendingUserTwoFactorAuthParams{
		UserID:           int64(userID),
		SecretCiphertext: secretCiphertext,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return userTwoFactorAuthFromRow(row), nil
}

// FindByUserIDはユーザーの二要素認証の設定を返す。無いときは (nil, nil) を返す。
// 登録の途中の設定も返すため、有効かどうかは呼び出し側がIsEnabledで確かめる。
func (r *UserTwoFactorAuthRepository) FindByUserID(ctx context.Context, userID model.UserID) (*model.UserTwoFactorAuth, error) {
	row, err := r.queries.GetUserTwoFactorAuthByUserID(ctx, int64(userID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return userTwoFactorAuthFromRow(row), nil
}

// Enableは登録の途中の設定を有効にし、照合したコードのタイムステップを使用済みとして記録する。
// 登録の途中の設定が無い・既に有効・照合後に秘密鍵が差し替わったときは更新せずfalseを返す。
func (r *UserTwoFactorAuthRepository) Enable(ctx context.Context, userID model.UserID, expectedSecretCiphertext []byte, step int64) (bool, error) {
	affected, err := r.queries.EnableUserTwoFactorAuth(ctx, query.EnableUserTwoFactorAuthParams{
		UserID:                   int64(userID),
		ExpectedSecretCiphertext: expectedSecretCiphertext,
		Step:                     step,
	})
	if err != nil {
		return false, err
	}

	return affected > 0, nil
}

// UseMatchingStepは照合した有効な設定に限ってタイムステップを記録する。
// 設定が入れ替わったときや、同じステップ以前を既に使っているときはfalseを返す。
//
// 照合と記録を分けずに条件付きのUPDATEで判定するのは、同じコードを同時に送られても一方しか通さないため。
func (r *UserTwoFactorAuthRepository) UseMatchingStep(ctx context.Context, setting *model.UserTwoFactorAuth, step int64) (bool, error) {
	affected, err := r.queries.UseMatchingUserTwoFactorAuthStep(ctx, query.UseMatchingUserTwoFactorAuthStepParams{
		UserID:                   int64(setting.UserID),
		ExpectedID:               int64(setting.ID),
		ExpectedSecretCiphertext: setting.SecretCiphertext,
		Step:                     step,
	})
	if err != nil {
		return false, err
	}

	return affected > 0, nil
}

// LockMatchingは、照合した有効な設定の行をトランザクションの終わりまでロックする。
// 設定が入れ替わったときや、既に無効になったときはロックせずfalseを返す。
// トランザクションの中で呼ばないとロックはすぐに外れるため、呼び出し側はWithTxのリポジトリで呼ぶ。
func (r *UserTwoFactorAuthRepository) LockMatching(ctx context.Context, setting *model.UserTwoFactorAuth) (bool, error) {
	affected, err := r.queries.LockMatchingUserTwoFactorAuth(ctx, query.LockMatchingUserTwoFactorAuthParams{
		UserID:                   int64(setting.UserID),
		ExpectedID:               int64(setting.ID),
		ExpectedSecretCiphertext: setting.SecretCiphertext,
	})
	if err != nil {
		return false, err
	}

	return affected > 0, nil
}

// DeleteMatchingは本人確認で照合した有効な設定だけを削除する。
// 設定が入れ替わっていたときや、既に削除されているときはfalseを返す。
func (r *UserTwoFactorAuthRepository) DeleteMatching(ctx context.Context, setting *model.UserTwoFactorAuth) (bool, error) {
	affected, err := r.queries.DeleteMatchingUserTwoFactorAuth(ctx, query.DeleteMatchingUserTwoFactorAuthParams{
		UserID:                   int64(setting.UserID),
		ExpectedID:               int64(setting.ID),
		ExpectedSecretCiphertext: setting.SecretCiphertext,
	})
	if err != nil {
		return false, err
	}

	return affected > 0, nil
}

// DeleteByUserIDはユーザーの二要素認証の設定を削除し、実際に削除した設定を返す。
// 設定が無い (同時に送られた別の無効化が先に削除した) ときは (nil, nil) を返す。
func (r *UserTwoFactorAuthRepository) DeleteByUserID(ctx context.Context, userID model.UserID) (*model.UserTwoFactorAuth, error) {
	row, err := r.queries.DeleteUserTwoFactorAuthByUserID(ctx, int64(userID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return userTwoFactorAuthFromRow(row), nil
}

func userTwoFactorAuthFromRow(row query.UserTwoFactorAuth) *model.UserTwoFactorAuth {
	return &model.UserTwoFactorAuth{
		ID:               model.UserTwoFactorAuthID(row.ID),
		UserID:           model.UserID(row.UserID),
		SecretCiphertext: row.SecretCiphertext,
		LastUsedStep:     row.LastUsedStep,
		EnabledAt:        row.EnabledAt,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}
