package email

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/model"
)

// TestEpisodeGenerationReportSender_Sendは、自動生成の報告を件名・HTML・テキストの両方に
// 書き、Annict DBのエピソード一覧へのリンクと飛ばした理由の説明を載せることをテストする。
func TestEpisodeGenerationReportSender_Send(t *testing.T) {
	t.Parallel()

	slotID := model.SlotID(493201)
	episodeID := model.EpisodeID(176556)
	report := &model.EpisodeGenerationReport{
		CreatedEpisodes: []model.EpisodeGenerationCreatedEpisode{
			{WorkID: 15816, WorkTitle: "幼馴染とはラブコメにならない", EpisodeNumber: "第13話"},
		},
		LinkedSlotCount: 3,
		SkippedSlots: []model.EpisodeGenerationSkippedSlot{
			{SlotID: 493202, WorkTitle: "幼馴染とはラブコメにならない", Reason: model.EpisodeGenerationSkipReasonCandidatesLinked},
		},
		FilledTitleCount: 2,
		Errors: []model.EpisodeGenerationError{
			{SlotID: &slotID, Message: "作品のロックに失敗しました"},
			{EpisodeID: &episodeID, Message: "エピソードのタイトルの更新に失敗しました"},
			{Message: "しょぼいカレンダーからのサブタイトルの取得に失敗"},
		},
	}

	noop := NewNoopSender()
	sender := NewEpisodeGenerationReportSender(noop, "https://example.com")

	ctx := context.Background()
	if err := sender.Send(ctx, "admin@example.com", report); err != nil {
		t.Fatalf("Send()のエラー = %v", err)
	}

	if len(noop.SentEmails) != 1 {
		t.Fatalf("SentEmailsの件数 = %d、期待値 = 1", len(noop.SentEmails))
	}
	sent := noop.SentEmails[0]
	if sent.To != "admin@example.com" {
		t.Errorf("To = %s、期待値 = admin@example.com", sent.To)
	}
	if sent.Subject != "Annict - エピソードの自動生成の結果" {
		t.Errorf("Subject = %q、期待値 = %q", sent.Subject, "Annict - エピソードの自動生成の結果")
	}

	if sent.HTMLBody == nil || sent.TextBody == nil {
		t.Fatal("HTMLBodyまたはTextBodyがnilです")
	}
	bodies := map[string]*bytes.Buffer{"HTMLBody": {}, "TextBody": {}}
	if err := sent.HTMLBody.Render(ctx, bodies["HTMLBody"]); err != nil {
		t.Fatalf("HTMLBody.Render()のエラー = %v", err)
	}
	if err := sent.TextBody.Render(ctx, bodies["TextBody"]); err != nil {
		t.Fatalf("TextBody.Render()のエラー = %v", err)
	}

	wants := []string{
		"作ったエピソード (1件)",
		"幼馴染とはラブコメにならない 第13話",
		"https://example.com/db/works/15816/episodes",
		"既存のエピソードに紐づけた放送枠: 3件",
		"放送枠493202 幼馴染とはラブコメにならない",
		"同じ番組の別の放送枠に紐づいています",
		"タイトルを入れたエピソード: 2件",
		"エラー (3件)",
		"放送枠493201: 作品のロックに失敗しました",
		"エピソード176556: エピソードのタイトルの更新に失敗しました",
		"しょぼいカレンダーからのサブタイトルの取得に失敗",
	}
	for name, body := range bodies {
		for _, want := range wants {
			if !strings.Contains(body.String(), want) {
				t.Errorf("%sに %q が含まれていません: %s", name, want, body.String())
			}
		}
	}
}
