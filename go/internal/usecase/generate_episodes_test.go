package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
)

type fakeSlotEpisodeCreator struct {
	output *CreateEpisodesFromSlotsOutput
	err    error
	input  *CreateEpisodesFromSlotsInput
}

func (f *fakeSlotEpisodeCreator) Execute(_ context.Context, input CreateEpisodesFromSlotsInput) (*CreateEpisodesFromSlotsOutput, error) {
	f.input = &input
	return f.output, f.err
}

type fakeEpisodeTitleFiller struct {
	output *FillEpisodeTitlesOutput
	err    error
	input  *FillEpisodeTitlesInput
}

func (f *fakeEpisodeTitleFiller) Execute(_ context.Context, input FillEpisodeTitlesInput) (*FillEpisodeTitlesOutput, error) {
	f.input = &input
	return f.output, f.err
}

type fakeReportSender struct {
	err     error
	to      string
	reports []*model.EpisodeGenerationReport
}

func (f *fakeReportSender) Send(_ context.Context, to string, report *model.EpisodeGenerationReport) error {
	f.to = to
	f.reports = append(f.reports, report)
	return f.err
}

func TestGenerateEpisodesUsecase_Execute_SendsReport(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 10, 31, 20, 0, 0, 0, time.UTC)
	creator := &fakeSlotEpisodeCreator{output: &CreateEpisodesFromSlotsOutput{
		CreatedEpisodes: []GeneratedEpisode{{SlotID: 1, WorkID: 10, WorkTitle: "作品A", EpisodeID: 100, Number: "第2話"}},
		LinkedSlotCount: 2,
		SkippedSlots:    []SkippedSlot{{SlotID: 3, WorkID: 11, WorkTitle: "作品B", Reason: model.EpisodeGenerationSkipReasonCandidatesLinked}},
		FailedSlots:     []FailedSlot{{SlotID: 4, Err: errors.New("作品のロックに失敗しました")}},
	}}
	filler := &fakeEpisodeTitleFiller{output: &FillEpisodeTitlesOutput{
		FilledCount:    5,
		FailedEpisodes: []FailedEpisode{{EpisodeID: 200, Err: errors.New("エピソードのタイトルの更新に失敗しました")}},
	}}
	sender := &fakeReportSender{}

	output, err := NewGenerateEpisodesUsecase(creator, filler, sender, "admin@example.com").Execute(context.Background(), GenerateEpisodesInput{Now: now})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	if creator.input == nil || !creator.input.Now.Equal(now) || filler.input == nil || !filler.input.Now.Equal(now) {
		t.Errorf("実行時刻 = (%v, %v)、期待値 = どちらも %v", creator.input, filler.input, now)
	}
	if !output.ReportSent || len(sender.reports) != 1 || sender.to != "admin@example.com" {
		t.Fatalf("送信 = (ReportSent %v, %d通, 宛先 %q)、期待値 = admin@example.comへ1通", output.ReportSent, len(sender.reports), sender.to)
	}

	report := sender.reports[0]
	if len(report.CreatedEpisodes) != 1 || report.CreatedEpisodes[0] != (model.EpisodeGenerationCreatedEpisode{WorkID: 10, WorkTitle: "作品A", EpisodeNumber: "第2話"}) {
		t.Errorf("CreatedEpisodes = %+v、期待値 = 作品Aの第2話", report.CreatedEpisodes)
	}
	if report.LinkedSlotCount != 2 || report.FilledTitleCount != 5 {
		t.Errorf("(LinkedSlotCount, FilledTitleCount) = (%d, %d)、期待値 = (2, 5)", report.LinkedSlotCount, report.FilledTitleCount)
	}
	if len(report.SkippedSlots) != 1 || report.SkippedSlots[0] != (model.EpisodeGenerationSkippedSlot{SlotID: 3, WorkTitle: "作品B", Reason: model.EpisodeGenerationSkipReasonCandidatesLinked}) {
		t.Errorf("SkippedSlots = %+v、期待値 = 放送枠3", report.SkippedSlots)
	}
	if len(report.Errors) != 2 {
		t.Fatalf("len(Errors) = %d、期待値 = 2", len(report.Errors))
	}
	if report.Errors[0].SlotID == nil || *report.Errors[0].SlotID != 4 || report.Errors[0].Message != "作品のロックに失敗しました" {
		t.Errorf("Errors[0] = %+v、期待値 = 放送枠4のエラー", report.Errors[0])
	}
	if report.Errors[1].EpisodeID == nil || *report.Errors[1].EpisodeID != 200 {
		t.Errorf("Errors[1] = %+v、期待値 = エピソード200のエラー", report.Errors[1])
	}
}

