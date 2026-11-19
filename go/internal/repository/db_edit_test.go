package repository_test

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

func TestDBEditRepository_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		actor      model.DBEditActor
		withUser   bool
		action     model.DBEditAction
		wantErr    bool
		wantUserID bool
	}{
		{name: "正常系: 編集者の編集を作れる", actor: model.DBEditActorUser, withUser: true, action: model.DBEditActionUpdate, wantUserID: true},
		{name: "正常系: システムの書き込みを編集者なしで作れる", actor: model.DBEditActorSystem, action: model.DBEditActionCreate},
		{name: "異常系: システムの書き込みに編集者を持たせられない", actor: model.DBEditActorSystem, withUser: true, action: model.DBEditActionCreate, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, tx := testutil.SetupTx(t)
			repo := repository.NewDBEditRepository(query.New(db)).WithTx(tx)
			ctx := context.Background()

			params := repository.CreateDBEditParams{Actor: tt.actor, Action: tt.action}
			userID := testutil.NewUserBuilder(t, tx).Build()
			if tt.withUser {
				params.UserID = &userID
			}

			edit, err := repo.Create(ctx, params)
			if tt.wantErr {
				if err == nil {
					t.Error("エラーを期待したが、nilだった")
				}
				return
			}
			if err != nil {
				t.Fatalf("Create()のエラー = %v", err)
			}

			if edit.ID == 0 {
				t.Error("ID = 0、非0を期待")
			}
			if edit.Actor != tt.actor {
				t.Errorf("Actor = %q、期待値 = %q", edit.Actor, tt.actor)
			}
			if edit.Action != tt.action {
				t.Errorf("Action = %q、期待値 = %q", edit.Action, tt.action)
			}
			if edit.CreatedAt.IsZero() {
				t.Error("CreatedAt = ゼロ値、非ゼロ値を期待")
			}
			if tt.wantUserID {
				if edit.UserID == nil || *edit.UserID != userID {
					t.Errorf("UserID = %v、期待値 = %d", edit.UserID, userID)
				}
			} else if edit.UserID != nil {
				t.Errorf("UserID = %d、nilを期待", *edit.UserID)
			}
		})
	}
}

func TestDBEditRepository_List(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewDBEditRepository(query.New(db)).WithTx(tx)

	userID := testutil.NewUserBuilder(t, tx).WithProfileImageData(editorImageData).Build()
	createdAt := time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)
	userEditID := withVisibleRevisionForTest(t, tx, testutil.NewDBEditBuilder(t, tx).WithUser(userID).WithAction(model.DBEditActionCreate).WithCreatedAt(createdAt).Build())
	systemEditID := withVisibleRevisionForTest(t, tx, testutil.NewDBEditBuilder(t, tx).Build())
	// 退会したユーザーの編集は、主体がuserのまま編集者が空になる。
	detachedEditID := withVisibleRevisionForTest(t, tx, testutil.NewDBEditBuilder(t, tx).WithActor(model.DBEditActorUser).Build())

	edits, err := repo.List(context.Background(), repository.DBEditListParams{Page: 1, PerPage: 100})
	if err != nil {
		t.Fatalf("List()のエラー = %v", err)
	}

	indexes := make(map[model.DBEditID]int, len(edits))
	for i, edit := range edits {
		indexes[edit.ID] = i
	}
	wantPresent := []model.DBEditID{detachedEditID, systemEditID, userEditID}
	for i, id := range wantPresent {
		index, ok := indexes[id]
		if !ok {
			t.Fatalf("操作 %d が一覧に含まれていません", id)
		}
		// wantPresentはIDの降順に並べてあるため、一覧でも前の要素より後ろに来る。
		if i > 0 && index <= indexes[wantPresent[i-1]] {
			t.Errorf("操作 %d が操作 %d より前に並んでいます", id, wantPresent[i-1])
		}
	}

	tests := []struct {
		name       string
		id         model.DBEditID
		wantActor  model.DBEditActor
		wantUserID *model.UserID
		wantAction model.DBEditAction
	}{
		{name: "編集者の編集", id: userEditID, wantActor: model.DBEditActorUser, wantUserID: &userID, wantAction: model.DBEditActionCreate},
		{name: "システムの書き込み", id: systemEditID, wantActor: model.DBEditActorSystem, wantAction: model.DBEditActionUpdate},
		{name: "退会したユーザーの編集", id: detachedEditID, wantActor: model.DBEditActorUser, wantAction: model.DBEditActionUpdate},
	}
	for _, tt := range tests {
		edit := edits[indexes[tt.id]]
		if edit.Actor != tt.wantActor {
			t.Errorf("%sのActor = %q、期待値 = %q", tt.name, edit.Actor, tt.wantActor)
		}
		if edit.Action != tt.wantAction {
			t.Errorf("%sのAction = %q、期待値 = %q", tt.name, edit.Action, tt.wantAction)
		}
		switch {
		case tt.wantUserID == nil && edit.UserID != nil:
			t.Errorf("%sのUserID = %d、nilを期待", tt.name, *edit.UserID)
		case tt.wantUserID != nil && (edit.UserID == nil || *edit.UserID != *tt.wantUserID):
			t.Errorf("%sのUserID = %v、期待値 = %d", tt.name, edit.UserID, *tt.wantUserID)
		}
	}
	if got := edits[indexes[userEditID]].CreatedAt; !got.Equal(createdAt) {
		t.Errorf("CreatedAt = %v、期待値 = %v", got, createdAt)
	}
	if user := edits[indexes[userEditID]].User; user == nil || !user.ProfileImageData.Valid || user.ProfileImageData.String != editorImageData {
		t.Errorf("編集者の編集のUser = %+v、期待値 = プロフィール画像%qを持つ編集者", user, editorImageData)
	}
}

