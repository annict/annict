package db_works

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/viewmodel"
)

// newTabIconPathはarrow-square-outアイコンのpathデータ先頭。描画されたページに
// 新規タブリンクの目印がいくつ出ているかを数えるために使う。
const newTabIconPath = "M224,104a8,8,0,0,1-16,0V59.32"

// TestIndex_Emptyは、作品が無いときに表が描画されず、その位置に空表示が出ることを、
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
				t.Errorf("作品が空の場合は空表示の見出し%qが表示されるべきです", want)
			}
		})
	}
}

// TestIndex_WithWorksは作品が存在する場合に表が表示されることをテスト
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
		// ID列は作品の公開ページを新しいタブで開くリンクになる。
		`href="/works/1"`,
		`target="_blank"`,
		`rel="noopener"`,
		// IDリンクは新しいタブで開くことを知らせるaria-labelを持つ。
		`aria-label="作品 1 を新しいタブで開く"`,
		// 追加のふりがな・英語タイトル。
		"てすとあにめいち",
		"Test Anime 1",
		// タイトル以外の作品属性は1つの「情報」列に統合され、属性ごとの列に分かれず
		// 1セル内のラベル / 値の組として描画される。
		`<th scope="col" class="text-left">情報</th>`,
		"<dd>TV</dd>",
		"<dd>OVA</dd>",
		"<dd>2024年春</dd>",
		"ウォッチ数",
		"<dd>100</dd>",
		// 画像がある作品はwebp / jpegのソースと遅延読み込みを備えたサムネイル <picture>
		// を、作品画像の比率のサイズで描画する。
		"<picture",
		`type="image/webp"`,
		`type="image/jpeg"`,
		`alt="テストアニメ1"`,
		`width="70"`,
		`height="93"`,
		`loading="lazy"`,
		// 画像がない作品は静的なプレースホルダーにフォールバックし、実サムネイルと同じ枠で描画される。
		`src="/static/images/no-work-image.png"`,
		`style="width:70px;height:93px;"`,
		// ステータスのバッジも情報列の組の1つとして、他と同じくラベル付きで表示される。
		"ステータス",
		`<span class="badge" data-variant="success">`,
		// 情報列は、IDを持つ作品のしょぼかる / MyAnimeListリンクを表示し、
		// それぞれrel="noopener" 付きで新しいタブで開く。
		"しょぼかる",
		"MyAnimeList",
		`href="http://cal.syoboi.jp/tid/3524"`,
		`href="https://myanimelist.net/anime/20"`,
		`>3524<svg aria-hidden="true"`,
		// タイトルセルはテーブル既定のwhitespace-nowrapを解除し、表専用の横スクロール
		// 領域内で長いタイトルが折り返すようにする。
		`class="whitespace-normal [overflow-wrap:anywhere]"`,
		// テーブルはcolgroup付きの固定レイアウトを使い、幅指定の無いタイトル列が残り幅を
		// 吸収する。autoレイアウトでは余白が全列へ分散してしまう。
		"table-fixed",
		// この閲覧者は行の操作を持たないため、テーブルは2つの下限のうち狭いほうを保つ
		// (操作列の幅を確保しない)。
		"min-w-[544px]",
		"<colgroup>",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}

	// ふりがな・英語タイトルが無い作品は該当行に "-" のプレースホルダーを表示する。
	if !strings.Contains(html, ">-</div>") {
		t.Error("ふりがな・英語タイトルが空の作品には '-' のプレースホルダーが表示されるべきです")
	}

	// 旧「画像」有無列 (素の "✓") はサムネイル列に置き換えられている。
	if strings.Contains(html, "✓") {
		t.Error("旧「画像」有無列の '✓' は表示されてはいけません")
	}

	// 外部サービスのIDを持たない作品は、セルに "-" のプレースホルダーを表示する。
	if !strings.Contains(html, "<dd>-</dd>") {
		t.Error("外部サービスのIDが無い作品には '-' のプレースホルダーが表示されるべきです")
	}

	// リリース時期が無い作品は、描画漏れに見える空の値ではなく、情報列の他の組と同じ "-"
	// のプレースホルダーを表示する。他の組の "-" で条件が満たされないよう、ラベルと併せて検証する。
	if !strings.Contains(html, `<dt class="text-muted-foreground">リリース時期</dt><dd>-</dd>`) {
		t.Error("リリース時期が空の作品には '-' のプレースホルダーが表示されるべきです")
	}

	// 外部サービスのリンクにはIDリンクと同じ新規タブアイコンを付ける。作品1は3つ
	// (ID・しょぼかる・MyAnimeList)、作品2は外部IDが無いためIDリンクの1つだけになる。
	if got := strings.Count(html, newTabIconPath); got != 4 {
		t.Errorf("新規タブアイコンの数 = %d、期待値 = 4", got)
	}

	// 外部サービス列とステータス列は情報列内の組になったため、専用の見出しはもう無い。
	if strings.Contains(html, `<th scope="col" class="text-left">外部サービス</th>`) {
		t.Error("外部サービス列の見出しは情報列への統合で無くなっているべきです")
	}
	if strings.Contains(html, `<th scope="col" class="text-left">ステータス</th>`) {
		t.Error("ステータス列の見出しは情報列への統合で無くなっているべきです")
	}
}

