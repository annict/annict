package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// createTestAnimeはリポジトリ経由で最小のアニメを1件作成し、そのIDを返す。
// 分類が結びつく同一性として使う。
func createTestAnime(t *testing.T, repo *repository.AnimeRepository, title string) model.AnimeID {
	t.Helper()
	anime, err := repo.Create(context.Background(), repository.CreateAnimeParams{
		Title: nullStr(title),
	})
	if err != nil {
		t.Fatalf("アニメの作成に失敗しました: %v", err)
	}
	return anime.ID
}

func TestAnimeClassificationRepository_Create(t *testing.T) {
	t.Parallel()

	t.Run("正常系: work分類を作成する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		animeRepo := repository.NewAnimeRepository(queries)
		repo := repository.NewAnimeClassificationRepository(queries)

		animeID := createTestAnime(t, animeRepo, "作品アニメ")

		// workは親とsort_numberを持たないが、生成設定
		// (episode_start_number / expected_episodes_count) は持てる。
		created, err := repo.Create(context.Background(), repository.CreateAnimeClassificationParams{
			AnimeID:               animeID,
			Kind:                  model.AnimeClassificationKindWork,
			Standalone:            true,
			EpisodeStartNumber:    nullStr("1"),
			ExpectedEpisodesCount: sql.NullInt32{Int32: 12, Valid: true},
		})
		if err != nil {
			t.Fatalf("Create()のエラー = %v", err)
		}
		if created.ID == 0 {
			t.Error("created.IDが採番されていない")
		}

		got, err := repo.GetByAnimeID(context.Background(), animeID)
		if err != nil {
			t.Fatalf("GetByAnimeID()のエラー = %v", err)
		}
		if got == nil {
			t.Fatal("既存のclassificationに対してGetByAnimeID()がnilを返した")
		}
		if got.Kind != model.AnimeClassificationKindWork {
			t.Errorf("Kind = %q、期待値 = work", got.Kind)
		}
		if got.ParentAnimeID != nil {
			t.Errorf("ParentAnimeID = %v、期待値 = nil (作品のため)", got.ParentAnimeID)
		}
		if !got.Standalone {
			t.Error("Standalone = false、期待値 = true")
		}
		if got.EpisodeStartNumber.String != "1" {
			t.Errorf("EpisodeStartNumber = %q、期待値 = 1", got.EpisodeStartNumber.String)
		}
		if got.ExpectedEpisodesCount.Int32 != 12 {
			t.Errorf("ExpectedEpisodesCount = %d、期待値 = 12", got.ExpectedEpisodesCount.Int32)
		}
	})

	t.Run("正常系: episode分類を親作品付きで作成する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		animeRepo := repository.NewAnimeRepository(queries)
		repo := repository.NewAnimeClassificationRepository(queries)

		parentAnimeID := createTestAnime(t, animeRepo, "親作品アニメ")
		episodeAnimeID := createTestAnime(t, animeRepo, "エピソードアニメ")

		// episodeは必ず親とsort_numberを持つ。数値のnumberは3.5の
		// ような小数の総集編をそのまま保つ。
		_, err := repo.Create(context.Background(), repository.CreateAnimeClassificationParams{
			AnimeID:       episodeAnimeID,
			Kind:          model.AnimeClassificationKindEpisode,
			ParentAnimeID: &parentAnimeID,
			Number:        nullStr("3.5"),
			NumberText:    nullStr("第3.5話"),
			SortNumber:    sql.NullInt32{Int32: 35, Valid: true},
		})
		if err != nil {
			t.Fatalf("Create()のエラー = %v", err)
		}

		got, err := repo.GetByAnimeID(context.Background(), episodeAnimeID)
		if err != nil {
			t.Fatalf("GetByAnimeID()のエラー = %v", err)
		}
		if got.Kind != model.AnimeClassificationKindEpisode {
			t.Errorf("Kind = %q、期待値 = episode", got.Kind)
		}
		if got.ParentAnimeID == nil || *got.ParentAnimeID != parentAnimeID {
			t.Errorf("ParentAnimeID = %v、期待値 = %d", got.ParentAnimeID, parentAnimeID)
		}
		if got.Number.String != "3.5" {
			t.Errorf("Number = %q、期待値 = 3.5", got.Number.String)
		}
		if got.NumberText.String != "第3.5話" {
			t.Errorf("NumberText = %q、期待値 = 第3.5話", got.NumberText.String)
		}
		if got.SortNumber.Int32 != 35 {
			t.Errorf("SortNumber = %d、期待値 = 35", got.SortNumber.Int32)
		}
	})
}

