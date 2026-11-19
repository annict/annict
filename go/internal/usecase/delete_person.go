package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// DeletePersonUsecaseはAnnict DB管理画面から人物を論理削除する。Rails版の物理削除
// (Person#destroy_in_batches) と異なり、people.statusをdeletedにする (Rails版への橋渡しの
// deleted_atも書く) だけで、キャスト・スタッフ・編集履歴は残す (作品・エピソードの削除と同じ)。
// 公開中・アーカイブの人物を削除でき、それ以外の状態の人物への送信はnot foundになる。db_activities
// には記録せず (Rails版と同じ)、Annict DBの編集履歴に記録する。
//
// 認可 (admin) は書き込みより先に本UseCaseで行い、ルートのRequireAdmin middlewareも同じ
// リクエストを手前で拒否する。UseCase側に検査を残すことで、そのルート以外から到達した呼び出し元
// にも同じ権限を要求する。
type DeletePersonUsecase struct {
	db                *sql.DB
	personRepo        *repository.PersonRepository
	dbEditRecordRepos DBEditRecordRepos
}

// NewDeletePersonUsecaseは新しいDeletePersonUsecaseを作成する。
func NewDeletePersonUsecase(db *sql.DB, personRepo *repository.PersonRepository, dbEditRecordRepos DBEditRecordRepos) *DeletePersonUsecase {
	return &DeletePersonUsecase{db: db, personRepo: personRepo, dbEditRecordRepos: dbEditRecordRepos}
}

// DeletePersonInputは削除する人物と、書き込みを認可するユーザーを指定する。
type DeletePersonInput struct {
	User     *model.User
	PersonID model.PersonID
}

// DeletePersonOutputはユースケースの出力。
type DeletePersonOutput struct {
	PersonID model.PersonID
}

// Executeはadminを認可してから人物を削除する。人物が存在しないか、削除できる状態でない
// 場合はAppErrCodeResourceNotFoundの *model.AppErrorを返し、Handler側で404に変換する。
func (uc *DeletePersonUsecase) Execute(ctx context.Context, input DeletePersonInput) (*DeletePersonOutput, error) {
	if input.User == nil || !input.User.IsAdmin() {
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
	if before != nil && (before.Status == model.PersonStatusPublished || before.Status == model.PersonStatusArchived) {
		recordPlan, err := planPersonStatusRecord(ctx, uc.dbEditRecordRepos, input.User, model.DBEditActionDelete, before, model.PersonStatusDeleted)
		if err != nil {
			return nil, err
		}
		written, err = writePersonWithRecord(ctx, uc.db, uc.personRepo, uc.dbEditRecordRepos, recordPlan, func(repo *repository.PersonRepository) (bool, error) {
			return repo.Delete(ctx, input.PersonID, before.UpdatedAt)
		})
		if err != nil {
			return nil, fmt.Errorf("人物の削除に失敗しました: %w", err)
		}
	}
	if !written {
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_person_not_found"),
			Metadata: map[string]string{"person_id": input.PersonID.String()},
		}
	}

	return &DeletePersonOutput{PersonID: input.PersonID}, nil
}
