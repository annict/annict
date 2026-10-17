package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBPeopleUsecaseはAnnict DB管理画面の人物一覧を取得するユースケース。
type GetDBPeopleUsecase struct {
	personRepo *repository.PersonRepository
}

// NewGetDBPeopleUsecaseは新しいGetDBPeopleUsecaseを作成する。
func NewGetDBPeopleUsecase(personRepo *repository.PersonRepository) *GetDBPeopleUsecase {
	return &GetDBPeopleUsecase{
		personRepo: personRepo,
	}
}

// GetDBPeopleInputはユースケースの入力。
type GetDBPeopleInput struct {
	Page    int32
	PerPage int32
}

// GetDBPeopleOutputはユースケースの出力。
type GetDBPeopleOutput struct {
	People     []*model.Person
	TotalCount int64
}

// ExecuteはAnnict DB管理画面の人物一覧と総数を取得する。一覧は誰でも閲覧でき、閲覧者に
// よらず公開中とアーカイブの人物だけを返す。
func (uc *GetDBPeopleUsecase) Execute(ctx context.Context, input GetDBPeopleInput) (*GetDBPeopleOutput, error) {
	params := repository.DBPersonListParams{
		Page:    input.Page,
		PerPage: input.PerPage,
	}

	people, err := uc.personRepo.ListForDB(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("DB人物一覧の取得に失敗: %w", err)
	}

	totalCount, err := uc.personRepo.CountForDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("DB人物総数の取得に失敗: %w", err)
	}

	return &GetDBPeopleOutput{
		People:     people,
		TotalCount: totalCount,
	}, nil
}
