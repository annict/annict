package testutil

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/annict/annict/go/internal/model"
)

// CreateTestWorkTagはテスト用の作品のタグをデータベースに作成し、IDを返す。
// タグ名は一意制約の対象のため、並列のテストで衝突しないようUUIDから作る。
func CreateTestWorkTag(t *testing.T, tx *sql.Tx) int64 {
	t.Helper()
	return createTestRow(t, tx, "作品のタグ",
		`INSERT INTO work_tags (name, created_at, updated_at) VALUES ($1, $2, $2) RETURNING id`,
		"タグ-"+uuid.New().String(), time.Now(),
	)
}

// CreateTestWorkTaggingはテスト用の作品へのタグ付けをデータベースに作成し、IDを返す。
func CreateTestWorkTagging(t *testing.T, tx *sql.Tx, userID model.UserID, workID model.WorkID, workTagID int64) int64 {
	t.Helper()
	return createTestRow(t, tx, "作品へのタグ付け",
		`INSERT INTO work_taggings (user_id, work_id, work_tag_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $4) RETURNING id`,
		int64(userID), int64(workID), workTagID, time.Now(),
	)
}

// CreateTestWorkTaggableはテスト用のユーザーが使ったタグの記録をデータベースに作成し、IDを返す。
func CreateTestWorkTaggable(t *testing.T, tx *sql.Tx, userID model.UserID, workTagID int64) int64 {
	t.Helper()
	return createTestRow(t, tx, "ユーザーが使ったタグ",
		`INSERT INTO work_taggables (user_id, work_tag_id, created_at, updated_at) VALUES ($1, $2, $3, $3) RETURNING id`,
		int64(userID), workTagID, time.Now(),
	)
}
