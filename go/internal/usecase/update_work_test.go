package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

// newUpdateWorkUsecaseは共有テストDBに対して作品更新UseCaseを組み立てる。
// 作成UseCaseと同じく内部で自前のトランザクションを開くため、テストはSetupTxではなく
// GetTestDBを使い、コミット済みの行がUseCaseの内側トランザクションと後続の同期不変条件
// チェックから見えるようにする。
func newUpdateWorkUsecase(db *sql.DB) *UpdateWorkUsecase {
	queries := query.New(db)
	workRepo := repository.NewWorkRepository(queries)
	numberFormatRepo := repository.NewNumberFormatRepository(queries)
	return NewUpdateWorkUsecase(
		db,
		workRepo,
		repository.NewAnimeRepository(queries),
		repository.NewAnimeClassificationRepository(queries),
		newTestWorkSatelliteRepos(queries),
		validator.NewDBWorkCreateValidator(workRepo, numberFormatRepo),
	)
}

// readWorkVersionは作品の編集フォームが運ぶ版を返す。送信が受理されるには、この版を名乗る
// 必要がある。存在しない作品には読むべき版が無いため、nullのセンチネルで代用する。そのような
// 作品への送信もバリデーションを通り、版の欠落ではなく作品の不在で却下されるようにするため。
func readWorkVersion(t *testing.T, db *sql.DB, workID model.WorkID) string {
	t.Helper()

	var updatedAt sql.NullTime
	err := db.QueryRow(`SELECT updated_at FROM works WHERE id = $1`, int64(workID)).Scan(&updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return validator.FormNullVersion
	}
	if err != nil {
		t.Fatalf("作品の版の読み込みに失敗: %v", err)
	}
	if !updatedAt.Valid {
		return validator.FormNullVersion
	}

	return updatedAt.Time.UTC().Format(validator.FormVersionLayout)
}

