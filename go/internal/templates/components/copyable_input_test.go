package components

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
)

// TestCopyableInputは、読み取り専用の欄と、JSが動くまで隠したコピーボタン、読み上げのためのステータス領域を
// 描画し、ボタンの名前と知らせる文言を両ロケールで持つことを検証する。
func TestCopyableInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		locale string
		wants  []string
	}{
		{
			name:   "日本語",
			locale: "ja",
			wants: []string{
				`aria-label="キーをコピー"`,
				`data-success-message="キーをコピーしました"`,
				`data-failure-message="コピーできませんでした。入力欄の値を選んでコピーしてください"`,
			},
		},
		{
			name:   "英語",
			locale: "en",
			wants: []string{
				`aria-label="Copy キー"`,
				`data-success-message="Copied キー to the clipboard"`,
				`data-failure-message="Couldn&#39;t copy. Select the value in the field and copy it instead"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), tt.locale)

			var buf strings.Builder
			err := CopyableInput(CopyableInputData{
				ID:    "secret",
				Label: "キー",
				Value: "ABCD EFGH",
			}).Render(ctx, &buf)
			if err != nil {
				t.Fatalf("Render()のエラー = %v", err)
			}

			html := buf.String()
			wants := append([]string{
				// web/copyable-input.tsが振る舞いを付ける目印と、読み取り専用の欄
				`<div role="group" class="field" data-copyable-input>`,
				`<label class="label" for="secret">キー</label>`,
				`data-copyable-input-value id="secret" readonly spellcheck="false" type="text" value="ABCD EFGH"`,
				`<button type="button" class="btn shrink-0" hidden`,
				`data-copyable-input-button`,
				// 読み上げは描画時からあるステータス領域で伝える
				`<p class="sr-only" role="status" data-copyable-input-status></p>`,
				// コピーボタンのアイコンは装飾で、名前はaria-labelが持つ
				`<svg aria-hidden="true" focusable="false"`,
			}, tt.wants...)
			for _, want := range wants {
				if !strings.Contains(html, want) {
					t.Errorf("出力に %qが含まれていない\n%s", want, html)
				}
			}
			// 説明を渡さないときは、存在しない要素を参照しない
			if strings.Contains(html, "aria-describedby") {
				t.Errorf("説明が無いのにaria-describedbyを出している\n%s", html)
			}
		})
	}
}

// TestCopyableInput_Descriptionは、説明を渡したときに欄の下へ表示し、欄から参照することを検証する。
func TestCopyableInput_Description(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	var buf strings.Builder
	err := CopyableInput(CopyableInputData{
		ID:          "secret",
		Label:       "キー",
		Value:       "ABCD EFGH",
		Description: "認証アプリに入力してください。",
	}).Render(ctx, &buf)
	if err != nil {
		t.Fatalf("Render()のエラー = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`aria-describedby="secret-description"`,
		`<p id="secret-description" class="text-sm text-muted-foreground">認証アプリに入力してください。</p>`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("出力に %qが含まれていない\n%s", want, html)
		}
	}
}
