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

// syncWorkInput holds the works columns relevant to the works -> animes sync. The
// helper inserts a real works row so the sync usecase (which opens its own
// transaction) can see it via GetTestDB.
//
// [Ja] syncWorkInput は works -> animes 同期に関係する works カラムを保持する。
// ヘルパーが実際の works 行を挿入し、(自前でトランザクションを開く) 同期 UseCase が
// GetTestDB 経由でその行を見られるようにする。
type syncWorkInput struct {
	title                 string
	titleKana             string
	titleRo               string
	titleEn               string
	titleAlter            string
	titleAlterEn          string
	media                 int32
	synopsis              string
	synopsisEn            string
	synopsisSource        string
	synopsisSourceEn      string
	unpublishedAt         sql.NullTime
	deletedAt             sql.NullTime
	noEpisodes            bool
	manualEpisodesCount   sql.NullInt32
	startEpisodeRawNumber float64
	animeID               sql.NullInt64
}

// defaultSyncWorkInput returns a minimal published TV work with the NOT NULL
// columns satisfied.
//
// [Ja] defaultSyncWorkInput は NOT NULL カラムを満たした最小の公開 TV 作品を返す。
func defaultSyncWorkInput() syncWorkInput {
	return syncWorkInput{
		title:                 "テストアニメ",
		media:                 workMediaTV,
		startEpisodeRawNumber: 1.0,
	}
}

func insertSyncWork(t *testing.T, db *sql.DB, in syncWorkInput) model.WorkID {
	t.Helper()

	var id int64
	err := db.QueryRow(`
		INSERT INTO works (
			title, title_kana, title_ro, title_en, title_alter, title_alter_en,
			media, synopsis, synopsis_en, synopsis_source, synopsis_source_en,
			unpublished_at, deleted_at, no_episodes,
			manual_episodes_count, start_episode_raw_number, anime_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, $13, $14,
			$15, $16, $17, NOW(), NOW()
		) RETURNING id
	`,
		in.title, in.titleKana, in.titleRo, in.titleEn, in.titleAlter, in.titleAlterEn,
		in.media, in.synopsis, in.synopsisEn, in.synopsisSource, in.synopsisSourceEn,
		in.unpublishedAt, in.deletedAt, in.noEpisodes,
		in.manualEpisodesCount, in.startEpisodeRawNumber, in.animeID,
	).Scan(&id)
	if err != nil {
		t.Fatalf("works の挿入に失敗: %v", err)
	}
	return model.WorkID(id)
}

func newSyncUsecase(db *sql.DB) *SyncWorksToAnimesUsecase {
	queries := query.New(db)
	return NewSyncWorksToAnimesUsecase(
		db,
		repository.NewWorkRepository(queries),
		repository.NewAnimeRepository(queries),
		repository.NewAnimeClassificationRepository(queries),
	)
}

// reloadSyncWork re-reads a work through the sync loader, mainly to observe the
// written-back anime_id.
//
// [Ja] reloadSyncWork は同期ローダー経由で work を読み直す。主に書き戻された
// anime_id を観測するため。
func reloadSyncWork(t *testing.T, db *sql.DB, workID model.WorkID) *model.Work {
	t.Helper()
	workRepo := repository.NewWorkRepository(query.New(db))
	works, err := workRepo.ListForAnimeSyncByIDs(context.Background(), []model.WorkID{workID})
	if err != nil {
		t.Fatalf("work の再取得に失敗: %v", err)
	}
	if len(works) != 1 {
		t.Fatalf("work の再取得件数 = %d, want 1", len(works))
	}
	return works[0]
}