// validUpdateWorkInputはDBWorkCreateValidatorを通過し、validCreateWorkInputとは
// 異なる値のフォーム入力を返す。更新が反映されたことをテストで観測できるようにする
// (メディアがOVAでなくTV、standaloneをoffに反転、話数も別の値)。版は作品が現在持つもの
// を運ぶ。編集者のフォームが開かれていたであろう版である。
func validUpdateWorkInput(t *testing.T, db *sql.DB, workID model.WorkID, title string) UpdateWorkInput {
	t.Helper()

	return UpdateWorkInput{
		WorkID:    workID,
		UpdatedAt: readWorkVersion(t, db, workID),
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
			// 別表ソースのフィールドをvalidCreateWorkInputと変えて全リコンサイル経路を
			// 動かす: 値の変更はその場更新 (公式サイト・xアカウント・syobocal id・放送)、自然
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

// createMappedWorkは作成UseCase経由でworkを作成し (新規animeにマッピング
// される)、そのIDを返す。更新対象の既存マッピング済みworkが要るテストで使う。
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
	if _, err := uc.Execute(context.Background(), validUpdateWorkInput(t, db, workID, newTitle)); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	// works行 (正本) が送信された変更を反映している。
	work := reloadSyncWork(t, db, workID)
	if work.Title != newTitle {
		t.Errorf("work.Title = %q、期待値 = %q", work.Title, newTitle)
	}
	if work.Media != workMediaTV {
		t.Errorf("work.Media = %d、期待値 = %d (tv)", work.Media, workMediaTV)
	}
	if work.NoEpisodes {
		t.Error("work.NoEpisodes = true、期待値 = false")
	}
	if work.AnimeID == nil {
		t.Fatal("work.AnimeID = nil、期待値 = 対応付けが維持されること")
	}
	animeID := *work.AnimeID

	// マッピング済みanimeが更新後の内容を写し、statusはpublishedのまま
	// (編集フォームはstatusを触らない)。
	anime, err := repository.NewAnimeRepository(query.New(db)).GetByID(context.Background(), animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.Title.String != newTitle {
		t.Errorf("anime.Title = %q、期待値 = %q", anime.Title.String, newTitle)
	}
	if anime.Media != model.AnimeMediaTV {
		t.Errorf("anime.Media = %q、期待値 = tv", anime.Media)
	}
	if anime.Synopsis.String != "更新後のあらすじ本文" {
		t.Errorf("anime.Synopsis = %q、期待値 = 更新後のあらすじ本文", anime.Synopsis.String)
	}
	if anime.Status != model.AnimeStatusPublished {
		t.Errorf("anime.Status = %q、期待値 = published", anime.Status)
	}

	classification, err := repository.NewAnimeClassificationRepository(query.New(db)).GetByAnimeID(context.Background(), animeID)
	if err != nil || classification == nil {
		t.Fatalf("GetByAnimeID()のclassification = %v、エラー = %v", classification, err)
	}
	if classification.Standalone {
		t.Error("classification.Standalone = true、期待値 = false (no_episodesのチェックを外したため)")
	}
	if classification.EpisodeStartNumber.String != "1" {
		t.Errorf("classification.EpisodeStartNumber = %q、期待値 = 1", classification.EpisodeStartNumber.String)
	}
	if !classification.ExpectedEpisodesCount.Valid || classification.ExpectedEpisodesCount.Int32 != 24 {
		t.Errorf("classification.ExpectedEpisodesCount = %+v、期待値 = {24 true}", classification.ExpectedEpisodesCount)
	}
}

// TestUpdateWorkUsecase_Execute_ProducesSyncConsistentMappingは作成側の不変条件の
// 更新版。更新直後の同期実行は差分なし (Unchanged) を検出しなければならず、更新が、同期が
// 更新後のworks行から導出するのと同じanime / 分類を書いていることを示す。
func TestUpdateWorkUsecase_Execute_ProducesSyncConsistentMapping(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateWorkUsecase(db)

	workID := createMappedWork(t, db, "更新前アニメ_"+t.Name())
	if _, err := uc.Execute(context.Background(), validUpdateWorkInput(t, db, workID, "更新後アニメ_"+t.Name())); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	result, err := newSyncUsecase(db).Execute(context.Background(), SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("同期のExecute()のエラー = %v", err)
	}
	if result.Processed != 1 || result.Created != 0 || result.Updated != 0 || result.Unchanged != 1 {
		t.Fatalf("同期の結果 = %+v、期待値 = {Processed:1 Created:0 Updated:0 Unchanged:1}", result)
	}
}

// TestUpdateWorkUsecase_Execute_ReconcilesSatelliteRowsは、更新の両書きが送信値に対して
// 各別表をリコンサイルすることを検証する。値の変更はその場更新 (公式サイト / xアカウント /
// syobocal id / 放送)、自然キーの変更は行の置換 (季節)、値のクリアは行削除 (wikipedia /
// ハッシュタグ / mal id / 放送終了日)。
func TestUpdateWorkUsecase_Execute_ReconcilesSatelliteRows(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateWorkUsecase(db)
	ctx := context.Background()

	workID := createMappedWork(t, db, "別表更新前_"+t.Name())
	if _, err := uc.Execute(ctx, validUpdateWorkInput(t, db, workID, "別表更新後_"+t.Name())); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	animeID := *reloadSyncWork(t, db, workID).AnimeID
	queries := query.New(db)
	animeIDs := []model.AnimeID{animeID}

	// syobocalはその場更新、malは削除 (ソースがクリアされた)。
	externalIDs, err := repository.NewAnimeExternalIDRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(external_ids)のエラー = %v", err)
	}
	if len(externalIDs) != 1 || externalIDs[0].Service != model.AnimeExternalServiceSyobocal || externalIDs[0].ExternalID != "9999" {
		t.Errorf("external_ids = %+v、期待値 = syobocal=9999のみ", externalIDs)
	}

	// official_siteはその場更新、wikipediaは削除 (ソースがクリアされた)。
	links, err := repository.NewAnimeLinkRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(links)のエラー = %v", err)
	}
	if len(links) != 1 || links[0].Kind != model.AnimeLinkKindOfficialSite || links[0].Language != model.LanguageJa || links[0].URL != "https://example.com/anime-v2" {
		t.Errorf("links = %+v、期待値 = official_site/ja=https://example.com/anime-v2のみ", links)
	}

	accounts, err := repository.NewAnimeOfficialAccountRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(official_accounts)のエラー = %v", err)
	}
	if len(accounts) != 1 || accounts[0].Account != "anime_official_2" {
		t.Errorf("accounts = %+v、期待値 = x=anime_official_2のみ", accounts)
	}

	hashtags, err := repository.NewAnimeHashtagRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(hashtags)のエラー = %v", err)
	}
	if len(hashtags) != 0 {
		t.Errorf("hashtags = %+v、期待値 = 0件 (同期元が空のため)", hashtags)
	}

	// 季節キーが変わった (2024春 -> 2025夏) ため旧行を削除し新行を作成、is_primary行は
	// 1つのまま。
	seasons, err := repository.NewAnimeSeasonRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(seasons)のエラー = %v", err)
	}
	if len(seasons) != 1 || seasons[0].Year != 2025 || seasons[0].Name == nil || *seasons[0].Name != model.SeasonNameSummer || !seasons[0].IsPrimary {
		t.Errorf("seasons = %+v、期待値 = 2025 summerのみis_primary", seasons)
	}

	// 放送イベントはその場更新: 開始日が変わり終了日は未定 (nil) にクリアされた。
	events, err := repository.NewAnimeEventRepository(queries).ListByAnimeIDs(ctx, animeIDs)
	if err != nil {
		t.Fatalf("ListByAnimeIDs(events)のエラー = %v", err)
	}
	if len(events) != 1 || events[0].StartedOn.Format("2006-01-02") != "2025-07-01" || events[0].EndedOn != nil {
		t.Errorf("events = %+v、期待値 = started_on 2025-07-01 / ended_onがnilのbroadcastが1件", events)
	}
}

