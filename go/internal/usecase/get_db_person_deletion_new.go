package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBPersonDeletionNewUsecaseはAnnict DB管理画面の人物の削除確認画面に必要なデータ
// (削除を確認する対象の人物) を取得するユースケース。削除できるのは一覧に表示する公開中と
// アーカイブの人物で、削除済み・掲載停止の人物はnot foundとして扱う。
type GetDBPersonDeletionNewUsecase struct {
	personRepo *repository.PersonRepository
}

// NewGetDBPersonDeletionNewUsecaseは新しいGetDBPersonDeletionNewUsecaseを作成する。
func NewGetDBPersonDeletionNewUsecase(personRepo *repository.PersonRepository) *GetDBPersonDeletionNewUsecase {
	return &GetDBPersonDeletionNewUsecase{personRepo: personRepo}
}

// GetDBPersonDeletionNewInputはユースケースの入力。
type GetDBPersonDeletionNewInput struct {
	User     *model.User
	PersonID model.PersonID
}

// GetDBPersonDeletionNewOutputはユースケースの出力。
type GetDBPersonDeletionNewOutput struct {
	Person *model.Person
}

// Executeは人物を取得する前に管理者を認可する。人物が存在しない、または削除できる状態でない
// 場合はAppErrCodeResourceNotFoundの *model.AppErrorを返し、Handler側で404に変換する。
func (uc *GetDBPersonDeletionNewUsecase) Execute(ctx context.Context, input GetDBPersonDeletionNewInput) (*GetDBPersonDeletionNewOutput, error) {
	if input.User == nil || !input.User.IsAdmin() {
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
	if person == nil || !person.Status.IsViewable() {
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_person_not_found"),
			Metadata: map[string]string{"person_id": input.PersonID.String()},
		}
	}

	return &GetDBPersonDeletionNewOutput{Person: person}, nil
}
