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

// newDeleteEpisodeUsecase wires the delete usecase against the shared test DB, on the pool rather
// than on a test transaction for the reason newArchiveEpisodeUsecase states.
//
// [Ja] newDeleteEpisodeUsecase は共有テスト DB に対してエピソード削除 UseCase を組み立てる。
// newArchiveEpisodeUsecase が述べる理由により、テスト用トランザクションではなくプールに対して
// 組み立てる。
func newDeleteEpisodeUsecase(db *sql.DB) *DeleteEpisodeUsecase {
	queries := query.New(db)
	return NewDeleteEpisodeUsecase(
		db,
		repository.NewEpisodeRepository(queries),
		repository.NewAnimeRepository(queries),
	)
}

// unsavedDeleteActor is the administrator the delete tests authorize with. Deleting is admin-only
// while archiving is open to committers (ADR 0009), so these tests cannot reuse the editor
// unsavedCreateActor returns. The row is never persisted: the delete records no db_activity, so
// nothing references the user.
//
// [Ja] unsavedDeleteActor は削除テストが認可に使う管理者。非公開が committer に開かれているのに
// 対し削除は admin 専用のため (ADR 0009)、これらのテストは unsavedCreateActor が返す編集者を
// 使えない。行は永続化しない。削除は db_activity を作らず、ユーザーを参照するものが無いため。
func unsavedDeleteActor() *model.User {
	return &model.User{ID: 1, Role: model.RoleAdmin}
}

// readDeletedEpisodeState returns the state column the delete writes, so a test can tell a
// deleted episode from one the submit left alone.
//
// [Ja] readDeletedEpisodeState は削除が書く状態カラムを返す。削除されたエピソードと、送信が手を
// 触れなかったエピソードをテストが区別できるようにするため。
func readDeletedEpisodeState(t *testing.T, db *sql.DB, episodeID model.EpisodeID) sql.NullTime {
	t.Helper()

	var deletedAt sql.NullTime
	if err := db.QueryRow(`SELECT deleted_at FROM episodes WHERE id = $1`, int64(episodeID)).Scan(&deletedAt); err != nil {
		t.Fatalf("エピソードの状態の読み込みに失敗: %v", err)
	}

	return deletedAt
}

