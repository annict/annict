package usecase

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

// newCreateWorkUsecaseは共有テストDBに対して作品作成UseCaseを組み立てる。
// 本UseCaseは内部で自前のトランザクションを開くため、テストはSetupTxではなく
// GetTestDBを使い、コミット済みの行がUseCaseの内側トランザクションと後続の同期
// 不変条件チェックから見えるようにする。
func newCreateWorkUsecase(db *sql.DB) *CreateWorkUsecase {
	queries := query.New(db)
	workRepo := repository.NewWorkRepository(queries)
	numberFormatRepo := repository.NewNumberFormatRepository(queries)
	return NewCreateWorkUsecase(
		db,
		workRepo,
		repository.NewAnimeRepository(queries),
		repository.NewAnimeClassificationRepository(queries),
		newTestWorkSatelliteRepos(queries),
		validator.NewDBWorkCreateValidator(workRepo, numberFormatRepo),
	)
}

// newTestWorkSatelliteReposは作成 / 更新UseCaseが両書きする別表リポジトリの束を作る。
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

// newSyncSatellitesUsecaseはフェーズ2の別表同期 (SyncWorkSatellitesUsecase) を6つ
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

// validCreateWorkInputはDBWorkCreateValidatorを通過するフォーム入力を返す。
// work -> anime / 分類 の写像を検証できるよう、非デフォルトのフィールドを十分にセットする。
// タイトルは引数で受け取り、各テストがユニークな値 (例: t.Name()) を渡せるようにする。
// 本テスト群はGetTestDBを使い行を共有DBにコミットするため、テストごとのタイトルで
// 並行テストがworks行を共有しないようにする。
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
			// createが6つの別表 (外部ID / リンク / 公式アカウント / ハッシュタグ /
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
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if output == nil || output.WorkID == 0 {
		t.Fatalf("output = %+v、期待値 = WorkIDが0以外", output)
	}

	// works.anime_idが書き戻され、作品が新規animeにマッピングされていること。
	work := reloadSyncWork(t, db, output.WorkID)
	if work.AnimeID == nil {
		t.Fatal("works.anime_id = nil、期待値 = 書き戻された値")
	}
	animeID := *work.AnimeID

	animeRepo := repository.NewAnimeRepository(query.New(db))
	anime, err := animeRepo.GetByID(context.Background(), animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Title.String != title {
		t.Errorf("anime.Title = %q、期待値 = %q", anime.Title.String, title)
	}
	if anime.TitleKana.String != "さくせいてすとあにめ" {
		t.Errorf("anime.TitleKana = %q、期待値 = さくせいてすとあにめ", anime.TitleKana.String)
	}
	if anime.TitleEn.String != "Create Test Anime" {
		t.Errorf("anime.TitleEn = %q、期待値 = Create Test Anime", anime.TitleEn.String)
	}
	if anime.Synopsis.String != "あらすじ本文" {
		t.Errorf("anime.Synopsis = %q、期待値 = あらすじ本文", anime.Synopsis.String)
	}
	if anime.Media != model.AnimeMediaOVA {
		t.Errorf("anime.Media = %q、期待値 = ova", anime.Media)
	}
	// 新規作成の作品は常にpublishedのため、そのanimeも同じステータスを写す。
	if anime.Status != model.AnimeStatusPublished {
		t.Errorf("anime.Status = %q、期待値 = published", anime.Status)
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
		t.Error("classification.Standalone = false、期待値 = true (no_episodes=1)")
	}
	if classification.EpisodeStartNumber.String != "2.5" {
		t.Errorf("classification.EpisodeStartNumber = %q、期待値 = 2.5", classification.EpisodeStartNumber.String)
	}
	if !classification.ExpectedEpisodesCount.Valid || classification.ExpectedEpisodesCount.Int32 != 12 {
		t.Errorf("classification.ExpectedEpisodesCount = %+v、期待値 = {12 true}", classification.ExpectedEpisodesCount)
	}
}

// TestCreateWorkUsecase_Execute_ProducesSyncConsistentMappingは同期の写像ヘルパー
// 再利用を正当化する不変条件。作成直後の同期実行は差分なし (Unchanged) を検出しなければ
// ならず、createと同期が同じanime / 分類をworkから導出していること、create経路が
// 差分メトリクスを水増ししないことを示す。
func TestCreateWorkUsecase_Execute_ProducesSyncConsistentMapping(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newCreateWorkUsecase(db)

	output, err := uc.Execute(context.Background(), validCreateWorkInput("作成テストアニメ_"+t.Name()))
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	syncUC := newSyncUsecase(db)
	result, err := syncUC.Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{output.WorkID}})
	if err != nil {
		t.Fatalf("同期のExecute()のエラー = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.Updated != 0 || result.Unchanged != 1 {
		t.Fatalf("同期の結果 = %+v、期待値 = {Processed:1 Created:0 Updated:0 Unchanged:1}", result)
	}
}

