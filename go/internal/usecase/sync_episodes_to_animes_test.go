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

// syncEpisodeInput holds the episodes columns relevant to the episodes -> animes
// sync. The helper inserts a real episodes row so the sync usecase (which opens its
// own transaction) can see it via GetTestDB.
//
// [Ja] syncEpisodeInput は episodes -> animes 同期に関係する episodes カラムを保持する。
// ヘルパーが実際の episodes 行を挿入し、(自前でトランザクションを開く) 同期 UseCase が
// GetTestDB 経由でその行を見られるようにする。
type syncEpisodeInput struct {
	workID         model.WorkID
	title          sql.NullString
	titleRo        string
	titleEn        string
	number         sql.NullString
	sortNumber     int32
	rawNumber      sql.NullFloat64
	status         string
	archiveMessage sql.NullString
	animeID        sql.NullInt64
}

// defaultSyncEpisodeInput returns a minimal published episode with a numeric
// number (raw_number=1, number="1") under the given parent work.
//
// [Ja] defaultSyncEpisodeInput は指定の親作品配下に、数値の話数 (raw_number=1,
// number="1") を持つ最小の公開エピソードを返す。
func defaultSyncEpisodeInput(workID model.WorkID) syncEpisodeInput {
	return syncEpisodeInput{
		workID:     workID,
		number:     sql.NullString{String: "1", Valid: true},
		sortNumber: 1,
		rawNumber:  sql.NullFloat64{Float64: 1, Valid: true},
		status:     "published",
	}
}

func insertSyncEpisode(t *testing.T, db *sql.DB, in syncEpisodeInput) model.EpisodeID {
	t.Helper()

	var id int64
	err := db.QueryRow(`
		INSERT INTO episodes (
			work_id, title, title_ro, title_en, number, sort_number,
			raw_number, status, archive_message, anime_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, NOW(), NOW()
		) RETURNING id
	`,
		int64(in.workID), in.title, in.titleRo, in.titleEn, in.number, in.sortNumber,
		in.rawNumber, in.status, in.archiveMessage, in.animeID,
	).Scan(&id)
	if err != nil {
		t.Fatalf("episodes の挿入に失敗: %v", err)
	}
	return model.EpisodeID(id)
}

// insertBareAnime inserts a bare anime row and returns its ID, used to stand in for
// a synced parent work's anime.
//
// [Ja] insertBareAnime は素の anime 行を挿入し ID を返す。同期済みの親作品の anime に
// 見立てて使う。
func insertBareAnime(t *testing.T, db *sql.DB) model.AnimeID {
	t.Helper()
	var id int64
	if err := db.QueryRow(`INSERT INTO animes (status) VALUES ('published') RETURNING id`).Scan(&id); err != nil {
		t.Fatalf("animes の挿入に失敗: %v", err)
	}
	return model.AnimeID(id)
}

// insertSyncedParentWork inserts a work already mapped to a fresh anime, so episodes
// under it resolve a non-NULL parent_anime_id.
//
// [Ja] insertSyncedParentWork は新規 anime にマッピング済みの work を挿入し、その配下の
// episodes が NOT NULL の parent_anime_id を解決できるようにする。
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

