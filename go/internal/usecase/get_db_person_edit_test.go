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

// newGetDBPersonEditUsecaseはテスト用トランザクション上にUseCaseを組み立てる。本
// UseCaseは読み取りのみで自らトランザクションを開かないためSetupTxを使う。
func newGetDBPersonEditUsecase(t *testing.T) (*GetDBPersonEditUsecase, *sql.Tx) {
	t.Helper()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	return NewGetDBPersonEditUsecase(repository.NewPersonRepository(queries), repository.NewPrefectureRepository(queries)), tx
}

// TestGetDBPersonEditUsecase_Execute_ReturnsPersonAndPrefecturesは、フォームの初期値に
// なる人物と出身地の選択肢を返し、アーカイブの人物も編集できることを検証する。
func TestGetDBPersonEditUsecase_Execute_ReturnsPersonAndPrefectures(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBPersonEditUsecase(t)

	prefectureID := testutil.CreateTestPrefecture(t, tx, "福岡県")
	name := "編集フォームの人物-" + t.Name()
	personID := testutil.NewPersonBuilder(t, tx).WithName(name).WithStatus(model.PersonStatusArchived).Build()

	output, err := uc.Execute(context.Background(), GetDBPersonEditInput{
		User:     &model.User{ID: 1, Role: model.RoleEditor},
		PersonID: personID,
	})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if output.Person == nil || output.Person.ID != personID || output.Person.Name != name {
		t.Errorf("Person = %+v、期待値 = ID %d の %q", output.Person, personID, name)
	}

	found := false
	for _, prefecture := range output.Prefectures {
		if prefecture.ID == prefectureID {
			found = true
		}
	}
	if !found {
		t.Errorf("Prefecturesに出身地 %d が含まれていません", prefectureID)
	}
}

// TestGetDBPersonEditUsecase_Execute_NotFoundは、存在しない人物と掲載停止・削除済みの人物が
// not foundのAppErrorとして報告されることを検証する (Handlerはこれを404として描画する)。
func TestGetDBPersonEditUsecase_Execute_NotFound(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBPersonEditUsecase(t)

	delistedID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDelisted).Build()
	deletedID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDeleted).Build()

	tests := []struct {
		name     string
		personID model.PersonID
	}{
		{name: "存在しない人物", personID: model.PersonID(999999999)},
		{name: "掲載停止の人物", personID: delistedID},
		{name: "削除済みの人物", personID: deletedID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := uc.Execute(context.Background(), GetDBPersonEditInput{
				User:     &model.User{ID: 1, Role: model.RoleEditor},
				PersonID: tt.personID,
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

// TestGetDBPersonEditUsecase_Execute_RequiresCommitterは、ユーザーを伴わない呼び出しと
// 一般ユーザーの呼び出しを、人物を読む前に拒否することを検証する。
func TestGetDBPersonEditUsecase_Execute_RequiresCommitter(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBPersonEditUsecase(t)
	personID := testutil.NewPersonBuilder(t, tx).Build()

	tests := []struct {
		name string
		user *model.User
	}{
		{name: "ユーザー無し", user: nil},
		{name: "一般ユーザー", user: &model.User{ID: 1, Role: model.RoleUser}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.Execute(context.Background(), GetDBPersonEditInput{User: tt.user, PersonID: personID})
			ae := model.AsAppError(err)
			if ae == nil || ae.Code != model.AppErrCodeForbidden {
				t.Errorf("Execute()のエラー = %v、期待値 = AppErrCodeForbidden", err)
			}
		})
	}
}