// TestCreateWorkUsecase_Execute_WritesSatelliteRowsは、createがworkのソース列から
// 6つの別表を両書きすることを検証する。フォーム入力から導出した外部ID・リンク・公式
// アカウント・ハッシュタグ・季節・放送イベントがすべて新規animeに載る。
func TestCreateWorkUsecase_Execute_WritesSatelliteRows(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newCreateWorkUsecase(db)
	ctx := context.Background()

	output, err := uc.Execute(ctx, validCreateWorkInput("別表作成テストアニメ_"+t.Name()))
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	animeID := *reloadSyncWork(t, db, output.WorkID).AnimeID
	queries := query.New(db)
	animeIDs := []model.AnimeID{animeID}

	externalIDs, err := repository.NewAnimeExternalIDRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(external_ids)のエラー = %v", err)
	}
	gotExternalIDs := map[model.AnimeExternalService]string{}
	for _, e := range externalIDs {
		gotExternalIDs[e.Service] = e.ExternalID
	}
	if gotExternalIDs[model.AnimeExternalServiceSyobocal] != "5678" {
		t.Errorf("syobocal external_id = %q、期待値 = 5678", gotExternalIDs[model.AnimeExternalServiceSyobocal])
	}
	if gotExternalIDs[model.AnimeExternalServiceMal] != "1234" {
		t.Errorf("mal external_id = %q、期待値 = 1234", gotExternalIDs[model.AnimeExternalServiceMal])
	}

	links, err := repository.NewAnimeLinkRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(links)のエラー = %v", err)
	}
	gotLinks := map[animeLinkKey]string{}
	for _, l := range links {
		gotLinks[animeLinkKey{animeID: l.AnimeID, kind: l.Kind, language: l.Language}] = l.URL
	}
	officialSiteJa := gotLinks[animeLinkKey{animeID: animeID, kind: model.AnimeLinkKindOfficialSite, language: model.LanguageJa}]
	if officialSiteJa != "https://example.com/anime" {
		t.Errorf("official_site/ja url = %q、期待値 = https://example.com/anime", officialSiteJa)
	}
	wikipediaJa := gotLinks[animeLinkKey{animeID: animeID, kind: model.AnimeLinkKindWikipedia, language: model.LanguageJa}]
	if wikipediaJa != "https://ja.wikipedia.org/wiki/anime" {
		t.Errorf("wikipedia/ja url = %q、期待値 = https://ja.wikipedia.org/wiki/anime", wikipediaJa)
	}

	accounts, err := repository.NewAnimeOfficialAccountRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(official_accounts)のエラー = %v", err)
	}
	if len(accounts) != 1 || accounts[0].Service != model.AnimeAccountServiceX || accounts[0].Account != "anime_official" {
		t.Errorf("accounts = %+v、期待値 = x=anime_officialが1件", accounts)
	}

	hashtags, err := repository.NewAnimeHashtagRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(hashtags)のエラー = %v", err)
	}
	if len(hashtags) != 1 || hashtags[0].Hashtag != "anime2024" {
		t.Errorf("hashtags = %+v、期待値 = anime2024が1件", hashtags)
	}

	seasons, err := repository.NewAnimeSeasonRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(seasons)のエラー = %v", err)
	}
	if len(seasons) != 1 {
		t.Fatalf("seasons = %+v、期待値 = 1件", seasons)
	}
	if seasons[0].Year != 2024 || seasons[0].Name == nil || *seasons[0].Name != model.SeasonNameSpring || !seasons[0].IsPrimary {
		t.Errorf("season = %+v、期待値 = year 2024 name spring is_primary true", seasons[0])
	}

	events, err := repository.NewAnimeEventRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(events)のエラー = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("events = %+v、期待値 = 1件", events)
	}
	if events[0].Kind != model.AnimeEventKindBroadcast || events[0].StartedOn.Format("2006-01-02") != "2024-04-05" ||
		events[0].EndedOn == nil || events[0].EndedOn.Format("2006-01-02") != "2024-06-28" {
		t.Errorf("event = %+v、期待値 = broadcast 2024-04-05..2024-06-28", events[0])
	}
}

// TestCreateWorkUsecase_Execute_ProducesSyncConsistentSatellitesはcreateの不変条件を
// 別表に広げる。作成直後のフェーズ2別表同期は差分を検出してはならず、createの両書きと同期が
// 同じ別表行をworkから導出していることを示す。
func TestCreateWorkUsecase_Execute_ProducesSyncConsistentSatellites(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newCreateWorkUsecase(db)

	output, err := uc.Execute(context.Background(), validCreateWorkInput("別表整合作成_"+t.Name()))
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	result, err := newSyncSatellitesUsecase(db).Execute(context.Background(), SyncWorkSatellitesInput{WorkIDs: []model.WorkID{output.WorkID}})
	if err != nil {
		t.Fatalf("サテライトの同期のExecute()のエラー = %v", err)
	}
	if result.Processed != 1 || result.SkippedNoAnime != 0 {
		t.Fatalf("サテライトの同期の結果 = %+v、期待値 = Processed:1 SkippedNoAnime:0", result)
	}
	if result.Created != 0 || result.Updated != 0 || result.Deleted != 0 {
		t.Fatalf("サテライトの同期結果 = %+v、期待値 = Created:0 Updated:0 Deleted:0", result)
	}
}

