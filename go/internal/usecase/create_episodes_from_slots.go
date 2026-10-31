package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// episodeGenerationWindowは、自動生成が対象にする放送枠の開始日時の範囲を、実行時刻の前後に
// 取る幅。後ろ側はRailsと同じく7日後まで。前側は、予定エピソード数との比較を直したことで
// 過去の放送枠が新しく対象に入り、リリース後の最初の実行でまとめて作らないように設ける。
// 7日あれば、実行が数日止まっても残りを拾える。
const episodeGenerationWindow = 7 * 24 * time.Hour

// SlotSkipReasonは、自動生成が放送枠を飛ばした理由を表す。どれも運用上の判断が要る状況で、
// 報告を見た編集者が紐づけや並び順を直す。
type SlotSkipReason string

const (
	// SlotSkipReasonCandidatesLinkedは、計算した話数のエピソードがあるものの、どれも同じ
	// 番組の別の放送枠に紐づいていることを表す。番組が途中の話数から放送していて、編集者が
	// 紐づけを手で付け替えた場合などに起きる。
	SlotSkipReasonCandidatesLinked SlotSkipReason = "candidates_linked"
	// SlotSkipReasonNoSortNumberGapは、新しいエピソードを入れる位置の前後のsort_numberの間が
	// 空いておらず、並び順を決められないことを表す。
	SlotSkipReasonNoSortNumberGap SlotSkipReason = "no_sort_number_gap"
)

// CreateEpisodesFromSlotsUsecaseは、エピソードが紐づいていない放送枠に、対応するエピソードを
// 紐づける。無ければエピソードを作る。
//
// 書き込みは一括作成と同じくanimesを基点とし、作るエピソードはanimeとそのkind='episode' の
// 分類になり、episodesの行は同一トランザクションで両書きする。移行期間中の正本はepisodes側。
type CreateEpisodesFromSlotsUsecase struct {
	db                      *sql.DB
	slotRepo                *repository.SlotRepository
	workRepo                *repository.WorkRepository
	episodeRepo             *repository.EpisodeRepository
	animeRepo               *repository.AnimeRepository
	animeClassificationRepo *repository.AnimeClassificationRepository
}

// NewCreateEpisodesFromSlotsUsecaseはCreateEpisodesFromSlotsUsecaseを生成する。
func NewCreateEpisodesFromSlotsUsecase(
	db *sql.DB,
	slotRepo *repository.SlotRepository,
	workRepo *repository.WorkRepository,
	episodeRepo *repository.EpisodeRepository,
	animeRepo *repository.AnimeRepository,
	animeClassificationRepo *repository.AnimeClassificationRepository,
) *CreateEpisodesFromSlotsUsecase {
	return &CreateEpisodesFromSlotsUsecase{
		db:                      db,
		slotRepo:                slotRepo,
		workRepo:                workRepo,
		episodeRepo:             episodeRepo,
		animeRepo:               animeRepo,
		animeClassificationRepo: animeClassificationRepo,
	}
}

// CreateEpisodesFromSlotsInputは実行時刻を持つ。対象にする放送枠の開始日時の範囲は、
// この時刻の前後に取る。
type CreateEpisodesFromSlotsInput struct {
	Now time.Time
}

// CreateEpisodesFromSlotsOutputは1回の実行の結果を、報告に使える形で返す。
type CreateEpisodesFromSlotsOutput struct {
	CreatedEpisodes []GeneratedEpisode
	// LinkedSlotCountは、既存のエピソードに紐づけた放送枠の件数。
	LinkedSlotCount int
	SkippedSlots    []SkippedSlot
	FailedSlots     []FailedSlot
}

// GeneratedEpisodeは自動生成で作ったエピソードを表す。
type GeneratedEpisode struct {
	SlotID    model.SlotID
	WorkID    model.WorkID
	WorkTitle string
	EpisodeID model.EpisodeID
	Number    string
}

// SkippedSlotは、紐づけ先を決められずに飛ばした放送枠を表す。
type SkippedSlot struct {
	SlotID    model.SlotID
	WorkID    model.WorkID
	WorkTitle string
	Reason    SlotSkipReason
}

// FailedSlotは、処理がエラーで失敗した放送枠を表す。
type FailedSlot struct {
	SlotID model.SlotID
	Err    error
}

