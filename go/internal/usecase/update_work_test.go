package usecase

import (
	"context"
	"database/sql"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

// newUpdateWorkUsecase wires the update-work usecase against the shared test DB. Like
// the create-work usecase it opens its own transaction, so its tests use GetTestDB (not
// SetupTx) so the committed rows are visible to the usecase's inner transaction and to
// the follow-up sync invariant check.
//
// [Ja] newUpdateWorkUsecase は共有テスト DB に対して作品更新 UseCase を組み立てる。
// 作成 UseCase と同じく内部で自前のトランザクションを開くため、テストは SetupTx ではなく
// GetTestDB を使い、コミット済みの行が UseCase の内側トランザクションと後続の同期不変条件
// チェックから見えるようにする。
func newUpdateWorkUsecase(db *sql.DB) *UpdateWorkUsecase {
	queries := query.New(db)
	return NewUpdateWorkUsecase(
		db,
		repository.NewWorkRepository(queries),
		repository.NewAnimeRepository(queries),
		repository.NewAnimeClassificationRepository(queries),
		newTestWorkSatelliteRepos(queries),
		validator.NewDBWorkCreateValidator(),
	)
}

// validUpdateWorkInput returns a form input that passes DBWorkCreateValidator with
// values that differ from validCreateWorkInput, so a test can observe the update taking
// effect (media TV instead of OVA, standalone flipped off, a different episode count).
//
// [Ja] validUpdateWorkInput は DBWorkCreateValidator を通過し、validCreateWorkInput とは
// 異なる値のフォーム入力を返す。更新が反映されたことをテストで観測できるようにする
// (メディアが OVA でなく TV、standalone を off に反転、話数も別の値)。
func validUpdateWorkInput(workID model.WorkID, title string) UpdateWorkInput {
	return UpdateWorkInput{
		WorkID: workID,
		WorkFormInput: WorkFormInput{
			Title:                 title,
			TitleKana:             "こうしんてすとあにめ",
			TitleEn:               "Update Test Anime",
			Media:                 "1", // TV
			Synopsis:              "更新後のあらすじ本文",
			SynopsisSource:        "更新後の出典",
			ManualEpisodesCount:   "24",
			StartEpisodeRawNumber: "1",
			NoEpisodes:            "",
			// Satellite source fields differ from validCreateWorkInput to exercise every
			// reconcile path: a changed value updates in place (official site, x account,
			// syobocal id, broadcast), a changed natural key deletes + creates (season), and
			// a cleared value deletes the row (wikipedia, hashtag, mal id, broadcast end).
			//
			// [Ja] 別表ソースのフィールドを validCreateWorkInput と変えて全リコンサイル経路を
			// 動かす: 値の変更はその場更新 (公式サイト・x アカウント・syobocal id・放送)、自然
			// キーの変更は削除 + 作成 (季節)、値のクリアは行削除 (wikipedia・ハッシュタグ・mal
			// id・放送終了日)。
			SeasonYear:      "2025",
			SeasonName:      "3", // summer
			StartedOn:       "2025-07-01",
			EndedOn:         "",
			OfficialSiteURL: "https://example.com/anime-v2",
			WikipediaURL:    "",
			TwitterUsername: "anime_official_2",
			TwitterHashtag:  "",
			ScTid:           "9999",
			MalAnimeID:      "",
		},
	}
}

// createMappedWork creates a work via the create usecase (which maps it to a fresh
// anime) and returns its ID, for tests that need an existing mapped work to update.
//
// [Ja] createMappedWork は作成 UseCase 経由で work を作成し (新規 anime にマッピング
// される)、その ID を返す。更新対象の既存マッピング済み work が要るテストで使う。
func createMappedWork(t *testing.T, db *sql.DB, title string) model.WorkID {
	t.Helper()
	output, err := newCreateWorkUsecase(db).Execute(context.Background(), validCreateWorkInput(title))
	if err != nil {
		t.Fatalf("前提の作品作成に失敗: %v", err)
	}
	return output.WorkID
}

func TestUpdateWorkUsecase_Execute_UpdatesWorkAnimeAndClassification(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateWorkUsecase(db)

	workID := createMappedWork(t, db, "更新前アニメ_"+t.Name())

	newTitle := "更新後アニメ_" + t.Name()
	if _, err := uc.Execute(context.Background(), validUpdateWorkInput(workID, newTitle)); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// The works row (the source of truth) reflects the submitted changes.
	//
	// [Ja] works 行 (正本) が送信された変更を反映している。
	work := reloadSyncWork(t, db, workID)
	if work.Title != newTitle {
		t.Errorf("work.Title = %q, want %q", work.Title, newTitle)
	}
	if work.Media != workMediaTV {
		t.Errorf("work.Media = %d, want %d (tv)", work.Media, workMediaTV)
	}
	if work.NoEpisodes {
		t.Error("work.NoEpisodes = true, want false")
	}
	if work.AnimeID == nil {
		t.Fatal("work.AnimeID should stay mapped, got nil")
	}
	animeID := *work.AnimeID

	// The mapped anime mirrors the updated content, and its status stays published
	// (the edit form does not touch status).
	//
	// [Ja] マッピング済み anime が更新後の内容を写し、status は published のまま
	// (編集フォームは status を触らない)。
	anime, err := repository.NewAnimeRepository(query.New(db)).GetByID(context.Background(), animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Title.String != newTitle {
		t.Errorf("anime.Title = %q, want %q", anime.Title.String, newTitle)
	}
	if anime.Media != model.AnimeMediaTV {
		t.Errorf("anime.Media = %q, want tv", anime.Media)
	}
	if anime.Synopsis.String != "更新後のあらすじ本文" {
		t.Errorf("anime.Synopsis = %q, want 更新後のあらすじ本文", anime.Synopsis.String)
	}
	if anime.Status != model.AnimeStatusPublished {
		t.Errorf("anime.Status = %q, want published", anime.Status)
	}

	classification, err := repository.NewAnimeClassificationRepository(query.New(db)).GetByAnimeID(context.Background(), animeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID() classification=%v err=%v", classification, err)
	}
	if classification.Standalone {
		t.Error("classification.Standalone = true, want false (no_episodes unchecked)")
	}
	if classification.EpisodeStartNumber.String != "1" {
		t.Errorf("classification.EpisodeStartNumber = %q, want 1", classification.EpisodeStartNumber.String)
	}
	if !classification.ExpectedEpisodesCount.Valid || classification.ExpectedEpisodesCount.Int32 != 24 {
		t.Errorf("classification.ExpectedEpisodesCount = %+v, want {24 true}", classification.ExpectedEpisodesCount)
	}
}

// TestUpdateWorkUsecase_Execute_ProducesSyncConsistentMapping is the update counterpart
// of the create invariant: a sync run right after an update must detect no diff
// (Unchanged), proving the update writes the same anime / classification the sync would
// derive from the updated works row.
//
// [Ja] TestUpdateWorkUsecase_Execute_ProducesSyncConsistentMapping は作成側の不変条件の
// 更新版。更新直後の同期実行は差分なし (Unchanged) を検出しなければならず、更新が、同期が
// 更新後の works 行から導出するのと同じ anime / 分類を書いていることを示す。
func TestUpdateWorkUsecase_Execute_ProducesSyncConsistentMapping(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateWorkUsecase(db)

	workID := createMappedWork(t, db, "更新前アニメ_"+t.Name())
	if _, err := uc.Execute(context.Background(), validUpdateWorkInput(workID, "更新後アニメ_"+t.Name())); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result, err := newSyncUsecase(db).Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("sync Execute() error = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.Updated != 0 || result.Unchanged != 1 {
		t.Fatalf("sync result = %+v, want {Processed:1 Created:0 Updated:0 Unchanged:1}", result)
	}
}

// TestUpdateWorkUsecase_Execute_ReconcilesSatelliteRows verifies the update dual-write
// reconciles every satellite table against the submitted values: a changed value updates
// in place (official site / x account / syobocal id / broadcast), a changed natural key
// replaces the row (season), and a cleared value deletes the row (wikipedia / hashtag /
// mal id / broadcast end).
//
// [Ja] TestUpdateWorkUsecase_Execute_ReconcilesSatelliteRows は、更新の両書きが送信値に対して
// 各別表をリコンサイルすることを検証する。値の変更はその場更新 (公式サイト / x アカウント /
// syobocal id / 放送)、自然キーの変更は行の置換 (季節)、値のクリアは行削除 (wikipedia /
// ハッシュタグ / mal id / 放送終了日)。
func TestUpdateWorkUsecase_Execute_ReconcilesSatelliteRows(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateWorkUsecase(db)
	ctx := context.Background()

	workID := createMappedWork(t, db, "別表更新前_"+t.Name())
	if _, err := uc.Execute(ctx, validUpdateWorkInput(workID, "別表更新後_"+t.Name())); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	animeID := *reloadSyncWork(t, db, workID).AnimeID
	queries := query.New(db)
	animeIDs := []model.AnimeID{animeID}

	// syobocal is updated in place; mal is deleted (its source was cleared).
	//
	// [Ja] syobocal はその場更新、mal は削除 (ソースがクリアされた)。
	externalIDs, err := repository.NewAnimeExternalIDRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(external_ids) error = %v", err)
	}
	if len(externalIDs) != 1 || externalIDs[0].Service != model.AnimeExternalServiceSyobocal || externalIDs[0].ExternalID != "9999" {
		t.Errorf("external_ids = %+v, want only syobocal=9999", externalIDs)
	}

	// official_site is updated in place; wikipedia is deleted (its source was cleared).
	//
	// [Ja] official_site はその場更新、wikipedia は削除 (ソースがクリアされた)。
	links, err := repository.NewAnimeLinkRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(links) error = %v", err)
	}
	if len(links) != 1 || links[0].Kind != model.AnimeLinkKindOfficialSite || links[0].Language != model.LanguageJa || links[0].URL != "https://example.com/anime-v2" {
		t.Errorf("links = %+v, want only official_site/ja=https://example.com/anime-v2", links)
	}

	accounts, err := repository.NewAnimeOfficialAccountRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(official_accounts) error = %v", err)
	}
	if len(accounts) != 1 || accounts[0].Account != "anime_official_2" {
		t.Errorf("accounts = %+v, want only x=anime_official_2", accounts)
	}

	hashtags, err := repository.NewAnimeHashtagRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(hashtags) error = %v", err)
	}
	if len(hashtags) != 0 {
		t.Errorf("hashtags = %+v, want none (source cleared)", hashtags)
	}

	// The season key changed (2024 spring -> 2025 summer), so the old row is deleted and a
	// new one created, leaving a single is_primary row.
	//
	// [Ja] 季節キーが変わった (2024 春 -> 2025 夏) ため旧行を削除し新行を作成、is_primary 行は
	// 1 つのまま。
	seasons, err := repository.NewAnimeSeasonRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(seasons) error = %v", err)
	}
	if len(seasons) != 1 || seasons[0].Year != 2025 || seasons[0].Name == nil || *seasons[0].Name != model.SeasonNameSummer || !seasons[0].IsPrimary {
		t.Errorf("seasons = %+v, want only 2025 summer is_primary", seasons)
	}

	// The broadcast event is updated in place: the start date changed and the end date was
	// cleared to open-ended (nil).
	//
	// [Ja] 放送イベントはその場更新: 開始日が変わり終了日は未定 (nil) にクリアされた。
	events, err := repository.NewAnimeEventRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(events) error = %v", err)
	}
	if len(events) != 1 || events[0].StartedOn.Format("2006-01-02") != "2025-07-01" || events[0].EndedOn != nil {
		t.Errorf("events = %+v, want one broadcast started 2025-07-01 ended nil", events)
	}
}

