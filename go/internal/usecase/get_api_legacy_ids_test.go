package usecase

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

func TestGetAPILegacyIDsUsecase_Execute(t *testing.T) {
	t.Parallel()

	f, _ := newAPIAnimeFixture(t)
	ctx := context.Background()
	uc := NewGetAPILegacyIDsUsecase(
		validator.NewAPILegacyIDIndexValidator(),
		f.animeRepo,
		f.classificationRep,
		repository.NewAnimeRedirectRepository(f.queries),
		repository.NewAnimeLegacyIDRepository(f.queries),
	)
	workRepo := repository.NewWorkRepository(f.queries)
	episodeRepo := repository.NewEpisodeRepository(f.queries)

	// legacyWorkは、旧作品を作り、指定したanimeに写す
	legacyWork := func(animeID model.AnimeID) model.WorkID {
		t.Helper()
		id := testutil.NewWorkBuilder(t, f.tx).WithTitle("旧作品").WithoutAnime().Build()
		if err := workRepo.UpdateAnimeID(ctx, id, animeID); err != nil {
			t.Fatalf("旧作品のanime_idの設定に失敗しました: %v", err)
		}
		return id
	}
	legacyEpisode := func(workID model.WorkID, sortNumber int32, animeID model.AnimeID) model.EpisodeID {
		t.Helper()
		id := testutil.NewEpisodeBuilder(t, f.tx, workID).WithSortNumber(sortNumber).Build()
		if err := episodeRepo.UpdateAnimeID(ctx, id, animeID); err != nil {
			t.Fatalf("旧エピソードのanime_idの設定に失敗しました: %v", err)
		}
		return id
	}

	published := f.createWork("公開中の作品", model.AnimeStatusPublished)
	merged := f.createWork("統合した作品", model.AnimeStatusMerged)
	f.redirect(merged, published)
	archived := f.createWork("非公開の作品", model.AnimeStatusArchived)
	episode := f.createEpisode(published, 1, model.AnimeStatusPublished)
	episodeOfArchived := f.createEpisode(archived, 1, model.AnimeStatusPublished)

	publishedWorkID := legacyWork(published)
	mergedWorkID := legacyWork(merged)
	archivedWorkID := legacyWork(archived)
	unmappedWorkID := testutil.NewWorkBuilder(t, f.tx).WithTitle("写していない旧作品").WithoutAnime().Build()
	episodeID := legacyEpisode(publishedWorkID, 1, episode)
	episodeOfArchivedID := legacyEpisode(archivedWorkID, 1, episodeOfArchived)

	// サブテストは親のトランザクションを共有するため、並列にしない

	t.Run("旧IDごとに、取得できるanimeのIDを要求の順に返す", func(t *testing.T) {
		output, err := uc.Execute(ctx, GetAPILegacyIDsInput{
			WorkIDs:    []model.WorkID{archivedWorkID, publishedWorkID, mergedWorkID, unmappedWorkID, model.WorkID(0), publishedWorkID},
			EpisodeIDs: []model.EpisodeID{episodeID, episodeOfArchivedID},
		})
		if err != nil {
			t.Fatalf("Execute()のエラー = %v", err)
		}

		wantWorks := []struct {
			id      model.WorkID
			animeID *model.AnimeID
		}{
			{id: archivedWorkID, animeID: nil},
			{id: publishedWorkID, animeID: &published},
			// 統合したanimeの旧IDは、統合先を返す
			{id: mergedWorkID, animeID: &published},
			{id: unmappedWorkID, animeID: nil},
			{id: model.WorkID(0), animeID: nil},
		}
		if len(output.Works) != len(wantWorks) {
			t.Fatalf("作品の結果 = %+v、期待値 = 重複をまとめた%d件", output.Works, len(wantWorks))
		}
		for i, want := range wantWorks {
			got := output.Works[i]
			if got.WorkID != want.id || !equalAnimeID(got.AnimeID, want.animeID) {
				t.Errorf("作品の結果[%d] = %d → %v、期待値 = %d → %v", i, got.WorkID, got.AnimeID, want.id, want.animeID)
			}
		}

		if len(output.Episodes) != 2 {
			t.Fatalf("エピソードの結果 = %+v、期待値 = 2件", output.Episodes)
		}
		if output.Episodes[0].EpisodeID != episodeID || !equalAnimeID(output.Episodes[0].AnimeID, &episode) {
			t.Errorf("公開中の作品のエピソードの結果 = %+v、期待値 = anime %d", output.Episodes[0], episode)
		}
		// 親の作品が公開中でないエピソードは、getAnimeと同じく返せない
		if output.Episodes[1].EpisodeID != episodeOfArchivedID || output.Episodes[1].AnimeID != nil {
			t.Errorf("非公開の作品のエピソードの結果 = %+v、期待値 = anime_idがnil", output.Episodes[1])
		}
	})

	t.Run("指定しなかった種類は空で返す", func(t *testing.T) {
		output, err := uc.Execute(ctx, GetAPILegacyIDsInput{EpisodeIDs: []model.EpisodeID{episodeID}})
		if err != nil {
			t.Fatalf("Execute()のエラー = %v", err)
		}
		if output.Works == nil || len(output.Works) != 0 || len(output.Episodes) != 1 {
			t.Errorf("結果 = %+v、期待値 = 作品は空の配列でエピソードは1件", output)
		}
	})

	t.Run("旧IDが1つも無いとValidationError", func(t *testing.T) {
		_, err := uc.Execute(ctx, GetAPILegacyIDsInput{})
		if model.AsValidationError(err) == nil {
			t.Errorf("エラー = %v、期待値 = ValidationError", err)
		}
	})
}

func equalAnimeID(got, want *model.AnimeID) bool {
	if got == nil || want == nil {
		return got == want
	}
	return *got == *want
}
