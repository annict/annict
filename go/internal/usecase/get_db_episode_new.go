package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBEpisodeNewUsecase is the use case for the bulk-create form on the DB admin screen. It
// retrieves the parent work the page's heading and subnav describe; the form itself has no
// stored state to load, since every row is typed into one textarea.
//
// [Ja] GetDBEpisodeNewUsecase は DB 管理画面の一括作成フォームのユースケース。ページの見出しと
// サブナビが示す親作品を取得する。フォーム自体は 1 つの textarea に全行を入力する形のため、
// 読み込む保存済みの状態を持たない。
type GetDBEpisodeNewUsecase struct {
	workRepo *repository.WorkRepository
}

// NewGetDBEpisodeNewUsecase creates a new GetDBEpisodeNewUsecase.
//
// [Ja] 新しい GetDBEpisodeNewUsecase を作成する。
func NewGetDBEpisodeNewUsecase(workRepo *repository.WorkRepository) *GetDBEpisodeNewUsecase {
	return &GetDBEpisodeNewUsecase{workRepo: workRepo}
}

// GetDBEpisodeNewInput is the input for the use case.
//
// [Ja] ユースケースの入力。
type GetDBEpisodeNewInput struct {
	WorkID model.WorkID
}

// GetDBEpisodeNewOutput is the output of the use case.
//
// [Ja] ユースケースの出力。
type GetDBEpisodeNewOutput struct {
	Work *model.Work
	// ManualCreationState travels as the domain state rather than as separate flags, so the
	// page and the rejected submit take the reason to state from the same place.
	//
	// [Ja] ManualCreationState は個別のフラグではなくドメインの状態のまま運ぶ。ページと
	// 却下された送信が、述べる理由を同じ場所から取れるようにするため。
	ManualCreationState model.ManualEpisodeCreationState
}

// Execute returns the parent work of the bulk-create form. It returns a *model.AppError with
// AppErrCodeResourceNotFound when the work does not exist or is deleted; the handler converts
// that to 404.
//
// [Ja] Execute は一括作成フォームの親作品を返す。作品が存在しない、または削除済みの場合は
// AppErrCodeResourceNotFound の *model.AppError を返し、Handler 側で 404 に変換する。
func (uc *GetDBEpisodeNewUsecase) Execute(ctx context.Context, input GetDBEpisodeNewInput) (*GetDBEpisodeNewOutput, error) {
	formWork, err := uc.workRepo.GetForEpisodeFormByID(ctx, input.WorkID)
	if err != nil {
		return nil, fmt.Errorf("作品の取得に失敗: %w", err)
	}
	if formWork == nil {
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_work_not_found"),
			Metadata: map[string]string{"work_id": input.WorkID.String()},
		}
	}

	return &GetDBEpisodeNewOutput{
		Work:                formWork.Work,
		ManualCreationState: formWork.ManualCreationState,
	}, nil
}
