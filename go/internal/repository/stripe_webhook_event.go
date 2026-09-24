package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// StripeWebhookEventはWebhookイベントの型エイリアスです
type StripeWebhookEvent = query.StripeWebhookEvent

// CreateStripeWebhookEventParamsはWebhookイベント作成時のパラメータです
type CreateStripeWebhookEventParams struct {
	StripeEventID   string
	StripeEventType string
	StripePayload   json.RawMessage
	Status          string
	ReceivedAt      time.Time
}

// StripeWebhookEventRepositoryはStripe Webhookイベントのリポジトリです
type StripeWebhookEventRepository struct {
	queries *query.Queries
}

// NewStripeWebhookEventRepositoryは新しいStripeWebhookEventRepositoryを作成します
func NewStripeWebhookEventRepository(queries *query.Queries) *StripeWebhookEventRepository {
	return &StripeWebhookEventRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *StripeWebhookEventRepository) WithTx(tx *sql.Tx) *StripeWebhookEventRepository {
	return &StripeWebhookEventRepository{queries: r.queries.WithTx(tx)}
}

// Createは新しいWebhookイベントを作成します
func (r *StripeWebhookEventRepository) Create(ctx context.Context, params CreateStripeWebhookEventParams) (StripeWebhookEvent, error) {
	return r.queries.CreateStripeWebhookEvent(ctx, query.CreateStripeWebhookEventParams{
		StripeEventID:   params.StripeEventID,
		StripeEventType: params.StripeEventType,
		StripePayload:   params.StripePayload,
		Status:          params.Status,
		ReceivedAt:      params.ReceivedAt,
	})
}

// GetByStripeEventIDはStripe Event IDでWebhookイベントを取得します
func (r *StripeWebhookEventRepository) GetByStripeEventID(ctx context.Context, stripeEventID string) (query.StripeWebhookEvent, error) {
	return r.queries.GetStripeWebhookEventByStripeEventID(ctx, stripeEventID)
}

// UpdateStatusはWebhookイベントのステータスを更新します
func (r *StripeWebhookEventRepository) UpdateStatus(ctx context.Context, params query.UpdateStripeWebhookEventStatusParams) error {
	return r.queries.UpdateStripeWebhookEventStatus(ctx, params)
}

// MarkAsProcessedはWebhookイベントを処理完了としてマークします
func (r *StripeWebhookEventRepository) MarkAsProcessed(ctx context.Context, id model.StripeWebhookEventID) error {
	return r.queries.UpdateStripeWebhookEventStatus(ctx, query.UpdateStripeWebhookEventStatusParams{
		ID:           int64(id),
		Status:       model.WebhookEventStatusProcessed.String(),
		ErrorMessage: sql.NullString{},
		ProcessedAt:  sql.NullTime{Time: time.Now(), Valid: true},
	})
}

// MarkAsFailedはWebhookイベントを処理失敗としてマークします
func (r *StripeWebhookEventRepository) MarkAsFailed(ctx context.Context, id model.StripeWebhookEventID, errorMessage string) error {
	return r.queries.UpdateStripeWebhookEventStatus(ctx, query.UpdateStripeWebhookEventStatusParams{
		ID:           int64(id),
		Status:       model.WebhookEventStatusFailed.String(),
		ErrorMessage: sql.NullString{String: errorMessage, Valid: true},
		ProcessedAt:  sql.NullTime{Time: time.Now(), Valid: true},
	})
}

// MarkAsSkippedはWebhookイベントを処理スキップとしてマークします
func (r *StripeWebhookEventRepository) MarkAsSkipped(ctx context.Context, id model.StripeWebhookEventID) error {
	return r.queries.UpdateStripeWebhookEventStatus(ctx, query.UpdateStripeWebhookEventStatusParams{
		ID:           int64(id),
		Status:       model.WebhookEventStatusSkipped.String(),
		ErrorMessage: sql.NullString{},
		ProcessedAt:  sql.NullTime{Time: time.Now(), Valid: true},
	})
}

// ExistsはStripe Event IDで既にイベントが存在するかを確認します (冪等性チェック用)
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