func TestSyncWorksToAnimesUsecase_Execute_CreatesAnimeForUnmappedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncUsecase(db)

	in := defaultSyncWorkInput()
	in.title = "進撃の巨人"
	in.titleRo = "Shingeki no Kyojin"
	in.synopsis = "あらすじ本文"
	in.media = workMediaOVA
	in.noEpisodes = true
	in.manualEpisodesCount = sql.NullInt32{Int32: 12, Valid: true}
	in.startEpisodeRawNumber = 2.5
	workID := insertSyncWork(t, db, in)

	result, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Processed != 1 || result.Created != 1 || result.Updated != 0 || result.Unchanged != 0 {
		t.Fatalf("result = %+v, want {Processed:1 Created:1 Updated:0 Unchanged:0}", result)
	}

	work := reloadSyncWork(t, db, workID)
	if work.AnimeID == nil {
		t.Fatal("works.anime_id should be written back, got nil")
	}
	animeID := *work.AnimeID

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Title.String != "進撃の巨人" {
		t.Errorf("anime.Title = %q, want 進撃の巨人", anime.Title.String)
	}
	if anime.TitleRo.String != "Shingeki no Kyojin" {
		t.Errorf("anime.TitleRo = %q, want Shingeki no Kyojin", anime.TitleRo.String)
	}
	if anime.Synopsis.String != "あらすじ本文" {
		t.Errorf("anime.Synopsis = %q, want あらすじ本文", anime.Synopsis.String)
	}
	if anime.Media != model.AnimeMediaOVA {
		t.Errorf("anime.Media = %q, want ova", anime.Media)
	}
	if anime.Status != model.AnimeStatusPublished {
		t.Errorf("anime.Status = %q, want published", anime.Status)
	}
	// works' NOT NULL DEFAULT '' columns map to NULL on the anime.
	//
	// [Ja] works の NOT NULL DEFAULT '' カラムは anime 上で NULL に写像される。
	if anime.TitleKana.Valid {
		t.Errorf("anime.TitleKana should be NULL, got %q", anime.TitleKana.String)
	}
	if anime.TitleEn.Valid {
		t.Errorf("anime.TitleEn should be NULL, got %q", anime.TitleEn.String)
	}
	// release_status has no source in works.
	//
	// [Ja] release_status は works に源泉がない。
	if anime.ReleaseStatus != "" {
		t.Errorf("anime.ReleaseStatus = %q, want empty (NULL)", anime.ReleaseStatus)
	}

	classRepo := repository.NewAnimeClassificationRepository(query.New(db))
	classification, err := classRepo.GetByAnimeID(context.Background(), animeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID() classification=%v err=%v", classification, err)
	}
	if classification.Kind != model.AnimeClassificationKindWork {
		t.Errorf("classification.Kind = %q, want work", classification.Kind)
	}
	if !classification.Standalone {
		t.Error("classification.Standalone = false, want true (no_episodes=true)")
	}
	if classification.EpisodeStartNumber.String != "2.5" {
		t.Errorf("classification.EpisodeStartNumber = %q, want 2.5", classification.EpisodeStartNumber.String)
	}
	if !classification.ExpectedEpisodesCount.Valid || classification.ExpectedEpisodesCount.Int32 != 12 {
		t.Errorf("classification.ExpectedEpisodesCount = %+v, want {12 true}", classification.ExpectedEpisodesCount)
	}
	// Episode-only fields stay NULL for a work classification.
	//
	// [Ja] work 分類では episode 専用フィールドは NULL のまま。
	if classification.ParentAnimeID != nil {
		t.Errorf("classification.ParentAnimeID = %v, want nil", classification.ParentAnimeID)
	}
	if classification.Number.Valid || classification.NumberText.Valid || classification.SortNumber.Valid {
		t.Error("episode-only fields (number / number_text / sort_number) should be NULL for a work")
	}
	if classification.NumberFormatID != nil {
		t.Errorf("classification.NumberFormatID = %v, want nil", classification.NumberFormatID)
	}
}

