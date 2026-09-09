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

// newUnarchiveEpisodeUsecase wires the re-publish usecase against the shared test DB, on the
// pool rather than on a test transaction for the reason newArchiveEpisodeUsecase states.
//
// [Ja] newUnarchiveEpisodeUsecase は共有テスト DB に対してエピソード再公開 UseCase を組み立てる。
// newArchiveEpisodeUsecase が述べる理由により、テスト用トランザクションではなくプールに対して
// 組み立てる。
func newUnarchiveEpisodeUsecase(db *sql.DB) *UnarchiveEpisodeUsecase {
	queries := query.New(db)
	return NewUnarchiveEpisodeUsecase(
		db,
		repository.NewEpisodeRepository(queries),
		repository.NewAnimeRepository(queries),
	)
}

// archiveFixtureEpisode puts an inserted episode into the state a re-publish starts from. The
// episodes fixtures are created published, matching how the bulk create leaves them, so the
// tests of the opposite direction unpublish them first.
//
// [Ja] archiveFixtureEpisode は挿入したエピソードを、再公開が起点とする状態にする。エピソードの
// フィクスチャは一括作成が残す形と同じく公開状態で作られるため、逆方向のテストでは先に非公開に
// する。
func archiveFixtureEpisode(t *testing.T, db *sql.DB, episodeID model.EpisodeID) {
	t.Helper()

	if _, err := db.Exec(`UPDATE episodes SET unpublished_at = NOW() WHERE id = $1`, int64(episodeID)); err != nil {
		t.Fatalf("エピソードの非公開化に失敗: %v", err)
	}
}

