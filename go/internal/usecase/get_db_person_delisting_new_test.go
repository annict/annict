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

// newGetDBPersonDelistingNewUsecaseはテスト用トランザクション上にUseCaseを組み立てる。本UseCaseは
// 読み取りのみで自らトランザクションを開かないためSetupTxを使う。
func newGetDBPersonDelistingNewUsecase(t *testing.T) (*GetDBPersonDelistingNewUsecase, *sql.Tx) {
	t.Helper()

	db, tx := testutil.SetupTx(t)
	return NewGetDBPersonDelistingNewUsecase(repository.NewPersonRepository(query.New(db).WithTx(tx))), tx
}

// TestGetDBPersonDelistingNewUsecase_Executeは、掲載を停止できる状態の人物を、確認画面が名指しする
// 名前と入力欄の初期値にする対応記録とともに返すことを検証する。
func TestGetDBPersonDelistingNewUsecase_Execute(t *testing.T) {
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

			uc, tx := newGetDBPersonDelistingNewUsecase(t)
			name := "確認画面の人物-" + t.Name()
			personID := tt.prepare(testutil.NewPersonBuilder(t, tx).WithName(name).WithDelistMessage("以前の対応記録")).Build()

			output, err := uc.Execute(context.Background(), GetDBPersonDelistingNewInput{
				User:     &model.User{ID: 1, Role: model.RoleAdmin},
				PersonID: personID,
			})
			if err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}
			if output.Person == nil || output.Person.ID != personID || output.Person.Name != name {
				t.Fatalf("Person = %+v、期待値 = ID %d の %q", output.Person, personID, name)
			}
			if got := output.Person.DelistMessage; got == nil || *got != "以前の対応記録" {
				t.Errorf("DelistMessage = %v、期待値 = 以前の対応記録", got)
			}
		})
	}
}

// TestGetDBPersonDelistingNewUsecase_Execute_NotFoundは、存在しない人物と掲載を停止できない状態の
// 人物を、not foundのAppErrorとして報告することを検証する。
func TestGetDBPersonDelistingNewUsecase_Execute_NotFound(t *testing.T) {
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

			uc, tx := newGetDBPersonDelistingNewUsecase(t)
			personID := model.PersonID(999999999)
			if tt.prepare != nil {
				personID = tt.prepare(testutil.NewPersonBuilder(t, tx)).Build()
			}

			output, err := uc.Execute(context.Background(), GetDBPersonDelistingNewInput{
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

// TestGetDBPersonDelistingNewUsecase_Execute_Forbiddenは、管理者でない呼び出しを、人物を読む前に
// 拒否することを検証する。
func TestGetDBPersonDelistingNewUsecase_Execute_Forbidden(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBPersonDelistingNewUsecase(t)
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
			_, err := uc.Execute(context.Background(), GetDBPersonDelistingNewInput{User: tt.user, PersonID: personID})
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeForbidden {
				t.Errorf("Execute()のエラー = %v、期待値 = AppErrCodeForbidden", err)
			}
		})
	}
}
