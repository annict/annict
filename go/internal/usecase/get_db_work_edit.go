package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDbWorkEditUsecase loads the data the Annict DB admin work edit form needs:
// the target work and the number-format options for its select box.
//
// [Ja] GetDbWorkEditUsecase は Annict DB 管理画面の作品編集フォームに必要なデータ
// (対象の work と、その選択肢となる number format) を取得するユースケース。
type GetDbWorkEditUsecase struct {
	workRepo         *repository.WorkRepository
	numberFormatRepo *repository.NumberFormatRepository
}

func NewGetDbWorkEditUsecase(
	workRepo *repository.WorkRepository,
	numberFormatRepo *repository.NumberFormatRepository,
) *GetDbWorkEditUsecase {
	return &GetDbWorkEditUsecase{
		workRepo:         workRepo,
		numberFormatRepo: numberFormatRepo,
	}
}

type GetDbWorkEditInput struct {
	WorkID model.WorkID
}

type GetDbWorkEditOutput struct {
	Work          *model.Work
	NumberFormats []model.NumberFormat
}

// Execute returns the work to edit and the form options. It returns a
// *model.AppError with AppErrCodeResourceNotFound when the work does not exist;
// the handler converts that to 404.
//
// [Ja] Execute は編集対象の work とフォームの選択肢を返す。work が存在しない場合は
// AppErrCodeResourceNotFound の *model.AppError を返し、Handler 側で 404 に変換する。
func (uc *GetDbWorkEditUsecase) Execute(ctx context.Context, input GetDbWorkEditInput) (*GetDbWorkEditOutput, error) {
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

	return &GetDbWorkEditOutput{
		Work:          work,
		NumberFormats: numberFormats,
	}, nil
}
