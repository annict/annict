package components

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/internal/viewmodel"
)

func TestHead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		meta         viewmodel.PageMeta
		assetVersion string
		wantContains []string
	}{
		{
			name: "基本的なメタタグが正しく出力される",
			meta: viewmodel.PageMeta{
				Title:       "テストページ | Annict",
				Description: "テストページの説明",
				OGType:      "website",
				OGURL:       "https://annict.com/test",
				OGImage:     "https://annict.com/test.png",
			},
			assetVersion: "v1.0.0",
			wantContains: []string{
				`<meta charset="UTF-8">`,
				`<meta name="description" content="テストページの説明">`,
				`<meta property="og:title" content="テストページ | Annict">`,
				`<meta property="og:type" content="website">`,
				`<meta property="og:url" content="https://annict.com/test">`,
				`<meta property="og:description" content="テストページの説明">`,
				`<meta property="og:site_name" content="Annict (アニクト)">`,
				`<meta property="og:image" content="https://annict.com/test.png">`,
				`<meta property="og:locale" content="ja_JP">`,
				`<meta name="twitter:card" content="summary">`,
				`<meta name="twitter:site" content="@AnnictJP">`,
				`<meta name="twitter:title" content="テストページ | Annict">`,
				`<meta name="twitter:description" content="テストページの説明">`,
				`<meta name="twitter:image" content="https://annict.com/test.png">`,
				`<meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover">`,
				`<title>テストページ | Annict</title>`,
				`<link rel="shortcut icon" href="/static/images/favicon.png" type="image/png">`,
				`<link rel="canonical" href="https://annict.com/test">`,
				`<link rel="manifest" href="/manifest.json">`,
				`<link rel="stylesheet" href="/static/css/style.css?v=v1.0.0">`,
				`<script type="module" src="/static/js/main.js?v=v1.0.0"></script>`,
			},
		},
		{
			name: "assetVersionが異なる場合",
			meta: viewmodel.PageMeta{
				Title:       "テストページ",
				Description: "説明",
				OGType:      "article",
				OGURL:       "https://annict.com",
				OGImage:     "https://annict.com/image.png",
			},
			assetVersion: "v2.0.0",
			wantContains: []string{
				`<link rel="stylesheet" href="/static/css/style.css?v=v2.0.0">`,
				`<script type="module" src="/static/js/main.js?v=v2.0.0"></script>`,
			},
		},
	}

	for _, tt := range tests {
		tt := tt // キャプチャ
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			// テンプレートをレンダリング
			var buf strings.Builder
			err := Head(tt.meta, tt.assetVersion).Render(ctx, &buf)
			if err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			html := buf.String()

			// 期待する文字列が含まれているか確認
			for _, want := range tt.wantContains {
				if !strings.Contains(html, want) {
					t.Errorf("期待する文字列が含まれていません: %q\nHTML: %s", want, html)
				}
			}
		})
	}
}

func TestHead_DarkMode(t *testing.T) {
	t.Parallel()

	meta := viewmodel.PageMeta{
		Title:       "テスト",
		Description: "説明",
		OGType:      "website",
		OGURL:       "https://annict.com",
		OGImage:     "https://annict.com/image.png",
	}

	ctx := context.Background()
	var buf strings.Builder
	err := Head(meta, "v1.0.0").Render(ctx, &buf)
	if err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	// ダークモード対応のスクリプトが含まれているか確認
	wantScript := `if (matchMedia("(prefers-color-scheme: dark)").matches)`
	if !strings.Contains(html, wantScript) {
		t.Errorf("ダークモード対応のスクリプトが含まれていません\nHTML: %s", html)
	}
}
