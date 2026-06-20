package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDbWorkFormOptionsUsecase はDB管理画面の作品フォーム用選択肢を取得するユースケースです
type GetDbWorkFormOptionsUsecase struct {
	numberFormatRepo *repository.NumberFormatRepository
}

// NewGetDbWorkFormOptionsUsecase は新しいGetDbWorkFormOptionsUsecaseを作成します
func NewGetDbWorkFormOptionsUsecase(numberFormatRepo *repository.NumberFormatRepository) *GetDbWorkFormOptionsUsecase {
	return &GetDbWorkFormOptionsUsecase{
		numberFormatRepo: numberFormatRepo,
	}
}

// GetDbWorkFormOptionsOutput はユースケースの出力です
type GetDbWorkFormOptionsOutput struct {
	NumberFormats []model.NumberFormat
}

// Execute はフォーム用の選択肢データを取得します
func (uc *GetDbWorkFormOptionsUsecase) Execute(ctx context.Context) (*GetDbWorkFormOptionsOutput, error) {
	numberFormats, err := uc.numberFormatRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("NumberFormatの取得に失敗: %w", err)
	}

	return &GetDbWorkFormOptionsOutput{
		NumberFormats: numberFormats,
	}, nil
}
