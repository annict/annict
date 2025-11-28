package components

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/internal/session"
)

func TestFormErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		formErrors    *session.FormErrors
		wantContains  []string
		wantNotRender bool
	}{
		{
			name: "グローバルエラーが1つ",
			formErrors: &session.FormErrors{
				Global: []string{"ログインに失敗しました"},
			},
			wantContains: []string{
				`<div class="alert-destructive">`,
				`<h2>ログインに失敗しました</h2>`,
			},
		},
		{
			name: "グローバルエラーが複数",
			formErrors: &session.FormErrors{
				Global: []string{
					"エラー1",
					"エラー2",
					"エラー3",
				},
			},
			wantContains: []string{
				`<h2>エラー1</h2>`,
				`<h2>エラー2</h2>`,
				`<h2>エラー3</h2>`,
			},
		},
		{
			name: "フィールドエラーのみ（グローバルエラーなし）",
			formErrors: &session.FormErrors{
				Fields: map[string][]string{
					"email": {"メールアドレスが不正です"},
				},
			},
			wantNotRender: true,
		},
		{
			name:          "formErrorsがnilの場合は何も表示しない",
			formErrors:    nil,
			wantNotRender: true,
		},
		{
			name: "formErrorsが空の場合は何も表示しない",
			formErrors: &session.FormErrors{
				Global: []string{},
			},
			wantNotRender: true,
		},
	}

	for _, tt := range tests {
		tt := tt // キャプチャ
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			// テンプレートをレンダリング
			var buf strings.Builder
			err := FormErrors(tt.formErrors).Render(ctx, &buf)
			if err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			html := buf.String()

			// 何も表示されないケース
			if tt.wantNotRender {
				if strings.TrimSpace(html) != "" {
					t.Errorf("何も表示されないはずだが、HTMLが生成されました: %s", html)
				}
				return
			}

			// 期待する文字列が含まれているか確認
			for _, want := range tt.wantContains {
				if !strings.Contains(html, want) {
					t.Errorf("期待する文字列が含まれていません: %q\nHTML: %s", want, html)
				}
			}
		})
	}
}

func TestFormErrors_HTMLStructure(t *testing.T) {
	t.Parallel()

	formErrors := &session.FormErrors{
		Global: []string{"エラーメッセージ"},
	}

	ctx := context.Background()
	var buf strings.Builder
	err := FormErrors(formErrors).Render(ctx, &buf)
	if err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	// 正しいHTML構造を持っているか確認
	expectedStructure := []string{
		`<div class="alert-destructive">`,
		`<h2>`,
		`</h2>`,
		`</div>`,
	}

	for _, expected := range expectedStructure {
		if !strings.Contains(html, expected) {
			t.Errorf("期待するHTML構造が含まれていません: %q\nHTML: %s", expected, html)
		}
	}
}
