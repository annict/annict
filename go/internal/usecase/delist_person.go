package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/validator"
)

// DelistPersonUsecaseはAnnict DB管理画面から、本人の申出に応じて人物の掲載を停止する。
// people.statusをdelistedにして対応記録 (delist_message) を書き、Rails版の表示とAPIから除外
// されるよう橋渡しのunpublished_atも書く。キャスト・スタッフは書き換えず、db_activitiesにも記録しない。
// Annict DBの編集履歴には記録し、対応記録は版に入れない。
// 掲載を停止できるのは公開中とアーカイブの人物で、状態と確認時の版の照合を書き込みと同じ文で
// 行う。確認後に編集された人物への送信は競合になり、停止できない状態になった人物はnot foundになる。
//
// 認可 (admin) は書き込みより先に本UseCaseで行い、ルートのRequireAdmin middlewareも同じ
// リクエストを手前で拒否する。UseCase側に検査を残すことで、そのルート以外から到達した呼び出し元
// にも同じ権限を要求する。
type DelistPersonUsecase struct {
	db                *sql.DB
	personRepo        *repository.PersonRepository
	dbEditRecordRepos DBEditRecordRepos
}

// NewDelistPersonUsecaseは新しいDelistPersonUsecaseを作成する。
func NewDelistPersonUsecase(db *sql.DB, personRepo *repository.PersonRepository, dbEditRecordRepos DBEditRecordRepos) *DelistPersonUsecase {
	return &DelistPersonUsecase{db: db, personRepo: personRepo, dbEditRecordRepos: dbEditRecordRepos}
}

// DelistPersonInputは掲載を停止する人物と確認時の版、対応記録、認可するユーザーを指定する。
type DelistPersonInput struct {
	User      *model.User
	PersonID  model.PersonID
	UpdatedAt string
	// DelistMessageは確認画面で入力された対応記録。前後の空白を除いて空なら、記録なし (NULL) として書く。
	DelistMessage string
}

// DelistPersonOutputはユースケースの出力。
type DelistPersonOutput struct {
	PersonID model.PersonID
}

// Executeはadminを認可してから人物の掲載を停止する。人物が存在しないか、掲載を停止できる状態でない
// 場合はAppErrCodeResourceNotFoundの *model.AppErrorを返し、Handler側で404に変換する。
// 確認後に版が進んだ場合はAppErrCodeConflictを返す。
func (uc *DelistPersonUsecase) Execute(ctx context.Context, input DelistPersonInput) (*DelistPersonOutput, error) {
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

	var delistMessage *string
	if message := strings.TrimSpace(input.DelistMessage); message != "" {
		delistMessage = &message
	}

	written, err := uc.delist(ctx, input, version, delistMessage)
	if err != nil {
		return nil, err
	}
	if !written {
		current, err := uc.personRepo.GetForDelistingByID(ctx, input.PersonID)
		if err != nil {
			return nil, fmt.Errorf("人物の取得に失敗しました: %w", err)
		}
		if current != nil && current.Status.IsViewable() {
			return nil, personConflictError(ctx, input.PersonID)
		}
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_person_not_found"),
			Metadata: map[string]string{"person_id": input.PersonID.String()},
		}
	}

	return &DelistPersonOutput{PersonID: input.PersonID}, nil
}

// delistは、掲載停止を編集履歴の記録と同じトランザクションで書く。編集履歴の版は事前に読んだ人物から組み立て、
// 書き込みは確認時の版を照合するため、読んだ人物が確認時の版と違えば書き込まない。掲載を停止できない人物と、
// 確認時の版が進んでいた人物ではfalseを返す。
func (uc *DelistPersonUsecase) delist(ctx context.Context, input DelistPersonInput, version time.Time, delistMessage *string) (bool, error) {
	before, err := uc.personRepo.GetForRevisionByID(ctx, input.PersonID)
	if err != nil {
		return false, fmt.Errorf("人物の取得に失敗しました: %w", err)
	}
	if before == nil || (before.Status != model.PersonStatusPublished && before.Status != model.PersonStatusArchived) {
		return false, nil
	}
	recordPlan, err := planPersonStatusRecord(ctx, uc.dbEditRecordRepos, input.User, model.DBEditActionDelist, before, model.PersonStatusDelisted)
	if err != nil {
		return false, err
	}
	written, err := writePersonWithRecord(ctx, uc.db, uc.personRepo, uc.dbEditRecordRepos, recordPlan, func(repo *repository.PersonRepository) (bool, error) {
		return repo.Delist(ctx, input.PersonID, version, delistMessage)
	})
	if err != nil {
		return false, fmt.Errorf("人物の掲載停止に失敗しました: %w", err)
	}
	return written, nil
}
