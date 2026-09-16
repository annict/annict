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

// newArchiveEpisodeUsecaseは共有テストDBに対してエピソード非公開UseCaseを組み立てる。
// 作成 / 更新UseCaseと同じく内部で自前のトランザクションを開くため、テストはSetupTxではなく
// GetTestDBを使い、コミット済みの行がUseCaseの内側トランザクションと後続の同期不変条件
// チェックから見えるようにする。
func newArchiveEpisodeUsecase(db *sql.DB) *ArchiveEpisodeUsecase {
	queries := query.New(db)
	return NewArchiveEpisodeUsecase(
		db,
		repository.NewEpisodeRepository(queries),
		repository.NewAnimeRepository(queries),
	)
}

// readArchivedEpisodeStateは非公開と再公開が書く状態カラムを返す。どちらかが変更した
// エピソードと、送信が手を触れなかったエピソードをテストが区別できるようにするため。
func readArchivedEpisodeState(t *testing.T, db *sql.DB, episodeID model.EpisodeID) sql.NullTime {
	t.Helper()

	var unpublishedAt sql.NullTime
	if err := db.QueryRow(`SELECT unpublished_at FROM episodes WHERE id = $1`, int64(episodeID)).Scan(&unpublishedAt); err != nil {
		t.Fatalf("エピソードの状態の読み込みに失敗: %v", err)
	}

	return unpublishedAt
}

// awaitBlockedByBackendは、blockerPIDのバックエンドが保持するロックを待っている
// バックエンドが現れるまで待機する。検証対象の書き込みが既に開始して待機していることを確かめて
// から、インターリーブさせたい変更をコミットできるようにするため。pg_blocking_pidsにより、
// UseCase内部のtransactionへ触れたりsleepに依存したりせず実行順を固定する。resultChは
// 早期失敗のためだけに監視する。待機に入る前に終わったUseCaseはテストが固定したいインター
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
			t.Fatalf("ブロッカーのコミット前に返ったExecute()の結果 = %+v、期待値 = 行ロックを待つこと", result)
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
			t.Fatal("Execute()がepisodeの行ロックを待たなかった")
		}
	}
}

// TestArchiveEpisodeUsecase_Execute_ArchivesEpisodeAndAnimeは、マッピング済みで公開中の
// エピソードを非公開にするとepisodes.unpublished_at (状態の正本) が立ち、導出された
// anime.status = archivedが両書きされること、および直後のフェーズ2同期がUnchangedを報告する
// ことを検証する (非公開とリコンシリエーションがunpublished_atから同じstatusを導出するため、
// 同期はアーカイブ済みanimeをpublishedに戻さない)。
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
		t.Fatalf("同期済みanimeの準備に失敗: %v", err)
	}

	output, err := uc.Execute(context.Background(), ArchiveEpisodeInput{EpisodeID: episodeID, User: unsavedCreateActor()})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if output.EpisodeID != episodeID || output.WorkID != workID {
		t.Errorf("output = %+v、期待値 = {EpisodeID:%d WorkID:%d}", output, int64(episodeID), int64(workID))
	}

	if unpublishedAt := readArchivedEpisodeState(t, db, episodeID); !unpublishedAt.Valid {
		t.Error("episodes.unpublished_at = NULL、期待値 = 非公開の時刻")
	}

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), episodeAnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Status != model.AnimeStatusArchived {
		t.Errorf("anime.Status = %q、期待値 = %q", anime.Status, model.AnimeStatusArchived)
	}
	// 非公開が写像するのはstatusだけなので、anime固有の内容はそのまま保持される。
	if anime.ArchiveMessage.String != "非公開前のメッセージ" {
		t.Errorf("anime.ArchiveMessage = %q、期待値 = %q", anime.ArchiveMessage.String, "非公開前のメッセージ")
	}

	syncUC := newSyncEpisodesUsecase(db)
	result, err := syncUC.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
	if err != nil {
		t.Fatalf("同期のExecute()のエラー = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.Updated != 0 || result.Unchanged != 1 {
		t.Fatalf("同期の結果 = %+v、期待値 = {Processed:1 Created:0 Updated:0 Unchanged:1}", result)
	}
}

