package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBEpisodesUsecase is the use case for retrieving one work's episode list on the DB
// admin screen, together with the parent work the page's heading and subnav describe.
//
// [Ja] GetDBEpisodesUsecase は DB 管理画面の、ある作品のエピソード一覧を、ページの見出しと
// サブナビが示す親作品と併せて取得するユースケース。
type GetDBEpisodesUsecase struct {
	workRepo    *repository.WorkRepository
	episodeRepo *repository.EpisodeRepository
}

// NewGetDBEpisodesUsecase creates a new GetDBEpisodesUsecase.
//
// [Ja] 新しい GetDBEpisodesUsecase を作成する。
func NewGetDBEpisodesUsecase(workRepo *repository.WorkRepository, episodeRepo *repository.EpisodeRepository) *GetDBEpisodesUsecase {
	return &GetDBEpisodesUsecase{
		workRepo:    workRepo,
		episodeRepo: episodeRepo,
	}
}

// GetDBEpisodesInput is the input for the use case. Page is 1-based.
//
// [Ja] ユースケースの入力。Page は 1 始まり。
type GetDBEpisodesInput struct {
	WorkID  model.WorkID
	Page    int32
	PerPage int32
}

// GetDBEpisodesOutput is the output of the use case.
//
// [Ja] ユースケースの出力。
type GetDBEpisodesOutput struct {
	Work *model.Work
	// PublishedEpisodeCount and MaxGeneratableEpisodeNumber back the page's auto-generation
	// notice: how many of the work's episodes are published now, and how far the Syobocal
	// auto-generation could number them. TotalCount is the list's own total instead, which
	// keeps unpublished episodes because the list shows them.
	//
	// [Ja] PublishedEpisodeCount と MaxGeneratableEpisodeNumber はページの
	// 自動生成の案内に使う。
	// 作品のエピソードのうち現在公開中の件数と、しょぼいカレンダー由来の自動生成がどこまで
	// 話数を振れるかを表す。一方 TotalCount は一覧自体の総件数で、一覧が表示する非公開の
	// エピソードも含む。
	PublishedEpisodeCount       int64
	MaxGeneratableEpisodeNumber int64
	Episodes                    []*model.Episode
	TotalCount                  int64
}

// Execute returns the parent work and one page of its episodes. It returns a *model.AppError
// with AppErrCodeResourceNotFound when the work does not exist or is deleted; the handler
// converts that to 404.
//
// [Ja] Execute は親作品と、そのエピソード 1 ページ分を返す。作品が存在しない、または削除済み
// の場合は AppErrCodeResourceNotFound の *model.AppError を返し、Handler 側で 404 に変換する。
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