// Executeは対象の放送枠を開始日時の順に、1件ずつ別のトランザクションで処理する。1件の
// 失敗で残りの放送枠を止めないよう、放送枠ごとのエラーは結果に積んで次へ進む。放送枠の一覧を
// 読めないときだけエラーを返す。
func (uc *CreateEpisodesFromSlotsUsecase) Execute(ctx context.Context, input CreateEpisodesFromSlotsInput) (*CreateEpisodesFromSlotsOutput, error) {
	slots, err := uc.slotRepo.ListForEpisodeGeneration(
		ctx,
		input.Now.Add(-episodeGenerationWindow),
		input.Now.Add(episodeGenerationWindow),
	)
	if err != nil {
		return nil, fmt.Errorf("自動生成の対象の放送枠の取得に失敗: %w", err)
	}

	output := &CreateEpisodesFromSlotsOutput{}
	for _, slot := range slots {
		result, err := uc.processSlot(ctx, slot, input.Now)
		if err != nil {
			output.FailedSlots = append(output.FailedSlots, FailedSlot{SlotID: slot.ID, Err: err})
			continue
		}

		switch {
		case result.created != nil:
			output.CreatedEpisodes = append(output.CreatedEpisodes, *result.created)
		case result.linked:
			output.LinkedSlotCount++
		case result.skipped != nil:
			output.SkippedSlots = append(output.SkippedSlots, *result.skipped)
		}
	}

	return output, nil
}

// slotResultは1件の放送枠の処理結果。どのフィールドも空のときは、報告するまでもなく何も
// しなかったこと (予定エピソード数に達している、一覧を読んだ後に紐づけられたなど) を表す。
type slotResult struct {
	created *GeneratedEpisode
	linked  bool
	skipped *SkippedSlot
}

// processSlotは1件の放送枠を1トランザクションで処理する。作品の行をロックしてから、照合と
// 並び順に使う値を読む。手動登録や別の実行と同じ話を二重に作らないため (一括作成と同じ)。
func (uc *CreateEpisodesFromSlotsUsecase) processSlot(ctx context.Context, candidate *model.Slot, now time.Time) (slotResult, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return slotResult{}, fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	workRepo := uc.workRepo.WithTx(tx)
	episodeRepo := uc.episodeRepo.WithTx(tx)
	slotRepo := uc.slotRepo.WithTx(tx)

	locked, err := workRepo.LockForEpisodeCreateByID(ctx, candidate.WorkID)
	if err != nil {
		return slotResult{}, fmt.Errorf("作品のロックに失敗しました: %w", err)
	}
	if !locked {
		return slotResult{}, nil
	}

	slot, err := slotRepo.GetForEpisodeGenerationByID(ctx, candidate.ID, candidate.WorkID, now.Add(-episodeGenerationWindow), now.Add(episodeGenerationWindow))
	if err != nil {
		return slotResult{}, fmt.Errorf("放送枠の再取得に失敗しました: %w", err)
	}
	if slot == nil {
		return slotResult{}, nil
	}

	work, err := workRepo.GetForEpisodeGenerationByID(ctx, slot.WorkID)
	if err != nil {
		return slotResult{}, fmt.Errorf("作品の取得に失敗しました: %w", err)
	}
	if work == nil || exceedsManualEpisodesCount(work.Work, slot) {
		return slotResult{}, nil
	}

	rawNumber := slot.EpisodeRawNumber(work.Work.StartEpisodeRawNumber)
	siblings, err := episodeRepo.ListForEpisodeGenerationByWorkID(ctx, work.Work.ID, slot.Program.ID)
	if err != nil {
		return slotResult{}, fmt.Errorf("作品のエピソードの取得に失敗しました: %w", err)
	}

	matchedID, candidatesFound := matchGeneratedEpisode(siblings, rawNumber)
	if matchedID != nil {
		linked, err := slotRepo.LinkEpisode(ctx, slot, *matchedID)
		if err != nil {
			return slotResult{}, fmt.Errorf("放送枠へのエピソードの紐づけに失敗しました: %w", err)
		}
		if !linked {
			return slotResult{}, nil
		}
		if err := tx.Commit(); err != nil {
			return slotResult{}, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
		}
		return slotResult{linked: true}, nil
	}
	if candidatesFound {
		return slotResult{skipped: skippedSlot(slot, work.Work, SlotSkipReasonCandidatesLinked)}, nil
	}

	position, ok := generatedEpisodePosition(siblings, work.MaxSortNumber, rawNumber)
	if !ok {
		return slotResult{skipped: skippedSlot(slot, work.Work, SlotSkipReasonNoSortNumberGap)}, nil
	}

	number := generatedEpisodeNumber(work.NumberFormat, rawNumber)
	episode := &model.Episode{
		WorkID:        work.Work.ID,
		Number:        &number,
		RawNumber:     &rawNumber,
		SortNumber:    position.SortNumber,
		ParentAnimeID: work.Work.AnimeID,
	}

	var animeID *model.AnimeID
	if episode.ParentAnimeID != nil {
		anime, err := uc.animeRepo.WithTx(tx).Create(ctx, animeCreateParamsFromEpisode(episode))
		if err != nil {
			return slotResult{}, fmt.Errorf("animeの作成に失敗しました: %w", err)
		}
		if _, err := uc.animeClassificationRepo.WithTx(tx).Create(ctx, classificationCreateParamsFromEpisode(episode, anime.ID)); err != nil {
			return slotResult{}, fmt.Errorf("anime_classificationの作成に失敗しました: %w", err)
		}
		animeID = &anime.ID
	}

	episodeID, err := episodeRepo.CreateGenerated(ctx, repository.CreateGeneratedEpisodeParams{
		WorkID:        episode.WorkID,
		Number:        number,
		RawNumber:     rawNumber,
		SortNumber:    position.SortNumber,
		PrevEpisodeID: position.PrevEpisodeID,
		AnimeID:       animeID,
	})
	if err != nil {
		return slotResult{}, fmt.Errorf("エピソードの作成に失敗しました: %w", err)
	}

	if position.NextEpisodeID != nil {
		if err := episodeRepo.UpdatePrevEpisodeID(ctx, work.Work.ID, *position.NextEpisodeID, episodeID); err != nil {
			return slotResult{}, fmt.Errorf("直後のエピソードの付け替えに失敗しました: %w", err)
		}
	}

	linked, err := slotRepo.LinkEpisode(ctx, slot, episodeID)
	if err != nil {
		return slotResult{}, fmt.Errorf("放送枠へのエピソードの紐づけに失敗しました: %w", err)
	}
	if !linked {
		// 照合時の属性や対象条件と一致しなければ、作ったエピソードはロールバックで消す。
		return slotResult{}, nil
	}

	if err := workRepo.IncrementEpisodesCount(ctx, work.Work.ID, 1); err != nil {
		return slotResult{}, fmt.Errorf("作品のエピソード件数の更新に失敗しました: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return slotResult{}, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return slotResult{created: &GeneratedEpisode{
		SlotID:    slot.ID,
		WorkID:    work.Work.ID,
		WorkTitle: work.Work.Title,
		EpisodeID: episodeID,
		Number:    number,
	}}, nil
}

