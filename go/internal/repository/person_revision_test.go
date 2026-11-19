package repository_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestPersonRevisionRepository_CreateAndListByPersonID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewPersonRevisionRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	personID := testutil.NewPersonBuilder(t, tx).Build()
	dbEditID := testutil.NewDBEditBuilder(t, tx).Build()
	prefectureID := model.PrefectureID(13)

	initial := model.PersonSnapshot{Name: "人物", Status: model.PersonStatusPublished}
	after := model.PersonSnapshot{
		Name:              "人物",
		NameKana:          "じんぶつ",
		NameEn:            "Person",
		Nickname:          revisionPtr("愛称"),
		NicknameEn:        "Nickname",
		Gender:            revisionPtr(model.PersonGenderFemale),
		BloodType:         revisionPtr(model.PersonBloodTypeAB),
		PrefectureID:      &prefectureID,
		Birthday:          revisionPtr(time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC)),
		Height:            revisionPtr(int32(160)),
		URL:               revisionPtr("https://example.com/"),
		URLEn:             "https://example.com/en",
		WikipediaURL:      revisionPtr("https://ja.wikipedia.org/wiki/example"),
		WikipediaURLEn:    "https://en.wikipedia.org/wiki/example",
		TwitterUsername:   revisionPtr("example"),
		TwitterUsernameEn: "example_en",
		Status:            model.PersonStatusArchived,
		ArchiveMessage:    revisionPtr("非公開の案内"),
	}
	if err := repo.CreateInitial(ctx, personID, initial); err != nil {
		t.Fatalf("CreateInitial()のエラー = %v", err)
	}
	// 起点の版がすでにあるときは足さない。
	if err := repo.CreateInitial(ctx, personID, after); err != nil {
		t.Fatalf("2回目のCreateInitial()のエラー = %v", err)
	}
	if err := repo.Create(ctx, dbEditID, personID, after); err != nil {
		t.Fatalf("Create()のエラー = %v", err)
	}

	revisions, err := repo.ListByPersonID(ctx, personID)
	if err != nil {
		t.Fatalf("ListByPersonID()のエラー = %v", err)
	}
	if len(revisions) != 2 {
		t.Fatalf("版の件数 = %d、期待値 = 2", len(revisions))
	}
	if revisions[0].DBEditID != nil || !reflect.DeepEqual(revisions[0].Snapshot, initial) {
		t.Errorf("起点の版 = %+v、期待値 = DBEditIDがnilで%+v", revisions[0], initial)
	}
	if revisions[1].DBEditID == nil || *revisions[1].DBEditID != dbEditID {
		t.Errorf("版のDBEditID = %v、期待値 = %d", revisions[1].DBEditID, dbEditID)
	}
	if !reflect.DeepEqual(revisions[1].Snapshot, after) {
		t.Errorf("読み戻した版 = %+v、期待値 = %+v", revisions[1].Snapshot, after)
	}
	// 掲載停止の対応記録は保存形式に入れない。
	if got := testutil.CountRows(t, tx, "person_revisions", "person_id = $1 AND snapshot ? 'delist_message'", int64(personID)); got != 0 {
		t.Errorf("delist_messageを持つ版の件数 = %d、期待値 = 0", got)
	}
}

func TestPersonRevisionRepository_ListPersonIDsWithRevisions(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewPersonRevisionRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	withRevisionID := testutil.NewPersonBuilder(t, tx).Build()
	withoutRevisionID := testutil.NewPersonBuilder(t, tx).Build()
	if err := repo.CreateInitial(ctx, withRevisionID, model.PersonSnapshot{Name: "人物", Status: model.PersonStatusPublished}); err != nil {
		t.Fatalf("CreateInitial()のエラー = %v", err)
	}

	got, err := repo.ListPersonIDsWithRevisions(ctx, []model.PersonID{withRevisionID, withoutRevisionID})
	if err != nil {
		t.Fatalf("ListPersonIDsWithRevisions()のエラー = %v", err)
	}
	want := map[model.PersonID]bool{withRevisionID: true}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("版を持つ人物 = %v、期待値 = %v", got, want)
	}
}

func TestPersonRevisionRepository_ListHistoryByPersonID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewPersonRevisionRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()
	personID := testutil.NewPersonBuilder(t, tx).Build()

	if err := repo.CreateInitial(ctx, personID, model.PersonSnapshot{Name: "起点", Status: model.PersonStatusPublished}); err != nil {
		t.Fatalf("CreateInitial()のエラー = %v", err)
	}
	userID := testutil.NewUserBuilder(t, tx).WithProfileImageData(editorImageData).Build()
	for _, v := range []struct {
		action model.DBEditAction
		name   string
		status model.PersonStatus
	}{
		{model.DBEditActionUpdate, "1", model.PersonStatusPublished},
		{model.DBEditActionDelist, "1", model.PersonStatusDelisted},
		{model.DBEditActionUndelist, "1", model.PersonStatusPublished},
	} {
		editID := testutil.NewDBEditBuilder(t, tx).WithUser(userID).WithAction(v.action).Build()
		if err := repo.Create(ctx, editID, personID, model.PersonSnapshot{Name: v.name, Status: v.status}); err != nil {
			t.Fatalf("Create()のエラー = %v", err)
		}
	}

	for _, tt := range []struct {
		name            string
		includeDelisted bool
		wantNames       []string
		wantPrevious    []string
	}{
		// 掲載停止・掲載再開の操作の版を除く。起点の版は並べない。
		{name: "管理者以外", wantNames: []string{"1"}, wantPrevious: []string{"起点"}},
		{name: "管理者", includeDelisted: true, wantNames: []string{"1", "1", "1"}, wantPrevious: []string{"1", "1", "起点"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			revisions, err := repo.ListHistoryByPersonID(ctx, repository.PersonRevisionHistoryParams{PersonID: personID, IncludeDelisted: tt.includeDelisted, Page: 1, PerPage: 10})
			if err != nil {
				t.Fatalf("ListHistoryByPersonID()のエラー = %v", err)
			}
			count, err := repo.CountHistoryByPersonID(ctx, personID, tt.includeDelisted)
			if err != nil {
				t.Fatalf("CountHistoryByPersonID()のエラー = %v", err)
			}
			if len(revisions) != len(tt.wantNames) || count != int64(len(tt.wantNames)) {
				t.Fatalf("版 = %d件 (総数%d)、期待値 = %d件", len(revisions), count, len(tt.wantNames))
			}
			for i, r := range revisions {
				if r.Snapshot.Name != tt.wantNames[i] || r.Previous == nil || r.Previous.Name != tt.wantPrevious[i] {
					t.Errorf("%d件目の版 = %q (前の版 %v)、期待値 = %q (前の版 %q)", i+1, r.Snapshot.Name, r.Previous, tt.wantNames[i], tt.wantPrevious[i])
				}
				if r.DBEdit == nil || r.DBEdit.User == nil || r.DBEdit.User.ID != userID || r.DBEdit.User.ProfileImageData.String != editorImageData {
					t.Errorf("%d件目の操作 = %+v、期待値 = プロフィール画像を持つユーザー%dの操作", i+1, r.DBEdit, userID)
				}
			}
		})
	}
}
