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

// syncEpisodeInputはepisodes -> animes同期に関係するepisodesカラムを保持する。
// ヘルパーが実際のepisodes行を挿入し、(自前でトランザクションを開く) 同期UseCaseが
// GetTestDB経由でその行を見られるようにする。
type syncEpisodeInput struct {
	workID        model.WorkID
	title         sql.NullString
	titleRo       string
	titleEn       string
	number        sql.NullString
	sortNumber    int32
	rawNumber     sql.NullFloat64
	unpublishedAt sql.NullTime
	deletedAt     sql.NullTime
	animeID       sql.NullInt64
}

// defaultSyncEpisodeInputは指定の親作品配下に、数値の話数 (raw_number=1,
// number="1") を持つ最小の公開エピソードを返す。archived / deletedの導出を検証する
// テストは、返った値にunpublished_at / deleted_atを立てて使う。
func defaultSyncEpisodeInput(workID model.WorkID) syncEpisodeInput {
	return syncEpisodeInput{
		workID:     workID,
		number:     sql.NullString{String: "1", Valid: true},
		sortNumber: 1,
		rawNumber:  sql.NullFloat64{Float64: 1, Valid: true},
	}
}

func insertSyncEpisode(t *testing.T, db *sql.DB, in syncEpisodeInput) model.EpisodeID {
	t.Helper()

	var id int64
	err := db.QueryRow(`
		INSERT INTO episodes (
			work_id, title, title_ro, title_en, number, sort_number,
			raw_number, unpublished_at, deleted_at,
			anime_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9,
			$10, NOW(), NOW()
		) RETURNING id
	`,
		int64(in.workID), in.title, in.titleRo, in.titleEn, in.number, in.sortNumber,
		in.rawNumber, in.unpublishedAt, in.deletedAt,
		in.animeID,
	).Scan(&id)
	if err != nil {
		t.Fatalf("episodesの挿入に失敗: %v", err)
	}
	return model.EpisodeID(id)
}

// insertBareAnimeは素のanime行を挿入しIDを返す。同期済みの親作品のanimeに
// 見立てて使う。
func insertBareAnime(t *testing.T, db *sql.DB) model.AnimeID {
	t.Helper()
	var id int64
	if err := db.QueryRow(`INSERT INTO animes (status) VALUES ('published') RETURNING id`).Scan(&id); err != nil {
		t.Fatalf("animesの挿入に失敗: %v", err)
	}
	return model.AnimeID(id)
}

// insertSyncedParentWorkは新規animeにマッピング済みのworkを挿入し、その配下の
// episodesがNOT NULLのparent_anime_idを解決できるようにする。
func insertSyncedParentWork(t *testing.T, db *sql.DB) (model.WorkID, model.AnimeID) {
	t.Helper()
	parentAnimeID := insertBareAnime(t, db)
	in := defaultSyncWorkInput()
	in.animeID = sql.NullInt64{Int64: int64(parentAnimeID), Valid: true}
	workID := insertSyncWork(t, db, in)
	return workID, parentAnimeID
}

func newSyncEpisodesUsecase(db *sql.DB) *SyncEpisodesToAnimesUsecase {
	queries := query.New(db)
	return NewSyncEpisodesToAnimesUsecase(
		db,
		repository.NewEpisodeRepository(queries),
		repository.NewAnimeRepository(queries),
		repository.NewAnimeClassificationRepository(queries),
	)
}

// reloadSyncEpisodeは同期ローダー経由でepisodeを読み直す。主に書き戻された
// anime_idを観測するため。
func reloadSyncEpisode(t *testing.T, db *sql.DB, episodeID model.EpisodeID) *model.Episode {
	t.Helper()
	episodeRepo := repository.NewEpisodeRepository(query.New(db))
	episodes, err := episodeRepo.ListForAnimeSyncByIDs(context.Background(), []model.EpisodeID{episodeID})
	if err != nil {
		t.Fatalf("episodeの再取得に失敗: %v", err)
	}
	if len(episodes) != 1 {
		t.Fatalf("episodeの再取得件数 = %d、期待値 = 1", len(episodes))
	}
	return episodes[0]
}

