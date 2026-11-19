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

// listAnimeRevisionsForTestはanimeの版を、起点の版を含めてIDの昇順に、操作とあわせて返す。
func listAnimeRevisionsForTest(t *testing.T, db *sql.DB, animeID model.AnimeID) []workRevisionForTest {
	t.Helper()

	revisions, err := repository.NewAnimeRevisionRepository(query.New(db)).ListByAnimeID(context.Background(), animeID)
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

// readEpisodeAnimeIDForTestはエピソードのanimeの写像を返す。
func readEpisodeAnimeIDForTest(t *testing.T, db *sql.DB, episodeID model.EpisodeID) model.AnimeID {
	t.Helper()

	var animeID sql.NullInt64
	if err := db.QueryRow(`SELECT anime_id FROM episodes WHERE id = $1`, int64(episodeID)).Scan(&animeID); err != nil {
		t.Fatalf("episodes.anime_idの取得に失敗: %v", err)
	}
	if !animeID.Valid {
		t.Fatalf("エピソード%dがanimeにマッピングされていない", episodeID)
	}
	return model.AnimeID(animeID.Int64)
}

func TestCreateEpisodesUsecase_Execute_RecordsOneDBEdit(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	editor := committedEditorForTest(t, model.RoleEditor)
	workID, parentAnimeID := insertMappedCreateTargetWork(t, db)

	output, err := newCreateEpisodesUsecase(db).Execute(context.Background(), CreateEpisodesInput{
		WorkID: workID,
		User:   editor,
		Rows:   "#1,1,はじまり\n#2,2,つづき",
	})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	var dbEditIDs []model.DBEditID
	for i, episodeID := range output.EpisodeIDs {
		revisions := listAnimeRevisionsForTest(t, db, readEpisodeAnimeIDForTest(t, db, episodeID))
		// 作成の操作には起点の版を作らない。
		if len(revisions) != 1 {
			t.Fatalf("%d行目の版の件数 = %d、期待値 = 1", i+1, len(revisions))
		}
		got := revisions[0]
		if got.actor != model.DBEditActorUser || model.UserID(got.userID.Int64) != editor.ID || got.action != model.DBEditActionCreate {
			t.Errorf("%d行目の操作 = %s・%v・%s、期待値 = user・%d・create", i+1, got.actor, got.userID, got.action, editor.ID)
		}
		s := got.revision.Snapshot
		if s.Kind != model.AnimeClassificationKindEpisode || s.ParentAnimeID == nil || *s.ParentAnimeID != parentAnimeID || s.Title == nil {
			t.Errorf("%d行目のスナップショット = Kind %q・ParentAnimeID %v・Title %v、期待値 = episode・%d・タイトルあり", i+1, s.Kind, s.ParentAnimeID, s.Title, parentAnimeID)
		}
		dbEditIDs = append(dbEditIDs, *got.revision.DBEditID)
	}
	// 一括作成は、作ったエピソードの版を1つの操作にまとめる。
	if len(dbEditIDs) != 2 || dbEditIDs[0] != dbEditIDs[1] {
		t.Errorf("版の操作 = %v、期待値 = 2行とも同じ操作", dbEditIDs)
	}
}

func TestCreateEpisodesUsecase_Execute_SkipsDBEditForUnmappedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	editor := committedEditorForTest(t, model.RoleEditor)
	workID := insertCreateTargetWork(t, db, sql.NullInt64{})

	var before int
	if err := db.QueryRow(`SELECT COUNT(*) FROM db_edits WHERE user_id = $1`, int64(editor.ID)).Scan(&before); err != nil {
		t.Fatalf("操作の数の取得に失敗: %v", err)
	}
	if _, err := newCreateEpisodesUsecase(db).Execute(context.Background(), CreateEpisodesInput{
		WorkID: workID,
		User:   editor,
		Rows:   "#1,1,はじまり",
	}); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	// 版はanimeを指すため、animeにマッピングされていない作品のエピソードは記録しない。ほかのテストも同じ編集者で
	// 記録するため、件数が増えないことではなく、この作品のエピソードの版が無いことを確かめる。
	var count int
	if err := db.QueryRow(`
		SELECT COUNT(*) FROM anime_revisions ar
		JOIN episodes e ON e.anime_id = ar.anime_id
		WHERE e.work_id = $1`, int64(workID)).Scan(&count); err != nil {
		t.Fatalf("版の数の取得に失敗: %v", err)
	}
	if count != 0 {
		t.Errorf("版の数 = %d、期待値 = 0", count)
	}
}

