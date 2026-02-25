// Package repository はデータアクセス層を提供します
package repository

import (
	"context"
	"database/sql"

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
func (r *EmailNotificationRepository) Create(ctx context.Context, userID int64, unsubscriptionKey string) error {
	_, err := r.queries.CreateEmailNotification(ctx, query.CreateEmailNotificationParams{
		UserID:            userID,
		UnsubscriptionKey: unsubscriptionKey,
	})
	return err
}
