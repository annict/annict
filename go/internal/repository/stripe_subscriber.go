package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// CreateStripeSubscriberParams はStripeサブスクライバー作成のパラメータの型エイリアスです
type CreateStripeSubscriberParams = query.CreateStripeSubscriberParams

// UpdateStripeSubscriberParams はStripeサブスクライバー更新のパラメータの型エイリアスです
type UpdateStripeSubscriberParams = query.UpdateStripeSubscriberParams

// StripeSubscriberRepository はStripeサブスクライバー関連のデータアクセスを担当します
type StripeSubscriberRepository struct {
	queries *query.Queries
}

// NewStripeSubscriberRepository はStripeSubscriberRepositoryを作成します
func NewStripeSubscriberRepository(queries *query.Queries) *StripeSubscriberRepository {
	return &StripeSubscriberRepository{queries: queries}
}

// WithTx はトランザクションを使用する新しいRepositoryを返します
func (r *StripeSubscriberRepository) WithTx(tx *sql.Tx) *StripeSubscriberRepository {
	return &StripeSubscriberRepository{queries: r.queries.WithTx(tx)}
}

// Create は新しいStripeサブスクライバーを作成します
func (r *StripeSubscriberRepository) Create(ctx context.Context, params query.CreateStripeSubscriberParams) (model.StripeSubscriber, error) {
	row, err := r.queries.CreateStripeSubscriber(ctx, params)
	if err != nil {
		return model.StripeSubscriber{}, err
	}
	return toStripeSubscriberModel(row), nil
}

// GetByID はIDでStripeサブスクライバーを検索します
func (r *StripeSubscriberRepository) GetByID(ctx context.Context, id int64) (model.StripeSubscriber, error) {
	row, err := r.queries.GetStripeSubscriberByID(ctx, id)
	if err != nil {
		return model.StripeSubscriber{}, err
	}
	return toStripeSubscriberModel(row), nil
}

// GetByStripeCustomerID はStripe顧客IDでStripeサブスクライバーを検索します
func (r *StripeSubscriberRepository) GetByStripeCustomerID(ctx context.Context, stripeCustomerID string) (query.StripeSubscriber, error) {
	return r.queries.GetStripeSubscriberByStripeCustomerID(ctx, stripeCustomerID)
}

// GetByStripeSubscriptionID はStripeサブスクリプションIDでStripeサブスクライバーを検索します
func (r *StripeSubscriberRepository) GetByStripeSubscriptionID(ctx context.Context, stripeSubscriptionID string) (query.StripeSubscriber, error) {
	return r.queries.GetStripeSubscriberByStripeSubscriptionID(ctx, stripeSubscriptionID)
}

// Update はStripeサブスクライバーの情報を更新します
func (r *StripeSubscriberRepository) Update(ctx context.Context, params query.UpdateStripeSubscriberParams) error {
	return r.queries.UpdateStripeSubscriber(ctx, params)
}

// UpdateStatus はStripeサブスクライバーのステータスのみを更新します
func (r *StripeSubscriberRepository) UpdateStatus(ctx context.Context, params query.UpdateStripeSubscriberStatusParams) error {
	return r.queries.UpdateStripeSubscriberStatus(ctx, params)
}

// IsActive はサブスクリプションがアクティブかどうかを判定します
// active または past_due 状態をアクティブとして扱います
// past_due は支払い遅延中だが、Stripeがリトライ中のため猶予期間として利用可能
func (r *StripeSubscriberRepository) IsActive(subscriber *model.StripeSubscriber) bool {
	status := model.StripeSubscriptionStatus(subscriber.StripeStatus)
	return status.IsActive()
}

// toStripeSubscriberModel はqueryの結果をモデルに変換します
func toStripeSubscriberModel(row query.StripeSubscriber) model.StripeSubscriber {
	return model.StripeSubscriber{
		ID:                       row.ID,
		StripeCustomerID:         row.StripeCustomerID,
		StripeSubscriptionID:     row.StripeSubscriptionID,
		StripePriceID:            row.StripePriceID,
		StripeStatus:             row.StripeStatus,
		StripeCurrentPeriodStart: row.StripeCurrentPeriodStart,
		StripeCurrentPeriodEnd:   row.StripeCurrentPeriodEnd,
		StripeCancelAt:           row.StripeCancelAt,
		StripeCanceledAt:         row.StripeCanceledAt,
		CreatedAt:                row.CreatedAt,
		UpdatedAt:                row.UpdatedAt,
	}
}
