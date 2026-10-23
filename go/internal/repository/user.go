// Package repositoryはデータアクセス層を提供します
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// Userはユーザーの型エイリアスです
// ハンドラーがqueryパッケージを直接参照しないようにするため、repositoryで公開します
type User = model.User

// UpdateUserPasswordParamsはパスワード更新のパラメータの型エイリアスです
type UpdateUserPasswordParams = query.UpdateUserPasswordParams

// GetActiveUserByIDRowはユーザーID検索結果の型エイリアスです
type GetActiveUserByIDRow = query.GetActiveUserByIDRow

// GetActiveUserByEmailOrUsernameRowはメールアドレス/ユーザー名検索結果の型エイリアスです
type GetActiveUserByEmailOrUsernameRow = query.GetActiveUserByEmailOrUsernameRow

// UserRepositoryはUser関連のデータアクセスを担当します
type UserRepository struct {
	queries               *query.Queries
	stripeSubscriberRepo  *StripeSubscriberRepository
	gumroadSubscriberRepo *GumroadSubscriberRepository
}

// NewUserRepositoryはUserRepositoryを作成します
func NewUserRepository(queries *query.Queries) *UserRepository {
	return &UserRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *UserRepository) WithTx(tx *sql.Tx) *UserRepository {
	return &UserRepository{queries: r.queries.WithTx(tx)}
}

// GetActiveByEmailForSignInは論理削除されていないユーザーをメールアドレスで検索します (サインイン用)
func (r *UserRepository) GetActiveByEmailForSignIn(ctx context.Context, email string) (query.GetActiveUserByEmailForSignInRow, error) {
	return r.queries.GetActiveUserByEmailForSignIn(ctx, email)
}

// GetActiveByEmailOrUsernameは論理削除されていないユーザーをメールアドレスまたはユーザー名で検索します
func (r *UserRepository) GetActiveByEmailOrUsername(ctx context.Context, emailOrUsername string) (query.GetActiveUserByEmailOrUsernameRow, error) {
	return r.queries.GetActiveUserByEmailOrUsername(ctx, emailOrUsername)
}

