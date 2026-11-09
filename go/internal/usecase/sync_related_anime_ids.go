package usecase

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/repository"
)

// SyncRelatedAnimeIDsUsecaseは、作品・エピソードのメタデータを持つ関連テーブルの
// anime_id (放送枠はepisode_anime_idも) を、work_id / episode_idが指す行のanime_idと
// 照合して直す。animes同期バッチの最後のパスで、works / episodesのパスがanime_idを
// 書き戻したあとに走らせる。
//
// 列ごとにテーブルをIDの順に1ページずつ進め、1ページを1つのUPDATE文で直す。明示的な
// トランザクションは張らず、差分の判定もUPDATE文の中で写し元の今の値と比べて行う。
// 照合は冪等なので、途中で失敗しても次の実行が残りを直す。
type SyncRelatedAnimeIDsUsecase struct {
	castRepo       *repository.CastRepository
	staffRepo      *repository.StaffRepository
	trailerRepo    *repository.TrailerRepository
	workImageRepo  *repository.WorkImageRepository
	vodTitleRepo   *repository.VodTitleRepository
	slotRepo       *repository.SlotRepository
	programRepo    *repository.ProgramRepository
	seriesWorkRepo *repository.SeriesWorkRepository
	batchSize      int
}

// NewSyncRelatedAnimeIDsUsecaseはSyncRelatedAnimeIDsUsecaseを生成する。batchSizeが
// 0以下の場合はDefaultSyncAnimesBatchSizeにフォールバックし、ページの走査が必ず前進する
// ようにする。
func NewSyncRelatedAnimeIDsUsecase(
	castRepo *repository.CastRepository,
	staffRepo *repository.StaffRepository,
	trailerRepo *repository.TrailerRepository,
	workImageRepo *repository.WorkImageRepository,
	vodTitleRepo *repository.VodTitleRepository,
	slotRepo *repository.SlotRepository,
	programRepo *repository.ProgramRepository,
	seriesWorkRepo *repository.SeriesWorkRepository,
	batchSize int,
) *SyncRelatedAnimeIDsUsecase {
	if batchSize <= 0 {
		batchSize = DefaultSyncAnimesBatchSize
	}
	return &SyncRelatedAnimeIDsUsecase{
		castRepo:       castRepo,
		staffRepo:      staffRepo,
		trailerRepo:    trailerRepo,
		workImageRepo:  workImageRepo,
		vodTitleRepo:   vodTitleRepo,
		slotRepo:       slotRepo,
		programRepo:    programRepo,
		seriesWorkRepo: seriesWorkRepo,
		batchSize:      batchSize,
	}
}

// SyncRelatedAnimeIDsResultは、列ごとに書き換えた行の件数を報告する。整合性の
// メトリクスで、既存の行を埋め終えたあとは、Rails版が写し損ねた行や、work_id /
// episode_idを書き換えた行の件数になる。
type SyncRelatedAnimeIDsResult struct {
	Casts               int
	Staffs              int
	Trailers            int
	WorkImages          int
	VodTitles           int
	SlotsAnimeID        int
	SlotsEpisodeAnimeID int
	Programs            int
	SeriesWorks         int
}

// relatedAnimeIDColumnは照合する1つの列。reconcileは1ページを照合し、updatedには
// 書き換えた件数を積算する。
type relatedAnimeIDColumn struct {
	name      string
	reconcile func(ctx context.Context, afterID int64, batchSize int) (repository.AnimeIDReconcilePage, error)
	updated   *int
}

// Executeは対象のすべての列を照合し、列ごとの件数を返す。
func (uc *SyncRelatedAnimeIDsUsecase) Execute(ctx context.Context) (*SyncRelatedAnimeIDsResult, error) {
	result := &SyncRelatedAnimeIDsResult{}
	columns := []relatedAnimeIDColumn{
		{name: "casts.anime_id", reconcile: uc.castRepo.ReconcileAnimeIDsAfter, updated: &result.Casts},
		{name: "staffs.anime_id", reconcile: uc.staffRepo.ReconcileAnimeIDsAfter, updated: &result.Staffs},
		{name: "trailers.anime_id", reconcile: uc.trailerRepo.ReconcileAnimeIDsAfter, updated: &result.Trailers},
		{name: "work_images.anime_id", reconcile: uc.workImageRepo.ReconcileAnimeIDsAfter, updated: &result.WorkImages},
		{name: "vod_titles.anime_id", reconcile: uc.vodTitleRepo.ReconcileAnimeIDsAfter, updated: &result.VodTitles},
		{name: "slots.anime_id", reconcile: uc.slotRepo.ReconcileAnimeIDsAfter, updated: &result.SlotsAnimeID},
		{name: "slots.episode_anime_id", reconcile: uc.slotRepo.ReconcileEpisodeAnimeIDsAfter, updated: &result.SlotsEpisodeAnimeID},
		{name: "programs.anime_id", reconcile: uc.programRepo.ReconcileAnimeIDsAfter, updated: &result.Programs},
		{name: "series_works.anime_id", reconcile: uc.seriesWorkRepo.ReconcileAnimeIDsAfter, updated: &result.SeriesWorks},
	}

	for _, column := range columns {
		if err := uc.reconcileColumn(ctx, column); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// reconcileColumnは1つの列をIDの順に最後のページまで照合する。カーソル (afterID) は
// ページの最後の行のIDで厳密に増加するため、ループは必ず終了し、空のページが終端を表す。
func (uc *SyncRelatedAnimeIDsUsecase) reconcileColumn(ctx context.Context, column relatedAnimeIDColumn) error {
	var afterID int64
	for {
		page, err := column.reconcile(ctx, afterID, uc.batchSize)
		if err != nil {
			return fmt.Errorf("%sの照合に失敗: %w", column.name, err)
		}
		if page.LastID == 0 {
			return nil
		}

		*column.updated += page.Updated
		afterID = page.LastID
	}
}
