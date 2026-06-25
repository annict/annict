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

// createTestAnime inserts a minimal anime via the repository and returns its ID,
// used as the identity a classification attaches to.
//
// [Ja] createTestAnime はリポジトリ経由で最小のアニメを 1 件作成し、その ID を返す。
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

	t.Run("正常系: work 分類を作成する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		animeRepo := repository.NewAnimeRepository(queries)
		repo := repository.NewAnimeClassificationRepository(queries)

		animeID := createTestAnime(t, animeRepo, "作品アニメ")

		// A work carries no parent and no sort_number, but may carry the
		// generation settings (episode_start_number / expected_episodes_count).
		//
		// [Ja] work は親と sort_number を持たないが、生成設定
		// (episode_start_number / expected_episodes_count) は持てる。
		created, err := repo.Create(context.Background(), repository.CreateAnimeClassificationParams{
			AnimeID:               animeID,
			Kind:                  model.AnimeClassificationKindWork,
			Standalone:            true,
			EpisodeStartNumber:    nullStr("1"),
			ExpectedEpisodesCount: sql.NullInt32{Int32: 12, Valid: true},
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if created.ID == 0 {
			t.Error("created.ID should be assigned")
		}

		got, err := repo.GetByAnimeID(context.Background(), animeID)
		if err != nil {
			t.Fatalf("GetByAnimeID() error = %v", err)
		}
		if got == nil {
			t.Fatal("GetByAnimeID() returned nil for an existing classification")
		}
		if got.Kind != model.AnimeClassificationKindWork {
			t.Errorf("Kind = %q, want work", got.Kind)
		}
		if got.ParentAnimeID != nil {
			t.Errorf("ParentAnimeID = %v, want nil for a work", got.ParentAnimeID)
		}
		if !got.Standalone {
			t.Error("Standalone = false, want true")
		}
		if got.EpisodeStartNumber.String != "1" {
			t.Errorf("EpisodeStartNumber = %q, want 1", got.EpisodeStartNumber.String)
		}
		if got.ExpectedEpisodesCount.Int32 != 12 {
			t.Errorf("ExpectedEpisodesCount = %d, want 12", got.ExpectedEpisodesCount.Int32)
		}
	})

	t.Run("正常系: episode 分類を親作品付きで作成する", func(t *testing.T) {
		t.Parallel()

		db, tx := testutil.SetupTx(t)
		queries := query.New(db).WithTx(tx)
		animeRepo := repository.NewAnimeRepository(queries)
		repo := repository.NewAnimeClassificationRepository(queries)

		parentAnimeID := createTestAnime(t, animeRepo, "親作品アニメ")
		episodeAnimeID := createTestAnime(t, animeRepo, "エピソードアニメ")

		// An episode always carries a parent and a sort_number; the numeric
		// number preserves a decimal recap such as 3.5.
		//
		// [Ja] episode は必ず親と sort_number を持つ。数値の number は 3.5 の
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
			t.Fatalf("Create() error = %v", err)
		}

		got, err := repo.GetByAnimeID(context.Background(), episodeAnimeID)
		if err != nil {
			t.Fatalf("GetByAnimeID() error = %v", err)
		}
		if got.Kind != model.AnimeClassificationKindEpisode {
			t.Errorf("Kind = %q, want episode", got.Kind)
		}
		if got.ParentAnimeID == nil || *got.ParentAnimeID != parentAnimeID {
			t.Errorf("ParentAnimeID = %v, want %d", got.ParentAnimeID, parentAnimeID)
		}
		if got.Number.String != "3.5" {
			t.Errorf("Number = %q, want 3.5", got.Number.String)
		}
		if got.NumberText.String != "第3.5話" {
			t.Errorf("NumberText = %q, want 第3.5話", got.NumberText.String)
		}
		if got.SortNumber.Int32 != 35 {
			t.Errorf("SortNumber = %d, want 35", got.SortNumber.Int32)
		}
	})
}

func TestAnimeClassificationRepository_Create_ConstraintViolation(t *testing.T) {
	t.Parallel()

	// The work/episode shape is enforced by CHECK constraints in the schema, not
	// by the repository (CreateAnimeClassificationParams documents that supplying a
	// consistent shape is the caller's responsibility). These cases verify that an
	// inconsistent shape is rejected by the database and the error is propagated
	// out of Create, so the param struct and the constraints cannot silently drift
	// apart.
	//
	// [Ja] work/episode の形状はリポジトリではなくスキーマの CHECK 制約で守られる
	// (整合した形での指定は呼び出し元の責務だと CreateAnimeClassificationParams が
	// 明記している)。以下のケースは、不整合な形状が DB に拒否され、そのエラーが
	// Create から伝搬することを確認する。param 構造体と CHECK 制約が気付かぬうちに
	// ドリフトしないようにする。
	tests := []struct {
		name   string
		params func(animeID model.AnimeID) repository.CreateAnimeClassificationParams
	}{
		{
			// A work must not carry sort_number (sort_number_check).
			//
			// [Ja] work は sort_number を持てない (sort_number_check)。
			name: "異常系: work に sort_number を設定すると CHECK 制約違反になる",
			params: func(animeID model.AnimeID) repository.CreateAnimeClassificationParams {
				return repository.CreateAnimeClassificationParams{
					AnimeID:    animeID,
					Kind:       model.AnimeClassificationKindWork,
					SortNumber: sql.NullInt32{Int32: 1, Valid: true},
				}
			},
		},
		{
			// An episode must carry a parent (parent_check).
			//
			// [Ja] episode は親を持たなければならない (parent_check)。
			name: "異常系: 親を持たない episode は CHECK 制約違反になる",
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
				t.Fatal("Create() error = nil, want a CHECK constraint violation")
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
		t.Fatalf("GetByAnimeID() error = %v", err)
	}
	if got != nil {
		t.Errorf("GetByAnimeID() = %+v, want nil for a missing classification", got)
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
		t.Fatalf("Create() error = %v", err)
	}

	err := repo.UpdateByAnimeID(context.Background(), repository.UpdateAnimeClassificationParams{
		AnimeID:               animeID,
		Kind:                  model.AnimeClassificationKindWork,
		Standalone:            true,
		ExpectedEpisodesCount: sql.NullInt32{Int32: 24, Valid: true},
	})
	if err != nil {
		t.Fatalf("UpdateByAnimeID() error = %v", err)
	}

	got, err := repo.GetByAnimeID(context.Background(), animeID)
	if err != nil {
		t.Fatalf("GetByAnimeID() error = %v", err)
	}
	if !got.Standalone {
		t.Error("Standalone = false, want true after update")
	}
	if got.ExpectedEpisodesCount.Int32 != 24 {
		t.Errorf("ExpectedEpisodesCount = %d, want 24 after update", got.ExpectedEpisodesCount.Int32)
	}
}
