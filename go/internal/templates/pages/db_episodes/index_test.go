package db_episodes

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/templates"
	"github.com/annict/annict/go/internal/viewmodel"
)

func TestIndex_Empty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		locale  string
		heading string
	}{
		{name: "日本語", locale: "ja", heading: "エピソードはありません"},
		{name: "英語", locale: "en", heading: "No episodes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), tt.locale)
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

			// The empty state names the missing content in the page's only <h2>, so the outline stays
			// h1 -> h2 with the list gone.
			//
			// [Ja] 空表示はページで唯一の <h2> で何が無いのかを述べるため、一覧が無い状態でもアウト
			// ラインは h1 → h2 のままになる。
			if !strings.Contains(html, `<section class="empty">`) {
				t.Error("エピソードが空の場合は空表示コンポーネントが表示されるべきです")
			}

			if want := "<h2>" + tt.heading + "</h2>"; !strings.Contains(html, want) {
				t.Errorf("エピソードが空の場合は空表示の見出し %q が表示されるべきです", want)
			}

			// The subnav still renders on an empty list, so the visitor can move back to the work.
			//
			// [Ja] 一覧が空でもサブナビは描画され、閲覧者は作品へ戻れる。
			if !strings.Contains(html, `href="/db/works/1/edit"`) {
				t.Error("エピソードが空でも作品へ戻るサブナビが表示されるべきです")
			}
		})
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
				ID:                  10,
				WorkID:              3,
				Number:              "第2話",
				RawNumber:           "2",
				Title:               "エピソードタイトル",
				TitleEn:             "Episode Title",
				PrevNumber:          "第1話",
				SortNumber:          200,
				EpisodeRecordsCount: 42,
				Status:              viewmodel.PublishingStatusPublished,
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
		// The preceding episode column names the neighbour the query derived.
		//
		// [Ja] 前のエピソードの列は、クエリが導出した隣接エピソードを名指しする。
		"前のエピソード",
		"第1話",
		// The sort number and the records count columns.
		//
		// [Ja] 並び順と記録数の列。
		"<td>200</td>",
		"<td>100</td>",
		"<td>42</td>",
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

	// The work's first episode has no preceding one, and its column reads as the same gap.
	//
	// [Ja] 作品の最初のエピソードには直前のエピソードが無く、その列も同じ欠落として読める。
	if !strings.Contains(html, `<td class="whitespace-normal [overflow-wrap:anywhere]">-</td>`) {
		t.Error("直前のエピソードが無い行はプレースホルダーを表示すべきです")
	}
}

// TestIndex_GenerationNotice covers the episode-planning notice in both supported locales.
// It states the three values the editor plans episodes by, and it renders on an empty list
// too, where those values are all the page has to show. The notice sits in a plain container
// rather than a Basecoat alert because that component requires a title element, and the
// notice deliberately carries no heading.
//
// [Ja] TestIndex_GenerationNotice は、対応する両ロケールでエピソード計画の案内を検証する。
// 編集者がエピソードを計画するための 3 つの値を述べ、一覧が空でも描画される (空の一覧では
// これらの値だけがページの情報になるため)。案内が Basecoat の alert ではなく素のコンテナに
// 載るのは、alert がタイトル要素を必須とする一方、この案内は意図的に見出しを持たないため。
func TestIndex_GenerationNotice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		locale     string
		wantLabels []string
	}{
		{
			name:       "日本語",
			locale:     "ja",
			wantLabels: []string{"予定総話数", "公開中のエピソード数", "自動生成されるエピソード数"},
		},
		{
			name:       "英語",
			locale:     "en",
			wantLabels: []string{"Expected episodes", "Published episodes", "Auto-generated episodes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), tt.locale)
			data := IndexPageData{
				WorkID:   1,
				WorkName: "テストアニメ",
				Generation: viewmodel.DBEpisodeGenerationSummary{
					PlannedCount:                "12",
					PublishedEpisodeCount:       5,
					MaxGeneratableEpisodeNumber: 9,
				},
				Episodes:   []viewmodel.DBEpisodeListItem{},
				Pagination: viewmodel.NewPagination(1, 0, 100, "/db/works/1/episodes"),
			}

			var buf strings.Builder
			if err := Index(data).Render(ctx, &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			html := buf.String()
			if strings.Contains(html, `class="alert"`) {
				t.Error("案内は Basecoat の alert を使うべきではありません (タイトル要素が必須のため)")
			}
			if !strings.Contains(html, "<dl") {
				t.Error("案内は定義リストで描画されるべきです")
			}
			for _, value := range []string{"12", "5", "9"} {
				if !strings.Contains(html, `<dd class="text-card-foreground">`+value+"</dd>") {
					t.Errorf("出力に値 %q が含まれていません", value)
				}
			}
			for _, label := range tt.wantLabels {
				if !strings.Contains(html, "<dt>"+label+"</dt>") {
					t.Errorf("出力にラベル %q が含まれていません", label)
				}
			}
		})
	}
}

