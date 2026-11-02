package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBEpisodeNewUsecaseはDB管理画面の一括作成フォームのユースケース。ページの見出しと
// サブナビが示す親作品を取得する。フォーム自体は1つのtextareaに全行を入力する形のため、
// 読み込む保存済みの状態を持たない。
type GetDBEpisodeNewUsecase struct {
	workRepo *repository.WorkRepository
}

// NewGetDBEpisodeNewUsecaseは新しいGetDBEpisodeNewUsecaseを作成する。
func NewGetDBEpisodeNewUsecase(workRepo *repository.WorkRepository) *GetDBEpisodeNewUsecase {
	return &GetDBEpisodeNewUsecase{workRepo: workRepo}
}

// GetDBEpisodeNewInputはユースケースの入力。
type GetDBEpisodeNewInput struct {
	User   *model.User
	WorkID model.WorkID
}

// GetDBEpisodeNewOutputはユースケースの出力。
type GetDBEpisodeNewOutput struct {
	Work *model.Work
	// ManualCreationStateは個別のフラグではなくドメインの状態のまま運ぶ。ページと
	// 却下された送信が、述べる理由を同じ場所から取れるようにするため。
	ManualCreationState model.ManualEpisodeCreationState
	// RawNumberRequiredは、作品に有効な放送枠があり、一括作成のすべての行に話数を求めるか。
	// フォームはこれを見て、話数が要ることを案内する。
	RawNumberRequired bool
}

// Executeは作品を取得する前にコミッターを認可し、一括作成フォームの親作品を返す。作品が
// 存在しない、または削除済みの場合はAppErrCodeResourceNotFoundの *model.AppErrorを返し、
// Handler側で404に変換する。
func (uc *GetDBEpisodeNewUsecase) Execute(ctx context.Context, input GetDBEpisodeNewInput) (*GetDBEpisodeNewOutput, error) {
	if input.User == nil || !input.User.IsCommitter() {
		return nil, &model.AppError{
			Code:     model.AppErrCodeForbidden,
			UserMsg:  i18n.T(ctx, "error_forbidden"),
			Metadata: map[string]string{"work_id": input.WorkID.String()},
		}
	}

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
		RawNumberRequired:   formWork.HasActiveSlots,
	}, nil
}
