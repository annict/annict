package components

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/viewmodel"
)

func TestHead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		meta            viewmodel.PageMeta
		assetVersion    string
		wantContains    []string
		wantNotContains []string
	}{
		{
			name: "基本的なメタタグが正しく出力される",
			meta: viewmodel.PageMeta{
				Title:        "テストページ | Annict",
				Description:  "テストページの説明",
				OGType:       "website",
				CanonicalURL: "https://annict.com/test",
				OGImage:      "https://annict.com/test.png",
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
			// A page that declares no origin gets no resource hint. Most public pages request
			// nothing from a third party, so warming up a connection for them would be a waste.
			//
			// [Ja] オリジンを宣言しないページにはリソースヒントを出さない。多くの公開ページは
			// 第三者に何もリクエストしないため、接続を張っても無駄になる。
			wantNotContains: []string{
				`rel="preconnect"`,
			},
		},
		{
			name: "宣言したオリジンが preconnect として出力される",
			meta: viewmodel.PageMeta{
				Title:             "Annictにログイン | Annict",
				Description:       "ログインページの説明",
				OGType:            "website",
				CanonicalURL:      "https://annict.com/sign_in",
				OGImage:           "https://annict.com/og-image.png",
				PreconnectOrigins: []string{"https://challenges.cloudflare.com"},
			},
			assetVersion: "v1.0.0",
			wantContains: []string{
				`<link rel="preconnect" href="https://challenges.cloudflare.com">`,
			},
		},
		{
			name: "assetVersionが異なる場合",
			meta: viewmodel.PageMeta{
				Title:        "テストページ",
				Description:  "説明",
				OGType:       "article",
				CanonicalURL: "https://annict.com",
				OGImage:      "https://annict.com/image.png",
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

			for _, notWant := range tt.wantNotContains {
				if strings.Contains(html, notWant) {
					t.Errorf("含まれてはいけない文字列が含まれています: %q\nHTML: %s", notWant, html)
				}
			}
		})
	}
}

// TestDBHead verifies the DB admin <head> renders the tags shared with the public pages
// plus the minimal Open Graph set, while dropping the rest of the tags addressing readers
// outside the page. The PageMeta it receives still carries description / OG values, so the
// assertions show the omission comes from DBHead rather than from empty input.
//
// [Ja] TestDBHead は DB 管理画面の <head> が公開ページと共通のタグと最小限の Open Graph を
// 描画し、それ以外のページ自身の外にいる読み手へ向けたタグを落とすことを検証する。渡す
// PageMeta は description や OG の値を持たせてあるため、出力されないのが入力の空ではなく
// DBHead によるものだと分かる。
func TestDBHead(t *testing.T) {
	t.Parallel()

	meta := viewmodel.PageMeta{
		Title:        "作品 | Annict DB",
		Description:  "テストページの説明",
		OGType:       "website",
		CanonicalURL: "https://annict.com/db/works",
		OGImage:      "https://annict.com/test.png",
	}

	ctx := context.Background()
	var buf strings.Builder
	if err := DBHead(meta, "v1.0.0").Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	wantContains := []string{
		`<meta charset="UTF-8">`,
		`<meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover">`,
		`<title>作品 | Annict DB</title>`,
		`<link rel="shortcut icon" href="/static/images/favicon.png" type="image/png">`,
		`<link rel="stylesheet" href="/static/css/style.css?v=v1.0.0">`,
		`<script type="module" src="/static/js/main.js?v=v1.0.0"></script>`,
		`if (matchMedia("(prefers-color-scheme: dark)").matches)`,
		`<meta property="og:title" content="作品 | Annict DB">`,
		`<meta property="og:type" content="website">`,
		`<meta property="og:url" content="https://annict.com/db/works">`,
		`<meta property="og:site_name" content="Annict (アニクト)">`,
	}
	for _, want := range wantContains {
		if !strings.Contains(html, want) {
			t.Errorf("期待する文字列が含まれていません: %q\nHTML: %s", want, html)
		}
	}

	// The Open Graph properties left out are the ones a titled card does not need:
	// description and image carry the public pages' generic copy and artwork, and locale
	// only matters for the pages that are actually shared.
	//
	// [Ja] 落とす Open Graph はタイトル付きカードに要らないもの。description と image は
	// 公開ページ向けの汎用の文言と画像であり、locale は実際に共有されるページでしか意味を
	// 持たない。
	wantNotContains := []string{
		`name="description"`,
		`property="og:description"`,
		`property="og:image"`,
		`property="og:locale"`,
		`name="twitter:`,
		`rel="canonical"`,
		`rel="manifest"`,
	}
	for _, notWant := range wantNotContains {
		if strings.Contains(html, notWant) {
			t.Errorf("公開ページ向けの出力が含まれてはいけません: %q\nHTML: %s", notWant, html)
		}
	}
}

// TestHead_PreconnectPosition verifies a declared origin is hinted before the document
// starts asking for anything. The hint only buys time while the rest of the head is still
// being parsed, so it has to precede the title and the assets rather than merely appear.
//
// [Ja] TestHead_PreconnectPosition は宣言したオリジンのヒントが、文書が何かを要求し始める
// 前に出ることを検証する。ヒントは head の残りを解析している間しか時間を稼げないため、
// 単に出力されるだけでなくタイトルやアセットより前にある必要がある。
func TestHead_PreconnectPosition(t *testing.T) {
	t.Parallel()

	meta := viewmodel.PageMeta{
		Title:             "Annictにログイン | Annict",
		Description:       "ログインページの説明",
		OGType:            "website",
		CanonicalURL:      "https://annict.com/sign_in",
		OGImage:           "https://annict.com/og-image.png",
		PreconnectOrigins: []string{"https://challenges.cloudflare.com"},
	}

	ctx := context.Background()
	var buf strings.Builder
	if err := Head(meta, "v1.0.0").Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	preconnectIndex := strings.Index(html, `rel="preconnect"`)
	if preconnectIndex < 0 {
		t.Fatalf("preconnectが出力されていません\nHTML: %s", html)
	}

	wantAfterPreconnect := []string{
		"<title>",
		`rel="stylesheet"`,
		"<script",
	}
	for _, want := range wantAfterPreconnect {
		index := strings.Index(html, want)
		if index < 0 {
			t.Errorf("比較対象が出力されていません: %q\nHTML: %s", want, html)
			continue
		}
		if index < preconnectIndex {
			t.Errorf("preconnectは %q より前に出力される必要があります\nHTML: %s", want, html)
		}
	}
}

func TestHead_DarkMode(t *testing.T) {
	t.Parallel()

	meta := viewmodel.PageMeta{
		Title:        "テスト",
		Description:  "説明",
		OGType:       "website",
		CanonicalURL: "https://annict.com",
		OGImage:      "https://annict.com/image.png",
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
