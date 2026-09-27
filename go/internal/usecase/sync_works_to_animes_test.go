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

// syncWorkInputはworks -> animes同期に関係するworksカラムを保持する。
// ヘルパーが実際のworks行を挿入し、(自前でトランザクションを開く) 同期UseCaseが
// GetTestDB経由でその行を見られるようにする。
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

// defaultSyncWorkInputはNOT NULLカラムを満たした最小の公開TV作品を返す。
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
		t.Fatalf("worksの挿入に失敗: %v", err)
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

// reloadSyncWorkは同期ローダー経由でworkを読み直す。主に書き戻された
// anime_idを観測するため。
func reloadSyncWork(t *testing.T, db *sql.DB, workID model.WorkID) *model.Work {
	t.Helper()
	workRepo := repository.NewWorkRepository(query.New(db))
	works, err := workRepo.ListForAnimeSyncByIDs(context.Background(), []model.WorkID{workID})
	if err != nil {
		t.Fatalf("workの再取得に失敗: %v", err)
	}
	if len(works) != 1 {
		t.Fatalf("workの再取得件数 = %d、期待値 = 1", len(works))
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
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if result.Processed != 1 || result.Created != 1 || result.Updated != 0 || result.Unchanged != 0 {
		t.Fatalf("result = %+v、期待値 = {Processed:1 Created:1 Updated:0 Unchanged:0}", result)
	}

	work := reloadSyncWork(t, db, workID)
	if work.AnimeID == nil {
		t.Fatal("works.anime_id = nil、期待値 = 書き戻された値")
	}
	animeID := *work.AnimeID

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Title.String != "進撃の巨人" {
		t.Errorf("anime.Title = %q、期待値 = 進撃の巨人", anime.Title.String)
	}
	if anime.TitleRo.String != "Shingeki no Kyojin" {
		t.Errorf("anime.TitleRo = %q、期待値 = Shingeki no Kyojin", anime.TitleRo.String)
	}
	if anime.Synopsis.String != "あらすじ本文" {
		t.Errorf("anime.Synopsis = %q、期待値 = あらすじ本文", anime.Synopsis.String)
	}
	if anime.Media != model.AnimeMediaOVA {
		t.Errorf("anime.Media = %q、期待値 = ova", anime.Media)
	}
	if anime.Status != model.AnimeStatusPublished {
		t.Errorf("anime.Status = %q、期待値 = published", anime.Status)
	}
	// worksのNOT NULL DEFAULT '' カラムはanime上でNULLに写像される。
	if anime.TitleKana.Valid {
		t.Errorf("anime.TitleKana = %q、期待値 = NULL", anime.TitleKana.String)
	}
	if anime.TitleEn.Valid {
		t.Errorf("anime.TitleEn = %q、期待値 = NULL", anime.TitleEn.String)
	}
	// release_statusはworksに源泉がない。
	if anime.ReleaseStatus != "" {
		t.Errorf("anime.ReleaseStatus = %q、期待値 = 空 (NULL)", anime.ReleaseStatus)
	}

	classRepo := repository.NewAnimeClassificationRepository(query.New(db))
	classification, err := classRepo.GetByAnimeID(context.Background(), animeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID()のclassification = %v、エラー = %v", classification, err)
	}
	if classification.Kind != model.AnimeClassificationKindWork {
		t.Errorf("classification.Kind = %q、期待値 = work", classification.Kind)
	}
	if !classification.Standalone {
		t.Error("classification.Standalone = false、期待値 = true (no_episodes=true)")
	}
	if classification.EpisodeStartNumber.String != "2.5" {
		t.Errorf("classification.EpisodeStartNumber = %q、期待値 = 2.5", classification.EpisodeStartNumber.String)
	}
	if !classification.ExpectedEpisodesCount.Valid || classification.ExpectedEpisodesCount.Int32 != 12 {
		t.Errorf("classification.ExpectedEpisodesCount = %+v、期待値 = {12 true}", classification.ExpectedEpisodesCount)
	}
	// work分類ではepisode専用フィールドはNULLのまま。
	if classification.ParentAnimeID != nil {
		t.Errorf("classification.ParentAnimeID = %v、期待値 = nil", classification.ParentAnimeID)
	}
	if classification.Number.Valid || classification.NumberText.Valid || classification.SortNumber.Valid {
		t.Error("作品ではエピソード専用のフィールド (number / number_text / sort_number) がNULLであるべきだが、値が入っていた")
	}
	if classification.NumberFormatID != nil {
		t.Errorf("classification.NumberFormatID = %v、期待値 = nil", classification.NumberFormatID)
	}
}

func TestSyncWorksToAnimesUsecase_Execute_IsIdempotent(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncUsecase(db)

	workID := insertSyncWork(t, db, defaultSyncWorkInput())

	if _, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}}); err != nil {
		t.Fatalf("1回目のExecute()のエラー = %v", err)
	}

	// 同じworkに対する2回目の実行は差分なしを検出しなければならない
	// (media / status / NUMERICのラウンドトリップがチャーンを生まないことの検証)。
	result, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("2回目のExecute()のエラー = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.Updated != 0 || result.Unchanged != 1 {
		t.Fatalf("result = %+v、期待値 = {Processed:1 Created:0 Updated:0 Unchanged:1}", result)
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
		t.Fatalf("1回目のExecute()のエラー = %v", err)
	}

	if _, err := db.Exec(
		`UPDATE works SET title = $1, media = $2, no_episodes = $3 WHERE id = $4`,
		"新タイトル", workMediaMovie, true, int64(workID),
	); err != nil {
		t.Fatalf("worksの更新に失敗: %v", err)
	}

	result, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("2回目のExecute()のエラー = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.Updated != 1 || result.Unchanged != 0 {
		t.Fatalf("result = %+v、期待値 = {Processed:1 Created:0 Updated:1 Unchanged:0}", result)
	}

	work := reloadSyncWork(t, db, workID)
	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), *work.AnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Title.String != "新タイトル" {
		t.Errorf("anime.Title = %q、期待値 = 新タイトル", anime.Title.String)
	}
	if anime.Media != model.AnimeMediaMovie {
		t.Errorf("anime.Media = %q、期待値 = movie", anime.Media)
	}

	classRepo := repository.NewAnimeClassificationRepository(query.New(db))
	classification, err := classRepo.GetByAnimeID(context.Background(), *work.AnimeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID()のclassification = %v、エラー = %v", classification, err)
	}
	if !classification.Standalone {
		t.Error("no_episodesの変更後のclassification.Standalone = false、期待値 = true")
	}
}

