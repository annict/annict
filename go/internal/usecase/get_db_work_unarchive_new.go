package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBWorkUnarchiveNewUsecaseはAnnict DB管理画面の公開確認画面に必要なデータ
// (公開を確認する対象のwork) を取得するユースケース。公開できるのは現在アーカイブ済みの
// workだけで、これは再公開自体が適用するRailsのscope Work.without_deleted.unpublishedに
// 一致する。公開中・削除済みのworkはnot foundとして扱う。
type GetDBWorkUnarchiveNewUsecase struct {
	workRepo *repository.WorkRepository
}

func NewGetDBWorkUnarchiveNewUsecase(workRepo *repository.WorkRepository) *GetDBWorkUnarchiveNewUsecase {
	return &GetDBWorkUnarchiveNewUsecase{workRepo: workRepo}
}

type GetDBWorkUnarchiveNewInput struct {
	User   *model.User
	WorkID model.WorkID
}

type GetDBWorkUnarchiveNewOutput struct {
	Work *model.Work
}

// Executeはworkを取得する前にコミッターを認可する。workが存在しない、または現在
// アーカイブ済みでない場合はAppErrCodeResourceNotFoundの *model.AppErrorを返し、Handler側で
// 404に変換する。
func (uc *GetDBWorkUnarchiveNewUsecase) Execute(ctx context.Context, input GetDBWorkUnarchiveNewInput) (*GetDBWorkUnarchiveNewOutput, error) {
	if input.User == nil || !input.User.IsCommitter() {
		return nil, &model.AppError{
			Code:     model.AppErrCodeForbidden,
			UserMsg:  i18n.T(ctx, "error_forbidden"),
			Metadata: map[string]string{"work_id": input.WorkID.String()},
		}
	}

	work, err := uc.workRepo.GetForStateChangeByID(ctx, input.WorkID)
	if err != nil {
		return nil, fmt.Errorf("作品の取得に失敗: %w", err)
	}
	if work == nil || work.DerivedStatus() != model.WorkStatusArchived {
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_work_not_found"),
			Metadata: map[string]string{"work_id": input.WorkID.String()},
		}
	}

	return &GetDBWorkUnarchiveNewOutput{Work: work}, nil
}