// reloadSyncEpisode re-reads an episode through the sync loader, mainly to observe
// the written-back anime_id.
//
// [Ja] reloadSyncEpisode は同期ローダー経由で episode を読み直す。主に書き戻された
// anime_id を観測するため。
func reloadSyncEpisode(t *testing.T, db *sql.DB, episodeID model.EpisodeID) *model.Episode {
	t.Helper()
	episodeRepo := repository.NewEpisodeRepository(query.New(db))
	episodes, err := episodeRepo.ListForAnimeSyncByIDs(context.Background(), []model.EpisodeID{episodeID})
	if err != nil {
		t.Fatalf("episode の再取得に失敗: %v", err)
	}
	if len(episodes) != 1 {
		t.Fatalf("episode の再取得件数 = %d, want 1", len(episodes))
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
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Processed != 1 || result.Created != 1 || result.Updated != 0 || result.Unchanged != 0 || result.SkippedNoParent != 0 {
		t.Fatalf("result = %+v, want {Processed:1 Created:1 Updated:0 Unchanged:0 SkippedNoParent:0}", result)
	}

	episode := reloadSyncEpisode(t, db, episodeID)
	if episode.AnimeID == nil {
		t.Fatal("episodes.anime_id should be written back, got nil")
	}
	animeID := *episode.AnimeID

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Title.String != "第3話タイトル" {
		t.Errorf("anime.Title = %q, want 第3話タイトル", anime.Title.String)
	}
	if anime.TitleRo.String != "Episode 3" {
		t.Errorf("anime.TitleRo = %q, want Episode 3", anime.TitleRo.String)
	}
	if anime.Status != model.AnimeStatusPublished {
		t.Errorf("anime.Status = %q, want published", anime.Status)
	}
	// Episodes do not source media; the layer-1 row leaves it NULL.
	//
	// [Ja] episode は media を源泉としないため、第 1 層の行では NULL のまま。
	if anime.Media != "" {
		t.Errorf("anime.Media = %q, want empty (NULL)", anime.Media)
	}

	classRepo := repository.NewAnimeClassificationRepository(query.New(db))
	classification, err := classRepo.GetByAnimeID(context.Background(), animeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID() classification=%v err=%v", classification, err)
	}
	if classification.Kind != model.AnimeClassificationKindEpisode {
		t.Errorf("classification.Kind = %q, want episode", classification.Kind)
	}
	if classification.ParentAnimeID == nil || *classification.ParentAnimeID != parentAnimeID {
		t.Errorf("classification.ParentAnimeID = %v, want %d", classification.ParentAnimeID, parentAnimeID)
	}
	if classification.Number.String != "3.5" {
		t.Errorf("classification.Number = %q, want 3.5", classification.Number.String)
	}
	if classification.NumberText.String != "第3話" {
		t.Errorf("classification.NumberText = %q, want 第3話", classification.NumberText.String)
	}
	if !classification.SortNumber.Valid || classification.SortNumber.Int32 != 3 {
		t.Errorf("classification.SortNumber = %+v, want {3 true}", classification.SortNumber)
	}
	if classification.Standalone {
		t.Error("classification.Standalone = true, want false for an episode")
	}
	// Work-only generation settings stay NULL for an episode classification.
	//
	// [Ja] work 専用の生成設定は episode 分類では NULL のまま。
	if classification.NumberFormatID != nil {
		t.Errorf("classification.NumberFormatID = %v, want nil", classification.NumberFormatID)
	}
	if classification.EpisodeStartNumber.Valid || classification.ExpectedEpisodesCount.Valid {
		t.Error("episode_start_number / expected_episodes_count should be NULL for an episode")
	}
}

func TestSyncEpisodesToAnimesUsecase_Execute_CreatesRecapEpisodeWithoutNumber(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncEpisodesUsecase(db)

	workID, _ := insertSyncedParentWork(t, db)

	// A recap-style episode (総集編) has no numeric raw_number but keeps a display
	// number_text. The episode classification allows number to be NULL
	// (anime_classifications_number_check), so the sync must create it without
	// tripping that constraint.
	//
	// [Ja] 総集編タイプの episode は数値の raw_number を持たないが表示用の number_text は
	// 持つ。episode 分類は number の NULL を許容する (anime_classifications_number_check)
	// ため、同期はその制約に触れずに作成しなければならない。
	in := defaultSyncEpisodeInput(workID)
	in.rawNumber = sql.NullFloat64{}
	in.number = sql.NullString{String: "総集編", Valid: true}
	episodeID := insertSyncEpisode(t, db, in)

	result, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Processed != 1 || result.Created != 1 || result.Updated != 0 || result.Unchanged != 0 || result.SkippedNoParent != 0 {
		t.Fatalf("result = %+v, want {Processed:1 Created:1 Updated:0 Unchanged:0 SkippedNoParent:0}", result)
	}

	episode := reloadSyncEpisode(t, db, episodeID)
	if episode.AnimeID == nil {
		t.Fatal("episodes.anime_id should be written back, got nil")
	}

	classRepo := repository.NewAnimeClassificationRepository(query.New(db))
	classification, err := classRepo.GetByAnimeID(context.Background(), *episode.AnimeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID() classification=%v err=%v", classification, err)
	}
	if classification.Number.Valid {
		t.Errorf("classification.Number = %+v, want NULL (no raw_number)", classification.Number)
	}
	if classification.NumberText.String != "総集編" {
		t.Errorf("classification.NumberText = %q, want 総集編", classification.NumberText.String)
	}

	// A second run must detect no diff: the NULL number must round-trip without
	// churning (NULL desired vs NULL existing compares equal).
	//
	// [Ja] 2 回目の実行は差分なしを検出しなければならない。NULL の number が
	// チャーンせずにラウンドトリップすること (NULL の desired と NULL の existing が
	// 等しく比較されること) の検証。
	second, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
	if err != nil {
		t.Fatalf("second Execute() error = %v", err)
	}
	if second.Processed != 1 || second.Created != 0 || second.Updated != 0 || second.Unchanged != 1 {
		t.Fatalf("second result = %+v, want {Processed:1 Created:0 Updated:0 Unchanged:1}", second)
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
		t.Fatalf("first Execute() error = %v", err)
	}

	// A second run over the same episode must detect no diff (validates that the
	// status / NUMERIC round-trips do not churn).
	//
	// [Ja] 同じ episode に対する 2 回目の実行は差分なしを検出しなければならない
	// (status / NUMERIC のラウンドトリップがチャーンを生まないことの検証)。
	result, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
	if err != nil {
		t.Fatalf("second Execute() error = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.Updated != 0 || result.Unchanged != 1 {
		t.Fatalf("result = %+v, want {Processed:1 Created:0 Updated:0 Unchanged:1}", result)
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
		t.Fatalf("first Execute() error = %v", err)
	}

	if _, err := db.Exec(
		`UPDATE episodes SET title = $1, number = $2, sort_number = $3, raw_number = $4 WHERE id = $5`,
		"新話タイトル", "第2話", 2, 2.0, int64(episodeID),
	); err != nil {
		t.Fatalf("episodes の更新に失敗: %v", err)
	}

	result, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
	if err != nil {
		t.Fatalf("second Execute() error = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.Updated != 1 || result.Unchanged != 0 {
		t.Fatalf("result = %+v, want {Processed:1 Created:0 Updated:1 Unchanged:0}", result)
	}

	episode := reloadSyncEpisode(t, db, episodeID)
	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), *episode.AnimeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Title.String != "新話タイトル" {
		t.Errorf("anime.Title = %q, want 新話タイトル", anime.Title.String)
	}

	classRepo := repository.NewAnimeClassificationRepository(query.New(db))
	classification, err := classRepo.GetByAnimeID(context.Background(), *episode.AnimeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID() classification=%v err=%v", classification, err)
	}
	if classification.Number.String != "2" {
		t.Errorf("classification.Number = %q, want 2", classification.Number.String)
	}
	if classification.NumberText.String != "第2話" {
		t.Errorf("classification.NumberText = %q, want 第2話", classification.NumberText.String)
	}
	if classification.SortNumber.Int32 != 2 {
		t.Errorf("classification.SortNumber = %d, want 2", classification.SortNumber.Int32)
	}
}

func TestSyncEpisodesToAnimesUsecase_Execute_UpdatesReparentedEpisode(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncEpisodesUsecase(db)

	// First sync maps the episode under parent work A.
	//
	// [Ja] 最初の同期で episode を親作品 A の配下にマッピングする。
	workA, parentAnimeA := insertSyncedParentWork(t, db)
	episodeID := insertSyncEpisode(t, db, defaultSyncEpisodeInput(workA))
	if _, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}}); err != nil {
		t.Fatalf("first Execute() error = %v", err)
	}
	episode := reloadSyncEpisode(t, db, episodeID)
	animeID := *episode.AnimeID

	classRepo := repository.NewAnimeClassificationRepository(query.New(db))
	classification, err := classRepo.GetByAnimeID(context.Background(), animeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID() classification=%v err=%v", classification, err)
	}
	if classification.ParentAnimeID == nil || *classification.ParentAnimeID != parentAnimeA {
		t.Fatalf("classification.ParentAnimeID = %v, want %d (parent A)", classification.ParentAnimeID, parentAnimeA)
	}

	// Re-parent the episode to a second synced work B. The resolved parent_anime_id
	// changes (works sync never exercises this, since a work classification keeps
	// parent_anime_id NULL), so the next sync must update the classification's
	// parent_anime_id. The episode's own anime row stays the same.
	//
	// [Ja] episode を 2 つ目の同期済み work B へ付け替える。解決される parent_anime_id が
	// 変わる (work 分類は parent_anime_id が常に NULL のため works 同期では通らない経路) ので、
	// 次回の同期は分類の parent_anime_id を更新しなければならない。episode 自身の anime 行は
	// 変わらない。
	workB, parentAnimeB := insertSyncedParentWork(t, db)
	if _, err := db.Exec(`UPDATE episodes SET work_id = $1 WHERE id = $2`, int64(workB), int64(episodeID)); err != nil {
		t.Fatalf("episodes の付け替えに失敗: %v", err)
	}

	result, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
	if err != nil {
		t.Fatalf("second Execute() error = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.Updated != 1 || result.Unchanged != 0 {
		t.Fatalf("result = %+v, want {Processed:1 Created:0 Updated:1 Unchanged:0}", result)
	}

	classification, err = classRepo.GetByAnimeID(context.Background(), animeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID() classification=%v err=%v", classification, err)
	}
	if classification.ParentAnimeID == nil || *classification.ParentAnimeID != parentAnimeB {
		t.Errorf("classification.ParentAnimeID = %v, want %d (re-parented to B)", classification.ParentAnimeID, parentAnimeB)
	}
}