func TestSyncWorksToAnimesUsecase_Execute_PreservesUnsourcedAnimeFields(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncUsecase(db)

	workID := insertSyncWork(t, db, defaultSyncWorkInput())
	if _, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}}); err != nil {
		t.Fatalf("1回目のExecute()のエラー = %v", err)
	}
	work := reloadSyncWork(t, db, workID)
	animeID := *work.AnimeID

	// works同期が源泉としないカラムに編集者の設定値があると仮定し、work側を
	// 変更して同期にUPDATEを発行させる。
	if _, err := db.Exec(
		`UPDATE animes SET release_status = $1, title_alter_ro = $2 WHERE id = $3`,
		string(model.ReleaseStatusReleased), "ローマ字別名", int64(animeID),
	); err != nil {
		t.Fatalf("animesの事前更新に失敗: %v", err)
	}
	if _, err := db.Exec(`UPDATE works SET title = $1 WHERE id = $2`, "改題", int64(workID)); err != nil {
		t.Fatalf("worksの更新に失敗: %v", err)
	}

	result, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("2回目のExecute()のエラー = %v", err)
	}
	if result.Updated != 1 {
		t.Fatalf("result.Updated = %d、期待値 = 1", result.Updated)
	}

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Title.String != "改題" {
		t.Errorf("anime.Title = %q、期待値 = 改題 (作品の変更が反映されること)", anime.Title.String)
	}
	if anime.ReleaseStatus != model.ReleaseStatusReleased {
		t.Errorf("anime.ReleaseStatus = %q、期待値 = released (保持されること)", anime.ReleaseStatus)
	}
	if anime.TitleAlterRo.String != "ローマ字別名" {
		t.Errorf("anime.TitleAlterRo = %q、期待値 = ローマ字別名 (保持されること)", anime.TitleAlterRo.String)
	}
}

func TestSyncWorksToAnimesUsecase_Execute_EmptyInput(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncUsecase(db)

	result, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: nil})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if result.Processed != 0 || result.Created != 0 || result.Updated != 0 || result.Unchanged != 0 {
		t.Fatalf("result = %+v、期待値 = すべて0", result)
	}
}

