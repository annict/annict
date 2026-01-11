// Package repository はデータアクセス層を提供します
package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/query"
)

// User はユーザーの型エイリアスです
// ハンドラーがqueryパッケージを直接参照しないようにするため、repositoryで公開します
type User = query.GetUserByIDRow

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
