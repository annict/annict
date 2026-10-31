package email

import (
	"context"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates/emails/episode_generation_report"
)

// EpisodeGenerationReportSenderはエピソードの自動生成の報告メールの送信を行う
type EpisodeGenerationReportSender struct {
	sender Sender
	appURL string
}

// NewEpisodeGenerationReportSenderは新しいEpisodeGenerationReportSenderを作成する。
// appURLは、報告に載せるAnnict DBのページへのリンクの起点になる。
func NewEpisodeGenerationReportSender(sender Sender, appURL string) *EpisodeGenerationReportSender {
	return &EpisodeGenerationReportSender{sender: sender, appURL: appURL}
}

// Sendは報告メールをレンダリングして送信する。読み手は運用する管理者に限られるため、本文は
// 日本語版だけを用意する。
func (s *EpisodeGenerationReportSender) Send(ctx context.Context, to string, report *model.EpisodeGenerationReport) error {
	ctx = i18n.SetLocale(ctx, "ja")
	data := s.reportData(report)

	return s.sender.Send(ctx, SendInput{
		To:       to,
		Subject:  i18n.T(ctx, "episode_generation_report_email_subject"),
		HTMLBody: episode_generation_report.JaHTML(data),
		TextBody: episode_generation_report.JaText(data),
	})
}

func (s *EpisodeGenerationReportSender) reportData(report *model.EpisodeGenerationReport) episode_generation_report.Data {
	data := episode_generation_report.Data{
		LinkedSlotCount:  report.LinkedSlotCount,
		FilledTitleCount: report.FilledTitleCount,
	}

	for _, episode := range report.CreatedEpisodes {
		data.CreatedEpisodes = append(data.CreatedEpisodes, episode_generation_report.CreatedEpisode{
			WorkTitle:     episode.WorkTitle,
			EpisodeNumber: episode.EpisodeNumber,
			EpisodesURL:   fmt.Sprintf("%s/db/works/%s/episodes", s.appURL, episode.WorkID),
		})
	}

	for _, slot := range report.SkippedSlots {
		data.SkippedSlots = append(data.SkippedSlots, episode_generation_report.SkippedSlot{
			SlotID:    slot.SlotID.String(),
			WorkTitle: slot.WorkTitle,
			Reason:    skipReasonText(slot.Reason),
		})
	}

	for _, e := range report.Errors {
		target := ""
		switch {
		case e.SlotID != nil:
			target = "放送枠" + e.SlotID.String()
		case e.EpisodeID != nil:
			target = "エピソード" + e.EpisodeID.String()
		}
		data.Errors = append(data.Errors, episode_generation_report.Error{Target: target, Message: e.Message})
	}

	return data
}

// skipReasonTextは飛ばした理由を、報告を読んだ編集者が次に何を直せばよいか分かる文にする。
func skipReasonText(reason model.EpisodeGenerationSkipReason) string {
	switch reason {
	case model.EpisodeGenerationSkipReasonCandidatesLinked:
		return "計算した話数のエピソードが、同じ番組の別の放送枠に紐づいています。放送枠の紐づけを確認してください"
	case model.EpisodeGenerationSkipReasonNoSortNumberGap:
		return "前後のエピソードの並び順 (sort_number) の間が空いておらず、新しいエピソードを入れる位置を決められません。並び順を空けてください"
	default:
		return string(reason)
	}
}
