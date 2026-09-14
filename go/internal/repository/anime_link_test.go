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

	// 1つのanimeは (kind, language) ごとに1リンクを持つ。公式サイト (ja) と
	// Wikipedia (en) を作成する。
	if _, err := repo.Create(context.Background(), repository.CreateAnimeLinkParams{
		AnimeID:  animeID,
		Kind:     model.AnimeLinkKindOfficialSite,
		Language: model.LanguageJa,
		URL:      "https://example.dev/official",
	}); err != nil {
		t.Fatalf("Create(official_site, ja)のエラー = %v", err)
	}
	if _, err := repo.Create(context.Background(), repository.CreateAnimeLinkParams{
		AnimeID:  animeID,
		Kind:     model.AnimeLinkKindWikipedia,
		Language: model.LanguageEn,
		URL:      "https://en.wikipedia.org/wiki/Example",
	}); err != nil {
		t.Fatalf("Create(wikipedia, en)のエラー = %v", err)
	}

	got, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs()のエラー = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d、期待値 = 2", len(got))
	}

	byKey := map[animeLinkTestKey]*model.AnimeLink{}
	for _, l := range got {
		if l.ID == 0 {
			t.Error("IDが採番されていない")
		}
		if l.AnimeID != animeID {
			t.Errorf("AnimeID = %d、期待値 = %d", l.AnimeID, animeID)
		}
		// worksはlabelをsourceしないため、作成した行ではnilのまま。
		if l.Label != nil || l.LabelEn != nil {
			t.Errorf("Label/LabelEn = %v/%v、期待値 = nil/nil", l.Label, l.LabelEn)
		}
		byKey[animeLinkTestKey{l.Kind, l.Language}] = l
	}

	if got := byKey[animeLinkTestKey{model.AnimeLinkKindOfficialSite, model.LanguageJa}]; got == nil || got.URL != "https://example.dev/official" {
		t.Errorf("official_site/ja = %+v、期待値 = url https://example.dev/official", got)
	}
	if got := byKey[animeLinkTestKey{model.AnimeLinkKindWikipedia, model.LanguageEn}]; got == nil || got.URL != "https://en.wikipedia.org/wiki/Example" {
		t.Errorf("wikipedia/en = %+v、期待値 = url https://en.wikipedia.org/wiki/Example", got)
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
		t.Fatalf("Create()のエラー = %v", err)
	}

	if err := repo.Update(context.Background(), repository.UpdateAnimeLinkParams{
		ID:  created.ID,
		URL: "https://example.dev/new",
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
	if got[0].URL != "https://example.dev/new" {
		t.Errorf("更新後のURL = %q、期待値 = https://example.dev/new", got[0].URL)
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

func TestAnimeLinkRepository_ListByAnimeIDs_EmptyInput(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeLinkRepository(query.New(db).WithTx(tx))

	got, err := repo.ListByAnimeIDs(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListByAnimeIDs()のエラー = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("空入力時のlen(got) = %d、期待値 = 0", len(got))
	}
}

// animeLinkTestKeyはアサーション用にanimeのリンクを (kind, language) でキーにする。
type animeLinkTestKey struct {
	kind     model.AnimeLinkKind
	language model.Language
}