func TestSyncEpisodesToAnimesUsecase_Execute_CreatesAnimeForUnmappedEpisode(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncEpisodesUsecase(db)

	workID, parentAnimeID := insertSyncedParentWork(t, db)

	in := defaultSyncEpisodeInput(workID)
	in.title = sql.NullString{String: "第3話タイトル", Valid: true}
	in.titleRo = "Episode 3"
	in.number = sql.NullString{String: "第3話", Valid: true}
	in.sortNumber = 3
	in.rawNumber = sql.NullFloat64{Float64: 3.5, Valid: true}
	episodeID := insertSyncEpisode(t, db, in)

	result, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if result.Processed != 1 || result.Created != 1 || result.Updated != 0 || result.Unchanged != 0 || result.SkippedNoParent != 0 {
		t.Fatalf("result = %+v、期待値 = {Processed:1 Created:1 Updated:0 Unchanged:0 SkippedNoParent:0}", result)
	}

	episode := reloadSyncEpisode(t, db, episodeID)
	if episode.AnimeID == nil {
		t.Fatal("episodes.anime_id = nil、期待値 = 書き戻された値")
	}
	animeID := *episode.AnimeID

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Title.String != "第3話タイトル" {
		t.Errorf("anime.Title = %q、期待値 = 第3話タイトル", anime.Title.String)
	}
	if anime.TitleRo.String != "Episode 3" {
		t.Errorf("anime.TitleRo = %q、期待値 = Episode 3", anime.TitleRo.String)
	}
	if anime.Status != model.AnimeStatusPublished {
		t.Errorf("anime.Status = %q、期待値 = published", anime.Status)
	}
	// episodeはmediaを源泉としないため、第1層の行ではNULLのまま。
	if anime.Media != "" {
		t.Errorf("anime.Media = %q、期待値 = 空 (NULL)", anime.Media)
	}

	classRepo := repository.NewAnimeClassificationRepository(query.New(db))
	classification, err := classRepo.GetByAnimeID(context.Background(), animeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID()のclassification = %v、エラー = %v", classification, err)
	}
	if classification.Kind != model.AnimeClassificationKindEpisode {
		t.Errorf("classification.Kind = %q、期待値 = episode", classification.Kind)
	}
	if classification.ParentAnimeID == nil || *classification.ParentAnimeID != parentAnimeID {
		t.Errorf("classification.ParentAnimeID = %v、期待値 = %d", classification.ParentAnimeID, parentAnimeID)
	}
	if classification.Number.String != "3.5" {
		t.Errorf("classification.Number = %q、期待値 = 3.5", classification.Number.String)
	}
	if classification.NumberText.String != "第3話" {
		t.Errorf("classification.NumberText = %q、期待値 = 第3話", classification.NumberText.String)
	}
	if !classification.SortNumber.Valid || classification.SortNumber.Int32 != 3 {
		t.Errorf("classification.SortNumber = %+v、期待値 = {3 true}", classification.SortNumber)
	}
	if classification.Standalone {
		t.Error("エピソードのclassification.Standalone = true、期待値 = false")
	}
	// work専用の生成設定はepisode分類ではNULLのまま。
	if classification.NumberFormatID != nil {
		t.Errorf("classification.NumberFormatID = %v、期待値 = nil", classification.NumberFormatID)
	}
	if classification.EpisodeStartNumber.Valid || classification.ExpectedEpisodesCount.Valid {
		t.Error("エピソードではepisode_start_number / expected_episodes_countがNULLであるべきだが、値が入っていた")
	}
}

