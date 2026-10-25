package testutil

import (
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
)

// Statusのkind (Rails enumの値)
const (
	StatusKindWannaWatch   int32 = 1
	StatusKindWatching     int32 = 2
	StatusKindWatched      int32 = 3
	StatusKindStopWatching int32 = 4
	StatusKindOnHold       int32 = 5
)

// StatusBuilderはステータスのテストデータビルダー
type StatusBuilder struct {
	t      *testing.T
	tx     *sql.Tx
	userID model.UserID
	workID model.WorkID
	kind   int32
}

// NewStatusBuilderは新しいStatusBuilderを作成
func NewStatusBuilder(t *testing.T, tx *sql.Tx, userID model.UserID, workID model.WorkID) *StatusBuilder {
	return &StatusBuilder{
		t:      t,
		tx:     tx,
		userID: userID,
		workID: workID,
		kind:   StatusKindWatching,
	}
}

// WithKindはステータスの種類 (StatusKind*) を設定
func (b *StatusBuilder) WithKind(kind int32) *StatusBuilder {
	b.kind = kind
	return b
}

// Buildはテスト用のステータスデータをデータベースに作成し、IDを返す
func (b *StatusBuilder) Build() int64 {
	b.t.Helper()

	var id int64
	err := b.tx.QueryRow(
		`INSERT INTO statuses (user_id, work_id, kind, created_at, updated_at) VALUES ($1, $2, $3, $4, $4) RETURNING id`,
		int64(b.userID),
		int64(b.workID),
		b.kind,
		time.Now(),
	).Scan(&id)
	if err != nil {
		b.t.Fatalf("ステータスデータの作成に失敗しました: %v", err)
	}

	return id
}
