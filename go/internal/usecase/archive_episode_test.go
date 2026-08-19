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

// newArchiveEpisodeUsecase wires the archive-episode usecase against the shared test DB. Like
// the create / update usecases it opens its own transaction, so its tests use GetTestDB (not
// SetupTx) so the committed rows are visible to the usecase's inner transaction and to the
// follow-up sync invariant check.
//
// [Ja] newArchiveEpisodeUsecase は共有テスト DB に対してエピソード非公開 UseCase を組み立てる。
// 作成 / 更新 UseCase と同じく内部で自前のトランザクションを開くため、テストは SetupTx ではなく
// GetTestDB を使い、コミット済みの行が UseCase の内側トランザクションと後続の同期不変条件
// チェックから見えるようにする。
func newArchiveEpisodeUsecase(db *sql.DB) *ArchiveEpisodeUsecase {
	queries := query.New(db)
	return NewArchiveEpisodeUsecase(
		db,
		repository.NewEpisodeRepository(queries),
		repository.NewAnimeRepository(queries),
	)
}

// readArchivedEpisodeState returns the state column the archive and the re-publish write, so a
// test can tell an episode either of them changed from one the submit left alone.
//
// [Ja] readArchivedEpisodeState は非公開と再公開が書く状態カラムを返す。どちらかが変更した
// エピソードと、送信が手を触れなかったエピソードをテストが区別できるようにするため。
func readArchivedEpisodeState(t *testing.T, db *sql.DB, episodeID model.EpisodeID) sql.NullTime {
	t.Helper()

	var unpublishedAt sql.NullTime
	if err := db.QueryRow(`SELECT unpublished_at FROM episodes WHERE id = $1`, int64(episodeID)).Scan(&unpublishedAt); err != nil {
		t.Fatalf("エピソードの状態の読み込みに失敗: %v", err)
	}

	return unpublishedAt
}

// awaitBlockedByBackend waits until some backend is blocked on a lock the backend with
// blockerPID holds, so an interleaving test can commit the conflicting change knowing the write
// under test already started and is waiting. pg_blocking_pids makes the order deterministic
// without a sleep or access to the usecase's private transaction. resultCh is watched only to
// fail fast: a usecase that finished before blocking never reached the interleaving the test
// means to fix, so asserting on its result would fix nothing.
//
// [Ja] awaitBlockedByBackend は、blockerPID のバックエンドが保持するロックを待っている
// バックエンドが現れるまで待機する。検証対象の書き込みが既に開始して待機していることを確かめて
// から、インターリーブさせたい変更をコミットできるようにするため。pg_blocking_pids により、
// UseCase 内部の transaction へ触れたり sleep に依存したりせず実行順を固定する。resultCh は
// 早期失敗のためだけに監視する。待機に入る前に終わった UseCase はテストが固定したいインター
// リーブに到達していないため、その結果に対するアサーションは何も固定しない。
func awaitBlockedByBackend[T any](t *testing.T, ctx context.Context, db *sql.DB, blockerPID int, resultCh <-chan T) {
	t.Helper()

	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case result := <-resultCh:
			t.Fatalf("Execute() completed before blocker committed: %+v", result)
		case <-ticker.C:
			var blocked bool
			if err := db.QueryRowContext(ctx, `
				SELECT EXISTS (
					SELECT 1
					FROM pg_stat_activity activity
					WHERE $1 = ANY(pg_blocking_pids(activity.pid))
						AND activity.wait_event_type = 'Lock'
				)`,
				blockerPID,
			).Scan(&blocked); err != nil {
				t.Fatalf("待機状態の取得に失敗: %v", err)
			}
			if blocked {
				return
			}
		case <-deadline.C:
			t.Fatal("Execute() did not wait for the episode row lock")
		}
	}
}