func TestSyncEpisodesToAnimesUsecase_Execute_CreatesRecapEpisodeWithoutNumber(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncEpisodesUsecase(db)

	workID, _ := insertSyncedParentWork(t, db)

	// 総集編タイプのepisodeは数値のraw_numberを持たないが表示用のnumber_textは
	// 持つ。episode分類はnumberのNULLを許容する (anime_classifications_number_check)
	// ため、同期はその制約に触れずに作成しなければならない。
	in := defaultSyncEpisodeInput(workID)
	in.rawNumber = sql.NullFloat64{}
	in.number = sql.NullString{String: "総集編", Valid: true}
	episodeID := insertSyncEpisode(t, db, in)

	result, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if result.Processed != 1 || result.Created != 1 || result.Updated != 0 || result.Unchanged != 0 || result.SkippedNoParent != 0 {
		t.Fatalf("result = %+v、期待値 = {Processed:1 Created:1 Updated:0 Unchanged:0 SkippedNoParent:0}", result)
	}

	episode := reloadSyncEpisode(t, db, episodeID)
	if episode.AnimeID == nil {
		t.Fatal("episodes.anime_id = nil、期待値 = 書き戻された値")
	}

	classRepo := repository.NewAnimeClassificationRepository(query.New(db))
	classification, err := classRepo.GetByAnimeID(context.Background(), *episode.AnimeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID()のclassification = %v、エラー = %v", classification, err)
	}
	if classification.Number.Valid {
		t.Errorf("classification.Number = %+v、期待値 = NULL (raw_numberが無いため)", classification.Number)
	}
	if classification.NumberText.String != "総集編" {
		t.Errorf("classification.NumberText = %q、期待値 = 総集編", classification.NumberText.String)
	}

	// 2回目の実行は差分なしを検出しなければならない。NULLのnumberが
	// チャーンせずにラウンドトリップすること (NULLのdesiredとNULLのexistingが
	// 等しく比較されること) の検証。
	second, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
	if err != nil {
		t.Fatalf("2回目のExecute()のエラー = %v", err)
	}
	if second.Processed != 1 || second.Created != 0 || second.Updated != 0 || second.Unchanged != 1 {
		t.Fatalf("2回目の結果 = %+v、期待値 = {Processed:1 Created:0 Updated:0 Unchanged:1}", second)
	}
}

func TestSyncEpisodesToAnimesUsecase_Execute_IsIdempotent(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncEpisodesUsecase(db)

	workID, _ := insertSyncedParentWork(t, db)
	in := defaultSyncEpisodeInput(workID)
	in.rawNumber = sql.NullFloat64{Float64: 3.5, Valid: true}
	in.number = sql.NullString{String: "3.5", Valid: true}
	episodeID := insertSyncEpisode(t, db, in)

	if _, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}}); err != nil {
		t.Fatalf("1回目のExecute()のエラー = %v", err)
	}

	// 同じepisodeに対する2回目の実行は差分なしを検出しなければならない
	// (status / NUMERICのラウンドトリップがチャーンを生まないことの検証)。
	result, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
	if err != nil {
		t.Fatalf("2回目のExecute()のエラー = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.Updated != 0 || result.Unchanged != 1 {
		t.Fatalf("result = %+v、期待値 = {Processed:1 Created:0 Updated:0 Unchanged:1}", result)
	}
}

func TestSyncEpisodesToAnimesUsecase_Execute_UpdatesChangedEpisode(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncEpisodesUsecase(db)

	workID, _ := insertSyncedParentWork(t, db)
	in := defaultSyncEpisodeInput(workID)
	in.title = sql.NullString{String: "旧話タイトル", Valid: true}
	episodeID := insertSyncEpisode(t, db, in)

	if _, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}}); err != nil {
		t.Fatalf("1回目のExecute()のエラー = %v", err)
	}

	if _, err := db.Exec(
		`UPDATE episodes SET title = $1, number = $2, sort_number = $3, raw_number = $4 WHERE id = $5`,
		"新話タイトル", "第2話", 2, 2.0, int64(episodeID),
	); err != nil {
		t.Fatalf("episodesの更新に失敗: %v", err)
	}

	result, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
	if err != nil {
		t.Fatalf("2回目のExecute()のエラー = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.Updated != 1 || result.Unchanged != 0 {
		t.Fatalf("result = %+v、期待値 = {Processed:1 Created:0 Updated:1 Unchanged:0}", result)
	}

	episode := reloadSyncEpisode(t, db, episodeID)
	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), *episode.AnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Title.String != "新話タイトル" {
		t.Errorf("anime.Title = %q、期待値 = 新話タイトル", anime.Title.String)
	}

	classRepo := repository.NewAnimeClassificationRepository(query.New(db))
	classification, err := classRepo.GetByAnimeID(context.Background(), *episode.AnimeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID()のclassification = %v、エラー = %v", classification, err)
	}
	if classification.Number.String != "2" {
		t.Errorf("classification.Number = %q、期待値 = 2", classification.Number.String)
	}
	if classification.NumberText.String != "第2話" {
		t.Errorf("classification.NumberText = %q、期待値 = 第2話", classification.NumberText.String)
	}
	if classification.SortNumber.Int32 != 2 {
		t.Errorf("classification.SortNumber = %d、期待値 = 2", classification.SortNumber.Int32)
	}
}