func TestSyncEpisodesToAnimesUsecase_Execute_SkipsEpisodeWithUnsyncedParent(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncEpisodesUsecase(db)

	// Parent work is not yet synced (no anime_id), so the episode cannot be
	// reconciled this run and is deferred.
	//
	// [Ja] 親作品が未同期 (anime_id なし) のため、episode は今回はリコンサイルできず
	// 繰り延べられる。
	workID := insertSyncWork(t, db, defaultSyncWorkInput())
	episodeID := insertSyncEpisode(t, db, defaultSyncEpisodeInput(workID))

	result, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.SkippedNoParent != 1 {
		t.Fatalf("result = %+v, want {Processed:1 Created:0 ... SkippedNoParent:1}", result)
	}

	episode := reloadSyncEpisode(t, db, episodeID)
	if episode.AnimeID != nil {
		t.Errorf("episodes.anime_id = %v, want nil (episode deferred)", episode.AnimeID)
	}
}

func TestSyncEpisodesToAnimesUsecase_Execute_PreservesUnsourcedAnimeFields(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncEpisodesUsecase(db)

	workID, _ := insertSyncedParentWork(t, db)
	episodeID := insertSyncEpisode(t, db, defaultSyncEpisodeInput(workID))
	if _, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}}); err != nil {
		t.Fatalf("first Execute() error = %v", err)
	}
	episode := reloadSyncEpisode(t, db, episodeID)
	animeID := *episode.AnimeID

	// Simulate values on the episode's anime that episodes do not source, then
	// force an episode change so the sync issues an UPDATE.
	//
	// [Ja] episode が源泉としないカラムに値があると仮定し、episode 側を変更して同期に
	// UPDATE を発行させる。
	if _, err := db.Exec(
		`UPDATE animes SET media = 'tv', synopsis = $1, release_status = $2 WHERE id = $3`,
		"編集者が書いたあらすじ", string(model.ReleaseStatusReleased), int64(animeID),
	); err != nil {
		t.Fatalf("animes の事前更新に失敗: %v", err)
	}
	if _, err := db.Exec(`UPDATE episodes SET title = $1 WHERE id = $2`, "改題", int64(episodeID)); err != nil {
		t.Fatalf("episodes の更新に失敗: %v", err)
	}

	result, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
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
		t.Errorf("anime.Title = %q, want 改題 (episode change applied)", anime.Title.String)
	}
	if anime.Media != model.AnimeMediaTV {
		t.Errorf("anime.Media = %q, want tv (preserved)", anime.Media)
	}
	if anime.Synopsis.String != "編集者が書いたあらすじ" {
		t.Errorf("anime.Synopsis = %q, want preserved", anime.Synopsis.String)
	}
	if anime.ReleaseStatus != model.ReleaseStatusReleased {
		t.Errorf("anime.ReleaseStatus = %q, want released (preserved)", anime.ReleaseStatus)
	}
}

