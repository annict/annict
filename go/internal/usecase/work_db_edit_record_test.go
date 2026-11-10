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

// workRevisionForTestは作品の版と、その版が属する操作。
type workRevisionForTest struct {
	revision *model.AnimeRevision
	actor    model.DBEditActor
	userID   sql.NullInt64
	action   model.DBEditAction
}

// listWorkRevisionsForTestは作品のanimeの版を、起点の版を含めてIDの昇順に、操作とあわせて返す。
func listWorkRevisionsForTest(t *testing.T, db *sql.DB, workID model.WorkID) []workRevisionForTest {
	t.Helper()
	ctx := context.Background()

	animeID := reloadSyncWork(t, db, workID).AnimeID
	if animeID == nil {
		t.Fatalf("作品%dがanimeにマッピングされていない", workID)
	}
	revisions, err := repository.NewAnimeRevisionRepository(query.New(db)).ListByAnimeID(ctx, *animeID)
	if err != nil {
		t.Fatalf("ListByAnimeID()のエラー = %v", err)
	}

	result := make([]workRevisionForTest, len(revisions))
	for i, r := range revisions {
		result[i] = workRevisionForTest{revision: r}
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

func TestCreateWorkUsecase_Execute_RecordsDBEdit(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	title := "編集履歴の作成テスト_" + t.Name()
	workID := createMappedWork(t, db, title)

	revisions := listWorkRevisionsForTest(t, db, workID)
	// 作成の操作には起点の版を作らない。
	if len(revisions) != 1 {
		t.Fatalf("版の件数 = %d、期待値 = 1", len(revisions))
	}
	got := revisions[0]
	editor := committedEditorForTest(t, model.RoleEditor)
	if got.actor != model.DBEditActorUser || !got.userID.Valid || model.UserID(got.userID.Int64) != editor.ID || got.action != model.DBEditActionCreate {
		t.Errorf("操作 = %s・%v・%s、期待値 = user・%d・create", got.actor, got.userID, got.action, editor.ID)
	}
	s := got.revision.Snapshot
	if s.Title == nil || *s.Title != title || s.Kind != model.AnimeClassificationKindWork || s.Status != model.AnimeStatusPublished {
		t.Errorf("スナップショット = Title %v・Kind %q・Status %q、期待値 = %q・work・published", s.Title, s.Kind, s.Status, title)
	}
	// 作成した別表の行も、書き込んだ後の状態として持つ。
	if len(s.Links) == 0 || len(s.Seasons) == 0 {
		t.Errorf("別表の行 = リンク%d件・シーズン%d件、どちらも1件以上を期待", len(s.Links), len(s.Seasons))
	}
}

func TestUpdateWorkUsecase_Execute_RecordsDBEdit(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateWorkUsecase(db)
	ctx := context.Background()
	workID := createMappedWork(t, db, "編集履歴の更新テスト_"+t.Name())

	newTitle := "編集履歴の更新テスト後_" + t.Name()
	if _, err := uc.Execute(ctx, validUpdateWorkInput(t, db, workID, newTitle)); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	revisions := listWorkRevisionsForTest(t, db, workID)
	// 作成の版を比べる相手にするため、起点の版は足さない。
	if len(revisions) != 2 {
		t.Fatalf("版の件数 = %d、期待値 = 2 (作成と更新)", len(revisions))
	}
	updated := revisions[1]
	if updated.action != model.DBEditActionUpdate {
		t.Errorf("2つ目の版の操作 = %q、期待値 = update", updated.action)
	}
	if got := updated.revision.Snapshot.Title; got == nil || *got != newTitle {
		t.Errorf("更新の版のTitle = %v、期待値 = %q", got, newTitle)
	}
	diff := model.DiffAnimeSnapshots(&revisions[0].revision.Snapshot, updated.revision.Snapshot)
	if diff.IsEmpty() {
		t.Error("作成と更新の版の差分が無い")
	}

	// 内容の変わらない送信では、操作も版も作らない。
	if _, err := uc.Execute(ctx, validUpdateWorkInput(t, db, workID, newTitle)); err != nil {
		t.Fatalf("2回目のExecute()のエラー = %v", err)
	}
	if got := len(listWorkRevisionsForTest(t, db, workID)); got != 2 {
		t.Errorf("内容の変わらない送信のあとの版の件数 = %d、期待値 = 2", got)
	}
}

func TestUpdateWorkUsecase_Execute_AddsInitialRevision(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	ctx := context.Background()
	workID := createMappedWork(t, db, "編集履歴の起点の版テスト_"+t.Name())

	// 記録を始める前に作られた作品を再現するため、作成の版と操作を消す。
	animeID := *reloadSyncWork(t, db, workID).AnimeID
	if _, err := db.Exec(`DELETE FROM db_edits WHERE id IN (SELECT db_edit_id FROM anime_revisions WHERE anime_id = $1)`, int64(animeID)); err != nil {
		t.Fatalf("作成の版の削除に失敗: %v", err)
	}

	newTitle := "編集履歴の起点の版テスト後_" + t.Name()
	if _, err := newUpdateWorkUsecase(db).Execute(ctx, validUpdateWorkInput(t, db, workID, newTitle)); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	revisions := listWorkRevisionsForTest(t, db, workID)
	if len(revisions) != 2 {
		t.Fatalf("版の件数 = %d、期待値 = 2 (起点の版と更新)", len(revisions))
	}
	initial := revisions[0].revision
	if initial.DBEditID != nil {
		t.Errorf("1つ目の版のDBEditID = %d、起点の版 (nil) を期待", *initial.DBEditID)
	}
	if initial.Snapshot.Title == nil || *initial.Snapshot.Title != "編集履歴の起点の版テスト_"+t.Name() {
		t.Errorf("起点の版のTitle = %v、更新する前のタイトルを期待", initial.Snapshot.Title)
	}
}

func TestArchiveUnarchiveDeleteWork_RecordsDBEdit(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	ctx := context.Background()
	editor := committedEditorForTest(t, model.RoleAdmin)
	workID := createMappedWork(t, db, "編集履歴の状態テスト_"+t.Name())

	if _, err := newArchiveWorkUsecase(db).Execute(ctx, ArchiveWorkInput{User: editor, WorkID: workID}); err != nil {
		t.Fatalf("非公開のExecute()のエラー = %v", err)
	}
	if _, err := newUnarchiveWorkUsecase(db).Execute(ctx, UnarchiveWorkInput{User: editor, WorkID: workID}); err != nil {
		t.Fatalf("再公開のExecute()のエラー = %v", err)
	}
	if _, err := newDeleteWorkUsecase(db).Execute(ctx, DeleteWorkInput{User: editor, WorkID: workID}); err != nil {
		t.Fatalf("削除のExecute()のエラー = %v", err)
	}

	revisions := listWorkRevisionsForTest(t, db, workID)
	want := []struct {
		action model.DBEditAction
		status model.AnimeStatus
	}{
		{action: model.DBEditActionCreate, status: model.AnimeStatusPublished},
		{action: model.DBEditActionArchive, status: model.AnimeStatusArchived},
		{action: model.DBEditActionUnarchive, status: model.AnimeStatusPublished},
		{action: model.DBEditActionDelete, status: model.AnimeStatusDeleted},
	}
	if len(revisions) != len(want) {
		t.Fatalf("版の件数 = %d、期待値 = %d", len(revisions), len(want))
	}
	for i, w := range want {
		got := revisions[i]
		if got.action != w.action || got.revision.Snapshot.Status != w.status {
			t.Errorf("%d番目の版 = %s・%s、期待値 = %s・%s", i+1, got.action, got.revision.Snapshot.Status, w.action, w.status)
		}
		if i > 0 && model.UserID(got.userID.Int64) != editor.ID {
			t.Errorf("%d番目の版の編集者 = %v、期待値 = %d", i+1, got.userID, editor.ID)
		}
	}
	// 状態だけを変える操作では、別表の行は書き込む前のまま残る。
	diff := model.DiffAnimeSnapshots(&revisions[0].revision.Snapshot, revisions[1].revision.Snapshot)
	if len(diff.Rows) != 0 || len(diff.Fields) != 1 || diff.Fields[0].Field != "status" {
		t.Errorf("非公開の差分 = %+v、statusの変更だけを期待", diff)
	}
}

func TestWorkStateChanges_RejectConcurrentUpdate(t *testing.T) {
	t.Parallel()
	for _, action := range []string{"非公開", "再公開", "削除"} {
		t.Run(action, func(t *testing.T) {
			t.Parallel()
			db := testutil.GetTestDB()
			ctx := context.Background()
			workID := createMappedWork(t, db, "状態変更の競合_"+t.Name())
			editor := committedEditorForTest(t, model.RoleAdmin)
			if action == "再公開" {
				if _, err := newArchiveWorkUsecase(db).Execute(ctx, ArchiveWorkInput{User: editor, WorkID: workID}); err != nil {
					t.Fatalf("競合テストの準備または取得に失敗: %v", err)
				}
			}

			// 状態変更が読む状態と版を確保したあと、別の編集者の通常の更新を挟む。
			uc := newArchiveWorkUsecase(db)
			current := reloadSyncWork(t, db, workID)
			wantStatus := current.DerivedStatus()
			anime, err := uc.animeRepo.GetByID(ctx, *current.AnimeID)
			if err != nil {
				t.Fatalf("競合テストの準備または取得に失敗: %v", err)
			}
			before, err := readAnimeSnapshot(ctx, uc.animeClassificationRepo, uc.satelliteRepos, anime)
			if err != nil {
				t.Fatalf("競合テストの準備または取得に失敗: %v", err)
			}
			input := validUpdateWorkInput(t, db, workID, current.Title)
			input.OfficialSiteURL = "https://example.com/concurrent-change"
			if _, err := newUpdateWorkUsecase(db).Execute(ctx, input); err != nil {
				t.Fatalf("競合テストの準備または取得に失敗: %v", err)
			}
			revisionCount := len(listWorkRevisionsForTest(t, db, workID))

			switch action {
			case "非公開":
				_, err = uc.archiveWork(ctx, editor.ID, current, anime, before, time.Now())
			case "再公開":
				_, err = newUnarchiveWorkUsecase(db).unarchiveWork(ctx, editor.ID, current, anime, before)
			case "削除":
				_, err = newDeleteWorkUsecase(db).deleteWork(ctx, editor.ID, current, anime, before, time.Now())
			}
			if ae := model.AsAppError(err); ae == nil || ae.Code != model.AppErrCodeConflict {
				t.Fatalf("状態変更のエラー = %v、競合を期待", err)
			}
			if got := reloadSyncWork(t, db, workID).DerivedStatus(); got != wantStatus {
				t.Errorf("競合後の状態 = %q、期待値 = %q", got, wantStatus)
			}
			if got := len(listWorkRevisionsForTest(t, db, workID)); got != revisionCount {
				t.Errorf("競合後の版の件数 = %d、期待値 = %d", got, revisionCount)
			}
			links, err := uc.satelliteRepos.Link.ListByAnimeIDs(ctx, []model.AnimeID{anime.ID})
			if err != nil {
				t.Fatalf("競合テストの準備または取得に失敗: %v", err)
			}
			for _, link := range links {
				if link.Kind == model.AnimeLinkKindOfficialSite && link.Language == model.LanguageJa && link.URL != input.OfficialSiteURL {
					t.Errorf("公式サイトURL = %q、期待値 = %q", link.URL, input.OfficialSiteURL)
				}
			}
		})
	}
}