// exceedsManualEpisodesCountは、放送枠が作品の予定エピソード数を超える話にあたるかを返す。
// 回数ではなく不定期放送を除いた作品内の順番で比べる。回数で比べると、不定期放送のある番組で
// 最後の話が作られないため。
func exceedsManualEpisodesCount(work *model.Work, slot *model.Slot) bool {
	return work.ManualEpisodesCount != nil && *work.ManualEpisodesCount < slot.EpisodeOrder()
}

// matchGeneratedEpisodeは、放送枠を紐づける既存のエピソードを返す。候補は公開中で話数が
// rawNumberと等しいエピソードで、同じ番組の別の放送枠に紐づいているものは除き、残りのうち
// 並びで最初のものを選ぶ。candidatesFoundは、除く前の候補が1件以上あったかを表し、
// 紐づけ先がnilでこれがtrueなら、すべて除かれたことを表す。
func matchGeneratedEpisode(siblings []repository.DBEpisodeGenerationSibling, rawNumber float64) (matchedID *model.EpisodeID, candidatesFound bool) {
	for _, sibling := range siblings {
		if !sibling.Published || sibling.RawNumber == nil || *sibling.RawNumber != rawNumber {
			continue
		}
		candidatesFound = true
		if !sibling.LinkedToProgram {
			return &sibling.ID, true
		}
	}
	return nil, candidatesFound
}

// generatedEpisodeNumberは自動生成するエピソードの表示の番号を返す。作品の書式に当てはめ
// られなければ「第N話」にする (RailsのEpisodeGeneratorServiceと同じ)。
func generatedEpisodeNumber(numberFormat *model.NumberFormat, rawNumber float64) string {
	if numberFormat != nil {
		if number := numberFormat.EpisodeNumber(rawNumber); number != "" {
			return number
		}
	}
	return fmt.Sprintf("第%d話", int(rawNumber))
}

func skippedSlot(slot *model.Slot, work *model.Work, reason SlotSkipReason) *SkippedSlot {
	return &SkippedSlot{
		SlotID:    slot.ID,
		WorkID:    work.ID,
		WorkTitle: work.Title,
		Reason:    reason,
	}
}
