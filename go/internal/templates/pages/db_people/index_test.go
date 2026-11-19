package db_people

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/viewmodel"
)

// renderIndexは人物一覧ページを描画し、生成されたHTMLを返す。
func renderIndex(t *testing.T, ctx context.Context, data IndexPageData) string {
	t.Helper()

	var buf strings.Builder
	if err := Index(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	return buf.String()
}

// TestIndex_Emptyは、人物が無いときに表が描画されず、その位置に空表示が出ることを、
// サイトが提供する各ロケールで検証する。
func TestIndex_Empty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		locale  string
		heading string
	}{
		{name: "日本語", locale: "ja", heading: "人物が見つかりませんでした"},
		{name: "英語", locale: "en", heading: "No people found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			html := renderIndex(t, i18n.SetLocale(context.Background(), tt.locale), IndexPageData{
				People:     []viewmodel.DBPersonListItem{},
				Pagination: viewmodel.NewPagination(1, 0, 100, "/db/people"),
			})

			if strings.Contains(html, "<table") {
				t.Error("人物が空の場合は <table> が含まれてはいけません")
			}
			if want := "<h2>" + tt.heading + "</h2>"; !strings.Contains(html, want) {
				t.Errorf("人物が空の場合は空表示の見出し%qが表示されるべきです", want)
			}
		})
	}
}

// TestIndex_WithPeopleは、人物の行がID・名前・読み・英語名・状態とともに、
// captionとscopeを持つ表として描画されることを検証する。
func TestIndex_WithPeople(t *testing.T) {
	t.Parallel()

	html := renderIndex(t, i18n.SetLocale(context.Background(), "ja"), IndexPageData{
		People: []viewmodel.DBPersonListItem{
			{ID: 2, Name: "人物2", NameKana: "じんぶつに", NameEn: "Person 2", Status: viewmodel.DBPersonStatusPublished},
			{ID: 1, Name: "人物1", Status: viewmodel.DBPersonStatusArchived},
		},
		Pagination: viewmodel.NewPagination(1, 2, 100, "/db/people"),
	})

	for _, expected := range []string{
		`<caption class="sr-only">DB人物一覧</caption>`,
		`<th scope="col" class="text-left">ID</th>`,
		`<th scope="col" class="text-left">名前</th>`,
		`<th scope="col" class="text-left">ステータス</th>`,
		// ID列は人物の公開ページを新しいタブで開くリンクになる。
		`href="/people/2"`,
		`aria-label="人物 2 を新しいタブで開く"`,
		"人物2",
		"じんぶつに",
		"Person 2",
		"人物1",
		// 読み・英語名の無い人物はプレースホルダーを表示する。
		`<div class="text-muted-foreground">-</div>`,
		`<span class="badge" data-variant="success">公開</span>`,
		`<span class="badge" data-variant="warning">非公開</span>`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}
}

// TestIndex_ActionsByRoleは、committerにだけ「新規登録」と各行の「編集」、操作列を
// 描画することを検証する。
func TestIndex_ActionsByRole(t *testing.T) {
	t.Parallel()

	newLink := `href="/db/people/new"`
	editLink := `href="/db/people/1/edit"`
	actionsHeader := `<th scope="col" class="text-center">メニュー</th>`

	tests := []struct {
		name        string
		isCommitter bool
		wantPresent bool
	}{
		{name: "committer", isCommitter: true, wantPresent: true},
		{name: "committer以外", isCommitter: false, wantPresent: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			html := renderIndex(t, i18n.SetLocale(context.Background(), "ja"), IndexPageData{
				People: []viewmodel.DBPersonListItem{
					{ID: 1, Name: "人物1", Status: viewmodel.DBPersonStatusPublished},
				},
				Pagination:  viewmodel.NewPagination(1, 1, 100, "/db/people"),
				IsCommitter: tt.isCommitter,
			})

			for _, s := range []string{newLink, editLink, actionsHeader} {
				if got := strings.Contains(html, s); got != tt.wantPresent {
					t.Errorf("%qを含むか = %v、期待値 = %v", s, got, tt.wantPresent)
				}
			}
			if tt.wantPresent {
				// 編集リンクは視覚的に隠したテキストで対象の人物を名指しする。
				if want := `<span class="sr-only"> 人物 1</span>`; !strings.Contains(html, want) {
					t.Errorf("編集リンクに対象の人物の名指し%qが含まれていません", want)
				}
			}
		})
	}
}

