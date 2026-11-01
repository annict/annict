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

func TestLibraryEntryRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewLibraryEntryRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()

	// 見たい・見てる・見たの作品だけ視聴者数が減り、それ以外の作品は変わらない。
	tests := []struct {
		status            string
		wantWatchersCount int32
	}{
		{status: "wanna_watch", wantWatchersCount: 9},
		{status: "watching", wantWatchersCount: 9},
		{status: "watched", wantWatchersCount: 9},
		{status: "on_hold", wantWatchersCount: 10},
		{status: "stop_watching", wantWatchersCount: 10},
	}
	workIDs := make([]model.WorkID, len(tests))
	for i, tt := range tests {
		workIDs[i] = testutil.NewWorkBuilder(t, tx).WithWatchersCount(10).Build()
		testutil.NewLibraryEntryBuilder(t, tx).
			WithUserID(userID).
			WithWorkID(workIDs[i]).
			WithStatus(tt.status).
			Build()
	}
	otherLibraryEntryID := testutil.NewLibraryEntryBuilder(t, tx).
		WithUserID(otherUserID).
		WithWorkID(workIDs[0]).
		Build()

	// 上限を超える分は残し、次の呼び出しで削除する。
	for i, want := range []int64{3, 2, 0} {
		got, err := repo.DeleteByUserID(ctx, userID, 3)
		if err != nil {
			t.Fatalf("%d回目のDeleteByUserIDのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の削除件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	if got := testutil.CountRows(t, tx, "library_entries", "user_id = $1", int64(userID)); got != 0 {
		t.Errorf("ユーザーのライブラリエントリの件数 = %d、期待値 = 0", got)
	}
	if got := testutil.CountRows(t, tx, "library_entries", "id = $1", otherLibraryEntryID); got != 1 {
		t.Errorf("他のユーザーのライブラリエントリの件数 = %d、期待値 = 1", got)
	}
	for i, tt := range tests {
		var watchersCount int32
		if err := tx.QueryRowContext(ctx, "SELECT watchers_count FROM works WHERE id = $1", int64(workIDs[i])).Scan(&watchersCount); err != nil {
			t.Fatalf("作品の視聴者数の取得に失敗: %v", err)
		}
		if watchersCount != tt.wantWatchersCount {
			t.Errorf("%sの作品の視聴者数 = %d、期待値 = %d", tt.status, watchersCount, tt.wantWatchersCount)
		}
	}
}

// libraryEntryNextResourcesは、テストで確かめるライブラリの次のエピソード・次の放送枠と更新日時。
type libraryEntryNextResources struct {
	nextEpisodeID sql.NullInt64
	nextSlotID    sql.NullInt64
	updatedAt     time.Time
}

func getLibraryEntryNextResources(t *testing.T, tx *sql.Tx, id int64) libraryEntryNextResources {
	t.Helper()

	var got libraryEntryNextResources
	if err := tx.QueryRow(
		"SELECT next_episode_id, next_slot_id, updated_at FROM library_entries WHERE id = $1", id,
	).Scan(&got.nextEpisodeID, &got.nextSlotID, &got.updatedAt); err != nil {
		t.Fatalf("ライブラリエントリの取得に失敗: %v", err)
	}
	return got
}

func TestLibraryEntryRepository_UpdateNextResources(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewLibraryEntryRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	// 共有DBにほかのテストがコミットした行を対象にしないよう、実行時刻を遠い未来に置く。
	now := time.Date(2100, 1, 10, 12, 0, 0, 0, time.UTC)
	updatedFrom := time.Date(2100, 1, 7, 0, 0, 0, 0, time.UTC)
	before := updatedFrom.Add(-time.Second)

	// 作品のエピソードは、削除済みと非公開のものを並び順の先頭に置き、次のエピソードに選ばれ
	// ないことを確かめる。
	workID := testutil.NewWorkBuilder(t, tx).Build()
	episode1 := testutil.NewEpisodeBuilder(t, tx, workID).WithSortNumber(10).Build()
	episode2 := testutil.NewEpisodeBuilder(t, tx, workID).WithSortNumber(20).Build()
	testutil.NewEpisodeBuilder(t, tx, workID).WithSortNumber(1).WithDeletedAt(now).Build()
	testutil.NewEpisodeBuilder(t, tx, workID).WithSortNumber(2).WithUnpublishedAt(now).Build()

	// 番組の放送枠は、1話を非公開、2話を公開中にする。
	channelGroupID := testutil.NewChannelGroupBuilder(t, tx).Build()
	channelID := testutil.NewChannelBuilder(t, tx).WithChannelGroupID(channelGroupID).Build()
	programID := testutil.NewProgramBuilder(t, tx).WithChannelID(channelID).WithWorkID(workID).Build()
	newSlot := func(episodeID model.EpisodeID) *testutil.SlotBuilder {
		return testutil.NewSlotBuilder(t, tx).WithWorkID(workID).WithChannelID(channelID).WithProgramID(programID).WithEpisodeID(episodeID)
	}
	newSlot(episode1).WithUnpublishedAt(now).Build()
	slot2 := newSlot(episode2).Build()

	newEntry := func() *testutil.LibraryEntryBuilder {
		userID := testutil.NewUserBuilder(t, tx).Build()
		return testutil.NewLibraryEntryBuilder(t, tx).WithUserID(userID).WithWorkID(workID).WithUpdatedAt(updatedFrom)
	}
	// idの順に作り、2件ずつのページに区切ったときの読み方を確かめる。
	withProgram := newEntry().WithProgramID(programID).WithWatchedEpisodeIDs(episode1).Build()
	withoutProgram := newEntry().Build()
	unchangedSlot := newEntry().WithProgramID(programID).WithNextEpisodeID(episode1).Build()
	allWatched := newEntry().WithWatchedEpisodeIDs(episode1, episode2).Build()
	staleNextEpisode := newEntry().WithProgramID(programID).WithWatchedEpisodeIDs(episode1).WithNextEpisodeID(episode1).Build()
	wannaWatch := newEntry().WithStatus("wanna_watch").Build()
	updatedBefore := newEntry().WithUpdatedAt(before).Build()
	hasNextEpisode := newEntry().WithNextEpisodeID(episode1).Build()

	// 5件の対象を2件ずつ読み、最後のページの次で読んだ行が無くなる。
	wantPages := []struct {
		lastID       int64
		updatedCount int64
	}{
		{lastID: withoutProgram, updatedCount: 2},
		{lastID: allWatched, updatedCount: 0},
		{lastID: staleNextEpisode, updatedCount: 1},
		{lastID: 0, updatedCount: 0},
	}
	afterID := withProgram - 1
	for i, want := range wantPages {
		lastID, updatedCount, err := repo.UpdateNextResources(ctx, afterID, updatedFrom, now, 2)
		if err != nil {
			t.Fatalf("%dページ目のUpdateNextResourcesのエラー = %v", i+1, err)
		}
		if lastID != want.lastID || updatedCount != want.updatedCount {
			t.Errorf("%dページ目の (最後のid, 更新件数) = (%d, %d)、期待値 = (%d, %d)", i+1, lastID, updatedCount, want.lastID, want.updatedCount)
		}
		afterID = lastID
	}

	episodeID := func(id model.EpisodeID) sql.NullInt64 { return sql.NullInt64{Int64: int64(id), Valid: true} }
	slotID := func(id model.SlotID) sql.NullInt64 { return sql.NullInt64{Int64: int64(id), Valid: true} }
	tests := []struct {
		name string
		id   int64
		want libraryEntryNextResources
	}{
		{name: "番組を選んでいる: 未視聴の最初のエピソードと、それを指す放送枠", id: withProgram, want: libraryEntryNextResources{episodeID(episode2), slotID(slot2), now}},
		{name: "番組を選んでいない: 削除済み・非公開を除いた最初のエピソード", id: withoutProgram, want: libraryEntryNextResources{nextEpisodeID: episodeID(episode1), updatedAt: now}},
		{name: "値が変わらない: 非公開の放送枠は選ばず、更新日時も進めない", id: unchangedSlot, want: libraryEntryNextResources{nextEpisodeID: episodeID(episode1), updatedAt: updatedFrom}},
		{name: "すべて視聴済み: 次のエピソードは無いまま", id: allWatched, want: libraryEntryNextResources{updatedAt: updatedFrom}},
		{name: "次の放送枠が無い: 次のエピソードも求め直す", id: staleNextEpisode, want: libraryEntryNextResources{episodeID(episode2), slotID(slot2), now}},
		{name: "見たい: 対象にしない", id: wannaWatch, want: libraryEntryNextResources{updatedAt: updatedFrom}},
		{name: "更新日時が下限より前: 対象にしない", id: updatedBefore, want: libraryEntryNextResources{updatedAt: before}},
		{name: "番組を選ばず次のエピソードがある: 対象にしない", id: hasNextEpisode, want: libraryEntryNextResources{nextEpisodeID: episodeID(episode1), updatedAt: updatedFrom}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getLibraryEntryNextResources(t, tx, tt.id)
			if got.nextEpisodeID != tt.want.nextEpisodeID || got.nextSlotID != tt.want.nextSlotID || !got.updatedAt.Equal(tt.want.updatedAt) {
				t.Errorf("(次のエピソード, 次の放送枠, 更新日時) = (%v, %v, %v)、期待値 = (%v, %v, %v)",
					got.nextEpisodeID, got.nextSlotID, got.updatedAt, tt.want.nextEpisodeID, tt.want.nextSlotID, tt.want.updatedAt)
			}
		})
	}
}
