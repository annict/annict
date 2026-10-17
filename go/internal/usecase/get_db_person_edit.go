package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBPersonEditUsecaseはDB管理画面の人物編集フォームのユースケース。フォームの初期値に
// なる保存済みの人物と、出身地の選択肢を取得する。
type GetDBPersonEditUsecase struct {
	personRepo     *repository.PersonRepository
	prefectureRepo *repository.PrefectureRepository
}

// NewGetDBPersonEditUsecaseは新しいGetDBPersonEditUsecaseを作成する。
func NewGetDBPersonEditUsecase(
	personRepo *repository.PersonRepository,
	prefectureRepo *repository.PrefectureRepository,
) *GetDBPersonEditUsecase {
	return &GetDBPersonEditUsecase{personRepo: personRepo, prefectureRepo: prefectureRepo}
}

// GetDBPersonEditInputはユースケースの入力。
type GetDBPersonEditInput struct {
	User     *model.User
	PersonID model.PersonID
}

// GetDBPersonEditOutputはユースケースの出力。
type GetDBPersonEditOutput struct {
	Person      *model.Person
	Prefectures []model.Prefecture
}

// Executeは人物を取得する前にコミッターを認可し、編集対象の人物と出身地の選択肢を返す。
// 人物が存在しないか編集できない状態 (削除・掲載停止) の場合はAppErrCodeResourceNotFoundの
// *model.AppErrorを返し、Handler側で404に変換する。アーカイブの人物は返す。
func (uc *GetDBPersonEditUsecase) Execute(ctx context.Context, input GetDBPersonEditInput) (*GetDBPersonEditOutput, error) {
	if input.User == nil || !input.User.IsCommitter() {
		return nil, &model.AppError{
			Code:     model.AppErrCodeForbidden,
			UserMsg:  i18n.T(ctx, "error_forbidden"),
			Metadata: map[string]string{"person_id": input.PersonID.String()},
		}
	}

	person, err := uc.personRepo.GetForEditByID(ctx, input.PersonID)
	if err != nil {
		return nil, fmt.Errorf("人物の取得に失敗: %w", err)
	}
	if person == nil {
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_person_not_found"),
			Metadata: map[string]string{"person_id": input.PersonID.String()},
		}
	}

	prefectures, err := uc.prefectureRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("出身地の取得に失敗: %w", err)
	}

	return &GetDBPersonEditOutput{
		Person:      person,
		Prefectures: prefectures,
	}, nil
}