// TestUpdateWorkUsecase_Execute_ProducesSyncConsistentSatellites extends the update
// invariant to the satellite tables: the phase 2 satellite sync run right after an update
// must detect no diff, proving the update reconcile leaves the anime's satellite rows equal
// to what the sync derives from the updated works row.
//
// [Ja] TestUpdateWorkUsecase_Execute_ProducesSyncConsistentSatellites は update の不変条件を
// 別表に広げる。更新直後のフェーズ 2 別表同期は差分を検出してはならず、更新のリコンサイルが
// anime の別表行を、同期が更新後の works 行から導出するのと等しく保つことを示す。
func TestUpdateWorkUsecase_Execute_ProducesSyncConsistentSatellites(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateWorkUsecase(db)

	workID := createMappedWork(t, db, "別表整合更新前_"+t.Name())
	if _, err := uc.Execute(context.Background(), validUpdateWorkInput(workID, "別表整合更新後_"+t.Name())); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result, err := newSyncSatellitesUsecase(db).Execute(context.Background(), SyncWorkSatellitesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("satellite sync Execute() error = %v", err)
	}
	if result.Processed != 1 || result.SkippedNoAnime != 0 {
		t.Fatalf("satellite sync result = %+v, want Processed:1 SkippedNoAnime:0", result)
	}
	if result.Created != 0 || result.Updated != 0 || result.Deleted != 0 {
		t.Fatalf("satellite sync detected a diff = %+v, want Created:0 Updated:0 Deleted:0", result)
	}
}