// TestIndex_StateActionsは、操作列が行の状態とロールに応じて、公開中の行に非公開の確認画面
// へのリンクを、非公開の行に公開のhtmxボタンを、管理者にだけ削除のhtmxボタンを
// 描画することを検証する。公開・削除のボタンは確認ダイアログとCSRFヘッダーを持ち、
// 確認文で人物を名指しする。
func TestIndex_StateActions(t *testing.T) {
	t.Parallel()

	archiveLink := `href="/db/people/1/archive/new"`
	unarchiveButton := `hx-delete="/db/people/2/archive"`
	unarchiveConfirm := `hx-confirm="人物 2 を公開しますか？"`
	deleteButtons := []string{`hx-delete="/db/people/1"`, `hx-delete="/db/people/2"`}
	deleteConfirm := `hx-confirm="人物 1 を削除しますか？"`
	csrfHeader := `hx-headers="{&#34;X-CSRF-Token&#34;:&#34;test-csrf&#34;}"`

	tests := []struct {
		name       string
		isAdmin    bool
		wantDelete bool
	}{
		{name: "編集者", isAdmin: false, wantDelete: false},
		{name: "管理者", isAdmin: true, wantDelete: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			html := renderIndex(t, i18n.SetLocale(context.Background(), "ja"), IndexPageData{
				People: []viewmodel.DBPersonListItem{
					{ID: 1, Name: "公開中の人物", Status: viewmodel.DBPersonStatusPublished},
					{ID: 2, Name: "非公開の人物", Status: viewmodel.DBPersonStatusArchived},
				},
				Pagination:  viewmodel.NewPagination(1, 2, 100, "/db/people"),
				IsCommitter: true,
				IsAdmin:     tt.isAdmin,
				CSRFToken:   "test-csrf",
			})

			for _, s := range []string{archiveLink, unarchiveButton, unarchiveConfirm, csrfHeader} {
				if !strings.Contains(html, s) {
					t.Errorf("レスポンスに含まれていない文字列 = %q", s)
				}
			}
			// 非公開の行に非公開の操作を、公開中の行に公開の操作を出さない。
			for _, s := range []string{`href="/db/people/2/archive/new"`, `hx-delete="/db/people/1/archive"`} {
				if strings.Contains(html, s) {
					t.Errorf("行の状態に合わない操作%qが含まれています", s)
				}
			}
			for _, s := range append(deleteButtons, deleteConfirm) {
				if got := strings.Contains(html, s); got != tt.wantDelete {
					t.Errorf("%qを含むか = %v、期待値 = %v", s, got, tt.wantDelete)
				}
			}
		})
	}
}

// TestIndex_DelistedRowは、管理者に表示する掲載停止の行が、掲載停止のバッジと掲載停止の解除の
// 確認画面へのリンクだけを持ち、編集・アーカイブ・掲載停止・削除の操作を持たないことを検証する。
// 閲覧できる状態の行には、管理者にだけ掲載停止の確認画面へのリンクを出す。
func TestIndex_DelistedRow(t *testing.T) {
	t.Parallel()

	html := renderIndex(t, i18n.SetLocale(context.Background(), "ja"), IndexPageData{
		People: []viewmodel.DBPersonListItem{
			{ID: 1, Name: "公開中の人物", Status: viewmodel.DBPersonStatusPublished},
			{ID: 3, Name: "掲載停止の人物", Status: viewmodel.DBPersonStatusDelisted},
		},
		Pagination:  viewmodel.NewPagination(1, 2, 100, "/db/people"),
		IsCommitter: true,
		IsAdmin:     true,
		CSRFToken:   "test-csrf",
	})

	for _, s := range []string{
		`<span class="badge" data-variant="destructive">掲載停止</span>`,
		`href="/people/1"`,
		`href="/db/people/3/undelisting/new"`,
		`href="/db/people/1/delisting/new"`,
	} {
		if !strings.Contains(html, s) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", s)
		}
	}
	for _, s := range []string{
		`href="/people/3"`,
		`href="/db/people/3/edit"`,
		`href="/db/people/3/archive/new"`,
		`href="/db/people/3/delisting/new"`,
		`hx-delete="/db/people/3"`,
		`href="/db/people/1/undelisting/new"`,
	} {
		if strings.Contains(html, s) {
			t.Errorf("行の状態に合わない操作%qが含まれています", s)
		}
	}
}

