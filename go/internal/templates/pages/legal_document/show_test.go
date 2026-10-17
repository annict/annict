package legal_document

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
)

func renderShow(t *testing.T, ctx context.Context, data ShowPageData) string {
	t.Helper()

	var buf strings.Builder
	if err := Show(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	return buf.String()
}

// TestShow_LanguageSwitcherは、言語版が2つある文書で、相手言語版へのリンクを含む言語切替リンクを
// 表示し、表示中の版に印を付けることを検証する。
func TestShow_LanguageSwitcher(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), i18n.LangEn)
	html := renderShow(t, ctx, ShowPageData{
		Heading: "利用規約",
		HTML:    "<p>本文</p>",
		LanguageVersions: []LanguageVersion{
			{Lang: i18n.LangJa, Path: "/terms"},
			{Lang: i18n.LangEn, Path: "/en/terms", Current: true},
		},
	})

	for _, expected := range []string{
		`<nav aria-label="Language">`,
		// 言語名はページの言語に関わらずその言語自身の表記で書き、言語名の言語とリンク先の言語を示す
		`href="/terms" lang="ja" hreflang="ja">日本語</a>`,
		`href="/en/terms" lang="en" hreflang="en" aria-current="page">English</a>`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}

	if count := strings.Count(html, `aria-current="page"`); count != 1 {
		t.Errorf("aria-currentの個数 = %d、期待値 = 1", count)
	}
}

// TestShow_NoLanguageSwitcherは、言語版が1つしかない文書では言語切替リンクを表示しないことを検証する。
// 切り替える相手の言語版が無いため。
func TestShow_NoLanguageSwitcher(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), i18n.LangJa)
	html := renderShow(t, ctx, ShowPageData{
		Heading: "利用規約",
		HTML:    "<p>本文</p>",
		LanguageVersions: []LanguageVersion{
			{Lang: i18n.LangJa, Path: "/terms", Current: true},
		},
	})

	if strings.Contains(html, "<nav") {
		t.Error("言語版が1つの文書で言語切替リンクが表示された")
	}
	if !strings.Contains(html, "<p>本文</p>") {
		t.Error("本文が描画されていません")
	}
}

// TestShow_ArchiveNoticeは、改定前の版のページで、本文の前に適用期間と現行の版へのリンクを示す注意書きを
// 表示することを検証する。
func TestShow_ArchiveNotice(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), i18n.LangJa)
	html := renderShow(t, ctx, ShowPageData{
		Heading: "利用規約 (2014年3月10日施行版)",
		HTML:    "<p>本文</p>",
		ArchiveNotice: &ArchiveNotice{
			EffectiveOn: "2014年3月10日",
			EndedOn:     "2026年10月31日",
			CurrentPath: "/terms",
		},
	})

	for _, expected := range []string{
		`<div class="alert" data-variant="warning">`,
		"<h2>この文書は改定前の版です</h2>",
		"この版の適用期間は2014年3月10日から2026年10月31日までです。",
		`<a class="link" href="/terms">現行の版を見る</a>`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}

	// 読み始める前に改定前の版だと分かるよう、注意書きを本文より前に置く
	if strings.Index(html, "この文書は改定前の版です") > strings.Index(html, "<p>本文</p>") {
		t.Error("注意書きが本文より後に表示された")
	}
}

// TestShow_NoArchiveNoticeは、現行の版のページでは改定前の版の注意書きを表示しないことを検証する。
func TestShow_NoArchiveNotice(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), i18n.LangJa)
	html := renderShow(t, ctx, ShowPageData{
		Heading: "利用規約",
		HTML:    "<p>本文</p>",
	})

	if strings.Contains(html, `class="alert"`) {
		t.Error("現行の版のページで改定前の版の注意書きが表示された")
	}
}