func TestSyncEpisodesToAnimesUsecase_Execute_UpdatesReparentedEpisode(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncEpisodesUsecase(db)

	// 最初の同期でepisodeを親作品Aの配下にマッピングする。
	workA, parentAnimeA := insertSyncedParentWork(t, db)
	episodeID := insertSyncEpisode(t, db, defaultSyncEpisodeInput(workA))
	if _, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}}); err != nil {
		t.Fatalf("1回目のExecute()のエラー = %v", err)
	}
	episode := reloadSyncEpisode(t, db, episodeID)
	animeID := *episode.AnimeID

	classRepo := repository.NewAnimeClassificationRepository(query.New(db))
	classification, err := classRepo.GetByAnimeID(context.Background(), animeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID()のclassification = %v、エラー = %v", classification, err)
	}
	if classification.ParentAnimeID == nil || *classification.ParentAnimeID != parentAnimeA {
		t.Fatalf("classification.ParentAnimeID = %v、期待値 = %d (親アニメA)", classification.ParentAnimeID, parentAnimeA)
	}

	// episodeを2つ目の同期済みwork Bへ付け替える。解決されるparent_anime_idが
	// 変わる (work分類はparent_anime_idが常にNULLのためworks同期では通らない経路) ので、
	// 次回の同期は分類のparent_anime_idを更新しなければならない。episode自身のanime行は
	// 変わらない。
	workB, parentAnimeB := insertSyncedParentWork(t, db)
	if _, err := db.Exec(`UPDATE episodes SET work_id = $1 WHERE id = $2`, int64(workB), int64(episodeID)); err != nil {
		t.Fatalf("episodesの付け替えに失敗: %v", err)
	}

	result, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
	if err != nil {
		t.Fatalf("2回目のExecute()のエラー = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.Updated != 1 || result.Unchanged != 0 {
		t.Fatalf("result = %+v、期待値 = {Processed:1 Created:0 Updated:1 Unchanged:0}", result)
	}

	classification, err = classRepo.GetByAnimeID(context.Background(), animeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID()のclassification = %v、エラー = %v", classification, err)
	}
	if classification.ParentAnimeID == nil || *classification.ParentAnimeID != parentAnimeB {
		t.Errorf("classification.ParentAnimeID = %v、期待値 = %d (Bへ付け替えられること)", classification.ParentAnimeID, parentAnimeB)
	}
}

func TestSyncEpisodesToAnimesUsecase_Execute_SkipsEpisodeWithUnsyncedParent(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncEpisodesUsecase(db)

	// 親作品が未同期 (anime_idなし) のため、episodeは今回はリコンサイルできず
	// 繰り延べられる。
	workID := insertSyncWork(t, db, defaultSyncWorkInput())
	episodeID := insertSyncEpisode(t, db, defaultSyncEpisodeInput(workID))

	result, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.SkippedNoParent != 1 {
		t.Fatalf("result = %+v、期待値 = {Processed:1 Created:0 ... SkippedNoParent:1}", result)
	}

	episode := reloadSyncEpisode(t, db, episodeID)
	if episode.AnimeID != nil {
		t.Errorf("episodes.anime_id = %v、期待値 = nil (エピソードの同期を保留するため)", episode.AnimeID)
	}
}

