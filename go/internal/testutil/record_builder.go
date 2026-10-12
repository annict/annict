package testutil

import (
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
)

// EpisodeRecordBuilderはエピソードの記録のテストデータビルダー。
// 記録 (records) とエピソードの記録 (episode_records) を組で作成する。
type EpisodeRecordBuilder struct {
	t                       *testing.T
	tx                      *sql.Tx
	userID                  model.UserID
	workID                  model.WorkID
	episodeID               model.EpisodeID
	body                    *string
	multipleEpisodeRecordID *int64
	deletedAt               *time.Time
}

// EpisodeRecordResultはEpisodeRecordBuilderで作成した記録のID
type EpisodeRecordResult struct {
	RecordID        int64
	EpisodeRecordID int64
}

// NewEpisodeRecordBuilderは新しいEpisodeRecordBuilderを作成
func NewEpisodeRecordBuilder(t *testing.T, tx *sql.Tx, userID model.UserID, workID model.WorkID, episodeID model.EpisodeID) *EpisodeRecordBuilder {
	return &EpisodeRecordBuilder{
		t:         t,
		tx:        tx,
		userID:    userID,
		workID:    workID,
		episodeID: episodeID,
	}
}

// WithBodyは記録の本文を設定
func (b *EpisodeRecordBuilder) WithBody(body string) *EpisodeRecordBuilder {
	b.body = &body
	return b
}

// WithMultipleEpisodeRecordIDは記録をまとめる複数エピソードの記録を設定
func (b *EpisodeRecordBuilder) WithMultipleEpisodeRecordID(id int64) *EpisodeRecordBuilder {
	b.multipleEpisodeRecordID = &id
	return b
}

// WithDeletedAtは記録とエピソードの記録の論理削除日時を設定
func (b *EpisodeRecordBuilder) WithDeletedAt(deletedAt time.Time) *EpisodeRecordBuilder {
	b.deletedAt = &deletedAt
	return b
}

// Buildはテスト用の記録とエピソードの記録をデータベースに作成し、IDを返す
func (b *EpisodeRecordBuilder) Build() EpisodeRecordResult {
	b.t.Helper()

	recordID := createTestRecord(b.t, b.tx, b.userID, b.workID, b.deletedAt)

	var id int64
	err := b.tx.QueryRow(
		`INSERT INTO episode_records (
			user_id, work_id, episode_id, record_id, body, multiple_episode_record_id,
			deleted_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8) RETURNING id`,
		int64(b.userID),
		int64(b.workID),
		int64(b.episodeID),
		recordID,
		b.body,
		b.multipleEpisodeRecordID,
		b.deletedAt,
		time.Now(),
	).Scan(&id)
	if err != nil {
		b.t.Fatalf("エピソードの記録データの作成に失敗しました: %v", err)
	}

	return EpisodeRecordResult{RecordID: recordID, EpisodeRecordID: id}
}

// WorkRecordBuilderは作品の記録のテストデータビルダー。
// 記録 (records) と作品の記録 (work_records) を組で作成する。
type WorkRecordBuilder struct {
	t         *testing.T
	tx        *sql.Tx
	userID    model.UserID
	workID    model.WorkID
	body      string
	deletedAt *time.Time
}

// WorkRecordResultはWorkRecordBuilderで作成した記録のID
type WorkRecordResult struct {
	RecordID     int64
	WorkRecordID int64
}

// NewWorkRecordBuilderは新しいWorkRecordBuilderを作成
func NewWorkRecordBuilder(t *testing.T, tx *sql.Tx, userID model.UserID, workID model.WorkID) *WorkRecordBuilder {
	return &WorkRecordBuilder{
		t:      t,
		tx:     tx,
		userID: userID,
		workID: workID,
	}
}

// WithBodyは記録の本文を設定
func (b *WorkRecordBuilder) WithBody(body string) *WorkRecordBuilder {
	b.body = body
	return b
}

// WithDeletedAtは記録と作品の記録の論理削除日時を設定
func (b *WorkRecordBuilder) WithDeletedAt(deletedAt time.Time) *WorkRecordBuilder {
	b.deletedAt = &deletedAt
	return b
}

// Buildはテスト用の記録と作品の記録をデータベースに作成し、IDを返す
func (b *WorkRecordBuilder) Build() WorkRecordResult {
	b.t.Helper()

	recordID := createTestRecord(b.t, b.tx, b.userID, b.workID, b.deletedAt)

	var id int64
	err := b.tx.QueryRow(
		`INSERT INTO work_records (
			user_id, work_id, record_id, body, deleted_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $6) RETURNING id`,
		int64(b.userID),
		int64(b.workID),
		recordID,
		b.body,
		b.deletedAt,
		time.Now(),
	).Scan(&id)
	if err != nil {
		b.t.Fatalf("作品の記録データの作成に失敗しました: %v", err)
	}

	return WorkRecordResult{RecordID: recordID, WorkRecordID: id}
}

// createTestRecordはテスト用の記録をデータベースに作成し、IDを返す
func createTestRecord(t *testing.T, tx *sql.Tx, userID model.UserID, workID model.WorkID, deletedAt *time.Time) int64 {
	t.Helper()

	var id int64
	now := time.Now()
	err := tx.QueryRow(
		`INSERT INTO records (user_id, work_id, watched_at, deleted_at, created_at, updated_at) VALUES ($1, $2, $3, $4, $3, $3) RETURNING id`,
		int64(userID),
		int64(workID),
		now,
		deletedAt,
	).Scan(&id)
	if err != nil {
		t.Fatalf("記録データの作成に失敗しました: %v", err)
	}

	return id
}

// CreateTestMultipleEpisodeRecordはテスト用の複数エピソードの記録をデータベースに作成し、IDを返す
func CreateTestMultipleEpisodeRecord(t *testing.T, tx *sql.Tx, userID model.UserID, workID model.WorkID) int64 {
	t.Helper()

	var id int64
	err := tx.QueryRow(
		`INSERT INTO multiple_episode_records (user_id, work_id, created_at, updated_at) VALUES ($1, $2, $3, $3) RETURNING id`,
		int64(userID),
		int64(workID),
		time.Now(),
	).Scan(&id)
	if err != nil {
		t.Fatalf("複数エピソードの記録データの作成に失敗しました: %v", err)
	}

	return id
}

// CreateTestCommentはテスト用のエピソードの記録へのコメントをデータベースに作成し、IDを返す。
// userIDはコメントを書いたユーザー。
func CreateTestComment(t *testing.T, tx *sql.Tx, userID model.UserID, episodeRecordID int64) int64 {
	t.Helper()

	var id int64
	err := tx.QueryRow(
		`INSERT INTO comments (user_id, episode_record_id, body, created_at, updated_at) VALUES ($1, $2, 'コメント', $3, $3) RETURNING id`,
		int64(userID),
		episodeRecordID,
		time.Now(),
	).Scan(&id)
	if err != nil {
		t.Fatalf("コメントデータの作成に失敗しました: %v", err)
	}

	return id
}
