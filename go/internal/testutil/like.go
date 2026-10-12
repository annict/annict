package testutil

import (
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
)

// CreateTestLikeはテスト用のいいねをデータベースに作成し、IDを返す。
// recipientTypeはいいね先のモデル名 (Status / EpisodeRecordなど)。
func CreateTestLike(t *testing.T, tx *sql.Tx, userID model.UserID, recipientType string, recipientID int64) int64 {
	t.Helper()

	var id int64
	err := tx.QueryRow(
		`INSERT INTO likes (user_id, recipient_type, recipient_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $4) RETURNING id`,
		int64(userID),
		recipientType,
		recipientID,
		time.Now(),
	).Scan(&id)
	if err != nil {
		t.Fatalf("いいねデータの作成に失敗しました: %v", err)
	}

	return id
}
