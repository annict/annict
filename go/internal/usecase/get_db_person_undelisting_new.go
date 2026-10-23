package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBPersonUndelistingNewUsecaseはAnnict DB管理画面の人物の掲載停止解除確認画面に必要な
// データ (掲載停止の解除を確認する対象の人物と、その対応記録) を取得するユースケース。掲載停止を
// 解除できるのは掲載停止の人物だけで、それ以外の状態の人物はnot foundとして扱う。
type GetDBPersonUndelistingNewUsecase struct {
	personRepo *repository.PersonRepository
}

// NewGetDBPersonUndelistingNewUsecaseは新しいGetDBPersonUndelistingNewUsecaseを作成する。
func NewGetDBPersonUndelistingNewUsecase(personRepo *repository.PersonRepository) *GetDBPersonUndelistingNewUsecase {
	return &GetDBPersonUndelistingNewUsecase{personRepo: personRepo}
}

// GetDBPersonUndelistingNewInputはユースケースの入力。
type GetDBPersonUndelistingNewInput struct {
	User     *model.User
	PersonID model.PersonID
}

// GetDBPersonUndelistingNewOutputはユースケースの出力。
type GetDBPersonUndelistingNewOutput struct {
	Person *model.Person
}

// Executeは人物を取得する前に管理者を認可する。人物が存在しない、または現在掲載停止でない場合は
// AppErrCodeResourceNotFoundの *model.AppErrorを返し、Handler側で404に変換する。
func (uc *GetDBPersonUndelistingNewUsecase) Execute(ctx context.Context, input GetDBPersonUndelistingNewInput) (*GetDBPersonUndelistingNewOutput, error) {
	if input.User == nil || !input.User.IsAdmin() {
		return nil, &model.AppError{
			Code:     model.AppErrCodeForbidden,
			UserMsg:  i18n.T(ctx, "error_forbidden"),
			Metadata: map[string]string{"person_id": input.PersonID.String()},
		}
	}

	person, err := uc.personRepo.GetForDelistingByID(ctx, input.PersonID)
	if err != nil {
		return nil, fmt.Errorf("人物の取得に失敗: %w", err)
	}
	if person == nil || person.Status != model.PersonStatusDelisted {
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_person_not_found"),
			Metadata: map[string]string{"person_id": input.PersonID.String()},
		}
	}

	return &GetDBPersonUndelistingNewOutput{Person: person}, nil
}
