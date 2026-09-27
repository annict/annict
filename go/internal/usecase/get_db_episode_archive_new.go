package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBEpisodeArchiveNewUsecaseはAnnict DBの非公開確認画面に必要なデータ (非公開を
// 確認する対象のエピソードと、その見出し・サブナビが示す親作品) を取得する。非公開にできるのは
// 現在公開中のエピソードだけで、これはRailsのscope Episode.without_deleted.publishedに
// 一致する。すでに非公開のエピソードはnot foundとして扱う。
type GetDBEpisodeArchiveNewUsecase struct {
	episodeRepo *repository.EpisodeRepository
}

// NewGetDBEpisodeArchiveNewUsecaseは新しいGetDBEpisodeArchiveNewUsecaseを作成する。
func NewGetDBEpisodeArchiveNewUsecase(episodeRepo *repository.EpisodeRepository) *GetDBEpisodeArchiveNewUsecase {
	return &GetDBEpisodeArchiveNewUsecase{episodeRepo: episodeRepo}
}

// GetDBEpisodeArchiveNewInputはユースケースの入力。
type GetDBEpisodeArchiveNewInput struct {
	EpisodeID model.EpisodeID
}

// GetDBEpisodeArchiveNewOutputはユースケースの出力。
type GetDBEpisodeArchiveNewOutput struct {
	Episode *model.Episode
	Work    *model.Work
}

// Executeは非公開を確認する対象のエピソードとその親作品を返す。エピソードが存在しない、
// 削除済み、削除済み作品に属する、または現在公開中でない場合はAppErrCodeResourceNotFoundの
// *model.AppErrorを返し、Handler側で404に変換する。
func (uc *GetDBEpisodeArchiveNewUsecase) Execute(ctx context.Context, input GetDBEpisodeArchiveNewInput) (*GetDBEpisodeArchiveNewOutput, error) {
	target, err := uc.episodeRepo.GetForArchiveByID(ctx, input.EpisodeID)
	if err != nil {
		return nil, fmt.Errorf("エピソードの取得に失敗: %w", err)
	}
	if target == nil || target.Episode.DerivedStatus() != model.EpisodeStatusPublished {
		return nil, episodeNotFoundError(ctx, input.EpisodeID)
	}

	return &GetDBEpisodeArchiveNewOutput{
		Episode: target.Episode,
		Work:    target.Work,
	}, nil
}

// episodeNotFoundErrorはHandlerが404に写像するリソース未存在エラーを組み立てる。
// 非公開の確認ページと、それに続く送信の双方が送出するため、両者の間に非公開にできる状態から
// 外れたエピソードは、どちらでも同じ形で報告される。
func episodeNotFoundError(ctx context.Context, episodeID model.EpisodeID) *model.AppError {
	return &model.AppError{
		Code:     model.AppErrCodeResourceNotFound,
		UserMsg:  i18n.T(ctx, "error_episode_not_found"),
		Metadata: map[string]string{"episode_id": episodeID.String()},
	}
}
