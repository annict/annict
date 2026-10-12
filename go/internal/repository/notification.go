package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// NotificationRepositoryは通知関連のデータアクセスを担当します
type NotificationRepository struct {
	queries *query.Queries
}

// NewNotificationRepositoryはNotificationRepositoryを作成します
func NewNotificationRepository(queries *query.Queries) *NotificationRepository {
	return &NotificationRepository{queries: queries}
}

// WithTxはトランザクションを使用する新しいRepositoryを返します
func (r *NotificationRepository) WithTx(tx *sql.Tx) *NotificationRepository {
	return &NotificationRepository{queries: r.queries.WithTx(tx)}
}

// DeleteByActionUserIDはユーザーが起こした通知を最大limit件削除し、削除した件数を返します。
// 他人宛ての未読の通知の分だけ、宛先の未読数を減算します。
func (r *NotificationRepository) DeleteByActionUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteNotificationsByActionUserID(ctx, query.DeleteNotificationsByActionUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}

// DeleteByUserIDはユーザー宛ての通知を最大limit件削除し、削除した件数を返します。
func (r *NotificationRepository) DeleteByUserID(ctx context.Context, userID model.UserID, limit int32) (int64, error) {
	return r.queries.DeleteNotificationsByUserID(ctx, query.DeleteNotificationsByUserIDParams{
		UserID:    int64(userID),
		BatchSize: limit,
	})
}