func TestSyncEpisodesToAnimesUsecase_Execute_RecreatesMissingClassification(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncEpisodesUsecase(db)

	workID, parentAnimeID := insertSyncedParentWork(t, db)
	episodeID := insertSyncEpisode(t, db, defaultSyncEpisodeInput(workID))
	if _, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}}); err != nil {
		t.Fatalf("first Execute() error = %v", err)
	}
	episode := reloadSyncEpisode(t, db, episodeID)
	animeID := *episode.AnimeID

	// Drop only the classification row, keeping the anime and the episodes.anime_id
	// mapping, to simulate a half-built mapping. The next sync must self-heal by
	// re-creating the classification, which counts as an update.
	//
	// [Ja] anime と episodes.anime_id のマッピングは残したまま分類行だけを削除し、半端な
	// マッピング状態を再現する。次回の同期は分類を再作成して自己修復しなければならず、
	// 更新として数えられる。
	if _, err := db.Exec(`DELETE FROM anime_classifications WHERE anime_id = $1`, int64(animeID)); err != nil {
		t.Fatalf("anime_classifications の削除に失敗: %v", err)
	}

	result, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: []model.EpisodeID{episodeID}})
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
	if classification.Kind != model.AnimeClassificationKindEpisode {
		t.Errorf("classification.Kind = %q, want episode", classification.Kind)
	}
	if classification.ParentAnimeID == nil || *classification.ParentAnimeID != parentAnimeID {
		t.Errorf("classification.ParentAnimeID = %v, want %d (re-created)", classification.ParentAnimeID, parentAnimeID)
	}
}

