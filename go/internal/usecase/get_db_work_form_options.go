package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBWorkFormOptionsUsecase is the use case for retrieving the select options for the work form on the DB admin screen.
//
// [Ja] DB 管理画面の作品フォーム用選択肢を取得するユースケース。
type GetDBWorkFormOptionsUsecase struct {
	numberFormatRepo *repository.NumberFormatRepository
}

// NewGetDBWorkFormOptionsUsecase creates a new GetDBWorkFormOptionsUsecase.
//
// [Ja] 新しい GetDBWorkFormOptionsUsecase を作成する。
func NewGetDBWorkFormOptionsUsecase(numberFormatRepo *repository.NumberFormatRepository) *GetDBWorkFormOptionsUsecase {
	return &GetDBWorkFormOptionsUsecase{
		numberFormatRepo: numberFormatRepo,
	}
}

// GetDBWorkFormOptionsOutput is the output of the use case.
//
// [Ja] ユースケースの出力。
type GetDBWorkFormOptionsOutput struct {
	NumberFormats []model.NumberFormat
}

// Execute retrieves the select option data for the form.
//
// [Ja] フォーム用の選択肢データを取得する。
func (uc *GetDBWorkFormOptionsUsecase) Execute(ctx context.Context) (*GetDBWorkFormOptionsOutput, error) {
	numberFormats, err := uc.numberFormatRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("NumberFormatの取得に失敗: %w", err)
	}

	return &GetDBWorkFormOptionsOutput{
		NumberFormats: numberFormats,
	}, nil
}
