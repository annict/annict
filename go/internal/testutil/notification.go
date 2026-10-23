package testutil

import (
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
)

// CreateTestNotificationはテスト用の通知をデータベースに作成し、IDを返す。
// userIDは通知の宛先、actionUserIDは通知のきっかけを作ったユーザー。
func CreateTestNotification(t *testing.T, tx *sql.Tx, userID, actionUserID model.UserID, trackableType string, trackableID int64) int64 {
	t.Helper()

	var id int64
	err := tx.QueryRow(
		`INSERT INTO notifications (
			user_id, action_user_id, trackable_type, trackable_id, action, created_at, updated_at
		) VALUES ($1, $2, $3, $4, 'likes.create', $5, $5) RETURNING id`,
		int64(userID),
		int64(actionUserID),
		trackableType,
		trackableID,
		time.Now(),
	).Scan(&id)
	if err != nil {
		t.Fatalf("通知データの作成に失敗しました: %v", err)
	}

	return id
}