// TestArchiveEpisodeUsecase_Execute_ArchivesEpisodeAndAnime verifies archiving a mapped,
// published episode sets episodes.unpublished_at (the state source of truth) and dual-writes
// the derived anime.status = archived, and that a phase 2 sync right after reports Unchanged
// (the archive and the reconciliation derive the same status from unpublished_at, so the sync
// does not clobber the archived anime back to published).
//
// [Ja] TestArchiveEpisodeUsecase_Execute_ArchivesEpisodeAndAnime は、マッピング済みで公開中の
// エピソードを非公開にすると episodes.unpublished_at (状態の正本) が立ち、導出された
// anime.status = archived が両書きされること、および直後のフェーズ 2 同期が Unchanged を報告する
// ことを検証する (非公開とリコンシリエーションが unpublished_at から同じ status を導出するため、
// 同期はアーカイブ済み anime を published に戻さない)。
func TestArchiveEpisodeUsecase_Execute_ArchivesEpisodeAndAnime(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newArchiveEpisodeUsecase(db)

	workID, parentAnimeID := insertMappedCreateTargetWork(t, db)
	episodeID, episodeAnimeID := insertMappedUpdateTargetEpisode(t, db, workID, parentAnimeID, 100)
	if _, err := db.Exec(`
		UPDATE animes
		SET
			title = '編集前のタイトル',
			title_ro = 'Before',
			title_en = 'Before EN',
			archive_message = '非公開前のメッセージ'
		WHERE id = $1`,
		int64(episodeAnimeID),
	); err != nil {
		t.Fatalf("同期済み anime の準備に失敗: %v", err)
	}

	output, err := uc.Execute(context.Background(), ArchiveEpisodeInput{EpisodeID: episodeID, User: unsavedCreateActor()})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if output.EpisodeID != episodeID || output.WorkID != workID {
		t.Errorf("output = %+v, want {EpisodeID:%d WorkID:%d}", output, int64(episodeID), int64(workID))
	}

	if unpublishedAt := readArchivedEpisodeState(t, db, episodeID); !unpublishedAt.Valid {
		t.Error("episodes.unpublished_at = NULL, want 非公開の時刻")
	}

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), episodeAnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Status != model.AnimeStatusArchived {
		t.Errorf("anime.Status = %q, want %q", anime.Status, model.AnimeStatusArchived)
	}
	// The archive maps status alone, so anime-owned content stays byte-for-byte unchanged.
	//
	// [Ja] 非公開が写像するのは status だけなので、anime 固有の内容はそのまま保持される。
	if anime.ArchiveMessage.String != "非公開前のメッセージ" {
		t.Errorf("anime.ArchiveMessage = %q, want %q", anime.ArchiveMessage.String, "非公開前のメッセージ")
	}

	syncUC := newSyncEpisodesUsecase(db)
	result, err := syncUC.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
	if err != nil {
		t.Fatalf("sync Execute() error = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.Updated != 0 || result.Unchanged != 1 {
		t.Fatalf("sync result = %+v, want {Processed:1 Created:0 Updated:0 Unchanged:1}", result)
	}
}

// TestArchiveEpisodeUsecase_Execute_SkipsAnimeForUnmappedEpisode covers an episode with no anime
// yet: only the episodes row is written, and the phase 2 sync creates the anime later with the
// status the archived episode now derives.
//
// [Ja] TestArchiveEpisodeUsecase_Execute_SkipsAnimeForUnmappedEpisode は、まだ anime を持たない
// エピソードを検証する。書かれるのは episodes の行だけで、anime は後でフェーズ 2 の同期が、
// 非公開になったエピソードが導出する status で作成する。
func TestArchiveEpisodeUsecase_Execute_SkipsAnimeForUnmappedEpisode(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newArchiveEpisodeUsecase(db)

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	episodeID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 100)

	if _, err := uc.Execute(context.Background(), ArchiveEpisodeInput{EpisodeID: episodeID, User: unsavedCreateActor()}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if unpublishedAt := readArchivedEpisodeState(t, db, episodeID); !unpublishedAt.Valid {
		t.Error("episodes.unpublished_at = NULL, want 非公開の時刻")
	}

	var animeID sql.NullInt64
	if err := db.QueryRow(`SELECT anime_id FROM episodes WHERE id = $1`, int64(episodeID)).Scan(&animeID); err != nil {
		t.Fatalf("episodes.anime_id の読み込みに失敗: %v", err)
	}
	if animeID.Valid {
		t.Errorf("episodes.anime_id = %d, want NULL のまま", animeID.Int64)
	}
}

