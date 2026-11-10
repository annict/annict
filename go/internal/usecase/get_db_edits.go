package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// GetDBEditsUsecaseはAnnict DB全体の編集履歴に必要なデータ (操作の1ページ分と、各操作の版) を取得する
// ユースケース。作品の履歴と同じく、ログインしていなくても見られる。
//
// 見せる版は各リソースの履歴と同じで、削除した作品・エピソード・人物の版は誰にも見せない。人物の掲載停止に
// かかわる版 (掲載停止中の人物の版と、掲載停止・掲載再開の操作の版と、掲載停止中に記録した版) は管理者にだけ
// 見せる。見せる版が1つも残らない操作は並べない。
type GetDBEditsUsecase struct {
	dbEditRepo         *repository.DBEditRepository
	animeRevisionRepo  *repository.AnimeRevisionRepository
	personRevisionRepo *repository.PersonRevisionRepository
	numberFormatRepo   *repository.NumberFormatRepository
	prefectureRepo     *repository.PrefectureRepository
}

func NewGetDBEditsUsecase(
	dbEditRepo *repository.DBEditRepository,
	animeRevisionRepo *repository.AnimeRevisionRepository,
	personRevisionRepo *repository.PersonRevisionRepository,
	numberFormatRepo *repository.NumberFormatRepository,
	prefectureRepo *repository.PrefectureRepository,
) *GetDBEditsUsecase {
	return &GetDBEditsUsecase{
		dbEditRepo:         dbEditRepo,
		animeRevisionRepo:  animeRevisionRepo,
		personRevisionRepo: personRevisionRepo,
		numberFormatRepo:   numberFormatRepo,
		prefectureRepo:     prefectureRepo,
	}
}

type GetDBEditsInput struct {
	// Userは閲覧者 (ログインしていないときはnil)。管理者かどうかで見せる版を決める。
	User    *model.User
	Page    int32
	PerPage int32
}

type GetDBEditsOutput struct {
	// Editsは操作を新しい順に並べたもの。各操作はAnimeRevisionsとPersonRevisionsに見せる版を持つ。
	Edits      []*model.DBEdit
	TotalCount int64
	// NumberFormatsとPrefecturesは、版の「話数フォーマット」と「出身地」をIDではなく名前で示すための一覧。
	NumberFormats []model.NumberFormat
	Prefectures   []model.Prefecture
}

// Executeは操作の1ページ分を返す。版は操作ごとにまとめず、1ページ分を作品・エピソードと人物のそれぞれ1回の
// クエリで読んでから操作に振り分ける。
func (uc *GetDBEditsUsecase) Execute(ctx context.Context, input GetDBEditsInput) (*GetDBEditsOutput, error) {
	isAdmin := input.User != nil && input.User.IsAdmin()

	edits, err := uc.dbEditRepo.List(ctx, repository.DBEditListParams{
		IncludeDelisted: isAdmin,
		Page:            input.Page,
		PerPage:         input.PerPage,
	})
	if err != nil {
		return nil, fmt.Errorf("編集履歴の操作の取得に失敗: %w", err)
	}
	totalCount, err := uc.dbEditRepo.Count(ctx, isAdmin)
	if err != nil {
		return nil, fmt.Errorf("編集履歴の操作の数の取得に失敗: %w", err)
	}

	ids := make([]model.DBEditID, len(edits))
	editByID := make(map[model.DBEditID]*model.DBEdit, len(edits))
	for i, edit := range edits {
		ids[i] = edit.ID
		editByID[edit.ID] = edit
	}

	animeRevisions, err := uc.animeRevisionRepo.ListForDBEdits(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("作品・エピソードの版の取得に失敗: %w", err)
	}
	for _, r := range animeRevisions {
		edit := editByID[*r.DBEditID]
		r.DBEdit = edit
		edit.AnimeRevisions = append(edit.AnimeRevisions, r)
	}
	personRevisions, err := uc.personRevisionRepo.ListForDBEdits(ctx, ids, isAdmin)
	if err != nil {
		return nil, fmt.Errorf("人物の版の取得に失敗: %w", err)
	}
	for _, r := range personRevisions {
		edit := editByID[*r.DBEditID]
		r.DBEdit = edit
		edit.PersonRevisions = append(edit.PersonRevisions, r)
	}

	numberFormats, err := uc.numberFormatRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("話数フォーマットの取得に失敗: %w", err)
	}
	prefectures, err := uc.prefectureRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("出身地の取得に失敗: %w", err)
	}

	return &GetDBEditsOutput{Edits: edits, TotalCount: totalCount, NumberFormats: numberFormats, Prefectures: prefectures}, nil
}
