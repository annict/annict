package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// APIWorkOrderは、REST API V2の作品の一覧の並び順
type APIWorkOrder string

const (
	APIWorkOrderID         APIWorkOrder = "id"
	APIWorkOrderIDDesc     APIWorkOrder = "-id"
	APIWorkOrderSeason     APIWorkOrder = "season"
	APIWorkOrderSeasonDesc APIWorkOrder = "-season"
)

// APIWorkCursorは、作品の一覧で前のページの最後の作品を指す
type APIWorkCursor struct {
	AnimeID model.AnimeID
	// SeasonKeyは代表の季節のキー (model.AnimeSeason.SortKey)。代表の季節の無い作品ではnil
	SeasonKey *int32
}

// GetAPIWorksUsecaseは、REST API V2で作品の一覧を取得するユースケース
type GetAPIWorksUsecase struct {
	animeRepo       *repository.AnimeRepository
	animeSeasonRepo *repository.AnimeSeasonRepository
	loader          *apiAnimeLoader
}

func NewGetAPIWorksUsecase(
	animeRepo *repository.AnimeRepository,
	animeClassificationRepo *repository.AnimeClassificationRepository,
	animeLegacyIDRepo *repository.AnimeLegacyIDRepository,
	satelliteRepos WorkSatelliteRepos,
) *GetAPIWorksUsecase {
	return &GetAPIWorksUsecase{
		animeRepo:       animeRepo,
		animeSeasonRepo: satelliteRepos.Season,
		loader: &apiAnimeLoader{
			animeRepo:               animeRepo,
			animeClassificationRepo: animeClassificationRepo,
			animeLegacyIDRepo:       animeLegacyIDRepo,
			satelliteRepos:          satelliteRepos,
		},
	}
}

// GetAPIWorksInputは作品の一覧の条件。絞り込みはnilなら使わず、指定したものをすべて満たす作品を返す
type GetAPIWorksInput struct {
	IDs []model.AnimeID
	// Seasonsは、どれかの季節に放送・公開した作品に絞る
	Seasons []model.SeasonBucket
	// Yearは、その年のどれかの季節に放送・公開した作品に絞る
	Year  *int32
	Title *string
	Order APIWorkOrder
	After *APIWorkCursor
	Limit int32
}

// GetAPIWorksOutputは作品の一覧の結果
type GetAPIWorksOutput struct {
	Works []*APIAnime
	// Nextは次のページを取得するカーソル。続きが無ければnil
	Next *APIWorkCursor
}

// Executeは、公開中の作品を条件で絞り込み、並び順のとおりに最大Limit件返す
func (uc *GetAPIWorksUsecase) Execute(ctx context.Context, input GetAPIWorksInput) (*GetAPIWorksOutput, error) {
	candidates, err := uc.candidateIDs(ctx, input)
	if err != nil {
		return nil, err
	}

	params := repository.ListAPIWorksParams{
		IDs:   candidates,
		Title: input.Title,
		Order: repository.APIWorkOrder(input.Order),
		// 続きがあるかを知るため、1件多く引く
		Limit: input.Limit + 1,
	}
	if input.After != nil {
		params.AfterID = &input.After.AnimeID
		params.AfterSeasonKey = input.After.SeasonKey
	}
	animes, err := uc.animeRepo.ListAPIWorks(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("作品の一覧の取得に失敗: %w", err)
	}

	hasNext := len(animes) > int(input.Limit)
	if hasNext {
		animes = animes[:input.Limit]
	}
	works, err := uc.loader.load(ctx, animes)
	if err != nil {
		return nil, err
	}

	output := &GetAPIWorksOutput{Works: works}
	if hasNext {
		output.Next = newAPIWorkCursor(works[len(works)-1])
	}
	return output, nil
}

// candidateIDsは、IDと年・季節の絞り込みから候補の作品のIDを求める。どれも指定されなければnilを返す。
// 年・季節の絞り込みは代表でない季節にも一致させるため、季節の行から先に候補を引く
func (uc *GetAPIWorksUsecase) candidateIDs(ctx context.Context, input GetAPIWorksInput) ([]model.AnimeID, error) {
	candidates := input.IDs
	if input.Year != nil {
		ids, err := uc.animeSeasonRepo.ListAnimeIDsBySeasons(ctx, input.Year, nil)
		if err != nil {
			return nil, fmt.Errorf("年に一致する作品の取得に失敗: %w", err)
		}
		candidates = intersectAnimeIDs(candidates, ids)
	}
	if input.Seasons != nil {
		ids, err := uc.animeSeasonRepo.ListAnimeIDsBySeasons(ctx, nil, input.Seasons)
		if err != nil {
			return nil, fmt.Errorf("季節に一致する作品の取得に失敗: %w", err)
		}
		candidates = intersectAnimeIDs(candidates, ids)
	}
	return candidates, nil
}

// intersectAnimeIDsは、候補 (nilなら絞らない) をidsに含まれるものに絞る
func intersectAnimeIDs(candidates, ids []model.AnimeID) []model.AnimeID {
	if candidates == nil {
		return ids
	}
	included := make(map[model.AnimeID]bool, len(ids))
	for _, id := range ids {
		included[id] = true
	}
	result := make([]model.AnimeID, 0, len(candidates))
	for _, id := range candidates {
		if included[id] {
			result = append(result, id)
		}
	}
	return result
}

// newAPIWorkCursorは、作品の並び順のキーからカーソルを作る
func newAPIWorkCursor(work *APIAnime) *APIWorkCursor {
	cursor := &APIWorkCursor{AnimeID: work.Anime.ID}
	for _, s := range work.Seasons {
		if s.IsPrimary {
			key := s.SortKey()
			cursor.SeasonKey = &key
		}
	}
	return cursor
}
