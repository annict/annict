package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestAnimeLegacyIDRepository_ListByAnimeIDs(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	ctx := context.Background()
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeLegacyIDRepository(queries)

	workAnimeID := createTestAnime(t, animeRepo, "旧作品から写した作品")
	episodeAnimeID := createTestAnime(t, animeRepo, "旧エピソードから写したエピソード")
	newAnimeID := createTestAnime(t, animeRepo, "旧IDの無い作品")

	workID := testutil.NewWorkBuilder(t, tx).WithTitle("旧作品").Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).Build()
	if err := repository.NewWorkRepository(queries).UpdateAnimeID(ctx, workID, workAnimeID); err != nil {
		t.Fatalf("旧作品のanime_idの設定に失敗しました: %v", err)
	}
	if err := repository.NewEpisodeRepository(queries).UpdateAnimeID(ctx, episodeID, episodeAnimeID); err != nil {
		t.Fatalf("旧エピソードのanime_idの設定に失敗しました: %v", err)
	}

	legacyIDs, err := repo.ListByAnimeIDs(ctx, []model.AnimeID{workAnimeID, episodeAnimeID, newAnimeID})
	if err != nil {
		t.Fatalf("ListByAnimeIDs()のエラー = %v", err)
	}

	byAnimeID := make(map[model.AnimeID]*model.AnimeLegacyID, len(legacyIDs))
	for _, l := range legacyIDs {
		byAnimeID[l.AnimeID] = l
	}
	if l := byAnimeID[workAnimeID]; l == nil || l.WorkID == nil || *l.WorkID != workID || l.EpisodeID != nil {
		t.Errorf("作品の旧ID = %+v、期待値 = 旧作品 %d だけ", l, workID)
	}
	if l := byAnimeID[episodeAnimeID]; l == nil || l.EpisodeID == nil || *l.EpisodeID != episodeID || l.WorkID != nil {
		t.Errorf("エピソードの旧ID = %+v、期待値 = 旧エピソード %d だけ", l, episodeID)
	}
	if l, ok := byAnimeID[newAnimeID]; ok {
		t.Errorf("旧IDの無いanimeの結果 = %+v、期待値 = 含まない", l)
	}

	empty, err := repo.ListByAnimeIDs(ctx, nil)
	if err != nil || len(empty) != 0 {
		t.Errorf("空入力の結果 = %v、エラー = %v、期待値 = 空", empty, err)
	}
}

func TestAnimeLegacyIDRepository_ListByWorkIDsAndEpisodeIDs(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	ctx := context.Background()
	queries := query.New(db).WithTx(tx)
	animeRepo := repository.NewAnimeRepository(queries)
	repo := repository.NewAnimeLegacyIDRepository(queries)

	workAnimeID := createTestAnime(t, animeRepo, "旧作品から写した作品")
	episodeAnimeID := createTestAnime(t, animeRepo, "旧エピソードから写したエピソード")
	workID := testutil.NewWorkBuilder(t, tx).WithTitle("旧作品").Build()
	unmappedWorkID := testutil.NewWorkBuilder(t, tx).WithTitle("写していない旧作品").WithoutAnime().Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).Build()
	unmappedEpisodeID := testutil.NewEpisodeBuilder(t, tx, workID).WithSortNumber(2).Build()
	if err := repository.NewWorkRepository(queries).UpdateAnimeID(ctx, workID, workAnimeID); err != nil {
		t.Fatalf("旧作品のanime_idの設定に失敗しました: %v", err)
	}
	if err := repository.NewEpisodeRepository(queries).UpdateAnimeID(ctx, episodeID, episodeAnimeID); err != nil {
		t.Fatalf("旧エピソードのanime_idの設定に失敗しました: %v", err)
	}

	t.Run("写した旧作品だけを返す", func(t *testing.T) {
		legacyIDs, err := repo.ListByWorkIDs(ctx, []model.WorkID{workID, unmappedWorkID, model.WorkID(0)})
		if err != nil {
			t.Fatalf("ListByWorkIDs()のエラー = %v", err)
		}
		if len(legacyIDs) != 1 || *legacyIDs[0].WorkID != workID || legacyIDs[0].AnimeID != workAnimeID || legacyIDs[0].EpisodeID != nil {
			t.Errorf("結果 = %+v、期待値 = 旧作品 %d とanime %d の組だけ", legacyIDs, workID, workAnimeID)
		}
	})

	t.Run("写した旧エピソードだけを返す", func(t *testing.T) {
		legacyIDs, err := repo.ListByEpisodeIDs(ctx, []model.EpisodeID{episodeID, unmappedEpisodeID})
		if err != nil {
			t.Fatalf("ListByEpisodeIDs()のエラー = %v", err)
		}
		if len(legacyIDs) != 1 || *legacyIDs[0].EpisodeID != episodeID || legacyIDs[0].AnimeID != episodeAnimeID || legacyIDs[0].WorkID != nil {
			t.Errorf("結果 = %+v、期待値 = 旧エピソード %d とanime %d の組だけ", legacyIDs, episodeID, episodeAnimeID)
		}
	})

	t.Run("空入力では空を返す", func(t *testing.T) {
		works, err1 := repo.ListByWorkIDs(ctx, nil)
		episodes, err2 := repo.ListByEpisodeIDs(ctx, nil)
		if err1 != nil || err2 != nil || len(works) != 0 || len(episodes) != 0 {
			t.Errorf("結果 = %v / %v、エラー = %v / %v、期待値 = 空", works, episodes, err1, err2)
		}
	})
}
