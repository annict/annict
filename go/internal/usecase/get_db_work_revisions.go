package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBWorkRevisionsUsecaseはAnnict DB管理画面の作品の履歴に必要なデータ (作品と、その作品と配下の
// エピソードの版の1ページ分) を取得するユースケース。履歴はRails版の /db/activitiesと同じく、ログインしていなくても見られる。
type GetDBWorkRevisionsUsecase struct {
	workRepo          *repository.WorkRepository
	animeRevisionRepo *repository.AnimeRevisionRepository
	numberFormatRepo  *repository.NumberFormatRepository
}

func NewGetDBWorkRevisionsUsecase(
	workRepo *repository.WorkRepository,
	animeRevisionRepo *repository.AnimeRevisionRepository,
	numberFormatRepo *repository.NumberFormatRepository,
) *GetDBWorkRevisionsUsecase {
	return &GetDBWorkRevisionsUsecase{
		workRepo:          workRepo,
		animeRevisionRepo: animeRevisionRepo,
		numberFormatRepo:  numberFormatRepo,
	}
}

type GetDBWorkRevisionsInput struct {
	WorkID  model.WorkID
	Page    int32
	PerPage int32
}

type GetDBWorkRevisionsOutput struct {
	Work *model.Work
	// Revisionsは作品と配下のエピソードの版を新しい順に並べたもの。各版はDBEditとPreviousを持ち、
	// エピソードの版はEpisodeも持つ。
	Revisions  []*model.AnimeRevision
	TotalCount int64
	// NumberFormatsは、版の「話数フォーマット」をIDではなく名前で示すための話数フォーマットの一覧。
	NumberFormats []model.NumberFormat
}

// Executeは作品と、その作品と配下のエピソードの版の1ページ分を返す。存在しない作品と削除した作品は、Annict DBの一覧と
// 同じく誰にも見せないため、AppErrCodeResourceNotFoundの *model.AppErrorを返す。animeにまだ
// マッピングされていない作品は版を持たないため、空の履歴を返す。
func (uc *GetDBWorkRevisionsUsecase) Execute(ctx context.Context, input GetDBWorkRevisionsInput) (*GetDBWorkRevisionsOutput, error) {
	works, err := uc.workRepo.ListForAnimeSyncByIDs(ctx, []model.WorkID{input.WorkID})
	if err != nil {
		return nil, fmt.Errorf("作品の取得に失敗: %w", err)
	}
	if len(works) == 0 || works[0].DerivedStatus() == model.WorkStatusDeleted {
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_work_not_found"),
			Metadata: map[string]string{"work_id": input.WorkID.String()},
		}
	}
	work := works[0]

	output := &GetDBWorkRevisionsOutput{Work: work, Revisions: []*model.AnimeRevision{}}
	if work.AnimeID == nil {
		return output, nil
	}

	output.Revisions, err = uc.animeRevisionRepo.ListHistoryByWorkAnimeID(ctx, repository.AnimeRevisionHistoryParams{
		AnimeID: *work.AnimeID,
		Page:    input.Page,
		PerPage: input.PerPage,
	})
	if err != nil {
		return nil, fmt.Errorf("作品の版の取得に失敗: %w", err)
	}
	output.TotalCount, err = uc.animeRevisionRepo.CountHistoryByWorkAnimeID(ctx, *work.AnimeID)
	if err != nil {
		return nil, fmt.Errorf("作品の版の数の取得に失敗: %w", err)
	}
	output.NumberFormats, err = uc.numberFormatRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("話数フォーマットの取得に失敗: %w", err)
	}
	return output, nil
}
