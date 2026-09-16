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

// datePtrは指定の暦日についてUTC午前0時の *time.Timeを作る。別表ローダーが射影する
// works.started_on / ended_onのdate列に合わせる。
func datePtr(year int, month time.Month, day int) *time.Time {
	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return &t
}

// newSyncAnimeEventsUsecaseは共有テストDB上にリコンサイラとそのリポジトリを組み立てる。
// 本UseCaseは自前でトランザクションを開くため、テストは前提データ (animes) をアウターtxで
// 包まずGetTestDB経由で直接コミットする。
func newSyncAnimeEventsUsecase(db *sql.DB) (*SyncAnimeEventsUsecase, *repository.AnimeEventRepository) {
	repo := repository.NewAnimeEventRepository(query.New(db))
	return NewSyncAnimeEventsUsecase(db, repo), repo
}

// workForEventSyncはイベントリコンサイラが読むカラム (started_onとended_on) だけを
// 持つanime解決済みのworkを組み立てる。
func workForEventSync(animeID model.AnimeID, startedOn, endedOn *time.Time) *model.Work {
	aid := animeID
	return &model.Work{AnimeID: &aid, StartedOn: startedOn, EndedOn: endedOn}
}

// eventsByKindはanimeのイベントをkindをキーに読み戻す。
func eventsByKind(t *testing.T, repo *repository.AnimeEventRepository, animeID model.AnimeID) map[model.AnimeEventKind]*model.AnimeEvent {
	t.Helper()
	rows, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs()のエラー = %v", err)
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
		t.Fatalf("Reconcile()のエラー = %v", err)
	}
	if counts.Created != 1 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = Createdのみ1", counts)
	}

	broadcast := eventsByKind(t, repo, animeID)[model.AnimeEventKindBroadcast]
	if broadcast == nil {
		t.Fatal("broadcastの行が作成されなかった")
	}
	if !sameDate(broadcast.StartedOn, *work.StartedOn) || broadcast.EndedOn == nil || !sameDate(*broadcast.EndedOn, *work.EndedOn) {
		t.Errorf("broadcast = %+v、期待値 = started_on 2024-01-06 / ended_on 2024-03-30", broadcast)
	}
}

func TestSyncAnimeEventsUsecase_Reconcile_CreatesOpenEndedRowWhenEndMissing(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeEventsUsecase(db)

	animeID := insertBareAnime(t, db)
	// started_onはあるがended_onはNULL: 終了未定のbroadcast行を作る。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForEventSync(animeID, datePtr(2024, time.January, 6), nil)})
	if err != nil {
		t.Fatalf("Reconcile()のエラー = %v", err)
	}
	if counts.Created != 1 {
		t.Errorf("counts = %+v、期待値 = Created 1", counts)
	}

	broadcast := eventsByKind(t, repo, animeID)[model.AnimeEventKindBroadcast]
	if broadcast == nil || broadcast.EndedOn != nil {
		t.Errorf("broadcast = %+v、期待値 = ended_onがNULLの行", broadcast)
	}
}

func TestSyncAnimeEventsUsecase_Reconcile_IsIdempotent(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, _ := newSyncAnimeEventsUsecase(db)

	animeID := insertBareAnime(t, db)
	works := []*model.Work{workForEventSync(animeID, datePtr(2024, time.January, 6), datePtr(2024, time.March, 30))}

	if _, err := uc.Reconcile(context.Background(), works); err != nil {
		t.Fatalf("1回目のReconcile()のエラー = %v", err)
	}

	// 同じソースで再実行すると差分は検出されず何も書かれない。正本切り替え判定が依拠する
	// 不変条件 (同期済みのページは差分ゼロを報告する)。これが成り立つにはdate列を経由した日付の
	// round-tripが等しく比較されなければならない。
	counts, err := uc.Reconcile(context.Background(), works)
	if err != nil {
		t.Fatalf("2回目のReconcile()のエラー = %v", err)
	}
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 1 {
		t.Errorf("counts = %+v、期待値 = Unchangedのみ1", counts)
	}
}

func TestSyncAnimeEventsUsecase_Reconcile_UpdatesChangedDates(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeEventsUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForEventSync(animeID, datePtr(2024, time.January, 6), datePtr(2024, time.March, 30))}); err != nil {
		t.Fatalf("1回目のReconcile()のエラー = %v", err)
	}

	// kindが自然キーのため、放送期間 (started_on / ended_on) の変更は同じ
	// (anime_id, broadcast) 行のその場更新で、削除 + 作成ではない。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForEventSync(animeID, datePtr(2024, time.April, 6), datePtr(2024, time.June, 29))})
	if err != nil {
		t.Fatalf("2回目のReconcile()のエラー = %v", err)
	}
	if counts.Updated != 1 || counts.Created != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = Updatedのみ1", counts)
	}

	broadcast := eventsByKind(t, repo, animeID)[model.AnimeEventKindBroadcast]
	if broadcast == nil || !sameDate(broadcast.StartedOn, time.Date(2024, time.April, 6, 0, 0, 0, 0, time.UTC)) || broadcast.EndedOn == nil || !sameDate(*broadcast.EndedOn, time.Date(2024, time.June, 29, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("broadcast = %+v、期待値 = 更新後にstarted_on 2024-04-06 / ended_on 2024-06-29", broadcast)
	}
}

