package usecase

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
)

// recordCallは、子のユースケースのフェイクが呼ばれた順をcallsに記録する。callsがnilなら記録しない。
func recordCall(calls *[]string, name string) {
	if calls != nil {
		*calls = append(*calls, name)
	}
}

type fakeSlotEpisodeCreator struct {
	output *CreateEpisodesFromSlotsOutput
	err    error
	input  *CreateEpisodesFromSlotsInput
	calls  *[]string
	// beforeReturnは、結果を返す前に呼ぶ。実行上限を超えた状況を作るため、ctxを終わらせるのに使う。
	beforeReturn func()
}

func (f *fakeSlotEpisodeCreator) Execute(_ context.Context, input CreateEpisodesFromSlotsInput) (*CreateEpisodesFromSlotsOutput, error) {
	f.input = &input
	recordCall(f.calls, "CreateEpisodesFromSlots")
	if f.beforeReturn != nil {
		f.beforeReturn()
	}
	return f.output, f.err
}

type fakeEpisodeTitleFiller struct {
	output *FillEpisodeTitlesOutput
	err    error
	input  *FillEpisodeTitlesInput
	calls  *[]string
	// beforeReturnは、結果を返す前に呼ぶ。実行上限を超えた状況を作るため、ctxを終わらせるのに使う。
	beforeReturn func()
}

func (f *fakeEpisodeTitleFiller) Execute(_ context.Context, input FillEpisodeTitlesInput) (*FillEpisodeTitlesOutput, error) {
	f.input = &input
	recordCall(f.calls, "FillEpisodeTitles")
	if f.beforeReturn != nil {
		f.beforeReturn()
	}
	return f.output, f.err
}

type fakeLibraryEntryUpdater struct {
	output *UpdateLibraryEntryNextResourcesOutput
	err    error
	input  *UpdateLibraryEntryNextResourcesInput
	calls  *[]string
	// beforeReturnは、結果を返す前に呼ぶ。実行上限を超えた状況を作るため、ctxを終わらせるのに使う。
	beforeReturn func()
}

func (f *fakeLibraryEntryUpdater) Execute(_ context.Context, input UpdateLibraryEntryNextResourcesInput) (*UpdateLibraryEntryNextResourcesOutput, error) {
	f.input = &input
	recordCall(f.calls, "UpdateLibraryEntryNextResources")
	if f.beforeReturn != nil {
		f.beforeReturn()
	}
	return f.output, f.err
}

// newFakeLibraryEntryUpdaterは、1件も更新しないライブラリの更新を返す。
func newFakeLibraryEntryUpdater() *fakeLibraryEntryUpdater {
	return &fakeLibraryEntryUpdater{output: &UpdateLibraryEntryNextResourcesOutput{}}
}

type fakeReportSender struct {
	err     error
	to      string
	reports []*model.EpisodeGenerationReport
	// ctxErrは、送るときに渡されたctxのErrの値。
	ctxErr error
}

func (f *fakeReportSender) Send(ctx context.Context, to string, report *model.EpisodeGenerationReport) error {
	f.ctxErr = ctx.Err()
	f.to = to
	f.reports = append(f.reports, report)
	return f.err
}

