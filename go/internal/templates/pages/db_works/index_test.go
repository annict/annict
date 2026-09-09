package db_works

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/viewmodel"
)

// newTabIconPath is the leading path data of the arrow-square-out icon, used to count how many
// new-tab link indicators a rendered page carries.
//
// [Ja] newTabIconPath は arrow-square-out アイコンの path データ先頭。描画されたページに
// 新規タブリンクの目印がいくつ出ているかを数えるために使う。
const newTabIconPath = "M224,104a8,8,0,0,1-16,0V59.32"

// TestIndex_Empty verifies that an empty work list renders no table and puts the empty state
// in its place, in each locale the site serves.
//
// [Ja] TestIndex_Empty は、作品が無いときに表が描画されず、その位置に空表示が出ることを、
// サイトが提供する各ロケールで検証する。
func TestIndex_Empty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		locale  string
		heading string
	}{
		{name: "日本語", locale: "ja", heading: "作品が見つかりませんでした"},
		{name: "英語", locale: "en", heading: "No works found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), tt.locale)
			data := IndexPageData{
				Works:      []viewmodel.DBWorkListItem{},
				Pagination: viewmodel.NewPagination(1, 0, 30, "/db/works"),
			}

			var buf strings.Builder
			if err := Index(data).Render(ctx, &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			html := buf.String()

			// テーブルが表示されないことを確認
			if strings.Contains(html, "<table") {
				t.Error("作品が空の場合は <table> が含まれてはいけません")
			}

			if !strings.Contains(html, `<section class="empty">`) {
				t.Error("作品が空の場合は空表示コンポーネントが表示されるべきです")
			}

			if want := "<h2>" + tt.heading + "</h2>"; !strings.Contains(html, want) {
				t.Errorf("作品が空の場合は空表示の見出し %q が表示されるべきです", want)
			}
		})
	}
}

