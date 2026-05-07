package repository

import (
	"context"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// GumroadSubscriberRepository はGumroadサブスクライバー関連のデータアクセスを担当します
type GumroadSubscriberRepository struct {
	queries *query.Queries
}

// NewGumroadSubscriberRepository はGumroadSubscriberRepositoryを作成します
func NewGumroadSubscriberRepository(queries *query.Queries) *GumroadSubscriberRepository {
	return &GumroadSubscriberRepository{queries: queries}
}

// GetByID はIDでGumroadサブスクライバーを検索します
func (r *GumroadSubscriberRepository) GetByID(ctx context.Context, id model.GumroadSubscriberID) (model.GumroadSubscriber, error) {
	row, err := r.queries.GetGumroadSubscriberByID(ctx, int64(id))
	if err != nil {
		return model.GumroadSubscriber{}, err
	}
	return toGumroadSubscriberModel(row), nil
}

// IsActive はサブスクリプションがアクティブかどうかを判定します
// Rails版のGumroadSubscriber.active?と同じロジック:
// !gumroad_cancelled_at&.past? && !gumroad_ended_at&.past?
func (r *GumroadSubscriberRepository) IsActive(subscriber *model.GumroadSubscriber) bool {
	now := time.Now()

	// キャンセル日時が過去でないこと
	if subscriber.GumroadCancelledAt.Valid && subscriber.GumroadCancelledAt.Time.Before(now) {
		return false
	}

	// 終了日時が過去でないこと
	if subscriber.GumroadEndedAt.Valid && subscriber.GumroadEndedAt.Time.Before(now) {
		return false
	}

	return true
}

// toGumroadSubscriberModel はqueryの結果をモデルに変換します
func toGumroadSubscriberModel(row query.GumroadSubscriber) model.GumroadSubscriber {
	return model.GumroadSubscriber{
		ID:                                 model.GumroadSubscriberID(row.ID),
		GumroadID:                          row.GumroadID,
		GumroadProductID:                   row.GumroadProductID,
		GumroadProductName:                 row.GumroadProductName,
		GumroadUserID:                      row.GumroadUserID,
		GumroadUserEmail:                   row.GumroadUserEmail,
		GumroadPurchaseIds:                 row.GumroadPurchaseIds,
		GumroadCreatedAt:                   row.GumroadCreatedAt,
		GumroadCancelledAt:                 row.GumroadCancelledAt,
		GumroadUserRequestedCancellationAt: row.GumroadUserRequestedCancellationAt,
		GumroadChargeOccurrenceCount:       row.GumroadChargeOccurrenceCount,
		GumroadEndedAt:                     row.GumroadEndedAt,
		CreatedAt:                          row.CreatedAt,
		UpdatedAt:                          row.UpdatedAt,
	}
}
