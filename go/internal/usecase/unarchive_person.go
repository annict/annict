package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// UnarchivePersonUsecaseはAnnict DB管理画面から非公開の人物を公開に戻す。
// ArchivePersonUsecaseの逆で、people.statusをpublishedに戻し、unpublished_atをNULLに戻す。
// 公開に戻せるのは非公開の人物だけで、それ以外の状態の人物への送信はnot foundになる。db_activities
// には記録せず、Annict DBの編集履歴に記録する。
//
// 認可 (committer) は書き込みより先に本UseCaseで行い、ルートのRequireCommitter middlewareも同じ
// リクエストを手前で拒否する。UseCase側に検査を残すことで、そのルート以外から到達した呼び出し元
// にも同じ権限を要求する。
type UnarchivePersonUsecase struct {
	db                *sql.DB
	personRepo        *repository.PersonRepository
	dbEditRecordRepos DBEditRecordRepos
}

// NewUnarchivePersonUsecaseは新しいUnarchivePersonUsecaseを作成する。
func NewUnarchivePersonUsecase(db *sql.DB, personRepo *repository.PersonRepository, dbEditRecordRepos DBEditRecordRepos) *UnarchivePersonUsecase {
	return &UnarchivePersonUsecase{db: db, personRepo: personRepo, dbEditRecordRepos: dbEditRecordRepos}
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

	// 編集履歴の版は事前に読んだ人物から組み立てるため、書き込みはその版を照合する。読んだ後に別の編集が入った
	// 人物と、状態を変えられない人物には書き込まない。
	before, err := uc.personRepo.GetForRevisionByID(ctx, input.PersonID)
	if err != nil {
		return nil, fmt.Errorf("人物の取得に失敗しました: %w", err)
	}
	written := false
	if before != nil && before.Status == model.PersonStatusArchived {
		recordPlan, err := planPersonStatusRecord(ctx, uc.dbEditRecordRepos, input.User, model.DBEditActionUnarchive, before, model.PersonStatusPublished)
		if err != nil {
			return nil, err
		}
		written, err = writePersonWithRecord(ctx, uc.db, uc.personRepo, uc.dbEditRecordRepos, recordPlan, func(repo *repository.PersonRepository) (bool, error) {
			return repo.Unarchive(ctx, input.PersonID, before.UpdatedAt)
		})
		if err != nil {
			return nil, fmt.Errorf("人物のアーカイブ解除に失敗しました: %w", err)
		}
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
