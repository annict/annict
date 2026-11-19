package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBPeopleUsecaseはAnnict DB管理画面の人物一覧を取得するユースケース。
type GetDBPeopleUsecase struct {
	personRepo      *repository.PersonRepository
	featureFlagRepo *repository.FeatureFlagRepository
}

// NewGetDBPeopleUsecaseは新しいGetDBPeopleUsecaseを作成する。
func NewGetDBPeopleUsecase(personRepo *repository.PersonRepository, featureFlagRepo *repository.FeatureFlagRepository) *GetDBPeopleUsecase {
	return &GetDBPeopleUsecase{
		personRepo:      personRepo,
		featureFlagRepo: featureFlagRepo,
	}
}

// GetDBPeopleInputはユースケースの入力。
type GetDBPeopleInput struct {
	// Userは一覧の閲覧者で、未ログインならnil。掲載停止の人物を一覧に含めるかと、履歴へのリンクを出すかを決める。
	User    *model.User
	Page    int32
	PerPage int32
	// DeviceTokenは閲覧中のデバイスのトークン (Cookieが無いときは空文字列)。履歴へのリンクを出すかの判定に使う。
	DeviceToken string
}

// GetDBPeopleOutputはユースケースの出力。
type GetDBPeopleOutput struct {
	People     []*model.Person
	TotalCount int64
	// CanShowRevisionsは各行に人物の履歴へのリンクを出すかを表す。
	CanShowRevisions bool
}

// ExecuteはAnnict DB管理画面の人物一覧と総数を取得する。一覧は誰でも閲覧でき、公開中と
// アーカイブの人物を返す。掲載停止の人物は、その設定と解除を担う管理者にだけ返す。
func (uc *GetDBPeopleUsecase) Execute(ctx context.Context, input GetDBPeopleInput) (*GetDBPeopleOutput, error) {
	params := repository.DBPersonListParams{
		Page:            input.Page,
		PerPage:         input.PerPage,
		IncludeDelisted: input.User != nil && input.User.IsAdmin(),
	}

	people, err := uc.personRepo.ListForDB(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("DB人物一覧の取得に失敗: %w", err)
	}

	totalCount, err := uc.personRepo.CountForDB(ctx, params.IncludeDelisted)
	if err != nil {
		return nil, fmt.Errorf("DB人物総数の取得に失敗: %w", err)
	}

	canShowRevisions, err := canShowDBRevisions(ctx, uc.featureFlagRepo, input.DeviceToken, input.User)
	if err != nil {
		return nil, err
	}

	return &GetDBPeopleOutput{
		People:           people,
		TotalCount:       totalCount,
		CanShowRevisions: canShowRevisions,
	}, nil
}
