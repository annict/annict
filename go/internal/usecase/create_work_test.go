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

// newCreateWorkUsecase wires the create-work usecase against the shared test DB. The
// usecase opens its own transaction, so its tests use GetTestDB (not SetupTx) so the
// committed rows are visible to the usecase's inner transaction and to the follow-up
// sync invariant check.
//
// [Ja] newCreateWorkUsecase は共有テスト DB に対して作品作成 UseCase を組み立てる。
// 本 UseCase は内部で自前のトランザクションを開くため、テストは SetupTx ではなく
// GetTestDB を使い、コミット済みの行が UseCase の内側トランザクションと後続の同期
// 不変条件チェックから見えるようにする。
func newCreateWorkUsecase(db *sql.DB) *CreateWorkUsecase {
	queries := query.New(db)
	return NewCreateWorkUsecase(
		db,
		repository.NewWorkRepository(queries),
		repository.NewAnimeRepository(queries),
		repository.NewAnimeClassificationRepository(queries),
		newTestWorkSatelliteRepos(queries),
		validator.NewDBWorkCreateValidator(),
	)
}

// newTestWorkSatelliteRepos builds the satellite repository bundle the create / update
// usecases dual-write through.
//
// [Ja] newTestWorkSatelliteRepos は作成 / 更新 UseCase が両書きする別表リポジトリの束を作る。
func newTestWorkSatelliteRepos(queries *query.Queries) WorkSatelliteRepos {
	return WorkSatelliteRepos{
		ExternalID:      repository.NewAnimeExternalIDRepository(queries),
		Link:            repository.NewAnimeLinkRepository(queries),
		OfficialAccount: repository.NewAnimeOfficialAccountRepository(queries),
		Hashtag:         repository.NewAnimeHashtagRepository(queries),
		Season:          repository.NewAnimeSeasonRepository(queries),
		Event:           repository.NewAnimeEventRepository(queries),
	}
}

// newSyncSatellitesUsecase wires the phase 2 satellite sync (SyncWorkSatellitesUsecase)
// with all six reconcilers, so a test can run it right after a create / update and assert
// no diff is detected (the invariant that the dual-write and the sync derive the same
// satellite rows).
//
// [Ja] newSyncSatellitesUsecase はフェーズ 2 の別表同期 (SyncWorkSatellitesUsecase) を 6 つ
// のリコンサイラすべてと組み立てる。作成 / 更新の直後にこれを走らせ差分ゼロを検証できるように
// する (両書きと同期が同じ別表行を導出するという不変条件)。
func newSyncSatellitesUsecase(db *sql.DB) *SyncWorkSatellitesUsecase {
	queries := query.New(db)
	return NewSyncWorkSatellitesUsecase(
		repository.NewWorkRepository(queries),
		NewSyncAnimeExternalIDsUsecase(db, repository.NewAnimeExternalIDRepository(queries)),
		NewSyncAnimeLinksUsecase(db, repository.NewAnimeLinkRepository(queries)),
		NewSyncAnimeOfficialAccountsUsecase(db, repository.NewAnimeOfficialAccountRepository(queries)),
		NewSyncAnimeHashtagsUsecase(db, repository.NewAnimeHashtagRepository(queries)),
		NewSyncAnimeSeasonsUsecase(db, repository.NewAnimeSeasonRepository(queries)),
		NewSyncAnimeEventsUsecase(db, repository.NewAnimeEventRepository(queries)),
	)
}

// validCreateWorkInput returns a form input that passes DBWorkCreateValidator, with
// enough non-default fields set to exercise the work -> anime / classification mapping.
// The title is taken as an argument so each test can pass a unique value (e.g. t.Name()):
// these tests use GetTestDB and commit their rows to the shared DB, so a per-test title
// keeps parallel tests from sharing works rows.
//
// [Ja] validCreateWorkInput は DBWorkCreateValidator を通過するフォーム入力を返す。
// work -> anime / 分類 の写像を検証できるよう、非デフォルトのフィールドを十分にセットする。
// タイトルは引数で受け取り、各テストがユニークな値 (例: t.Name()) を渡せるようにする。
// 本テスト群は GetTestDB を使い行を共有 DB にコミットするため、テストごとのタイトルで
// 並行テストが works 行を共有しないようにする。
func validCreateWorkInput(title string) CreateWorkInput {
	return CreateWorkInput{
		WorkFormInput: WorkFormInput{
			Title:                 title,
			TitleKana:             "さくせいてすとあにめ",
			TitleEn:               "Create Test Anime",
			Media:                 "2", // OVA
			Synopsis:              "あらすじ本文",
			SynopsisSource:        "出典",
			ManualEpisodesCount:   "12",
			StartEpisodeRawNumber: "2.5",
			NoEpisodes:            "1",
			// Satellite source fields so create dual-writes rows into all six satellite
			// tables (external IDs / links / official account / hashtag / season / event).
			//
			// [Ja] create が 6 つの別表 (外部 ID / リンク / 公式アカウント / ハッシュタグ /
			// 季節 / イベント) に行を両書きするよう、別表ソースのフィールドをセットする。
			SeasonYear:      "2024",
			SeasonName:      "2", // spring
			StartedOn:       "2024-04-05",
			EndedOn:         "2024-06-28",
			OfficialSiteURL: "https://example.com/anime",
			WikipediaURL:    "https://ja.wikipedia.org/wiki/anime",
			TwitterUsername: "anime_official",
			TwitterHashtag:  "anime2024",
			ScTid:           "5678",
			MalAnimeID:      "1234",
		},
	}
}

