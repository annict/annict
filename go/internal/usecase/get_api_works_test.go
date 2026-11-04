package usecase

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

func TestGetAPIWorksUsecase_Execute(t *testing.T) {
	t.Parallel()

	f, _ := newAPIAnimeFixture(t)
	ctx := context.Background()
	uc := NewGetAPIWorksUsecase(f.animeRepo, f.classificationRep, repository.NewAnimeLegacyIDRepository(f.queries), f.satelliteRepos)

	addSeason := func(id model.AnimeID, year int32, name model.SeasonName, primary bool) {
		t.Helper()
		if _, err := f.satelliteRepos.Season.Create(ctx, repository.CreateAnimeSeasonParams{AnimeID: id, Year: year, Name: &name, IsPrimary: primary}); err != nil {
			t.Fatalf("季節の作成に失敗しました: %v", err)
		}
	}
	// 年を4001年以降にして、ほかのテストのデータと重ならないようにする
	springOnly := f.createWork("春の作品", model.AnimeStatusPublished)
	addSeason(springOnly, 4001, model.SeasonNameSpring, true)
	both := f.createWork("春と翌年の冬の作品", model.AnimeStatusPublished)
	addSeason(both, 4002, model.SeasonNameWinter, true)
	addSeason(both, 4001, model.SeasonNameSpring, false)
	winterOnly := f.createWork("翌年の冬の作品", model.AnimeStatusPublished)
	addSeason(winterOnly, 4002, model.SeasonNameWinter, true)
	episode := f.createEpisode(both, 1, model.AnimeStatusPublished)

	year := int32(4002)
	tests := []struct {
		name  string
		input GetAPIWorksInput
		want  []model.AnimeID
	}{
		{
			name:  "年と季節は別の行に一致してもよく、すべてを満たす作品に絞る",
			input: GetAPIWorksInput{Year: &year, Seasons: []model.SeasonBucket{{Year: 4001, Name: ptrSeasonName(model.SeasonNameSpring)}}},
			want:  []model.AnimeID{both},
		},
		{
			name:  "idsと季節の両方で絞る",
			input: GetAPIWorksInput{IDs: []model.AnimeID{springOnly, winterOnly, episode}, Seasons: []model.SeasonBucket{{Year: 4001, Name: ptrSeasonName(model.SeasonNameSpring)}, {Year: 4002, Name: ptrSeasonName(model.SeasonNameWinter)}}},
			want:  []model.AnimeID{springOnly, winterOnly},
		},
		{
			name:  "一致する季節が無ければ空",
			input: GetAPIWorksInput{Seasons: []model.SeasonBucket{{Year: 4009}}},
			want:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.Order = APIWorkOrderID
			tt.input.Limit = 10
			output, err := uc.Execute(ctx, tt.input)
			if err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}
			var got []model.AnimeID
			for _, w := range output.Works {
				got = append(got, w.Anime.ID)
			}
			assertAPIAnimeIDs(t, got, tt.want)
			if output.Next != nil {
				t.Errorf("Next = %+v、期待値 = nil", output.Next)
			}
		})
	}

	t.Run("続きがあればNextに最後の作品の季節のキーを入れる", func(t *testing.T) {
		input := GetAPIWorksInput{IDs: []model.AnimeID{springOnly, both, winterOnly}, Order: APIWorkOrderSeason, Limit: 1}
		var got []model.AnimeID
		for range 4 {
			output, err := uc.Execute(ctx, input)
			if err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}
			for _, w := range output.Works {
				got = append(got, w.Anime.ID)
			}
			if output.Next == nil {
				break
			}
			if output.Next.SeasonKey == nil {
				t.Fatalf("Next = %+v、期待値 = 季節のキーのあるカーソル", output.Next)
			}
			input.After = output.Next
		}
		assertAPIAnimeIDs(t, got, []model.AnimeID{springOnly, both, winterOnly})
	})
}

func ptrSeasonName(n model.SeasonName) *model.SeasonName { return &n }

func assertAPIAnimeIDs(t *testing.T, got, want []model.AnimeID) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("anime = %v、期待値 = %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("anime = %v、期待値 = %v", got, want)
		}
	}
}