// TestExternalServiceLinkI18nは外部サービスリンクが可視IDを名前に残しつつ、新しい
// タブで開くことを各対応言語で伝えることを検証する。
func TestExternalServiceLinkI18n(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		locale string
		want   string
	}{
		{name: "英語", locale: "en", want: `aria-label="Open 3524 in a new tab"`},
		{name: "日本語", locale: "ja", want: `aria-label="3524 を新しいタブで開く"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), tt.locale)
			var buf strings.Builder
			link := viewmodel.ExternalServiceLink{Label: "3524", URL: "http://cal.syoboi.jp/tid/3524"}
			if err := externalServiceLink(link).Render(ctx, &buf); err != nil {
				t.Fatalf("外部サービスのリンクの描画エラー = %v", err)
			}

			if html := buf.String(); !strings.Contains(html, tt.want) {
				t.Errorf("翻訳済みのアクセシブル名が無い。期待する文字列 = %q、HTML = %q", tt.want, html)
			}
		})
	}
}

// TestIndex_DecorativeIconsAreHiddenは作品一覧の新規タブアイコン (IDリンクと外部サービス
// リンクのそれぞれ) を検証する。どのリンクもアクセシブルネームで行き先を既に伝えるため、
// アイコンはアクセシビリティツリーにもフォーカス順序にも出ない。アイコン自身のpathデータと
// 個数を突き合わせることで、先頭行だけでなく列全体を固定する (後からヘルパーを通さずに足された
// アイコンがあれば、2つの個数がずれる)。
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
		t.Errorf("装飾アイコンとして描画された数 = %d、期待値 = %d", got, icons)
	}
	if strings.Contains(html, iconWrapperMarkup) {
		t.Error("装飾アイコンはラッパー要素ではなくSVG自体で隠すべきです")
	}
}

// actionColumnWorksは操作列テスト用に、公開中と非公開の作品を1件ずつ返す。
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

	// テーブル構造の3部分をまとめて数える。条件付きの列はcol・見出し・各行のデータ
	// セルを一体として増減させ、支援技術にも視覚的な配置と同じ関係を伝える必要がある。
	if got := strings.Count(html, "<col ") + strings.Count(html, "<col>"); got != wantColumns {
		t.Errorf("列要素数 = %d、期待値 = %d", got, wantColumns)
	}
	if got := strings.Count(html, "<th "); got != wantColumns {
		t.Errorf("列見出し数 = %d、期待値 = %d", got, wantColumns)
	}
	if got, want := strings.Count(html, "<td"), wantColumns*len(actionColumnWorks()); got != want {
		t.Errorf("データセル数 = %d、期待値 = %d", got, want)
	}
	if !strings.Contains(html, wantMinWidth) {
		t.Errorf("テーブルに最小幅%qがありません", wantMinWidth)
	}
	if strings.Contains(html, wantAbsentMinWidth) {
		t.Errorf("テーブルに不要な最小幅%qが残っています", wantAbsentMinWidth)
	}
}

// TestIndex_ActionColumn_Committerはcommitter (非admin) が各行に編集リンクを、公開中の
// 行に非公開リンク (確認画面へ) を、非公開の行に公開のhtmx DELETEボタンを見る一方、admin
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
		// 列は専用の見出しで示される。
		`<th scope="col" class="text-center">操作</th>`,
		// 両行の編集リンク。
		`href="/db/works/1/edit"`,
		`href="/db/works/2/edit"`,
		// 公開中の行 (1): 確認画面への非公開リンク。
		`href="/db/works/1/archive/new"`,
		// 非公開の行 (2): 公開はarchiveパスへのhtmx DELETEで、対象作品を名指しする
		// 確認ダイアログとX-CSRF-Tokenヘッダーで送るCSRFトークンを伴う。
		`hx-delete="/db/works/2/archive"`,
		"作品 2 を公開しますか",
		"X-CSRF-Token",
		"test-csrf-token",
		// 各コントロールはアクセシブルネームに対象の作品を足し、行が並ぶページで同じ名前の
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
		// 公開中の行に公開ボタンは無く、非公開の行に非公開リンクは無い。
		`hx-delete="/db/works/1/archive"`,
		`href="/db/works/2/archive/new"`,
		// どちらの行にもadmin専用の削除ボタンは無い。
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

// TestIndex_ActionColumn_Adminはadminがさらに各行に削除のhtmx DELETEボタン
// (workパスへのDELETE) を見ることを検証する。
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
		// 両行の削除ボタンはworkパス (archiveパスとは別) をDELETE対象にする。
		`hx-delete="/db/works/1"`,
		`hx-delete="/db/works/2"`,
		"作品 1 を削除しますか",
		"作品 2 を削除しますか",
		`削除<span class="sr-only"> 作品 1</span>`,
		// committerの操作も引き続き表示される。
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

// TestIndex_ActionColumn_Anonymousは未ログインや一般ユーザーに操作列そのものが
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

// TestIndex_NewWorkLinkIsCommitterOnlyは見出しの「新規登録」をcommitterにだけ出すことを
// 検証する。一覧は公開のため、登録画面に入れない閲覧者には403が返るだけのリンクを出さない。
func TestIndex_NewWorkLinkIsCommitterOnly(t *testing.T) {
	t.Parallel()

	newLink := `href="/db/works/new"`
	// actionsContainerは見出しが操作の周りに描画するラッパー。操作の無い閲覧者にはこれも
	// 出さない。空のラッパーはモバイル幅では単独で全幅のflex行になり、公開されている一覧の
	// 見出しの下に余白を足してしまうため。
	actionsContainer := `<div class="flex w-full flex-none justify-end gap-2 md:w-auto">`

	tests := []struct {
		name        string
		isCommitter bool
		want        bool
	}{
		{name: "committerには出す", isCommitter: true, want: true},
		{name: "committerでない閲覧者には出さない", isCommitter: false, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			html := renderIndex(t, IndexPageData{
				Works:       []viewmodel.DBWorkListItem{},
				Pagination:  viewmodel.NewPagination(1, 0, 30, "/db/works"),
				IsCommitter: tt.isCommitter,
			})

			if got := strings.Contains(html, newLink); got != tt.want {
				t.Errorf("登録画面へのリンクの有無 = %v、期待値 = %v", got, tt.want)
			}
			if got := strings.Contains(html, actionsContainer); got != tt.want {
				t.Errorf("見出しの操作コンテナの有無 = %v、期待値 = %v", got, tt.want)
			}
		})
	}
}

// TestIndex_SidebarToggleはページがタイトル行にサイドバートグルを描画する
// ことを検証する。
func TestIndex_SidebarToggle(t *testing.T) {
	t.Parallel()

	html := renderIndex(t, IndexPageData{
		Works:      []viewmodel.DBWorkListItem{},
		Pagination: viewmodel.NewPagination(1, 0, 30, "/db/works"),
	})

	// トグルはサイドバーに結線され、全画面幅で利用できる。
	for _, expected := range []string{`data-sidebar-toggle="db-sidebar"`} {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}
}

// TestIndex_FilterUIはリリース時期のcomboboxと放送予定未登録チェックボックスの
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
		// ネイティブの複数選択selectはname付きの常用可能なフォームコントロールとして残り、
		// サーバー描画時の初期選択も保持する。
		`<label id="db-works-season-slugs-label" for="db-works-season-slugs-select" class="label">`,
		`<select id="db-works-season-slugs-select" name="season_slugs" class="select w-full" data-season-slugs-select multiple size="6">`,
		`<option value="2024-spring" selected>2024年春</option>`,
		`<option value="2024-winter">2024年冬</option>`,
		// 非表示のリリース時期comboboxにも同じ初期選択を渡す。クライアントコードは
		// Basecoat初期化後、selectへ同期できる状態になってからだけ表示する。
		`data-season-slugs-combobox`,
		"hidden",
		`aria-multiselectable="true"`,
		`<div role="option" data-value="2024-spring" aria-selected="true">2024年春</div>`,
		`<div role="option" data-value="2024-winter">2024年冬</div>`,
		// FilterNoSlotsがtrueなので放送予定未登録チェックボックスがchecked状態で描画される。
		`<input type="checkbox" name="filter_no_slots" value="1" checked>`,
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}

	// BasecoatのJSON配列hidden inputにはnameを付けない。ネイティブselectだけを
	// season_slugsのname付きコントロールとし、どのモードでも繰り返しクエリを送信する。
	if strings.Contains(html, `<input type="hidden" name="season_slugs"`) {
		t.Errorf("Basecoatのhidden inputにseason_slugsのnameを付けてはいけません")
	}

	// フィルタカードは既定のoverflow-hiddenを外し、comboboxのpopoverがカード境界で
	// 切られないようにする。
	if !strings.Contains(html, "overflow-visible") {
		t.Errorf("フィルタカードにoverflow-visibleが付いていません (popoverがクリップされます)")
	}
}

// TestSeasonSlugsFilter_RemoveLabelLocaleはBasecoatが生成するチップの選択解除ラベルに
// bridgeが使うローカライズ済みテンプレートを受け取ることを検証する。
func TestSeasonSlugsFilter_RemoveLabelLocale(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		locale string
		want   string
	}{
		{name: "英語", locale: "en", want: `data-season-slugs-remove-label="Remove {label}"`},
		{name: "日本語", locale: "ja", want: `data-season-slugs-remove-label="{label}の選択を解除"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), tt.locale)
			var buf strings.Builder
			if err := seasonSlugsFilter(nil).Render(ctx, &buf); err != nil {
				t.Fatalf("放送シーズンの絞り込みの描画エラー = %v", err)
			}

			if html := buf.String(); !strings.Contains(html, tt.want) {
				t.Errorf("翻訳済みの削除ラベルのテンプレートが無い。期待する文字列 = %q、HTML = %q", tt.want, html)
			}
		})
	}
}
