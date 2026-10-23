package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBPersonDelistingNewUsecaseはAnnict DB管理画面の人物の掲載停止確認画面に必要なデータ
// (掲載停止を確認する対象の人物と、入力欄の初期値にする対応記録) を取得するユースケース。
// 掲載を停止できるのは公開中とアーカイブの人物だけで、それ以外の状態の人物はnot foundとして扱う。
type GetDBPersonDelistingNewUsecase struct {
	personRepo *repository.PersonRepository
}

// NewGetDBPersonDelistingNewUsecaseは新しいGetDBPersonDelistingNewUsecaseを作成する。
func NewGetDBPersonDelistingNewUsecase(personRepo *repository.PersonRepository) *GetDBPersonDelistingNewUsecase {
	return &GetDBPersonDelistingNewUsecase{personRepo: personRepo}
}

// GetDBPersonDelistingNewInputはユースケースの入力。
type GetDBPersonDelistingNewInput struct {
	User     *model.User
	PersonID model.PersonID
}

// GetDBPersonDelistingNewOutputはユースケースの出力。
type GetDBPersonDelistingNewOutput struct {
	Person *model.Person
}

// Executeは人物を取得する前に管理者を認可する。人物が存在しない、または掲載を停止できる状態で
// ない場合はAppErrCodeResourceNotFoundの *model.AppErrorを返し、Handler側で404に変換する。
func (uc *GetDBPersonDelistingNewUsecase) Execute(ctx context.Context, input GetDBPersonDelistingNewInput) (*GetDBPersonDelistingNewOutput, error) {
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
	if person == nil || !person.Status.IsViewable() {
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_person_not_found"),
			Metadata: map[string]string{"person_id": input.PersonID.String()},
		}
	}

	return &GetDBPersonDelistingNewOutput{Person: person}, nil
}