// TestIndex_WithWorks は作品が存在する場合に表が表示されることをテスト
func TestIndex_WithWorks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	data := IndexPageData{
		Works: []viewmodel.DBWorkListItem{
			{
				ID:            1,
				Title:         "テストアニメ1",
				TitleKana:     "てすとあにめいち",
				TitleEn:       "Test Anime 1",
				Media:         "TV",
				Season:        "2024年春",
				Syobocal:      viewmodel.ExternalServiceLink{Label: "3524", URL: "http://cal.syoboi.jp/tid/3524"},
				MalAnime:      viewmodel.ExternalServiceLink{Label: "20", URL: "https://myanimelist.net/anime/20"},
				WatchersCount: 100,
				Status:        viewmodel.PublishingStatusPublished,
				Image:         viewmodel.NewWorkImage(`{"master":{"id":"workimage/1/image/master-abc.jpg","storage":"store"}}`, testutil.NewTestImageHelper()),
			},
			{
				ID:            2,
				Title:         "テストアニメ2",
				TitleKana:     "",
				TitleEn:       "",
				Media:         "OVA",
				Season:        "",
				WatchersCount: 50,
				Status:        viewmodel.PublishingStatusPublished,
				Image:         viewmodel.NewWorkImage("", testutil.NewTestImageHelper()),
			},
		},
		Pagination: viewmodel.NewPagination(1, 2, 30, "/db/works"),
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
		`<div class="overflow-x-auto" role="region" aria-label="DB作品一覧" tabindex="0">`,
		`<caption class="sr-only">DB作品一覧</caption>`,
		"テストアニメ1",
		"テストアニメ2",
		// ID column links to the work's public page in a new tab.
		//
		// [Ja] ID 列は作品の公開ページを新しいタブで開くリンクになる。
		`href="/works/1"`,
		`target="_blank"`,
		`rel="noopener"`,
		// The ID link exposes an aria-label announcing that it opens in a new tab.
		//
		// [Ja] ID リンクは新しいタブで開くことを知らせる aria-label を持つ。
		`aria-label="作品 1 を新しいタブで開く"`,
		// The extra kana / English titles.
		//
		// [Ja] 追加のふりがな・英語タイトル。
		"てすとあにめいち",
		"Test Anime 1",
		// Every work attribute other than the title is merged into a single "info" column,
		// rendered as label / value pairs inside one cell rather than one column each.
		//
		// [Ja] タイトル以外の作品属性は 1 つの「情報」列に統合され、属性ごとの列に分かれず
		// 1 セル内のラベル / 値の組として描画される。
		`<th scope="col" class="text-left">情報</th>`,
		"<dd>TV</dd>",
		"<dd>OVA</dd>",
		"<dd>2024年春</dd>",
		"ウォッチ数",
		"<dd>100</dd>",
		// A work with an image renders a thumbnail <picture> with webp / jpeg sources and
		// lazy loading, sized to the work-image ratio.
		//
		// [Ja] 画像がある作品は webp / jpeg のソースと遅延読み込みを備えたサムネイル <picture>
		// を、作品画像の比率のサイズで描画する。
		"<picture",
		`type="image/webp"`,
		`type="image/jpeg"`,
		`alt="テストアニメ1"`,
		`width="70"`,
		`height="93"`,
		`loading="lazy"`,
		// A work without an image falls back to the static placeholder, drawn in the same
		// box as a real thumbnail.
		//
		// [Ja] 画像がない作品は静的なプレースホルダーにフォールバックし、実サムネイルと同じ枠で描画される。
		`src="/static/images/no-work-image.png"`,
		`style="width:70px;height:93px;"`,
		// The status badge is one of the info column's pairs, labelled like the others.
		//
		// [Ja] ステータスのバッジも情報列の組の 1 つとして、他と同じくラベル付きで表示される。
		"ステータス",
		`<span class="badge" data-variant="success">`,
		// The info column shows the しょぼかる / MyAnimeList links for a work that has
		// ids, each opening in a new tab with rel="noopener".
		//
		// [Ja] 情報列は、ID を持つ作品のしょぼかる / MyAnimeList リンクを表示し、
		// それぞれ rel="noopener" 付きで新しいタブで開く。
		"しょぼかる",
		"MyAnimeList",
		`href="http://cal.syoboi.jp/tid/3524"`,
		`href="https://myanimelist.net/anime/20"`,
		`>3524<svg aria-hidden="true"`,
		// The title cell opts out of the table's default whitespace-nowrap so long titles wrap
		// inside the table's dedicated horizontal scroll region.
		//
		// [Ja] タイトルセルはテーブル既定の whitespace-nowrap を解除し、表専用の横スクロール
		// 領域内で長いタイトルが折り返すようにする。
		`class="whitespace-normal [overflow-wrap:anywhere]"`,
		// The table uses a fixed layout with a colgroup so the width-less title column absorbs
		// the remaining space; auto layout would spread the slack across all columns instead.
		//
		// [Ja] テーブルは colgroup 付きの固定レイアウトを使い、幅指定の無いタイトル列が残り幅を
		// 吸収する。auto レイアウトでは余白が全列へ分散してしまう。
		"table-fixed",
		// This viewer takes no row action, so the table keeps the narrower of the two floors
		// (the action column's width is not reserved).
		//
		// [Ja] この閲覧者は行の操作を持たないため、テーブルは 2 つの下限のうち狭いほうを保つ
		// (操作列の幅を確保しない)。
		"min-w-[544px]",
		"<colgroup>",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}

	// A work without kana / English titles renders a "-" placeholder in those lines.
	//
	// [Ja] ふりがな・英語タイトルが無い作品は該当行に "-" のプレースホルダーを表示する。
	if !strings.Contains(html, ">-</div>") {
		t.Error("ふりがな・英語タイトルが空の作品には '-' のプレースホルダーが表示されるべきです")
	}

	// The old image-existence column (a bare "✓") is replaced by the thumbnail column.
	//
	// [Ja] 旧「画像」有無列 (素の "✓") はサムネイル列に置き換えられている。
	if strings.Contains(html, "✓") {
		t.Error("旧「画像」有無列の '✓' は表示されてはいけません")
	}

	// A work with no external-service ids renders a "-" placeholder in the cell.
	//
	// [Ja] 外部サービスの ID を持たない作品は、セルに "-" のプレースホルダーを表示する。
	if !strings.Contains(html, "<dd>-</dd>") {
		t.Error("外部サービスの ID が無い作品には '-' のプレースホルダーが表示されるべきです")
	}

	// A work with no release season renders the same "-" placeholder as the other info
	// pairs, rather than an empty value that reads as a rendering gap. The label is matched
	// alongside it so this cannot be satisfied by some other pair's dash.
	//
	// [Ja] リリース時期が無い作品は、描画漏れに見える空の値ではなく、情報列の他の組と同じ "-"
	// のプレースホルダーを表示する。他の組の "-" で条件が満たされないよう、ラベルと併せて検証する。
	if !strings.Contains(html, `<dt class="text-muted-foreground">リリース時期</dt><dd>-</dd>`) {
		t.Error("リリース時期が空の作品には '-' のプレースホルダーが表示されるべきです")
	}

	// External-service links carry the same new-tab icon as the ID link. Work 1 renders
	// three icons (ID, しょぼかる, MyAnimeList); work 2 has no external ids, so only its
	// ID link renders one.
	//
	// [Ja] 外部サービスのリンクには ID リンクと同じ新規タブアイコンを付ける。作品 1 は 3 つ
	// (ID・しょぼかる・MyAnimeList)、作品 2 は外部 ID が無いため ID リンクの 1 つだけになる。
	if got := strings.Count(html, newTabIconPath); got != 4 {
		t.Errorf("新規タブアイコンの数 = %d, want 4", got)
	}

	// The external services and status columns are gone: both are now pairs inside the
	// info column, so neither has a header of its own anymore.
	//
	// [Ja] 外部サービス列とステータス列は情報列内の組になったため、専用の見出しはもう無い。
	if strings.Contains(html, `<th scope="col" class="text-left">外部サービス</th>`) {
		t.Error("外部サービス列の見出しは情報列への統合で無くなっているべきです")
	}
	if strings.Contains(html, `<th scope="col" class="text-left">ステータス</th>`) {
		t.Error("ステータス列の見出しは情報列への統合で無くなっているべきです")
	}
}

