package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// APIEpisodeCursorは、エピソードの一覧で前のページの最後のエピソードを指す
type APIEpisodeCursor struct {
	SortNumber int32
	AnimeID    model.AnimeID
}

// GetAPIEpisodesUsecaseは、REST API V2で作品のエピソードの一覧を取得するユースケース
type GetAPIEpisodesUsecase struct {
	animeRepo               *repository.AnimeRepository
	animeClassificationRepo *repository.AnimeClassificationRepository
	loader                  *apiAnimeLoader
}

func NewGetAPIEpisodesUsecase(
	animeRepo *repository.AnimeRepository,
	animeClassificationRepo *repository.AnimeClassificationRepository,
	animeLegacyIDRepo *repository.AnimeLegacyIDRepository,
	satelliteRepos WorkSatelliteRepos,
) *GetAPIEpisodesUsecase {
	return &GetAPIEpisodesUsecase{
		animeRepo:               animeRepo,
		animeClassificationRepo: animeClassificationRepo,
		loader: &apiAnimeLoader{
			animeRepo:               animeRepo,
			animeClassificationRepo: animeClassificationRepo,
			animeLegacyIDRepo:       animeLegacyIDRepo,
			satelliteRepos:          satelliteRepos,
		},
	}
}

// GetAPIEpisodesInputはエピソードの一覧の条件
type GetAPIEpisodesInput struct {
	// WorkIDはエピソードの親の作品
	WorkID model.AnimeID
	After  *APIEpisodeCursor
	Limit  int32
}

// GetAPIEpisodesOutputはエピソードの一覧の結果
type GetAPIEpisodesOutput struct {
	Episodes []*APIAnime
	// Nextは次のページを取得するカーソル。続きが無ければnil
	Next *APIEpisodeCursor
}

// Executeは、作品の公開中のエピソードを (sort_number, anime_id) の順に最大Limit件返す。
// 作品が存在しない・公開中でない・エピソードである場合は、未存在の `*model.AppError` を返す
func (uc *GetAPIEpisodesUsecase) Execute(ctx context.Context, input GetAPIEpisodesInput) (*GetAPIEpisodesOutput, error) {
	isWork, err := uc.isPublishedWork(ctx, input.WorkID)
	if err != nil {
		return nil, err
	}
	if !isWork {
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_not_found_message"),
			Metadata: map[string]string{"anime_id": input.WorkID.String()},
		}
	}

	var after *repository.AfterEpisode
	if input.After != nil {
		after = &repository.AfterEpisode{SortNumber: input.After.SortNumber, AnimeID: input.After.AnimeID}
	}
	// 続きがあるかを知るため、1件多く引く
	animes, err := uc.animeRepo.ListAPIEpisodes(ctx, input.WorkID, after, input.Limit+1)
	if err != nil {
		return nil, fmt.Errorf("エピソードの一覧の取得に失敗: %w", err)
	}

	hasNext := len(animes) > int(input.Limit)
	if hasNext {
		animes = animes[:input.Limit]
	}
	episodes, err := uc.loader.load(ctx, animes)
	if err != nil {
		return nil, err
	}

	output := &GetAPIEpisodesOutput{Episodes: episodes}
	if hasNext {
		last := episodes[len(episodes)-1]
		output.Next = &APIEpisodeCursor{SortNumber: last.Classification.SortNumber.Int32, AnimeID: last.Anime.ID}
	}
	return output, nil
}

// isPublishedWorkは、IDが公開中の作品を指すかを返す
func (uc *GetAPIEpisodesUsecase) isPublishedWork(ctx context.Context, id model.AnimeID) (bool, error) {
	anime, err := uc.animeRepo.GetByID(ctx, id)
	if err != nil {
		return false, fmt.Errorf("作品の取得に失敗: %w", err)
	}
	if anime == nil || anime.Status != model.AnimeStatusPublished {
		return false, nil
	}

	classification, err := uc.animeClassificationRepo.GetByAnimeID(ctx, id)
	if err != nil {
		return false, fmt.Errorf("作品の分類の取得に失敗: %w", err)
	}
	return classification != nil && classification.Kind == model.AnimeClassificationKindWork, nil
}