func TestSyncWorksToAnimesUsecase_Execute_IsIdempotent(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncUsecase(db)

	workID := insertSyncWork(t, db, defaultSyncWorkInput())

	if _, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}}); err != nil {
		t.Fatalf("first Execute() error = %v", err)
	}

	// A second run over the same work must detect no diff (validates that the
	// media / status / NUMERIC round-trips do not churn).
	//
	// [Ja] 同じ work に対する 2 回目の実行は差分なしを検出しなければならない
	// (media / status / NUMERIC のラウンドトリップがチャーンを生まないことの検証)。
	result, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("second Execute() error = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.Updated != 0 || result.Unchanged != 1 {
		t.Fatalf("result = %+v, want {Processed:1 Created:0 Updated:0 Unchanged:1}", result)
	}
}

func TestSyncWorksToAnimesUsecase_Execute_UpdatesChangedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncUsecase(db)

	in := defaultSyncWorkInput()
	in.title = "旧タイトル"
	in.media = workMediaTV
	workID := insertSyncWork(t, db, in)

	if _, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}}); err != nil {
		t.Fatalf("first Execute() error = %v", err)
	}

	if _, err := db.Exec(
		`UPDATE works SET title = $1, media = $2, no_episodes = $3 WHERE id = $4`,
		"新タイトル", workMediaMovie, true, int64(workID),
	); err != nil {
		t.Fatalf("works の更新に失敗: %v", err)
	}

	result, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("second Execute() error = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.Updated != 1 || result.Unchanged != 0 {
		t.Fatalf("result = %+v, want {Processed:1 Created:0 Updated:1 Unchanged:0}", result)
	}

	work := reloadSyncWork(t, db, workID)
	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), *work.AnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Title.String != "新タイトル" {
		t.Errorf("anime.Title = %q, want 新タイトル", anime.Title.String)
	}
	if anime.Media != model.AnimeMediaMovie {
		t.Errorf("anime.Media = %q, want movie", anime.Media)
	}

	classRepo := repository.NewAnimeClassificationRepository(query.New(db))
	classification, err := classRepo.GetByAnimeID(context.Background(), *work.AnimeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID() classification=%v err=%v", classification, err)
	}
	if !classification.Standalone {
		t.Error("classification.Standalone = false, want true after no_episodes change")
	}
}

func TestSyncWorksToAnimesUsecase_Execute_PreservesUnsourcedAnimeFields(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncUsecase(db)

	workID := insertSyncWork(t, db, defaultSyncWorkInput())
	if _, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}}); err != nil {
		t.Fatalf("first Execute() error = %v", err)
	}
	work := reloadSyncWork(t, db, workID)
	animeID := *work.AnimeID

	// Simulate editor-set values on columns the works sync does not source, then
	// force a work change so the sync issues an UPDATE.
	//
	// [Ja] works 同期が源泉としないカラムに編集者の設定値があると仮定し、work 側を
	// 変更して同期に UPDATE を発行させる。
	if _, err := db.Exec(
		`UPDATE animes SET release_status = $1, title_alter_ro = $2 WHERE id = $3`,
		string(model.ReleaseStatusReleased), "ローマ字別名", int64(animeID),
	); err != nil {
		t.Fatalf("animes の事前更新に失敗: %v", err)
	}
	if _, err := db.Exec(`UPDATE works SET title = $1 WHERE id = $2`, "改題", int64(workID)); err != nil {
		t.Fatalf("works の更新に失敗: %v", err)
	}

	result, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("second Execute() error = %v", err)
	}
	if result.Updated != 1 {
		t.Fatalf("result.Updated = %d, want 1", result.Updated)
	}

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Title.String != "改題" {
		t.Errorf("anime.Title = %q, want 改題 (work change applied)", anime.Title.String)
	}
	if anime.ReleaseStatus != model.ReleaseStatusReleased {
		t.Errorf("anime.ReleaseStatus = %q, want released (preserved)", anime.ReleaseStatus)
	}
	if anime.TitleAlterRo.String != "ローマ字別名" {
		t.Errorf("anime.TitleAlterRo = %q, want ローマ字別名 (preserved)", anime.TitleAlterRo.String)
	}
}

