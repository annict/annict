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

	// (anime_id, hashtag) はユニークなので、1つのanimeはタグ値ごとに高々1行を持つ。
	// 異なる2つのタグを作成する。
	for _, tag := range []string{"rezero", "rezero2nd"} {
		if _, err := repo.Create(context.Background(), repository.CreateAnimeHashtagParams{
			AnimeID: animeID,
			Hashtag: tag,
		}); err != nil {
			t.Fatalf("Create(%s)のエラー = %v", tag, err)
		}
	}

	got, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs()のエラー = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d、期待値 = 2", len(got))
	}

	tags := map[string]bool{}
	for _, h := range got {
		if h.ID == 0 {
			t.Error("IDが採番されていない")
		}
		if h.AnimeID != animeID {
			t.Errorf("AnimeID = %d、期待値 = %d", h.AnimeID, animeID)
		}
		// worksはsort_numberをsourceしないため、作成直後の行はDB既定値の0のまま。
		if h.SortNumber != 0 {
			t.Errorf("SortNumber = %d、期待値 = 0", h.SortNumber)
		}
		tags[h.Hashtag] = true
	}
	if !tags["rezero"] || !tags["rezero2nd"] {
		t.Errorf("tags = %v、期待値 = rezeroとrezero2nd", tags)
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

func TestAnimeHashtagRepository_ListByAnimeIDs_EmptyInput(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeHashtagRepository(query.New(db).WithTx(tx))

	got, err := repo.ListByAnimeIDs(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListByAnimeIDs()のエラー = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("空入力時のlen(got) = %d、期待値 = 0", len(got))
	}
}
