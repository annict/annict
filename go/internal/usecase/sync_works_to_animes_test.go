package usecase

import (
	"context"
	"database/sql"
	"testing"

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
	status                string
	archiveMessage        sql.NullString
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
		status:                "published",
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
			status, archive_message, no_episodes, manual_episodes_count,
			start_episode_raw_number, anime_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, $13, $14, $15,
			$16, $17, NOW(), NOW()
		) RETURNING id
	`,
		in.title, in.titleKana, in.titleRo, in.titleEn, in.titleAlter, in.titleAlterEn,
		in.media, in.synopsis, in.synopsisEn, in.synopsisSource, in.synopsisSourceEn,
		in.status, in.archiveMessage, in.noEpisodes, in.manualEpisodesCount,
		in.startEpisodeRawNumber, in.animeID,
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
		Status:                model.WorkStatusPublished,
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

func TestWorkStatusToAnimeStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status model.WorkStatus
		want   model.AnimeStatus
	}{
		{model.WorkStatusPublished, model.AnimeStatusPublished},
		{model.WorkStatusArchived, model.AnimeStatusArchived},
		{model.WorkStatusDeleted, model.AnimeStatusDeleted},
		{model.WorkStatus(""), model.AnimeStatusPublished},
	}
	for _, tt := range tests {
		if got := workStatusToAnimeStatus(tt.status); got != tt.want {
			t.Errorf("workStatusToAnimeStatus(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}
