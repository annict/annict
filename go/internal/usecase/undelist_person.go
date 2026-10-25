package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/validator"
)

// UndelistPersonUsecaseはAnnict DB管理画面から人物の掲載停止を解除する。DelistPersonUsecaseの
// 逆で、people.statusをpublishedに戻し、橋渡しのunpublished_atをNULLに戻す。対応記録は申出に
// 応じた経緯として残す。掲載停止を解除できるのは掲載停止の人物だけで、それ以外の状態の人物への
// 送信はnot foundになる。確認後に版が進んだ送信は競合として扱う。編集履歴には記録しない。
//
// 認可 (admin) は書き込みより先に本UseCaseで行い、ルートのRequireAdmin middlewareも同じ
// リクエストを手前で拒否する。UseCase側に検査を残すことで、そのルート以外から到達した呼び出し元
// にも同じ権限を要求する。
type UndelistPersonUsecase struct {
	personRepo *repository.PersonRepository
}

// NewUndelistPersonUsecaseは新しいUndelistPersonUsecaseを作成する。
func NewUndelistPersonUsecase(personRepo *repository.PersonRepository) *UndelistPersonUsecase {
	return &UndelistPersonUsecase{personRepo: personRepo}
}

// UndelistPersonInputは掲載停止を解除する人物と確認時の版、認可するユーザーを指定する。
type UndelistPersonInput struct {
	User      *model.User
	PersonID  model.PersonID
	UpdatedAt string
}

// UndelistPersonOutputはユースケースの出力。
type UndelistPersonOutput struct {
	PersonID model.PersonID
}

// Executeはadminを認可してから人物の掲載停止を解除する。人物が存在しないか、掲載停止でない場合は
// AppErrCodeResourceNotFoundの *model.AppErrorを返し、Handler側で404に変換する。
// 確認後に版が進んだ場合はAppErrCodeConflictを返す。
func (uc *UndelistPersonUsecase) Execute(ctx context.Context, input UndelistPersonInput) (*UndelistPersonOutput, error) {
	if input.User == nil || !input.User.IsAdmin() {
		return nil, &model.AppError{
			Code:     model.AppErrCodeForbidden,
			UserMsg:  i18n.T(ctx, "error_forbidden"),
			Metadata: map[string]string{"person_id": input.PersonID.String()},
		}
	}
	version, err := time.Parse(validator.FormVersionLayout, input.UpdatedAt)
	if err != nil {
		ve := model.NewValidationError()
		ve.AddGlobal(i18n.T(ctx, "validation_version_missing"))
		return nil, ve
	}

	written, err := uc.personRepo.Undelist(ctx, input.PersonID, version)
	if err != nil {
		return nil, fmt.Errorf("人物の掲載停止の解除に失敗しました: %w", err)
	}
	if !written {
		current, err := uc.personRepo.GetForDelistingByID(ctx, input.PersonID)
		if err != nil {
			return nil, fmt.Errorf("人物の取得に失敗しました: %w", err)
		}
		if current != nil && current.Status == model.PersonStatusDelisted {
			return nil, personConflictError(ctx, input.PersonID)
		}
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_person_not_found"),
			Metadata: map[string]string{"person_id": input.PersonID.String()},
		}
	}

	return &UndelistPersonOutput{PersonID: input.PersonID}, nil
}
