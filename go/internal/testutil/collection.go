package testutil

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/annict/annict/go/internal/model"
)

// CreateTestCollectionはテスト用のコレクションをデータベースに作成し、IDを返す。
// 名前はユーザーごとの一意制約の対象のため、同じユーザーで複数作れるようUUIDから作る。
func CreateTestCollection(t *testing.T, tx *sql.Tx, userID model.UserID) int64 {
	t.Helper()

	var id int64
	err := tx.QueryRow(
		`INSERT INTO collections (user_id, name, created_at, updated_at) VALUES ($1, $2, $3, $3) RETURNING id`,
		int64(userID),
		"コレクション-"+uuid.New().String(),
		time.Now(),
	).Scan(&id)
	if err != nil {
		t.Fatalf("コレクションデータの作成に失敗しました: %v", err)
	}

	return id
}

// CreateTestCollectionItemはテスト用のコレクション項目をデータベースに作成し、IDを返す。
func CreateTestCollectionItem(t *testing.T, tx *sql.Tx, userID model.UserID, collectionID int64, workID model.WorkID) int64 {
	t.Helper()
	return createTestRow(t, tx, "コレクション項目",
		`INSERT INTO collection_items (user_id, collection_id, work_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $4) RETURNING id`,
		int64(userID), collectionID, int64(workID), time.Now(),
	)
}