// TestArchiveEpisodeUsecase_Execute_ArchivesEpisodeWithUnmappedParent covers an episode that
// carries an anime while its parent work no longer does. Archiving maps the status alone, which
// needs no parent_anime_id, so the anime follows the episode instead of waiting for the parent
// to be mapped again.
//
// [Ja] TestArchiveEpisodeUsecase_Execute_ArchivesEpisodeWithUnmappedParent は、エピソード自身は
// anime を持つが親作品が持たなくなった場合を検証する。非公開が写像するのは status だけで
// parent_anime_id を必要としないため、anime は親が再びマッピングされるのを待たずエピソードに
// 追従する。
func TestArchiveEpisodeUsecase_Execute_ArchivesEpisodeWithUnmappedParent(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newArchiveEpisodeUsecase(db)

	workID, parentAnimeID := insertMappedCreateTargetWork(t, db)
	episodeID, episodeAnimeID := insertMappedUpdateTargetEpisode(t, db, workID, parentAnimeID, 100)
	if _, err := db.Exec(`UPDATE works SET anime_id = NULL WHERE id = $1`, int64(workID)); err != nil {
		t.Fatalf("親作品の写像の解除に失敗: %v", err)
	}

	if _, err := uc.Execute(context.Background(), ArchiveEpisodeInput{EpisodeID: episodeID, User: unsavedCreateActor()}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), episodeAnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Status != model.AnimeStatusArchived {
		t.Errorf("anime.Status = %q, want %q", anime.Status, model.AnimeStatusArchived)
	}
}

// TestArchiveEpisodeUsecase_Execute_RequiresCommitter verifies authorization belongs to the
// write usecase as well as the HTTP boundary. Rejected callers cannot archive the episode;
// an editor can.
//
// [Ja] TestArchiveEpisodeUsecase_Execute_RequiresCommitter は認可が HTTP 境界だけでなく書き込み
// UseCase にも属することを検証する。拒否された呼び出し元はエピソードを非公開にできず、編集者は
// 実行できる。
func TestArchiveEpisodeUsecase_Execute_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newArchiveEpisodeUsecase(db)
	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	episodeID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 100)

	tests := []struct {
		name          string
		user          *model.User
		wantForbidden bool
	}{
		{name: "未認証", user: nil, wantForbidden: true},
		{name: "一般ユーザー", user: &model.User{ID: 1, Role: model.RoleUser}, wantForbidden: true},
		{name: "編集者", user: &model.User{ID: 1, Role: model.RoleEditor}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.Execute(context.Background(), ArchiveEpisodeInput{
				EpisodeID: episodeID,
				User:      tt.user,
			})
			if tt.wantForbidden {
				appErr := model.AsAppError(err)
				if appErr == nil || appErr.Code != model.AppErrCodeForbidden {
					t.Fatalf("Execute() error = %v, want AppErrCodeForbidden", err)
				}
				if unpublishedAt := readArchivedEpisodeState(t, db, episodeID); unpublishedAt.Valid {
					t.Error("拒否された送信が episodes.unpublished_at を更新しました")
				}
				return
			}
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
		})
	}
}