func TestSyncEpisodesToAnimesUsecase_Execute_PreservesUnsourcedAnimeFields(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncEpisodesUsecase(db)

	workID, _ := insertSyncedParentWork(t, db)
	episodeID := insertSyncEpisode(t, db, defaultSyncEpisodeInput(workID))
	if _, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}}); err != nil {
		t.Fatalf("1回目のExecute()のエラー = %v", err)
	}
	episode := reloadSyncEpisode(t, db, episodeID)
	animeID := *episode.AnimeID

	// episodeが源泉としないカラムに値があると仮定し、episode側を変更して同期に
	// UPDATEを発行させる。
	if _, err := db.Exec(
		`UPDATE animes SET media = 'tv', synopsis = $1, release_status = $2 WHERE id = $3`,
		"編集者が書いたあらすじ", string(model.ReleaseStatusReleased), int64(animeID),
	); err != nil {
		t.Fatalf("animesの事前更新に失敗: %v", err)
	}
	if _, err := db.Exec(`UPDATE episodes SET title = $1 WHERE id = $2`, "改題", int64(episodeID)); err != nil {
		t.Fatalf("episodesの更新に失敗: %v", err)
	}

	result, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
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
		t.Errorf("anime.Title = %q、期待値 = 改題 (エピソードの変更が反映されること)", anime.Title.String)
	}
	if anime.Media != model.AnimeMediaTV {
		t.Errorf("anime.Media = %q、期待値 = tv (保持されること)", anime.Media)
	}
	if anime.Synopsis.String != "編集者が書いたあらすじ" {
		t.Errorf("anime.Synopsis = %q、期待値 = 保持されること", anime.Synopsis.String)
	}
	if anime.ReleaseStatus != model.ReleaseStatusReleased {
		t.Errorf("anime.ReleaseStatus = %q、期待値 = released (保持されること)", anime.ReleaseStatus)
	}
}

func TestSyncEpisodesToAnimesUsecase_Execute_RecreatesMissingClassification(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncEpisodesUsecase(db)

	workID, parentAnimeID := insertSyncedParentWork(t, db)
	episodeID := insertSyncEpisode(t, db, defaultSyncEpisodeInput(workID))
	if _, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}}); err != nil {
		t.Fatalf("1回目のExecute()のエラー = %v", err)
	}
	episode := reloadSyncEpisode(t, db, episodeID)
	animeID := *episode.AnimeID

	// animeとepisodes.anime_idのマッピングは残したまま分類行だけを削除し、半端な
	// マッピング状態を再現する。次回の同期は分類を再作成して自己修復しなければならず、
	// 更新として数えられる。
	if _, err := db.Exec(`DELETE FROM anime_classifications WHERE anime_id = $1`, int64(animeID)); err != nil {
		t.Fatalf("anime_classificationsの削除に失敗: %v", err)
	}

	result, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
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
	if classification.Kind != model.AnimeClassificationKindEpisode {
		t.Errorf("classification.Kind = %q、期待値 = episode", classification.Kind)
	}
	if classification.ParentAnimeID == nil || *classification.ParentAnimeID != parentAnimeID {
		t.Errorf("classification.ParentAnimeID = %v、期待値 = %d (再作成されること)", classification.ParentAnimeID, parentAnimeID)
	}
}

func TestSyncEpisodesToAnimesUsecase_Execute_CreatesArchivedAnimeForUnpublishedEpisode(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncEpisodesUsecase(db)

	workID, _ := insertSyncedParentWork(t, db)

	// episodes.unpublished_atが立っている (エピソードレベルで非公開) 場合は
	// anime.status = archivedに写像される。
	in := defaultSyncEpisodeInput(workID)
	in.unpublishedAt = sql.NullTime{Time: time.Now(), Valid: true}
	episodeID := insertSyncEpisode(t, db, in)

	if _, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}}); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	episode := reloadSyncEpisode(t, db, episodeID)
	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), *episode.AnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Status != model.AnimeStatusArchived {
		t.Errorf("anime.Status = %q、期待値 = archived (episodes.unpublished_atが設定済みのため)", anime.Status)
	}
}

func TestSyncEpisodesToAnimesUsecase_Execute_CreatesDeletedAnimeForDeletedEpisode(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncEpisodesUsecase(db)

	workID, _ := insertSyncedParentWork(t, db)

	// episodes.deleted_atが立っている (エピソードレベルでソフトデリート) 場合は
	// anime.status = deletedに写像される。
	in := defaultSyncEpisodeInput(workID)
	in.deletedAt = sql.NullTime{Time: time.Now(), Valid: true}
	episodeID := insertSyncEpisode(t, db, in)

	if _, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}}); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	episode := reloadSyncEpisode(t, db, episodeID)
	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), *episode.AnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Status != model.AnimeStatusDeleted {
		t.Errorf("anime.Status = %q、期待値 = deleted (episodes.deleted_atが設定済みのため)", anime.Status)
	}
}

