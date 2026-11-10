package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBEpisodeRevisionsUsecaseはAnnict DB管理画面のエピソードの履歴に必要なデータ (エピソードと親作品、
// そのエピソードの版の1ページ分) を取得するユースケース。作品の履歴と同じく、ログインしていなくても見られる。
type GetDBEpisodeRevisionsUsecase struct {
	episodeRepo       *repository.EpisodeRepository
	animeRevisionRepo *repository.AnimeRevisionRepository
}

func NewGetDBEpisodeRevisionsUsecase(
	episodeRepo *repository.EpisodeRepository,
	animeRevisionRepo *repository.AnimeRevisionRepository,
) *GetDBEpisodeRevisionsUsecase {
	return &GetDBEpisodeRevisionsUsecase{
		episodeRepo:       episodeRepo,
		animeRevisionRepo: animeRevisionRepo,
	}
}

type GetDBEpisodeRevisionsInput struct {
	EpisodeID model.EpisodeID
	Page      int32
	PerPage   int32
}

type GetDBEpisodeRevisionsOutput struct {
	Episode *model.Episode
	// Workはエピソードの親作品で、見出しとサブナビに使う。
	Work *model.Work
	// Revisionsはエピソードの版を新しい順に並べたもの。各版はDBEditとPreviousを持つ。
	Revisions  []*model.AnimeRevision
	TotalCount int64
}

// Executeはエピソードと、そのエピソードの版の1ページ分を返す。エピソードの編集画面と同じく、存在しない
// エピソード・削除したエピソード・削除した作品のエピソードは誰にも見せないため、AppErrCodeResourceNotFoundの
// *model.AppErrorを返す。animeにまだマッピングされていないエピソードは版を持たないため、空の履歴を返す。
func (uc *GetDBEpisodeRevisionsUsecase) Execute(ctx context.Context, input GetDBEpisodeRevisionsInput) (*GetDBEpisodeRevisionsOutput, error) {
	target, err := uc.episodeRepo.GetForEditByID(ctx, input.EpisodeID)
	if err != nil {
		return nil, fmt.Errorf("エピソードの取得に失敗: %w", err)
	}
	if target == nil {
		return nil, episodeNotFoundError(ctx, input.EpisodeID)
	}

	output := &GetDBEpisodeRevisionsOutput{Episode: target.Episode, Work: target.Work, Revisions: []*model.AnimeRevision{}}
	if target.Episode.AnimeID == nil {
		return output, nil
	}

	output.Revisions, err = uc.animeRevisionRepo.ListHistoryByAnimeID(ctx, repository.AnimeRevisionHistoryParams{
		AnimeID: *target.Episode.AnimeID,
		Page:    input.Page,
		PerPage: input.PerPage,
	})
	if err != nil {
		return nil, fmt.Errorf("エピソードの版の取得に失敗: %w", err)
	}
	output.TotalCount, err = uc.animeRevisionRepo.CountHistoryByAnimeID(ctx, *target.Episode.AnimeID)
	if err != nil {
		return nil, fmt.Errorf("エピソードの版の数の取得に失敗: %w", err)
	}
	return output, nil
}
