package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestAnimeSeasonRepository_CreateAndListByAnimeIDs(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeSeasonRepository(queries)

	animeID := createTestAnime(t, animeRepo, "季節同期アニメ")

	// One works-managed primary season with a name, plus a name-less row for a
	// different year. is_primary differs so the partial UNIQUE index on (anime_id)
	// WHERE is_primary is satisfied, and the two (year, name) keys are distinct.
	//
	// [Ja] 名前付きの works 管理下の主季節 1 つと、別の年の名前なし行を作る。is_primary が
	// 異なるので (anime_id) WHERE is_primary の部分 UNIQUE インデックスを満たし、2 つの
	// (year, name) キーも異なる。
	spring := model.SeasonNameSpring
	if _, err := repo.Create(context.Background(), repository.CreateAnimeSeasonParams{
		AnimeID: animeID, Year: 2024, Name: &spring, IsPrimary: true,
	}); err != nil {
		t.Fatalf("Create(named) error = %v", err)
	}
	if _, err := repo.Create(context.Background(), repository.CreateAnimeSeasonParams{
		AnimeID: animeID, Year: 2023, Name: nil, IsPrimary: false,
	}); err != nil {
		t.Fatalf("Create(name-less) error = %v", err)
	}

	got, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}

	byYear := map[int32]*model.AnimeSeason{}
	for _, s := range got {
		if s.ID == 0 {
			t.Error("ID should be assigned")
		}
		if s.AnimeID != animeID {
			t.Errorf("AnimeID = %d, want %d", s.AnimeID, animeID)
		}
		byYear[s.Year] = s
	}

	named := byYear[2024]
	if named == nil || named.Name == nil || *named.Name != model.SeasonNameSpring || !named.IsPrimary {
		t.Errorf("2024 row = %+v, want spring / is_primary true", named)
	}
	nameless := byYear[2023]
	if nameless == nil || nameless.Name != nil || nameless.IsPrimary {
		t.Errorf("2023 row = %+v, want NULL name / is_primary false", nameless)
	}
}

func TestAnimeSeasonRepository_Delete(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeSeasonRepository(queries)

	animeID := createTestAnime(t, animeRepo, "季節削除アニメ")

	summer := model.SeasonNameSummer
	created, err := repo.Create(context.Background(), repository.CreateAnimeSeasonParams{
		AnimeID: animeID, Year: 2025, Name: &summer, IsPrimary: true,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	got, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("len(got) = %d, want 0 after delete", len(got))
	}
}

func TestAnimeSeasonRepository_ListByAnimeIDs_EmptyInput(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeSeasonRepository(query.New(db).WithTx(tx))

	got, err := repo.ListByAnimeIDs(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListByAnimeIDs() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("len(got) = %d, want 0 for empty input", len(got))
	}
}
