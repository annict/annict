package usecase

import (
	"context"
	"slices"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestGetDBPersonRevisionsUsecase_Execute_HidesDelistingは、掲載を再開した人物の履歴で、管理者以外には
// 掲載停止・掲載再開の操作の版と掲載停止中に記録した版を出さず、管理者には出すことを検証する。見せる版の
// 差分は隠した版と比べるため、掲載再開のあとの版に状態の変化は出ない。
func TestGetDBPersonRevisionsUsecase_Execute_HidesDelisting(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	ctx := context.Background()
	queries := query.New(db).WithTx(tx)
	repo := repository.NewPersonRevisionRepository(queries)
	uc := NewGetDBPersonRevisionsUsecase(repository.NewPersonRepository(queries), repo, repository.NewPrefectureRepository(queries))
	personID := testutil.NewPersonBuilder(t, tx).Build()

	snapshot := func(name string, status model.PersonStatus) model.PersonSnapshot {
		return model.PersonSnapshot{Name: name, Status: status}
	}
	record := func(action model.DBEditAction, s model.PersonSnapshot) {
		if err := repo.Create(ctx, testutil.NewDBEditBuilder(t, tx).WithAction(action).Build(), personID, s); err != nil {
			t.Fatalf("版の作成に失敗: %v", err)
		}
	}
	record(model.DBEditActionCreate, snapshot("掲載停止前", model.PersonStatusPublished))
	record(model.DBEditActionDelist, snapshot("掲載停止前", model.PersonStatusDelisted))
	record(model.DBEditActionUpdate, snapshot("掲載停止中の編集", model.PersonStatusDelisted))
	record(model.DBEditActionUndelist, snapshot("掲載停止中の編集", model.PersonStatusPublished))
	record(model.DBEditActionUpdate, snapshot("掲載再開後の編集", model.PersonStatusPublished))

	for _, tt := range []struct {
		name        string
		user        *model.User
		wantActions []model.DBEditAction
	}{
		{name: "ログインしていない閲覧者", wantActions: []model.DBEditAction{model.DBEditActionUpdate, model.DBEditActionCreate}},
		{name: "編集者", user: &model.User{ID: 1, Role: model.RoleEditor}, wantActions: []model.DBEditAction{model.DBEditActionUpdate, model.DBEditActionCreate}},
		{name: "管理者", user: &model.User{ID: 1, Role: model.RoleAdmin}, wantActions: []model.DBEditAction{
			model.DBEditActionUpdate, model.DBEditActionUndelist, model.DBEditActionUpdate, model.DBEditActionDelist, model.DBEditActionCreate,
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			output, err := uc.Execute(ctx, GetDBPersonRevisionsInput{User: tt.user, PersonID: personID, Page: 1, PerPage: 30})
			if err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}
			if output.TotalCount != int64(len(tt.wantActions)) || len(output.Revisions) != len(tt.wantActions) {
				t.Fatalf("版 = %d件 (総数%d)、期待値 = %d件", len(output.Revisions), output.TotalCount, len(tt.wantActions))
			}
			for i, want := range tt.wantActions {
				if got := output.Revisions[i].DBEdit.Action; got != want {
					t.Errorf("%d件目の操作 = %s、期待値 = %s", i+1, got, want)
				}
			}
			// 掲載再開のあとの版は、掲載再開の版と比べるため、状態の変化が出ない。
			latest := output.Revisions[0]
			diff := model.DiffPersonSnapshots(latest.Previous, latest.Snapshot)
			if len(diff.Fields) != 1 || diff.Fields[0].Field != model.RevisionField("name") {
				t.Errorf("最新の版の差分 = %+v、期待値 = 名前だけ", diff)
			}
		})
	}
}

func TestGetDBPersonRevisionsUsecase_Execute_NotFound(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name     string
		status   model.PersonStatus
		user     *model.User
		notFound bool
	}{
		{name: "削除した人物は管理者にも見せない", status: model.PersonStatusDeleted, user: &model.User{ID: 1, Role: model.RoleAdmin}, notFound: true},
		{name: "掲載停止中の人物は編集者に見せない", status: model.PersonStatusDelisted, user: &model.User{ID: 1, Role: model.RoleEditor}, notFound: true},
		{name: "掲載停止中の人物はログインしていない閲覧者に見せない", status: model.PersonStatusDelisted, notFound: true},
		{name: "掲載停止中の人物は管理者に見せる", status: model.PersonStatusDelisted, user: &model.User{ID: 1, Role: model.RoleAdmin}},
		{name: "非公開の人物は誰にでも見せる", status: model.PersonStatusArchived},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, tx := testutil.SetupTx(t)
			queries := query.New(db).WithTx(tx)
			uc := NewGetDBPersonRevisionsUsecase(repository.NewPersonRepository(queries), repository.NewPersonRevisionRepository(queries), repository.NewPrefectureRepository(queries))
			personID := testutil.NewPersonBuilder(t, tx).WithStatus(tt.status).Build()

			_, err := uc.Execute(context.Background(), GetDBPersonRevisionsInput{User: tt.user, PersonID: personID, Page: 1, PerPage: 30})
			ae := model.AsAppError(err)
			if tt.notFound != (ae != nil && ae.Code == model.AppErrCodeResourceNotFound) {
				t.Errorf("Execute()のエラー = %v、404を期待するか = %v", err, tt.notFound)
			}
			if !tt.notFound && err != nil {
				t.Errorf("Execute()のエラー = %v", err)
			}
		})
	}

	t.Run("存在しない人物", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		uc := NewGetDBPersonRevisionsUsecase(repository.NewPersonRepository(queries), repository.NewPersonRevisionRepository(queries), repository.NewPrefectureRepository(queries))
		_, err := uc.Execute(context.Background(), GetDBPersonRevisionsInput{PersonID: model.PersonID(1 << 62), Page: 1, PerPage: 30})
		if ae := model.AsAppError(err); ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
			t.Errorf("Execute()のエラー = %v、期待値 = AppErrCodeResourceNotFound", err)
		}
	})
}

// 版の「出身地」を名前で示すため、都道府県の一覧も返すことを確かめる。
func TestGetDBPersonRevisionsUsecase_Execute_Prefectures(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	uc := NewGetDBPersonRevisionsUsecase(repository.NewPersonRepository(queries), repository.NewPersonRevisionRepository(queries), repository.NewPrefectureRepository(queries))
	personID := testutil.NewPersonBuilder(t, tx).Build()
	prefectureID := testutil.CreateTestPrefecture(t, tx, "履歴の出身地")

	output, err := uc.Execute(context.Background(), GetDBPersonRevisionsInput{PersonID: personID, Page: 1, PerPage: 30})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if !slices.ContainsFunc(output.Prefectures, func(p model.Prefecture) bool {
		return p.ID == prefectureID && p.Name == "履歴の出身地"
	}) {
		t.Errorf("都道府県の一覧 = %+v、作成した出身地を含むことを期待", output.Prefectures)
	}
}
