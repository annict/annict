package db_episodes

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/viewmodel"
)

func TestIndex_Empty(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	data := IndexPageData{
		WorkID:     1,
		WorkName:   "テストアニメ",
		Episodes:   []viewmodel.DBEpisodeListItem{},
		Pagination: viewmodel.NewPagination(1, 0, 100, "/db/works/1/episodes"),
	}

	var buf strings.Builder
	if err := Index(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	if strings.Contains(html, "<table") {
		t.Error("エピソードが空の場合は <table> が含まれてはいけません")
	}

	if !strings.Contains(html, "エピソードが見つかりませんでした") {
		t.Error("エピソードが空の場合は空メッセージが表示されるべきです")
	}

	// The subnav still renders on an empty list, so the visitor can move back to the work.
	//
	// [Ja] 一覧が空でもサブナビは描画され、閲覧者は作品へ戻れる。
	if !strings.Contains(html, `href="/db/works/1/edit"`) {
		t.Error("エピソードが空でも作品へ戻るサブナビが表示されるべきです")
	}
}

func TestIndex_WithEpisodes(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	data := IndexPageData{
		WorkID:   3,
		WorkName: "テストアニメ",
		Episodes: []viewmodel.DBEpisodeListItem{
			{
				ID:         10,
				WorkID:     3,
				Number:     "第2話",
				RawNumber:  "2",
				Title:      "エピソードタイトル",
				TitleEn:    "Episode Title",
				SortNumber: 200,
				Status:     viewmodel.PublishingStatusPublished,
			},
			{
				ID:         11,
				WorkID:     3,
				SortNumber: 100,
				Status:     viewmodel.PublishingStatusArchived,
			},
		},
		Pagination: viewmodel.NewPagination(1, 2, 100, "/db/works/3/episodes"),
	}

	var buf strings.Builder
	if err := Index(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	expectedContents := []string{
		"<table",
		"<thead",
		"<tbody",
		`<div class="overflow-x-auto" role="region" aria-label="DBエピソード一覧" tabindex="0">`,
		`<caption class="sr-only">DBエピソード一覧</caption>`,
		// The heading names the parent work, and the subnav links back to it.
		//
		// [Ja] 見出しは親作品を名指しし、サブナビはそこへ戻るリンクを持つ。
		"テストアニメ",
		`href="/db/works/3/edit"`,
		// The ID column links to the episode's public page in a new tab, and says so in its
		// accessible name.
		//
		// [Ja] ID 列はエピソードの公開ページを新しいタブで開くリンクで、そのことを
		// アクセシブルネームで伝える。
		`href="/works/3/episodes/10"`,
		`target="_blank"`,
		`rel="noopener"`,
		`aria-label="エピソード 10 を新しいタブで開く"`,
		// Both number representations and both titles.
		//
		// [Ja] 2 系統の話数と 2 つのタイトル。
		"第2話",
		"エピソードタイトル",
		"Episode Title",
		// Each stacked value carries a localized label, and titles carry their known
		// language so assistive technology can distinguish and pronounce them.
		//
		// [Ja] 縦積みした各値にはローカライズ済みのラベルを付け、タイトルには既知の言語を
		// 指定して、支援技術が値を区別して正しく発音できるようにする。
		"表示用話数:",
		"数値話数:",
		"日本語タイトル:",
		"英語タイトル:",
		`lang="ja">エピソードタイトル</span>`,
		`lang="en">Episode Title</span>`,
		// The sort number column.
		//
		// [Ja] 並び順の列。
		"<td>200</td>",
		"<td>100</td>",
		// The status badge comes from the shared status label component.
		//
		// [Ja] 状態のバッジは共有のステータスラベルコンポーネントが描画する。
		`<span class="badge" data-variant="success">公開</span>`,
		`<span class="badge" data-variant="warning">非公開</span>`,
	}

	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("出力に %q が含まれていません", expected)
		}
	}
}

func TestIndex_MissingValuesRenderPlaceholder(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	data := IndexPageData{
		WorkID:   1,
		WorkName: "テストアニメ",
		Episodes: []viewmodel.DBEpisodeListItem{
			{ID: 1, WorkID: 1, SortNumber: 100, Status: viewmodel.PublishingStatusPublished},
		},
		Pagination: viewmodel.NewPagination(1, 1, 100, "/db/works/1/episodes"),
	}

	var buf strings.Builder
	if err := Index(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	// Four unset attributes (display number, numeric number, Japanese title, English title)
	// each render the placeholder instead of an empty cell.
	//
	// [Ja] 未設定の 4 つの属性 (表示用話数・数値話数・日本語タイトル・英語タイトル) は
	// いずれも空のセルではなくプレースホルダーを描画する。
	if got := strings.Count(html, ": </span> -</div>"); got != 4 {
		t.Errorf("プレースホルダーの数 = %d, want 4", got)
	}
	if strings.Contains(html, `lang="ja">-`) || strings.Contains(html, `lang="en">-`) {
		t.Error("プレースホルダーには言語指定を付けてはいけません")
	}
}

// TestIndex_HeadingFallsBackWhenWorkNameEmpty covers the empty WorkName that
// viewmodel.DBEpisodeListWorkName produces for a work with no name to show. The document
// title falls back on the same signal, so the heading and the title stay in step.
//
// [Ja] TestIndex_HeadingFallsBackWhenWorkNameEmpty は、表示できる名前が無い作品に対して
// viewmodel.DBEpisodeListWorkName が返す空の WorkName を検証する。文書タイトルも同じ合図で
// フォールバックするため、見出しとタイトルの歩調が揃う。
func TestIndex_HeadingFallsBackWhenWorkNameEmpty(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	data := IndexPageData{
		WorkID:     1,
		WorkName:   "",
		Episodes:   []viewmodel.DBEpisodeListItem{},
		Pagination: viewmodel.NewPagination(1, 0, 100, "/db/works/1/episodes"),
	}

	var buf strings.Builder
	if err := Index(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	if !strings.Contains(buf.String(), "エピソード") {
		t.Error("作品の名前が無い場合は汎用のページタイトルが見出しになるべきです")
	}
}

func TestIndex_NoEpisodesWorkDropsSubnavEpisodeEntry(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	data := IndexPageData{
		WorkID:     1,
		WorkName:   "テストアニメ",
		NoEpisodes: true,
		Episodes:   []viewmodel.DBEpisodeListItem{},
		Pagination: viewmodel.NewPagination(1, 0, 100, "/db/works/1/episodes"),
	}

	var buf strings.Builder
	if err := Index(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	if strings.Contains(buf.String(), `href="/db/works/1/episodes"`) {
		t.Error("no_episodes の作品ではサブナビのエピソード項目が落ちるべきです")
	}
}
