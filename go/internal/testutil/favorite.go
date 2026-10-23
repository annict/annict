package testutil

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/annict/annict/go/internal/model"
)

// CreateTestCharacterはテスト用のキャラクターをデータベースに作成し、IDを返す。
func CreateTestCharacter(t *testing.T, tx *sql.Tx) int64 {
	t.Helper()
	return createTestRow(t, tx, "キャラクター",
		`INSERT INTO characters (name, created_at, updated_at) VALUES ($1, $2, $2) RETURNING id`,
		"テストキャラクター-"+uuid.New().String(), time.Now(),
	)
}

// CreateTestPersonはテスト用の人物をデータベースに作成し、IDを返す。
// 名前は一意制約の対象のため、並列のテストで衝突しないようUUIDから作る。
func CreateTestPerson(t *testing.T, tx *sql.Tx) int64 {
	t.Helper()
	return createTestRow(t, tx, "人物",
		`INSERT INTO people (name, created_at, updated_at) VALUES ($1, $2, $2) RETURNING id`,
		"テスト人物-"+uuid.New().String(), time.Now(),
	)
}

// CreateTestOrganizationはテスト用の団体をデータベースに作成し、IDを返す。
// 名前は一意制約の対象のため、並列のテストで衝突しないようUUIDから作る。
func CreateTestOrganization(t *testing.T, tx *sql.Tx) int64 {
	t.Helper()
	return createTestRow(t, tx, "団体",
		`INSERT INTO organizations (name, created_at, updated_at) VALUES ($1, $2, $2) RETURNING id`,
		"テスト団体-"+uuid.New().String(), time.Now(),
	)
}

// CreateTestCharacterFavoriteはテスト用のキャラクターのお気に入りをデータベースに作成し、IDを返す。
func CreateTestCharacterFavorite(t *testing.T, tx *sql.Tx, userID model.UserID, characterID int64) int64 {
	t.Helper()
	return createTestRow(t, tx, "キャラクターのお気に入り",
		`INSERT INTO character_favorites (user_id, character_id, created_at, updated_at) VALUES ($1, $2, $3, $3) RETURNING id`,
		int64(userID), characterID, time.Now(),
	)
}

// CreateTestPersonFavoriteはテスト用の人物のお気に入りをデータベースに作成し、IDを返す。
func CreateTestPersonFavorite(t *testing.T, tx *sql.Tx, userID model.UserID, personID int64) int64 {
	t.Helper()
	return createTestRow(t, tx, "人物のお気に入り",
		`INSERT INTO person_favorites (user_id, person_id, created_at, updated_at) VALUES ($1, $2, $3, $3) RETURNING id`,
		int64(userID), personID, time.Now(),
	)
}

// CreateTestOrganizationFavoriteはテスト用の団体のお気に入りをデータベースに作成し、IDを返す。
func CreateTestOrganizationFavorite(t *testing.T, tx *sql.Tx, userID model.UserID, organizationID int64) int64 {
	t.Helper()
	return createTestRow(t, tx, "団体のお気に入り",
		`INSERT INTO organization_favorites (user_id, organization_id, created_at, updated_at) VALUES ($1, $2, $3, $3) RETURNING id`,
		int64(userID), organizationID, time.Now(),
	)
}

// createTestRowはIDを返すINSERT文を実行し、作成した行のIDを返す。
func createTestRow(t *testing.T, tx *sql.Tx, name, insertSQL string, args ...any) int64 {
	t.Helper()

	var id int64
	if err := tx.QueryRow(insertSQL, args...).Scan(&id); err != nil {
		t.Fatalf("%sデータの作成に失敗しました: %v", name, err)
	}
	return id
}
