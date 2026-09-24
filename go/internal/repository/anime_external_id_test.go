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

	animeID := createTestAnime(t, animeRepo, "外部ID同期アニメ")

	// 1つのanimeはサービスごとに高々1行を持つ。syobocalとmalの両方を作成する。
	if _, err := repo.Create(context.Background(), repository.CreateAnimeExternalIDParams{
		AnimeID:    animeID,
		Service:    model.AnimeExternalServiceSyobocal,
		ExternalID: "12345",
	}); err != nil {
		t.Fatalf("Create(syobocal)のエラー = %v", err)
	}
	if _, err := repo.Create(context.Background(), repository.CreateAnimeExternalIDParams{
		AnimeID:    animeID,
		Service:    model.AnimeExternalServiceMal,
		ExternalID: "678",
	}); err != nil {
		t.Fatalf("Create(mal)のエラー = %v", err)
	}

	got, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs()のエラー = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d、期待値 = 2", len(got))
	}

	// 行は (anime_id, service) 順で返る。'mal' は 'syobocal' より前に並ぶ。
	byService := map[model.AnimeExternalService]string{}
	for _, e := range got {
		if e.ID == 0 {
			t.Error("IDが採番されていない")
		}
		if e.AnimeID != animeID {
			t.Errorf("AnimeID = %d、期待値 = %d", e.AnimeID, animeID)
		}
		byService[e.Service] = e.ExternalID
	}
	if byService[model.AnimeExternalServiceSyobocal] != "12345" {
		t.Errorf("syobocal external_id = %q、期待値 = 12345", byService[model.AnimeExternalServiceSyobocal])
	}
	if byService[model.AnimeExternalServiceMal] != "678" {
		t.Errorf("mal external_id = %q、期待値 = 678", byService[model.AnimeExternalServiceMal])
	}
}

func TestAnimeExternalIDRepository_Update(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeExternalIDRepository(queries)

	animeID := createTestAnime(t, animeRepo, "外部ID更新アニメ")

	created, err := repo.Create(context.Background(), repository.CreateAnimeExternalIDParams{
		AnimeID:    animeID,
		Service:    model.AnimeExternalServiceSyobocal,
		ExternalID: "100",
	})
	if err != nil {
		t.Fatalf("Create()のエラー = %v", err)
	}

	if err := repo.Update(context.Background(), repository.UpdateAnimeExternalIDParams{
		ID:         created.ID,
		ExternalID: "200",
	}); err != nil {
		t.Fatalf("Update()のエラー = %v", err)
	}

	got, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs()のエラー = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d、期待値 = 1", len(got))
	}
	if got[0].ExternalID != "200" {
		t.Errorf("更新後のExternalID = %q、期待値 = 200", got[0].ExternalID)
	}
}

func TestAnimeExternalIDRepository_Delete(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeExternalIDRepository(queries)

	animeID := createTestAnime(t, animeRepo, "外部ID削除アニメ")

	created, err := repo.Create(context.Background(), repository.CreateAnimeExternalIDParams{
		AnimeID:    animeID,
		Service:    model.AnimeExternalServiceMal,
		ExternalID: "999",
	})
	if err != nil {
		t.Fatalf("Create()のエラー = %v", err)
	}

	if err := repo.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("Delete()のエラー = %v", err)
	}

	got, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs()のエラー = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("削除後のlen(got) = %d、期待値 = 0", len(got))
	}
}

func TestAnimeExternalIDRepository_ListByAnimeIDs_EmptyInput(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeExternalIDRepository(query.New(db).WithTx(tx))

	got, err := repo.ListByAnimeIDs(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListByAnimeIDs()のエラー = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("空入力時のlen(got) = %d、期待値 = 0", len(got))
	}
}