// TestExternalServiceLinkI18n verifies that an external-service link announces that it opens
// in a new tab in each supported locale while retaining its visible identifier in the name.
//
// [Ja] TestExternalServiceLinkI18n は外部サービスリンクが可視 ID を名前に残しつつ、新しい
// タブで開くことを各対応言語で伝えることを検証する。
func TestExternalServiceLinkI18n(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		locale string
		want   string
	}{
		{name: "English", locale: "en", want: `aria-label="Open 3524 in a new tab"`},
		{name: "Japanese", locale: "ja", want: `aria-label="3524 を新しいタブで開く"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), tt.locale)
			var buf strings.Builder
			link := viewmodel.ExternalServiceLink{Label: "3524", URL: "http://cal.syoboi.jp/tid/3524"}
			if err := externalServiceLink(link).Render(ctx, &buf); err != nil {
				t.Fatalf("failed to render external-service link: %v", err)
			}

			if html := buf.String(); !strings.Contains(html, tt.want) {
				t.Errorf("localized accessible name is missing: want %q in %q", tt.want, html)
			}
		})
	}
}

// TestIndex_DecorativeIconsAreHidden covers the new-tab icons of the work list: the one in the
// ID link and the one in each external-service link. Every link already announces where it goes
// in its accessible name, so none of the icons reaches the accessibility tree or the focus
// order. Counting them against the icon's own path data holds the whole column rather than the
// first row: an icon added later without the helper would leave the two counts apart.
//
// [Ja] TestIndex_DecorativeIconsAreHidden は作品一覧の新規タブアイコン (ID リンクと外部サービス
// リンクのそれぞれ) を検証する。どのリンクもアクセシブルネームで行き先を既に伝えるため、
// アイコンはアクセシビリティツリーにもフォーカス順序にも出ない。アイコン自身の path データと
// 個数を突き合わせることで、先頭行だけでなく列全体を固定する (後からヘルパーを通さずに足された
// アイコンがあれば、2 つの個数がずれる)。
func TestIndex_DecorativeIconsAreHidden(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	data := IndexPageData{
		Works: []viewmodel.DBWorkListItem{
			{
				ID:       1,
				Title:    "テストアニメ1",
				Media:    "TV",
				Status:   viewmodel.PublishingStatusPublished,
				Syobocal: viewmodel.ExternalServiceLink{Label: "3524", URL: "http://cal.syoboi.jp/tid/3524"},
				MalAnime: viewmodel.ExternalServiceLink{Label: "20", URL: "https://myanimelist.net/anime/20"},
				Image:    viewmodel.NewWorkImage("", testutil.NewTestImageHelper()),
			},
		},
		Pagination: viewmodel.NewPagination(1, 1, 30, "/db/works"),
	}

	var buf strings.Builder
	if err := Index(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	icons := strings.Count(html, newTabIconPath)
	if icons == 0 {
		t.Fatal("新規タブアイコンが描画されていません")
	}
	if got := strings.Count(html, decorativeIconMarkup(t, ctx, "arrow-square-out-regular", "w-[18px] h-[18px]")); got != icons {
		t.Errorf("装飾アイコンとして描画された数が一致しません: got %d, want %d", got, icons)
	}
	if strings.Contains(html, iconWrapperMarkup) {
		t.Error("装飾アイコンはラッパー要素ではなく SVG 自体で隠すべきです")
	}
}

// actionColumnWorks returns one published and one archived work for the action-column tests.
//
// [Ja] actionColumnWorks は操作列テスト用に、公開中と非公開の作品を 1 件ずつ返す。
func actionColumnWorks() []viewmodel.DBWorkListItem {
	return []viewmodel.DBWorkListItem{
		{ID: 1, Title: "公開作品", Media: "TV", Status: viewmodel.PublishingStatusPublished},
		{ID: 2, Title: "非公開作品", Media: "TV", Status: viewmodel.PublishingStatusArchived},
	}
}

func renderIndex(t *testing.T, data IndexPageData) string {
	t.Helper()
	var buf strings.Builder
	if err := Index(data).Render(context.Background(), &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	return buf.String()
}

func assertActionColumnStructure(t *testing.T, html string, wantActions bool) {
	t.Helper()

	wantColumns := 4
	wantMinWidth := "min-w-[544px]"
	wantAbsentMinWidth := "min-w-[640px]"
	if wantActions {
		wantColumns = 5
		wantMinWidth = "min-w-[640px]"
		wantAbsentMinWidth = "min-w-[544px]"
	}

	// Count all three parts of the table structure together. A conditional column must add or
	// remove its col, header and one data cell per row as a unit so assistive technology sees the
	// same relationships as the visual layout.
	//
	// [Ja] テーブル構造の 3 部分をまとめて数える。条件付きの列は col・見出し・各行のデータ
	// セルを一体として増減させ、支援技術にも視覚的な配置と同じ関係を伝える必要がある。
	if got := strings.Count(html, "<col ") + strings.Count(html, "<col>"); got != wantColumns {
		t.Errorf("列要素数 = %d, want %d", got, wantColumns)
	}
	if got := strings.Count(html, "<th "); got != wantColumns {
		t.Errorf("列見出し数 = %d, want %d", got, wantColumns)
	}
	if got, want := strings.Count(html, "<td"), wantColumns*len(actionColumnWorks()); got != want {
		t.Errorf("データセル数 = %d, want %d", got, want)
	}
	if !strings.Contains(html, wantMinWidth) {
		t.Errorf("テーブルに最小幅 %q がありません", wantMinWidth)
	}
	if strings.Contains(html, wantAbsentMinWidth) {
		t.Errorf("テーブルに不要な最小幅 %q が残っています", wantAbsentMinWidth)
	}
}

// TestIndex_ActionColumn_Committer verifies that a committer (non-admin) sees the edit link on
// every row, the unpublish link (to the confirmation screen) on published rows, and the publish
// htmx DELETE button on archived rows, but not the admin-only delete button.
//
// [Ja] TestIndex_ActionColumn_Committer は committer (非 admin) が各行に編集リンクを、公開中の
// 行に非公開リンク (確認画面へ) を、非公開の行に公開の htmx DELETE ボタンを見る一方、admin
// 専用の削除ボタンは見えないことを検証する。
func TestIndex_ActionColumn_Committer(t *testing.T) {
	t.Parallel()

	html := renderIndex(t, IndexPageData{
		Works:       actionColumnWorks(),
		Pagination:  viewmodel.NewPagination(1, 2, 30, "/db/works"),
		IsCommitter: true,
		IsAdmin:     false,
		CSRFToken:   "test-csrf-token",
	})
	assertActionColumnStructure(t, html, true)

	wantPresent := []string{
		// The column is announced by its own header.
		//
		// [Ja] 列は専用の見出しで示される。
		`<th scope="col" class="text-center">操作</th>`,
		// Edit link on both rows.
		//
		// [Ja] 両行の編集リンク。
		`href="/db/works/1/edit"`,
		`href="/db/works/2/edit"`,
		// Published row (1): unpublish link to the confirmation screen.
		//
		// [Ja] 公開中の行 (1): 確認画面への非公開リンク。
		`href="/db/works/1/archive/new"`,
		// Archived row (2): publish is an htmx DELETE against the archive path, with a confirm
		// dialog naming the work and the CSRF token carried in the X-CSRF-Token header.
		//
		// [Ja] 非公開の行 (2): 公開は archive パスへの htmx DELETE で、対象作品を名指しする
		// 確認ダイアログと X-CSRF-Token ヘッダーで送る CSRF トークンを伴う。
		`hx-delete="/db/works/2/archive"`,
		"作品 2 を公開しますか",
		"X-CSRF-Token",
		"test-csrf-token",
		// Every control extends its accessible name with the work it acts on, so a page of rows
		// does not present controls that all read the same. The visible label stays at the start
		// of the name.
		//
		// [Ja] 各コントロールはアクセシブルネームに対象の作品を足し、行が並ぶページで同じ名前の
		// コントロールばかりにならないようにする。可視ラベルは名前の先頭に残る。
		`編集<span class="sr-only"> 作品 1</span>`,
		`編集<span class="sr-only"> 作品 2</span>`,
		`非公開<span class="sr-only"> 作品 1</span>`,
		`公開<span class="sr-only"> 作品 2</span>`,
	}
	for _, expected := range wantPresent {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}

	wantAbsent := []string{
		// Published row has no publish button, archived row has no unpublish link.
		//
		// [Ja] 公開中の行に公開ボタンは無く、非公開の行に非公開リンクは無い。
		`hx-delete="/db/works/1/archive"`,
		`href="/db/works/2/archive/new"`,
		// No admin-only delete buttons for either row.
		//
		// [Ja] どちらの行にも admin 専用の削除ボタンは無い。
		`hx-delete="/db/works/1"`,
		`hx-delete="/db/works/2"`,
		"を削除しますか",
	}
	for _, unexpected := range wantAbsent {
		if strings.Contains(html, unexpected) {
			t.Errorf("含まれてはいけない文字列が含まれています: %q", unexpected)
		}
	}
}

// TestIndex_ActionColumn_Admin verifies that an admin additionally sees the delete htmx DELETE
// button (against the work path) on every row.
//
// [Ja] TestIndex_ActionColumn_Admin は admin がさらに各行に削除の htmx DELETE ボタン
// (work パスへの DELETE) を見ることを検証する。
func TestIndex_ActionColumn_Admin(t *testing.T) {
	t.Parallel()

	html := renderIndex(t, IndexPageData{
		Works:       actionColumnWorks(),
		Pagination:  viewmodel.NewPagination(1, 2, 30, "/db/works"),
		IsCommitter: true,
		IsAdmin:     true,
		CSRFToken:   "test-csrf-token",
	})

	wantPresent := []string{
		// Delete button on both rows targets the work path (distinct from the archive path).
		//
		// [Ja] 両行の削除ボタンは work パス (archive パスとは別) を DELETE 対象にする。
		`hx-delete="/db/works/1"`,
		`hx-delete="/db/works/2"`,
		"作品 1 を削除しますか",
		"作品 2 を削除しますか",
		`削除<span class="sr-only"> 作品 1</span>`,
		// The committer actions are still present.
		//
		// [Ja] committer の操作も引き続き表示される。
		`href="/db/works/1/archive/new"`,
		`hx-delete="/db/works/2/archive"`,
		"X-CSRF-Token",
	}
	for _, expected := range wantPresent {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}
}

// TestIndex_ActionColumn_Anonymous verifies that a signed-out or regular visitor gets no action
// column at all (the list itself stays public). The column is dropped rather than left empty:
// a header announcing a column that holds nothing is read out on every row, and the table
// already needs horizontal scrolling at mobile widths.
//
// [Ja] TestIndex_ActionColumn_Anonymous は未ログインや一般ユーザーに操作列そのものが
// 出ない (一覧自体は公開のまま) ことを検証する。列を空のまま残さず落とすのは、何も入らない列の
// 見出しが行ごとに読み上げられるうえ、このテーブルがモバイル幅では既に横スクロールを
// 要するため。
func TestIndex_ActionColumn_Anonymous(t *testing.T) {
	t.Parallel()

	html := renderIndex(t, IndexPageData{
		Works:       actionColumnWorks(),
		Pagination:  viewmodel.NewPagination(1, 2, 30, "/db/works"),
		IsCommitter: false,
		IsAdmin:     false,
	})
	assertActionColumnStructure(t, html, false)

	wantAbsent := []string{
		`<th scope="col" class="text-center">操作</th>`,
		`href="/db/works/1/edit"`,
		`href="/db/works/1/archive/new"`,
		"hx-delete=",
	}
	for _, unexpected := range wantAbsent {
		if strings.Contains(html, unexpected) {
			t.Errorf("含まれてはいけない文字列が含まれています: %q", unexpected)
		}
	}
}

// TestIndex_SidebarToggle verifies the page renders the sidebar toggle in its title row.
//
// [Ja] TestIndex_SidebarToggle はページがタイトル行にサイドバートグルを描画する
// ことを検証する。
func TestIndex_SidebarToggle(t *testing.T) {
	t.Parallel()

	html := renderIndex(t, IndexPageData{
		Works:      []viewmodel.DBWorkListItem{},
		Pagination: viewmodel.NewPagination(1, 0, 30, "/db/works"),
	})

	// The toggle is wired to the sidebar at every viewport size.
	//
	// [Ja] トグルはサイドバーに結線され、全画面幅で利用できる。
	for _, expected := range []string{`data-sidebar-toggle="db-sidebar"`} {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}
}

// TestIndex_FilterUI verifies the release-season combobox and no-slots checkbox markup.
//
// [Ja] TestIndex_FilterUI はリリース時期の combobox と放送予定未登録チェックボックスの
// マークアップを検証する。
func TestIndex_FilterUI(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	data := IndexPageData{
		Works:         []viewmodel.DBWorkListItem{},
		Pagination:    viewmodel.NewPagination(1, 0, 100, "/db/works"),
		FilterNoSlots: true,
		SeasonFilterOptions: []viewmodel.SeasonFilterOption{
			{Slug: "2024-spring", Label: "2024年春", Selected: true},
			{Slug: "2024-winter", Label: "2024年冬", Selected: false},
		},
	}

	var buf strings.Builder
	if err := Index(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	expectedContents := []string{
		// The native multiple select remains the durable named form control, including the
		// server-rendered initial selection.
		//
		// [Ja] ネイティブの複数選択 select は name 付きの常用可能なフォームコントロールとして残り、
		// サーバー描画時の初期選択も保持する。
		`<label id="db-works-season-slugs-label" for="db-works-season-slugs-select" class="label">`,
		`<select id="db-works-season-slugs-select" name="season_slugs" class="select w-full" data-season-slugs-select multiple size="6">`,
		`<option value="2024-spring" selected>2024年春</option>`,
		`<option value="2024-winter">2024年冬</option>`,
		// The hidden release-season combobox carries the same initial selection. Client code
		// reveals it only after Basecoat initializes and can synchronize it to the select.
		//
		// [Ja] 非表示のリリース時期 combobox にも同じ初期選択を渡す。クライアントコードは
		// Basecoat 初期化後、select へ同期できる状態になってからだけ表示する。
		`data-season-slugs-combobox`,
		"hidden",
		`aria-multiselectable="true"`,
		`<div role="option" data-value="2024-spring" aria-selected="true">2024年春</div>`,
		`<div role="option" data-value="2024-winter">2024年冬</div>`,
		// The no-slots checkbox renders in its checked state because FilterNoSlots is true.
		//
		// [Ja] FilterNoSlots が true なので放送予定未登録チェックボックスが checked 状態で描画される。
		`<input type="checkbox" name="filter_no_slots" value="1" checked>`,
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}

	// Basecoat's JSON-array hidden input stays unnamed: the native select is the sole named
	// season_slugs control and therefore submits repeated query parameters in every mode.
	//
	// [Ja] Basecoat の JSON 配列 hidden input には name を付けない。ネイティブ select だけを
	// season_slugs の name 付きコントロールとし、どのモードでも繰り返しクエリを送信する。
	if strings.Contains(html, `<input type="hidden" name="season_slugs"`) {
		t.Errorf("Basecoat の hidden input に season_slugs の name を付けてはいけません")
	}

	// The filter card opts out of the default overflow-hidden so the combobox popover is not
	// clipped by the card boundary.
	//
	// [Ja] フィルタカードは既定の overflow-hidden を外し、combobox の popover がカード境界で
	// 切られないようにする。
	if !strings.Contains(html, "overflow-visible") {
		t.Errorf("フィルタカードに overflow-visible が付いていません (popover がクリップされます)")
	}
}

// TestSeasonSlugsFilter_RemoveLabelLocale verifies that the bridge receives the localized
// template used for Basecoat-generated chip remove labels.
//
// [Ja] TestSeasonSlugsFilter_RemoveLabelLocale は Basecoat が生成するチップの選択解除ラベルに
// bridge が使うローカライズ済みテンプレートを受け取ることを検証する。
func TestSeasonSlugsFilter_RemoveLabelLocale(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		locale string
		want   string
	}{
		{name: "English", locale: "en", want: `data-season-slugs-remove-label="Remove {label}"`},
		{name: "Japanese", locale: "ja", want: `data-season-slugs-remove-label="{label}の選択を解除"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), tt.locale)
			var buf strings.Builder
			if err := seasonSlugsFilter(nil).Render(ctx, &buf); err != nil {
				t.Fatalf("failed to render release-season filter: %v", err)
			}

			if html := buf.String(); !strings.Contains(html, tt.want) {
				t.Errorf("localized remove-label template is missing: want %q in %q", tt.want, html)
			}
		})
	}
}
