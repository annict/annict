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
	episodeRepo     *repository.EpisodeRepository
	featureFlagRepo *repository.FeatureFlagRepository
}

// NewGetDBEpisodeEditUsecaseは新しいGetDBEpisodeEditUsecaseを作成する。
func NewGetDBEpisodeEditUsecase(episodeRepo *repository.EpisodeRepository, featureFlagRepo *repository.FeatureFlagRepository) *GetDBEpisodeEditUsecase {
	return &GetDBEpisodeEditUsecase{episodeRepo: episodeRepo, featureFlagRepo: featureFlagRepo}
}

// GetDBEpisodeEditInputはユースケースの入力。
type GetDBEpisodeEditInput struct {
	User      *model.User
	EpisodeID model.EpisodeID
	// DeviceTokenは閲覧中のデバイスのトークン (Cookieが無いときは空文字列)。履歴へのリンクを出すかの判定に使う。
	DeviceToken string
}

// GetDBEpisodeEditOutputはユースケースの出力。
type GetDBEpisodeEditOutput struct {
	Episode *model.Episode
	Work    *model.Work
	// CanShowRevisionsはエピソードの編集画面の見出しに、そのエピソードの履歴へのリンクを出すかを表す。
	CanShowRevisions bool
}

// Executeはエピソードを取得する前にコミッターを認可し、編集対象のエピソードとその親作品を
// 返す。エピソードが存在しない、削除済み、または削除済み作品に属する場合は
// AppErrCodeResourceNotFoundの *model.AppErrorを返し、Handler側で404に変換する。
func (uc *GetDBEpisodeEditUsecase) Execute(ctx context.Context, input GetDBEpisodeEditInput) (*GetDBEpisodeEditOutput, error) {
	if input.User == nil || !input.User.IsCommitter() {
		return nil, &model.AppError{
			Code:     model.AppErrCodeForbidden,
			UserMsg:  i18n.T(ctx, "error_forbidden"),
			Metadata: map[string]string{"episode_id": input.EpisodeID.String()},
		}
	}

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

	canShowRevisions, err := canShowDBRevisions(ctx, uc.featureFlagRepo, input.DeviceToken, input.User)
	if err != nil {
		return nil, err
	}

	return &GetDBEpisodeEditOutput{
		Episode:          target.Episode,
		Work:             target.Work,
		CanShowRevisions: canShowRevisions,
	}, nil
}