func TestCreateWorkUsecase_Execute_CreatesWorkAnimeAndClassification(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newCreateWorkUsecase(db)

	title := "作成テストアニメ_" + t.Name()
	output, err := uc.Execute(context.Background(), validCreateWorkInput(title))
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if output == nil || output.WorkID == 0 {
		t.Fatalf("output = %+v, want a non-zero WorkID", output)
	}

	// works.anime_id must be written back so the work is mapped to the new anime.
	//
	// [Ja] works.anime_id が書き戻され、作品が新規 anime にマッピングされていること。
	work := reloadSyncWork(t, db, output.WorkID)
	if work.AnimeID == nil {
		t.Fatal("works.anime_id should be written back, got nil")
	}
	animeID := *work.AnimeID

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID() anime=%v err=%v", anime, err)
	}
	if anime.Title.String != title {
		t.Errorf("anime.Title = %q, want %q", anime.Title.String, title)
	}
	if anime.TitleKana.String != "さくせいてすとあにめ" {
		t.Errorf("anime.TitleKana = %q, want さくせいてすとあにめ", anime.TitleKana.String)
	}
	if anime.TitleEn.String != "Create Test Anime" {
		t.Errorf("anime.TitleEn = %q, want Create Test Anime", anime.TitleEn.String)
	}
	if anime.Synopsis.String != "あらすじ本文" {
		t.Errorf("anime.Synopsis = %q, want あらすじ本文", anime.Synopsis.String)
	}
	if anime.Media != model.AnimeMediaOVA {
		t.Errorf("anime.Media = %q, want ova", anime.Media)
	}
	// A newly created work is always published, so its anime mirrors that status.
	//
	// [Ja] 新規作成の作品は常に published のため、その anime も同じステータスを写す。
	if anime.Status != model.AnimeStatusPublished {
		t.Errorf("anime.Status = %q, want published", anime.Status)
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
		t.Error("classification.Standalone = false, want true (no_episodes=1)")
	}
	if classification.EpisodeStartNumber.String != "2.5" {
		t.Errorf("classification.EpisodeStartNumber = %q, want 2.5", classification.EpisodeStartNumber.String)
	}
	if !classification.ExpectedEpisodesCount.Valid || classification.ExpectedEpisodesCount.Int32 != 12 {
		t.Errorf("classification.ExpectedEpisodesCount = %+v, want {12 true}", classification.ExpectedEpisodesCount)
	}
}

// TestCreateWorkUsecase_Execute_ProducesSyncConsistentMapping is the invariant that
// justifies reusing the sync mapping helpers: a sync run right after a create must
// detect no diff (Unchanged), proving create and sync derive the same anime /
// classification from the work and the create path never inflates the diff metric.
//
// [Ja] TestCreateWorkUsecase_Execute_ProducesSyncConsistentMapping は同期の写像ヘルパー
// 再利用を正当化する不変条件。作成直後の同期実行は差分なし (Unchanged) を検出しなければ
// ならず、create と同期が同じ anime / 分類を work から導出していること、create 経路が
// 差分メトリクスを水増ししないことを示す。
func TestCreateWorkUsecase_Execute_ProducesSyncConsistentMapping(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newCreateWorkUsecase(db)

	output, err := uc.Execute(context.Background(), validCreateWorkInput("作成テストアニメ_"+t.Name()))
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	syncUC := newSyncUsecase(db)
	result, err := syncUC.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{output.WorkID}})
	if err != nil {
		t.Fatalf("sync Execute() error = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.Updated != 0 || result.Unchanged != 1 {
		t.Fatalf("sync result = %+v, want {Processed:1 Created:0 Updated:0 Unchanged:1}", result)
	}
}