func TestSyncAnimeEventsUsecase_Reconcile_UpdatesWhenEndDateAddedOrRemoved(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeEventsUsecase(db)

	animeID := insertBareAnime(t, db)

	// 終了未定で始まり、その後終了日が現れる: ended_onのNULL -> 非NULL変化が差分として
	// 検出され、その場で更新される。
	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForEventSync(animeID, datePtr(2024, time.January, 6), nil)}); err != nil {
		t.Fatalf("1回目のReconcile()のエラー = %v", err)
	}

	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForEventSync(animeID, datePtr(2024, time.January, 6), datePtr(2024, time.March, 30))})
	if err != nil {
		t.Fatalf("2回目のReconcile()のエラー = %v", err)
	}
	if counts.Updated != 1 || counts.Created != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("終了日の追加後のcounts = %+v、期待値 = Updatedのみ1", counts)
	}

	broadcast := eventsByKind(t, repo, animeID)[model.AnimeEventKindBroadcast]
	if broadcast == nil || broadcast.EndedOn == nil {
		t.Errorf("broadcast = %+v、期待値 = 更新後のended_onが入った行", broadcast)
	}
}

func TestSyncAnimeEventsUsecase_Reconcile_DeletesRowWhenSourceRemoved(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeEventsUsecase(db)

	animeID := insertBareAnime(t, db)

	if _, err := uc.Reconcile(context.Background(), []*model.Work{workForEventSync(animeID, datePtr(2024, time.January, 6), datePtr(2024, time.March, 30))}); err != nil {
		t.Fatalf("1回目のReconcile()のエラー = %v", err)
	}

	// started_onが消えた (NULL)。broadcastの行は削除されるべき。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForEventSync(animeID, nil, nil)})
	if err != nil {
		t.Fatalf("2回目のReconcile()のエラー = %v", err)
	}
	if counts.Deleted != 1 || counts.Created != 0 || counts.Updated != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = Deletedのみ1", counts)
	}

	if _, ok := eventsByKind(t, repo, animeID)[model.AnimeEventKindBroadcast]; ok {
		t.Error("broadcastの行が削除されなかった")
	}
}

func TestSyncAnimeEventsUsecase_Reconcile_TreatsNullStartedOnAsNoRow(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeEventsUsecase(db)

	// started_onがNULLならended_onがあっても行は作られない: 行の有無は開始日で決まる
	// (anime_events.started_onはNOT NULL)。
	animeID := insertBareAnime(t, db)

	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForEventSync(animeID, nil, datePtr(2024, time.March, 30))})
	if err != nil {
		t.Fatalf("Reconcile()のエラー = %v", err)
	}
	if counts.Created != 0 || counts.Updated != 0 || counts.Deleted != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = すべて0", counts)
	}

	if rows := eventsByKind(t, repo, animeID); len(rows) != 0 {
		t.Errorf("started_onがNULLのときのrows = %v、期待値 = 0件", rows)
	}
}

func TestSyncAnimeEventsUsecase_Reconcile_PreservesEditorAddedRows(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc, repo := newSyncAnimeEventsUsecase(db)

	animeID := insertBareAnime(t, db)

	// 既存行を2つ用意する: works管理下の1つ (broadcast) と、works管理下のキー空間の
	// 外の1つ (編集者が直接足しうるrevival_screening)。anime_eventsにはis_primary /
	// sort_numberの目印が無いため、kindが「works管理下 (= 削除対象)」であることを示す。
	started := time.Date(2024, time.January, 6, 0, 0, 0, 0, time.UTC)
	revivalStarted := time.Date(2034, time.January, 6, 0, 0, 0, 0, time.UTC)
	seed := []repository.CreateAnimeEventParams{
		{AnimeID: animeID, Kind: model.AnimeEventKindBroadcast, StartedOn: started, EndedOn: nil},
		{AnimeID: animeID, Kind: model.AnimeEventKindRevivalScreening, StartedOn: revivalStarted, EndedOn: nil},
	}
	for _, s := range seed {
		if _, err := repo.Create(context.Background(), s); err != nil {
			t.Fatalf("シードのCreate(%+v)のエラー = %v", s, err)
		}
	}

	// workは何もsourceしないため、管理下のbroadcastは削除されるが、キー空間の外の
	// 編集者追加のrevival_screening行は保全される。
	counts, err := uc.Reconcile(context.Background(), []*model.Work{workForEventSync(animeID, nil, nil)})
	if err != nil {
		t.Fatalf("Reconcile()のエラー = %v", err)
	}
	if counts.Deleted != 1 || counts.Created != 0 || counts.Updated != 0 || counts.Unchanged != 0 {
		t.Errorf("counts = %+v、期待値 = Deletedのみ1", counts)
	}

	byKind := eventsByKind(t, repo, animeID)
	if _, ok := byKind[model.AnimeEventKindBroadcast]; ok {
		t.Error("管理対象のbroadcastの行が削除されなかった")
	}
	if byKind[model.AnimeEventKindRevivalScreening] == nil {
		t.Error("編集者が追加したrevival_screeningの行が保持されなかった")
	}
}