// TestUpdateWorkUsecase_Execute_ProducesSyncConsistentSatellitesはupdateの不変条件を
// 別表に広げる。更新直後のフェーズ2別表同期は差分を検出してはならず、更新のリコンサイルが
// animeの別表行を、同期が更新後のworks行から導出するのと等しく保つことを示す。
func TestUpdateWorkUsecase_Execute_ProducesSyncConsistentSatellites(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateWorkUsecase(db)

	workID := createMappedWork(t, db, "別表整合更新前_"+t.Name())
	if _, err := uc.Execute(context.Background(), validUpdateWorkInput(t, db, workID, "別表整合更新後_"+t.Name())); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	result, err := newSyncSatellitesUsecase(db).Execute(context.Background(), SyncWorkSatellitesInput{WorkIDs: []model.WorkID{workID}})
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

// TestUpdateWorkUsecase_Execute_PreservesNonFormAnimeColumnsは、更新が編集フォーム
// の送信しないanime写像カラムを潰さないことを検証する: title_ro、archive_message、
// および作品状態のsource (unpublished_at / deleted_at) から導出されるanime.status。
// これらはworks行から引き継がれるため、ローマ字タイトルを持つアーカイブ済みwork
// (unpublished_atあり) は内容編集後もアーカイブ状態とメッセージを保ち、後続の同期も
// Unchangedのままになる。
func TestUpdateWorkUsecase_Execute_PreservesNonFormAnimeColumns(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateWorkUsecase(db)

	workID := createMappedWork(t, db, "更新前アニメ_"+t.Name())
	work := reloadSyncWork(t, db, workID)
	animeID := *work.AnimeID

	// unpublished_atを立てたアーカイブ済みworkにRailsが付けたローマ字タイトルを
	// 持たせ、worksとマッピング済みanimeを (同期が残すのと同じく) 整合させておく: worksは
	// unpublished_atを持ち、animeは導出されたstatus = archivedとanimes専用の
	// archive_messageを持つ。
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, "UPDATE works SET title_ro = $2, unpublished_at = NOW() WHERE id = $1", int64(workID), "Koushin Anime"); err != nil {
		t.Fatalf("worksの前提更新に失敗: %v", err)
	}
	if _, err := db.ExecContext(ctx, "UPDATE animes SET title_ro = $2, status = 'archived', archive_message = $3 WHERE id = $1", int64(animeID), "Koushin Anime", "凍結中"); err != nil {
		t.Fatalf("animesの前提更新に失敗: %v", err)
	}

	if _, err := uc.Execute(ctx, validUpdateWorkInput(t, db, workID, "更新後アニメ_"+t.Name())); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	anime, err := repository.NewAnimeRepository(query.New(db)).GetByID(ctx, animeID)
	if err != nil || anime == nil {
		t.Fatalf("GetByID()のanime = %v、エラー = %v", anime, err)
	}
	if anime.TitleRo.String != "Koushin Anime" {
		t.Errorf("anime.TitleRo = %q、期待値 = Koushin Anime (保持されること)", anime.TitleRo.String)
	}
	if anime.Status != model.AnimeStatusArchived {
		t.Errorf("anime.Status = %q、期待値 = archived (保持されること)", anime.Status)
	}
	if anime.ArchiveMessage.String != "凍結中" {
		t.Errorf("anime.ArchiveMessage = %q、期待値 = 凍結中 (保持されること)", anime.ArchiveMessage.String)
	}

	result, err := newSyncUsecase(db).Execute(ctx, SyncWorksToAnimesInput{WorkIDs: []model.WorkID{workID}})
	if err != nil {
		t.Fatalf("同期のExecute()のエラー = %v", err)
	}
	if result.Unchanged != 1 || result.Updated != 0 {
		t.Fatalf("同期の結果 = %+v、期待値 = Unchanged:1 Updated:0", result)
	}
}

