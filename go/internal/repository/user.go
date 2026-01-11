// Package repository はデータアクセス層を提供します
package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/query"
)

// UserRepository はUser関連のデータアクセスを担当します
type UserRepository struct {
	queries *query.Queries
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