// TestDeleteEpisodeUsecase_Execute_DeletesEpisodeAndAnime verifies deleting a mapped, published
// episode sets episodes.deleted_at (the state source of truth) and dual-writes the derived
// anime.status = deleted, and that a phase 2 sync right after reports Unchanged (the delete and
// the reconciliation derive the same status from deleted_at, so the sync does not clobber the
// deleted anime back to published).
//
// [Ja] TestDeleteEpisodeUsecase_Execute_DeletesEpisodeAndAnime は、マッピング済みで公開中の
// エピソードを削除すると episodes.deleted_at (状態の正本) が立ち、導出された anime.status =
// deleted が両書きされること、および直後のフェーズ 2 同期が Unchanged を報告することを検証する
// (削除とリコンシリエーションが deleted_at から同じ status を導出するため、同期は削除済み anime
// を published に戻さない)。
func TestDeleteEpisodeUsecase_Execute_DeletesEpisodeAndAnime(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newDeleteEpisodeUsecase(db)

	workID, parentAnimeID := insertMappedCreateTargetWork(t, db)
	episodeID, episodeAnimeID := insertMappedUpdateTargetEpisode(t, db, workID, parentAnimeID, 100)
	if _, err := db.Exec(`
		UPDATE animes
		SET
			title = '編集前のタイトル',
			title_ro = 'Before',
			title_en = 'Before EN'
		WHERE id = $1`,
		int64(episodeAnimeID),
	); err != nil {
		t.Fatalf("anime の準備に失敗: %v", err)
	}

	output, err := uc.Execute(context.Background(), DeleteEpisodeInput{EpisodeID: episodeID, User: unsavedDeleteActor()})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if output.EpisodeID != episodeID || output.WorkID != workID {
		t.Errorf("output = %+v, want {EpisodeID:%d WorkID:%d}", output, int64(episodeID), int64(workID))
	}

	if deletedAt := readDeletedEpisodeState(t, db, episodeID); !deletedAt.Valid {
		t.Error("episodes.deleted_at = NULL, want 削除の時刻")
	}

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), episodeAnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Status != model.AnimeStatusDeleted {
		t.Errorf("anime.Status = %q, want %q", anime.Status, model.AnimeStatusDeleted)
	}
	// The delete maps status alone, so anime-owned content stays byte-for-byte unchanged. The
	// row survives the delete because animes has no physical delete (ADR 0004).
	//
	// [Ja] 削除が写像するのは status だけなので、anime 固有の内容はそのまま保持される。animes は
	// 物理削除を持たない (ADR 0004) ため、行は削除後も残る。
	if anime.Title.String != "編集前のタイトル" {
		t.Errorf("anime.Title = %q, want %q", anime.Title.String, "編集前のタイトル")
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

// TestDeleteEpisodeUsecase_Execute_DeletesArchivedEpisode covers the state the archive left an
// episode in. An archived episode is deletable, unlike the archive submit which refuses it, so
// the administrator does not have to re-publish a row before removing it.
//
// [Ja] TestDeleteEpisodeUsecase_Execute_DeletesArchivedEpisode は非公開がエピソードに残した状態
// を検証する。非公開の送信が拒否するのとは異なり、非公開のエピソードも削除できる。管理者が行を
// 消す前に再公開しなくて済むようにするため。
func TestDeleteEpisodeUsecase_Execute_DeletesArchivedEpisode(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newDeleteEpisodeUsecase(db)

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	episodeID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 100)
	archiveFixtureEpisode(t, db, episodeID)

	if _, err := uc.Execute(context.Background(), DeleteEpisodeInput{EpisodeID: episodeID, User: unsavedDeleteActor()}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if deletedAt := readDeletedEpisodeState(t, db, episodeID); !deletedAt.Valid {
		t.Error("episodes.deleted_at = NULL, want 削除の時刻")
	}
	// The archive timestamp is left where it is: deleted_at wins in DerivedStatus, and keeping
	// unpublished_at records that the row was already out of the counter when it was deleted.
	//
	// [Ja] 非公開の時刻はそのまま残す。DerivedStatus では deleted_at が優先され、unpublished_at
	// を残すことで、削除時点でその行が既にカウンターから外れていたことが記録される。
	if unpublishedAt := readArchivedEpisodeState(t, db, episodeID); !unpublishedAt.Valid {
		t.Error("episodes.unpublished_at = NULL, want 非公開の時刻のまま")
	}
}

// TestDeleteEpisodeUsecase_Execute_SkipsAnimeForUnmappedEpisode covers an episode with no anime
// yet: only the episodes row is written, and the phase 2 sync creates the anime later with the
// status the deleted episode now derives.
//
// [Ja] TestDeleteEpisodeUsecase_Execute_SkipsAnimeForUnmappedEpisode は、まだ anime を持たない
// エピソードを検証する。書かれるのは episodes の行だけで、anime は後でフェーズ 2 の同期が、
// 削除されたエピソードが導出する status で作成する。
func TestDeleteEpisodeUsecase_Execute_SkipsAnimeForUnmappedEpisode(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newDeleteEpisodeUsecase(db)

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	episodeID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 100)

	if _, err := uc.Execute(context.Background(), DeleteEpisodeInput{EpisodeID: episodeID, User: unsavedDeleteActor()}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if deletedAt := readDeletedEpisodeState(t, db, episodeID); !deletedAt.Valid {
		t.Error("episodes.deleted_at = NULL, want 削除の時刻")
	}

	var animeID sql.NullInt64
	if err := db.QueryRow(`SELECT anime_id FROM episodes WHERE id = $1`, int64(episodeID)).Scan(&animeID); err != nil {
		t.Fatalf("episodes.anime_id の読み込みに失敗: %v", err)
	}
	if animeID.Valid {
		t.Errorf("episodes.anime_id = %d, want NULL のまま", animeID.Int64)
	}
}

// TestDeleteEpisodeUsecase_Execute_RequiresAdmin verifies authorization belongs to the write
// usecase as well as the HTTP boundary, and that it is the admin rule rather than the committer
// one the archive endpoints apply: an editor who may archive an episode may not delete it
// (ADR 0009).
//
// [Ja] TestDeleteEpisodeUsecase_Execute_RequiresAdmin は認可が HTTP 境界だけでなく書き込み
// UseCase にも属すること、そしてそれが非公開エンドポイントの committer の規則ではなく admin の
// 規則であることを検証する。エピソードを非公開にできる編集者も、削除はできない (ADR 0009)。
func TestDeleteEpisodeUsecase_Execute_RequiresAdmin(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newDeleteEpisodeUsecase(db)
	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	episodeID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 100)

	tests := []struct {
		name          string
		user          *model.User
		wantForbidden bool
	}{
		{name: "未認証", user: nil, wantForbidden: true},
		{name: "一般ユーザー", user: &model.User{ID: 1, Role: model.RoleUser}, wantForbidden: true},
		{name: "編集者", user: &model.User{ID: 1, Role: model.RoleEditor}, wantForbidden: true},
		{name: "管理者", user: unsavedDeleteActor()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.Execute(context.Background(), DeleteEpisodeInput{
				EpisodeID: episodeID,
				User:      tt.user,
			})
			if tt.wantForbidden {
				appErr := model.AsAppError(err)
				if appErr == nil || appErr.Code != model.AppErrCodeForbidden {
					t.Fatalf("Execute() error = %v, want AppErrCodeForbidden", err)
				}
				if deletedAt := readDeletedEpisodeState(t, db, episodeID); deletedAt.Valid {
					t.Error("拒否された送信が episodes.deleted_at を立てました")
				}
				return
			}
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
		})
	}
}

