package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBPersonArchiveNewUsecaseはAnnict DB管理画面の人物のアーカイブ確認画面に必要なデータ
// (アーカイブを確認する対象の人物) を取得するユースケース。アーカイブできるのは公開中の人物
// だけで、それ以外の状態の人物はnot foundとして扱う。
type GetDBPersonArchiveNewUsecase struct {
	personRepo *repository.PersonRepository
}

// NewGetDBPersonArchiveNewUsecaseは新しいGetDBPersonArchiveNewUsecaseを作成する。
func NewGetDBPersonArchiveNewUsecase(personRepo *repository.PersonRepository) *GetDBPersonArchiveNewUsecase {
	return &GetDBPersonArchiveNewUsecase{personRepo: personRepo}
}

// GetDBPersonArchiveNewInputはユースケースの入力。
type GetDBPersonArchiveNewInput struct {
	User     *model.User
	PersonID model.PersonID
}

// GetDBPersonArchiveNewOutputはユースケースの出力。
type GetDBPersonArchiveNewOutput struct {
	Person *model.Person
}

// Executeは人物を取得する前にコミッターを認可する。人物が存在しない、または現在公開中でない
// 場合はAppErrCodeResourceNotFoundの *model.AppErrorを返し、Handler側で404に変換する。
func (uc *GetDBPersonArchiveNewUsecase) Execute(ctx context.Context, input GetDBPersonArchiveNewInput) (*GetDBPersonArchiveNewOutput, error) {
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
	if person == nil || person.Status != model.PersonStatusPublished {
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_person_not_found"),
			Metadata: map[string]string{"person_id": input.PersonID.String()},
		}
	}

	return &GetDBPersonArchiveNewOutput{Person: person}, nil
}