func TestSyncWorksToAnimesUsecase_Execute_EmptyInput(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncUsecase(db)

	result, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: nil})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Processed != 0 || result.Created != 0 || result.Updated != 0 || result.Unchanged != 0 {
		t.Fatalf("result = %+v, want all zero", result)
	}
}

func TestSyncWorksToAnimesUsecase_Execute_RecreatesMissingClassification(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncUsecase(db)

	// no_episodes=true so the re-created classification carries a non-default
	// standalone value derived from the work, making the assertion meaningful.
	//
	// [Ja] no_episodes=true にして、再作成される分類が work 由来の非デフォルトな
	// standalone 値を持つようにし、アサーションを有意にする。
	in := defaultSyncWorkInput()
	in.noEpisodes = true
	workID := insertSyncWork(t, db, in)
	if _, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}}); err != nil {
		t.Fatalf("first Execute() error = %v", err)
	}
	work := reloadSyncWork(t, db, workID)
	animeID := *work.AnimeID

	// Drop only the classification row, keeping the anime and the works.anime_id
	// mapping, to simulate a half-built mapping. The next sync must self-heal by
	// re-creating the classification (the classificationCreate path on an already
	// existing anime), which counts as an update.
	//
	// [Ja] anime と works.anime_id のマッピングは残したまま分類行だけを削除し、半端な
	// マッピング状態を再現する。次回の同期は分類を再作成して自己修復しなければならない
	// (既存 anime に対する classificationCreate 経路で、更新として数えられる)。
	if _, err := db.Exec(`DELETE FROM anime_classifications WHERE anime_id = $1`, int64(animeID)); err != nil {
		t.Fatalf("anime_classifications の削除に失敗: %v", err)
	}

	result, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("second Execute() error = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.Updated != 1 || result.Unchanged != 0 {
		t.Fatalf("result = %+v, want {Processed:1 Created:0 Updated:1 Unchanged:0}", result)
	}

	classRepo := repository.NewAnimeClassificationRepository(query.New(db))
	classification, err := classRepo.GetByAnimeID(context.Background(), animeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID() classification=%v err=%v", classification, err)
	}
	if classification.Kind != model.AnimeClassificationKindWork {
		t.Errorf("classification.Kind = %q, want work", classification.Kind)
	}
	if !classification.Standalone {
		t.Error("classification.Standalone = false, want true (no_episodes=true mapped on re-create)")
	}
}

func TestSyncWorksToAnimesUsecase_Execute_CreatesArchivedAnimeForUnpublishedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncUsecase(db)

	// works.unpublished_at set (archived at the work level) must map to
	// anime.status = archived, not published.
	//
	// [Ja] works.unpublished_at が立っている (作品レベルで非公開) 場合は
	// anime.status = archived に写像され、published にはならない。
	in := defaultSyncWorkInput()
	in.unpublishedAt = sql.NullTime{Time: time.Now(), Valid: true}
	workID := insertSyncWork(t, db, in)

	if _, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	work := reloadSyncWork(t, db, workID)
	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), *work.AnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Status != model.AnimeStatusArchived {
		t.Errorf("anime.Status = %q, want archived (works.unpublished_at set)", anime.Status)
	}
}

func TestSyncWorksToAnimesUsecase_Execute_CreatesDeletedAnimeForDeletedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncUsecase(db)

	// works.deleted_at set (soft-deleted at the work level) must map to
	// anime.status = deleted.
	//
	// [Ja] works.deleted_at が立っている (作品レベルでソフトデリート) 場合は
	// anime.status = deleted に写像される。
	in := defaultSyncWorkInput()
	in.deletedAt = sql.NullTime{Time: time.Now(), Valid: true}
	workID := insertSyncWork(t, db, in)

	if _, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	work := reloadSyncWork(t, db, workID)
	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), *work.AnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Status != model.AnimeStatusDeleted {
		t.Errorf("anime.Status = %q, want deleted (works.deleted_at set)", anime.Status)
	}
}

