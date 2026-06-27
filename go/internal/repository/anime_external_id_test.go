package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestAnimeExternalIDRepository_CreateAndListByAnimeIDs(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeExternalIDRepository(queries)

	animeID := createTestAnime(t, animeRepo, "外部 ID 同期アニメ")

	// One anime can hold at most one row per service; create both syobocal and mal.
	//
	// [Ja] 1 つの anime はサービスごとに高々 1 行を持つ。syobocal と mal の両方を作成する。
	if _, err := repo.Create(context.Background(), repository.CreateAnimeExternalIDParams{
		AnimeID:    animeID,
		Service:    model.AnimeExternalServiceSyobocal,
		ExternalID: "12345",
	}); err != nil {
		t.Fatalf("Create(syobocal) error = %v", err)
	}
	if _, err := repo.Create(context.Background(), repository.CreateAnimeExternalIDParams{
		AnimeID:    animeID,
		Service:    model.AnimeExternalServiceMal,
		ExternalID: "678",
	}); err != nil {
		t.Fatalf("Create(mal) error = %v", err)
	}

	got, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}

	// Rows come back ordered by (anime_id, service); 'mal' sorts before 'syobocal'.
	//
	// [Ja] 行は (anime_id, service) 順で返る。'mal' は 'syobocal' より前に並ぶ。
	byService := map[model.AnimeExternalService]string{}
	for _, e := range got {
		if e.ID == 0 {
			t.Error("ID should be assigned")
		}
		if e.AnimeID != animeID {
			t.Errorf("AnimeID = %d, want %d", e.AnimeID, animeID)
		}
		byService[e.Service] = e.ExternalID
	}
	if byService[model.AnimeExternalServiceSyobocal] != "12345" {
		t.Errorf("syobocal external_id = %q, want 12345", byService[model.AnimeExternalServiceSyobocal])
	}
	if byService[model.AnimeExternalServiceMal] != "678" {
		t.Errorf("mal external_id = %q, want 678", byService[model.AnimeExternalServiceMal])
	}
}

func TestAnimeExternalIDRepository_Update(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeExternalIDRepository(queries)

	animeID := createTestAnime(t, animeRepo, "外部 ID 更新アニメ")

	created, err := repo.Create(context.Background(), repository.CreateAnimeExternalIDParams{
		AnimeID:    animeID,
		Service:    model.AnimeExternalServiceSyobocal,
		ExternalID: "100",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.Update(context.Background(), repository.UpdateAnimeExternalIDParams{
		ID:         created.ID,
		ExternalID: "200",
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}
	if got[0].ExternalID != "200" {
		t.Errorf("ExternalID = %q, want 200 after update", got[0].ExternalID)
	}
}

func TestAnimeExternalIDRepository_Delete(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeExternalIDRepository(queries)

	animeID := createTestAnime(t, animeRepo, "外部 ID 削除アニメ")

	created, err := repo.Create(context.Background(), repository.CreateAnimeExternalIDParams{
		AnimeID:    animeID,
		Service:    model.AnimeExternalServiceMal,
		ExternalID: "999",
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

func TestAnimeExternalIDRepository_ListByAnimeIDs_EmptyInput(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeExternalIDRepository(query.New(db).WithTx(tx))

	got, err := repo.ListByAnimeIDs(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListByAnimeIDs() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("len(got) = %d, want 0 for empty input", len(got))
	}
}
