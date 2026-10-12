package testutil

import (
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
)

// CreateTestFollowはテスト用のフォローをデータベースに作成し、IDを返す。
// userIDがfollowingIDをフォローする。
func CreateTestFollow(t *testing.T, tx *sql.Tx, userID, followingID model.UserID) int64 {
	t.Helper()

	var id int64
	err := tx.QueryRow(
		`INSERT INTO follows (user_id, following_id, created_at, updated_at) VALUES ($1, $2, $3, $3) RETURNING id`,
		int64(userID),
		int64(followingID),
		time.Now(),
	).Scan(&id)
	if err != nil {
		t.Fatalf("フォローデータの作成に失敗しました: %v", err)
	}

	return id
}

// CreateTestMuteUserはテスト用のミュートをデータベースに作成し、IDを返す。
// userIDがmutedUserIDをミュートする。
func CreateTestMuteUser(t *testing.T, tx *sql.Tx, userID, mutedUserID model.UserID) int64 {
	t.Helper()

	var id int64
	err := tx.QueryRow(
		`INSERT INTO mute_users (user_id, muted_user_id, created_at, updated_at) VALUES ($1, $2, $3, $3) RETURNING id`,
		int64(userID),
		int64(mutedUserID),
		time.Now(),
	).Scan(&id)
	if err != nil {
		t.Fatalf("ミュートデータの作成に失敗しました: %v", err)
	}

	return id
}
