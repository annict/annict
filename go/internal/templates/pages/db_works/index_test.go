package db_works

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/viewmodel"
)

// TestIndex_Empty は作品が存在しない場合に表が表示されず空メッセージが表示されることをテスト
func TestIndex_Empty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
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

	// 空メッセージが表示されることを確認
	if !strings.Contains(html, "作品が見つかりませんでした") {
		t.Error("作品が空の場合は空メッセージが表示されるべきです")
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
				Season:        "2024 春",
				Syobocal:      viewmodel.ExternalServiceLink{Label: "3524", URL: "http://cal.syoboi.jp/tid/3524"},
				MalAnime:      viewmodel.ExternalServiceLink{Label: "20", URL: "https://myanimelist.net/anime/20"},
				WatchersCount: 100,
				Status:        viewmodel.WorkStatusPublished,
				ImageURL:      "https://imgproxy.test/thumb1.jpg",
			},
			{
				ID:            2,
				Title:         "テストアニメ2",
				TitleKana:     "",
				TitleEn:       "",
				Media:         "OVA",
				Season:        "2024 夏",
				WatchersCount: 50,
				Status:        viewmodel.WorkStatusPublished,
				ImageURL:      "",
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
		// The media column and the extra kana / English titles.
		//
		// [Ja] メディア列と追加のふりがな・英語タイトル。
		"てすとあにめいち",
		"Test Anime 1",
		"TV",
		"OVA",
		// A work with an image renders a thumbnail <picture> with the imgproxy URL and lazy loading.
		//
		// [Ja] 画像がある作品は imgproxy URL と遅延読み込み付きのサムネイル <picture> を描画する。
		"<picture",
		`src="https://imgproxy.test/thumb1.jpg"`,
		`alt="テストアニメ1"`,
		`width="70"`,
		`height="52"`,
		`loading="lazy"`,
		// A work without an image renders a muted placeholder box.
		//
		// [Ja] 画像がない作品はミュートされたプレースホルダーの箱を描画する。
		"bg-muted",
		// The external services column shows the しょぼかる / MyAnimeList links for a
		// work that has ids, each opening in a new tab with rel="noopener".
		//
		// [Ja] 外部サービス列は、ID を持つ作品のしょぼかる / MyAnimeList リンクを表示し、
		// それぞれ rel="noopener" 付きで新しいタブで開く。
		"外部サービス",
		"しょぼかる",
		"MyAnimeList",
		`href="http://cal.syoboi.jp/tid/3524"`,
		`href="https://myanimelist.net/anime/20"`,
		">3524</a>",
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
}

// actionColumnWorks returns one published and one archived work for the action-column tests.
//
// [Ja] actionColumnWorks は操作列テスト用に、公開中と非公開の作品を 1 件ずつ返す。
func actionColumnWorks() []viewmodel.DBWorkListItem {
	return []viewmodel.DBWorkListItem{
		{ID: 1, Title: "公開作品", Media: "TV", Status: viewmodel.WorkStatusPublished},
		{ID: 2, Title: "非公開作品", Media: "TV", Status: viewmodel.WorkStatusArchived},
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

	wantPresent := []string{
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
		// dialog and the CSRF token carried in the X-CSRF-Token header.
		//
		// [Ja] 非公開の行 (2): 公開は archive パスへの htmx DELETE で、確認ダイアログと
		// X-CSRF-Token ヘッダーで送る CSRF トークンを伴う。
		`hx-delete="/db/works/2/archive"`,
		"この作品を公開しますか",
		"X-CSRF-Token",
		"test-csrf-token",
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
		"この作品を削除しますか",
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
		"この作品を削除しますか",
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

// TestIndex_ActionColumn_Anonymous verifies that a signed-out or regular visitor sees no action
// controls (the list itself stays public).
//
// [Ja] TestIndex_ActionColumn_Anonymous は未ログインや一般ユーザーが操作コントロールを一切
// 見ない (一覧自体は公開のまま) ことを検証する。
func TestIndex_ActionColumn_Anonymous(t *testing.T) {
	t.Parallel()

	html := renderIndex(t, IndexPageData{
		Works:       actionColumnWorks(),
		Pagination:  viewmodel.NewPagination(1, 2, 30, "/db/works"),
		IsCommitter: false,
		IsAdmin:     false,
	})

	wantAbsent := []string{
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

// TestIndex_FilterUI はリリース時期の複数選択と放送予定未登録チェックボックスの描画をテスト
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
		// The release-season multi-select and its selected / unselected options.
		//
		// [Ja] リリース時期の複数選択とその選択済み / 未選択オプション。
		`name="season_slugs"`,
		"multiple",
		`<option value="2024-spring" selected>2024年春</option>`,
		`<option value="2024-winter">2024年冬</option>`,
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
}