func TestGenerateEpisodesUsecase_Execute_SendsReport(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 10, 31, 20, 0, 0, 0, time.UTC)
	var calls []string
	creator := &fakeSlotEpisodeCreator{output: &CreateEpisodesFromSlotsOutput{
		CreatedEpisodes: []GeneratedEpisode{{SlotID: 1, WorkID: 10, WorkTitle: "作品A", EpisodeID: 100, Number: "第2話"}},
		LinkedSlotCount: 2,
		SkippedSlots:    []SkippedSlot{{SlotID: 3, WorkID: 11, WorkTitle: "作品B", Reason: model.EpisodeGenerationSkipReasonCandidatesLinked}},
		FailedSlots:     []FailedSlot{{SlotID: 4, Err: errors.New("作品のロックに失敗しました")}},
	}, calls: &calls}
	filler := &fakeEpisodeTitleFiller{output: &FillEpisodeTitlesOutput{
		FilledCount:    5,
		FailedEpisodes: []FailedEpisode{{EpisodeID: 200, Err: errors.New("エピソードのタイトルの更新に失敗しました")}},
	}, calls: &calls}
	updater := &fakeLibraryEntryUpdater{output: &UpdateLibraryEntryNextResourcesOutput{UpdatedCount: 7}, calls: &calls}
	sender := &fakeReportSender{}

	output, err := NewGenerateEpisodesUsecase(creator, filler, updater, sender, "admin@example.com").Execute(context.Background(), GenerateEpisodesInput{Now: now})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	// その日に作ったエピソードと紐づけた放送枠をライブラリに反映するため、ライブラリの更新は最後に呼ぶ。
	if want := []string{"CreateEpisodesFromSlots", "FillEpisodeTitles", "UpdateLibraryEntryNextResources"}; !slices.Equal(calls, want) {
		t.Errorf("子のユースケースを呼んだ順 = %v、期待値 = %v", calls, want)
	}
	if creator.input == nil || !creator.input.Now.Equal(now) || filler.input == nil || !filler.input.Now.Equal(now) || updater.input == nil || !updater.input.Now.Equal(now) {
		t.Errorf("実行時刻 = (%v, %v, %v)、期待値 = どれも %v", creator.input, filler.input, updater.input, now)
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

// TestGenerateEpisodesUsecase_Execute_ReportsStepErrorsは、放送枠の一覧の取得・しょぼいカレンダー
// への問い合わせ・ライブラリの更新が失敗しても、後の手順を続け、エラーを返さずに報告へ残すことを
// 検証する。
func TestGenerateEpisodesUsecase_Execute_ReportsStepErrors(t *testing.T) {
	t.Parallel()

	creator := &fakeSlotEpisodeCreator{err: errors.New("自動生成の対象の放送枠の取得に失敗")}
	filler := &fakeEpisodeTitleFiller{err: errors.New("しょぼいカレンダーからのサブタイトルの取得に失敗")}
	updater := &fakeLibraryEntryUpdater{err: errors.New("ライブラリの次のエピソード・次の放送枠の更新に失敗")}
	sender := &fakeReportSender{}

	output, err := NewGenerateEpisodesUsecase(creator, filler, updater, sender, "admin@example.com").Execute(context.Background(), GenerateEpisodesInput{Now: time.Now()})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	if !output.ReportSent || len(sender.reports) != 1 {
		t.Fatalf("送信 = (ReportSent %v, %d通)、期待値 = 1通", output.ReportSent, len(sender.reports))
	}
	if updater.input == nil {
		t.Error("ライブラリの更新が呼ばれなかった")
	}
	errs := sender.reports[0].Errors
	if len(errs) != 3 ||
		errs[0].Message != "自動生成の対象の放送枠の取得に失敗" ||
		errs[1].Message != "しょぼいカレンダーからのサブタイトルの取得に失敗" ||
		errs[2].Message != "ライブラリの次のエピソード・次の放送枠の更新に失敗" {
		t.Errorf("Errors = %+v、期待値 = 3つの手順のエラー", errs)
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

	// ライブラリを更新した件数は報告に含めないため、更新があっても送らない。
	t.Run("報告することが無い", func(t *testing.T) {
		t.Parallel()

		sender := &fakeReportSender{}
		uc := NewGenerateEpisodesUsecase(
			&fakeSlotEpisodeCreator{output: &CreateEpisodesFromSlotsOutput{}},
			&fakeEpisodeTitleFiller{output: &FillEpisodeTitlesOutput{}},
			&fakeLibraryEntryUpdater{output: &UpdateLibraryEntryNextResourcesOutput{UpdatedCount: 3}},
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
		output, err := NewGenerateEpisodesUsecase(creator, filler, newFakeLibraryEntryUpdater(), sender, "").Execute(context.Background(), GenerateEpisodesInput{Now: time.Now()})
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
		output, err := NewGenerateEpisodesUsecase(creator, filler, newFakeLibraryEntryUpdater(), nil, "admin@example.com").Execute(context.Background(), GenerateEpisodesInput{Now: time.Now()})
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
		newFakeLibraryEntryUpdater(),
		&fakeReportSender{err: sendErr},
		"admin@example.com",
	)
	if _, err := uc.Execute(context.Background(), GenerateEpisodesInput{Now: time.Now()}); !errors.Is(err, sendErr) {
		t.Fatalf("Execute()のエラー = %v、期待値 = %v をラップしたエラー", err, sendErr)
	}
}

// TestGenerateEpisodesUsecase_Execute_ReportsTimeoutは、実行上限を超えてctxが終わったら残りの
// 手順を呼ばず、処理しなかったものを1件のエラーにまとめ、ジョブのctxから切り離して報告メールを
// 送ることを検証する。
func TestGenerateEpisodesUsecase_Execute_ReportsTimeout(t *testing.T) {
	t.Parallel()

	const prefix = "実行上限の時間を超えたため、次の処理をしませんでした: "
	tests := []struct {
		name        string
		setup       func(cancel context.CancelFunc, creator *fakeSlotEpisodeCreator, filler *fakeEpisodeTitleFiller, updater *fakeLibraryEntryUpdater)
		wantCalls   []string
		wantMessage string
	}{
		{
			name: "放送枠の途中で時間切れ",
			setup: func(cancel context.CancelFunc, creator *fakeSlotEpisodeCreator, _ *fakeEpisodeTitleFiller, _ *fakeLibraryEntryUpdater) {
				creator.output = &CreateEpisodesFromSlotsOutput{UnprocessedSlotCount: 2}
				creator.beforeReturn = cancel
			},
			wantCalls:   []string{"CreateEpisodesFromSlots"},
			wantMessage: prefix + "放送枠2件、タイトルの取得、ライブラリの更新",
		},
		{
			name: "放送枠の一覧を読む間に時間切れ",
			setup: func(cancel context.CancelFunc, creator *fakeSlotEpisodeCreator, _ *fakeEpisodeTitleFiller, _ *fakeLibraryEntryUpdater) {
				creator.err = context.DeadlineExceeded
				creator.beforeReturn = cancel
			},
			wantCalls:   []string{"CreateEpisodesFromSlots"},
			wantMessage: prefix + "放送枠の処理、タイトルの取得、ライブラリの更新",
		},
		{
			name: "タイトルとライブラリの更新の途中で時間切れ",
			setup: func(cancel context.CancelFunc, _ *fakeSlotEpisodeCreator, filler *fakeEpisodeTitleFiller, updater *fakeLibraryEntryUpdater) {
				filler.output = &FillEpisodeTitlesOutput{UnprocessedEpisodeCount: 3}
				updater.output = &UpdateLibraryEntryNextResourcesOutput{UpdatedCount: 1, Interrupted: true}
				updater.beforeReturn = cancel
			},
			wantCalls:   []string{"CreateEpisodesFromSlots", "FillEpisodeTitles", "UpdateLibraryEntryNextResources"},
			wantMessage: prefix + "タイトルを入れるエピソード3件、ライブラリの更新の残り",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var calls []string
			creator := &fakeSlotEpisodeCreator{output: &CreateEpisodesFromSlotsOutput{}, calls: &calls}
			filler := &fakeEpisodeTitleFiller{output: &FillEpisodeTitlesOutput{}, calls: &calls}
			updater := &fakeLibraryEntryUpdater{output: &UpdateLibraryEntryNextResourcesOutput{}, calls: &calls}
			tt.setup(cancel, creator, filler, updater)
			sender := &fakeReportSender{}

			output, err := NewGenerateEpisodesUsecase(creator, filler, updater, sender, "admin@example.com").Execute(ctx, GenerateEpisodesInput{Now: time.Now()})
			if err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}

			if !slices.Equal(calls, tt.wantCalls) {
				t.Errorf("呼んだ手順 = %v、期待値 = %v", calls, tt.wantCalls)
			}
			if !output.ReportSent || len(sender.reports) != 1 {
				t.Fatalf("送信 = (ReportSent %v, %d通)、期待値 = 1通", output.ReportSent, len(sender.reports))
			}
			if sender.ctxErr != nil {
				t.Errorf("送るときのctxのErr = %v、期待値 = nil", sender.ctxErr)
			}
			errs := sender.reports[0].Errors
			if len(errs) != 1 || errs[0].Message != tt.wantMessage {
				t.Errorf("Errors = %+v、期待値 = %qの1件", errs, tt.wantMessage)
			}
		})
	}
}
