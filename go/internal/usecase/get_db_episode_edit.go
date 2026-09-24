package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBEpisodeEditUsecaseはDB管理画面のエピソード編集フォームのユースケース。
// フォームの初期値になる保存済みのエピソードと、ページの見出しとサブナビが示す親作品を
// 取得する。
type GetDBEpisodeEditUsecase struct {
	episodeRepo *repository.EpisodeRepository
}

// NewGetDBEpisodeEditUsecaseは新しいGetDBEpisodeEditUsecaseを作成する。
func NewGetDBEpisodeEditUsecase(episodeRepo *repository.EpisodeRepository) *GetDBEpisodeEditUsecase {
	return &GetDBEpisodeEditUsecase{episodeRepo: episodeRepo}
}

// GetDBEpisodeEditInputはユースケースの入力。
type GetDBEpisodeEditInput struct {
	EpisodeID model.EpisodeID
}

// GetDBEpisodeEditOutputはユースケースの出力。
type GetDBEpisodeEditOutput struct {
	Episode *model.Episode
	Work    *model.Work
}

// Executeは編集対象のエピソードとその親作品を返す。エピソードが存在しない、削除済み、
// または削除済み作品に属する場合はAppErrCodeResourceNotFoundの *model.AppErrorを返し、
// Handler側で404に変換する。
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
