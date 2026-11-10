package usecase

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

// personRevisionForTestは人物の版と、その版が属する操作。
type personRevisionForTest struct {
	revision *model.PersonRevision
	actor    model.DBEditActor
	userID   sql.NullInt64
	action   model.DBEditAction
}

// listPersonRevisionsForTestは人物の版を、起点の版を含めてIDの昇順に、操作とあわせて返す。
func listPersonRevisionsForTest(t *testing.T, db *sql.DB, personID model.PersonID) []personRevisionForTest {
	t.Helper()

	revisions, err := repository.NewPersonRevisionRepository(query.New(db)).ListByPersonID(context.Background(), personID)
	if err != nil {
		t.Fatalf("ListByPersonID()のエラー = %v", err)
	}
	result := make([]personRevisionForTest, len(revisions))
	for i, r := range revisions {
		result[i] = personRevisionForTest{revision: r}
		if r.DBEditID == nil {
			continue
		}
		if err := db.QueryRow(`SELECT actor, user_id, action FROM db_edits WHERE id = $1`, int64(*r.DBEditID)).
			Scan(&result[i].actor, &result[i].userID, &result[i].action); err != nil {
			t.Fatalf("操作%dの取得に失敗: %v", *r.DBEditID, err)
		}
	}
	return result
}

func TestCreatePeopleUsecase_Execute_RecordsOneDBEdit(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	editor := committedEditorForTest(t, model.RoleEditor)
	first := createPeopleTestName("記録の人物1")
	second := createPeopleTestName("記録の人物2")
	cleanupCreatedPeople(t, db, first, second)

	output, err := newCreatePeopleUsecase(db).Execute(context.Background(), CreatePeopleInput{
		User: editor,
		Rows: first + ",きろくのじんぶついち\n" + second,
	})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	var dbEditIDs []model.DBEditID
	for i, personID := range output.PersonIDs {
		revisions := listPersonRevisionsForTest(t, db, personID)
		// 作成の操作には起点の版を作らない。
		if len(revisions) != 1 {
			t.Fatalf("%d件目の版の件数 = %d、期待値 = 1", i+1, len(revisions))
		}
		got := revisions[0]
		if got.actor != model.DBEditActorUser || model.UserID(got.userID.Int64) != editor.ID || got.action != model.DBEditActionCreate {
			t.Errorf("%d件目の操作 = %s・%v・%s、期待値 = user・%d・create", i+1, got.actor, got.userID, got.action, editor.ID)
		}
		if s := got.revision.Snapshot; !strings.HasPrefix(s.Name, "記録の人物") || s.Status != model.PersonStatusPublished {
			t.Errorf("%d件目のスナップショット = Name %q・Status %q、期待値 = 記録の人物・published", i+1, s.Name, s.Status)
		}
		dbEditIDs = append(dbEditIDs, *got.revision.DBEditID)
	}
	// 一括作成は、作った人物の版を1つの操作にまとめる。
	if len(dbEditIDs) != 2 || dbEditIDs[0] != dbEditIDs[1] {
		t.Errorf("版の操作 = %v、期待値 = 2件とも同じ操作", dbEditIDs)
	}
}

func TestUpdatePersonUsecase_Execute_RecordsDBEditWithInitialRevision(t *testing.T) {
	t.Parallel()

	uc, db := newUpdatePersonUsecase(t)
	editor := committedEditorForTest(t, model.RoleEditor)
	personID := testutil.CreateCommittedPerson(t, func(b *testutil.PersonBuilder) *testutil.PersonBuilder {
		return b.WithUpdatedAt(updatePersonTestVersion)
	})
	newName := "記録の更新後の人物-" + t.Name()

	if _, err := uc.Execute(context.Background(), UpdatePersonInput{
		PersonID:  personID,
		User:      editor,
		Name:      newName,
		Height:    "170",
		UpdatedAt: updatePersonTestVersion.Format(validator.FormVersionLayout),
	}); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	revisions := listPersonRevisionsForTest(t, db, personID)
	// 記録を始める前からあった人物には、書き込む前の状態を起点の版として足す。
	if len(revisions) != 2 || revisions[0].revision.DBEditID != nil {
		t.Fatalf("版 = %d件、期待値 = 起点の版と更新の版の2件", len(revisions))
	}
	updated := revisions[1]
	if updated.actor != model.DBEditActorUser || model.UserID(updated.userID.Int64) != editor.ID || updated.action != model.DBEditActionUpdate {
		t.Errorf("操作 = %s・%v・%s、期待値 = user・%d・update", updated.actor, updated.userID, updated.action, editor.ID)
	}
	if s := updated.revision.Snapshot; s.Name != newName || s.Height == nil || *s.Height != 170 {
		t.Errorf("更新の版 = Name %q・Height %v、期待値 = %q・170", s.Name, s.Height, newName)
	}
	if model.DiffPersonSnapshots(&revisions[0].revision.Snapshot, updated.revision.Snapshot).IsEmpty() {
		t.Error("起点の版と更新の版の差分が無い")
	}
}

