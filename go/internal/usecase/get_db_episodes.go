package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBEpisodesUsecaseはDB管理画面の、ある作品のエピソード一覧を、ページの見出しと
// サブナビが示す親作品と併せて取得するユースケース。
type GetDBEpisodesUsecase struct {
	workRepo    *repository.WorkRepository
	episodeRepo *repository.EpisodeRepository
}

// NewGetDBEpisodesUsecaseは新しいGetDBEpisodesUsecaseを作成する。
func NewGetDBEpisodesUsecase(workRepo *repository.WorkRepository, episodeRepo *repository.EpisodeRepository) *GetDBEpisodesUsecase {
	return &GetDBEpisodesUsecase{
		workRepo:    workRepo,
		episodeRepo: episodeRepo,
	}
}

// GetDBEpisodesInputはユースケースの入力。Pageは1始まり。
type GetDBEpisodesInput struct {
	WorkID  model.WorkID
	Page    int32
	PerPage int32
}

// GetDBEpisodesOutputはユースケースの出力。
type GetDBEpisodesOutput struct {
	Work *model.Work
	// PublishedEpisodeCountとMaxGeneratableEpisodeNumberはページの
	// 自動生成の案内に使う。
	// 作品のエピソードのうち現在公開中の件数と、しょぼいカレンダー由来の自動生成がどこまで
	// 話数を振れるかを表す。一方TotalCountは一覧自体の総件数で、一覧が表示する非公開の
	// エピソードも含む。
	PublishedEpisodeCount       int64
	MaxGeneratableEpisodeNumber int64
	Episodes                    []*model.Episode
	TotalCount                  int64
}

// Executeは親作品と、そのエピソード1ページ分を返す。作品が存在しない、または削除済み
// の場合はAppErrCodeResourceNotFoundの *model.AppErrorを返し、Handler側で404に変換する。
func (uc *GetDBEpisodesUsecase) Execute(ctx context.Context, input GetDBEpisodesInput) (*GetDBEpisodesOutput, error) {
	listWork, err := uc.workRepo.GetForEpisodeListByID(ctx, input.WorkID)
	if err != nil {
		return nil, fmt.Errorf("作品の取得に失敗: %w", err)
	}
	if listWork == nil {
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_work_not_found"),
			Metadata: map[string]string{"work_id": input.WorkID.String()},
		}
	}

	params := repository.DBEpisodeListParams{
		WorkID:  input.WorkID,
		Page:    input.Page,
		PerPage: input.PerPage,
	}

	episodes, err := uc.episodeRepo.ListForDB(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("DBエピソード一覧の取得に失敗: %w", err)
	}

	totalCount, err := uc.episodeRepo.CountForDB(ctx, input.WorkID)
	if err != nil {
		return nil, fmt.Errorf("DBエピソード総数の取得に失敗: %w", err)
	}

	return &GetDBEpisodesOutput{
		Work:                        listWork.Work,
		PublishedEpisodeCount:       listWork.PublishedEpisodeCount,
		MaxGeneratableEpisodeNumber: listWork.MaxGeneratableEpisodeNumber,
		Episodes:                    episodes,
		TotalCount:                  totalCount,
	}, nil
}
