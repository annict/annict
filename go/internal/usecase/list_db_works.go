package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// ListDbWorksUsecase はDB管理画面の作品一覧を取得するユースケースです
type ListDbWorksUsecase struct {
	workRepo *repository.WorkRepository
}

// NewListDbWorksUsecase は新しいListDbWorksUsecaseを作成します
func NewListDbWorksUsecase(workRepo *repository.WorkRepository) *ListDbWorksUsecase {
	return &ListDbWorksUsecase{
		workRepo: workRepo,
	}
}

// ListDbWorksInput はユースケースの入力です
type ListDbWorksInput struct {
	FilterNoEpisodes bool
	FilterNoImage    bool
	FilterNoSeason   bool
	Page             int32
	PerPage          int32
}

// ListDbWorksOutput はユースケースの出力です
type ListDbWorksOutput struct {
	Works      []*model.Work
	TotalCount int64
}

// Execute はDB管理画面の作品一覧と総数を取得します
func (uc *ListDbWorksUsecase) Execute(ctx context.Context, input ListDbWorksInput) (*ListDbWorksOutput, error) {
	params := repository.DBWorkListParams{
		FilterNoEpisodes: input.FilterNoEpisodes,
		FilterNoImage:    input.FilterNoImage,
		FilterNoSeason:   input.FilterNoSeason,
		Page:             input.Page,
		PerPage:          input.PerPage,
	}

	works, err := uc.workRepo.ListForDB(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("DB作品一覧の取得に失敗: %w", err)
	}

	totalCount, err := uc.workRepo.CountForDB(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("DB作品総数の取得に失敗: %w", err)
	}

	return &ListDbWorksOutput{
		Works:      works,
		TotalCount: totalCount,
	}, nil
}