func TestSyncWorksToAnimesUsecase_Execute_ReconcilesWorkStateChangeToAnime(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncUsecase(db)

	workID := insertSyncWork(t, db, defaultSyncWorkInput())
	if _, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}}); err != nil {
		t.Fatalf("first Execute() error = %v", err)
	}
	work := reloadSyncWork(t, db, workID)
	animeID := *work.AnimeID
	animeRepo := repository.NewAnimeRepository(query.New(db))

	// Setting works.unpublished_at must drive a reconciliation to anime.status =
	// archived (the state derivation propagates work-level changes to animes).
	//
	// [Ja] works.unpublished_at を立てると anime.status = archived へリコンサイルされる
	// (状態導出が作品レベルの変更を animes へ伝播する)。
	if _, err := db.Exec(`UPDATE works SET unpublished_at = NOW() WHERE id = $1`, int64(workID)); err != nil {
		t.Fatalf("works.unpublished_at の更新に失敗: %v", err)
	}
	result, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Updated != 1 {
		t.Fatalf("result.Updated = %d, want 1 (unpublished_at reconciled)", result.Updated)
	}
	anime, err := animeRepo.GetByID(context.Background(), animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Status != model.AnimeStatusArchived {
		t.Errorf("anime.Status = %q, want archived", anime.Status)
	}

	// Then setting works.deleted_at must reconcile anime.status = deleted (deleted_at
	// wins over unpublished_at).
	//
	// [Ja] 続けて works.deleted_at を立てると anime.status = deleted へリコンサイルされる
	// (deleted_at が unpublished_at より優先される)。
	if _, err := db.Exec(`UPDATE works SET deleted_at = NOW() WHERE id = $1`, int64(workID)); err != nil {
		t.Fatalf("works.deleted_at の更新に失敗: %v", err)
	}
	result, err = uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Updated != 1 {
		t.Fatalf("result.Updated = %d, want 1 (deleted_at reconciled)", result.Updated)
	}
	anime, err = animeRepo.GetByID(context.Background(), animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Status != model.AnimeStatusDeleted {
		t.Errorf("anime.Status = %q, want deleted", anime.Status)
	}
}

func TestSyncWorksToAnimesUsecase_Execute_DoesNotClobberAnimeArchivedState(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncUsecase(db)

	workID := insertSyncWork(t, db, defaultSyncWorkInput())
	if _, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}}); err != nil {
		t.Fatalf("first Execute() error = %v", err)
	}
	work := reloadSyncWork(t, db, workID)
	animeID := *work.AnimeID
	animeRepo := repository.NewAnimeRepository(query.New(db))

	// Simulate the animes-first + works two-write archive path: animes.status is set
	// to archived and works.unpublished_at is set together. This is the invariant the
	// derivation fix protects: the reconciler must derive archived from
	// works.unpublished_at and therefore leave animes.status = archived untouched
	// instead of reverting it.
	//
	// [Ja] animes-first + works 両書きによる非公開経路を再現する。animes.status を
	// archived に、works.unpublished_at を同時に立てる。これは導出の是正が守る不変条件で、
	// リコンシラーは archived を works.unpublished_at から導出し、animes.status = archived
	// を published に戻さず据え置く。
	if _, err := db.Exec(`UPDATE works SET unpublished_at = NOW() WHERE id = $1`, int64(workID)); err != nil {
		t.Fatalf("works.unpublished_at の更新に失敗: %v", err)
	}
	if _, err := db.Exec(`UPDATE animes SET status = 'archived' WHERE id = $1`, int64(animeID)); err != nil {
		t.Fatalf("animes.status の更新に失敗: %v", err)
	}

	result, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Unchanged != 1 || result.Updated != 0 {
		t.Fatalf("result = %+v, want {Unchanged:1 Updated:0} (no clobber)", result)
	}

	anime, err := animeRepo.GetByID(context.Background(), animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Status != model.AnimeStatusArchived {
		t.Errorf("anime.Status = %q, want archived (must not be clobbered to published)", anime.Status)
	}
}

