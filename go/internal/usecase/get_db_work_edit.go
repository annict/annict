package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBWorkEditUsecaseはAnnict DB管理画面の作品編集フォームに必要なデータ
// (対象のworkと、その選択肢となるnumber format) を取得するユースケース。
type GetDBWorkEditUsecase struct {
	workRepo         *repository.WorkRepository
	numberFormatRepo *repository.NumberFormatRepository
}

func NewGetDBWorkEditUsecase(
	workRepo *repository.WorkRepository,
	numberFormatRepo *repository.NumberFormatRepository,
) *GetDBWorkEditUsecase {
	return &GetDBWorkEditUsecase{
		workRepo:         workRepo,
		numberFormatRepo: numberFormatRepo,
	}
}

type GetDBWorkEditInput struct {
	WorkID model.WorkID
}

type GetDBWorkEditOutput struct {
	Work          *model.Work
	NumberFormats []model.NumberFormat
}

// Executeは編集対象のworkとフォームの選択肢を返す。workが存在しない場合は
// AppErrCodeResourceNotFoundの *model.AppErrorを返し、Handler側で404に変換する。
func (uc *GetDBWorkEditUsecase) Execute(ctx context.Context, input GetDBWorkEditInput) (*GetDBWorkEditOutput, error) {
	work, err := uc.workRepo.GetForEditByID(ctx, input.WorkID)
	if err != nil {
		return nil, fmt.Errorf("作品の取得に失敗: %w", err)
	}
	if work == nil {
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_work_not_found"),
			Metadata: map[string]string{"work_id": input.WorkID.String()},
		}
	}

	numberFormats, err := uc.numberFormatRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("NumberFormatの取得に失敗: %w", err)
	}

	return &GetDBWorkEditOutput{
		Work:          work,
		NumberFormats: numberFormats,
	}, nil
}