// GetByEmailはメールアドレスでユーザーを検索します
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (query.GetUserByEmailRow, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

// GetActiveByIDは論理削除されていないユーザーをユーザーIDで検索します
func (r *UserRepository) GetActiveByID(ctx context.Context, id model.UserID) (query.GetActiveUserByIDRow, error) {
	return r.queries.GetActiveUserByID(ctx, int64(id))
}

// UserCreateParamsはユーザー作成のパラメータ
type UserCreateParams struct {
	Username          string
	Email             string
	EncryptedPassword string
	Locale            string
}

// Createはユーザーを作成します
func (r *UserRepository) Create(ctx context.Context, params UserCreateParams) (*model.User, error) {
	row, err := r.queries.CreateUser(ctx, query.CreateUserParams{
		Username:          params.Username,
		Email:             params.Email,
		EncryptedPassword: params.EncryptedPassword,
		Locale:            params.Locale,
	})
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:        model.UserID(row.ID),
		Username:  row.Username,
		Email:     row.Email,
		Role:      row.Role,
		Locale:    row.Locale,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

// GetByUsernameはユーザー名でユーザーの存在を確認します
// ユーザーが存在しない場合はsql.ErrNoRowsを返します
func (r *UserRepository) GetByUsername(ctx context.Context, username string) error {
	_, err := r.queries.GetUserByUsername(ctx, username)
	return err
}

// FindActiveIDByUsernameは論理削除されていないユーザーのIDをusernameで取得する。
// ユーザーが存在しない場合や削除済み (deleted_at IS NOT NULL) の場合は
// (nil, nil) を返す。
func (r *UserRepository) FindActiveIDByUsername(ctx context.Context, username string) (*model.UserID, error) {
	id, err := r.queries.GetActiveUserIDByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	userID := model.UserID(id)
	return &userID, nil
}

// FindActiveIDByEmailは論理削除されていないユーザーのIDをメールアドレスで取得する。
// ユーザーが存在しない場合や削除済み (deleted_at IS NOT NULL) の場合は
// (nil, nil) を返す。
func (r *UserRepository) FindActiveIDByEmail(ctx context.Context, email string) (*model.UserID, error) {
	id, err := r.queries.GetActiveUserIDByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	userID := model.UserID(id)
	return &userID, nil
}

// FindDeletionTargetは退会の削除処理の対象として、論理削除済みのユーザーも含めて
// IDで取得する。ユーザーが存在しない (削除を終えた) 場合は (nil, nil) を返す。
func (r *UserRepository) FindDeletionTarget(ctx context.Context, id model.UserID) (*model.UserDeletionTarget, error) {
	row, err := r.queries.GetUserForDeletion(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	target := &model.UserDeletionTarget{
		ID:      model.UserID(row.ID),
		Deleted: row.DeletedAt.Valid,
	}
	if row.StripeSubscriberID.Valid {
		stripeSubscriberID := model.StripeSubscriberID(row.StripeSubscriberID.Int64)
		target.StripeSubscriberID = &stripeSubscriberID
	}
	return target, nil
}

// AnonymizeAndSoftDeleteは退会を受け付けたユーザーのユーザー名とメールアドレスを置き換え、
// 論理削除します。検証済みのユーザー名と一致しない場合や、すでに論理削除された場合は
// 変更せず、変更した件数を返します。
func (r *UserRepository) AnonymizeAndSoftDelete(ctx context.Context, id model.UserID, expectedUsername, username, email string) (int64, error) {
	return r.queries.AnonymizeAndSoftDeleteUser(ctx, query.AnonymizeAndSoftDeleteUserParams{
		ID:               int64(id),
		ExpectedUsername: expectedUsername,
		Username:         username,
		Email:            email,
	})
}

// LockForAccountDeletionは退会の受付中にユーザー行をロックし、匿名化する前の最新の属性と、
// 論理削除済みかどうかを返します。ユーザーが存在しない場合はsql.ErrNoRowsを返します。
func (r *UserRepository) LockForAccountDeletion(ctx context.Context, id model.UserID) (*model.AccountDeletionUser, error) {
	row, err := r.queries.LockUserForAccountDeletion(ctx, int64(id))
	if err != nil {
		return nil, err
	}

	user := &model.AccountDeletionUser{
		Username: row.Username,
		Email:    row.Email,
		Locale:   row.Locale,
		Deleted:  row.DeletedAt.Valid,
	}
	if row.StripeSubscriberID.Valid {
		stripeSubscriberID := model.StripeSubscriberID(row.StripeSubscriberID.Int64)
		user.StripeSubscriberID = &stripeSubscriberID
	}
	return user, nil
}

// UpdateStripeSubscriberIDはユーザーのStripeサブスクライバーIDを更新します
func (r *UserRepository) UpdateStripeSubscriberID(ctx context.Context, userID model.UserID, stripeSubscriberID *model.StripeSubscriberID) error {
	var nullableID sql.NullInt64
	if stripeSubscriberID != nil {
		nullableID = sql.NullInt64{Int64: int64(*stripeSubscriberID), Valid: true}
	}
	return r.queries.UpdateUserStripeSubscriberID(ctx, query.UpdateUserStripeSubscriberIDParams{
		ID:                 int64(userID),
		StripeSubscriberID: nullableID,
	})
}

// GetByStripeSubscriberIDはStripeサブスクライバーIDでユーザーを検索します
func (r *UserRepository) GetByStripeSubscriberID(ctx context.Context, stripeSubscriberID model.StripeSubscriberID) (query.GetUserByStripeSubscriberIDRow, error) {
	return r.queries.GetUserByStripeSubscriberID(ctx, sql.NullInt64{Int64: int64(stripeSubscriberID), Valid: true})
}

// ClearActiveUserStripeSubscriberIDは論理削除されていないユーザーの紐付けを外し、
// 対象のユーザーIDを返します。対象が無い場合はnilを返します。
func (r *UserRepository) ClearActiveUserStripeSubscriberID(ctx context.Context, stripeSubscriberID model.StripeSubscriberID) (*model.UserID, error) {
	id, err := r.queries.ClearActiveUserStripeSubscriberID(ctx, sql.NullInt64{Int64: int64(stripeSubscriberID), Valid: true})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	userID := model.UserID(id)
	return &userID, nil
}

// WithStripeSubscriberRepoはStripeSubscriberRepositoryを設定します
func (r *UserRepository) WithStripeSubscriberRepo(repo *StripeSubscriberRepository) *UserRepository {
	r.stripeSubscriberRepo = repo
	return r
}

// WithGumroadSubscriberRepoはGumroadSubscriberRepositoryを設定します
func (r *UserRepository) WithGumroadSubscriberRepo(repo *GumroadSubscriberRepository) *UserRepository {
	r.gumroadSubscriberRepo = repo
	return r
}

// UpdatePasswordはユーザーのパスワードを更新します
func (r *UserRepository) UpdatePassword(ctx context.Context, params query.UpdateUserPasswordParams) error {
	return r.queries.UpdateUserPassword(ctx, params)
}

// IsSupporterはユーザーがサポーターかどうかを判定します
// Stripeサブスクリプションまたは (移行期間中は) Gumroadサブスクリプションがアクティブな場合にtrueを返します
func (r *UserRepository) IsSupporter(ctx context.Context, user *model.User) (bool, error) {
	// Stripeサブスクリプションをチェック
	if user.StripeSubscriberID != nil && r.stripeSubscriberRepo != nil {
		stripeSubscriber, err := r.stripeSubscriberRepo.GetByID(ctx, *user.StripeSubscriberID)
		if err == nil && stripeSubscriber != nil && r.stripeSubscriberRepo.IsActive(stripeSubscriber) {
			return true, nil
		}
	}

	// Gumroadサブスクリプションをチェック (移行期間中)
	if user.GumroadSubscriberID != nil && r.gumroadSubscriberRepo != nil {
		gumroadSubscriber, err := r.gumroadSubscriberRepo.GetByID(ctx, *user.GumroadSubscriberID)
		if err == nil && r.gumroadSubscriberRepo.IsActive(&gumroadSubscriber) {
			return true, nil
		}
	}

	return false, nil
}

// DeleteByIDはユーザー行を、外部キーにON DELETEの無い従属データ (設定・メール通知設定・
// メールアドレスの確認) とほかのユーザーから参照されていない購読情報と一緒に削除し、
// 削除した件数を返します。そのほかの従属データは外部キーのON DELETEに従って扱われます。
// 購読情報を共有するユーザーの削除を直列化するため、購読情報の行をロックしてから、別の文で
// 最後の参照者かを確かめて削除します。ロックを削除の文まで保持するよう、WithTxで
// トランザクションを渡したRepositoryから呼んでください。
func (r *UserRepository) DeleteByID(ctx context.Context, id model.UserID) (int64, error) {
	user, err := r.queries.GetUserForDeletion(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("削除対象ユーザーの取得に失敗: %w", err)
	}
	// StripeとGumroadの購読情報を両方持つユーザーどうしが互いのロックを待ってデッドロックしないよう、
	// 常にStripe → Gumroadの順でロックする。
	if user.StripeSubscriberID.Valid {
		if _, err := r.queries.LockStripeSubscriberForUserDeletion(ctx, user.StripeSubscriberID.Int64); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("購読情報(Stripe)のロックに失敗: %w", err)
		}
	}
	if user.GumroadSubscriberID.Valid {
		if _, err := r.queries.LockGumroadSubscriberForUserDeletion(ctx, user.GumroadSubscriberID.Int64); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("購読情報(Gumroad)のロックに失敗: %w", err)
		}
	}
	count, err := r.queries.DeleteUser(ctx, int64(id))
	if err != nil {
		return 0, fmt.Errorf("ユーザー行の削除に失敗: %w", err)
	}
	return count, nil
}