func TestSyncEpisodesToAnimesUsecase_Execute_ReconcilesEpisodeStateChangeToAnime(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncEpisodesUsecase(db)

	workID, _ := insertSyncedParentWork(t, db)
	episodeID := insertSyncEpisode(t, db, defaultSyncEpisodeInput(workID))
	if _, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}}); err != nil {
		t.Fatalf("1回目のExecute()のエラー = %v", err)
	}
	episode := reloadSyncEpisode(t, db, episodeID)
	animeID := *episode.AnimeID
	animeRepo := repository.NewAnimeRepository(query.New(db))

	// episodes.unpublished_atを立てるとanime.status = archivedへリコンサイルされる
	// (状態導出がエピソードレベルの変更をanimesへ伝播する)。
	if _, err := db.Exec(`UPDATE episodes SET unpublished_at = NOW() WHERE id = $1`, int64(episodeID)); err != nil {
		t.Fatalf("episodes.unpublished_atの更新に失敗: %v", err)
	}
	result, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
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

	// 続けてepisodes.deleted_atを立てるとanime.status = deletedへリコンサイルされる
	// (deleted_atがunpublished_atより優先される)。
	if _, err := db.Exec(`UPDATE episodes SET deleted_at = NOW() WHERE id = $1`, int64(episodeID)); err != nil {
		t.Fatalf("episodes.deleted_atの更新に失敗: %v", err)
	}
	result, err = uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
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

func TestSyncEpisodesToAnimesUsecase_Execute_DoesNotClobberAnimeArchivedState(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncEpisodesUsecase(db)

	workID, _ := insertSyncedParentWork(t, db)
	episodeID := insertSyncEpisode(t, db, defaultSyncEpisodeInput(workID))
	if _, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}}); err != nil {
		t.Fatalf("1回目のExecute()のエラー = %v", err)
	}
	episode := reloadSyncEpisode(t, db, episodeID)
	animeID := *episode.AnimeID
	animeRepo := repository.NewAnimeRepository(query.New(db))

	// animes-first + episodes両書きによる非公開経路を再現する。animes.statusを
	// archivedに (編集者が入力したarchive_message付きで)、episodes.unpublished_atを
	// 同時に立てる。これは導出の是正が守る不変条件で、リコンシラーはarchivedを
	// episodes.unpublished_atから導出し、animes.status = archivedをpublishedに戻さず
	// 据え置く。archive_messageはanimes専用のため、リコンシラーはepisodeから上書きせず
	// 既存の値を引き継ぐ。
	const archiveMessage = "編集者が入力したアーカイブメッセージ"
	if _, err := db.Exec(`UPDATE episodes SET unpublished_at = NOW() WHERE id = $1`, int64(episodeID)); err != nil {
		t.Fatalf("episodes.unpublished_atの更新に失敗: %v", err)
	}
	if _, err := db.Exec(
		`UPDATE animes SET status = 'archived', archive_message = $2 WHERE id = $1`,
		int64(animeID), archiveMessage,
	); err != nil {
		t.Fatalf("animes.status / archive_messageの更新に失敗: %v", err)
	}

	result, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
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
	if anime.ArchiveMessage.String != archiveMessage {
		t.Errorf("anime.ArchiveMessage = %q、期待値 = %q (上書きされないこと)", anime.ArchiveMessage.String, archiveMessage)
	}
}

func TestSyncEpisodesToAnimesUsecase_Execute_EmptyInput(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncEpisodesUsecase(db)

	result, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: nil})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if result.Processed != 0 || result.Created != 0 || result.Updated != 0 || result.Unchanged != 0 || result.SkippedNoParent != 0 {
		t.Fatalf("result = %+v、期待値 = すべて0", result)
	}
}

