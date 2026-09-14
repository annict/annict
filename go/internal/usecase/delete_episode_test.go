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

// newDeleteEpisodeUsecaseは共有テストDBに対してエピソード削除UseCaseを組み立てる。
// newArchiveEpisodeUsecaseが述べる理由により、テスト用トランザクションではなくプールに対して
// 組み立てる。
func newDeleteEpisodeUsecase(db *sql.DB) *DeleteEpisodeUsecase {
	queries := query.New(db)
	return NewDeleteEpisodeUsecase(
		db,
		repository.NewEpisodeRepository(queries),
		repository.NewAnimeRepository(queries),
	)
}

// unsavedDeleteActorは削除テストが認可に使う管理者。非公開がcommitterに開かれているのに
// 対し削除はadmin専用のため (ADR 0009)、これらのテストはunsavedCreateActorが返す編集者を
// 使えない。行は永続化しない。削除はdb_activityを作らず、ユーザーを参照するものが無いため。
func unsavedDeleteActor() *model.User {
	return &model.User{ID: 1, Role: model.RoleAdmin}
}

// readDeletedEpisodeStateは削除が書く状態カラムを返す。削除されたエピソードと、送信が手を
// 触れなかったエピソードをテストが区別できるようにするため。
func readDeletedEpisodeState(t *testing.T, db *sql.DB, episodeID model.EpisodeID) sql.NullTime {
	t.Helper()

	var deletedAt sql.NullTime
	if err := db.QueryRow(`SELECT deleted_at FROM episodes WHERE id = $1`, int64(episodeID)).Scan(&deletedAt); err != nil {
		t.Fatalf("エピソードの状態の読み込みに失敗: %v", err)
	}

	return deletedAt
}

// TestDeleteEpisodeUsecase_Execute_DeletesEpisodeAndAnimeは、マッピング済みで公開中の
// エピソードを削除するとepisodes.deleted_at (状態の正本) が立ち、導出されたanime.status =
// deletedが両書きされること、および直後のフェーズ2同期がUnchangedを報告することを検証する
// (削除とリコンシリエーションがdeleted_atから同じstatusを導出するため、同期は削除済みanime
// をpublishedに戻さない)。
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
		t.Fatalf("animeの準備に失敗: %v", err)
	}

	output, err := uc.Execute(context.Background(), DeleteEpisodeInput{EpisodeID: episodeID, User: unsavedDeleteActor()})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if output.EpisodeID != episodeID || output.WorkID != workID {
		t.Errorf("output = %+v、期待値 = {EpisodeID:%d WorkID:%d}", output, int64(episodeID), int64(workID))
	}

	if deletedAt := readDeletedEpisodeState(t, db, episodeID); !deletedAt.Valid {
		t.Error("episodes.deleted_at = NULL、期待値 = 削除の時刻")
	}

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), episodeAnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Status != model.AnimeStatusDeleted {
		t.Errorf("anime.Status = %q、期待値 = %q", anime.Status, model.AnimeStatusDeleted)
	}
	// 削除が写像するのはstatusだけなので、anime固有の内容はそのまま保持される。animesは
	// 物理削除を持たない (ADR 0004) ため、行は削除後も残る。
	if anime.Title.String != "編集前のタイトル" {
		t.Errorf("anime.Title = %q、期待値 = %q", anime.Title.String, "編集前のタイトル")
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

// TestDeleteEpisodeUsecase_Execute_DeletesArchivedEpisodeは非公開がエピソードに残した状態
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
		t.Fatalf("Execute()のエラー = %v", err)
	}

	if deletedAt := readDeletedEpisodeState(t, db, episodeID); !deletedAt.Valid {
		t.Error("episodes.deleted_at = NULL、期待値 = 削除の時刻")
	}
	// 非公開の時刻はそのまま残す。DerivedStatusではdeleted_atが優先され、unpublished_at
	// を残すことで、削除時点でその行が既にカウンターから外れていたことが記録される。
	if unpublishedAt := readArchivedEpisodeState(t, db, episodeID); !unpublishedAt.Valid {
		t.Error("episodes.unpublished_at = NULL、期待値 = 非公開の時刻のまま")
	}
}

// TestDeleteEpisodeUsecase_Execute_SkipsAnimeForUnmappedEpisodeは、まだanimeを持たない
// エピソードを検証する。書かれるのはepisodesの行だけで、animeは後でフェーズ2の同期が、
// 削除されたエピソードが導出するstatusで作成する。
func TestDeleteEpisodeUsecase_Execute_SkipsAnimeForUnmappedEpisode(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newDeleteEpisodeUsecase(db)

	workID := insertCreateTargetWork(t, db, sql.NullInt64{})
	episodeID := insertUpdateTargetEpisode(t, db, workID, sql.NullInt64{}, 100)

	if _, err := uc.Execute(context.Background(), DeleteEpisodeInput{EpisodeID: episodeID, User: unsavedDeleteActor()}); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	if deletedAt := readDeletedEpisodeState(t, db, episodeID); !deletedAt.Valid {
		t.Error("episodes.deleted_at = NULL、期待値 = 削除の時刻")
	}

	var animeID sql.NullInt64
	if err := db.QueryRow(`SELECT anime_id FROM episodes WHERE id = $1`, int64(episodeID)).Scan(&animeID); err != nil {
		t.Fatalf("episodes.anime_idの読み込みに失敗: %v", err)
	}
	if animeID.Valid {
		t.Errorf("episodes.anime_id = %d、期待値 = NULLのまま", animeID.Int64)
	}
}

// TestDeleteEpisodeUsecase_Execute_RequiresAdminは認可がHTTP境界だけでなく書き込み
// UseCaseにも属すること、そしてそれが非公開エンドポイントのcommitterの規則ではなくadminの
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
					t.Fatalf("Execute()のエラー = %v、期待値 = AppErrCodeForbidden", err)
				}
				if deletedAt := readDeletedEpisodeState(t, db, episodeID); deletedAt.Valid {
					t.Error("拒否された送信がepisodes.deleted_atを立てました")
				}
				return
			}
			if err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}
		})
	}
}

// このテストはトランザクション前の射影と削除の書き込みの間の実行順を固定する。事前読み取り
// 後の削除がロック済みのepisodeを待つ間に、ロック元のトランザクションが親作品を削除する。作品の
// ガードによりnot foundを返し、UseCaseのロールバックによりepisode、カウンター、animeの状態を
// 保持しなければならない。静的なフィクスチャではこれを固定できない。ステートメント開始前に削除
// された親はepisodeの更新側のEXISTSガードが捉えるが、同時削除を捉えるのはREAD COMMITTEDが
// 再検査するworksの更新側のガードであるため。
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

	// 作品を削除する前に、削除がblockerTxを待っていることを観測する。sleepやテスト用
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
		t.Fatalf("ブロッカー用トランザクションのCommit()のエラー = %v", err)
	}

	var result deleteResult
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

	if deletedAt := readDeletedEpisodeState(t, db, episodeID); deletedAt.Valid {
		t.Error("episodes.deleted_atに値が入った、期待値 = NULLのまま")
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

// TestDeleteEpisodeUsecase_Execute_NotFoundは、エピソード一覧が削除の操作を出せない送信を
// 検証する。存在しなかったエピソード、すでに削除済みのエピソード、作品が削除されたエピソードの
// 3つ。いずれもnot foundとして報告され、Handlerはそれを404に変換する。
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
				t.Fatalf("Execute()のエラー = %v、期待値 = AppErrCodeResourceNotFound", err)
			}
		})
	}
}
