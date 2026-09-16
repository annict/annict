package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBWorkFormOptionsUsecaseはDB管理画面の作品フォーム用選択肢を取得するユースケース。
type GetDBWorkFormOptionsUsecase struct {
	numberFormatRepo *repository.NumberFormatRepository
}

// NewGetDBWorkFormOptionsUsecaseは新しいGetDBWorkFormOptionsUsecaseを作成する。
func NewGetDBWorkFormOptionsUsecase(numberFormatRepo *repository.NumberFormatRepository) *GetDBWorkFormOptionsUsecase {
	return &GetDBWorkFormOptionsUsecase{
		numberFormatRepo: numberFormatRepo,
	}
}

// GetDBWorkFormOptionsOutputはユースケースの出力。
type GetDBWorkFormOptionsOutput struct {
	NumberFormats []model.NumberFormat
}

// Executeはフォーム用の選択肢データを取得する。
func (uc *GetDBWorkFormOptionsUsecase) Execute(ctx context.Context) (*GetDBWorkFormOptionsOutput, error) {
	numberFormats, err := uc.numberFormatRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("NumberFormatの取得に失敗: %w", err)
	}

	return &GetDBWorkFormOptionsOutput{
		NumberFormats: numberFormats,
	}, nil
}
