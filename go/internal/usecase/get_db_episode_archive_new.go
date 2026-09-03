package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBEpisodeArchiveNewUsecase loads the data the Annict DB archive-confirmation screen
// needs: the episode whose archiving is being confirmed, together with the parent work its
// heading and subnav describe. Only a currently published episode is archivable, matching the
// Rails scope Episode.without_deleted.published, so an already-archived episode is reported as
// not found.
//
// [Ja] GetDBEpisodeArchiveNewUsecase は Annict DB の非公開確認画面に必要なデータ (非公開を
// 確認する対象のエピソードと、その見出し・サブナビが示す親作品) を取得する。非公開にできるのは
// 現在公開中のエピソードだけで、これは Rails の scope Episode.without_deleted.published に
// 一致する。すでに非公開のエピソードは not found として扱う。
type GetDBEpisodeArchiveNewUsecase struct {
	episodeRepo *repository.EpisodeRepository
}

// NewGetDBEpisodeArchiveNewUsecase creates a new GetDBEpisodeArchiveNewUsecase.
//
// [Ja] 新しい GetDBEpisodeArchiveNewUsecase を作成する。
func NewGetDBEpisodeArchiveNewUsecase(episodeRepo *repository.EpisodeRepository) *GetDBEpisodeArchiveNewUsecase {
	return &GetDBEpisodeArchiveNewUsecase{episodeRepo: episodeRepo}
}

// GetDBEpisodeArchiveNewInput is the input for the use case.
//
// [Ja] ユースケースの入力。
type GetDBEpisodeArchiveNewInput struct {
	EpisodeID model.EpisodeID
}

// GetDBEpisodeArchiveNewOutput is the output of the use case.
//
// [Ja] ユースケースの出力。
type GetDBEpisodeArchiveNewOutput struct {
	Episode *model.Episode
	Work    *model.Work
}

// Execute returns the episode to confirm archiving for and its parent work. It returns a
// *model.AppError with AppErrCodeResourceNotFound when the episode does not exist, is deleted,
// belongs to a deleted work, or is not currently published; the handler converts that to 404.
//
// [Ja] Execute は非公開を確認する対象のエピソードとその親作品を返す。エピソードが存在しない、
// 削除済み、削除済み作品に属する、または現在公開中でない場合は AppErrCodeResourceNotFound の
// *model.AppError を返し、Handler 側で 404 に変換する。
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

// episodeNotFoundError builds the resource-not-found error the handler maps to a 404. The
// archive confirmation page and the submit that follows it both raise it, so an episode that
// left the archivable state between the two is reported the same way in either place.
//
// [Ja] episodeNotFoundError は Handler が 404 に写像するリソース未存在エラーを組み立てる。
// 非公開の確認ページと、それに続く送信の双方が送出するため、両者の間に非公開にできる状態から
// 外れたエピソードは、どちらでも同じ形で報告される。
func episodeNotFoundError(ctx context.Context, episodeID model.EpisodeID) *model.AppError {
	return &model.AppError{
		Code:     model.AppErrCodeResourceNotFound,
		UserMsg:  i18n.T(ctx, "error_episode_not_found"),
		Metadata: map[string]string{"episode_id": episodeID.String()},
	}
}
