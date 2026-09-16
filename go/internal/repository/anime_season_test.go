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

	// 名前付きのworks管理下の主季節1つと、別の年の名前なし行を作る。is_primaryが
	// 異なるので (anime_id) WHERE is_primaryの部分UNIQUEインデックスを満たし、2つの
	// (year, name) キーも異なる。
	spring := model.SeasonNameSpring
	if _, err := repo.Create(context.Background(), repository.CreateAnimeSeasonParams{
		AnimeID: animeID, Year: 2024, Name: &spring, IsPrimary: true,
	}); err != nil {
		t.Fatalf("Nameがある場合のCreate()のエラー = %v", err)
	}
	if _, err := repo.Create(context.Background(), repository.CreateAnimeSeasonParams{
		AnimeID: animeID, Year: 2023, Name: nil, IsPrimary: false,
	}); err != nil {
		t.Fatalf("Nameがnilの場合のCreate()のエラー = %v", err)
	}

	got, err := repo.ListByAnimeIDs(context.Background(), []model.AnimeID{animeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs()のエラー = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d、期待値 = 2", len(got))
	}

	byYear := map[int32]*model.AnimeSeason{}
	for _, s := range got {
		if s.ID == 0 {
			t.Error("IDが採番されていない")
		}
		if s.AnimeID != animeID {
			t.Errorf("AnimeID = %d、期待値 = %d", s.AnimeID, animeID)
		}
		byYear[s.Year] = s
	}

	named := byYear[2024]
	if named == nil || named.Name == nil || *named.Name != model.SeasonNameSpring || !named.IsPrimary {
		t.Errorf("2024の行 = %+v、期待値 = spring / is_primaryがtrue", named)
	}
	nameless := byYear[2023]
	if nameless == nil || nameless.Name != nil || nameless.IsPrimary {
		t.Errorf("2023の行 = %+v、期待値 = nameがNULL / is_primaryがfalse", nameless)
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

func TestAnimeSeasonRepository_ListByAnimeIDs_EmptyInput(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewAnimeSeasonRepository(query.New(db).WithTx(tx))

	got, err := repo.ListByAnimeIDs(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListByAnimeIDs()のエラー = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("空入力時のlen(got) = %d、期待値 = 0", len(got))
	}
}
