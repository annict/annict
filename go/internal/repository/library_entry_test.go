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

// TestLibraryEntryRepository_UpdateNextResources_PreservesConcurrentChangesは、更新SQLが行ロックを
// 待つ間に視聴済み配列か番組が変わった場合、読み取り時点の次のエピソード・放送枠を書き戻さない
// ことを検証する。
func TestLibraryEntryRepository_UpdateNextResources_PreservesConcurrentChanges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		changesWatched bool
	}{
		{name: "視聴済みのエピソードが増えた", changesWatched: true},
		{name: "番組を変更した"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, setupTx := testutil.SetupTx(t)
			// 2100年を下限にするほかのライブラリ更新テストに、このコミット済みの行が入らない
			// ようにする。このテスト自身はidと1件の上限でも対象を絞る。
			now := time.Date(2099, 1, 10, 12, 0, 0, 0, time.UTC)
			updatedFrom := now.AddDate(0, 0, -3)
			changedAt := now.Add(time.Minute)
			userID := testutil.NewUserBuilder(t, setupTx).Build()
			workID := testutil.NewWorkBuilder(t, setupTx).Build()
			episode1 := testutil.NewEpisodeBuilder(t, setupTx, workID).WithSortNumber(10).Build()
			episode2 := testutil.NewEpisodeBuilder(t, setupTx, workID).WithSortNumber(20).Build()
			channelGroupID := testutil.NewChannelGroupBuilder(t, setupTx).Build()
			channelID := testutil.NewChannelBuilder(t, setupTx).WithChannelGroupID(channelGroupID).Build()
			program1 := testutil.NewProgramBuilder(t, setupTx).WithChannelID(channelID).WithWorkID(workID).Build()
			program2 := testutil.NewProgramBuilder(t, setupTx).WithChannelID(channelID).WithWorkID(workID).Build()
			newSlot := func(programID int64, episodeID model.EpisodeID) model.SlotID {
				return testutil.NewSlotBuilder(t, setupTx).WithWorkID(workID).WithChannelID(channelID).
					WithProgramID(programID).WithEpisodeID(episodeID).Build()
			}
			newSlot(program1, episode1)
			slot2 := newSlot(program1, episode2)
			otherProgramSlot := newSlot(program2, episode1)
			entryID := testutil.NewLibraryEntryBuilder(t, setupTx).WithUserID(userID).WithWorkID(workID).
				WithProgramID(program1).WithUpdatedAt(updatedFrom).Build()
			// 競合する2接続から前提データが見えるようにコミットし、終了時に依存行から消す。
			if err := setupTx.Commit(); err != nil {
				t.Fatalf("テストデータのコミットに失敗: %v", err)
			}
			t.Cleanup(func() {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				statements := []struct {
					query string
					id    int64
				}{
					{"DELETE FROM library_entries WHERE id = $1", entryID},
					{"DELETE FROM statuses WHERE user_id = $1", int64(userID)},
					{"DELETE FROM slots WHERE work_id = $1", int64(workID)},
					{"DELETE FROM programs WHERE work_id = $1", int64(workID)},
					{"DELETE FROM episodes WHERE work_id = $1", int64(workID)},
					{"DELETE FROM works WHERE id = $1", int64(workID)},
					{"DELETE FROM channels WHERE id = $1", channelID},
					{"DELETE FROM channel_groups WHERE id = $1", channelGroupID},
				}
				for _, statement := range statements {
					if _, err := db.ExecContext(ctx, statement.query, statement.id); err != nil {
						t.Errorf("競合テストの後始末に失敗 (%s): %v", statement.query, err)
					}
				}
				testutil.DeleteUser(t, db, userID)
			})

			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			updateTx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatalf("更新用トランザクションの開始に失敗: %v", err)
			}
			defer func() { _ = updateTx.Rollback() }()
			var updatePID int
			if err := updateTx.QueryRowContext(ctx, "SELECT pg_backend_pid()").Scan(&updatePID); err != nil {
				t.Fatalf("更新用接続のPIDの取得に失敗: %v", err)
			}
			changeTx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatalf("先行する変更用トランザクションの開始に失敗: %v", err)
			}
			// 失敗時も行ロックを先に解放し、実行中の更新SQLを待つRollbackを止めない。
			defer func() { _ = changeTx.Rollback() }()
			var changePID int
			if err := changeTx.QueryRowContext(ctx, "SELECT pg_backend_pid()").Scan(&changePID); err != nil {
				t.Fatalf("変更用接続のPIDの取得に失敗: %v", err)
			}

			wantEpisode, wantSlot := episode1, otherProgramSlot
			if tt.changesWatched {
				_, err = changeTx.ExecContext(ctx, `
					UPDATE library_entries
					SET watched_episode_ids = ARRAY[$2::bigint], next_episode_id = $3,
					    next_slot_id = $4, updated_at = $5
					WHERE id = $1`, entryID, int64(episode1), int64(episode2), int64(slot2), changedAt)
				wantEpisode, wantSlot = episode2, slot2
			} else {
				_, err = changeTx.ExecContext(ctx, `
					UPDATE library_entries
					SET program_id = $2, next_episode_id = $3, next_slot_id = $4, updated_at = $5
					WHERE id = $1`, entryID, program2, int64(episode1), int64(otherProgramSlot), changedAt)
			}
			if err != nil {
				t.Fatalf("先行するライブラリの変更に失敗: %v", err)
			}

			type updateResult struct {
				lastID       int64
				updatedCount int64
				err          error
			}
			resultCh := make(chan updateResult, 1)
			repo := repository.NewLibraryEntryRepository(query.New(db)).WithTx(updateTx)
			go func() {
				lastID, updatedCount, err := repo.UpdateNextResources(ctx, entryID-1, updatedFrom, now, 1)
				resultCh <- updateResult{lastID: lastID, updatedCount: updatedCount, err: err}
			}()

			// CTEが変更前の行を読んでからUPDATEで待機したことを確認し、先行する変更を確定する。
			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()
			waiting := false
			for !waiting {
				select {
				case result := <-resultCh:
					t.Fatalf("先行する変更のコミット前に更新が完了: %+v", result)
				case <-ticker.C:
					if err := changeTx.QueryRowContext(ctx, "SELECT $1 = ANY(pg_blocking_pids($2))", changePID, updatePID).Scan(&waiting); err != nil {
						t.Fatalf("更新SQLのロック待ちの確認に失敗: %v", err)
					}
				case <-ctx.Done():
					t.Fatalf("更新SQLが行ロックを待たなかった: %v", ctx.Err())
				}
			}
			if err := changeTx.Commit(); err != nil {
				t.Fatalf("先行する変更のコミットに失敗: %v", err)
			}
			select {
			case result := <-resultCh:
				if result.err != nil {
					t.Fatalf("UpdateNextResources()のエラー = %v", result.err)
				}
				if result.lastID != entryID || result.updatedCount != 0 {
					t.Errorf("(最後のid, 更新件数) = (%d, %d)、期待値 = (%d, 0)", result.lastID, result.updatedCount, entryID)
				}
			case <-ctx.Done():
				t.Fatalf("先行する変更のコミット後も更新が完了しない: %v", ctx.Err())
			}
			got := getLibraryEntryNextResources(t, updateTx, entryID)
			if got.nextEpisodeID != (sql.NullInt64{Int64: int64(wantEpisode), Valid: true}) ||
				got.nextSlotID != (sql.NullInt64{Int64: int64(wantSlot), Valid: true}) || !got.updatedAt.Equal(changedAt) {
				t.Errorf("(次のエピソード, 次の放送枠, 更新日時) = (%v, %v, %v)、期待値 = (%d, %d, %v)",
					got.nextEpisodeID, got.nextSlotID, got.updatedAt, wantEpisode, wantSlot, changedAt)
			}
		})
	}
}