func TestSyncWorksToAnimesUsecase_Execute_RecreatesMissingClassification(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncUsecase(db)

	// no_episodes=trueにして、再作成される分類がwork由来の非デフォルトな
	// standalone値を持つようにし、アサーションを有意にする。
	in := defaultSyncWorkInput()
	in.noEpisodes = true
	workID := insertSyncWork(t, db, in)
	if _, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}}); err != nil {
		t.Fatalf("1回目のExecute()のエラー = %v", err)
	}
	work := reloadSyncWork(t, db, workID)
	animeID := *work.AnimeID

	// animeとworks.anime_idのマッピングは残したまま分類行だけを削除し、半端な
	// マッピング状態を再現する。次回の同期は分類を再作成して自己修復しなければならない
	// (既存animeに対するclassificationCreate経路で、更新として数えられる)。
	if _, err := db.Exec(`DELETE FROM anime_classifications WHERE anime_id = $1`, int64(animeID)); err != nil {
		t.Fatalf("anime_classificationsの削除に失敗: %v", err)
	}

	result, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("2回目のExecute()のエラー = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.Updated != 1 || result.Unchanged != 0 {
		t.Fatalf("result = %+v、期待値 = {Processed:1 Created:0 Updated:1 Unchanged:0}", result)
	}

	classRepo := repository.NewAnimeClassificationRepository(query.New(db))
	classification, err := classRepo.GetByAnimeID(context.Background(), animeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID()のclassification = %v、エラー = %v", classification, err)
	}
	if classification.Kind != model.AnimeClassificationKindWork {
		t.Errorf("classification.Kind = %q、期待値 = work", classification.Kind)
	}
	if !classification.Standalone {
		t.Error("classification.Standalone = false、期待値 = true (再作成時にno_episodes=trueが写ること)")
	}
}

func TestSyncWorksToAnimesUsecase_Execute_CreatesArchivedAnimeForUnpublishedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncUsecase(db)

	// works.unpublished_atが立っている (作品レベルで非公開) 場合は
	// anime.status = archivedに写像され、publishedにはならない。
	in := defaultSyncWorkInput()
	in.unpublishedAt = sql.NullTime{Time: time.Now(), Valid: true}
	workID := insertSyncWork(t, db, in)

	if _, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}}); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	work := reloadSyncWork(t, db, workID)
	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), *work.AnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Status != model.AnimeStatusArchived {
		t.Errorf("anime.Status = %q、期待値 = archived (works.unpublished_atが設定済みのため)", anime.Status)
	}
}

func TestSyncWorksToAnimesUsecase_Execute_CreatesDeletedAnimeForDeletedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncUsecase(db)

	// works.deleted_atが立っている (作品レベルでソフトデリート) 場合は
	// anime.status = deletedに写像される。
	in := defaultSyncWorkInput()
	in.deletedAt = sql.NullTime{Time: time.Now(), Valid: true}
	workID := insertSyncWork(t, db, in)

	if _, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}}); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	work := reloadSyncWork(t, db, workID)
	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), *work.AnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Status != model.AnimeStatusDeleted {
		t.Errorf("anime.Status = %q、期待値 = deleted (works.deleted_atが設定済みのため)", anime.Status)
	}
}

