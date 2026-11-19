package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBPersonRevisionsUsecaseはAnnict DB管理画面の人物の履歴に必要なデータ (人物と、その人物の版の1ページ分) を
// 取得するユースケース。作品の履歴と同じく、ログインしていなくても見られる。
//
// 人物の掲載停止は本人からの申出によるもので、掲載停止の事実を第三者に伝えない。そのため、掲載停止中の人物の
// 履歴は管理者にだけ見せ、掲載を再開したあとも、掲載停止・掲載再開の操作の版と掲載停止中に記録した版は
// 管理者にだけ見せる。
type GetDBPersonRevisionsUsecase struct {
	personRepo         *repository.PersonRepository
	personRevisionRepo *repository.PersonRevisionRepository
	prefectureRepo     *repository.PrefectureRepository
}

func NewGetDBPersonRevisionsUsecase(
	personRepo *repository.PersonRepository,
	personRevisionRepo *repository.PersonRevisionRepository,
	prefectureRepo *repository.PrefectureRepository,
) *GetDBPersonRevisionsUsecase {
	return &GetDBPersonRevisionsUsecase{
		personRepo:         personRepo,
		personRevisionRepo: personRevisionRepo,
		prefectureRepo:     prefectureRepo,
	}
}

type GetDBPersonRevisionsInput struct {
	// Userは閲覧者 (ログインしていないときはnil)。管理者かどうかで見せる版を決める。
	User     *model.User
	PersonID model.PersonID
	Page     int32
	PerPage  int32
}

type GetDBPersonRevisionsOutput struct {
	Person *model.Person
	// Revisionsは人物の版を新しい順に並べたもの。各版はDBEditとPreviousを持つ。
	Revisions  []*model.PersonRevision
	TotalCount int64
	// Prefecturesは、版の「出身地」をIDではなく名前で示すための都道府県の一覧。
	Prefectures []model.Prefecture
}

// Executeは人物と、その人物の版の1ページ分を返す。存在しない人物と削除した人物は誰にも見せず、掲載停止中の人物は
// 管理者以外に見せないため、AppErrCodeResourceNotFoundの *model.AppErrorを返す (Annict DBの人物の一覧と同じ)。
func (uc *GetDBPersonRevisionsUsecase) Execute(ctx context.Context, input GetDBPersonRevisionsInput) (*GetDBPersonRevisionsOutput, error) {
	isAdmin := input.User != nil && input.User.IsAdmin()

	person, err := uc.personRepo.GetForRevisionByID(ctx, input.PersonID)
	if err != nil {
		return nil, fmt.Errorf("人物の取得に失敗: %w", err)
	}
	if person == nil || person.Status == model.PersonStatusDeleted || (person.Status == model.PersonStatusDelisted && !isAdmin) {
		return nil, personNotFoundError(ctx, input.PersonID)
	}

	revisions, err := uc.personRevisionRepo.ListHistoryByPersonID(ctx, repository.PersonRevisionHistoryParams{
		PersonID:        input.PersonID,
		IncludeDelisted: isAdmin,
		Page:            input.Page,
		PerPage:         input.PerPage,
	})
	if err != nil {
		return nil, fmt.Errorf("人物の版の取得に失敗: %w", err)
	}
	totalCount, err := uc.personRevisionRepo.CountHistoryByPersonID(ctx, input.PersonID, isAdmin)
	if err != nil {
		return nil, fmt.Errorf("人物の版の数の取得に失敗: %w", err)
	}
	prefectures, err := uc.prefectureRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("出身地の取得に失敗: %w", err)
	}
	return &GetDBPersonRevisionsOutput{Person: person, Revisions: revisions, TotalCount: totalCount, Prefectures: prefectures}, nil
}
