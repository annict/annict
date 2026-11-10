package usecase

import (
	"context"
	"database/sql"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// newDeletePersonUsecaseは共有テストDBに対してUseCaseを組み立てる。本UseCaseは編集履歴と同じトランザクションで
// 書くため自前のトランザクションを開く。前提の人物はtestutil.CreateCommittedPersonでコミットする。
func newDeletePersonUsecase(t *testing.T) (*DeletePersonUsecase, *sql.DB) {
	t.Helper()

	db := testutil.GetTestDB()
	queries := query.New(db)
	return NewDeletePersonUsecase(db, repository.NewPersonRepository(queries), newTestDBEditRecordRepos(queries)), db
}

// TestDeletePersonUsecase_Executeは、削除できる状態の人物を削除することを検証する。
func TestDeletePersonUsecase_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prepare func(b *testutil.PersonBuilder) *testutil.PersonBuilder
	}{
		{name: "公開中の人物", prepare: func(b *testutil.PersonBuilder) *testutil.PersonBuilder { return b }},
		{name: "アーカイブの人物", prepare: func(b *testutil.PersonBuilder) *testutil.PersonBuilder {
			return b.WithStatus(model.PersonStatusArchived)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc, db := newDeletePersonUsecase(t)
			personID := testutil.CreateCommittedPerson(t, tt.prepare)

			output, err := uc.Execute(context.Background(), DeletePersonInput{
				User:     committedEditorForTest(t, model.RoleAdmin),
				PersonID: personID,
			})
			if err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}
			if output.PersonID != personID {
				t.Errorf("PersonID = %d、期待値 = %d", output.PersonID, personID)
			}

			var deleted bool
			if err := db.QueryRow(`SELECT status = 'deleted' AND deleted_at IS NOT NULL FROM people WHERE id = $1`, int64(personID)).Scan(&deleted); err != nil {
				t.Fatalf("人物の取得に失敗: %v", err)
			}
			if !deleted {
				t.Errorf("statusがdeletedでないか、deleted_atが立っていません")
			}
		})
	}
}

// TestDeletePersonUsecase_Execute_NotFoundは、存在しない人物と削除できない状態の人物を、not foundの
// AppErrorとして報告することを検証する (Handlerはこれを404として描画する)。
func TestDeletePersonUsecase_Execute_NotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prepare func(b *testutil.PersonBuilder) *testutil.PersonBuilder
	}{
		{name: "存在しない人物"},
		{name: "掲載停止の人物", prepare: func(b *testutil.PersonBuilder) *testutil.PersonBuilder {
			return b.WithStatus(model.PersonStatusDelisted)
		}},
		{name: "削除済みの人物", prepare: func(b *testutil.PersonBuilder) *testutil.PersonBuilder {
			return b.WithStatus(model.PersonStatusDeleted)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc, _ := newDeletePersonUsecase(t)
			personID := model.PersonID(999999999)
			if tt.prepare != nil {
				personID = testutil.CreateCommittedPerson(t, tt.prepare)
			}

			output, err := uc.Execute(context.Background(), DeletePersonInput{
				User:     committedEditorForTest(t, model.RoleAdmin),
				PersonID: personID,
			})
			if output != nil {
				t.Errorf("Execute() = %+v、期待値 = nil", output)
			}
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
				t.Errorf("Execute()のエラー = %v、期待値 = AppErrCodeResourceNotFound", err)
			}
		})
	}
}

// TestDeletePersonUsecase_Execute_RequiresAdminは、管理者でない呼び出しを、人物を書き換える前に
// 拒否することを検証する。
func TestDeletePersonUsecase_Execute_RequiresAdmin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		user *model.User
	}{
		{name: "ユーザー無し", user: nil},
		{name: "一般ユーザー", user: committedEditorForTest(t, model.RoleUser)},
		{name: "編集者", user: committedEditorForTest(t, model.RoleEditor)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc, db := newDeletePersonUsecase(t)
			personID := testutil.CreateCommittedPerson(t, nil)

			_, err := uc.Execute(context.Background(), DeletePersonInput{User: tt.user, PersonID: personID})
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeForbidden {
				t.Errorf("Execute()のエラー = %v、期待値 = AppErrCodeForbidden", err)
			}

			var unchanged bool
			if err := db.QueryRow(`SELECT status <> 'deleted' FROM people WHERE id = $1`, int64(personID)).Scan(&unchanged); err != nil {
				t.Fatalf("人物の取得に失敗: %v", err)
			}
			if !unchanged {
				t.Error("拒否した呼び出しが人物の状態を書き換えました")
			}
		})
	}
}
