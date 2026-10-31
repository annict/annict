package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/annict/annict/go/internal/model"
)

// slotEpisodeCreatorは放送枠にエピソードを紐づけ、無ければ作る。
// 実体は *CreateEpisodesFromSlotsUsecase。
type slotEpisodeCreator interface {
	Execute(ctx context.Context, input CreateEpisodesFromSlotsInput) (*CreateEpisodesFromSlotsOutput, error)
}

// episodeTitleFillerはタイトルが空のエピソードにしょぼいカレンダーのサブタイトルを入れる。
// 実体は *FillEpisodeTitlesUsecase。
type episodeTitleFiller interface {
	Execute(ctx context.Context, input FillEpisodeTitlesInput) (*FillEpisodeTitlesOutput, error)
}

// EpisodeGenerationReportSenderは、エピソードの自動生成の結果を報告するメールを送る。
type EpisodeGenerationReportSender interface {
	Send(ctx context.Context, to string, report *model.EpisodeGenerationReport) error
}

// GenerateEpisodesUsecaseはエピソードの自動生成をまとめる。放送枠へのエピソードの紐づけと作成、
// しょぼいカレンダーのサブタイトルの取得を順に行い、1回の実行の結果を1通のメールで管理者へ
// 報告する。
type GenerateEpisodesUsecase struct {
	creator      slotEpisodeCreator
	titleFiller  episodeTitleFiller
	reportSender EpisodeGenerationReportSender
	adminEmail   string
}

// NewGenerateEpisodesUsecaseはGenerateEpisodesUsecaseを生成する。reportSenderがnilか
// adminEmailが空のときは、報告メールを送らずにログに残す。
func NewGenerateEpisodesUsecase(
	creator slotEpisodeCreator,
	titleFiller episodeTitleFiller,
	reportSender EpisodeGenerationReportSender,
	adminEmail string,
) *GenerateEpisodesUsecase {
	return &GenerateEpisodesUsecase{
		creator:      creator,
		titleFiller:  titleFiller,
		reportSender: reportSender,
		adminEmail:   adminEmail,
	}
}

// GenerateEpisodesInputは実行時刻を持つ。
type GenerateEpisodesInput struct {
	Now time.Time
}

// GenerateEpisodesOutputは1回の実行の報告と、それをメールで送ったかを返す。
type GenerateEpisodesOutput struct {
	Report     *model.EpisodeGenerationReport
	ReportSent bool
}

// Executeは自動生成を実行し、結果を報告する。放送枠の処理やしょぼいカレンダーの失敗は、
// 残りの処理を止めずに報告とログ (Sentryにも送られる) に残す。放送枠の処理は冪等で、翌日の
// 実行で残りを拾えるため、これらの失敗ではエラーを返さない。エラーを返すのは報告メールの
// 送信に失敗したときだけ。
func (uc *GenerateEpisodesUsecase) Execute(ctx context.Context, input GenerateEpisodesInput) (*GenerateEpisodesOutput, error) {
	report := &model.EpisodeGenerationReport{}

	created, err := uc.creator.Execute(ctx, CreateEpisodesFromSlotsInput(input))
	if err != nil {
		slog.ErrorContext(ctx, "放送枠からのエピソードの自動生成に失敗しました", "error", err)
		report.Errors = append(report.Errors, model.EpisodeGenerationError{Message: err.Error()})
	} else {
		addCreatedEpisodesToReport(ctx, report, created)
	}

	filled, err := uc.titleFiller.Execute(ctx, FillEpisodeTitlesInput(input))
	if err != nil {
		slog.ErrorContext(ctx, "しょぼいカレンダーからのエピソードのタイトルの取得に失敗しました", "error", err)
		report.Errors = append(report.Errors, model.EpisodeGenerationError{Message: err.Error()})
	} else {
		addFilledTitlesToReport(ctx, report, filled)
	}

	slog.InfoContext(ctx, "エピソードの自動生成が完了しました",
		"created_episodes", len(report.CreatedEpisodes),
		"linked_slots", report.LinkedSlotCount,
		"skipped_slots", len(report.SkippedSlots),
		"filled_titles", report.FilledTitleCount,
		"errors", len(report.Errors),
	)

	sent, err := uc.sendReport(ctx, report)
	if err != nil {
		return nil, err
	}
	return &GenerateEpisodesOutput{Report: report, ReportSent: sent}, nil
}

// addCreatedEpisodesToReportは放送枠の処理の結果を報告に積む。放送枠ごとのエラーはSentryに
// 送られるErrorで、飛ばした放送枠は運用上の判断が要る状況でありエラーではないためWarnで記録する。
func addCreatedEpisodesToReport(ctx context.Context, report *model.EpisodeGenerationReport, output *CreateEpisodesFromSlotsOutput) {
	for _, episode := range output.CreatedEpisodes {
		report.CreatedEpisodes = append(report.CreatedEpisodes, model.EpisodeGenerationCreatedEpisode{
			WorkID:        episode.WorkID,
			WorkTitle:     episode.WorkTitle,
			EpisodeNumber: episode.Number,
		})
	}
	report.LinkedSlotCount = output.LinkedSlotCount

	for _, skipped := range output.SkippedSlots {
		slog.WarnContext(ctx, "紐づけ先を決められない放送枠を飛ばしました",
			"slot_id", skipped.SlotID,
			"work_id", skipped.WorkID,
			"reason", skipped.Reason,
		)
		report.SkippedSlots = append(report.SkippedSlots, model.EpisodeGenerationSkippedSlot{
			SlotID:    skipped.SlotID,
			WorkTitle: skipped.WorkTitle,
			Reason:    skipped.Reason,
		})
	}

	for _, failed := range output.FailedSlots {
		slog.ErrorContext(ctx, "放送枠からのエピソードの自動生成に失敗しました", "slot_id", failed.SlotID, "error", failed.Err)
		slotID := failed.SlotID
		report.Errors = append(report.Errors, model.EpisodeGenerationError{SlotID: &slotID, Message: failed.Err.Error()})
	}
}

// addFilledTitlesToReportはタイトルの取得の結果を報告に積む。
func addFilledTitlesToReport(ctx context.Context, report *model.EpisodeGenerationReport, output *FillEpisodeTitlesOutput) {
	report.FilledTitleCount = output.FilledCount

	for _, failed := range output.FailedEpisodes {
		slog.ErrorContext(ctx, "エピソードへのタイトルの書き込みに失敗しました", "episode_id", failed.EpisodeID, "error", failed.Err)
		episodeID := failed.EpisodeID
		report.Errors = append(report.Errors, model.EpisodeGenerationError{EpisodeID: &episodeID, Message: failed.Err.Error()})
	}
}

// sendReportは報告メールを送り、送ったかを返す。報告することが何も無い実行では送らない。
func (uc *GenerateEpisodesUsecase) sendReport(ctx context.Context, report *model.EpisodeGenerationReport) (bool, error) {
	if report.IsEmpty() {
		return false, nil
	}
	if uc.reportSender == nil || uc.adminEmail == "" {
		slog.InfoContext(ctx, "報告メールの送り先が設定されていないため、自動生成の報告メールを送りません")
		return false, nil
	}

	if err := uc.reportSender.Send(ctx, uc.adminEmail, report); err != nil {
		return false, fmt.Errorf("自動生成の報告メールの送信に失敗: %w", err)
	}
	return true, nil
}
