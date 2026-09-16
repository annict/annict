package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBWorkDeletionNewUsecaseはAnnict DB管理画面の削除確認画面に必要なデータ
// (削除を確認する対象のwork) を取得するユースケース。公開中のworkもアーカイブ済みのworkも
// 削除できるため、これは削除自体が適用するRailsのscope Work.without_deletedに一致する。
// すでに削除済みのworkだけをnot foundとして扱う。
type GetDBWorkDeletionNewUsecase struct {
	workRepo *repository.WorkRepository
}

func NewGetDBWorkDeletionNewUsecase(workRepo *repository.WorkRepository) *GetDBWorkDeletionNewUsecase {
	return &GetDBWorkDeletionNewUsecase{workRepo: workRepo}
}

type GetDBWorkDeletionNewInput struct {
	User   *model.User
	WorkID model.WorkID
}

type GetDBWorkDeletionNewOutput struct {
	Work *model.Work
}

// Executeはworkを取得する前に管理者を認可する。workが存在しない、またはすでに
// 削除済みの場合はAppErrCodeResourceNotFoundの *model.AppErrorを返し、Handler側で404に
// 変換する。
func (uc *GetDBWorkDeletionNewUsecase) Execute(ctx context.Context, input GetDBWorkDeletionNewInput) (*GetDBWorkDeletionNewOutput, error) {
	if input.User == nil || !input.User.IsAdmin() {
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
	if work == nil || work.DerivedStatus() == model.WorkStatusDeleted {
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_work_not_found"),
			Metadata: map[string]string{"work_id": input.WorkID.String()},
		}
	}

	return &GetDBWorkDeletionNewOutput{Work: work}, nil
}
