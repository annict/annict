package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBWorksUsecaseはDB管理画面の作品一覧を取得するユースケース。
type GetDBWorksUsecase struct {
	workRepo        *repository.WorkRepository
	featureFlagRepo *repository.FeatureFlagRepository
}

// NewGetDBWorksUsecaseは新しいGetDBWorksUsecaseを作成する。
func NewGetDBWorksUsecase(workRepo *repository.WorkRepository, featureFlagRepo *repository.FeatureFlagRepository) *GetDBWorksUsecase {
	return &GetDBWorksUsecase{
		workRepo:        workRepo,
		featureFlagRepo: featureFlagRepo,
	}
}

// GetDBWorksInputはユースケースの入力。
type GetDBWorksInput struct {
	FilterNoEpisodes bool
	FilterNoImage    bool
	FilterNoSeason   bool
	FilterNoSlots    bool
	// SeasonYears / SeasonNamesはリリース時期の複数選択フィルタの並列 (年, 季節)
	// ペア (空でフィルタ無効)。SeasonNamesの0は「季節未登録」を表す。
	SeasonYears []int32
	SeasonNames []int32
	Page        int32
	PerPage     int32
	// Userは閲覧者 (ログインしていないときはnil)。履歴へのリンクを出すかの判定に使う。
	User *model.User
	// DeviceTokenは閲覧中のデバイスのトークン (Cookieが無いときは空文字列)。履歴へのリンクを出すかの判定に使う。
	DeviceToken string
}

// GetDBWorksOutputはユースケースの出力。
type GetDBWorksOutput struct {
	Works      []*model.Work
	TotalCount int64
	// CanShowRevisionsは各行に作品の履歴へのリンクを出すかを表す。
	CanShowRevisions bool
}

// ExecuteはDB管理画面の作品一覧と総数を取得する。
func (uc *GetDBWorksUsecase) Execute(ctx context.Context, input GetDBWorksInput) (*GetDBWorksOutput, error) {
	params := repository.DBWorkListParams{
		FilterNoEpisodes: input.FilterNoEpisodes,
		FilterNoImage:    input.FilterNoImage,
		FilterNoSeason:   input.FilterNoSeason,
		FilterNoSlots:    input.FilterNoSlots,
		SeasonYears:      input.SeasonYears,
		SeasonNames:      input.SeasonNames,
		Page:             input.Page,
		PerPage:          input.PerPage,
	}

	works, err := uc.workRepo.ListForDB(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("DB作品一覧の取得に失敗: %w", err)
	}

	totalCount, err := uc.workRepo.CountForDB(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("DB作品総数の取得に失敗: %w", err)
	}

	canShowRevisions, err := canShowDBRevisions(ctx, uc.featureFlagRepo, input.DeviceToken, input.User)
	if err != nil {
		return nil, err
	}

	return &GetDBWorksOutput{
		Works:            works,
		TotalCount:       totalCount,
		CanShowRevisions: canShowRevisions,
	}, nil
}
