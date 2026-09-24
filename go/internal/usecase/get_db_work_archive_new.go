package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBWorkArchiveNewUsecaseはAnnict DB管理画面の非公開確認画面に必要なデータ
// (非公開を確認する対象のwork) を取得するユースケース。アーカイブできるのは現在公開中の
// workだけで、これはRailsのscope Work.without_deleted.publishedに一致する。すでに
// アーカイブ済み・削除済みのworkはnot foundとして扱う。
type GetDBWorkArchiveNewUsecase struct {
	workRepo *repository.WorkRepository
}

func NewGetDBWorkArchiveNewUsecase(workRepo *repository.WorkRepository) *GetDBWorkArchiveNewUsecase {
	return &GetDBWorkArchiveNewUsecase{workRepo: workRepo}
}

type GetDBWorkArchiveNewInput struct {
	User   *model.User
	WorkID model.WorkID
}

type GetDBWorkArchiveNewOutput struct {
	Work *model.Work
}

// Executeはworkを取得する前にコミッターを認可する。workが存在しない、または現在
// 公開中でない場合はAppErrCodeResourceNotFoundの *model.AppErrorを返し、Handler側で404に
// 変換する。
func (uc *GetDBWorkArchiveNewUsecase) Execute(ctx context.Context, input GetDBWorkArchiveNewInput) (*GetDBWorkArchiveNewOutput, error) {
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
	if work == nil || work.DerivedStatus() != model.WorkStatusPublished {
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_work_not_found"),
			Metadata: map[string]string{"work_id": input.WorkID.String()},
		}
	}

	return &GetDBWorkArchiveNewOutput{Work: work}, nil
}
