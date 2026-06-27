package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestAnimeLinkRepository_CreateAndListByAnimeIDs(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeLinkRepository(queries)

	animeID := createTestAnime(t, animeRepo, "リンク同期アニメ")

	// An anime can hold one link per (kind, language); create an official site (ja)
	// and a Wikipedia page (en).
	//
	// [Ja] 1 つの anime は (kind, language) ごとに 1 リンクを持つ。公式サイト (ja) と
	// Wikipedia (en) を作成する。
	if _, err := repo.Create(context.Background(), repository.CreateAnimeLinkParams{
		AnimeID:  animeID,
		Kind:     model.AnimeLinkKindOfficialSite,
		Language: model.LanguageJa,
		URL:      "https://example.dev/official",
	}); err != nil {
		t.Fatalf("Create(official_site, ja) error = %v", err)
	}
	if _, err := repo.Create(context.Background(), repository.CreateAnimeLinkParams{
		AnimeID:  animeID,
		Kind:     model.AnimeLinkKindWikipedia,
		Language: model.LanguageEn,
		URL:      "https://en.wikipedia.org/wiki/Example",
	}); err != nil {
		t.Fatalf("Create(wikipedia, en) error = %v", err)
	}

	got, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}

	byKey := map[animeLinkTestKey]*model.AnimeLink{}
	for _, l := range got {
		if l.ID == 0 {
			t.Error("ID should be assigned")
		}
		if l.AnimeID != animeID {
			t.Errorf("AnimeID = %d, want %d", l.AnimeID, animeID)
		}
		// Works do not source labels, so created rows leave them nil.
		//
		// [Ja] works は label を source しないため、作成した行では nil のまま。
		if l.Label != nil || l.LabelEn != nil {
			t.Errorf("Label/LabelEn = %v/%v, want nil/nil", l.Label, l.LabelEn)
		}
		byKey[animeLinkTestKey{l.Kind, l.Language}] = l
	}

	if got := byKey[animeLinkTestKey{model.AnimeLinkKindOfficialSite, model.LanguageJa}]; got == nil || got.URL != "https://example.dev/official" {
		t.Errorf("official_site/ja = %+v, want url https://example.dev/official", got)
	}
	if got := byKey[animeLinkTestKey{model.AnimeLinkKindWikipedia, model.LanguageEn}]; got == nil || got.URL != "https://en.wikipedia.org/wiki/Example" {
		t.Errorf("wikipedia/en = %+v, want url https://en.wikipedia.org/wiki/Example", got)
	}
}

func TestAnimeLinkRepository_Update(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeLinkRepository(queries)

	animeID := createTestAnime(t, animeRepo, "リンク更新アニメ")

	created, err := repo.Create(context.Background(), repository.CreateAnimeLinkParams{
		AnimeID:  animeID,
		Kind:     model.AnimeLinkKindOfficialSite,
		Language: model.LanguageJa,
		URL:      "https://example.dev/old",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.Update(context.Background(), repository.UpdateAnimeLinkParams{
		ID:  created.ID,
		URL: "https://example.dev/new",
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
	if got[0].URL != "https://example.dev/new" {
		t.Errorf("URL = %q, want https://example.dev/new after update", got[0].URL)
	}
}

func TestAnimeLinkRepository_Delete(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeLinkRepository(queries)

	animeID := createTestAnime(t, animeRepo, "リンク削除アニメ")

	created, err := repo.Create(context.Background(), repository.CreateAnimeLinkParams{
		AnimeID:  animeID,
		Kind:     model.AnimeLinkKindWikipedia,
		Language: model.LanguageJa,
		URL:      "https://ja.wikipedia.org/wiki/Example",
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

func TestAnimeLinkRepository_ListByAnimeIDs_EmptyInput(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeLinkRepository(query.New(db).WithTx(tx))

	got, err := repo.ListByAnimeIDs(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListByAnimeIDs() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("len(got) = %d, want 0 for empty input", len(got))
	}
}

// animeLinkTestKey keys an anime's links by (kind, language) for assertions.
//
// [Ja] animeLinkTestKey はアサーション用に anime のリンクを (kind, language) でキーにする。
type animeLinkTestKey struct {
	kind     model.AnimeLinkKind
	language model.Language
}