// TestCreateWorkUsecase_Execute_WritesSatelliteRows verifies the create dual-writes the
// six satellite tables from the work's source columns: the external IDs, links, official
// account, hashtag, season and broadcast event derived from the form input all land on the
// new anime.
//
// [Ja] TestCreateWorkUsecase_Execute_WritesSatelliteRows は、create が work のソース列から
// 6 つの別表を両書きすることを検証する。フォーム入力から導出した外部 ID・リンク・公式
// アカウント・ハッシュタグ・季節・放送イベントがすべて新規 anime に載る。
func TestCreateWorkUsecase_Execute_WritesSatelliteRows(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newCreateWorkUsecase(db)
	ctx := context.Background()

	output, err := uc.Execute(ctx, validCreateWorkInput("別表作成テストアニメ_"+t.Name()))
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	animeID := *reloadSyncWork(t, db, output.WorkID).AnimeID
	queries := query.New(db)
	animeIDs := []model.AnimeID{animeID}

	externalIDs, err := repository.NewAnimeExternalIDRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(external_ids) error = %v", err)
	}
	gotExternalIDs := map[model.AnimeExternalService]string{}
	for _, e := range externalIDs {
		gotExternalIDs[e.Service] = e.ExternalID
	}
	if gotExternalIDs[model.AnimeExternalServiceSyobocal] != "5678" {
		t.Errorf("syobocal external_id = %q, want 5678", gotExternalIDs[model.AnimeExternalServiceSyobocal])
	}
	if gotExternalIDs[model.AnimeExternalServiceMal] != "1234" {
		t.Errorf("mal external_id = %q, want 1234", gotExternalIDs[model.AnimeExternalServiceMal])
	}

	links, err := repository.NewAnimeLinkRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(links) error = %v", err)
	}
	gotLinks := map[animeLinkKey]string{}
	for _, l := range links {
		gotLinks[animeLinkKey{animeID: l.AnimeID, kind: l.Kind, language: l.Language}] = l.URL
	}
	officialSiteJa := gotLinks[animeLinkKey{animeID: animeID, kind: model.AnimeLinkKindOfficialSite, language: model.LanguageJa}]
	if officialSiteJa != "https://example.com/anime" {
		t.Errorf("official_site/ja url = %q, want https://example.com/anime", officialSiteJa)
	}
	wikipediaJa := gotLinks[animeLinkKey{animeID: animeID, kind: model.AnimeLinkKindWikipedia, language: model.LanguageJa}]
	if wikipediaJa != "https://ja.wikipedia.org/wiki/anime" {
		t.Errorf("wikipedia/ja url = %q, want https://ja.wikipedia.org/wiki/anime", wikipediaJa)
	}

	accounts, err := repository.NewAnimeOfficialAccountRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(official_accounts) error = %v", err)
	}
	if len(accounts) != 1 || accounts[0].Service != model.AnimeAccountServiceX || accounts[0].Account != "anime_official" {
		t.Errorf("official accounts = %+v, want one x=anime_official", accounts)
	}

	hashtags, err := repository.NewAnimeHashtagRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(hashtags) error = %v", err)
	}
	if len(hashtags) != 1 || hashtags[0].Hashtag != "anime2024" {
		t.Errorf("hashtags = %+v, want one anime2024", hashtags)
	}

	seasons, err := repository.NewAnimeSeasonRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(seasons) error = %v", err)
	}
	if len(seasons) != 1 {
		t.Fatalf("seasons = %+v, want one row", seasons)
	}
	if seasons[0].Year != 2024 || seasons[0].Name == nil || *seasons[0].Name != model.SeasonNameSpring || !seasons[0].IsPrimary {
		t.Errorf("season = %+v, want year 2024 name spring is_primary true", seasons[0])
	}

	events, err := repository.NewAnimeEventRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(events) error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("events = %+v, want one row", events)
	}
	if events[0].Kind != model.AnimeEventKindBroadcast || events[0].StartedOn.Format("2006-01-02") != "2024-04-05" ||
		events[0].EndedOn == nil || events[0].EndedOn.Format("2006-01-02") != "2024-06-28" {
		t.Errorf("event = %+v, want broadcast 2024-04-05..2024-06-28", events[0])
	}
}

// TestCreateWorkUsecase_Execute_ProducesSyncConsistentSatellites extends the create
// invariant to the satellite tables: the phase 2 satellite sync run right after a create
// must detect no diff, proving the create dual-write and the sync derive the same
// satellite rows from the work.
//
// [Ja] TestCreateWorkUsecase_Execute_ProducesSyncConsistentSatellites は create の不変条件を
// 別表に広げる。作成直後のフェーズ 2 別表同期は差分を検出してはならず、create の両書きと同期が
// 同じ別表行を work から導出していることを示す。
func TestCreateWorkUsecase_Execute_ProducesSyncConsistentSatellites(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newCreateWorkUsecase(db)

	output, err := uc.Execute(context.Background(), validCreateWorkInput("別表整合作成_"+t.Name()))
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result, err := newSyncSatellitesUsecase(db).Execute(context.Background(), SyncWorkSatellitesInput{WorkIDs: []model.WorkID{output.WorkID}})
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

func TestCreateWorkUsecase_Execute_ReturnsValidationError(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newCreateWorkUsecase(db)

	input := validCreateWorkInput("作成テストアニメ_" + t.Name())
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
