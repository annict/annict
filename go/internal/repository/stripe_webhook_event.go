package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// StripeWebhookEventRepository はStripe Webhookイベントのリポジトリです
type StripeWebhookEventRepository struct {
	queries *query.Queries
}

// NewStripeWebhookEventRepository は新しいStripeWebhookEventRepositoryを作成します
func NewStripeWebhookEventRepository(queries *query.Queries) *StripeWebhookEventRepository {
	return &StripeWebhookEventRepository{queries: queries}
}

// Create は新しいWebhookイベントを作成します
func (r *StripeWebhookEventRepository) Create(ctx context.Context, params query.CreateStripeWebhookEventParams) (query.StripeWebhookEvent, error) {
	return r.queries.CreateStripeWebhookEvent(ctx, params)
}

// GetByStripeEventID はStripe Event IDでWebhookイベントを取得します
func (r *StripeWebhookEventRepository) GetByStripeEventID(ctx context.Context, stripeEventID string) (query.StripeWebhookEvent, error) {
	return r.queries.GetStripeWebhookEventByStripeEventID(ctx, stripeEventID)
}

// UpdateStatus はWebhookイベントのステータスを更新します
func (r *StripeWebhookEventRepository) UpdateStatus(ctx context.Context, params query.UpdateStripeWebhookEventStatusParams) error {
	return r.queries.UpdateStripeWebhookEventStatus(ctx, params)
}

// MarkAsProcessed はWebhookイベントを処理完了としてマークします
func (r *StripeWebhookEventRepository) MarkAsProcessed(ctx context.Context, id int64) error {
	return r.queries.UpdateStripeWebhookEventStatus(ctx, query.UpdateStripeWebhookEventStatusParams{
		ID:           id,
		Status:       model.WebhookEventStatusProcessed.String(),
		ErrorMessage: sql.NullString{},
		ProcessedAt:  sql.NullTime{Time: time.Now(), Valid: true},
	})
}

// MarkAsFailed はWebhookイベントを処理失敗としてマークします
func (r *StripeWebhookEventRepository) MarkAsFailed(ctx context.Context, id int64, errorMessage string) error {
	return r.queries.UpdateStripeWebhookEventStatus(ctx, query.UpdateStripeWebhookEventStatusParams{
		ID:           id,
		Status:       model.WebhookEventStatusFailed.String(),
		ErrorMessage: sql.NullString{String: errorMessage, Valid: true},
		ProcessedAt:  sql.NullTime{Time: time.Now(), Valid: true},
	})
}

// MarkAsSkipped はWebhookイベントを処理スキップとしてマークします
func (r *StripeWebhookEventRepository) MarkAsSkipped(ctx context.Context, id int64) error {
	return r.queries.UpdateStripeWebhookEventStatus(ctx, query.UpdateStripeWebhookEventStatusParams{
		ID:           id,
		Status:       model.WebhookEventStatusSkipped.String(),
		ErrorMessage: sql.NullString{},
		ProcessedAt:  sql.NullTime{Time: time.Now(), Valid: true},
	})
}

// Exists はStripe Event IDで既にイベントが存在するかを確認します（冪等性チェック用）
func (r *StripeWebhookEventRepository) Exists(ctx context.Context, stripeEventID string) (bool, error) {
	_, err := r.queries.GetStripeWebhookEventByStripeEventID(ctx, stripeEventID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
