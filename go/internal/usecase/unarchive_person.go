package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// UnarchivePersonUsecaseはAnnict DB管理画面から非公開の人物を公開に戻す。
// ArchivePersonUsecaseの逆で、people.statusをpublishedに戻し、unpublished_atをNULLに戻す。
// 公開に戻せるのは非公開の人物だけで、それ以外の状態の人物への送信はnot foundになる。編集履歴
// には記録しない。
//
// 認可 (committer) は書き込みより先に本UseCaseで行い、ルートのRequireCommitter middlewareも同じ
// リクエストを手前で拒否する。UseCase側に検査を残すことで、そのルート以外から到達した呼び出し元
// にも同じ権限を要求する。
type UnarchivePersonUsecase struct {
	personRepo *repository.PersonRepository
}

// NewUnarchivePersonUsecaseは新しいUnarchivePersonUsecaseを作成する。
func NewUnarchivePersonUsecase(personRepo *repository.PersonRepository) *UnarchivePersonUsecase {
	return &UnarchivePersonUsecase{personRepo: personRepo}
}

// UnarchivePersonInputはアーカイブ解除する人物と、書き込みを認可するユーザーを指定する。
type UnarchivePersonInput struct {
	User     *model.User
	PersonID model.PersonID
}

// UnarchivePersonOutputはユースケースの出力。
type UnarchivePersonOutput struct {
	PersonID model.PersonID
}

// Executeはcommitterを認可してから人物のアーカイブを解除する。人物が存在しないか、アーカイブ解除できる状態でない
// 場合はAppErrCodeResourceNotFoundの *model.AppErrorを返し、Handler側で404に変換する。
func (uc *UnarchivePersonUsecase) Execute(ctx context.Context, input UnarchivePersonInput) (*UnarchivePersonOutput, error) {
	if input.User == nil || !input.User.IsCommitter() {
		return nil, &model.AppError{
			Code:     model.AppErrCodeForbidden,
			UserMsg:  i18n.T(ctx, "error_forbidden"),
			Metadata: map[string]string{"person_id": input.PersonID.String()},
		}
	}

	written, err := uc.personRepo.Unarchive(ctx, input.PersonID)
	if err != nil {
		return nil, fmt.Errorf("人物のアーカイブ解除に失敗しました: %w", err)
	}
	if !written {
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_person_not_found"),
			Metadata: map[string]string{"person_id": input.PersonID.String()},
		}
	}

	return &UnarchivePersonOutput{PersonID: input.PersonID}, nil
}
