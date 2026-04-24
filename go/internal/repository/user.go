// Package repository はデータアクセス層を提供します
package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// User はユーザーの型エイリアスです
// ハンドラーがqueryパッケージを直接参照しないようにするため、repositoryで公開します
type User = model.User

// UpdateUserPasswordParams はパスワード更新のパラメータの型エイリアスです
type UpdateUserPasswordParams = query.UpdateUserPasswordParams

// GetUserByIDRow はユーザーID検索結果の型エイリアスです
type GetUserByIDRow = query.GetUserByIDRow

// GetUserByEmailOrUsernameRow はメールアドレス/ユーザー名検索結果の型エイリアスです
type GetUserByEmailOrUsernameRow = query.GetUserByEmailOrUsernameRow

// UserRepository はUser関連のデータアクセスを担当します
type UserRepository struct {
	queries               *query.Queries
	stripeSubscriberRepo  *StripeSubscriberRepository
	gumroadSubscriberRepo *GumroadSubscriberRepository
}

// NewUserRepository はUserRepositoryを作成します
func NewUserRepository(queries *query.Queries) *UserRepository {
	return &UserRepository{queries: queries}
}

// WithTx はトランザクションを使用する新しいRepositoryを返します
func (r *UserRepository) WithTx(tx *sql.Tx) *UserRepository {
	return &UserRepository{queries: r.queries.WithTx(tx)}
}

// GetByEmailForSignIn はメールアドレスでユーザーを検索します（サインイン用）
func (r *UserRepository) GetByEmailForSignIn(ctx context.Context, email string) (query.GetUserByEmailForSignInRow, error) {
	return r.queries.GetUserByEmailForSignIn(ctx, email)
}

// GetByEmailOrUsername はメールアドレスまたはユーザー名でユーザーを検索します
func (r *UserRepository) GetByEmailOrUsername(ctx context.Context, emailOrUsername string) (query.GetUserByEmailOrUsernameRow, error) {
	return r.queries.GetUserByEmailOrUsername(ctx, emailOrUsername)
}

// GetByEmail はメールアドレスでユーザーを検索します
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (query.GetUserByEmailRow, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

// GetByID はユーザーIDでユーザーを検索します
func (r *UserRepository) GetByID(ctx context.Context, id int64) (query.GetUserByIDRow, error) {
	return r.queries.GetUserByID(ctx, id)
}

// UserCreateParams はユーザー作成のパラメータ
type UserCreateParams struct {
	Username          string
	Email             string
	EncryptedPassword string
	Locale            string
}

// Create はユーザーを作成します
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
		ID:        row.ID,
		Username:  row.Username,
		Email:     row.Email,
		Role:      row.Role,
		Locale:    row.Locale,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

// GetByUsername はユーザー名でユーザーの存在を確認します
// ユーザーが存在しない場合はsql.ErrNoRowsを返します
func (r *UserRepository) GetByUsername(ctx context.Context, username string) error {
	_, err := r.queries.GetUserByUsername(ctx, username)
	return err
}

// UpdateStripeSubscriberID はユーザーのStripeサブスクライバーIDを更新します
func (r *UserRepository) UpdateStripeSubscriberID(ctx context.Context, userID int64, stripeSubscriberID *int64) error {
	var nullableID sql.NullInt64
	if stripeSubscriberID != nil {
		nullableID = sql.NullInt64{Int64: *stripeSubscriberID, Valid: true}
	}
	return r.queries.UpdateUserStripeSubscriberID(ctx, query.UpdateUserStripeSubscriberIDParams{
		ID:                 userID,
		StripeSubscriberID: nullableID,
	})
}

// GetByStripeSubscriberID はStripeサブスクライバーIDでユーザーを検索します
func (r *UserRepository) GetByStripeSubscriberID(ctx context.Context, stripeSubscriberID int64) (query.GetUserByStripeSubscriberIDRow, error) {
	return r.queries.GetUserByStripeSubscriberID(ctx, sql.NullInt64{Int64: stripeSubscriberID, Valid: true})
}

// FindUserIDByStripeSubscriberID はStripeサブスクライバーIDからユーザーIDを検索します
// ユーザーが見つからない場合はnilを返します（sql.ErrNoRowsの場合）
func (r *UserRepository) FindUserIDByStripeSubscriberID(ctx context.Context, stripeSubscriberID int64) (*int64, error) {
	user, err := r.queries.GetUserByStripeSubscriberID(ctx, sql.NullInt64{Int64: stripeSubscriberID, Valid: true})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user.ID, nil
}

// WithStripeSubscriberRepo はStripeSubscriberRepositoryを設定します
func (r *UserRepository) WithStripeSubscriberRepo(repo *StripeSubscriberRepository) *UserRepository {
	r.stripeSubscriberRepo = repo
	return r
}

// WithGumroadSubscriberRepo はGumroadSubscriberRepositoryを設定します
func (r *UserRepository) WithGumroadSubscriberRepo(repo *GumroadSubscriberRepository) *UserRepository {
	r.gumroadSubscriberRepo = repo
	return r
}

// UpdatePassword はユーザーのパスワードを更新します
func (r *UserRepository) UpdatePassword(ctx context.Context, params query.UpdateUserPasswordParams) error {
	return r.queries.UpdateUserPassword(ctx, params)
}

// IsSupporter はユーザーがサポーターかどうかを判定します
// Stripeサブスクリプションまたは（移行期間中は）Gumroadサブスクリプションがアクティブな場合にtrueを返します
func (r *UserRepository) IsSupporter(ctx context.Context, user *query.User) (bool, error) {
	// Stripeサブスクリプションをチェック
	if user.StripeSubscriberID.Valid && r.stripeSubscriberRepo != nil {
		stripeSubscriber, err := r.stripeSubscriberRepo.GetByID(ctx, user.StripeSubscriberID.Int64)
		if err == nil && r.stripeSubscriberRepo.IsActive(&stripeSubscriber) {
			return true, nil
		}
	}

	// Gumroadサブスクリプションをチェック（移行期間中）
	if user.GumroadSubscriberID.Valid && r.gumroadSubscriberRepo != nil {
		gumroadSubscriber, err := r.gumroadSubscriberRepo.GetByID(ctx, user.GumroadSubscriberID.Int64)
		if err == nil && r.gumroadSubscriberRepo.IsActive(&gumroadSubscriber) {
			return true, nil
		}
	}

	return false, nil
}