// TestIndex_GenerationNoticeUnknownPlannedCount verifies that a work with no expected
// episode count recorded says so in words. The notice states three values side by side, so
// the gap must not read as a count of its own.
//
// [Ja] TestIndex_GenerationNoticeUnknownPlannedCount は、予定総話数が未登録の作品でその旨を
// 言葉で示すことを検証する。案内は 3 つの値を並べて述べるため、欠落がそれ自体で件数のように
// 読めてはならない。
func TestIndex_GenerationNoticeUnknownPlannedCount(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	data := IndexPageData{
		WorkID:   1,
		WorkName: "テストアニメ",
		Generation: viewmodel.DBEpisodeGenerationSummary{
			PlannedCount:                "",
			PublishedEpisodeCount:       0,
			MaxGeneratableEpisodeNumber: 0,
		},
		Episodes:   []viewmodel.DBEpisodeListItem{},
		Pagination: viewmodel.NewPagination(1, 0, 100, "/db/works/1/episodes"),
	}

	var buf strings.Builder
	if err := Index(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	if !strings.Contains(buf.String(), `<dd class="text-card-foreground">不明</dd>`) {
		t.Error("予定総話数が未登録なら「不明」と表示すべきです")
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

// actionColumnEpisodes returns one published and one archived episode for the action-column
// tests, so both branches of the status-dependent action appear in one render.
//
// [Ja] actionColumnEpisodes は操作列テスト用に、公開中と非公開のエピソードを 1 件ずつ返す。
// 状態で分かれる操作の両方が 1 回の描画に現れるようにするため。
func actionColumnEpisodes() []viewmodel.DBEpisodeListItem {
	return []viewmodel.DBEpisodeListItem{
		{ID: 10, WorkID: 3, Number: "第2話", SortNumber: 200, Status: viewmodel.PublishingStatusPublished},
		{ID: 11, WorkID: 3, Number: "第1話", SortNumber: 100, Status: viewmodel.PublishingStatusArchived},
	}
}

func renderActionColumnIndex(t *testing.T, isCommitter, isAdmin bool) string {
	t.Helper()

	ctx := i18n.SetLocale(context.Background(), "ja")
	data := IndexPageData{
		WorkID:      3,
		WorkName:    "テストアニメ",
		Episodes:    actionColumnEpisodes(),
		Pagination:  viewmodel.NewPagination(1, 2, 100, "/db/works/3/episodes"),
		IsCommitter: isCommitter,
		IsAdmin:     isAdmin,
		CSRFToken:   "test-csrf-token",
	}

	var buf strings.Builder
	if err := Index(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	return buf.String()
}

func assertActionColumnStructure(t *testing.T, html string, wantActions bool) {
	t.Helper()

	wantColumns := 7
	wantMinWidth := "min-w-[860px]"
	wantAbsentMinWidth := "min-w-[960px]"
	if wantActions {
		wantColumns = 8
		wantMinWidth = "min-w-[960px]"
		wantAbsentMinWidth = "min-w-[860px]"
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
	if got, want := strings.Count(html, "<td"), wantColumns*len(actionColumnEpisodes()); got != want {
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
// every row, the unpublish link (to the confirmation screen) on published rows and the publish
// htmx DELETE button on archived rows, but not the admin-only delete button.
//
// [Ja] TestIndex_ActionColumn_Committer は committer (非 admin) が各行に編集リンクを、公開中の
// 行に非公開リンク (確認画面へ) を、非公開の行に公開の htmx DELETE ボタンを見る一方、admin
// 専用の削除ボタンは見えないことを検証する。
func TestIndex_ActionColumn_Committer(t *testing.T) {
	t.Parallel()

	html := renderActionColumnIndex(t, true, false)
	assertActionColumnStructure(t, html, true)

	wantPresent := []string{
		// The column is announced by its own header.
		//
		// [Ja] 列は専用の見出しで示される。
		`<th scope="col" class="text-center">操作</th>`,
		// Edit link on both rows.
		//
		// [Ja] 両行の編集リンク。
		`href="/db/episodes/10/edit"`,
		`href="/db/episodes/11/edit"`,
		// Published row (10): unpublish link to the confirmation screen.
		//
		// [Ja] 公開中の行 (10): 確認画面への非公開リンク。
		`href="/db/episodes/10/archive/new"`,
		// Archived row (11): publish is an htmx DELETE against the archive path, with a confirm
		// dialog and the CSRF token carried in the X-CSRF-Token header.
		//
		// [Ja] 非公開の行 (11): 公開は archive パスへの htmx DELETE で、確認ダイアログと
		// X-CSRF-Token ヘッダーで送る CSRF トークンを伴う。
		`hx-delete="/db/episodes/11/archive"`,
		"エピソード 11 を公開しますか",
		"X-CSRF-Token",
		"test-csrf-token",
		// Every control extends its accessible name with the episode it acts on, so a page of
		// rows does not present controls that all read the same. The visible label stays at
		// the start of the name.
		//
		// [Ja] 各コントロールはアクセシブルネームに対象のエピソードを足し、行が並ぶページで
		// 同じ名前のコントロールばかりにならないようにする。可視ラベルは名前の先頭に残る。
		`編集<span class="sr-only"> エピソード 10</span>`,
		`編集<span class="sr-only"> エピソード 11</span>`,
		`非公開<span class="sr-only"> エピソード 10</span>`,
		`公開<span class="sr-only"> エピソード 11</span>`,
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
		`hx-delete="/db/episodes/10/archive"`,
		`href="/db/episodes/11/archive/new"`,
		// No admin-only delete buttons for either row.
		//
		// [Ja] どちらの行にも admin 専用の削除ボタンは無い。
		`hx-delete="/db/episodes/10"`,
		`hx-delete="/db/episodes/11"`,
		"を削除しますか",
	}
	for _, unexpected := range wantAbsent {
		if strings.Contains(html, unexpected) {
			t.Errorf("含まれてはいけない文字列が含まれています: %q", unexpected)
		}
	}
}

// TestIndex_ActionColumn_Admin verifies that an admin additionally sees the delete htmx DELETE
// button (against the episode path) on every row.
//
// [Ja] TestIndex_ActionColumn_Admin は admin がさらに各行に削除の htmx DELETE ボタン
// (エピソードのパスへの DELETE) を見ることを検証する。
func TestIndex_ActionColumn_Admin(t *testing.T) {
	t.Parallel()

	html := renderActionColumnIndex(t, true, true)

	wantPresent := []string{
		// The delete button targets the episode path (distinct from the archive path).
		//
		// [Ja] 削除ボタンはエピソードのパス (archive パスとは別) を DELETE 対象にする。
		`hx-delete="/db/episodes/10"`,
		`hx-delete="/db/episodes/11"`,
		"エピソード 10 を削除しますか",
		"エピソード 11 を削除しますか",
		`削除<span class="sr-only"> エピソード 10</span>`,
		// The committer actions are still present.
		//
		// [Ja] committer の操作も引き続き表示される。
		`href="/db/episodes/10/archive/new"`,
		`hx-delete="/db/episodes/11/archive"`,
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

	html := renderActionColumnIndex(t, false, false)
	assertActionColumnStructure(t, html, false)

	wantAbsent := []string{
		`<th scope="col" class="text-center">操作</th>`,
		`href="/db/episodes/10/edit"`,
		`href="/db/episodes/10/archive/new"`,
		"hx-delete=",
	}
	for _, unexpected := range wantAbsent {
		if strings.Contains(html, unexpected) {
			t.Errorf("含まれてはいけない文字列が含まれています: %q", unexpected)
		}
	}
}

// TestIndex_DecorativeIconsAreHidden covers the icons of the create link and of each row's ID
// link, both of which repeat what the link around them already announces. The SVGs stay out of
// the accessibility tree and out of the focus order instead of adding a second,
// browser-dependent representation. The create link is a Basecoat text button, so its icon also
// declares the inline-start position that lets Basecoat apply the button's icon-aware spacing;
// the ID link is a plain link and takes no position.
//
// [Ja] TestIndex_DecorativeIconsAreHidden は作成リンクと各行の ID リンクのアイコンを検証する。
// どちらも囲むリンクが既に伝えている内容を繰り返す。SVG はアクセシビリティツリーとフォーカス
// 順序から除外し、ブラウザー依存の別表現を重ねないようにする。作成リンクは Basecoat のテキスト
// 付きボタンのため、そのアイコンは Basecoat がボタンのアイコン用間隔を適用できる inline-start
// の位置も宣言する。ID リンクは通常のリンクのため位置を持たない。
func TestIndex_DecorativeIconsAreHidden(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	var linkBuf strings.Builder
	if err := indexNewEpisodesLink(IndexPageData{WorkID: 1}).Render(ctx, &linkBuf); err != nil {
		t.Fatalf("一覧アクションのレンダリングエラー: %v", err)
	}

	if !strings.Contains(linkBuf.String(), decorativeIconMarkup(t, ctx, "plus-regular", "", templates.InlineIconStart)) {
		t.Error(`装飾アイコン "plus-regular" が aria-hidden の要素内にありません`)
	}

	var pageBuf strings.Builder
	if err := Index(IndexPageData{
		WorkID:     3,
		WorkName:   "テストアニメ",
		Episodes:   []viewmodel.DBEpisodeListItem{{ID: 10, WorkID: 3, Status: viewmodel.PublishingStatusPublished}},
		Pagination: viewmodel.NewPagination(1, 1, 100, "/db/works/3/episodes"),
	}).Render(ctx, &pageBuf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	html := pageBuf.String()

	if !strings.Contains(html, decorativeIconMarkup(t, ctx, "arrow-square-out-regular", "w-[18px] h-[18px]")) {
		t.Error(`装飾アイコン "arrow-square-out-regular" が aria-hidden かつ focusable="false" ではありません`)
	}
	if strings.Contains(html, iconWrapperMarkup) {
		t.Error("装飾アイコンはラッパー要素ではなく SVG 自体で隠すべきです")
	}
}
