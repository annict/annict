package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestAnimeOfficialAccountRepository_CreateAndListByAnimeIDs(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeOfficialAccountRepository(queries)

	animeID := createTestAnime(t, animeRepo, "公式アカウント同期アニメ")

	// (anime_id, service) is unique, so an anime holds at most one row per service;
	// create the works-sourced x account and an editor-style youtube account.
	//
	// [Ja] (anime_id, service) はユニークなので、1 つの anime はサービスごとに高々 1 行を持つ。
	// works 由来の x アカウントと、編集者風の youtube アカウントを作成する。
	if _, err := repo.Create(context.Background(), repository.CreateAnimeOfficialAccountParams{
		AnimeID: animeID,
		Service: model.AnimeAccountServiceX,
		Account: "rezero_official",
	}); err != nil {
		t.Fatalf("Create(x) error = %v", err)
	}
	if _, err := repo.Create(context.Background(), repository.CreateAnimeOfficialAccountParams{
		AnimeID: animeID,
		Service: model.AnimeAccountServiceYoutube,
		Account: "rezeroanime",
	}); err != nil {
		t.Fatalf("Create(youtube) error = %v", err)
	}

	got, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}

	byService := map[model.AnimeAccountService]string{}
	for _, a := range got {
		if a.ID == 0 {
			t.Error("ID should be assigned")
		}
		if a.AnimeID != animeID {
			t.Errorf("AnimeID = %d, want %d", a.AnimeID, animeID)
		}
		// works do not source label / label_en, so they stay nil on a freshly created row.
		//
		// [Ja] works は label / label_en を source しないため、作成直後の行では nil のまま。
		if a.Label != nil || a.LabelEn != nil {
			t.Errorf("Label/LabelEn = %v/%v, want nil/nil", a.Label, a.LabelEn)
		}
		byService[a.Service] = a.Account
	}
	if byService[model.AnimeAccountServiceX] != "rezero_official" {
		t.Errorf("x account = %q, want rezero_official", byService[model.AnimeAccountServiceX])
	}
	if byService[model.AnimeAccountServiceYoutube] != "rezeroanime" {
		t.Errorf("youtube account = %q, want rezeroanime", byService[model.AnimeAccountServiceYoutube])
	}
}

func TestAnimeOfficialAccountRepository_Update(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeOfficialAccountRepository(queries)

	animeID := createTestAnime(t, animeRepo, "公式アカウント更新アニメ")

	created, err := repo.Create(context.Background(), repository.CreateAnimeOfficialAccountParams{
		AnimeID: animeID,
		Service: model.AnimeAccountServiceX,
		Account: "old_handle",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.Update(context.Background(), repository.UpdateAnimeOfficialAccountParams{
		ID:      created.ID,
		Account: "new_handle",
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
	if got[0].Account != "new_handle" {
		t.Errorf("Account = %q, want new_handle after update", got[0].Account)
	}
}

func TestAnimeOfficialAccountRepository_Delete(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeOfficialAccountRepository(queries)

	animeID := createTestAnime(t, animeRepo, "公式アカウント削除アニメ")

	created, err := repo.Create(context.Background(), repository.CreateAnimeOfficialAccountParams{
		AnimeID: animeID,
		Service: model.AnimeAccountServiceX,
		Account: "to_be_deleted",
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

func TestAnimeOfficialAccountRepository_ListByAnimeIDs_EmptyInput(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeOfficialAccountRepository(query.New(db).WithTx(tx))

	got, err := repo.ListByAnimeIDs(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListByAnimeIDs() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("len(got) = %d, want 0 for empty input", len(got))
	}
}
