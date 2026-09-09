package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBEpisodeEditUsecase is the use case for the episode edit form on the DB admin screen.
// It retrieves the episode whose stored values the form starts from, together with the
// parent work the page's heading and subnav describe.
//
// [Ja] GetDBEpisodeEditUsecase は DB 管理画面のエピソード編集フォームのユースケース。
// フォームの初期値になる保存済みのエピソードと、ページの見出しとサブナビが示す親作品を
// 取得する。
type GetDBEpisodeEditUsecase struct {
	episodeRepo *repository.EpisodeRepository
}

// NewGetDBEpisodeEditUsecase creates a new GetDBEpisodeEditUsecase.
//
// [Ja] 新しい GetDBEpisodeEditUsecase を作成する。
func NewGetDBEpisodeEditUsecase(episodeRepo *repository.EpisodeRepository) *GetDBEpisodeEditUsecase {
	return &GetDBEpisodeEditUsecase{episodeRepo: episodeRepo}
}

// GetDBEpisodeEditInput is the input for the use case.
//
// [Ja] ユースケースの入力。
type GetDBEpisodeEditInput struct {
	EpisodeID model.EpisodeID
}

// GetDBEpisodeEditOutput is the output of the use case.
//
// [Ja] ユースケースの出力。
type GetDBEpisodeEditOutput struct {
	Episode *model.Episode
	Work    *model.Work
}

// Execute returns the episode to edit and its parent work. It returns a *model.AppError with
// AppErrCodeResourceNotFound when the episode does not exist, is deleted, or belongs to a
// deleted work; the handler converts that to 404.
//
// [Ja] Execute は編集対象のエピソードとその親作品を返す。エピソードが存在しない、削除済み、
// または削除済み作品に属する場合は AppErrCodeResourceNotFound の *model.AppError を返し、
// Handler 側で 404 に変換する。
func (uc *GetDBEpisodeEditUsecase) Execute(ctx context.Context, input GetDBEpisodeEditInput) (*GetDBEpisodeEditOutput, error) {
	target, err := uc.episodeRepo.GetForEditByID(ctx, input.EpisodeID)
	if err != nil {
		return nil, fmt.Errorf("エピソードの取得に失敗: %w", err)
	}
	if target == nil {
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_episode_not_found"),
			Metadata: map[string]string{"episode_id": input.EpisodeID.String()},
		}
	}

	return &GetDBEpisodeEditOutput{
		Episode: target.Episode,
		Work:    target.Work,
	}, nil
}