// TestArchiveEpisodeUsecase_Execute_UsesMappingFromArchivedRowAndPreservesConcurrentAnimeEdit
// fixes the interleaving between the confirmation projection and the archive write. The
// archive waits on a locked episode while another transaction changes its anime mapping, and
// an anime-only edit commits in that window. Once released, the status write must follow the
// mapping returned by the updated episode and must preserve the concurrent content.
//
// [Ja] このテストは確認用の射影と非公開の書き込みの間の実行順を固定する。非公開がロック済みの
// episode を待つ間に
// 別トランザクションが anime の写像を変え、その窓で anime 固有の編集もコミットする。解放後の
// status 書き込みは更新した episode が返す写像へ追従し、競合した内容を保持しなければならない。
func TestArchiveEpisodeUsecase_Execute_UsesMappingFromArchivedRowAndPreservesConcurrentAnimeEdit(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newArchiveEpisodeUsecase(db)
	workID, parentAnimeID := insertMappedCreateTargetWork(t, db)
	episodeID, formerAnimeID := insertMappedUpdateTargetEpisode(t, db, workID, parentAnimeID, 100)
	currentAnimeID := insertBareAnime(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	blockerTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("blocker transaction BeginTx() error = %v", err)
	}
	defer func() { _ = blockerTx.Rollback() }()

	var blockerPID int
	if err := blockerTx.QueryRowContext(ctx, `SELECT pg_backend_pid()`).Scan(&blockerPID); err != nil {
		t.Fatalf("blocker backend PID の取得に失敗: %v", err)
	}
	if _, err := blockerTx.ExecContext(ctx, `SELECT id FROM episodes WHERE id = $1 FOR UPDATE`, int64(episodeID)); err != nil {
		t.Fatalf("episode のロック取得に失敗: %v", err)
	}

	type archiveResult struct {
		output *ArchiveEpisodeOutput
		err    error
	}
	resultCh := make(chan archiveResult, 1)
	go func() {
		output, err := uc.Execute(ctx, ArchiveEpisodeInput{
			EpisodeID: episodeID,
			User:      unsavedCreateActor(),
		})
		resultCh <- archiveResult{output: output, err: err}
	}()

	// Observe the archive waiting on blockerTx before changing the mapping.
	//
	// [Ja] 写像を変更する前に、非公開が blockerTx を待っていることを観測する。
	awaitBlockedByBackend(t, ctx, db, blockerPID, resultCh)

	if _, err := blockerTx.ExecContext(
		ctx,
		`UPDATE episodes SET anime_id = $1 WHERE id = $2`,
		int64(currentAnimeID),
		int64(episodeID),
	); err != nil {
		t.Fatalf("episode の anime 写像変更に失敗: %v", err)
	}
	if _, err := db.ExecContext(
		ctx,
		`UPDATE animes SET archive_message = '同時編集後のメッセージ' WHERE id = $1`,
		int64(currentAnimeID),
	); err != nil {
		t.Fatalf("anime 固有属性の同時編集に失敗: %v", err)
	}
	if err := blockerTx.Commit(); err != nil {
		t.Fatalf("blocker transaction Commit() error = %v", err)
	}

	var result archiveResult
	select {
	case result = <-resultCh:
	case <-ctx.Done():
		t.Fatalf("Execute() did not finish after blocker committed: %v", ctx.Err())
	}
	if result.err != nil {
		t.Fatalf("Execute() error = %v", result.err)
	}
	if result.output == nil || result.output.EpisodeID != episodeID || result.output.WorkID != workID {
		t.Fatalf("Execute() output = %+v, want episode %d work %d", result.output, int64(episodeID), int64(workID))
	}

	animeRepo := repository.NewAnimeRepository(query.New(db))
	formerAnime, err := animeRepo.GetByID(ctx, formerAnimeID)
	if err != nil || formerAnime == nil {
		t.Fatalf("以前の anime の GetByID() anime=%v err=%v", formerAnime, err)
	}
	if formerAnime.Status != model.AnimeStatusPublished {
		t.Errorf("以前の anime.Status = %q, want %q", formerAnime.Status, model.AnimeStatusPublished)
	}

	currentAnime, err := animeRepo.GetByID(ctx, currentAnimeID)
	if err != nil || currentAnime == nil {
		t.Fatalf("現在の anime の GetByID() anime=%v err=%v", currentAnime, err)
	}
	if currentAnime.Status != model.AnimeStatusArchived {
		t.Errorf("現在の anime.Status = %q, want %q", currentAnime.Status, model.AnimeStatusArchived)
	}
	if currentAnime.ArchiveMessage.String != "同時編集後のメッセージ" {
		t.Errorf(
			"現在の anime.ArchiveMessage = %q, want %q",
			currentAnime.ArchiveMessage.String,
			"同時編集後のメッセージ",
		)
	}
}