// TestIndex_DelistLinkRequiresAdminは、編集者には掲載停止の確認画面へのリンクを出さないことを
// 検証する。
func TestIndex_DelistLinkRequiresAdmin(t *testing.T) {
	t.Parallel()

	html := renderIndex(t, i18n.SetLocale(context.Background(), "ja"), IndexPageData{
		People: []viewmodel.DBPersonListItem{
			{ID: 1, Name: "公開中の人物", Status: viewmodel.DBPersonStatusPublished},
		},
		Pagination:  viewmodel.NewPagination(1, 1, 100, "/db/people"),
		IsCommitter: true,
		CSRFToken:   "test-csrf",
	})

	if strings.Contains(html, `href="/db/people/1/delisting/new"`) {
		t.Error("編集者に掲載停止の確認画面へのリンクが表示されています")
	}
}

// TestIndex_RevisionsLinkは、フィーチャーフラグdb_edit_historyが有効な閲覧者には、ログインしていなくても
// 各行に「履歴」とメニューの列を出し、committer以外には操作を出さないことを検証する。
func TestIndex_RevisionsLink(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name          string
		isCommitter   bool
		showRevisions bool
		wantRevisions bool
		wantEdit      bool
		wantColumn    bool
		wantMinWidth  string
	}{
		{name: "ログインしていない閲覧者でフラグが有効", showRevisions: true, wantRevisions: true, wantColumn: true, wantMinWidth: "min-w-[480px]"},
		{name: "ログインしていない閲覧者でフラグが無効", wantMinWidth: "min-w-[384px]"},
		{name: "committerでフラグが有効", isCommitter: true, showRevisions: true, wantRevisions: true, wantEdit: true, wantColumn: true, wantMinWidth: "min-w-[480px]"},
		{name: "committerでフラグが無効", isCommitter: true, wantEdit: true, wantColumn: true, wantMinWidth: "min-w-[480px]"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			html := renderIndex(t, i18n.SetLocale(context.Background(), "ja"), IndexPageData{
				People: []viewmodel.DBPersonListItem{
					{ID: 1, Name: "人物1", Status: viewmodel.DBPersonStatusPublished},
					{ID: 2, Name: "人物2", Status: viewmodel.DBPersonStatusArchived},
				},
				Pagination:    viewmodel.NewPagination(1, 2, 100, "/db/people"),
				IsCommitter:   tt.isCommitter,
				ShowRevisions: tt.showRevisions,
			})

			for _, c := range []struct {
				s    string
				want bool
			}{
				{`href="/db/people/1/revisions"`, tt.wantRevisions},
				{`href="/db/people/2/revisions"`, tt.wantRevisions},
				{`href="/db/people/1/edit"`, tt.wantEdit},
				// 非公開の人物の再公開のボタンも、committerにだけ出す。
				{`hx-delete="/db/people/2/archive"`, tt.wantEdit},
				{`<th scope="col" class="text-center">メニュー</th>`, tt.wantColumn},
				{tt.wantMinWidth, true},
			} {
				if got := strings.Contains(html, c.s); got != c.want {
					t.Errorf("%qを含むか = %v、期待値 = %v", c.s, got, c.want)
				}
			}
		})
	}
}

// TestIndex_RevisionsLinkOrderは、行のメニューで「履歴」を操作の下に置くことを検証する。編集などの操作を
// 選ぶときに、操作ではない「履歴」を誤って選ばないようにするため。掲載停止中の行でも、掲載停止の解除の下に置く。
func TestIndex_RevisionsLinkOrder(t *testing.T) {
	t.Parallel()

	html := renderIndex(t, i18n.SetLocale(context.Background(), "ja"), IndexPageData{
		People: []viewmodel.DBPersonListItem{
			{ID: 1, Name: "人物1", Status: viewmodel.DBPersonStatusPublished},
			{ID: 3, Name: "人物3", Status: viewmodel.DBPersonStatusDelisted},
		},
		Pagination:    viewmodel.NewPagination(1, 2, 100, "/db/people"),
		IsCommitter:   true,
		IsAdmin:       true,
		ShowRevisions: true,
	})
	for _, row := range []struct {
		revisions  string
		operations []string
	}{
		{revisions: `href="/db/people/1/revisions"`, operations: []string{`href="/db/people/1/edit"`, `href="/db/people/1/archive/new"`, `href="/db/people/1/delisting/new"`, `hx-delete="/db/people/1"`}},
		{revisions: `href="/db/people/3/revisions"`, operations: []string{`href="/db/people/3/undelisting/new"`}},
	} {
		revisionsAt := strings.Index(html, row.revisions)
		if revisionsAt < 0 {
			t.Fatalf("%qが含まれていない", row.revisions)
		}
		for _, op := range row.operations {
			if at := strings.Index(html, op); at < 0 || at > revisionsAt {
				t.Errorf("%qの位置 = %d、期待値 = 「履歴」(%d) より前", op, at, revisionsAt)
			}
		}
	}
}
