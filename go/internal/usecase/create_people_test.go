package usecase

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

// newCreatePeopleUsecaseは共有テストDBに対して一括登録UseCaseを組み立てる。本UseCaseは
// 内部で自前のトランザクションを開くため、テストはSetupTxではなくGetTestDBを使う。
func newCreatePeopleUsecase(db *sql.DB) *CreatePeopleUsecase {
	personRepo := repository.NewPersonRepository(query.New(db))
	return NewCreatePeopleUsecase(db, personRepo, validator.NewDBPersonCreateValidator(personRepo))
}

// createPeopleTestNameは他のテストがコミットした人物と重ならない名前を返す。
func createPeopleTestName(label string) string {
	return label + "-" + uuid.New().String()
}

// cleanupCreatedPeopleは送信がコミットした人物とその編集履歴をテスト終了時に削除する。
// 送信者のユーザーの後始末より先に走るよう、ユーザーを挿入した後に呼ぶ。
func cleanupCreatedPeople(t *testing.T, db *sql.DB, names ...string) {
	t.Helper()

	t.Cleanup(func() {
		if _, err := db.Exec(`
			DELETE FROM db_activities
			WHERE trackable_type = 'Person' AND trackable_id IN (SELECT id FROM people WHERE name = ANY($1))
		`, pq.Array(names)); err != nil {
			t.Errorf("人物の編集履歴の後始末に失敗: %v", err)
		}
		if _, err := db.Exec(`DELETE FROM people WHERE name = ANY($1)`, pq.Array(names)); err != nil {
			t.Errorf("人物の後始末に失敗: %v", err)
		}
	})
}

func countCreatedPeople(t *testing.T, db *sql.DB, names ...string) int {
	t.Helper()

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM people WHERE name = ANY($1)`, pq.Array(names)).Scan(&count); err != nil {
		t.Fatalf("人物件数の取得に失敗: %v", err)
	}
	return count
}

// TestCreatePeopleUsecase_Execute_CreatesPeopleWithActivitiesは、送信された全行の人物を
// 入力順に作成し、人物ごとに送信者へ帰属するpeople.createの編集履歴を記録することを検証する。
func TestCreatePeopleUsecase_Execute_CreatesPeopleWithActivities(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	user := insertCreateActor(t, db, model.RoleEditor)
	first := createPeopleTestName("一括登録の人物1")
	second := createPeopleTestName("一括登録の人物2")
	cleanupCreatedPeople(t, db, first, second)

	output, err := newCreatePeopleUsecase(db).Execute(context.Background(), CreatePeopleInput{
		User: user,
		Rows: first + ",いっかつとうろくのじんぶついち\n" + second,
	})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if len(output.PersonIDs) != 2 {
		t.Fatalf("作成した人物 = %d件、期待値 = 2", len(output.PersonIDs))
	}

	tests := []struct {
		id           model.PersonID
		wantName     string
		wantNameKana string
	}{
		{id: output.PersonIDs[0], wantName: first, wantNameKana: "いっかつとうろくのじんぶついち"},
		{id: output.PersonIDs[1], wantName: second, wantNameKana: ""},
	}
	for _, tt := range tests {
		var name, nameKana string
		if err := db.QueryRow(`SELECT name, name_kana FROM people WHERE id = $1`, int64(tt.id)).Scan(&name, &nameKana); err != nil {
			t.Fatalf("人物 %d の取得に失敗: %v", int64(tt.id), err)
		}
		if name != tt.wantName || nameKana != tt.wantNameKana {
			t.Errorf("人物 %d = %q / %q、期待値 = %q / %q", int64(tt.id), name, nameKana, tt.wantName, tt.wantNameKana)
		}

		var count int
		if err := db.QueryRow(`
			SELECT COUNT(*) FROM db_activities
			WHERE trackable_type = 'Person' AND trackable_id = $1 AND action = 'people.create' AND user_id = $2
		`, int64(tt.id), int64(user.ID)).Scan(&count); err != nil {
			t.Fatalf("編集履歴の取得に失敗: %v", err)
		}
		if count != 1 {
			t.Errorf("人物 %d のpeople.createの編集履歴 = %d件、期待値 = 1", int64(tt.id), count)
		}
	}
}

// TestCreatePeopleUsecase_Execute_ValidationErrorは、1行でも不正な送信では1件も作成
// しないことを検証する。
func TestCreatePeopleUsecase_Execute_ValidationError(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	valid := createPeopleTestName("正しい行の人物")
	cleanupCreatedPeople(t, db, valid)

	_, err := newCreatePeopleUsecase(db).Execute(context.Background(), CreatePeopleInput{
		User: unsavedCreateActor(),
		Rows: valid + "\n,よみだけ",
	})
	if ve := model.AsValidationError(err); ve == nil {
		t.Fatalf("*model.ValidationErrorを期待したが得られなかった: %v", err)
	}

	if count := countCreatedPeople(t, db, valid); count != 0 {
		t.Errorf("作成された人物 = %d件、期待値 = 0", count)
	}
}

// TestCreatePeopleUsecase_Execute_RequiresCommitterは、committerでない呼び出しを何も
// 書かずに拒否することを検証する。
func TestCreatePeopleUsecase_Execute_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newCreatePeopleUsecase(db)

	tests := []struct {
		name string
		user *model.User
	}{
		{name: "未ログイン", user: nil},
		{name: "一般ユーザー", user: &model.User{ID: 1, Role: model.RoleUser}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			name := createPeopleTestName("認可テストの人物")
			cleanupCreatedPeople(t, db, name)

			_, err := uc.Execute(context.Background(), CreatePeopleInput{User: tt.user, Rows: name})

			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeForbidden {
				t.Fatalf("エラー = %v、期待値 = AppErrCodeForbidden", err)
			}
			if count := countCreatedPeople(t, db, name); count != 0 {
				t.Errorf("作成された人物 = %d件、期待値 = 0", count)
			}
		})
	}
}
