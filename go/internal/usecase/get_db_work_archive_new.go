package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBWorkArchiveNewUsecase loads the data the Annict DB admin archive-confirmation
// screen needs: the work whose archiving is being confirmed. Only a currently published
// work is archivable, matching the Rails scope Work.without_deleted.published, so an
// already-archived or deleted work is reported as not found.
//
// [Ja] GetDBWorkArchiveNewUsecase は Annict DB 管理画面の非公開確認画面に必要なデータ
// (非公開を確認する対象の work) を取得するユースケース。アーカイブできるのは現在公開中の
// work だけで、これは Rails の scope Work.without_deleted.published に一致する。すでに
// アーカイブ済み・削除済みの work は not found として扱う。
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

// Execute first authorizes a committer before looking up the work. It returns a
// *model.AppError with AppErrCodeResourceNotFound when the work does not exist or is not
// currently published; the handler converts that to 404.
//
// [Ja] Execute は work を取得する前にコミッターを認可する。work が存在しない、または現在
// 公開中でない場合は AppErrCodeResourceNotFound の *model.AppError を返し、Handler 側で 404 に
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
