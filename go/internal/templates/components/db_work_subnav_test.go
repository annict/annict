package components

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/templates"
	"github.com/annict/annict/go/internal/viewmodel"
)

// TestDBWorkSubnav verifies the subnav links to every work sub-resource, labels the
// entries, exposes a navigation landmark, and marks the current (work edit) entry with
// aria-current.
//
// [Ja] TestDBWorkSubnav は、サブナビが作品の各サブリソースへリンクし、項目にラベルを付け、
// ナビゲーションランドマークを持ち、現在ページ (作品編集) の項目に aria-current を付ける
// ことを検証する。
func TestDBWorkSubnav(t *testing.T) {
	t.Parallel()

	ctx := templates.SetCurrentPath(i18n.SetLocale(context.Background(), "ja"), "/db/works/1/edit")

	var buf strings.Builder
	if err := DBWorkSubnav(DBWorkSubnavData{WorkID: viewmodel.WorkID(1)}).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	expectedContents := []string{
		`aria-label="作品ナビゲーション"`,
		`href="/db/works/1/edit"`,
		`href="/db/works/1/episodes"`,
		`href="/db/works/1/programs"`,
		`href="/db/works/1/slots"`,
		`href="/db/works/1/casts"`,
		`href="/db/works/1/staffs"`,
		`href="/db/works/1/image"`,
		`href="/db/works/1/trailers"`,
		// Labels mirror the Rails work subnav (noun.*): programs / slots keep the
		// broadcast-oriented wording of the Japanese source.
		//
		// [Ja] ラベルは Rails の作品サブナビ (noun.*) に合わせ、番組情報 / 放送予定は日本語の
		// 放送寄りの表現を保つ。
		"エピソード",
		"番組情報",
		"放送予定",
		"キャスト",
		"スタッフ",
		"作品画像",
		// The work edit page is the current page, so its entry is marked and filled.
		//
		// [Ja] 作品編集ページが現在ページのため、その項目に印が付き塗りつぶされる。
		`aria-current="page"`,
		`aria-[current=page]:bg-primary`,
		// The entries stay on one line and the list scrolls on its own, so the document
		// itself never overflows horizontally on small screens.
		//
		// [Ja] 項目は折り返さずリスト自身が横スクロールするため、小さい画面でも文書全体が
		// 横に溢れることはない。
		`flex-nowrap`,
		`overflow-x-auto`,
		`whitespace-nowrap`,
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}
}

// TestDBWorkSubnav_MarksCurrentEntryOnSubPage verifies an entry stays marked while a page
// below it is open, so adding an episode still shows the episodes entry as current and no
// other entry is marked.
//
// [Ja] TestDBWorkSubnav_MarksCurrentEntryOnSubPage は、項目の配下のページを開いている間も
// その項目に印が残ることを検証する。エピソードの追加中もエピソードの項目が現在ページとして
// 表示され、他の項目には印が付かない。
func TestDBWorkSubnav_MarksCurrentEntryOnSubPage(t *testing.T) {
	t.Parallel()

	ctx := templates.SetCurrentPath(i18n.SetLocale(context.Background(), "ja"), "/db/works/1/episodes/new")

	var buf strings.Builder
	if err := DBWorkSubnav(DBWorkSubnavData{WorkID: viewmodel.WorkID(1)}).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	episodes := subnavItemHTML(t, html, "/db/works/1/episodes")
	if !strings.Contains(episodes, `aria-current="page"`) {
		t.Error("エピソードの配下のページでエピソードの項目に aria-current が付いていません")
	}

	for _, path := range []string{
		"/db/works/1/edit",
		"/db/works/1/programs",
		"/db/works/1/slots",
		"/db/works/1/casts",
		"/db/works/1/staffs",
		"/db/works/1/image",
		"/db/works/1/trailers",
	} {
		if strings.Contains(subnavItemHTML(t, html, path), `aria-current="page"`) {
			t.Errorf("現在ページではない %q の項目に aria-current が付いています", path)
		}
	}
}

// subnavItemHTML returns the markup of the subnav entry linking to the given path.
//
// [Ja] subnavItemHTML は指定したパスへリンクするサブナビ項目のマークアップを返す。
func subnavItemHTML(t *testing.T, html string, path string) string {
	t.Helper()

	start := strings.Index(html, `<a href="`+path+`"`)
	if start < 0 {
		t.Fatalf("%q へのリンクが描画されていません", path)
	}
	end := strings.Index(html[start:], "</a>")
	if end < 0 {
		t.Fatalf("%q へのリンクが閉じられていません", path)
	}
	return html[start : start+end]
}

// TestDBWorkSubnav_OmitsEpisodeItemsWhenNoEpisodes verifies that the episode-derived
// entries (episodes, broadcast slots) are hidden when the work has no episodes, mirroring
// Rails.
//
// [Ja] TestDBWorkSubnav_OmitsEpisodeItemsWhenNoEpisodes は、作品が「エピソード無し」の
// ときエピソード由来の項目 (エピソード・放送予定) が隠れることを検証する (Rails に
// 合わせている)。
func TestDBWorkSubnav_OmitsEpisodeItemsWhenNoEpisodes(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	var buf strings.Builder
	if err := DBWorkSubnav(DBWorkSubnavData{WorkID: viewmodel.WorkID(1), NoEpisodes: true}).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	if strings.Contains(html, "/db/works/1/episodes") {
		t.Error("エピソード無しのときエピソード項目は描画されてはいけません")
	}
	if strings.Contains(html, "/db/works/1/slots") {
		t.Error("エピソード無しのとき放送予定項目は描画されてはいけません")
	}

	for _, expected := range []string{
		"/db/works/1/programs",
		"/db/works/1/casts",
		"/db/works/1/staffs",
		"/db/works/1/image",
		"/db/works/1/trailers",
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("エピソード非依存の項目は描画されるべきです: %q", expected)
		}
	}
}

// TestDBWorkSubnav_I18n verifies the landmark name and the entry labels switch per locale.
//
// [Ja] TestDBWorkSubnav_I18n はランドマーク名と項目ラベルが言語ごとに切り替わることを
// 検証する。
func TestDBWorkSubnav_I18n(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		locale   string
		expected []string
	}{
		{"日本語", "ja", []string{`aria-label="作品ナビゲーション"`, "キャスト"}},
		{"英語", "en", []string{`aria-label="Work navigation"`, "Casts"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), tt.locale)

			var buf strings.Builder
			if err := DBWorkSubnav(DBWorkSubnavData{WorkID: viewmodel.WorkID(1)}).Render(ctx, &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}
			for _, expected := range tt.expected {
				if !strings.Contains(buf.String(), expected) {
					t.Errorf("期待する文字列が含まれていません: %q", expected)
				}
			}
		})
	}
}
