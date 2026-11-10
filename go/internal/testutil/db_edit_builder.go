package testutil

import (
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
)

// DBEditBuilderはAnnict DBの編集履歴の操作 (db_edits) のテストデータのビルダー。
// 既定では、編集者を持たないシステムの書き込みの更新 (actorがsystem、actionがupdate) を作る。
type DBEditBuilder struct {
	t         *testing.T
	tx        *sql.Tx
	actor     model.DBEditActor
	userID    *model.UserID
	action    model.DBEditAction
	createdAt *time.Time
}

// NewDBEditBuilderは新しいDBEditBuilderを作成する。
func NewDBEditBuilder(t *testing.T, tx *sql.Tx) *DBEditBuilder {
	return &DBEditBuilder{
		t:      t,
		tx:     tx,
		actor:  model.DBEditActorSystem,
		action: model.DBEditActionUpdate,
	}
}

// WithUserは操作を編集者userIDの編集にする (actorをuserにする)。
func (b *DBEditBuilder) WithUser(userID model.UserID) *DBEditBuilder {
	b.actor = model.DBEditActorUser
	b.userID = &userID
	return b
}

// WithActorは主体を設定する。退会したユーザーの編集 (actorがuserで編集者が空) を作るときに使う。
func (b *DBEditBuilder) WithActor(actor model.DBEditActor) *DBEditBuilder {
	b.actor = actor
	return b
}

// WithActionは操作の種類を設定する。
func (b *DBEditBuilder) WithAction(action model.DBEditAction) *DBEditBuilder {
	b.action = action
	return b
}

// WithCreatedAtは作成日時を設定する。
func (b *DBEditBuilder) WithCreatedAt(createdAt time.Time) *DBEditBuilder {
	b.createdAt = &createdAt
	return b
}

// Buildは操作をデータベースに作成し、IDを返す。
func (b *DBEditBuilder) Build() model.DBEditID {
	b.t.Helper()

	var userID any
	if b.userID != nil {
		userID = int64(*b.userID)
	}
	createdAt := time.Now()
	if b.createdAt != nil {
		createdAt = *b.createdAt
	}

	id := createTestRow(b.t, b.tx, "Annict DBの編集履歴の操作",
		`INSERT INTO db_edits (actor, user_id, action, created_at) VALUES ($1, $2, $3, $4) RETURNING id`,
		string(b.actor), userID, string(b.action), createdAt,
	)
	return model.DBEditID(id)
}
