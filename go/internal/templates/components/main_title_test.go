package components

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestMainTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		data            MainTitleData
		wantContains    []string
		wantNotContains []string
	}{
		{
			name: "Titleのみ指定",
			data: MainTitleData{
				Title: "作品一覧",
			},
			wantContains: []string{
				`<h1 class="flex items-center gap-2 text-2xl font-bold antialiased">`,
				`作品一覧`,
				`</h1>`,
			},
			wantNotContains: []string{
				`text-sm text-gray-600`,
				`flex w-full flex-none items-center justify-end gap-2 md:w-auto`,
			},
		},
		{
			name: "TitleとSubtitleを指定",
			data: MainTitleData{
				Title:    "作品一覧",
				Subtitle: "アニメ作品の管理",
			},
			wantContains: []string{
				`作品一覧`,
				`<div class="text-sm text-gray-600">`,
				`アニメ作品の管理`,
			},
		},
		{
			name: "TitleとContentを指定",
			data: MainTitleData{
				Title:   "作品一覧",
				Content: rawComponent(`<p class="custom-content">説明文</p>`),
			},
			wantContains: []string{
				`作品一覧`,
				`<p class="custom-content">説明文</p>`,
			},
		},
		{
			name: "TitleとActionsを指定",
			data: MainTitleData{
				Title:   "作品一覧",
				Actions: rawComponent(`<button class="btn-primary">新規作成</button>`),
			},
			wantContains: []string{
				`作品一覧`,
				`<div class="flex w-full flex-none items-center justify-end gap-2 md:w-auto">`,
				`<button class="btn-primary">新規作成</button>`,
			},
		},
		{
			name: "全フィールド指定",
			data: MainTitleData{
				Title:    "作品一覧",
				Subtitle: "アニメ作品の管理",
				Content:  rawComponent(`<p>追加コンテンツ</p>`),
				Actions:  rawComponent(`<button>アクション</button>`),
			},
			wantContains: []string{
				`作品一覧`,
				`アニメ作品の管理`,
				`<p>追加コンテンツ</p>`,
				`<button>アクション</button>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			var buf strings.Builder
			if err := MainTitle(tt.data).Render(ctx, &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			html := buf.String()

			for _, want := range tt.wantContains {
				if !strings.Contains(html, want) {
					t.Errorf("期待する文字列が含まれていません: %q\nHTML: %s", want, html)
				}
			}

			for _, notWant := range tt.wantNotContains {
				if strings.Contains(html, notWant) {
					t.Errorf("含まれてはいけない文字列が含まれています: %q\nHTML: %s", notWant, html)
				}
			}
		})
	}
}

// rawComponent はテスト用に文字列をそのままレンダリングするtempl.Componentを返します
func rawComponent(html string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, html)
		return err
	})
}
