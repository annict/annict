package testutil

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/annict/annict/go/internal/model"
)

// CreateTestForumPostはテスト用のフォーラムのスレッドを、カテゴリと一緒にデータベースに作成し、IDを返す。
// カテゴリのスラッグは一意制約の対象のため、並列のテストで衝突しないようUUIDから作る。
func CreateTestForumPost(t *testing.T, tx *sql.Tx, userID model.UserID) int64 {
	t.Helper()

	now := time.Now()
	categoryID := createTestRow(t, tx, "フォーラムのカテゴリ",
		`INSERT INTO forum_categories (
			slug, name, name_en, description, description_en, postable_role, sort_number, created_at, updated_at
		) VALUES ($1, 'カテゴリ', 'Category', '', '', 'user', 0, $2, $2) RETURNING id`,
		"test-"+uuid.New().String(), now,
	)
	return createTestRow(t, tx, "フォーラムのスレッド",
		`INSERT INTO forum_posts (user_id, forum_category_id, title, last_commented_at, created_at, updated_at)
		VALUES ($1, $2, 'スレッド', $3, $3, $3) RETURNING id`,
		int64(userID), categoryID, now,
	)
}

// CreateTestForumCommentはテスト用のフォーラムのコメントをデータベースに作成し、IDを返す。
func CreateTestForumComment(t *testing.T, tx *sql.Tx, userID model.UserID, forumPostID int64) int64 {
	t.Helper()
	return createTestRow(t, tx, "フォーラムのコメント",
		`INSERT INTO forum_comments (user_id, forum_post_id, body, created_at, updated_at) VALUES ($1, $2, 'コメント', $3, $3) RETURNING id`,
		int64(userID), forumPostID, time.Now(),
	)
}

// CreateTestForumPostParticipantはテスト用のフォーラムのスレッドの参加者をデータベースに作成し、IDを返す。
func CreateTestForumPostParticipant(t *testing.T, tx *sql.Tx, userID model.UserID, forumPostID int64) int64 {
	t.Helper()
	return createTestRow(t, tx, "フォーラムのスレッドの参加者",
		`INSERT INTO forum_post_participants (user_id, forum_post_id, created_at, updated_at) VALUES ($1, $2, $3, $3) RETURNING id`,
		int64(userID), forumPostID, time.Now(),
	)
}