// TestGenerateEpisodesUsecase_Execute_ReportsStepErrorsは、放送枠の一覧の取得としょぼいカレンダー
// への問い合わせが失敗しても、エラーを返さずに報告へ残すことを検証する。
func TestGenerateEpisodesUsecase_Execute_ReportsStepErrors(t *testing.T) {
	t.Parallel()

	creator := &fakeSlotEpisodeCreator{err: errors.New("自動生成の対象の放送枠の取得に失敗")}
	filler := &fakeEpisodeTitleFiller{err: errors.New("しょぼいカレンダーからのサブタイトルの取得に失敗")}
	sender := &fakeReportSender{}

	output, err := NewGenerateEpisodesUsecase(creator, filler, sender, "admin@example.com").Execute(context.Background(), GenerateEpisodesInput{Now: time.Now()})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	if !output.ReportSent || len(sender.reports) != 1 {
		t.Fatalf("送信 = (ReportSent %v, %d通)、期待値 = 1通", output.ReportSent, len(sender.reports))
	}
	errs := sender.reports[0].Errors
	if len(errs) != 2 || errs[0].Message != "自動生成の対象の放送枠の取得に失敗" || errs[1].Message != "しょぼいカレンダーからのサブタイトルの取得に失敗" {
		t.Errorf("Errors = %+v、期待値 = 2つの手順のエラー", errs)
	}
	if errs[0].SlotID != nil || errs[0].EpisodeID != nil {
		t.Errorf("Errors[0] = %+v、期待値 = 放送枠もエピソードも持たないエラー", errs[0])
	}
}

// TestGenerateEpisodesUsecase_Execute_SkipsReportは、報告することが無い実行と、送り先が
// 設定されていない実行では、報告メールを送らないことを検証する。
func TestGenerateEpisodesUsecase_Execute_SkipsReport(t *testing.T) {
	t.Parallel()

	withResult := func() (*fakeSlotEpisodeCreator, *fakeEpisodeTitleFiller) {
		return &fakeSlotEpisodeCreator{output: &CreateEpisodesFromSlotsOutput{LinkedSlotCount: 1}},
			&fakeEpisodeTitleFiller{output: &FillEpisodeTitlesOutput{}}
	}

	t.Run("報告することが無い", func(t *testing.T) {
		t.Parallel()

		sender := &fakeReportSender{}
		uc := NewGenerateEpisodesUsecase(
			&fakeSlotEpisodeCreator{output: &CreateEpisodesFromSlotsOutput{}},
			&fakeEpisodeTitleFiller{output: &FillEpisodeTitlesOutput{}},
			sender,
			"admin@example.com",
		)
		output, err := uc.Execute(context.Background(), GenerateEpisodesInput{Now: time.Now()})
		if err != nil {
			t.Fatalf("Execute()のエラー = %v", err)
		}
		if output.ReportSent || len(sender.reports) != 0 {
			t.Errorf("送信 = (ReportSent %v, %d通)、期待値 = 送らない", output.ReportSent, len(sender.reports))
		}
	})

	t.Run("管理者のメールアドレスが未設定", func(t *testing.T) {
		t.Parallel()

		creator, filler := withResult()
		sender := &fakeReportSender{}
		output, err := NewGenerateEpisodesUsecase(creator, filler, sender, "").Execute(context.Background(), GenerateEpisodesInput{Now: time.Now()})
		if err != nil {
			t.Fatalf("Execute()のエラー = %v", err)
		}
		if output.ReportSent || len(sender.reports) != 0 {
			t.Errorf("送信 = (ReportSent %v, %d通)、期待値 = 送らない", output.ReportSent, len(sender.reports))
		}
	})

	t.Run("メール送信クライアントが無い", func(t *testing.T) {
		t.Parallel()

		creator, filler := withResult()
		output, err := NewGenerateEpisodesUsecase(creator, filler, nil, "admin@example.com").Execute(context.Background(), GenerateEpisodesInput{Now: time.Now()})
		if err != nil {
			t.Fatalf("Execute()のエラー = %v", err)
		}
		if output.ReportSent {
			t.Error("ReportSent = true、期待値 = false")
		}
	})
}

func TestGenerateEpisodesUsecase_Execute_ReturnsSendError(t *testing.T) {
	t.Parallel()

	sendErr := errors.New("Resendへの送信に失敗")
	uc := NewGenerateEpisodesUsecase(
		&fakeSlotEpisodeCreator{output: &CreateEpisodesFromSlotsOutput{LinkedSlotCount: 1}},
		&fakeEpisodeTitleFiller{output: &FillEpisodeTitlesOutput{}},
		&fakeReportSender{err: sendErr},
		"admin@example.com",
	)
	if _, err := uc.Execute(context.Background(), GenerateEpisodesInput{Now: time.Now()}); !errors.Is(err, sendErr) {
		t.Fatalf("Execute()のエラー = %v、期待値 = %v をラップしたエラー", err, sendErr)
	}
}