// TestUnarchiveEpisodeUsecase_Execute_UnarchivesEpisodeAndAnime verifies re-publishing a mapped,
// archived episode clears episodes.unpublished_at (the state source of truth) and dual-writes
// the derived anime.status = published, and that a phase 2 sync right after reports Unchanged
// (the re-publish and the reconciliation derive the same status from unpublished_at, so the sync
// does not clobber the published anime back to archived).
//
// [Ja] TestUnarchiveEpisodeUsecase_Execute_UnarchivesEpisodeAndAnime は、マッピング済みで非公開
// のエピソードを再公開すると episodes.unpublished_at (状態の正本) がクリアされ、導出された
// anime.status = published が両書きされること、および直後のフェーズ 2 同期が Unchanged を報告
// することを検証する (再公開とリコンシリエーションが unpublished_at から同じ status を導出する
// ため、同期は公開済み anime を archived に戻さない)。
func TestUnarchiveEpisodeUsecase_Execute_UnarchivesEpisodeAndAnime(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUnarchiveEpisodeUsecase(db)

	workID, parentAnimeID := insertMappedCreateTargetWork(t, db)
	episodeID, episodeAnimeID := insertMappedUpdateTargetEpisode(t, db, workID, parentAnimeID, 100)
	archiveFixtureEpisode(t, db, episodeID)
	if _, err := db.Exec(`
		UPDATE animes
		SET
			status = 'archived',
			title = '編集前のタイトル',
			title_ro = 'Before',
			title_en = 'Before EN',
			archive_message = '非公開時のメッセージ'
		WHERE id = $1`,
		int64(episodeAnimeID),
	); err != nil {
		t.Fatalf("非公開済み anime の準備に失敗: %v", err)
	}

	output, err := uc.Execute(context.Background(), UnarchiveEpisodeInput{EpisodeID: episodeID, User: unsavedCreateActor()})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if output.EpisodeID != episodeID || output.WorkID != workID {
		t.Errorf("output = %+v, want {EpisodeID:%d WorkID:%d}", output, int64(episodeID), int64(workID))
	}

	if unpublishedAt := readArchivedEpisodeState(t, db, episodeID); unpublishedAt.Valid {
		t.Errorf("episodes.unpublished_at = %v, want NULL", unpublishedAt.Time)
	}

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), episodeAnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Status != model.AnimeStatusPublished {
		t.Errorf("anime.Status = %q, want %q", anime.Status, model.AnimeStatusPublished)
	}
	// The re-publish maps status alone, so anime-owned content stays byte-for-byte unchanged,
	// including the message the archive left behind.
	//
	// [Ja] 再公開が写像するのは status だけなので、非公開時のメッセージを含め anime 固有の内容は
	// そのまま保持される。
	if anime.ArchiveMessage.String != "非公開時のメッセージ" {
		t.Errorf("anime.ArchiveMessage = %q, want %q", anime.ArchiveMessage.String, "非公開時のメッセージ")
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

// TestUnarchiveEpisodeUsecase_Execute_SkipsAnimeForUnmappedEpisode covers an episode with no
// anime yet: only the episodes row is written, and the phase 2 sync creates the anime later with
// the status the re-published episode now derives.
//
// [Ja] TestUnarchiveEpisodeUsecase_Execute_SkipsAnimeForUnmappedEpisode は、まだ anime を持たない
// エピソードを検証する。書かれるのは episodes の行だけで、anime は後でフェーズ 2 の同期が、
// 再公開されたエピソードが導出する status で作成する。
func TestUnarchiveEpisodeUsecase_Execute_SkipsAnimeForUnmappedEpisode(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUnarchiveEpisodeUsecase(db)

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	episodeID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 100)
	archiveFixtureEpisode(t, db, episodeID)

	if _, err := uc.Execute(context.Background(), UnarchiveEpisodeInput{EpisodeID: episodeID, User: unsavedCreateActor()}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if unpublishedAt := readArchivedEpisodeState(t, db, episodeID); unpublishedAt.Valid {
		t.Errorf("episodes.unpublished_at = %v, want NULL", unpublishedAt.Time)
	}

	var animeID sql.NullInt64
	if err := db.QueryRow(`SELECT anime_id FROM episodes WHERE id = $1`, int64(episodeID)).Scan(&animeID); err != nil {
		t.Fatalf("episodes.anime_id の読み込みに失敗: %v", err)
	}
	if animeID.Valid {
		t.Errorf("episodes.anime_id = %d, want NULL のまま", animeID.Int64)
	}
}

// TestUnarchiveEpisodeUsecase_Execute_UnarchivesEpisodeWithUnmappedParent covers an episode that
// carries an anime while its parent work no longer does. Re-publishing maps the status alone,
// which needs no parent_anime_id, so the anime follows the episode instead of waiting for the
// parent to be mapped again.
//
// [Ja] TestUnarchiveEpisodeUsecase_Execute_UnarchivesEpisodeWithUnmappedParent は、エピソード
// 自身は anime を持つが親作品が持たなくなった場合を検証する。再公開が写像するのは status だけで
// parent_anime_id を必要としないため、anime は親が再びマッピングされるのを待たずエピソードに
// 追従する。
func TestUnarchiveEpisodeUsecase_Execute_UnarchivesEpisodeWithUnmappedParent(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUnarchiveEpisodeUsecase(db)

	workID, parentAnimeID := insertMappedCreateTargetWork(t, db)
	episodeID, episodeAnimeID := insertMappedUpdateTargetEpisode(t, db, workID, parentAnimeID, 100)
	archiveFixtureEpisode(t, db, episodeID)
	if _, err := db.Exec(`UPDATE animes SET status = 'archived' WHERE id = $1`, int64(episodeAnimeID)); err != nil {
		t.Fatalf("非公開済み anime の準備に失敗: %v", err)
	}
	if _, err := db.Exec(`UPDATE works SET anime_id = NULL WHERE id = $1`, int64(workID)); err != nil {
		t.Fatalf("親作品の写像の解除に失敗: %v", err)
	}

	if _, err := uc.Execute(context.Background(), UnarchiveEpisodeInput{EpisodeID: episodeID, User: unsavedCreateActor()}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), episodeAnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Status != model.AnimeStatusPublished {
		t.Errorf("anime.Status = %q, want %q", anime.Status, model.AnimeStatusPublished)
	}
}

// TestUnarchiveEpisodeUsecase_Execute_RequiresCommitter verifies authorization belongs to the
// write usecase as well as the HTTP boundary. Rejected callers cannot re-publish the episode;
// an editor can.
//
// [Ja] TestUnarchiveEpisodeUsecase_Execute_RequiresCommitter は認可が HTTP 境界だけでなく書き込み
// UseCase にも属することを検証する。拒否された呼び出し元はエピソードを再公開できず、編集者は
// 実行できる。
func TestUnarchiveEpisodeUsecase_Execute_RequiresCommitter(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUnarchiveEpisodeUsecase(db)
	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	episodeID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 100)
	archiveFixtureEpisode(t, db, episodeID)

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
			_, err := uc.Execute(context.Background(), UnarchiveEpisodeInput{
				EpisodeID: episodeID,
				User:      tt.user,
			})
			if tt.wantForbidden {
				appErr := model.AsAppError(err)
				if appErr == nil || appErr.Code != model.AppErrCodeForbidden {
					t.Fatalf("Execute() error = %v, want AppErrCodeForbidden", err)
				}
				if unpublishedAt := readArchivedEpisodeState(t, db, episodeID); !unpublishedAt.Valid {
					t.Error("拒否された送信が episodes.unpublished_at をクリアしました")
				}
				return
			}
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
		})
	}
}

