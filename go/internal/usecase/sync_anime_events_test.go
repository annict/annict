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

// datePtr builds a *time.Time at midnight UTC for the given calendar date, matching the
// works.started_on / ended_on date columns the satellite loader projects.
//
// [Ja] datePtr は指定の暦日について UTC 午前 0 時の *time.Time を作る。別表ローダーが射影する
// works.started_on / ended_on の date 列に合わせる。
func datePtr(year int, month time.Month, day int) *time.Time {
	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return &t
}

// newSyncAnimeEventsUsecase builds the reconciler and its repository over the shared test
// DB. The usecase opens its own transaction, so the test commits its setup (animes)
// directly via GetTestDB rather than wrapping it in an outer tx.
//
// [Ja] newSyncAnimeEventsUsecase は共有テスト DB 上にリコンサイラとそのリポジトリを組み立てる。
// 本 UseCase は自前でトランザクションを開くため、テストは前提データ (animes) をアウター tx で
// 包まず GetTestDB 経由で直接コミットする。
func newSyncAnimeEventsUsecase(db *sql.DB) (*SyncAnimeEventsUsecase, *repository.AnimeEventRepository) {
	repo := repository.NewAnimeEventRepository(query.New(db))
	return NewSyncAnimeEventsUsecase(db, repo), repo
}

// workForEventSync builds an anime-resolved work carrying only the columns the event
// reconciler reads (started_on and ended_on).
//
// [Ja] workForEventSync はイベントリコンサイラが読むカラム (started_on と ended_on) だけを
// 持つ anime 解決済みの work を組み立てる。
func workForEventSync(animeID model.AnimeID, startedOn, endedOn *time.Time) *model.Work {
	aid := animeID
	return &model.Work{AnimeID: &aid, StartedOn: startedOn, EndedOn: endedOn}
}

// eventsByKind reads back an anime's events keyed by kind.
//
// [Ja] eventsByKind は anime のイベントを kind をキーに読み戻す。
func eventsByKind(t *testing.T, repo *repository.AnimeEventRepository, animeID model.AnimeID) map[model.AnimeEventKind]*model.AnimeEvent {
	t.Helper()
	rows, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs() error = %v", err)
	}
	byKind := make(map[model.AnimeEventKind]*model.AnimeEvent, len(rows))
	for _, r := range rows {
		byKind[r.Kind] = r
	}
	return byKind
}

func TestSyncAnimeEventsUsecase_Reconcile_CreatesRowFromWorkColumns(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeEventsUsecase(db)

	animeID := insertBareAnime(t, db)
	work := workForEventSync(animeID, datePtr(2024, time.January, 6), datePtr(2024, time.March, 30))

	counts, err := uc.Reconcile(context.Background(), []*model.Work{work})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if counts.Created != 1 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want Created 1 only", counts)
	}

	broadcast := eventsByKind(t, repo, animeID)[model.AnimeEventKindBroadcast]
	if broadcast == nil {
		t.Fatal("broadcast row should have been created")
	}
	if !sameDate(broadcast.StartedOn, *work.StartedOn) || broadcast.EndedOn == nil || !sameDate(*broadcast.EndedOn, *work.EndedOn) {
		t.Errorf("broadcast = %+v, want started 2024-01-06 ended 2024-03-30", broadcast)
	}
}

func TestSyncAnimeEventsUsecase_Reconcile_CreatesOpenEndedRowWhenEndMissing(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeEventsUsecase(db)

	animeID := insertBareAnime(t, db)
	// started_on present but ended_on NULL: an open-ended broadcast row is created.
	//
	// [Ja] started_on はあるが ended_on は NULL: 終了未定の broadcast 行を作る。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForEventSync(animeID, datePtr(2024, time.January, 6), nil)})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if counts.Created != 1 {
		t.Errorf("counts = %+v, want Created 1", counts)
	}

	broadcast := eventsByKind(t, repo, animeID)[model.AnimeEventKindBroadcast]
	if broadcast == nil || broadcast.EndedOn != nil {
		t.Errorf("broadcast = %+v, want a row with NULL ended_on", broadcast)
	}
}

