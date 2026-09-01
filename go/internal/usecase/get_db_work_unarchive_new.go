package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBWorkUnarchiveNewUsecase loads the data the Annict DB admin publish-confirmation
// screen needs: the work whose re-publish is being confirmed. Only a currently archived
// work is publishable, matching the Rails scope Work.without_deleted.unpublished the
// re-publish itself applies, so a published or deleted work is reported as not found.
//
// [Ja] GetDBWorkUnarchiveNewUsecase は Annict DB 管理画面の公開確認画面に必要なデータ
// (公開を確認する対象の work) を取得するユースケース。公開できるのは現在アーカイブ済みの
// work だけで、これは再公開自体が適用する Rails の scope Work.without_deleted.unpublished に
// 一致する。公開中・削除済みの work は not found として扱う。
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

// Execute first authorizes a committer before looking up the work. It returns a *model.AppError
// with AppErrCodeResourceNotFound when the work does not exist or is not currently archived; the
// handler converts that to 404.
//
// [Ja] Execute は work を取得する前にコミッターを認可する。work が存在しない、または現在
// アーカイブ済みでない場合は AppErrCodeResourceNotFound の *model.AppError を返し、Handler 側で
// 404 に変換する。
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
