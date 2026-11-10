package testutil

import (
	"testing"

	"github.com/annict/annict/go/internal/model"
)

// 以下は、自前のトランザクションを開くUseCaseを通るテストが、前提データを共有テストDBにコミットして作る
// ヘルパー。テスト用のトランザクションに作った行は、UseCaseのトランザクションから見えないため。

// CreateCommittedPersonは、共有テストDBに人物をコミットして作る。configureで属性を設定でき、nilなら既定の
// まま作る。人物と、人物を指す編集履歴 (db_activities) は、テストの終了時に消す。版 (person_revisions) は
// 外部キーで人物と一緒に消える。
func CreateCommittedPerson(t *testing.T, configure func(*PersonBuilder) *PersonBuilder) model.PersonID {
	t.Helper()

	db := GetTestDB()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("人物の作成トランザクションの開始に失敗: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	builder := NewPersonBuilder(t, tx)
	if configure != nil {
		builder = configure(builder)
	}
	id := builder.Build()
	if err := tx.Commit(); err != nil {
		t.Fatalf("人物の作成トランザクションのコミットに失敗: %v", err)
	}

	t.Cleanup(func() {
		if _, err := db.Exec(`DELETE FROM db_activities WHERE trackable_type = 'Person' AND trackable_id = $1`, int64(id)); err != nil {
			t.Errorf("人物の編集履歴の後始末に失敗: %v", err)
		}
		if _, err := db.Exec(`DELETE FROM people WHERE id = $1`, int64(id)); err != nil {
			t.Errorf("人物の後始末に失敗: %v", err)
		}
	})
	return id
}

// CreateCommittedUserは、共有テストDBに指定した役割のユーザーをコミットして作り、テストの終了時に消す。
// 編集履歴の操作がusersを外部キーで指すため、書き込みが成功する経路の編集者に使う。
func CreateCommittedUser(t *testing.T, role int32) *model.User {
	t.Helper()

	db := GetTestDB()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("ユーザーの作成トランザクションの開始に失敗: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	id := NewUserBuilder(t, tx).WithRole(role).Build()
	if err := tx.Commit(); err != nil {
		t.Fatalf("ユーザーの作成トランザクションのコミットに失敗: %v", err)
	}
	t.Cleanup(func() { DeleteUser(t, db, id) })
	return &model.User{ID: id, Role: role}
}