// TestUpdateWorkUsecase_Execute_PreservesNonFormAnimeColumns verifies the update does
// not clobber the anime-mapped columns the edit form does not submit (title_ro / status
// / archive_message): they are carried over from the works row, so an archived work with
// a romanized title keeps both after a content edit, and a follow-up sync stays Unchanged.
//
// [Ja] TestUpdateWorkUsecase_Execute_PreservesNonFormAnimeColumns は、更新が編集フォーム
// の送信しない anime 写像カラム (title_ro / status / archive_message) を潰さないことを
// 検証する。これらは works 行から引き継がれるため、ローマ字タイトルを持つアーカイブ済み
// work は内容編集後も両方を保ち、後続の同期も Unchanged のままになる。
func TestUpdateWorkUsecase_Execute_PreservesNonFormAnimeColumns(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateWorkUsecase(db)

	workID := createMappedWork(t, db, "更新前アニメ_"+t.Name())
	work := reloadSyncWork(t, db, workID)
	animeID := *work.AnimeID

	// Simulate an archived work with a Rails-set romanized title, keeping works and the
	// mapped anime consistent (as the sync would leave them).
	//
	// [Ja] Rails が付けたローマ字タイトルを持つアーカイブ済み work を模し、works と
	// マッピング済み anime を (同期が残すのと同じく) 整合させておく。
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, "UPDATE works SET title_ro = $2, status = 'archived', archive_message = $3 WHERE id = $1", int64(workID), "Koushin Anime", "凍結中"); err != nil {
		t.Fatalf("works の前提更新に失敗: %v", err)
	}
	if _, err := db.ExecContext(ctx, "UPDATE animes SET title_ro = $2, status = 'archived', archive_message = $3 WHERE id = $1", int64(animeID), "Koushin Anime", "凍結中"); err != nil {
		t.Fatalf("animes の前提更新に失敗: %v", err)
	}

	if _, err := uc.Execute(ctx, validUpdateWorkInput(workID, "更新後アニメ_"+t.Name())); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	anime, err := repository.NewAnimeRepository(query.New(db)).GetByID(ctx, animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.TitleRo.String != "Koushin Anime" {
		t.Errorf("anime.TitleRo = %q, want Koushin Anime (preserved)", anime.TitleRo.String)
	}
	if anime.Status != model.AnimeStatusArchived {
		t.Errorf("anime.Status = %q, want archived (preserved)", anime.Status)
	}
	if anime.ArchiveMessage.String != "凍結中" {
		t.Errorf("anime.ArchiveMessage = %q, want 凍結中 (preserved)", anime.ArchiveMessage.String)
	}

	result, err := newSyncUsecase(db).Execute(ctx, SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("sync Execute() error = %v", err)
	}
	if result.Unchanged != 1 || result.Updated != 0 {
		t.Fatalf("sync result = %+v, want Unchanged:1 Updated:0", result)
	}
}

