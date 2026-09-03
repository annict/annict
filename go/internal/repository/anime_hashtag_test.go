package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestAnimeHashtagRepository_CreateAndListByAnimeIDs(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeHashtagRepository(queries)

	animeID := createTestAnime(t, animeRepo, "ハッシュタグ同期アニメ")

	// (anime_id, hashtag) is unique, so an anime holds at most one row per tag value;
	// create two distinct tags.
	//
	// [Ja] (anime_id, hashtag) はユニークなので、1 つの anime はタグ値ごとに高々 1 行を持つ。
	// 異なる 2 つのタグを作成する。
	for _, tag := range []string{"rezero", "rezero2nd"} {
		if _, err := repo.Create(context.Background(), repository.CreateAnimeHashtagParams{
			AnimeID: animeID,
			Hashtag: tag,
		}); err != nil {
			t.Fatalf("Create(%s) error = %v", tag, err)
		}
	}

	got, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}

	tags := map[string]bool{}
	for _, h := range got {
		if h.ID == 0 {
			t.Error("ID should be assigned")
		}
		if h.AnimeID != animeID {
			t.Errorf("AnimeID = %d, want %d", h.AnimeID, animeID)
		}
		// works do not source sort_number, so a freshly created row keeps the DB default 0.
		//
		// [Ja] works は sort_number を source しないため、作成直後の行は DB 既定値の 0 のまま。
		if h.SortNumber != 0 {
			t.Errorf("SortNumber = %d, want 0", h.SortNumber)
		}
		tags[h.Hashtag] = true
	}
	if !tags["rezero"] || !tags["rezero2nd"] {
		t.Errorf("tags = %v, want rezero and rezero2nd", tags)
	}
}

func TestAnimeHashtagRepository_Delete(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeHashtagRepository(queries)

	animeID := createTestAnime(t, animeRepo, "ハッシュタグ削除アニメ")

	created, err := repo.Create(context.Background(), repository.CreateAnimeHashtagParams{
		AnimeID: animeID,
		Hashtag: "to_be_deleted",
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

func TestAnimeHashtagRepository_ListByAnimeIDs_EmptyInput(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeHashtagRepository(query.New(db).WithTx(tx))

	got, err := repo.ListByAnimeIDs(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListByAnimeIDs() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("len(got) = %d, want 0 for empty input", len(got))
	}
}
