package testutil

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/annict/annict/go/internal/model"
)

// CreateTestWorkCommentはテスト用の作品へのコメントをデータベースに作成し、IDを返す。
func CreateTestWorkComment(t *testing.T, tx *sql.Tx, userID model.UserID, workID model.WorkID) int64 {
	t.Helper()
	return createTestRow(t, tx, "作品へのコメント",
		`INSERT INTO work_comments (user_id, work_id, body, created_at, updated_at) VALUES ($1, $2, 'コメント', $3, $3) RETURNING id`,
		int64(userID), int64(workID), time.Now(),
	)
}

// CreateTestUserlandProjectMemberはテスト用のUserlandプロジェクトへの参加を、プロジェクトと
// カテゴリと一緒にデータベースに作成し、IDを返す。
func CreateTestUserlandProjectMember(t *testing.T, tx *sql.Tx, userID model.UserID) int64 {
	t.Helper()

	now := time.Now()
	categoryID := createTestRow(t, tx, "Userlandのカテゴリ",
		`INSERT INTO userland_categories (name, name_en, created_at, updated_at) VALUES ('カテゴリ', 'Category', $1, $1) RETURNING id`,
		now,
	)
	projectID := createTestRow(t, tx, "Userlandプロジェクト",
		`INSERT INTO userland_projects (userland_category_id, name, summary, description, url, created_at, updated_at)
		VALUES ($1, 'プロジェクト', '', '', 'https://example.com', $2, $2) RETURNING id`,
		categoryID, now,
	)
	return createTestRow(t, tx, "Userlandプロジェクトへの参加",
		`INSERT INTO userland_project_members (user_id, userland_project_id, created_at, updated_at) VALUES ($1, $2, $3, $3) RETURNING id`,
		int64(userID), projectID, now,
	)
}

// CreateTestEmailConfirmationはテスト用のメールアドレスの確認をデータベースに作成し、IDを返す。
// トークンは一意制約の対象のため、並列のテストで衝突しないようUUIDから作る。
func CreateTestEmailConfirmation(t *testing.T, tx *sql.Tx, userID model.UserID) int64 {
	t.Helper()
	return createTestRow(t, tx, "メールアドレスの確認",
		`INSERT INTO email_confirmations (user_id, email, event, token, expires_at, created_at, updated_at)
		VALUES ($1, 'confirmation@example.com', 'update_email', $2, $3, $3, $3) RETURNING id`,
		int64(userID), uuid.New().String(), time.Now(),
	)
}