func TestSyncAnimeEventsUsecase_Reconcile_IsIdempotent(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, _ := newSyncAnimeEventsUsecase(db)

	animeID := insertBareAnime(t, db)
	works := []*model.Work{workForEventSync(animeID, datePtr(2024, time.January, 6), datePtr(2024, time.March, 30))}

	if _, err := uc.Reconcile(context.Background(), works); err != nil {
		t.Fatalf("first Reconcile() error = %v", err)
	}

	// Re-running with the same source must detect no diff and write nothing. This is the
	// invariant the cutover decision depends on (a synced page reports zero diff). The
	// date round-trip through the date columns must compare equal for this to hold.
	//
	// [Ja] 同じソースで再実行すると差分は検出されず何も書かれない。正本切り替え判定が依拠する
	// 不変条件 (同期済みのページは差分ゼロを報告する)。これが成り立つには date 列を経由した日付の
	// round-trip が等しく比較されなければならない。
	counts, err := uc.Reconcile(context.Background(), works)
	if err != nil {
		t.Fatalf("second Reconcile() error = %v", err)
	}
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 1 {
		t.Errorf("counts = %+v, want Unchanged 1 only", counts)
	}
}

func TestSyncAnimeEventsUsecase_Reconcile_UpdatesChangedDates(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeEventsUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForEventSync(animeID, datePtr(2024, time.January, 6), datePtr(2024, time.March, 30))}); err != nil {
		t.Fatalf("first Reconcile() error = %v", err)
	}

	// kind is the natural key, so a changed broadcast period (started_on / ended_on) is an
	// in-place update of the same (anime_id, broadcast) row, not a delete plus a create.
	//
	// [Ja] kind が自然キーのため、放送期間 (started_on / ended_on) の変更は同じ
	// (anime_id, broadcast) 行のその場更新で、削除 + 作成ではない。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForEventSync(animeID, datePtr(2024, time.April, 6), datePtr(2024, time.June, 29))})
	if err != nil {
		t.Fatalf("second Reconcile() error = %v", err)
	}
	if counts.Updated != 1 || counts.Created != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want Updated 1 only", counts)
	}

	broadcast := eventsByKind(t, repo, animeID)[model.AnimeEventKindBroadcast]
	if broadcast == nil || !sameDate(broadcast.StartedOn, time.Date(2024, time.April, 6, 0, 0, 0, 0, time.UTC)) || broadcast.EndedOn == nil || !sameDate(*broadcast.EndedOn, time.Date(2024, time.June, 29, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("broadcast = %+v, want started 2024-04-06 ended 2024-06-29 after update", broadcast)
	}
}

func TestSyncAnimeEventsUsecase_Reconcile_UpdatesWhenEndDateAddedOrRemoved(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeEventsUsecase(db)

	animeID := insertBareAnime(t, db)

	// Start open-ended, then an end date appears: a NULL -> non-NULL ended_on change is
	// detected as a diff and updated in place.
	//
	// [Ja] 終了未定で始まり、その後終了日が現れる: ended_on の NULL -> 非 NULL 変化が差分として
	// 検出され、その場で更新される。
	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForEventSync(animeID, datePtr(2024, time.January, 6), nil)}); err != nil {
		t.Fatalf("first Reconcile() error = %v", err)
	}

	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForEventSync(animeID, datePtr(2024, time.January, 6), datePtr(2024, time.March, 30))})
	if err != nil {
		t.Fatalf("second Reconcile() error = %v", err)
	}
	if counts.Updated != 1 || counts.Created != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want Updated 1 only after end date added", counts)
	}

	broadcast := eventsByKind(t, repo, animeID)[model.AnimeEventKindBroadcast]
	if broadcast == nil || broadcast.EndedOn == nil {
		t.Errorf("broadcast = %+v, want an ended_on after update", broadcast)
	}
}