func TestSyncEpisodesToAnimesUsecase_Execute_EmptyInput(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newSyncEpisodesUsecase(db)

	result, err := uc.Execute(context.Background(), SyncEpisodesToAnimesInput{EpisodeIDs: nil})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Processed != 0 || result.Created != 0 || result.Updated != 0 || result.Unchanged != 0 || result.SkippedNoParent != 0 {
		t.Fatalf("result = %+v, want all zero", result)
	}
}

func TestPlanEpisodeAnimeSync_SkipsWhenParentUnresolved(t *testing.T) {
	t.Parallel()

	// An episode whose parent work is not yet synced (ParentAnimeID nil) must be
	// deferred, not created, because the episode classification requires a non-NULL
	// parent_anime_id.
	//
	// [Ja] 親作品が未同期 (ParentAnimeID nil) の episode は、episode 分類が NOT NULL の
	// parent_anime_id を要するため、作成せず繰り延べなければならない。
	episode := &model.Episode{
		ID:         model.EpisodeID(1),
		WorkID:     model.WorkID(1),
		SortNumber: 1,
		Status:     model.EpisodeStatusPublished,
	}

	plan := planEpisodeAnimeSync(
		[]*model.Episode{episode},
		map[model.AnimeID]*model.Anime{},
		map[model.AnimeID]*model.AnimeClassification{},
	)

	if plan.processed != 1 || len(plan.creates) != 0 || len(plan.updates) != 0 || plan.unchanged != 0 || plan.skippedNoParent != 1 {
		t.Fatalf("plan = {processed:%d creates:%d updates:%d unchanged:%d skippedNoParent:%d}, want {1 0 0 0 1}",
			plan.processed, len(plan.creates), len(plan.updates), plan.unchanged, plan.skippedNoParent)
	}
}

