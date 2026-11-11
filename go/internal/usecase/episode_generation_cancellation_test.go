package usecase

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// cancelBlockedEpisodeGenerationは、子ユースケースのDB処理がblockerのロックを待つことを
// 確かめてからctxを取り消す。観測もblockerの接続を使い、共有プールの接続を増やさない。
func cancelBlockedEpisodeGeneration[T any](t *testing.T, ctx context.Context, blocker *sql.Tx, cancel context.CancelFunc, resultCh <-chan T) T {
	t.Helper()
	var blockerPID int
	if err := blocker.QueryRowContext(ctx, "SELECT pg_backend_pid()").Scan(&blockerPID); err != nil {
		t.Fatalf("ロック用接続のPIDの取得に失敗: %v", err)
	}
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case result := <-resultCh:
			t.Fatalf("ロック待ちより前に処理が完了: %+v", result)
		case <-ctx.Done():
			t.Fatalf("DB処理がロックを待たなかった: %v", ctx.Err())
		case <-ticker.C:
			var blocked bool
			if err := blocker.QueryRowContext(ctx, `
				SELECT EXISTS (
					SELECT 1 FROM pg_stat_activity
					WHERE $1 = ANY(pg_blocking_pids(pid))
				)`, blockerPID).Scan(&blocked); err != nil {
				t.Fatalf("DB処理のロック待ちの確認に失敗: %v", err)
			}
			if blocked {
				cancel()
				select {
				case result := <-resultCh:
					return result
				case <-ctx.Done():
					t.Fatalf("ctxの取り消し後もDB処理が完了しない: %v", ctx.Err())
				}
			}
		}
	}
}

// TestCreateEpisodesFromSlotsUsecase_Execute_CancelsDuringWriteは、先の放送枠の作成を保持し、
// 分類の外部キー検査で待っていた放送枠とその後を未処理として返すことを検証する。
func TestCreateEpisodesFromSlotsUsecase_Execute_CancelsDuringWrite(t *testing.T) {
	t.Parallel()
	db := testutil.GetTestDB()
	now := generationNow()
	goodWorkID, _ := insertMappedCreateTargetWork(t, db)
	goodProgram := insertGenerationProgram(t, db, goodWorkID, 1)
	goodSlot := goodProgram.insertSlot(1, now, nil)
	workID, parentAnimeID := insertMappedCreateTargetWork(t, db)
	program := insertGenerationProgram(t, db, workID, 1)
	blockedSlot := program.insertSlot(1, now.Add(time.Hour), nil)
	laterSlot := program.insertSlot(2, now.Add(2*time.Hour), nil)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	blocker, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("ロック用トランザクションの開始に失敗: %v", err)
	}
	defer func() { _ = blocker.Rollback() }()
	var id int64
	// 分類の挿入が親animeのキー共有ロックを取る前に、排他の行ロックを保持する。
	if err := blocker.QueryRowContext(ctx, "SELECT id FROM animes WHERE id = $1 FOR UPDATE", int64(parentAnimeID)).Scan(&id); err != nil {
		t.Fatalf("親animeのロックに失敗: %v", err)
	}
	workCtx, cancelWork := context.WithCancel(ctx)
	defer cancelWork()
	type result struct {
		output *CreateEpisodesFromSlotsOutput
		err    error
	}
	results := make(chan result, 1)
	go func() {
		output, err := newCreateEpisodesFromSlotsUsecase(db).Execute(workCtx, CreateEpisodesFromSlotsInput{Now: now})
		results <- result{output, err}
	}()
	got := cancelBlockedEpisodeGeneration(t, ctx, blocker, cancelWork, results)
	if got.err != nil || got.output == nil {
		t.Fatalf("結果 = %+v、エラー = %v、期待値 = 中断時の途中結果", got.output, got.err)
	}
	if len(got.output.CreatedEpisodes) != 1 || got.output.CreatedEpisodes[0].SlotID != goodSlot ||
		got.output.UnprocessedSlotCount != 2 || len(got.output.FailedSlots) != 0 {
		t.Errorf("結果 = %+v、期待値 = 作成1件・未処理2件・個別エラーなし", got.output)
	}
	if !readSlotEpisodeID(t, db, goodSlot).Valid || countWorkEpisodes(t, db, goodWorkID) != 1 {
		t.Error("中断前にコミットした放送枠の結果が保持されていません")
	}
	if readSlotEpisodeID(t, db, blockedSlot).Valid || readSlotEpisodeID(t, db, laterSlot).Valid || countWorkEpisodes(t, db, workID) != 0 {
		t.Error("中断した放送枠のエピソードや紐づけが残っています")
	}
	var classifications int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM anime_classifications WHERE parent_anime_id = $1", int64(parentAnimeID)).Scan(&classifications); err != nil {
		t.Fatalf("分類の読み込みに失敗: %v", err)
	}
	if classifications != 0 {
		t.Errorf("分類の件数 = %d、期待値 = 0 (ロールバック)", classifications)
	}
}

