package episode_generation_report

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/a-h/templ"
)

// JaTextは日本語版のエピソードの自動生成の報告メール (テキスト形式) です。templのテキストは
// 改行を詰めて1行につなげるため、一覧を1行ずつ並べられるよう本文をGoで組み立てる。
func JaText(data Data) templ.Component {
	return templ.ComponentFunc(func(_ context.Context, w io.Writer) error {
		var b strings.Builder

		fmt.Fprintf(&b, "作ったエピソード (%d件)\n", len(data.CreatedEpisodes))
		for _, episode := range data.CreatedEpisodes {
			fmt.Fprintf(&b, "- %s %s %s\n", episode.WorkTitle, episode.EpisodeNumber, episode.EpisodesURL)
		}

		fmt.Fprintf(&b, "\n既存のエピソードに紐づけた放送枠: %d件\n", data.LinkedSlotCount)

		fmt.Fprintf(&b, "\n飛ばした放送枠 (%d件)\n", len(data.SkippedSlots))
		for _, slot := range data.SkippedSlots {
			fmt.Fprintf(&b, "- 放送枠%s %s: %s\n", slot.SlotID, slot.WorkTitle, slot.Reason)
		}

		fmt.Fprintf(&b, "\nタイトルを入れたエピソード: %d件\n", data.FilledTitleCount)

		fmt.Fprintf(&b, "\nエラー (%d件)\n", len(data.Errors))
		for _, e := range data.Errors {
			if e.Target != "" {
				fmt.Fprintf(&b, "- %s: %s\n", e.Target, e.Message)
			} else {
				fmt.Fprintf(&b, "- %s\n", e.Message)
			}
		}

		_, err := io.WriteString(w, b.String())
		return err
	})
}
