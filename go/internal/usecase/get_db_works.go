package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBWorksUsecaseはDB管理画面の作品一覧を取得するユースケース。
type GetDBWorksUsecase struct {
	workRepo *repository.WorkRepository
}

// NewGetDBWorksUsecaseは新しいGetDBWorksUsecaseを作成する。
func NewGetDBWorksUsecase(workRepo *repository.WorkRepository) *GetDBWorksUsecase {
	return &GetDBWorksUsecase{
		workRepo: workRepo,
	}
}

// GetDBWorksInputはユースケースの入力。
type GetDBWorksInput struct {
	FilterNoEpisodes bool
	FilterNoImage    bool
	FilterNoSeason   bool
	FilterNoSlots    bool
	// SeasonYears / SeasonNamesはリリース時期の複数選択フィルタの並列 (年, 季節)
	// ペア (空でフィルタ無効)。
	SeasonYears []int32
	SeasonNames []int32
	Page        int32
	PerPage     int32
}

// GetDBWorksOutputはユースケースの出力。
type GetDBWorksOutput struct {
	Works      []*model.Work
	TotalCount int64
}

// ExecuteはDB管理画面の作品一覧と総数を取得する。
func (uc *GetDBWorksUsecase) Execute(ctx context.Context, input GetDBWorksInput) (*GetDBWorksOutput, error) {
	params := repository.DBWorkListParams{
		FilterNoEpisodes: input.FilterNoEpisodes,
		FilterNoImage:    input.FilterNoImage,
		FilterNoSeason:   input.FilterNoSeason,
		FilterNoSlots:    input.FilterNoSlots,
		SeasonYears:      input.SeasonYears,
		SeasonNames:      input.SeasonNames,
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

	return &GetDBWorksOutput{
		Works:      works,
		TotalCount: totalCount,
	}, nil
}
