// Package repository はデータアクセス層を提供します
package repository

import (
	"context"
	"database/sql"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
)

// EmailNotificationRepository はEmailNotification関連のデータアクセスを担当します
type EmailNotificationRepository struct {
	queries *query.Queries
}

// NewEmailNotificationRepository はEmailNotificationRepositoryを作成します
func NewEmailNotificationRepository(queries *query.Queries) *EmailNotificationRepository {
	return &EmailNotificationRepository{queries: queries}
}

// WithTx はトランザクションを使用する新しいRepositoryを返します
func (r *EmailNotificationRepository) WithTx(tx *sql.Tx) *EmailNotificationRepository {
	return &EmailNotificationRepository{queries: r.queries.WithTx(tx)}
}

// Create はメール通知設定を作成します
func (r *EmailNotificationRepository) Create(ctx context.Context, userID model.UserID, unsubscriptionKey string) (*model.EmailNotification, error) {
	row, err := r.queries.CreateEmailNotification(ctx, query.CreateEmailNotificationParams{
		UserID:            int64(userID),
		UnsubscriptionKey: unsubscriptionKey,
	})
	if err != nil {
		return nil, err
	}

	return &model.EmailNotification{
		ID:                      model.EmailNotificationID(row.ID),
		UserID:                  model.UserID(row.UserID),
		UnsubscriptionKey:       row.UnsubscriptionKey,
		EventFollowedUser:       row.EventFollowedUser,
		EventLikedEpisodeRecord: row.EventLikedEpisodeRecord,
		CreatedAt:               row.CreatedAt,
		UpdatedAt:               row.UpdatedAt,
	}, nil
}