// TestUpdateWorkUsecase_Execute_SkipsAnimeWriteForUnmappedWork verifies that an unmapped
// work (anime_id NULL) is updated on the works side only: the usecase does not create an
// anime, leaving that to the sync batch (the arbiter).
//
// [Ja] TestUpdateWorkUsecase_Execute_SkipsAnimeWriteForUnmappedWork は、未マッピングの
// work (anime_id NULL) が works 側だけ更新されることを検証する。UseCase は anime を作らず、
// 同期バッチ (裁定者) に委ねる。
func TestUpdateWorkUsecase_Execute_SkipsAnimeWriteForUnmappedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateWorkUsecase(db)

	// Create a mapped work, then clear its mapping to model a Rails-created work not yet
	// picked up by the sync batch.
	//
	// [Ja] マッピング済み work を作ってからマッピングを外し、同期バッチに未取り込みの
	// Rails 由来 work を模す。
	workID := createMappedWork(t, db, "未マッピング更新前_"+t.Name())
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, "UPDATE works SET anime_id = NULL WHERE id = $1", int64(workID)); err != nil {
		t.Fatalf("anime_id のクリアに失敗: %v", err)
	}

	newTitle := "未マッピング更新後_" + t.Name()
	if _, err := uc.Execute(ctx, validUpdateWorkInput(workID, newTitle)); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// The works row is updated, and anime_id stays NULL (no anime was created here).
	//
	// [Ja] works 行は更新され、anime_id は NULL のまま (ここでは anime を作らない)。
	work := reloadSyncWork(t, db, workID)
	if work.Title != newTitle {
		t.Errorf("work.Title = %q, want %q", work.Title, newTitle)
	}
	if work.AnimeID != nil {
		t.Errorf("work.AnimeID = %v, want nil (unmapped work stays unmapped)", *work.AnimeID)
	}
}

func TestUpdateWorkUsecase_Execute_ReturnsValidationError(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateWorkUsecase(db)

	workID := createMappedWork(t, db, "更新前アニメ_"+t.Name())

	input := validUpdateWorkInput(workID, "更新後アニメ_"+t.Name())
	input.Title = "" // required

	output, err := uc.Execute(context.Background(), input)
	if output != nil {
		t.Errorf("output = %+v, want nil on validation error", output)
	}
	ve := model.AsValidationError(err)
	if ve == nil {
		t.Fatalf("expected *model.ValidationError, got %v", err)
	}
	if len(ve.GetFieldErrors("title")) == 0 {
		t.Error("expected a validation error on the title field")
	}
}

func TestUpdateWorkUsecase_Execute_ReturnsNotFoundForMissingWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateWorkUsecase(db)

	// An id far above any real works.id never matches a row.
	//
	// [Ja] 実在する works.id より遥かに大きい id はどの行にも一致しない。
	output, err := uc.Execute(context.Background(), validUpdateWorkInput(model.WorkID(1<<62), "存在しない_"+t.Name()))
	if output != nil {
		t.Errorf("output = %+v, want nil for a missing work", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("expected AppErrCodeResourceNotFound, got %v", err)
	}
}