// TestDeleteEpisodeUsecase_Execute_RollsBackWhenParentIsDeletedWhileEpisodeWriteWaits fixes the
// interleaving between the pre-transaction projection and the delete write. The delete waits on a
// locked episode after its pre-read, while the locking transaction deletes the parent work. The
// work guard must make the write report not found, and the usecase rollback must preserve the
// episode, counter, and anime status. A static fixture cannot fix this: a parent deleted before
// the statement starts is caught by the EXISTS guard on the episode update, while the guard that
// catches a concurrent deletion is the one on the works update, which READ COMMITTED re-checks.
//
// [Ja] このテストはトランザクション前の射影と削除の書き込みの間の実行順を固定する。事前読み取り
// 後の削除がロック済みの episode を待つ間に、ロック元のトランザクションが親作品を削除する。作品の
// ガードにより not found を返し、UseCase のロールバックにより episode、カウンター、anime の状態を
// 保持しなければならない。静的なフィクスチャではこれを固定できない。ステートメント開始前に削除
// された親は episode の更新側の EXISTS ガードが捉えるが、同時削除を捉えるのは READ COMMITTED が
// 再検査する works の更新側のガードであるため。
func TestDeleteEpisodeUsecase_Execute_RollsBackWhenParentIsDeletedWhileEpisodeWriteWaits(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newDeleteEpisodeUsecase(db)
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

	type deleteResult struct {
		output *DeleteEpisodeOutput
		err    error
	}
	resultCh := make(chan deleteResult, 1)
	go func() {
		output, err := uc.Execute(ctx, DeleteEpisodeInput{
			EpisodeID: episodeID,
			User:      unsavedDeleteActor(),
		})
		resultCh <- deleteResult{output: output, err: err}
	}()

	// Observe the delete waiting on blockerTx before deleting the work. This proves the pre-read
	// completed and the guarded write started without relying on a sleep or test hook.
	//
	// [Ja] 作品を削除する前に、削除が blockerTx を待っていることを観測する。sleep やテスト用
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

	var result deleteResult
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

	if deletedAt := readDeletedEpisodeState(t, db, episodeID); deletedAt.Valid {
		t.Error("episodes.deleted_at に値が入った, want NULL のまま")
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

// TestDeleteEpisodeUsecase_Execute_NotFound covers the submits the episode list cannot offer the
// delete action for: an episode that never existed, one already deleted, and one whose work was
// deleted. All three are reported as not found, which the handler turns into a 404.
//
// [Ja] TestDeleteEpisodeUsecase_Execute_NotFound は、エピソード一覧が削除の操作を出せない送信を
// 検証する。存在しなかったエピソード、すでに削除済みのエピソード、作品が削除されたエピソードの
// 3 つ。いずれも not found として報告され、Handler はそれを 404 に変換する。
func TestDeleteEpisodeUsecase_Execute_NotFound(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newDeleteEpisodeUsecase(db)

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	deletedID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 100)
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
		"削除済みのエピソード":   deletedID,
		"削除済み作品のエピソード": orphanID,
	}
	for name, episodeID := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := uc.Execute(context.Background(), DeleteEpisodeInput{EpisodeID: episodeID, User: unsavedDeleteActor()})
			appErr := model.AsAppError(err)
			if appErr == nil || appErr.Code != model.AppErrCodeResourceNotFound {
				t.Fatalf("Execute() error = %v, want AppErrCodeResourceNotFound", err)
			}
		})
	}
}