// TestArchiveEpisodeUsecase_Execute_SkipsAnimeForUnmappedEpisodeは、まだanimeを持たない
// エピソードを検証する。書かれるのはepisodesの行だけで、animeは後でフェーズ2の同期が、
// 非公開になったエピソードが導出するstatusで作成する。
func TestArchiveEpisodeUsecase_Execute_SkipsAnimeForUnmappedEpisode(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newArchiveEpisodeUsecase(db)

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	episodeID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 100)

	if _, err := uc.Execute(context.Background(), ArchiveEpisodeInput{EpisodeID: episodeID, User: unsavedCreateActor()}); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	if unpublishedAt := readArchivedEpisodeState(t, db, episodeID); !unpublishedAt.Valid {
		t.Error("episodes.unpublished_at = NULL、期待値 = 非公開の時刻")
	}

	var animeID sql.NullInt64
	if err := db.QueryRow(`SELECT anime_id FROM episodes WHERE id = $1`, int64(episodeID)).Scan(&animeID); err != nil {
		t.Fatalf("episodes.anime_idの読み込みに失敗: %v", err)
	}
	if animeID.Valid {
		t.Errorf("episodes.anime_id = %d、期待値 = NULLのまま", animeID.Int64)
	}
}

// TestArchiveEpisodeUsecase_Execute_ArchivesEpisodeWithUnmappedParentは、エピソード自身は
// animeを持つが親作品が持たなくなった場合を検証する。非公開が写像するのはstatusだけで
// parent_anime_idを必要としないため、animeは親が再びマッピングされるのを待たずエピソードに
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
		t.Fatalf("Execute()のエラー = %v", err)
	}

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), episodeAnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Status != model.AnimeStatusArchived {
		t.Errorf("anime.Status = %q、期待値 = %q", anime.Status, model.AnimeStatusArchived)
	}
}

// TestArchiveEpisodeUsecase_Execute_RequiresCommitterは認可がHTTP境界だけでなく書き込み
// UseCaseにも属することを検証する。拒否された呼び出し元はエピソードを非公開にできず、編集者は
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
					t.Fatalf("Execute()のエラー = %v、期待値 = AppErrCodeForbidden", err)
				}
				if unpublishedAt := readArchivedEpisodeState(t, db, episodeID); unpublishedAt.Valid {
					t.Error("拒否された送信がepisodes.unpublished_atを更新しました")
				}
				return
			}
			if err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}
		})
	}
}

// このテストは確認用の射影と非公開の書き込みの間の実行順を固定する。非公開がロック済みの
// episodeを待つ間に
// 別トランザクションがanimeの写像を変え、その窓でanime固有の編集もコミットする。解放後の
// status書き込みは更新したepisodeが返す写像へ追従し、競合した内容を保持しなければならない。
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
		t.Fatalf("ブロッカー用トランザクションのBeginTx()のエラー = %v", err)
	}
	defer func() { _ = blockerTx.Rollback() }()

	var blockerPID int
	if err := blockerTx.QueryRowContext(ctx, `SELECT pg_backend_pid()`).Scan(&blockerPID); err != nil {
		t.Fatalf("ブロッカーのbackend PIDの取得に失敗: %v", err)
	}
	if _, err := blockerTx.ExecContext(ctx, `SELECT id FROM episodes WHERE id = $1 FOR UPDATE`, int64(episodeID)); err != nil {
		t.Fatalf("episodeのロック取得に失敗: %v", err)
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

	// 写像を変更する前に、非公開がblockerTxを待っていることを観測する。
	awaitBlockedByBackend(t, ctx, db, blockerPID, resultCh)

	if _, err := blockerTx.ExecContext(
		ctx,
		`UPDATE episodes SET anime_id = $1 WHERE id = $2`,
		int64(currentAnimeID),
		int64(episodeID),
	); err != nil {
		t.Fatalf("episodeのanime写像変更に失敗: %v", err)
	}
	if _, err := db.ExecContext(
		ctx,
		`UPDATE animes SET archive_message = '同時編集後のメッセージ' WHERE id = $1`,
		int64(currentAnimeID),
	); err != nil {
		t.Fatalf("anime固有属性の同時編集に失敗: %v", err)
	}
	if err := blockerTx.Commit(); err != nil {
		t.Fatalf("ブロッカー用トランザクションのCommit()のエラー = %v", err)
	}

	var result archiveResult
	select {
	case result = <-resultCh:
	case <-ctx.Done():
		t.Fatalf("ブロッカーのコミット後にExecute()を待つcontextのエラー = %v、期待値 = nil (Execute()が完了すること)", ctx.Err())
	}
	if result.err != nil {
		t.Fatalf("Execute()のエラー = %v", result.err)
	}
	if result.output == nil || result.output.EpisodeID != episodeID || result.output.WorkID != workID {
		t.Fatalf("Execute()のoutput = %+v、期待値 = {EpisodeID:%d WorkID:%d}", result.output, int64(episodeID), int64(workID))
	}

	animeRepo := repository.NewAnimeRepository(query.New(db))
	formerAnime, err := animeRepo.GetByID(ctx, formerAnimeID)
	if err != nil || formerAnime == nil {
		t.Fatalf("以前のanimeのGetByID() anime=%v err=%v", formerAnime, err)
	}
	if formerAnime.Status != model.AnimeStatusPublished {
		t.Errorf("以前のanime.Status = %q、期待値 = %q", formerAnime.Status, model.AnimeStatusPublished)
	}

	currentAnime, err := animeRepo.GetByID(ctx, currentAnimeID)
	if err != nil || currentAnime == nil {
		t.Fatalf("現在のanimeのGetByID() anime=%v err=%v", currentAnime, err)
	}
	if currentAnime.Status != model.AnimeStatusArchived {
		t.Errorf("現在のanime.Status = %q、期待値 = %q", currentAnime.Status, model.AnimeStatusArchived)
	}
	if currentAnime.ArchiveMessage.String != "同時編集後のメッセージ" {
		t.Errorf(
			"現在のanime.ArchiveMessage = %q、期待値 = %q",
			currentAnime.ArchiveMessage.String,
			"同時編集後のメッセージ",
		)
	}
}

