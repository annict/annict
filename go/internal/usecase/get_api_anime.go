package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetAPIAnimeUsecaseは、REST API V2で作品かエピソード1件を取得するユースケース
type GetAPIAnimeUsecase struct {
	animeRepo         *repository.AnimeRepository
	animeRedirectRepo *repository.AnimeRedirectRepository
	loader            *apiAnimeLoader
}

func NewGetAPIAnimeUsecase(
	animeRepo *repository.AnimeRepository,
	animeClassificationRepo *repository.AnimeClassificationRepository,
	animeRedirectRepo *repository.AnimeRedirectRepository,
	animeLegacyIDRepo *repository.AnimeLegacyIDRepository,
	satelliteRepos WorkSatelliteRepos,
) *GetAPIAnimeUsecase {
	return &GetAPIAnimeUsecase{
		animeRepo:         animeRepo,
		animeRedirectRepo: animeRedirectRepo,
		loader: &apiAnimeLoader{
			animeRepo:               animeRepo,
			animeClassificationRepo: animeClassificationRepo,
			animeLegacyIDRepo:       animeLegacyIDRepo,
			satelliteRepos:          satelliteRepos,
		},
	}
}

// GetAPIAnimeInputは作品かエピソードの取得に必要な入力パラメータ
type GetAPIAnimeInput struct {
	AnimeID model.AnimeID
}

// GetAPIAnimeOutputは作品かエピソードの取得結果
type GetAPIAnimeOutput struct {
	Anime *APIAnime
}

// Executeは、指定したIDの公開中の作品かエピソードを返す。
// 統合したanimeのIDを指定したときは、統合先を返す。
// 存在しない・非公開・削除済みのanimeは、未存在の `*model.AppError` を返す
func (uc *GetAPIAnimeUsecase) Execute(ctx context.Context, input GetAPIAnimeInput) (*GetAPIAnimeOutput, error) {
	anime, err := uc.findPublishedAnime(ctx, input.AnimeID)
	if err != nil {
		return nil, err
	}
	if anime == nil {
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_not_found_message"),
			Metadata: map[string]string{"anime_id": input.AnimeID.String()},
		}
	}

	results, err := uc.loader.load(ctx, []*model.Anime{anime})
	if err != nil {
		return nil, err
	}
	return &GetAPIAnimeOutput{Anime: results[0]}, nil
}

// findPublishedAnimeは、統合をたどったうえで公開中のanimeを返す。返せるanimeが無ければnilを返す
func (uc *GetAPIAnimeUsecase) findPublishedAnime(ctx context.Context, id model.AnimeID) (*model.Anime, error) {
	anime, err := uc.animeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("animeの取得に失敗: %w", err)
	}

	if anime != nil && anime.Status == model.AnimeStatusMerged {
		canonicalID, err := uc.animeRedirectRepo.FindCanonicalAnimeID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("統合先の取得に失敗: %w", err)
		}
		if canonicalID == nil {
			return nil, nil
		}
		// 統合の記録は常に最終の統合先を指すため、1回たどれば足りる
		anime, err = uc.animeRepo.GetByID(ctx, *canonicalID)
		if err != nil {
			return nil, fmt.Errorf("統合先のanimeの取得に失敗: %w", err)
		}
	}

	if anime == nil || anime.Status != model.AnimeStatusPublished {
		return nil, nil
	}
	return anime, nil
}