func TestPlanEpisodeAnimeSync_SkipsWhenParentUnresolved(t *testing.T) {
	t.Parallel()

	// 親作品が未同期 (ParentAnimeID nil) のepisodeは、episode分類がNOT NULLの
	// parent_anime_idを要するため、作成せず繰り延べなければならない。
	episode := &model.Episode{
		ID:         model.EpisodeID(1),
		WorkID:     model.WorkID(1),
		SortNumber: 1,
	}

	plan := planEpisodeAnimeSync(
		[]*model.Episode{episode},
		map[model.AnimeID]*model.Anime{},
		map[model.AnimeID]*model.AnimeClassification{},
	)

	if plan.processed != 1 || len(plan.creates) != 0 || len(plan.updates) != 0 || plan.unchanged != 0 || plan.skippedNoParent != 1 {
		t.Fatalf("plan = {processed:%d creates:%d updates:%d unchanged:%d skippedNoParent:%d}、期待値 = {1 0 0 0 1}",
			plan.processed, len(plan.creates), len(plan.updates), plan.unchanged, plan.skippedNoParent)
	}
}

func TestPlanEpisodeAnimeSync_CreatesAnimeWhenMappedRowMissing(t *testing.T) {
	t.Parallel()

	// anime_idがロード済み集合に存在しないanimeを指すepisode (宙ぶらりんの
	// マッピング)。episodes.anime_idの外部キーによりExecute経由ではこの状態に到達
	// できないため、純粋なプランナーを直接呼ぶ。更新ではなく新規作成 (anime + 分類) に
	// フォールバックしなければならない。
	parentAnimeID := model.AnimeID(1 << 40)
	danglingAnimeID := model.AnimeID(1<<40 + 1)
	episode := &model.Episode{
		ID:            model.EpisodeID(1),
		WorkID:        model.WorkID(1),
		SortNumber:    1,
		AnimeID:       &danglingAnimeID,
		ParentAnimeID: &parentAnimeID,
	}

	plan := planEpisodeAnimeSync(
		[]*model.Episode{episode},
		map[model.AnimeID]*model.Anime{},
		map[model.AnimeID]*model.AnimeClassification{},
	)

	if plan.processed != 1 || len(plan.creates) != 1 || len(plan.updates) != 0 || plan.skippedNoParent != 0 {
		t.Fatalf("plan = {processed:%d creates:%d updates:%d skippedNoParent:%d}、期待値 = {1 1 0 0}",
			plan.processed, len(plan.creates), len(plan.updates), plan.skippedNoParent)
	}
	create := plan.creates[0]
	if create.episodeID != episode.ID {
		t.Errorf("create.episodeID = %d、期待値 = %d", create.episodeID, episode.ID)
	}
	if create.classification.Kind != model.AnimeClassificationKindEpisode {
		t.Errorf("create.classification.Kind = %q、期待値 = episode", create.classification.Kind)
	}
	if create.classification.ParentAnimeID == nil || *create.classification.ParentAnimeID != parentAnimeID {
		t.Errorf("create.classification.ParentAnimeID = %v、期待値 = %d", create.classification.ParentAnimeID, parentAnimeID)
	}
}

// TestAnimeStatusFromEpisodeStatusはepisodeの導出ライフサイクル状態をanimeの
// status enumに写像する純粋なenumアダプタを検証する。timestampsからstatusへの
// 優先順位自体はmodel.Episode.DerivedStatusが持ち、TestEpisode_DerivedStatusで担保する。
func TestAnimeStatusFromEpisodeStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status model.EpisodeStatus
		want   model.AnimeStatus
	}{
		{model.EpisodeStatusPublished, model.AnimeStatusPublished},
		{model.EpisodeStatusArchived, model.AnimeStatusArchived},
		{model.EpisodeStatusDeleted, model.AnimeStatusDeleted},
	}
	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			if got := animeStatusFromEpisodeStatus(tt.status); got != tt.want {
				t.Errorf("animeStatusFromEpisodeStatus(%q) = %q、期待値 = %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestNumericStringFromFloatPtr(t *testing.T) {
	t.Parallel()

	if got := numericStringFromFloatPtr(nil); got.Valid {
		t.Errorf("numericStringFromFloatPtr(nil) = %+v、期待値 = NULL", got)
	}

	tests := []struct {
		in   float64
		want string
	}{
		{1.0, "1"},
		{2.5, "2.5"},
		{3.0, "3"},
	}
	for _, tt := range tests {
		got := numericStringFromFloatPtr(&tt.in)
		if !got.Valid || got.String != tt.want {
			t.Errorf("numericStringFromFloatPtr(%v) = %+v、期待値 = %q", tt.in, got, tt.want)
		}
	}
}