func TestUpdateEpisodeUsecase_Execute_RecordsDBEditWithInitialRevision(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	editor := committedEditorForTest(t, model.RoleEditor)
	workID, parentAnimeID := insertMappedCreateTargetWork(t, db)
	episodeID, animeID := insertMappedUpdateTargetEpisode(t, db, workID, parentAnimeID, 100)

	if _, err := newUpdateEpisodeUsecase(db).Execute(context.Background(), updateEpisodeSubmit(t, db, episodeID, editor)); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	revisions := listAnimeRevisionsForTest(t, db, animeID)
	// 記録を始める前からあったエピソードには、書き込む前の状態を起点の版として足す。
	if len(revisions) != 2 || revisions[0].revision.DBEditID != nil {
		t.Fatalf("版 = %d件、期待値 = 起点の版と更新の版の2件", len(revisions))
	}
	updated := revisions[1]
	if updated.actor != model.DBEditActorUser || model.UserID(updated.userID.Int64) != editor.ID || updated.action != model.DBEditActionUpdate {
		t.Errorf("操作 = %s・%v・%s、期待値 = user・%d・update", updated.actor, updated.userID, updated.action, editor.ID)
	}
	if got := updated.revision.Snapshot.Title; got == nil || *got != "もう、お婿にいけません" {
		t.Errorf("更新の版のTitle = %v、期待値 = もう、お婿にいけません", got)
	}
	if got := updated.revision.Snapshot.SortNumber; got == nil || *got != 250 {
		t.Errorf("更新の版のSortNumber = %v、期待値 = 250", got)
	}
	if model.DiffAnimeSnapshots(&revisions[0].revision.Snapshot, updated.revision.Snapshot).IsEmpty() {
		t.Error("起点の版と更新の版の差分が無い")
	}

	// 内容の変わらない送信では、操作も版も作らない。
	if _, err := newUpdateEpisodeUsecase(db).Execute(context.Background(), updateEpisodeSubmit(t, db, episodeID, editor)); err != nil {
		t.Fatalf("2回目のExecute()のエラー = %v", err)
	}
	if got := len(listAnimeRevisionsForTest(t, db, animeID)); got != 2 {
		t.Errorf("内容の変わらない送信のあとの版の件数 = %d、期待値 = 2", got)
	}
}

func TestEpisodeStatusUsecases_RecordDBEdits(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	ctx := context.Background()
	editor := committedEditorForTest(t, model.RoleEditor)
	admin := committedEditorForTest(t, model.RoleAdmin)
	workID, parentAnimeID := insertMappedCreateTargetWork(t, db)
	episodeID, animeID := insertMappedUpdateTargetEpisode(t, db, workID, parentAnimeID, 100)

	if _, err := newArchiveEpisodeUsecase(db).Execute(ctx, ArchiveEpisodeInput{EpisodeID: episodeID, User: editor}); err != nil {
		t.Fatalf("非公開のExecute()のエラー = %v", err)
	}
	if _, err := newUnarchiveEpisodeUsecase(db).Execute(ctx, UnarchiveEpisodeInput{EpisodeID: episodeID, User: editor}); err != nil {
		t.Fatalf("再公開のExecute()のエラー = %v", err)
	}
	if _, err := newDeleteEpisodeUsecase(db).Execute(ctx, DeleteEpisodeInput{EpisodeID: episodeID, User: admin}); err != nil {
		t.Fatalf("削除のExecute()のエラー = %v", err)
	}

	revisions := listAnimeRevisionsForTest(t, db, animeID)
	if len(revisions) != 4 || revisions[0].revision.DBEditID != nil {
		t.Fatalf("版 = %d件、期待値 = 起点の版と3つの操作の版の4件", len(revisions))
	}
	for i, want := range []struct {
		action model.DBEditAction
		userID model.UserID
		status model.AnimeStatus
	}{
		{model.DBEditActionArchive, editor.ID, model.AnimeStatusArchived},
		{model.DBEditActionUnarchive, editor.ID, model.AnimeStatusPublished},
		{model.DBEditActionDelete, admin.ID, model.AnimeStatusDeleted},
	} {
		got := revisions[i+1]
		if got.action != want.action || model.UserID(got.userID.Int64) != want.userID || got.revision.Snapshot.Status != want.status {
			t.Errorf("%d番目の版 = %s・%v・%s、期待値 = %s・%d・%s", i+1, got.action, got.userID, got.revision.Snapshot.Status, want.action, want.userID, want.status)
		}
		// 状態だけを書き換えるため、状態のほかの項目は変わらない。
		diff := model.DiffAnimeSnapshots(&revisions[i].revision.Snapshot, got.revision.Snapshot)
		if len(diff.Fields) != 1 || diff.Fields[0].Field != model.RevisionField("status") || len(diff.Rows) != 0 {
			t.Errorf("%d番目の版の差分 = %+v、期待値 = statusだけ", i+1, diff)
		}
	}
}

