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

			// 空表示はページで唯一の <h2> で何が無いのかを述べるため、一覧が無い状態でもアウト
			// ラインはh1 → h2のままになる。
			if !strings.Contains(html, `<section class="empty">`) {
				t.Error("エピソードが空の場合は空表示コンポーネントが表示されるべきです")
			}

			if want := "<h2>" + tt.heading + "</h2>"; !strings.Contains(html, want) {
				t.Errorf("エピソードが空の場合は空表示の見出し%qが表示されるべきです", want)
			}

			// 一覧が空でもサブナビは描画され、閲覧者は作品へ戻れる。
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
		// 見出しは親作品を名指しし、サブナビはそこへ戻るリンクを持つ。
		"テストアニメ",
		`href="/db/works/3/edit"`,
		// ID列はエピソードの公開ページを新しいタブで開くリンクで、そのことを
		// アクセシブルネームで伝える。
		`href="/works/3/episodes/10"`,
		`target="_blank"`,
		`rel="noopener"`,
		`aria-label="エピソード 10 を新しいタブで開く"`,
		// 2系統の話数と2つのタイトル。
		"第2話",
		"エピソードタイトル",
		"Episode Title",
		// 縦積みした各値にはローカライズ済みのラベルを付け、タイトルには既知の言語を
		// 指定して、支援技術が値を区別して正しく発音できるようにする。
		"表示用話数:",
		// 「表示用話数:」の部分一致にならないよう、話数のラベルは要素の先頭から照合する。
		`<span class="sr-only">話数: </span>`,
		"日本語タイトル:",
		"英語タイトル:",
		`lang="ja">エピソードタイトル</span>`,
		`lang="en">Episode Title</span>`,
		// 前のエピソードの列は、クエリが導出した隣接エピソードを名指しする。
		"前のエピソード",
		"第1話",
		// ソート番号と記録数の列。
		`<th scope="col" class="text-left">ソート番号</th>`,
		"<td>200</td>",
		"<td>100</td>",
		"<td>42</td>",
		// 状態のバッジは共有のステータスラベルコンポーネントが描画する。
		`<span class="badge" data-variant="success">公開</span>`,
		`<span class="badge" data-variant="warning">非公開</span>`,
	}

	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("出力に%qが含まれていません", expected)
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

	// 未設定の4つの属性 (表示用話数・話数・日本語タイトル・英語タイトル) は
	// いずれも空のセルではなくプレースホルダーを描画する。
	if got := strings.Count(html, ": </span> -</div>"); got != 4 {
		t.Errorf("プレースホルダーの数 = %d、期待値 = 4", got)
	}
	if strings.Contains(html, `lang="ja">-`) || strings.Contains(html, `lang="en">-`) {
		t.Error("プレースホルダーには言語指定を付けてはいけません")
	}

	// 作品の最初のエピソードには直前のエピソードが無く、その列も同じ欠落として読める。
	if !strings.Contains(html, `<td class="whitespace-normal [overflow-wrap:anywhere]">-</td>`) {
		t.Error("直前のエピソードが無い行はプレースホルダーを表示すべきです")
	}
}

// TestIndex_GenerationNoticeは、対応する両ロケールでエピソード計画の案内を検証する。
// 編集者がエピソードを計画するための3つの値を述べ、一覧が空でも描画される (空の一覧では
// これらの値だけがページの情報になるため)。案内がBasecoatのalertではなく素のコンテナに
// 載るのは、alertがタイトル要素を必須とする一方、この案内は意図的に見出しを持たないため。
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
			wantLabels: []string{"予定エピソード数", "公開中のエピソード数", "自動生成されるエピソード数"},
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
				t.Error("案内はBasecoatのalertを使うべきではありません (タイトル要素が必須のため)")
			}
			if !strings.Contains(html, "<dl") {
				t.Error("案内は定義リストで描画されるべきです")
			}
			for _, value := range []string{"12", "5", "9"} {
				if !strings.Contains(html, `<dd class="text-card-foreground">`+value+"</dd>") {
					t.Errorf("出力に値%qが含まれていません", value)
				}
			}
			for _, label := range tt.wantLabels {
				if !strings.Contains(html, "<dt>"+label+"</dt>") {
					t.Errorf("出力にラベル%qが含まれていません", label)
				}
			}
		})
	}
}

// TestIndex_GenerationNoticeUnknownPlannedCountは、予定エピソード数が未登録の作品でその旨を
// 言葉で示すことを検証する。案内は3つの値を並べて述べるため、欠落がそれ自体で件数のように
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
		t.Error("予定エピソード数が未登録なら「不明」と表示すべきです")
	}
}

