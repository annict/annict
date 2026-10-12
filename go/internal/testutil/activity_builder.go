package testutil

import (
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
)

// ActivityBuilderはアクティビティのテストデータビルダー。
// アクティビティグループを指定しない場合は、アクティビティごとにグループを作成する。
type ActivityBuilder struct {
	t               *testing.T
	tx              *sql.Tx
	userID          model.UserID
	workID          model.WorkID
	activityGroupID *int64
	statusID        *int64
	episodeRecordID *int64
}

// NewActivityBuilderは新しいActivityBuilderを作成
func NewActivityBuilder(t *testing.T, tx *sql.Tx, userID model.UserID, workID model.WorkID) *ActivityBuilder {
	return &ActivityBuilder{
		t:      t,
		tx:     tx,
		userID: userID,
		workID: workID,
	}
}

// WithActivityGroupIDはアクティビティグループを設定
func (b *ActivityBuilder) WithActivityGroupID(activityGroupID int64) *ActivityBuilder {
	b.activityGroupID = &activityGroupID
	return b
}

// WithStatusIDはアクティビティの対象となるステータスを設定
func (b *ActivityBuilder) WithStatusID(statusID int64) *ActivityBuilder {
	b.statusID = &statusID
	return b
}

// WithEpisodeRecordIDはアクティビティの対象となるエピソードの記録を設定
func (b *ActivityBuilder) WithEpisodeRecordID(episodeRecordID int64) *ActivityBuilder {
	b.episodeRecordID = &episodeRecordID
	return b
}

// Buildはテスト用のアクティビティデータをデータベースに作成し、IDを返す
func (b *ActivityBuilder) Build() int64 {
	b.t.Helper()

	activityGroupID := b.activityGroupID
	if activityGroupID == nil {
		id := CreateTestActivityGroup(b.t, b.tx, b.userID)
		activityGroupID = &id
	}

	// ステータス・エピソードの記録の指定が無い場合は、それらを持たない作品のアクティビティとして作る。
	trackableType := "Work"
	trackableID := int64(b.workID)
	if b.statusID != nil {
		trackableType = "Status"
		trackableID = *b.statusID
	}
	if b.episodeRecordID != nil {
		trackableType = "EpisodeRecord"
		trackableID = *b.episodeRecordID
	}

	var id int64
	err := b.tx.QueryRow(
		`INSERT INTO activities (
			user_id, trackable_id, trackable_type, action, work_id, status_id,
			episode_record_id, activity_group_id, created_at, updated_at
		) VALUES ($1, $2, $3, 'create_status', $4, $5, $6, $7, $8, $8) RETURNING id`,
		int64(b.userID),
		trackableID,
		trackableType,
		int64(b.workID),
		b.statusID,
		b.episodeRecordID,
		*activityGroupID,
		time.Now(),
	).Scan(&id)
	if err != nil {
		b.t.Fatalf("アクティビティデータの作成に失敗しました: %v", err)
	}

	return id
}

// CreateTestActivityGroupはテスト用のアクティビティグループをデータベースに作成し、IDを返す
func CreateTestActivityGroup(t *testing.T, tx *sql.Tx, userID model.UserID) int64 {
	t.Helper()

	var id int64
	err := tx.QueryRow(
		`INSERT INTO activity_groups (user_id, itemable_type, created_at, updated_at) VALUES ($1, 'Status', $2, $2) RETURNING id`,
		int64(userID),
		time.Now(),
	).Scan(&id)
	if err != nil {
		t.Fatalf("アクティビティグループデータの作成に失敗しました: %v", err)
	}

	return id
}
