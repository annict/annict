package usecase

import (
	"context"
	"database/sql"
	"sync"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

var (
	committedEditorsMu sync.Mutex
	committedEditors   = map[int32]model.UserID{}
)

// committedEditorForTestは、自前のトランザクションで書くUseCaseのテストが編集者に使う、共有テストDBに
// コミット済みのユーザーを返す。編集履歴の操作がusersを外部キーで指すため、行の無いユーザーでは書けない。
// 役割ごとに1人をパッケージのテストで共有し、行は次のテストDBのリセットまで残す。
func committedEditorForTest(t *testing.T, role int32) *model.User {
	t.Helper()

	committedEditorsMu.Lock()
	defer committedEditorsMu.Unlock()
	if id, ok := committedEditors[role]; ok {
		return &model.User{ID: id, Role: role}
	}

	tx, err := testutil.GetTestDB().Begin()
	if err != nil {
		t.Fatalf("編集者の作成トランザクションの開始に失敗: %v", err)
	}
	id := testutil.NewUserBuilder(t, tx).WithRole(role).Build()
	if err := tx.Commit(); err != nil {
		t.Fatalf("編集者の作成トランザクションのコミットに失敗: %v", err)
	}
	committedEditors[role] = id
	return &model.User{ID: id, Role: role}
}

// newTestDBEditRecordReposは編集履歴を記録するリポジトリの束を作る。
func newTestDBEditRecordRepos(queries *query.Queries) DBEditRecordRepos {
	return DBEditRecordRepos{
		DBEdit:         repository.NewDBEditRepository(queries),
		AnimeRevision:  repository.NewAnimeRevisionRepository(queries),
		PersonRevision: repository.NewPersonRevisionRepository(queries),
	}
}

func newDBEditRecordReposForTest(t *testing.T) (DBEditRecordRepos, *sql.Tx) {
	t.Helper()
	db, tx := testutil.SetupTx(t)
	return newTestDBEditRecordRepos(query.New(db)).WithTx(tx), tx
}

// recordDBEditForTestは計画を立てて書く。テストでは計画も書き込みも同じテスト用のトランザクションで行う。
func recordDBEditForTest(t *testing.T, repos DBEditRecordRepos, input DBEditRecordInput) {
	t.Helper()
	ctx := context.Background()
	plan, err := planDBEditRecord(ctx, repos, input)
	if err != nil {
		t.Fatalf("planDBEditRecord()のエラー = %v", err)
	}
	if err := writeDBEditRecord(ctx, repos, plan); err != nil {
		t.Fatalf("writeDBEditRecord()のエラー = %v", err)
	}
}

func dbEditRecordTitle(title string) model.AnimeSnapshot {
	return model.AnimeSnapshot{Title: &title, Status: model.AnimeStatusPublished, Kind: model.AnimeClassificationKindWork}
}

func TestPlanDBEditRecord_FirstUpdateAddsInitialRevision(t *testing.T) {
	t.Parallel()

	repos, tx := newDBEditRecordReposForTest(t)
	ctx := context.Background()
	userID := testutil.NewUserBuilder(t, tx).Build()
	animeID := testutil.CreateTestAnime(t, tx)
	personID := testutil.NewPersonBuilder(t, tx).Build()

	before := dbEditRecordTitle("前")
	personBefore := model.PersonSnapshot{Name: "人物", Status: model.PersonStatusPublished}
	personAfter := model.PersonSnapshot{Name: "人物", Status: model.PersonStatusArchived}
	recordDBEditForTest(t, repos, DBEditRecordInput{
		Actor:  model.DBEditActorUser,
		UserID: &userID,
		Action: model.DBEditActionUpdate,
		Animes: []AnimeRevisionInput{{AnimeID: animeID, Before: &before, After: dbEditRecordTitle("後")}},
		People: []PersonRevisionInput{{PersonID: personID, Before: &personBefore, After: personAfter}},
	})

	edits := testutil.CountRows(t, tx, "db_edits", "user_id = $1 AND actor = 'user' AND action = 'update'", int64(userID))
	if edits != 1 {
		t.Fatalf("操作の件数 = %d、期待値 = 1", edits)
	}
	revisions, err := repos.AnimeRevision.ListByAnimeID(ctx, animeID)
	if err != nil {
		t.Fatalf("ListByAnimeID()のエラー = %v", err)
	}
	// 書き込む前の状態を起点の版にし、その後ろに操作の版を足す。
	if len(revisions) != 2 {
		t.Fatalf("作品の版の件数 = %d、期待値 = 2", len(revisions))
	}
	if revisions[0].DBEditID != nil || *revisions[0].Snapshot.Title != "前" {
		t.Errorf("1つ目の版 = DBEditID %v・Title %q、期待値 = 起点の版・前", revisions[0].DBEditID, *revisions[0].Snapshot.Title)
	}
	if revisions[1].DBEditID == nil || *revisions[1].Snapshot.Title != "後" {
		t.Errorf("2つ目の版 = DBEditID %v・Title %q、期待値 = 操作の版・後", revisions[1].DBEditID, *revisions[1].Snapshot.Title)
	}
	personRevisions, err := repos.PersonRevision.ListByPersonID(ctx, personID)
	if err != nil {
		t.Fatalf("ListByPersonID()のエラー = %v", err)
	}
	if len(personRevisions) != 2 || personRevisions[0].DBEditID != nil || personRevisions[1].DBEditID == nil {
		t.Fatalf("人物の版の件数 = %d、期待値 = 起点の版と操作の版の2件", len(personRevisions))
	}
	if *personRevisions[1].DBEditID != *revisions[1].DBEditID {
		t.Errorf("人物の版の操作 = %d、作品の版の操作 = %d、同じ操作を期待", *personRevisions[1].DBEditID, *revisions[1].DBEditID)
	}
}

func TestPlanDBEditRecord_LaterUpdateAddsOnlyRevision(t *testing.T) {
	t.Parallel()

	repos, tx := newDBEditRecordReposForTest(t)
	ctx := context.Background()
	userID := testutil.NewUserBuilder(t, tx).Build()
	animeID := testutil.CreateTestAnime(t, tx)

	first := dbEditRecordTitle("1")
	second := dbEditRecordTitle("2")
	recordDBEditForTest(t, repos, DBEditRecordInput{
		Actor: model.DBEditActorUser, UserID: &userID, Action: model.DBEditActionUpdate,
		Animes: []AnimeRevisionInput{{AnimeID: animeID, Before: &first, After: second}},
	})
	recordDBEditForTest(t, repos, DBEditRecordInput{
		Actor: model.DBEditActorUser, UserID: &userID, Action: model.DBEditActionUpdate,
		Animes: []AnimeRevisionInput{{AnimeID: animeID, Before: &second, After: dbEditRecordTitle("3")}},
	})

	revisions, err := repos.AnimeRevision.ListByAnimeID(ctx, animeID)
	if err != nil {
		t.Fatalf("ListByAnimeID()のエラー = %v", err)
	}
	got := make([]string, len(revisions))
	for i, r := range revisions {
		got[i] = *r.Snapshot.Title
	}
	if len(revisions) != 3 || got[0] != "1" || got[1] != "2" || got[2] != "3" {
		t.Errorf("版のTitle = %v、期待値 = [1 2 3] (起点の版は1つだけ)", got)
	}
}

func TestPlanDBEditRecord_UnchangedWritesNothing(t *testing.T) {
	t.Parallel()

	repos, tx := newDBEditRecordReposForTest(t)
	ctx := context.Background()
	userID := testutil.NewUserBuilder(t, tx).Build()
	unchangedID := testutil.CreateTestAnime(t, tx)
	changedID := testutil.CreateTestAnime(t, tx)
	same := dbEditRecordTitle("同じ")

	t.Run("すべてのリソースが変わらないときは計画が無い", func(t *testing.T) {
		plan, err := planDBEditRecord(ctx, repos, DBEditRecordInput{
			Actor: model.DBEditActorUser, UserID: &userID, Action: model.DBEditActionUpdate,
			Animes: []AnimeRevisionInput{{AnimeID: unchangedID, Before: &same, After: same}},
		})
		if err != nil {
			t.Fatalf("planDBEditRecord()のエラー = %v", err)
		}
		if plan != nil {
			t.Errorf("計画 = %+v、nilを期待", plan)
		}
		if err := writeDBEditRecord(ctx, repos, plan); err != nil {
			t.Fatalf("writeDBEditRecord()のエラー = %v", err)
		}
		if got := testutil.CountRows(t, tx, "db_edits", "user_id = $1", int64(userID)); got != 0 {
			t.Errorf("操作の件数 = %d、期待値 = 0", got)
		}
	})

	t.Run("変わらないリソースの版だけを除く", func(t *testing.T) {
		before := dbEditRecordTitle("前")
		recordDBEditForTest(t, repos, DBEditRecordInput{
			Actor: model.DBEditActorUser, UserID: &userID, Action: model.DBEditActionUpdate,
			Animes: []AnimeRevisionInput{
				{AnimeID: unchangedID, Before: &same, After: same},
				{AnimeID: changedID, Before: &before, After: dbEditRecordTitle("後")},
			},
		})
		if got := testutil.CountRows(t, tx, "anime_revisions", "anime_id = $1", int64(unchangedID)); got != 0 {
			t.Errorf("変わらない作品の版の件数 = %d、期待値 = 0", got)
		}
		if got := testutil.CountRows(t, tx, "anime_revisions", "anime_id = $1", int64(changedID)); got != 2 {
			t.Errorf("変わった作品の版の件数 = %d、期待値 = 2", got)
		}
	})
}

func TestPlanDBEditRecord_CreateHasNoInitialRevision(t *testing.T) {
	t.Parallel()

	repos, tx := newDBEditRecordReposForTest(t)
	animeIDs := []model.AnimeID{testutil.CreateTestAnime(t, tx), testutil.CreateTestAnime(t, tx)}

	// エピソードの自動生成のような、システムの書き込みの一括作成。
	recordDBEditForTest(t, repos, DBEditRecordInput{
		Actor:  model.DBEditActorSystem,
		Action: model.DBEditActionCreate,
		Animes: []AnimeRevisionInput{
			{AnimeID: animeIDs[0], After: dbEditRecordTitle("第1話")},
			{AnimeID: animeIDs[1], After: dbEditRecordTitle("第2話")},
		},
	})

	for _, id := range animeIDs {
		if got := testutil.CountRows(t, tx, "anime_revisions", "anime_id = $1 AND db_edit_id IS NULL", int64(id)); got != 0 {
			t.Errorf("作品%dの起点の版の件数 = %d、期待値 = 0", id, got)
		}
		// 一括作成の版は1つの操作にまとまる。
		if got := testutil.CountRows(t, tx, "anime_revisions ar JOIN db_edits de ON de.id = ar.db_edit_id",
			"ar.anime_id = $1 AND de.actor = 'system' AND de.user_id IS NULL AND de.action = 'create' AND de.id = (SELECT db_edit_id FROM anime_revisions WHERE anime_id = $2)",
			int64(id), int64(animeIDs[0])); got != 1 {
			t.Errorf("作品%dのシステムの作成の版の件数 = %d、期待値 = 1", id, got)
		}
	}
}

func TestPlanDBEditRecord_ActorAndUserMismatch(t *testing.T) {
	t.Parallel()

	repos, tx := newDBEditRecordReposForTest(t)
	userID := testutil.NewUserBuilder(t, tx).Build()
	after := dbEditRecordTitle("後")
	animes := []AnimeRevisionInput{{AnimeID: testutil.CreateTestAnime(t, tx), After: after}}

	tests := []struct {
		name   string
		actor  model.DBEditActor
		userID *model.UserID
	}{
		{name: "異常系: 編集者の編集に編集者が無い", actor: model.DBEditActorUser},
		{name: "異常系: システムの書き込みに編集者がある", actor: model.DBEditActorSystem, userID: &userID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := planDBEditRecord(context.Background(), repos, DBEditRecordInput{
				Actor: tt.actor, UserID: tt.userID, Action: model.DBEditActionCreate, Animes: animes,
			})
			if err == nil {
				t.Error("エラーを期待したが、nilだった")
			}
		})
	}
}
