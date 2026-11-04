package usecase

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

func TestGetAPIEpisodesUsecase_Execute(t *testing.T) {
	t.Parallel()

	f, _ := newAPIAnimeFixture(t)
	ctx := context.Background()
	uc := NewGetAPIEpisodesUsecase(f.animeRepo, f.classificationRep, repository.NewAnimeLegacyIDRepository(f.queries), f.satelliteRepos)

	workID := f.createWork("作品", model.AnimeStatusPublished)
	first := f.createEpisode(workID, 1, model.AnimeStatusPublished)
	f.createEpisode(workID, 2, model.AnimeStatusArchived)
	third := f.createEpisode(workID, 3, model.AnimeStatusPublished)

	t.Run("公開中のエピソードを順に返し、続きがあればNextに並び順のキーを入れる", func(t *testing.T) {
		output, err := uc.Execute(ctx, GetAPIEpisodesInput{WorkID: workID, Limit: 1})
		if err != nil {
			t.Fatalf("Execute()のエラー = %v", err)
		}
		if len(output.Episodes) != 1 || output.Episodes[0].Anime.ID != first {
			t.Fatalf("1ページめ = %+v、期待値 = %d", output.Episodes, first)
		}
		if output.Next == nil || output.Next.SortNumber != 1 || output.Next.AnimeID != first {
			t.Fatalf("Next = %+v、期待値 = 1話めを指すカーソル", output.Next)
		}
		// 次のエピソードは、ページの外でも前後として返す
		if output.Episodes[0].NextEpisode == nil || output.Episodes[0].NextEpisode.ID != third {
			t.Errorf("次のエピソード = %+v、期待値 = %d", output.Episodes[0].NextEpisode, third)
		}

		output, err = uc.Execute(ctx, GetAPIEpisodesInput{WorkID: workID, After: output.Next, Limit: 1})
		if err != nil {
			t.Fatalf("Execute()のエラー = %v", err)
		}
		if len(output.Episodes) != 1 || output.Episodes[0].Anime.ID != third || output.Next != nil {
			t.Errorf("2ページめ = %+v、Next = %+v、期待値 = %d で続き無し", output.Episodes, output.Next, third)
		}
	})

	archivedWork := f.createWork("非公開の作品", model.AnimeStatusArchived)
	mergedWork := f.createWork("統合した作品", model.AnimeStatusMerged)
	f.redirect(mergedWork, workID)
	notFoundTests := []struct {
		name string
		id   model.AnimeID
	}{
		{name: "エピソードのIDは404", id: first},
		{name: "非公開の作品は404", id: archivedWork},
		{name: "統合した作品は404", id: mergedWork},
		{name: "存在しないIDは404", id: model.AnimeID(0)},
	}
	for _, tt := range notFoundTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.Execute(ctx, GetAPIEpisodesInput{WorkID: tt.id, Limit: 10})
			if ae := model.AsAppError(err); ae == nil || ae.Code != model.AppErrCodeResourceNotFound {
				t.Errorf("エラー = %v、期待値 = 未存在のAppError", err)
			}
		})
	}
}
