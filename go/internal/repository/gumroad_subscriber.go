package repository

import (
	"context"
	"time"

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
func (r *GumroadSubscriberRepository) GetByID(ctx context.Context, id int64) (query.GumroadSubscriber, error) {
	return r.queries.GetGumroadSubscriberByID(ctx, id)
}

// IsActive はサブスクリプションがアクティブかどうかを判定します
// Rails版のGumroadSubscriber.active?と同じロジック:
// !gumroad_cancelled_at&.past? && !gumroad_ended_at&.past?
func (r *GumroadSubscriberRepository) IsActive(subscriber *query.GumroadSubscriber) bool {
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
