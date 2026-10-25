package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBPersonUnarchiveNewUsecaseはAnnict DB管理画面の人物のアーカイブ解除確認画面に必要な
// データ (アーカイブ解除を確認する対象の人物) を取得するユースケース。アーカイブ解除できるのは
// アーカイブの人物だけで、それ以外の状態の人物はnot foundとして扱う。
type GetDBPersonUnarchiveNewUsecase struct {
	personRepo *repository.PersonRepository
}

// NewGetDBPersonUnarchiveNewUsecaseは新しいGetDBPersonUnarchiveNewUsecaseを作成する。
func NewGetDBPersonUnarchiveNewUsecase(personRepo *repository.PersonRepository) *GetDBPersonUnarchiveNewUsecase {
	return &GetDBPersonUnarchiveNewUsecase{personRepo: personRepo}
}

// GetDBPersonUnarchiveNewInputはユースケースの入力。
type GetDBPersonUnarchiveNewInput struct {
	User     *model.User
	PersonID model.PersonID
}

// GetDBPersonUnarchiveNewOutputはユースケースの出力。
type GetDBPersonUnarchiveNewOutput struct {
	Person *model.Person
}

// Executeは人物を取得する前にコミッターを認可する。人物が存在しない、または現在アーカイブでない
// 場合はAppErrCodeResourceNotFoundの *model.AppErrorを返し、Handler側で404に変換する。
func (uc *GetDBPersonUnarchiveNewUsecase) Execute(ctx context.Context, input GetDBPersonUnarchiveNewInput) (*GetDBPersonUnarchiveNewOutput, error) {
	if input.User == nil || !input.User.IsCommitter() {
		return nil, &model.AppError{
			Code:     model.AppErrCodeForbidden,
			UserMsg:  i18n.T(ctx, "error_forbidden"),
			Metadata: map[string]string{"person_id": input.PersonID.String()},
		}
	}

	person, err := uc.personRepo.GetForStateChangeByID(ctx, input.PersonID)
	if err != nil {
		return nil, fmt.Errorf("人物の取得に失敗: %w", err)
	}
	if person == nil || person.Status != model.PersonStatusArchived {
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_person_not_found"),
			Metadata: map[string]string{"person_id": input.PersonID.String()},
		}
	}

	return &GetDBPersonUnarchiveNewOutput{Person: person}, nil
}