func TestPlanEpisodeAnimeSync_CreatesAnimeWhenMappedRowMissing(t *testing.T) {
	t.Parallel()

	// An episode whose anime_id points at an anime absent from the loaded set (a
	// dangling mapping). The episodes.anime_id foreign key makes this unreachable
	// through Execute, so the pure planner is exercised directly: it must fall back
	// to creating a fresh anime + classification rather than attempting an update.
	//
	// [Ja] anime_id がロード済み集合に存在しない anime を指す episode (宙ぶらりんの
	// マッピング)。episodes.anime_id の外部キーにより Execute 経由ではこの状態に到達
	// できないため、純粋なプランナーを直接呼ぶ。更新ではなく新規作成 (anime + 分類) に
	// フォールバックしなければならない。
	parentAnimeID := model.AnimeID(1 << 40)
	danglingAnimeID := model.AnimeID(1<<40 + 1)
	episode := &model.Episode{
		ID:            model.EpisodeID(1),
		WorkID:        model.WorkID(1),
		SortNumber:    1,
		Status:        model.EpisodeStatusPublished,
		AnimeID:       &danglingAnimeID,
		ParentAnimeID: &parentAnimeID,
	}

	plan := planEpisodeAnimeSync(
		[]*model.Episode{episode},
		map[model.AnimeID]*model.Anime{},
		map[model.AnimeID]*model.AnimeClassification{},
	)

	if plan.processed != 1 || len(plan.creates) != 1 || len(plan.updates) != 0 || plan.skippedNoParent != 0 {
		t.Fatalf("plan = {processed:%d creates:%d updates:%d skippedNoParent:%d}, want {1 1 0 0}",
			plan.processed, len(plan.creates), len(plan.updates), plan.skippedNoParent)
	}
	create := plan.creates[0]
	if create.episodeID != episode.ID {
		t.Errorf("create.episodeID = %d, want %d", create.episodeID, episode.ID)
	}
	if create.classification.Kind != model.AnimeClassificationKindEpisode {
		t.Errorf("create.classification.Kind = %q, want episode", create.classification.Kind)
	}
	if create.classification.ParentAnimeID == nil || *create.classification.ParentAnimeID != parentAnimeID {
		t.Errorf("create.classification.ParentAnimeID = %v, want %d", create.classification.ParentAnimeID, parentAnimeID)
	}
}

func TestEpisodeStatusToAnimeStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status model.EpisodeStatus
		want   model.AnimeStatus
	}{
		{model.EpisodeStatusPublished, model.AnimeStatusPublished},
		{model.EpisodeStatusArchived, model.AnimeStatusArchived},
		{model.EpisodeStatusDeleted, model.AnimeStatusDeleted},
		{model.EpisodeStatus(""), model.AnimeStatusPublished},
	}
	for _, tt := range tests {
		if got := episodeStatusToAnimeStatus(tt.status); got != tt.want {
			t.Errorf("episodeStatusToAnimeStatus(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestNumericStringFromFloatPtr(t *testing.T) {
	t.Parallel()

	if got := numericStringFromFloatPtr(nil); got.Valid {
		t.Errorf("numericStringFromFloatPtr(nil) = %+v, want NULL", got)
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
			t.Errorf("numericStringFromFloatPtr(%v) = %+v, want %q", tt.in, got, tt.want)
		}
	}
}