func TestPlanWorkAnimeSync_CreatesAnimeWhenMappedRowMissing(t *testing.T) {
	t.Parallel()

	// A work whose anime_id points at an anime absent from the loaded set (a dangling
	// mapping). The works.anime_id foreign key makes this state unreachable through
	// Execute, so the pure planner is exercised directly: it must fall back to creating
	// a fresh anime + classification rather than attempting an update.
	//
	// [Ja] anime_id がロード済み集合に存在しない anime を指す work (宙ぶらりんの
	// マッピング)。works.anime_id の外部キーにより Execute 経由ではこの状態に到達できない
	// ため、純粋なプランナーを直接呼ぶ。更新ではなく新規作成 (anime + 分類) に
	// フォールバックしなければならない。
	danglingAnimeID := model.AnimeID(1 << 40)
	work := &model.Work{
		ID:                    model.WorkID(1),
		Title:                 "宙ぶらりん作品",
		Media:                 workMediaTV,
		StartEpisodeRawNumber: 1,
		AnimeID:               &danglingAnimeID,
	}

	plan := planWorkAnimeSync(
		[]*model.Work{work},
		map[model.AnimeID]*model.Anime{},
		map[model.AnimeID]*model.AnimeClassification{},
	)

	if plan.processed != 1 || len(plan.creates) != 1 || len(plan.updates) != 0 || plan.unchanged != 0 {
		t.Fatalf("plan = {processed:%d creates:%d updates:%d unchanged:%d}, want {1 1 0 0}",
			plan.processed, len(plan.creates), len(plan.updates), plan.unchanged)
	}
	create := plan.creates[0]
	if create.workID != work.ID {
		t.Errorf("create.workID = %d, want %d", create.workID, work.ID)
	}
	if create.anime.Title.String != "宙ぶらりん作品" {
		t.Errorf("create.anime.Title = %q, want 宙ぶらりん作品", create.anime.Title.String)
	}
	if create.classification.Kind != model.AnimeClassificationKindWork {
		t.Errorf("create.classification.Kind = %q, want work", create.classification.Kind)
	}
}

func TestMediaToAnimeMedia(t *testing.T) {
	t.Parallel()

	tests := []struct {
		media int32
		want  model.AnimeMedia
	}{
		{workMediaOther, model.AnimeMediaOther},
		{workMediaTV, model.AnimeMediaTV},
		{workMediaOVA, model.AnimeMediaOVA},
		{workMediaMovie, model.AnimeMediaMovie},
		{workMediaONA, model.AnimeMediaONA},
		{99, model.AnimeMediaOther},
	}
	for _, tt := range tests {
		if got := mediaToAnimeMedia(tt.media); got != tt.want {
			t.Errorf("mediaToAnimeMedia(%d) = %q, want %q", tt.media, got, tt.want)
		}
	}
}

// TestAnimeStatusFromWorkStatus verifies the pure enum adapter that maps a work's
// derived lifecycle status onto the anime status enum. The timestamp-to-status priority
// itself is owned by model.Work.DerivedStatus and covered by TestWork_DerivedStatus.
//
// [Ja] TestAnimeStatusFromWorkStatus は work の導出ライフサイクル状態を anime の status
// enum に写像する純粋な enum アダプタを検証する。timestamps から status への優先順位自体は
// model.Work.DerivedStatus が持ち、TestWork_DerivedStatus で担保する。
func TestAnimeStatusFromWorkStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status model.WorkStatus
		want   model.AnimeStatus
	}{
		{model.WorkStatusPublished, model.AnimeStatusPublished},
		{model.WorkStatusArchived, model.AnimeStatusArchived},
		{model.WorkStatusDeleted, model.AnimeStatusDeleted},
	}
	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			if got := animeStatusFromWorkStatus(tt.status); got != tt.want {
				t.Errorf("animeStatusFromWorkStatus(%q) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}
