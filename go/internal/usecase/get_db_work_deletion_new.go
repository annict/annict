package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBWorkDeletionNewUsecase loads the data the Annict DB admin delete-confirmation screen
// needs: the work whose deletion is being confirmed. Both a published and an archived work
// can be deleted, matching the Rails scope Work.without_deleted the delete itself applies,
// so only an already deleted work is reported as not found.
//
// [Ja] GetDBWorkDeletionNewUsecase は Annict DB 管理画面の削除確認画面に必要なデータ
// (削除を確認する対象の work) を取得するユースケース。公開中の work もアーカイブ済みの work も
// 削除できるため、これは削除自体が適用する Rails の scope Work.without_deleted に一致する。
// すでに削除済みの work だけを not found として扱う。
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

// Execute first authorizes an administrator before looking up the work. It returns a
// *model.AppError with AppErrCodeResourceNotFound when the work does not exist or is already
// deleted; the handler converts that to 404.
//
// [Ja] Execute は work を取得する前に管理者を認可する。work が存在しない、またはすでに
// 削除済みの場合は AppErrCodeResourceNotFound の *model.AppError を返し、Handler 側で 404 に
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