// TestUpdateWorkUsecase_Execute_SkipsAnimeWriteForUnmappedWorkは、未マッピングの
// work (anime_id NULL) がworks側だけ更新されることを検証する。UseCaseはanimeを作らず、
// 同期バッチ (裁定者) に委ねる。
func TestUpdateWorkUsecase_Execute_SkipsAnimeWriteForUnmappedWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateWorkUsecase(db)

	// マッピング済みworkを作ってからマッピングを外し、同期バッチに未取り込みの
	// Rails由来workを模す。
	workID := createMappedWork(t, db, "未マッピング更新前_"+t.Name())
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, "UPDATE works SET anime_id = NULL WHERE id = $1", int64(workID)); err != nil {
		t.Fatalf("anime_idのクリアに失敗: %v", err)
	}

	newTitle := "未マッピング更新後_" + t.Name()
	if _, err := uc.Execute(ctx, validUpdateWorkInput(t, db, workID, newTitle)); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	// works行は更新され、anime_idはNULLのまま (ここではanimeを作らない)。
	work := reloadSyncWork(t, db, workID)
	if work.Title != newTitle {
		t.Errorf("work.Title = %q、期待値 = %q", work.Title, newTitle)
	}
	if work.AnimeID != nil {
		t.Errorf("work.AnimeID = %v、期待値 = nil (対応付けの無い作品はそのままであること)", *work.AnimeID)
	}
}

