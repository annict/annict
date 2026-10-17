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

// newGetDBPersonDeletionNewUsecaseはテスト用トランザクション上にUseCaseを組み立てる。本UseCaseは
// 読み取りのみで自らトランザクションを開かないためSetupTxを使う。
func newGetDBPersonDeletionNewUsecase(t *testing.T) (*GetDBPersonDeletionNewUsecase, *sql.Tx) {
	t.Helper()

	db, tx := testutil.SetupTx(t)
	return NewGetDBPersonDeletionNewUsecase(repository.NewPersonRepository(query.New(db).WithTx(tx))), tx
}

// TestGetDBPersonDeletionNewUsecase_Executeは、削除できる状態の人物を、確認画面が名指しする名前とともに
// 返すことを検証する。
func TestGetDBPersonDeletionNewUsecase_Execute(t *testing.T) {
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

			uc, tx := newGetDBPersonDeletionNewUsecase(t)
			name := "確認画面の人物-" + t.Name()
			personID := tt.prepare(testutil.NewPersonBuilder(t, tx).WithName(name)).Build()

			output, err := uc.Execute(context.Background(), GetDBPersonDeletionNewInput{
				User:     &model.User{ID: 1, Role: model.RoleAdmin},
				PersonID: personID,
			})
			if err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}
			if output.Person == nil || output.Person.ID != personID || output.Person.Name != name {
				t.Errorf("Person = %+v、期待値 = ID %d の %q", output.Person, personID, name)
			}
		})
	}
}

// TestGetDBPersonDeletionNewUsecase_Execute_NotFoundは、存在しない人物と削除できない状態の人物を、not foundの
// AppErrorとして報告することを検証する。古いリンクが、操作できない人物の確認画面ではなく404に
// なるようにするため。
func TestGetDBPersonDeletionNewUsecase_Execute_NotFound(t *testing.T) {
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

			uc, tx := newGetDBPersonDeletionNewUsecase(t)
			personID := model.PersonID(999999999)
			if tt.prepare != nil {
				personID = tt.prepare(testutil.NewPersonBuilder(t, tx)).Build()
			}

			output, err := uc.Execute(context.Background(), GetDBPersonDeletionNewInput{
				User:     &model.User{ID: 1, Role: model.RoleAdmin},
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

// TestGetDBPersonDeletionNewUsecase_Execute_Forbiddenは、操作する権限の無い呼び出しを、人物を読む前に拒否する
// ことを検証する。
func TestGetDBPersonDeletionNewUsecase_Execute_Forbidden(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBPersonDeletionNewUsecase(t)
	personID := testutil.NewPersonBuilder(t, tx).Build()

	tests := []struct {
		name string
		user *model.User
	}{
		{name: "ユーザー無し", user: nil},
		{name: "一般ユーザー", user: &model.User{ID: 1, Role: model.RoleUser}},
		{name: "編集者", user: &model.User{ID: 1, Role: model.RoleEditor}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.Execute(context.Background(), GetDBPersonDeletionNewInput{User: tt.user, PersonID: personID})
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeForbidden {
				t.Errorf("Execute()のエラー = %v、期待値 = AppErrCodeForbidden", err)
			}
		})
	}
}