// このテストは確認用の射影と非公開の書き込みの間の実行順を固定する。事前読み取り後の非公開
// がロック済みのepisodeを待つ間に、ロック元のトランザクションが親作品を削除する。作品のガード
// によりnot foundを返し、UseCaseのロールバックにより公開中のepisode、カウンター、animeの
// 状態を保持しなければならない。同名の再公開テストと対になる。ガードは2つのステートメントが
// それぞれ持つため、各方向が自分のSQLの退行を検出するインターリーブを必要とする。
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
		t.Fatalf("ブロッカー用トランザクションのBeginTx()のエラー = %v", err)
	}
	defer func() { _ = blockerTx.Rollback() }()

	var blockerPID int
	if err := blockerTx.QueryRowContext(ctx, `SELECT pg_backend_pid()`).Scan(&blockerPID); err != nil {
		t.Fatalf("ブロッカーのbackend PIDの取得に失敗: %v", err)
	}
	if _, err := blockerTx.ExecContext(
		ctx,
		`SELECT id FROM episodes WHERE id = $1 FOR UPDATE`,
		int64(episodeID),
	); err != nil {
		t.Fatalf("episodeのロック取得に失敗: %v", err)
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

	// 作品を削除する前に、非公開がblockerTxを待っていることを観測する。事前読み取りが完了
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
		t.Fatalf("ブロッカー用トランザクションのCommit()のエラー = %v", err)
	}

	var result archiveResult
	select {
	case result = <-resultCh:
	case <-ctx.Done():
		t.Fatalf("ブロッカーのコミット後にExecute()を待つcontextのエラー = %v、期待値 = nil (Execute()が完了すること)", ctx.Err())
	}
	if result.output != nil {
		t.Fatalf("Execute()のoutput = %+v、期待値 = nil", result.output)
	}
	appErr := model.AsAppError(result.err)
	if appErr == nil || appErr.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("Execute()のエラー = %v、期待値 = AppErrCodeResourceNotFound", result.err)
	}

	if unpublishedAt := readArchivedEpisodeState(t, db, episodeID); unpublishedAt.Valid {
		t.Errorf("episodes.unpublished_at = %v、期待値 = NULLのまま", unpublishedAt.Time)
	}
	var episodesCountAfter int32
	if err := db.QueryRow(
		`SELECT episodes_count FROM works WHERE id = $1`,
		int64(workID),
	).Scan(&episodesCountAfter); err != nil {
		t.Fatalf("作品のカウンターの読み込みに失敗: %v", err)
	}
	if episodesCountAfter != episodesCountBefore {
		t.Errorf("works.episodes_count = %d、期待値 = %d", episodesCountAfter, episodesCountBefore)
	}

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(ctx, episodeAnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Status != model.AnimeStatusPublished {
		t.Errorf("anime.Status = %q、期待値 = %q", anime.Status, model.AnimeStatusPublished)
	}
}

// TestArchiveEpisodeUsecase_Execute_NotFoundは、確認ページを出せない送信を検証する。存在
// しなかったエピソード、すでに非公開のエピソード、削除済みのエピソード、作品が削除された
// エピソードの4つ。いずれもnot foundとして報告され、Handlerはそれをページ自身が返すのと同じ
// 404に変換する。
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
				t.Fatalf("Execute()のエラー = %v、期待値 = AppErrCodeResourceNotFound", err)
			}
		})
	}
}