func TestUpdateWorkUsecase_Execute_ReturnsValidationError(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateWorkUsecase(db)

	workID := createMappedWork(t, db, "更新前アニメ_"+t.Name())

	input := validUpdateWorkInput(t, db, workID, "更新後アニメ_"+t.Name())
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

func TestUpdateWorkUsecase_Execute_ReturnsNotFoundForMissingWork(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateWorkUsecase(db)

	// 実在するworks.idより遥かに大きいidはどの行にも一致しない。
	output, err := uc.Execute(context.Background(), validUpdateWorkInput(t, db, model.WorkID(1<<62), "存在しない_"+t.Name()))
	if output != nil {
		t.Errorf("output = %+v、期待値 = nil (存在しない作品のため)", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
		t.Fatalf("エラーコード = %v、期待値 = AppErrCodeResourceNotFound", err)
	}
}

// TestUpdateWorkUsecase_Execute_RejectsStaleVersionは2人の編集者が同じフォームから送信
// する場合を検証する。2件目は行が既に進んだ版を名乗るため、1人目の値を上書きせず競合として
// 却下される。
func TestUpdateWorkUsecase_Execute_RejectsStaleVersion(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateWorkUsecase(db)

	workID := createMappedWork(t, db, "版競合前アニメ_"+t.Name())

	// 2人の編集者は同じ版でフォームを開き、1件目の送信がその版を進める。
	shared := validUpdateWorkInput(t, db, workID, "先に保存したタイトル_"+t.Name())
	if _, err := uc.Execute(context.Background(), shared); err != nil {
		t.Fatalf("1件目のExecute()のエラー = %v", err)
	}

	second := shared
	second.Title = "後から届いたタイトル_" + t.Name()
	output, err := uc.Execute(context.Background(), second)
	if output != nil {
		t.Errorf("output = %+v、期待値 = nil (versionの競合のため)", output)
	}
	ae := model.AsAppError(err)
	if ae == nil || ae.Code != model.AppErrCodeConflict {
		t.Fatalf("エラーコード = %v、期待値 = AppErrCodeConflict", err)
	}

	if work := reloadSyncWork(t, db, workID); work.Title != shared.Title {
		t.Errorf("work.Title = %q、期待値 = %q (後の送信は上書きしない)", work.Title, shared.Title)
	}
}

// TestUpdateWorkUsecase_Execute_RejectsMissingVersionは版をまったく示さない送信を検証する。
// 何も書かれる前に却下されるため、hiddenフィールド無しで組み立てられた要求が、その時点の行の
// 内容に対して適用されることはない。
func TestUpdateWorkUsecase_Execute_RejectsMissingVersion(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateWorkUsecase(db)

	title := "版なし送信アニメ_" + t.Name()
	workID := createMappedWork(t, db, title)

	input := validUpdateWorkInput(t, db, workID, "版なし更新後_"+t.Name())
	input.UpdatedAt = ""
	output, err := uc.Execute(context.Background(), input)
	if output != nil {
		t.Errorf("output = %+v、期待値 = nil (versionが無いため)", output)
	}
	ve := model.AsValidationError(err)
	if ve == nil {
		t.Fatalf("エラーの型 = %v、期待値 = *model.ValidationError", err)
	}
	if len(ve.Global) == 0 {
		t.Error("バージョンを読めなかった旨の全体エラーを期待したが、無かった")
	}

	if work := reloadSyncWork(t, db, workID); work.Title != title {
		t.Errorf("work.Title = %q、期待値 = %q (版を示さない送信は適用されない)", work.Title, title)
	}
}

// TestUpdateWorkUsecase_Execute_MatchesNullVersionはupdated_atがNULLの作品を検証する
// (共有カラムはこれを許す)。nullのセンチネルはそれ自体が1つの版であり、これを名乗る最初の送信は
// 適用されてカラムを進める。したがって同じセンチネルを名乗る2件目は、1件目を上書きせず競合する。
func TestUpdateWorkUsecase_Execute_MatchesNullVersion(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateWorkUsecase(db)

	workID := createMappedWork(t, db, "版NULLアニメ_"+t.Name())
	if _, err := db.Exec(`UPDATE works SET updated_at = NULL WHERE id = $1`, int64(workID)); err != nil {
		t.Fatalf("updated_atのNULL化に失敗: %v", err)
	}

	input := validUpdateWorkInput(t, db, workID, "版NULL更新後_"+t.Name())
	if input.UpdatedAt != validator.FormNullVersion {
		t.Fatalf("UpdatedAt = %q、期待値 = %q", input.UpdatedAt, validator.FormNullVersion)
	}
	if _, err := uc.Execute(context.Background(), input); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if work := reloadSyncWork(t, db, workID); work.Title != input.Title {
		t.Errorf("work.Title = %q、期待値 = %q", work.Title, input.Title)
	}

	second := input
	second.Title = "版NULL二件目_" + t.Name()
	ae := model.AsAppError(mustExecuteError(t, uc, second))
	if ae == nil || ae.Code != model.AppErrCodeConflict {
		t.Fatalf("NULLバージョンからの2回目の送信でAppErrCodeConflictを期待したが、違った")
	}
}

// mustExecuteErrorは失敗が期待される送信を実行し、そのエラーを返す。
func mustExecuteError(t *testing.T, uc *UpdateWorkUsecase, input UpdateWorkInput) error {
	t.Helper()

	output, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatalf("Execute() = %+v、期待値 = エラーあり", output)
	}

	return err
}

// TestUpdateWorkUsecase_Execute_TitleUniquenessExcludesItselfはタイトル一意性検査の
// 自分自身の除外を対象とする。除外が無いと、作品の他のフィールドを編集するだけで自分自身の
// タイトルで失敗する。検査が必ず見つける行だからである。
func TestUpdateWorkUsecase_Execute_TitleUniquenessExcludesItself(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateWorkUsecase(db)

	title := "自己重複テストアニメ_" + t.Name()
	workID := createMappedWork(t, db, title)

	// 他のフィールドを変えつつタイトルはそのまま送る。タイトル以外を編集したときに
	// 送信される内容そのもの。
	if _, err := uc.Execute(context.Background(), validUpdateWorkInput(t, db, workID, title)); err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
}

// TestUpdateWorkUsecase_Execute_RejectsDuplicateTitleは、生存中の別の作品が既に持つ
// タイトルへ改名するケースを対象とする。
func TestUpdateWorkUsecase_Execute_RejectsDuplicateTitle(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	uc := newUpdateWorkUsecase(db)

	existingTitle := "重複相手アニメ_" + t.Name()
	createMappedWork(t, db, existingTitle)
	workID := createMappedWork(t, db, "改名元アニメ_"+t.Name())

	output, err := uc.Execute(context.Background(), validUpdateWorkInput(t, db, workID, existingTitle))
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
