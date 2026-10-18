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

// newGetDBPeopleUsecaseはテスト用トランザクション上にUseCaseを組み立てる。本
// UseCaseは読み取りのみで自らトランザクションを開かないためSetupTxを使う。
func newGetDBPeopleUsecase(t *testing.T) (*GetDBPeopleUsecase, *sql.Tx) {
	t.Helper()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	return NewGetDBPeopleUsecase(repository.NewPersonRepository(queries)), tx
}

// TestGetDBPeopleUsecase_Executeは、公開中とアーカイブの人物を返し、掲載停止と削除の人物は
// 返さないことを検証する。
func TestGetDBPeopleUsecase_Execute(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBPeopleUsecase(t)

	publishedID := testutil.NewPersonBuilder(t, tx).Build()
	archivedID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusArchived).Build()
	delistedID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDelisted).Build()
	deletedID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDeleted).Build()

	output, err := uc.Execute(context.Background(), GetDBPeopleInput{
		Page:    1,
		PerPage: 100,
	})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	ids := make(map[model.PersonID]bool, len(output.People))
	for _, p := range output.People {
		ids[p.ID] = true
	}
	if !ids[publishedID] {
		t.Error("公開中の人物が含まれていません")
	}
	if !ids[archivedID] {
		t.Error("アーカイブの人物が含まれていません")
	}
	if ids[delistedID] {
		t.Error("掲載停止の人物が含まれてはいけません")
	}
	if ids[deletedID] {
		t.Error("削除済みの人物が含まれてはいけません")
	}
	if output.TotalCount < int64(len(output.People)) {
		t.Errorf("TotalCount = %d が1ページ分の件数 %d より少ない", output.TotalCount, len(output.People))
	}
}

// TestGetDBPeopleUsecase_Execute_Delistedは、掲載停止の人物を管理者にだけ返し、編集者・一般
// ユーザー・未ログインの閲覧者には返さないことを検証する。削除済みの人物は管理者にも返さない。
func TestGetDBPeopleUsecase_Execute_Delisted(t *testing.T) {
	t.Parallel()

	uc, tx := newGetDBPeopleUsecase(t)
	delistedID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDelisted).Build()
	deletedID := testutil.NewPersonBuilder(t, tx).WithStatus(model.PersonStatusDeleted).Build()

	tests := []struct {
		name         string
		user         *model.User
		wantDelisted bool
	}{
		{name: "管理者", user: &model.User{ID: 1, Role: model.RoleAdmin}, wantDelisted: true},
		{name: "編集者", user: &model.User{ID: 1, Role: model.RoleEditor}},
		{name: "一般ユーザー", user: &model.User{ID: 1, Role: model.RoleUser}},
		{name: "未ログイン", user: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := uc.Execute(context.Background(), GetDBPeopleInput{
				User:    tt.user,
				Page:    1,
				PerPage: 100,
			})
			if err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}

			ids := make(map[model.PersonID]bool, len(output.People))
			for _, p := range output.People {
				ids[p.ID] = true
			}
			if ids[delistedID] != tt.wantDelisted {
				t.Errorf("掲載停止の人物を含む = %v、期待値 = %v", ids[delistedID], tt.wantDelisted)
			}
			if ids[deletedID] {
				t.Error("削除済みの人物が含まれてはいけません")
			}
		})
	}
}