// TestFillEpisodeTitlesUsecase_Execute_CancelsDuringWriteは、episodeの更新後にanimeの更新が
// 中断した場合、その両書きを戻し、先に確定したタイトルだけを保持することを検証する。
func TestFillEpisodeTitlesUsecase_Execute_CancelsDuringWrite(t *testing.T) {
	t.Parallel()
	db := testutil.GetTestDB()
	workID, tid := insertTitleFillWork(t, db)
	program := insertGenerationProgram(t, db, workID, 1)
	firstID := insertTitleFillEpisode(t, db, program, 1)
	blockedID := insertTitleFillEpisode(t, db, program, 2)
	animeID := insertBareAnime(t, db)
	if _, err := db.Exec("UPDATE episodes SET anime_id = $2 WHERE id = $1", int64(blockedID), int64(animeID)); err != nil {
		t.Fatalf("animeの紐づけに失敗: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	blocker, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("ロック用トランザクションの開始に失敗: %v", err)
	}
	defer func() { _ = blocker.Rollback() }()
	if _, err := blocker.ExecContext(ctx, "UPDATE animes SET title = title WHERE id = $1", int64(animeID)); err != nil {
		t.Fatalf("animeのロックに失敗: %v", err)
	}
	workCtx, cancelWork := context.WithCancel(ctx)
	defer cancelWork()
	lookup := &fakeSubTitleLookup{subTitles: map[int32]map[int32]string{tid: {1: "先に確定するタイトル", 2: "戻るタイトル"}}}
	type result struct {
		output *FillEpisodeTitlesOutput
		err    error
	}
	results := make(chan result, 1)
	go func() {
		output, err := newFillEpisodeTitlesUsecase(db, lookup).Execute(workCtx, FillEpisodeTitlesInput{Now: time.Now()})
		results <- result{output, err}
	}()
	got := cancelBlockedEpisodeGeneration(t, ctx, blocker, cancelWork, results)
	if got.err != nil || got.output == nil {
		t.Fatalf("結果 = %+v、エラー = %v、期待値 = 中断時の途中結果", got.output, got.err)
	}
	if got.output.FilledCount != 1 || got.output.UnprocessedEpisodeCount < 1 || len(got.output.FailedEpisodes) != 0 {
		t.Errorf("結果 = %+v、期待値 = 更新1件・未処理1件以上・個別エラーなし", got.output)
	}
	if title, _ := readEpisodeTitle(t, db, firstID); title.String != "先に確定するタイトル" {
		t.Errorf("先に確定したタイトル = %q、期待値 = 先に確定するタイトル", title.String)
	}
	if title, _ := readEpisodeTitle(t, db, blockedID); title.Valid {
		t.Errorf("中断したepisodeのタイトル = %q、期待値 = NULL (ロールバック)", title.String)
	}
	var animeTitle sql.NullString
	if err := db.QueryRowContext(ctx, "SELECT title FROM animes WHERE id = $1", int64(animeID)).Scan(&animeTitle); err != nil {
		t.Fatalf("animeのタイトルの取得に失敗: %v", err)
	}
	if animeTitle.Valid {
		t.Errorf("中断したanimeのタイトル = %q、期待値 = NULL", animeTitle.String)
	}
}

// TestUpdateLibraryEntryNextResourcesUsecase_Execute_CancelsDuringWriteは、ページのUPDATE中に
// ctxが終わった場合、個別エラーにせず中断を返し、ページの書き込みを戻すことを検証する。
func TestUpdateLibraryEntryNextResourcesUsecase_Execute_CancelsDuringWrite(t *testing.T) {
	t.Parallel()
	db, setupTx := testutil.SetupTx(t)
	// 2100年のテストや2099年1月の競合テストと、コミットする行の対象期間を分ける。
	now := time.Date(2099, 12, 31, 12, 0, 0, 0, time.UTC)
	userID := testutil.NewUserBuilder(t, setupTx).Build()
	workID := testutil.NewWorkBuilder(t, setupTx).Build()
	testutil.NewEpisodeBuilder(t, setupTx, workID).Build()
	entryID := testutil.NewLibraryEntryBuilder(t, setupTx).WithUserID(userID).WithWorkID(workID).WithUpdatedAt(now).Build()
	if err := setupTx.Commit(); err != nil {
		t.Fatalf("テストデータのコミットに失敗: %v", err)
	}
	t.Cleanup(func() {
		for _, statement := range []struct {
			sql string
			id  int64
		}{
			{"DELETE FROM library_entries WHERE id = $1", entryID},
			{"DELETE FROM statuses WHERE user_id = $1", int64(userID)},
			{"DELETE FROM episodes WHERE work_id = $1", int64(workID)},
			{"DELETE FROM works WHERE id = $1", int64(workID)},
		} {
			if _, err := db.Exec(statement.sql, statement.id); err != nil {
				t.Errorf("中断テストの後始末に失敗 (%s): %v", statement.sql, err)
			}
		}
		testutil.DeleteUser(t, db, userID)
	})
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	blocker, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("ロック用トランザクションの開始に失敗: %v", err)
	}
	defer func() { _ = blocker.Rollback() }()
	var id int64
	if err := blocker.QueryRowContext(ctx, "SELECT id FROM library_entries WHERE id = $1 FOR UPDATE", entryID).Scan(&id); err != nil {
		t.Fatalf("ライブラリのロックに失敗: %v", err)
	}
	workCtx, cancelWork := context.WithCancel(ctx)
	defer cancelWork()
	uc := NewUpdateLibraryEntryNextResourcesUsecase(repository.NewLibraryEntryRepository(query.New(db)))
	type result struct {
		output *UpdateLibraryEntryNextResourcesOutput
		err    error
	}
	results := make(chan result, 1)
	go func() {
		output, err := uc.Execute(workCtx, UpdateLibraryEntryNextResourcesInput{Now: now})
		results <- result{output, err}
	}()
	got := cancelBlockedEpisodeGeneration(t, ctx, blocker, cancelWork, results)
	if got.err != nil || got.output == nil || !got.output.Interrupted || got.output.UpdatedCount != 0 {
		t.Fatalf("結果 = %+v、エラー = %v、期待値 = 更新0件で中断・エラーなし", got.output, got.err)
	}
	var nextEpisodeID sql.NullInt64
	if err := db.QueryRowContext(ctx, "SELECT next_episode_id FROM library_entries WHERE id = $1", entryID).Scan(&nextEpisodeID); err != nil {
		t.Fatalf("ライブラリの再読み込みに失敗: %v", err)
	}
	if nextEpisodeID.Valid {
		t.Errorf("中断したページの次のエピソード = %v、期待値 = NULL", nextEpisodeID)
	}
}