// TestIndex_HeadingFallsBackWhenWorkNameEmptyは、表示できる名前が無い作品に対して
// viewmodel.DBEpisodeListWorkNameが返す空のWorkNameを検証する。文書タイトルも同じ合図で
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
		t.Error("no_episodesの作品ではサブナビのエピソード項目が落ちるべきです")
	}
}

// actionColumnEpisodesは操作列テスト用に、公開中と非公開のエピソードを1件ずつ返す。
// 状態で分かれる操作の両方が1回の描画に現れるようにするため。
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

	// テーブル構造の3部分をまとめて数える。条件付きの列はcol・見出し・各行のデータ
	// セルを一体として増減させ、支援技術にも視覚的な配置と同じ関係を伝える必要がある。
	if got := strings.Count(html, "<col ") + strings.Count(html, "<col>"); got != wantColumns {
		t.Errorf("列要素数 = %d、期待値 = %d", got, wantColumns)
	}
	if got := strings.Count(html, "<th "); got != wantColumns {
		t.Errorf("列見出し数 = %d、期待値 = %d", got, wantColumns)
	}
	if got, want := strings.Count(html, "<td"), wantColumns*len(actionColumnEpisodes()); got != want {
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

	html := renderActionColumnIndex(t, true, false)
	assertActionColumnStructure(t, html, true)

	wantPresent := []string{
		// 列は専用の見出しで示される。
		`<th scope="col" class="text-center">操作</th>`,
		// 両行の編集リンク。
		`href="/db/episodes/10/edit"`,
		`href="/db/episodes/11/edit"`,
		// 公開中の行 (10): 確認画面への非公開リンク。
		`href="/db/episodes/10/archive/new"`,
		// 非公開の行 (11): 公開はarchiveパスへのhtmx DELETEで、確認ダイアログと
		// X-CSRF-Tokenヘッダーで送るCSRFトークンを伴う。
		`hx-delete="/db/episodes/11/archive"`,
		"エピソード 11 を公開しますか",
		"X-CSRF-Token",
		"test-csrf-token",
		// 各コントロールはアクセシブルネームに対象のエピソードを足し、行が並ぶページで
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
		// 公開中の行に公開ボタンは無く、非公開の行に非公開リンクは無い。
		`hx-delete="/db/episodes/10/archive"`,
		`href="/db/episodes/11/archive/new"`,
		// どちらの行にもadmin専用の削除ボタンは無い。
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

// TestIndex_ActionColumn_Adminはadminがさらに各行に削除のhtmx DELETEボタン
// (エピソードのパスへのDELETE) を見ることを検証する。
func TestIndex_ActionColumn_Admin(t *testing.T) {
	t.Parallel()

	html := renderActionColumnIndex(t, true, true)

	wantPresent := []string{
		// 削除ボタンはエピソードのパス (archiveパスとは別) をDELETE対象にする。
		`hx-delete="/db/episodes/10"`,
		`hx-delete="/db/episodes/11"`,
		"エピソード 10 を削除しますか",
		"エピソード 11 を削除しますか",
		`削除<span class="sr-only"> エピソード 10</span>`,
		// committerの操作も引き続き表示される。
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

// TestIndex_ActionColumn_Anonymousは未ログインや一般ユーザーに操作列そのものが
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

// TestIndex_DecorativeIconsAreHiddenは作成リンクと各行のIDリンクのアイコンを検証する。
// どちらも囲むリンクが既に伝えている内容を繰り返す。SVGはアクセシビリティツリーとフォーカス
// 順序から除外し、ブラウザー依存の別表現を重ねないようにする。作成リンクはBasecoatのテキスト
// 付きボタンのため、そのアイコンはBasecoatがボタンのアイコン用間隔を適用できるinline-start
// の位置も宣言する。IDリンクは通常のリンクのため位置を持たない。
func TestIndex_DecorativeIconsAreHidden(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	var linkBuf strings.Builder
	if err := indexNewEpisodesLink(IndexPageData{WorkID: 1}).Render(ctx, &linkBuf); err != nil {
		t.Fatalf("一覧アクションのレンダリングエラー: %v", err)
	}

	if !strings.Contains(linkBuf.String(), decorativeIconMarkup(t, ctx, "plus-regular", "", templates.InlineIconStart)) {
		t.Error(`装飾アイコン "plus-regular" がaria-hiddenの要素内にありません`)
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
		t.Error(`装飾アイコン "arrow-square-out-regular" がaria-hiddenかつfocusable="false" ではありません`)
	}
	if strings.Contains(html, iconWrapperMarkup) {
		t.Error("装飾アイコンはラッパー要素ではなくSVG自体で隠すべきです")
	}
}
