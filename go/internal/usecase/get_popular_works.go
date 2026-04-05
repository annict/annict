package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetPopularWorksUsecase は人気作品を取得するユースケースです
type GetPopularWorksUsecase struct {
	workRepo *repository.WorkRepository
}

// NewGetPopularWorksUsecase は新しいGetPopularWorksUsecaseを作成します
func NewGetPopularWorksUsecase(workRepo *repository.WorkRepository) *GetPopularWorksUsecase {
	return &GetPopularWorksUsecase{
		workRepo: workRepo,
	}
}

// GetPopularWorksOutput はユースケースの出力です
type GetPopularWorksOutput struct {
	Works []model.WorkWithDetails
}

// Execute は人気作品を取得します
func (uc *GetPopularWorksUsecase) Execute(ctx context.Context) (*GetPopularWorksOutput, error) {
	works, err := uc.workRepo.GetPopularWorksWithDetails(ctx)
	if err != nil {
		return nil, fmt.Errorf("人気作品の取得に失敗: %w", err)
	}
	return &GetPopularWorksOutput{Works: works}, nil
}