func TestAnimeClassificationRepository_Create_ConstraintViolation(t *testing.T) {
	t.Parallel()

	// work/episodeの形状はリポジトリではなくスキーマのCHECK制約で守られる
	// (整合した形での指定は呼び出し元の責務だとCreateAnimeClassificationParamsが
	// 明記している)。以下のケースは、不整合な形状がDBに拒否され、そのエラーが
	// Createから伝搬することを確認する。param構造体とCHECK制約が気付かぬうちに
	// ドリフトしないようにする。
	tests := []struct {
		name   string
		params func(animeID model.AnimeID) repository.CreateAnimeClassificationParams
	}{
		{
			// workはsort_numberを持てない (sort_number_check)。
			name: "異常系: workにsort_numberを設定するとCHECK制約違反になる",
			params: func(animeID model.AnimeID) repository.CreateAnimeClassificationParams {
				return repository.CreateAnimeClassificationParams{
					AnimeID:    animeID,
					Kind:       model.AnimeClassificationKindWork,
					SortNumber: sql.NullInt32{Int32: 1, Valid: true},
				}
			},
		},
		{
			// episodeは親を持たなければならない (parent_check)。
			name: "異常系: 親を持たないepisodeはCHECK制約違反になる",
			params: func(animeID model.AnimeID) repository.CreateAnimeClassificationParams {
				return repository.CreateAnimeClassificationParams{
					AnimeID:    animeID,
					Kind:       model.AnimeClassificationKindEpisode,
					SortNumber: sql.NullInt32{Int32: 1, Valid: true},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, tx := testutil.SetupTx(t)
			queries := query.New(db).WithTx(tx)
			animeRepo := repository.NewAnimeRepository(queries)
			repo := repository.NewAnimeClassificationRepository(queries)

			animeID := createTestAnime(t, animeRepo, "制約違反テストアニメ")

			if _, err := repo.Create(context.Background(), tt.params(animeID)); err == nil {
				t.Fatal("Create()のエラー = nil、期待値 = CHECK制約違反")
			}
		})
	}
}

func TestAnimeClassificationRepository_GetByAnimeID_NotFound(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeClassificationRepository(query.New(db).WithTx(tx))

	got, err := repo.GetByAnimeID(context.Background(), model.AnimeID(999999999))
	if err != nil {
		t.Fatalf("GetByAnimeID()のエラー = %v", err)
	}
	if got != nil {
		t.Errorf("GetByAnimeID() = %+v、期待値 = nil (存在しないclassificationのため)", got)
	}
}

func TestAnimeClassificationRepository_UpdateByAnimeID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeClassificationRepository(queries)

	animeID := createTestAnime(t, animeRepo, "更新対象アニメ")

	if _, err := repo.Create(context.Background(), repository.CreateAnimeClassificationParams{
		AnimeID:               animeID,
		Kind:                  model.AnimeClassificationKindWork,
		Standalone:            false,
		ExpectedEpisodesCount: sql.NullInt32{Int32: 12, Valid: true},
	}); err != nil {
		t.Fatalf("Create()のエラー = %v", err)
	}

	err := repo.UpdateByAnimeID(context.Background(), repository.UpdateAnimeClassificationParams{
		AnimeID:               animeID,
		Kind:                  model.AnimeClassificationKindWork,
		Standalone:            true,
		ExpectedEpisodesCount: sql.NullInt32{Int32: 24, Valid: true},
	})
	if err != nil {
		t.Fatalf("UpdateByAnimeID()のエラー = %v", err)
	}

	got, err := repo.GetByAnimeID(context.Background(), animeID)
	if err != nil {
		t.Fatalf("GetByAnimeID()のエラー = %v", err)
	}
	if !got.Standalone {
		t.Error("更新後のStandalone = false、期待値 = true")
	}
	if got.ExpectedEpisodesCount.Int32 != 24 {
		t.Errorf("更新後のExpectedEpisodesCount = %d、期待値 = 24", got.ExpectedEpisodesCount.Int32)
	}
}

// TestAnimeClassificationRepository_Upsertはエピソード編集に必要な2経路を検証する。
// 欠損した分類を挿入し、後続の呼び出しでは同じ行をその場で更新する。
func TestAnimeClassificationRepository_Upsert(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeClassificationRepository(queries)

	parentAnimeID := createTestAnime(t, animeRepo, "親アニメ")
	animeID := createTestAnime(t, animeRepo, "エピソードアニメ")
	params := repository.CreateAnimeClassificationParams{
		AnimeID:       animeID,
		Kind:          model.AnimeClassificationKindEpisode,
		ParentAnimeID: &parentAnimeID,
		Number:        nullStr("1"),
		NumberText:    nullStr("#1"),
		SortNumber:    sql.NullInt32{Int32: 100, Valid: true},
	}

	if err := repo.Upsert(context.Background(), params); err != nil {
		t.Fatalf("1回目のUpsert()のエラー = %v", err)
	}
	created, err := repo.GetByAnimeID(context.Background(), animeID)
	if err != nil || created == nil {
		t.Fatalf("1回目のGetByAnimeID() classification=%v err=%v", created, err)
	}
	if created.NumberText.String != "#1" || created.SortNumber.Int32 != 100 {
		t.Errorf("1回目のclassification = %+v、期待値 = number_text=#1 sort_number=100", created)
	}

	params.Number = nullStr("2.5")
	params.NumberText = nullStr("第2話")
	params.SortNumber = sql.NullInt32{Int32: 250, Valid: true}
	if err := repo.Upsert(context.Background(), params); err != nil {
		t.Fatalf("2回目のUpsert()のエラー = %v", err)
	}
	updated, err := repo.GetByAnimeID(context.Background(), animeID)
	if err != nil || updated == nil {
		t.Fatalf("2回目のGetByAnimeID() classification=%v err=%v", updated, err)
	}
	if updated.ID != created.ID {
		t.Errorf("2回目のID = %d、期待値 = %d (同じ行を更新)", int64(updated.ID), int64(created.ID))
	}
	if updated.Number.String != "2.5" || updated.NumberText.String != "第2話" {
		t.Errorf("2回目のclassification = %+v、期待値 = 送信した話数情報", updated)
	}
	if !updated.SortNumber.Valid || updated.SortNumber.Int32 != 250 {
		t.Errorf("2回目のSortNumber = %+v、期待値 = {250 true}", updated.SortNumber)
	}
}