func TestSyncWorksToAnimesUsecase_Execute_ReconcilesWorkStateChangeToAnime(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncUsecase(db)

	workID := insertSyncWork(t, db, defaultSyncWorkInput())
	if _, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}}); err != nil {
		t.Fatalf("1回目のExecute()のエラー = %v", err)
	}
	work := reloadSyncWork(t, db, workID)
	animeID := *work.AnimeID
	animeRepo := repository.NewAnimeRepository(query.New(db))

	// works.unpublished_atを立てるとanime.status = archivedへリコンサイルされる
	// (状態導出が作品レベルの変更をanimesへ伝播する)。
	if _, err := db.Exec(`UPDATE works SET unpublished_at = NOW() WHERE id = $1`, int64(workID)); err != nil {
		t.Fatalf("works.unpublished_atの更新に失敗: %v", err)
	}
	result, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if result.Updated != 1 {
		t.Fatalf("result.Updated = %d、期待値 = 1 (unpublished_atが揃えられること)", result.Updated)
	}
	anime, err := animeRepo.GetByID(context.Background(), animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Status != model.AnimeStatusArchived {
		t.Errorf("anime.Status = %q、期待値 = archived", anime.Status)
	}

	// 続けてworks.deleted_atを立てるとanime.status = deletedへリコンサイルされる
	// (deleted_atがunpublished_atより優先される)。
	if _, err := db.Exec(`UPDATE works SET deleted_at = NOW() WHERE id = $1`, int64(workID)); err != nil {
		t.Fatalf("works.deleted_atの更新に失敗: %v", err)
	}
	result, err = uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if result.Updated != 1 {
		t.Fatalf("result.Updated = %d、期待値 = 1 (deleted_atが揃えられること)", result.Updated)
	}
	anime, err = animeRepo.GetByID(context.Background(), animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Status != model.AnimeStatusDeleted {
		t.Errorf("anime.Status = %q、期待値 = deleted", anime.Status)
	}
}

func TestSyncWorksToAnimesUsecase_Execute_DoesNotClobberAnimeArchivedState(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncUsecase(db)

	workID := insertSyncWork(t, db, defaultSyncWorkInput())
	if _, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}}); err != nil {
		t.Fatalf("1回目のExecute()のエラー = %v", err)
	}
	work := reloadSyncWork(t, db, workID)
	animeID := *work.AnimeID
	animeRepo := repository.NewAnimeRepository(query.New(db))

	// animes-first + works両書きによる非公開経路を再現する。animes.statusを
	// archivedに、works.unpublished_atを同時に立てる。これは導出の是正が守る不変条件で、
	// リコンシラーはarchivedをworks.unpublished_atから導出し、animes.status = archived
	// をpublishedに戻さず据え置く。
	if _, err := db.Exec(`UPDATE works SET unpublished_at = NOW() WHERE id = $1`, int64(workID)); err != nil {
		t.Fatalf("works.unpublished_atの更新に失敗: %v", err)
	}
	if _, err := db.Exec(`UPDATE animes SET status = 'archived' WHERE id = $1`, int64(animeID)); err != nil {
		t.Fatalf("animes.statusの更新に失敗: %v", err)
	}

	result, err := uc.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if result.Unchanged != 1 || result.Updated != 0 {
		t.Fatalf("result = %+v、期待値 = {Unchanged:1 Updated:0} (上書きされないこと)", result)
	}

	anime, err := animeRepo.GetByID(context.Background(), animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Status != model.AnimeStatusArchived {
		t.Errorf("anime.Status = %q、期待値 = archived (publishedへ上書きされないこと)", anime.Status)
	}
}

func TestPlanWorkAnimeSync_CreatesAnimeWhenMappedRowMissing(t *testing.T) {
	t.Parallel()

	// anime_idがロード済み集合に存在しないanimeを指すwork (宙ぶらりんの
	// マッピング)。works.anime_idの外部キーによりExecute経由ではこの状態に到達できない
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
		t.Fatalf("plan = {processed:%d creates:%d updates:%d unchanged:%d}、期待値 = {1 1 0 0}",
			plan.processed, len(plan.creates), len(plan.updates), plan.unchanged)
	}
	create := plan.creates[0]
	if create.workID != work.ID {
		t.Errorf("create.workID = %d、期待値 = %d", create.workID, work.ID)
	}
	if create.anime.Title.String != "宙ぶらりん作品" {
		t.Errorf("create.anime.Title = %q、期待値 = 宙ぶらりん作品", create.anime.Title.String)
	}
	if create.classification.Kind != model.AnimeClassificationKindWork {
		t.Errorf("create.classification.Kind = %q、期待値 = work", create.classification.Kind)
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
			t.Errorf("mediaToAnimeMedia(%d) = %q、期待値 = %q", tt.media, got, tt.want)
		}
	}
}

// TestAnimeStatusFromWorkStatusはworkの導出ライフサイクル状態をanimeのstatus
// enumに写像する純粋なenumアダプタを検証する。timestampsからstatusへの優先順位自体は
// model.Work.DerivedStatusが持ち、TestWork_DerivedStatusで担保する。
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
				t.Errorf("animeStatusFromWorkStatus(%q) = %q、期待値 = %q", tt.status, got, tt.want)
			}
		})
	}
}