func TestCreateWorkUsecase_Execute_ReturnsValidationError(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newCreateWorkUsecase(db)

	input := validCreateWorkInput("作成テストアニメ_" + t.Name())
	input.Title = "" // 必須項目

	output, err := uc.Execute(context.Background(), input)
	if output != nil {
		t.Errorf("output = %+v、期待値 = nil (バリデーションエラーのため)", output)
	}
	ve := model.AsValidationError(err)
	if ve == nil {
		t.Fatalf("エラーの型 = %v、期待値 = *model.ValidationError", err)
	}
	if len(ve.GetFieldErrors("title")) == 0 {
		t.Error("titleフィールドのバリデーションエラーを期待したが、無かった")
	}
}

// TestCreateWorkUsecase_Execute_RejectsDuplicateTitleはUseCaseがバリデーターから
// 引き継ぐタイトル一意性を対象とする。RailsのWorkモデルに対応する規則で、同じDBに
// 書き込む両実装が「どのタイトルが空いているか」で食い違わないようにする。
func TestCreateWorkUsecase_Execute_RejectsDuplicateTitle(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newCreateWorkUsecase(db)

	title := "作成テストアニメ_" + t.Name()
	if _, err := uc.Execute(context.Background(), validCreateWorkInput(title)); err != nil {
		t.Fatalf("1件目のExecute()のエラー = %v", err)
	}

	output, err := uc.Execute(context.Background(), validCreateWorkInput(title))
	if output != nil {
		t.Errorf("output = %+v、期待値 = nil (バリデーションエラーのため)", output)
	}
	ve := model.AsValidationError(err)
	if ve == nil {
		t.Fatalf("エラーの型 = %v、期待値 = *model.ValidationError", err)
	}
	if len(ve.GetFieldErrors("title")) == 0 {
		t.Error("titleフィールドのバリデーションエラーを期待したが、無かった")
	}
}

// TestCreateWorkUsecase_Execute_RejectsUnstorableValuesは、以前は受け付けたうえで
// 失われていた値を対象とする。長すぎるタイトルはINSERTで失敗して500になり、数値として
// 不正な値はカラムへ渡る途中で捨てられ、そのフィールドが黙って空のまま作品が保存されていた。
// どちらも今はフィールドを名指ししたバリデーションエラーとして返る。
func TestCreateWorkUsecase_Execute_RejectsUnstorableValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		mutate    func(input *CreateWorkInput)
		wantField string
	}{
		{
			name: "タイトルがカラムの上限を超える",
			mutate: func(input *CreateWorkInput) {
				input.Title = strings.Repeat("あ", 501)
			},
			wantField: "title",
		},
		{
			name: "話数が数値でない",
			mutate: func(input *CreateWorkInput) {
				input.ManualEpisodesCount = "十二"
			},
			wantField: "manual_episodes_count",
		},
		{
			name: "開始話数が数値でない",
			mutate: func(input *CreateWorkInput) {
				input.StartEpisodeRawNumber = "第1話"
			},
			wantField: "start_episode_raw_number",
		},
		{
			name: "開始日が日付でない",
			mutate: func(input *CreateWorkInput) {
				input.StartedOn = "2024/04/01"
			},
			wantField: "started_on",
		},
		{
			// works.number_format_idとanime_classifications.number_format_idは
			// number_formatsへの外部キーで、どの行も持たないidは以前はトランザクション内の
			// INSERTで失敗していた。シーケンスの採番範囲を超えたidは存在し得ない。
			name: "話数フォーマットが登録されていない",
			mutate: func(input *CreateWorkInput) {
				input.NumberFormatID = "9000000000000000000"
			},
			wantField: "number_format_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := testutil.GetTestDB()
			uc := newCreateWorkUsecase(db)

			input := validCreateWorkInput("作成テストアニメ_" + t.Name())
			tt.mutate(&input)

			output, err := uc.Execute(context.Background(), input)
			if output != nil {
				t.Errorf("output = %+v、期待値 = nil (バリデーションエラーのため)", output)
			}
			ve := model.AsValidationError(err)
			if ve == nil {
				t.Fatalf("エラーの型 = %v、期待値 = *model.ValidationError", err)
			}
			if len(ve.GetFieldErrors(tt.wantField)) == 0 {
				t.Errorf("%sフィールドのエラー = %+v、期待値 = バリデーションエラー", tt.wantField, ve)
			}
		})
	}
}