func TestSyncAnimeEventsUsecase_Reconcile_DeletesRowWhenSourceRemoved(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeEventsUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForEventSync(animeID, datePtr(2024, time.January, 6), datePtr(2024, time.March, 30))}); err != nil {
		t.Fatalf("first Reconcile() error = %v", err)
	}

	// started_on is gone (NULL); the broadcast row should be deleted.
	//
	// [Ja] started_on が消えた (NULL)。broadcast の行は削除されるべき。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForEventSync(animeID, nil, nil)})
	if err != nil {
		t.Fatalf("second Reconcile() error = %v", err)
	}
	if counts.Deleted != 1 || counts.Created != 0 || counts.Updated != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want Deleted 1 only", counts)
	}

	if _, ok := eventsByKind(t, repo, animeID)[model.AnimeEventKindBroadcast]; ok {
		t.Error("broadcast row should have been deleted")
	}
}

func TestSyncAnimeEventsUsecase_Reconcile_TreatsNullStartedOnAsNoRow(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeEventsUsecase(db)

	// A NULL started_on yields no row even when ended_on is present: the row's existence
	// is gated on the start date (anime_events.started_on is NOT NULL).
	//
	// [Ja] started_on が NULL なら ended_on があっても行は作られない: 行の有無は開始日で決まる
	// (anime_events.started_on は NOT NULL)。
	animeID := insertBareAnime(t, db)

	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForEventSync(animeID, nil, datePtr(2024, time.March, 30))})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want all zero", counts)
	}

	if rows := eventsByKind(t, repo, animeID); len(rows) != 0 {
		t.Errorf("rows = %v, want none for NULL started_on", rows)
	}
}

func TestSyncAnimeEventsUsecase_Reconcile_PreservesEditorAddedRows(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeEventsUsecase(db)

	animeID := insertBareAnime(t, db)

	// Seed two existing rows: one works-managed (broadcast) and one outside the
	// works-managed key space (a revival_screening an editor might add directly).
	// anime_events has no is_primary / sort_number marker, so the kind is what marks a row
	// as works-managed and thus deletable.
	//
	// [Ja] 既存行を 2 つ用意する: works 管理下の 1 つ (broadcast) と、works 管理下のキー空間の
	// 外の 1 つ (編集者が直接足しうる revival_screening)。anime_events には is_primary /
	// sort_number の目印が無いため、kind が「works 管理下 (= 削除対象)」であることを示す。
	started := time.Date(2024, time.January, 6, 0, 0, 0, 0, time.UTC)
	revivalStarted := time.Date(2034, time.January, 6, 0, 0, 0, 0, time.UTC)
	seed := []repository.CreateAnimeEventParams{
		{AnimeID: animeID, Kind: model.AnimeEventKindBroadcast, StartedOn: started, EndedOn: nil},
		{AnimeID: animeID, Kind: model.AnimeEventKindRevivalScreening, StartedOn: revivalStarted, EndedOn: nil},
	}
	for _, s := range seed {
		if _, err := repo.Create(context.Background(), s); err != nil {
			t.Fatalf("seed Create(%+v) error = %v", s, err)
		}
	}

	// The work sources nothing, so the managed broadcast row is deleted while the
	// editor-added revival_screening row outside the key space is preserved.
	//
	// [Ja] work は何も source しないため、管理下の broadcast は削除されるが、キー空間の外の
	// 編集者追加の revival_screening 行は保全される。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForEventSync(animeID, nil, nil)})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if counts.Deleted != 1 || counts.Created != 0 || counts.Updated != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v, want Deleted 1 only", counts)
	}

	byKind := eventsByKind(t, repo, animeID)
	if _, ok := byKind[model.AnimeEventKindBroadcast]; ok {
		t.Error("managed broadcast row should have been deleted")
	}
	if byKind[model.AnimeEventKindRevivalScreening] == nil {
		t.Error("editor-added revival_screening row should be preserved")
	}
}
