package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/validator"
)

// CreatePeopleUsecaseは一括登録の1回の送信で作られる人物を作成する。1回の送信は複数行を
// 運び、それらをまとめて作成する (送信された全行が保存されるか、1行も保存されないかの
// どちらか)。Rails版は行ごとに保存しており、途中の行が失敗すると前の行だけが残っていた。
type CreatePeopleUsecase struct {
	db         *sql.DB
	personRepo *repository.PersonRepository
	validator  *validator.DBPersonCreateValidator
}

// NewCreatePeopleUsecaseはCreatePeopleUsecaseを生成する。
func NewCreatePeopleUsecase(
	db *sql.DB,
	personRepo *repository.PersonRepository,
	validator *validator.DBPersonCreateValidator,
) *CreatePeopleUsecase {
	return &CreatePeopleUsecase{
		db:         db,
		personRepo: personRepo,
		validator:  validator,
	}
}

// CreatePeopleInputは送信されたフォームを保持する。
type CreatePeopleInput struct {
	// Userは送信の帰属先となる編集者で、編集履歴 (db_activities) のuser_idになる。
	User *model.User
	// Rowsはtextareaに入力されたままの行 (1行1人)。
	Rows string
}

// CreatePeopleOutputは送信によって作成された人物を入力順で報告する。
type CreatePeopleOutput struct {
	PersonIDs []model.PersonID
}

// Executeは送信された全行の人物を作成する。認可を確認してから行を検証し、1つの
// トランザクションで作成する。作成した人物ごとに、編集履歴へpeople.createを記録する。
func (uc *CreatePeopleUsecase) Execute(ctx context.Context, input CreatePeopleInput) (*CreatePeopleOutput, error) {
	// 送信はcommitterに限られ、作成者に帰属するため、ユーザーを伴わない呼び出しと一般
	// ユーザーの呼び出しは何かを読む前に拒否する。
	if input.User == nil || !input.User.IsCommitter() {
		return nil, &model.AppError{
			Code:    model.AppErrCodeForbidden,
			UserMsg: i18n.T(ctx, "error_forbidden"),
		}
	}

	rows, err := uc.validator.Validate(ctx, validator.DBPersonCreateValidatorInput{Rows: input.Rows})
	if err != nil {
		return nil, err
	}

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	personRepo := uc.personRepo.WithTx(tx)
	personIDs := make([]model.PersonID, 0, len(rows))
	for i, row := range rows {
		personID, err := personRepo.Create(ctx, repository.CreatePersonParams{
			Name:     row.Name,
			NameKana: row.NameKana,
			UserID:   input.User.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("人物の作成に失敗しました (%d件目): %w", i+1, err)
		}
		personIDs = append(personIDs, personID)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return &CreatePeopleOutput{PersonIDs: personIDs}, nil
}
