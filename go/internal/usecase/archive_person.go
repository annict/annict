package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// ArchivePersonUsecaseはAnnict DB管理画面から人物をアーカイブする。people.statusを
// archivedにするだけで、キャスト・スタッフは書き換えず、編集履歴にも記録しない。archivedは
// 閲覧を保つ凍結のため、Rails版への橋渡し (unpublished_at) も書かない。アーカイブできるのは
// 公開中の人物だけで、状態の確認と書き込みはリポジトリの1文の更新で行う。確認画面を開いた後に
// 他者が状態を変えた人物への送信は、何も書かずにnot foundになる。
//
// 認可 (committer) は書き込みより先に本UseCaseで行い、ルートのRequireCommitter middlewareも同じ
// リクエストを手前で拒否する。UseCase側に検査を残すことで、そのルート以外から到達した呼び出し元
// にも同じ権限を要求する。
type ArchivePersonUsecase struct {
	personRepo *repository.PersonRepository
}

// NewArchivePersonUsecaseは新しいArchivePersonUsecaseを作成する。
func NewArchivePersonUsecase(personRepo *repository.PersonRepository) *ArchivePersonUsecase {
	return &ArchivePersonUsecase{personRepo: personRepo}
}

// ArchivePersonInputはアーカイブする人物と、書き込みを認可するユーザーを指定する。
type ArchivePersonInput struct {
	User     *model.User
	PersonID model.PersonID
}

// ArchivePersonOutputはユースケースの出力。
type ArchivePersonOutput struct {
	PersonID model.PersonID
}

// Executeはcommitterを認可してから人物をアーカイブする。人物が存在しないか、アーカイブできる状態でない
// 場合はAppErrCodeResourceNotFoundの *model.AppErrorを返し、Handler側で404に変換する。
func (uc *ArchivePersonUsecase) Execute(ctx context.Context, input ArchivePersonInput) (*ArchivePersonOutput, error) {
	if input.User == nil || !input.User.IsCommitter() {
		return nil, &model.AppError{
			Code:     model.AppErrCodeForbidden,
			UserMsg:  i18n.T(ctx, "error_forbidden"),
			Metadata: map[string]string{"person_id": input.PersonID.String()},
		}
	}

	written, err := uc.personRepo.Archive(ctx, input.PersonID)
	if err != nil {
		return nil, fmt.Errorf("人物のアーカイブに失敗しました: %w", err)
	}
	if !written {
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_person_not_found"),
			Metadata: map[string]string{"person_id": input.PersonID.String()},
		}
	}

	return &ArchivePersonOutput{PersonID: input.PersonID}, nil
}