func TestPersonStatusUsecases_RecordDBEdits(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	queries := query.New(db)
	personRepo := repository.NewPersonRepository(queries)
	recordRepos := newTestDBEditRecordRepos(queries)
	ctx := context.Background()
	editor := committedEditorForTest(t, model.RoleEditor)
	admin := committedEditorForTest(t, model.RoleAdmin)
	personID := testutil.CreateCommittedPerson(t, nil)
	version := func() string { return testutil.PersonVersion(t, db, int64(personID)) }

	steps := []struct {
		action model.DBEditAction
		run    func() error
	}{
		{model.DBEditActionArchive, func() error {
			_, err := NewArchivePersonUsecase(db, personRepo, recordRepos).Execute(ctx, ArchivePersonInput{User: editor, PersonID: personID})
			return err
		}},
		{model.DBEditActionUnarchive, func() error {
			_, err := NewUnarchivePersonUsecase(db, personRepo, recordRepos).Execute(ctx, UnarchivePersonInput{User: editor, PersonID: personID})
			return err
		}},
		{model.DBEditActionDelist, func() error {
			_, err := NewDelistPersonUsecase(db, personRepo, recordRepos).Execute(ctx, DelistPersonInput{User: admin, PersonID: personID, UpdatedAt: version(), DelistMessage: "本人からの申出"})
			return err
		}},
		{model.DBEditActionUndelist, func() error {
			_, err := NewUndelistPersonUsecase(db, personRepo, recordRepos).Execute(ctx, UndelistPersonInput{User: admin, PersonID: personID, UpdatedAt: version()})
			return err
		}},
		{model.DBEditActionDelete, func() error {
			_, err := NewDeletePersonUsecase(db, personRepo, recordRepos).Execute(ctx, DeletePersonInput{User: admin, PersonID: personID})
			return err
		}},
	}
	for _, step := range steps {
		if err := step.run(); err != nil {
			t.Fatalf("%sのExecute()のエラー = %v", step.action, err)
		}
	}

	revisions := listPersonRevisionsForTest(t, db, personID)
	if len(revisions) != len(steps)+1 || revisions[0].revision.DBEditID != nil {
		t.Fatalf("版 = %d件、期待値 = 起点の版と%dつの操作の版", len(revisions), len(steps))
	}
	wantStatuses := []model.PersonStatus{model.PersonStatusArchived, model.PersonStatusPublished, model.PersonStatusDelisted, model.PersonStatusPublished, model.PersonStatusDeleted}
	for i, step := range steps {
		got := revisions[i+1]
		if got.action != step.action || got.revision.Snapshot.Status != wantStatuses[i] {
			t.Errorf("%d番目の版 = %s・%s、期待値 = %s・%s", i+1, got.action, got.revision.Snapshot.Status, step.action, wantStatuses[i])
		}
		// 状態だけを書き換えるため、状態のほかの項目は変わらない。掲載停止の対応記録は版に入れない。
		diff := model.DiffPersonSnapshots(&revisions[i].revision.Snapshot, got.revision.Snapshot)
		if len(diff.Fields) != 1 || diff.Fields[0].Field != model.RevisionField("status") {
			t.Errorf("%d番目の版の差分 = %+v、期待値 = statusだけ", i+1, diff)
		}
	}
}