func TestDBEditRepository_List_Pagination(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewDBEditRepository(query.New(db)).WithTx(tx)

	for range 3 {
		withVisibleRevisionForTest(t, tx, testutil.NewDBEditBuilder(t, tx).Build())
	}

	first, err := repo.List(context.Background(), repository.DBEditListParams{Page: 1, PerPage: 1})
	if err != nil {
		t.Fatalf("1ページ目のList()のエラー = %v", err)
	}
	second, err := repo.List(context.Background(), repository.DBEditListParams{Page: 2, PerPage: 1})
	if err != nil {
		t.Fatalf("2ページ目のList()のエラー = %v", err)
	}

	if len(first) != 1 || len(second) != 1 {
		t.Fatalf("len(first) = %d、len(second) = %d、期待値 = 1、1", len(first), len(second))
	}
	// 2ページ目は1ページ目の続きのため、IDの降順で1ページ目より小さいIDから始まる。
	if first[0].ID <= second[0].ID {
		t.Errorf("1ページ目の操作 %d が2ページ目の操作 %d より新しくありません", first[0].ID, second[0].ID)
	}
}

func TestDBEditRepository_DetachFromUser(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewDBEditRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	editIDs := []model.DBEditID{
		testutil.NewDBEditBuilder(t, tx).WithUser(userID).Build(),
		testutil.NewDBEditBuilder(t, tx).WithUser(userID).Build(),
	}
	// 他人の編集は編集者を残す。
	otherID := testutil.NewDBEditBuilder(t, tx).WithUser(otherUserID).Build()

	// 上限を超える分は残し、次の呼び出しで外す。
	for i, want := range []int64{1, 1, 0} {
		got, err := repo.DetachFromUser(ctx, userID, 1)
		if err != nil {
			t.Fatalf("%d回目のDetachFromUserのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の外した件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	for _, id := range editIDs {
		if got := testutil.CountRows(t, tx, "db_edits", "id = $1 AND actor = 'user' AND user_id IS NULL", int64(id)); got != 1 {
			t.Errorf("操作%dの編集者が外れていない", id)
		}
	}
	if got := testutil.CountRows(t, tx, "db_edits", "id = $1 AND user_id = $2", int64(otherID), int64(otherUserID)); got != 1 {
		t.Errorf("他人の編集の編集者の件数 = %d、期待値 = 1", got)
	}
}

// withVisibleRevisionForTestは、操作に公開中の人物の版を1つ足す。操作の一覧は見せる版を持つ操作だけを並べるため。
func withVisibleRevisionForTest(t *testing.T, tx *sql.Tx, editID model.DBEditID) model.DBEditID {
	t.Helper()

	personID := testutil.NewPersonBuilder(t, tx).Build()
	repo := repository.NewPersonRevisionRepository(query.New(tx))
	if err := repo.Create(context.Background(), editID, personID, model.PersonSnapshot{Name: "版の人物", Status: model.PersonStatusPublished}); err != nil {
		t.Fatalf("操作%dの版の作成に失敗: %v", editID, err)
	}
	return editID
}

// TestDBEditRepository_List_VisibleRevisionsは、見せる版を持たない操作を並べず、管理者以外には人物の掲載停止に
// かかわる版しか持たない操作を並べないことを検証する。
func TestDBEditRepository_List_VisibleRevisions(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	ctx := context.Background()
	queries := query.New(db).WithTx(tx)
	repo := repository.NewDBEditRepository(queries)
	personRevisionRepo := repository.NewPersonRevisionRepository(queries)
	animeRevisionRepo := repository.NewAnimeRevisionRepository(queries)

	record := func(action model.DBEditAction, personStatus model.PersonStatus, snapshotStatus model.PersonStatus) model.DBEditID {
		editID := testutil.NewDBEditBuilder(t, tx).WithAction(action).Build()
		personID := testutil.NewPersonBuilder(t, tx).WithStatus(personStatus).Build()
		if err := personRevisionRepo.Create(ctx, editID, personID, model.PersonSnapshot{Name: "人物", Status: snapshotStatus}); err != nil {
			t.Fatalf("版の作成に失敗: %v", err)
		}
		return editID
	}
	visible := record(model.DBEditActionUpdate, model.PersonStatusPublished, model.PersonStatusPublished)
	delisting := record(model.DBEditActionDelist, model.PersonStatusPublished, model.PersonStatusDelisted)
	whileDelisted := record(model.DBEditActionUpdate, model.PersonStatusDelisted, model.PersonStatusDelisted)
	deletedPerson := record(model.DBEditActionUpdate, model.PersonStatusDeleted, model.PersonStatusPublished)
	empty := testutil.NewDBEditBuilder(t, tx).Build()

	// 削除した作品の版だけを持つ操作も並べない。
	deletedWorkEdit := testutil.NewDBEditBuilder(t, tx).Build()
	animeID := testutil.CreateTestAnime(t, tx)
	workID := testutil.NewWorkBuilder(t, tx).WithMedia(1).WithDeletedAt(time.Now()).Build()
	if _, err := tx.Exec(`UPDATE works SET anime_id = $1 WHERE id = $2`, int64(animeID), int64(workID)); err != nil {
		t.Fatalf("works.anime_idの設定に失敗: %v", err)
	}
	if err := animeRevisionRepo.Create(ctx, deletedWorkEdit, animeID, model.AnimeSnapshot{Status: model.AnimeStatusDeleted, Kind: model.AnimeClassificationKindWork}); err != nil {
		t.Fatalf("作品の版の作成に失敗: %v", err)
	}

	for _, tt := range []struct {
		name            string
		includeDelisted bool
		want            map[model.DBEditID]bool
	}{
		{name: "管理者以外", want: map[model.DBEditID]bool{visible: true, delisting: false, whileDelisted: false, deletedPerson: false, empty: false, deletedWorkEdit: false}},
		{name: "管理者", includeDelisted: true, want: map[model.DBEditID]bool{visible: true, delisting: true, whileDelisted: true, deletedPerson: false, empty: false, deletedWorkEdit: false}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			edits, err := repo.List(ctx, repository.DBEditListParams{IncludeDelisted: tt.includeDelisted, Page: 1, PerPage: 1000})
			if err != nil {
				t.Fatalf("List()のエラー = %v", err)
			}
			listed := make(map[model.DBEditID]bool, len(edits))
			for _, e := range edits {
				listed[e.ID] = true
			}
			for id, want := range tt.want {
				if listed[id] != want {
					t.Errorf("操作%dが並ぶか = %v、期待値 = %v", id, listed[id], want)
				}
			}
		})
	}
}