func TestCreateEpisodesFromSlotsUsecase_Execute_RecordsSystemDBEdit(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	now := generationNow()
	workID, parentAnimeID := insertMappedCreateTargetWork(t, db)
	program := insertGenerationProgram(t, db, workID, 1)
	program.insertSlot(1, now.Add(time.Hour), nil)

	output, err := newCreateEpisodesFromSlotsUsecase(db).Execute(context.Background(), CreateEpisodesFromSlotsInput{Now: now})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	var created *GeneratedEpisode
	for i := range output.CreatedEpisodes {
		if output.CreatedEpisodes[i].WorkID == workID {
			created = &output.CreatedEpisodes[i]
		}
	}
	if created == nil {
		t.Fatalf("output = %+v、期待値 = 作品%dのエピソードの作成", output, workID)
	}

	revisions := listAnimeRevisionsForTest(t, db, readEpisodeAnimeIDForTest(t, db, created.EpisodeID))
	if len(revisions) != 1 {
		t.Fatalf("版の件数 = %d、期待値 = 1", len(revisions))
	}
	got := revisions[0]
	if got.actor != model.DBEditActorSystem || got.userID.Valid || got.action != model.DBEditActionCreate {
		t.Errorf("操作 = %s・%v・%s、期待値 = system・NULL・create", got.actor, got.userID, got.action)
	}
	if s := got.revision.Snapshot; s.ParentAnimeID == nil || *s.ParentAnimeID != parentAnimeID || s.NumberText == nil || *s.NumberText != created.Number {
		t.Errorf("スナップショット = ParentAnimeID %v・NumberText %v、期待値 = %d・%q", s.ParentAnimeID, s.NumberText, parentAnimeID, created.Number)
	}
}

func TestFillEpisodeTitlesUsecase_Execute_RecordsSystemDBEdit(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	workID, tid := insertTitleFillWork(t, db)
	program := insertGenerationProgram(t, db, workID, 1)
	episodeID := insertTitleFillEpisode(t, db, program, 1)
	animeID := insertBareAnime(t, db)
	if _, err := db.Exec(`UPDATE episodes SET anime_id = $2 WHERE id = $1`, int64(episodeID), int64(animeID)); err != nil {
		t.Fatalf("animeの対応付けに失敗: %v", err)
	}

	lookup := &fakeSubTitleLookup{subTitles: map[int32]map[int32]string{tid: {1: "記録のテストのサブタイトル"}}}
	if _, err := newFillEpisodeTitlesUsecase(db, lookup).Execute(context.Background(), FillEpisodeTitlesInput{Now: time.Now()}); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	revisions := listAnimeRevisionsForTest(t, db, animeID)
	if len(revisions) != 2 || revisions[0].revision.DBEditID != nil {
		t.Fatalf("版 = %d件、期待値 = 起点の版と補完の版の2件", len(revisions))
	}
	got := revisions[1]
	if got.actor != model.DBEditActorSystem || got.userID.Valid || got.action != model.DBEditActionUpdate {
		t.Errorf("操作 = %s・%v・%s、期待値 = system・NULL・update", got.actor, got.userID, got.action)
	}
	diff := model.DiffAnimeSnapshots(&revisions[0].revision.Snapshot, got.revision.Snapshot)
	if len(diff.Fields) != 1 || diff.Fields[0].Field != model.RevisionField("title") || diff.Fields[0].After == nil || *diff.Fields[0].After != "記録のテストのサブタイトル" {
		t.Errorf("差分 = %+v、期待値 = タイトルだけ", diff)
	}
}