// TestArchiveEpisodeUsecase_Execute_RollsBackWhenParentIsDeletedWhileEpisodeWriteWaits fixes the
// interleaving between the confirmation projection and the archive write. The archive waits on a
// locked episode after its pre-read, while the locking transaction deletes the parent work. The
// work guard must make the write report not found, and the usecase rollback must preserve the
// published episode, counter, and anime status. It is the mirror of the re-publish test of the
// same name: the two statements carry the guard separately, so each direction needs its own
// interleaving to detect a regression in its own SQL.
//
// [Ja] このテストは確認用の射影と非公開の書き込みの間の実行順を固定する。事前読み取り後の非公開
// がロック済みの episode を待つ間に、ロック元のトランザクションが親作品を削除する。作品のガード
// により not found を返し、UseCase のロールバックにより公開中の episode、カウンター、anime の
// 状態を保持しなければならない。同名の再公開テストと対になる。ガードは 2 つのステートメントが
// それぞれ持つため、各方向が自分の SQL の退行を検出するインターリーブを必要とする。
func TestArchiveEpisodeUsecase_Execute_RollsBackWhenParentIsDeletedWhileEpisodeWriteWaits(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newArchiveEpisodeUsecase(db)
	workID, parentAnimeID := insertMappedCreateTargetWork(t, db)
	episodeID, episodeAnimeID := insertMappedUpdateTargetEpisode(t, db, workID, parentAnimeID, 100)

	var episodesCountBefore int32
	if err := db.QueryRow(
		`SELECT episodes_count FROM works WHERE id = $1`,
		int64(workID),
	).Scan(&episodesCountBefore); err != nil {
		t.Fatalf("作品のカウンターの読み込みに失敗: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	blockerTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("blocker transaction BeginTx() error = %v", err)
	}
	defer func() { _ = blockerTx.Rollback() }()

	var blockerPID int
	if err := blockerTx.QueryRowContext(ctx, `SELECT pg_backend_pid()`).Scan(&blockerPID); err != nil {
		t.Fatalf("blocker backend PID の取得に失敗: %v", err)
	}
	if _, err := blockerTx.ExecContext(
		ctx,
		`SELECT id FROM episodes WHERE id = $1 FOR UPDATE`,
		int64(episodeID),
	); err != nil {
		t.Fatalf("episode のロック取得に失敗: %v", err)
	}

	type archiveResult struct {
		output *ArchiveEpisodeOutput
		err    error
	}
	resultCh := make(chan archiveResult, 1)
	go func() {
		output, err := uc.Execute(ctx, ArchiveEpisodeInput{
			EpisodeID: episodeID,
			User:      unsavedCreateActor(),
		})
		resultCh <- archiveResult{output: output, err: err}
	}()

	// Observe the archive waiting on blockerTx before deleting the work. This proves the
	// pre-read completed and the guarded write started.
	//
	// [Ja] 作品を削除する前に、非公開が blockerTx を待っていることを観測する。事前読み取りが完了
	// してガード付きの書き込みが始まったことを証明する。
	awaitBlockedByBackend(t, ctx, db, blockerPID, resultCh)

	if _, err := blockerTx.ExecContext(
		ctx,
		`UPDATE works SET deleted_at = NOW() WHERE id = $1`,
		int64(workID),
	); err != nil {
		t.Fatalf("親作品の削除に失敗: %v", err)
	}
	if err := blockerTx.Commit(); err != nil {
		t.Fatalf("blocker transaction Commit() error = %v", err)
	}

	var result archiveResult
	select {
	case result = <-resultCh:
	case <-ctx.Done():
		t.Fatalf("Execute() did not finish after blocker committed: %v", ctx.Err())
	}
	if result.output != nil {
		t.Fatalf("Execute() output = %+v, want nil", result.output)
	}
	appErr := model.AsAppError(result.err)
	if appErr == nil || appErr.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("Execute() error = %v, want AppErrCodeResourceNotFound", result.err)
	}

	if unpublishedAt := readArchivedEpisodeState(t, db, episodeID); unpublishedAt.Valid {
		t.Errorf("episodes.unpublished_at = %v, want NULL のまま", unpublishedAt.Time)
	}
	var episodesCountAfter int32
	if err := db.QueryRow(
		`SELECT episodes_count FROM works WHERE id = $1`,
		int64(workID),
	).Scan(&episodesCountAfter); err != nil {
		t.Fatalf("作品のカウンターの読み込みに失敗: %v", err)
	}
	if episodesCountAfter != episodesCountBefore {
		t.Errorf("works.episodes_count = %d, want %d", episodesCountAfter, episodesCountBefore)
	}

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(ctx, episodeAnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Status != model.AnimeStatusPublished {
		t.Errorf("anime.Status = %q, want %q", anime.Status, model.AnimeStatusPublished)
	}
}

// TestArchiveEpisodeUsecase_Execute_NotFound covers the submits the confirmation page cannot be
// shown for: an episode that never existed, one already archived, a deleted one, and one whose
// work was deleted. All four are reported as not found, which the handler turns into the same
// 404 the page itself gives.
//
// [Ja] TestArchiveEpisodeUsecase_Execute_NotFound は、確認ページを出せない送信を検証する。存在
// しなかったエピソード、すでに非公開のエピソード、削除済みのエピソード、作品が削除された
// エピソードの 4 つ。いずれも not found として報告され、Handler はそれをページ自身が返すのと同じ
// 404 に変換する。
func TestArchiveEpisodeUsecase_Execute_NotFound(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newArchiveEpisodeUsecase(db)

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	archivedID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 100)
	if _, err := db.Exec(`UPDATE episodes SET unpublished_at = NOW() WHERE id = $1`, int64(archivedID)); err != nil {
		t.Fatalf("エピソードの非公開化に失敗: %v", err)
	}
	deletedID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 200)
	if _, err := db.Exec(`UPDATE episodes SET deleted_at = NOW() WHERE id = $1`, int64(deletedID)); err != nil {
		t.Fatalf("エピソードの削除に失敗: %v", err)
	}

	deletedWorkID := insertCreateTargetWork(t, db, sql.NullInt64{})
	orphanID := insertUpdateTargetEpisode(t, db, deletedWorkID, sql.NullInt64{}, 100)
	if _, err := db.Exec(`UPDATE works SET deleted_at = NOW() WHERE id = $1`, int64(deletedWorkID)); err != nil {
		t.Fatalf("作品の削除に失敗: %v", err)
	}

	tests := map[string]model.EpisodeID{
		"存在しないエピソード":   model.EpisodeID(-1),
		"非公開済みのエピソード":  archivedID,
		"削除済みのエピソード":   deletedID,
		"削除済み作品のエピソード": orphanID,
	}
	for name, episodeID := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := uc.Execute(context.Background(), ArchiveEpisodeInput{EpisodeID: episodeID, User: unsavedCreateActor()})
			appErr := model.AsAppError(err)
			if appErr == nil || appErr.Code != model.AppErrCodeResourceNotFound {
				t.Fatalf("Execute() error = %v, want AppErrCodeResourceNotFound", err)
			}
		})
	}
}
