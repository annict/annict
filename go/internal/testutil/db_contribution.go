package testutil

import (
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
)

// CreateTestDbActivityはテスト用のAnnict DBの編集履歴をデータベースに作成し、IDを返す。
func CreateTestDbActivity(t *testing.T, tx *sql.Tx, userID model.UserID, workID model.WorkID) int64 {
	t.Helper()
	return createTestRow(t, tx, "Annict DBの編集履歴",
		`INSERT INTO db_activities (user_id, trackable_id, trackable_type, action, created_at, updated_at)
		VALUES ($1, $2, 'Work', 'works.update', $3, $3) RETURNING id`,
		int64(userID), int64(workID), time.Now(),
	)
}

// CreateTestDbCommentはテスト用のAnnict DBのコメントをデータベースに作成し、IDを返す。
func CreateTestDbComment(t *testing.T, tx *sql.Tx, userID model.UserID, workID model.WorkID) int64 {
	t.Helper()
	return createTestRow(t, tx, "Annict DBのコメント",
		`INSERT INTO db_comments (user_id, resource_id, resource_type, body, created_at, updated_at)
		VALUES ($1, $2, 'Work', 'コメント', $3, $3) RETURNING id`,
		int64(userID), int64(workID), time.Now(),
	)
}
