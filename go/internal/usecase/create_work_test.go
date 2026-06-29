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
		validator.NewDbWorkCreateValidator(),
	)
}

// validCreateWorkInput returns a form input that passes DbWorkCreateValidator, with
// enough non-default fields set to exercise the work -> anime / classification mapping.
// The title is taken as an argument so each test can pass a unique value (e.g. t.Name()):
// these tests use GetTestDB and commit their rows to the shared DB, so a per-test title
// keeps parallel tests from sharing works rows.
//
// [Ja] validCreateWorkInput は DbWorkCreateValidator を通過するフォーム入力を返す。
// work -> anime / 分類 の写像を検証できるよう、非デフォルトのフィールドを十分にセットする。
// タイトルは引数で受け取り、各テストがユニークな値 (例: t.Name()) を渡せるようにする。
// 本テスト群は GetTestDB を使い行を共有 DB にコミットするため、テストごとのタイトルで
// 並行テストが works 行を共有しないようにする。
func validCreateWorkInput(title string) CreateWorkInput {
	return CreateWorkInput{
		Title:                 title,
		TitleKana:             "さくせいてすとあにめ",
		TitleEn:               "Create Test Anime",
		Media:                 "2", // OVA
		Synopsis:              "あらすじ本文",
		SynopsisSource:        "出典",
		ManualEpisodesCount:   "12",
		StartEpisodeRawNumber: "2.5",
		NoEpisodes:            "1",
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