// TestUnarchiveEpisodeUsecase_Execute_RollsBackWhenParentIsDeletedWhileEpisodeWriteWaits
// fixes the interleaving between the confirmation projection and the re-publish write. The
// re-publish waits on a locked episode after its pre-read, while the locking transaction deletes
// the parent work. The work guard must make the write report not found, and the usecase rollback
// must preserve the archived episode, counter, and anime status.
//
// [Ja] このテストは確認用の射影と再公開の書き込みの間の実行順を固定する。事前読み取り後の
// 再公開がロック済みの episode を待つ間に、ロック元のトランザクションが親作品を削除する。作品の
// ガードにより not found を返し、UseCase のロールバックにより非公開の episode、カウンター、anime
// の状態を保持しなければならない。
func TestUnarchiveEpisodeUsecase_Execute_RollsBackWhenParentIsDeletedWhileEpisodeWriteWaits(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUnarchiveEpisodeUsecase(db)
	workID, parentAnimeID := insertMappedCreateTargetWork(t, db)
	episodeID, episodeAnimeID := insertMappedUpdateTargetEpisode(t, db, workID, parentAnimeID, 100)
	archiveFixtureEpisode(t, db, episodeID)
	if _, err := db.Exec(
		`UPDATE animes SET status = $1 WHERE id = $2`,
		model.AnimeStatusArchived,
		int64(episodeAnimeID),
	); err != nil {
		t.Fatalf("非公開済み anime の準備に失敗: %v", err)
	}

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

	type unarchiveResult struct {
		output *UnarchiveEpisodeOutput
		err    error
	}
	resultCh := make(chan unarchiveResult, 1)
	go func() {
		output, err := uc.Execute(ctx, UnarchiveEpisodeInput{
			EpisodeID: episodeID,
			User:      unsavedCreateActor(),
		})
		resultCh <- unarchiveResult{output: output, err: err}
	}()

	// Observe the re-publish waiting on blockerTx before deleting the work. This proves the
	// pre-read completed and the guarded write started without relying on a sleep or test hook.
	//
	// [Ja] 作品を削除する前に、再公開が blockerTx を待っていることを観測する。sleep やテスト用
	// フックに依存せず、事前読み取りが完了してガード付きの書き込みが始まったことを証明する。
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

	var result unarchiveResult
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

	if unpublishedAt := readArchivedEpisodeState(t, db, episodeID); !unpublishedAt.Valid {
		t.Error("episodes.unpublished_at = NULL, want 非公開の時刻のまま")
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
	if anime.Status != model.AnimeStatusArchived {
		t.Errorf("anime.Status = %q, want %q", anime.Status, model.AnimeStatusArchived)
	}
}

// TestUnarchiveEpisodeUsecase_Execute_NotFound covers the submits the episode list cannot offer
// the re-publish action for: an episode that never existed, one already published, a deleted
// one, and one whose work was deleted. All four are reported as not found, which the handler
// turns into a 404.
//
// [Ja] TestUnarchiveEpisodeUsecase_Execute_NotFound は、エピソード一覧が再公開の操作を出せない
// 送信を検証する。存在しなかったエピソード、すでに公開中のエピソード、削除済みのエピソード、
// 作品が削除されたエピソードの 4 つ。いずれも not found として報告され、Handler はそれを 404 に
// 変換する。
func TestUnarchiveEpisodeUsecase_Execute_NotFound(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUnarchiveEpisodeUsecase(db)

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	publishedID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 100)
	deletedID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 200)
	archiveFixtureEpisode(t, db, deletedID)
	if _, err := db.Exec(`UPDATE episodes SET deleted_at = NOW() WHERE id = $1`, int64(deletedID)); err != nil {
		t.Fatalf("エピソードの削除に失敗: %v", err)
	}

	deletedWorkID := insertCreateTargetWork(t, db, sql.NullInt64{})
	orphanID := insertUpdateTargetEpisode(t, db, deletedWorkID, sql.NullInt64{}, 100)
	archiveFixtureEpisode(t, db, orphanID)
	if _, err := db.Exec(`UPDATE works SET deleted_at = NOW() WHERE id = $1`, int64(deletedWorkID)); err != nil {
		t.Fatalf("作品の削除に失敗: %v", err)
	}

	tests := map[string]model.EpisodeID{
		"存在しないエピソード":   model.EpisodeID(-1),
		"公開中のエピソード":    publishedID,
		"削除済みのエピソード":   deletedID,
		"削除済み作品のエピソード": orphanID,
	}
	for name, episodeID := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := uc.Execute(context.Background(), UnarchiveEpisodeInput{EpisodeID: episodeID, User: unsavedCreateActor()})
			appErr := model.AsAppError(err)
			if appErr == nil || appErr.Code != model.AppErrCodeResourceNotFound {
				t.Fatalf("Execute() error = %v, want AppErrCodeResourceNotFound", err)
			}
		})
	}
}
