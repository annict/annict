package usecase

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// newUnarchivePersonUsecaseは共有テストDBに対してUseCaseを組み立てる。本UseCaseは編集履歴と同じトランザクションで
// 書くため自前のトランザクションを開く。前提の人物はtestutil.CreateCommittedPersonでコミットする。
func newUnarchivePersonUsecase(t *testing.T) (*UnarchivePersonUsecase, *sql.DB) {
	t.Helper()

	db := testutil.GetTestDB()
	queries := query.New(db)
	return NewUnarchivePersonUsecase(db, repository.NewPersonRepository(queries), newTestDBEditRecordRepos(queries)), db
}

// TestUnarchivePersonUsecase_Executeは、非公開の人物を公開に戻し、unpublished_atをNULLに戻すことを検証する。
func TestUnarchivePersonUsecase_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prepare func(b *testutil.PersonBuilder) *testutil.PersonBuilder
	}{
		{name: "非公開の人物", prepare: func(b *testutil.PersonBuilder) *testutil.PersonBuilder {
			return b.WithStatus(model.PersonStatusArchived).WithUnpublishedAt(time.Now())
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc, db := newUnarchivePersonUsecase(t)
			personID := testutil.CreateCommittedPerson(t, tt.prepare)

			output, err := uc.Execute(context.Background(), UnarchivePersonInput{
				User:     committedEditorForTest(t, model.RoleEditor),
				PersonID: personID,
			})
			if err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}
			if output.PersonID != personID {
				t.Errorf("PersonID = %d、期待値 = %d", output.PersonID, personID)
			}

			var published bool
			if err := db.QueryRow(`SELECT status = 'published' AND unpublished_at IS NULL FROM people WHERE id = $1`, int64(personID)).Scan(&published); err != nil {
				t.Fatalf("人物の取得に失敗: %v", err)
			}
			if !published {
				t.Errorf("statusがpublishedに戻っていないか、unpublished_atが残っています")
			}
		})
	}
}

// TestUnarchivePersonUsecase_Execute_NotFoundは、存在しない人物とアーカイブ解除できない状態の人物を、not foundの
// AppErrorとして報告することを検証する (Handlerはこれを404として描画する)。
func TestUnarchivePersonUsecase_Execute_NotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prepare func(b *testutil.PersonBuilder) *testutil.PersonBuilder
	}{
		{name: "存在しない人物"},
		{name: "公開中の人物", prepare: func(b *testutil.PersonBuilder) *testutil.PersonBuilder { return b }},
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

			uc, _ := newUnarchivePersonUsecase(t)
			personID := model.PersonID(999999999)
			if tt.prepare != nil {
				personID = testutil.CreateCommittedPerson(t, tt.prepare)
			}

			output, err := uc.Execute(context.Background(), UnarchivePersonInput{
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

// TestUnarchivePersonUsecase_Execute_RequiresCommitterは、committerでない呼び出しを、人物を書き換える前に
// 拒否することを検証する。
func TestUnarchivePersonUsecase_Execute_RequiresCommitter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		user *model.User
	}{
		{name: "ユーザー無し", user: nil},
		{name: "一般ユーザー", user: committedEditorForTest(t, model.RoleUser)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc, db := newUnarchivePersonUsecase(t)
			personID := testutil.CreateCommittedPerson(t, func(b *testutil.PersonBuilder) *testutil.PersonBuilder {
				return b.WithStatus(model.PersonStatusArchived)
			})

			_, err := uc.Execute(context.Background(), UnarchivePersonInput{User: tt.user, PersonID: personID})
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeForbidden {
				t.Errorf("Execute()のエラー = %v、期待値 = AppErrCodeForbidden", err)
			}

			var unchanged bool
			if err := db.QueryRow(`SELECT status = 'archived' FROM people WHERE id = $1`, int64(personID)).Scan(&unchanged); err != nil {
				t.Fatalf("人物の取得に失敗: %v", err)
			}
			if !unchanged {
				t.Error("拒否した呼び出しが人物の状態を書き換えました")
			}
		})
	}
}
