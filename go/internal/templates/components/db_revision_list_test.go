package components

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/viewmodel"
)

// 縦に並ぶ差分の表と行の表が、同じ固定レイアウトと同じ幅の1列目を持つことを確かめる。作成の版は変更前の列を持たない。
func TestDBRevisionList_ColumnWidths(t *testing.T) {
	t.Parallel()
	ctx := i18n.SetLocale(context.Background(), "ja")
	before, after := "前", "後"
	revisions := viewmodel.NewDBRevisionsFromAnimeRevisions(ctx, []*model.AnimeRevision{{
		DBEdit:   &model.DBEdit{Actor: model.DBEditActorSystem, Action: model.DBEditActionUpdate},
		Previous: &model.AnimeSnapshot{Title: &before},
		Snapshot: model.AnimeSnapshot{Title: &after, Seasons: []model.AnimeSnapshotSeason{{Year: 2026}}},
	}, {
		DBEdit:   &model.DBEdit{Actor: model.DBEditActorSystem, Action: model.DBEditActionCreate},
		Snapshot: model.AnimeSnapshot{Title: &after},
	}}, time.UTC, nil, viewmodel.DBRevisionIDNames{})

	var buf strings.Builder
	if err := DBRevisionList(revisions).Render(ctx, &buf); err != nil {
		t.Fatalf("描画に失敗: %v", err)
	}
	body := buf.String()

	table := `<table class="table table-fixed min-w-[480px]">`
	if got := strings.Count(body, table); got != 3 {
		t.Errorf("固定レイアウトの表の数 = %d、期待値 = 3", got)
	}
	// templは要素の間に空白を出すため、空白を除いて列の定義を比べる。
	columns := strings.ReplaceAll(body, "> <", "><")
	firstColumn := `<colgroup><col class="w-32 sm:w-48">`
	for _, expected := range []struct{ name, after string }{
		{name: "更新の版の差分の表", after: firstColumn + `<col><col></colgroup>`},
		{name: "作成の版の差分の表", after: firstColumn + `<col></colgroup>`},
		{name: "行の表", after: firstColumn + `<col class="w-20"><col></colgroup>`},
	} {
		if got := strings.Count(columns, expected.after); got != 1 {
			t.Errorf("%sの列の定義の数 = %d、期待値 = 1", expected.name, got)
		}
	}
}

// 作品の履歴で、作品の版とエピソードの版のどちらも、編集者と操作の文と日時のあとに版の対象の行が並ぶことを確かめる。
// 作品の名前はリンクにせず、エピソードの名前はエピソードの履歴へリンクする。
func TestDBRevisionList_WorkRevisionsTarget(t *testing.T) {
	t.Parallel()
	ctx := i18n.SetLocale(context.Background(), "ja")
	workTitle, episodeNumber := "作品のタイトル", "第1話"
	edit := &model.DBEdit{Actor: model.DBEditActorSystem, Action: model.DBEditActionUpdate}
	revisions := viewmodel.NewDBRevisionsFromAnimeRevisions(ctx, []*model.AnimeRevision{{
		DBEdit:   edit,
		Snapshot: model.AnimeSnapshot{Kind: model.AnimeClassificationKindWork, Title: &workTitle},
	}, {
		DBEdit:   edit,
		Snapshot: model.AnimeSnapshot{Kind: model.AnimeClassificationKindEpisode, NumberText: &episodeNumber},
		Episode:  &model.Episode{ID: 7},
	}}, time.UTC, nil, viewmodel.DBRevisionIDNames{})

	var buf strings.Builder
	if err := DBRevisionList(revisions).Render(ctx, &buf); err != nil {
		t.Fatalf("描画に失敗: %v", err)
	}
	items := strings.Split(buf.String(), "<li ")[1:]
	if len(items) != 2 {
		t.Fatalf("版の数 = %d、期待値 = 2", len(items))
	}
	for i, expected := range []struct{ name, target string }{
		{name: "作品の版", target: `<span class="text-muted-foreground">作品:</span> <span class="font-medium">作品のタイトル</span>`},
		{name: "エピソードの版", target: `<span class="text-muted-foreground">エピソード:</span> <a href="/db/episodes/7/revisions"`},
	} {
		item := items[i]
		heading := strings.Index(item, "</time>")
		target := strings.Index(item, expected.target)
		if heading < 0 || target < 0 {
			t.Errorf("%sに日時 (%d) か対象の行 (%d) が無い", expected.name, heading, target)
			continue
		}
		if target < heading {
			t.Errorf("%sの対象の行が、編集者と操作の文より前にある", expected.name)
		}
	}
	if strings.Contains(items[0], "/db/works/") {
		t.Error("作品の版の作品名が作品の履歴へのリンクになっている")
	}
}

// 名前の無い人物の版でも、リンクの名前と表のcaptionで対象が分かることを確かめる。
func TestDBEditList_UnnamedPerson(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ locale, name string }{
		{"ja", "名前の無い人物"},
		{"en", "Unnamed person"},
	} {
		for _, rawName := range []string{"", " \t"} {
			t.Run(tt.locale+"/"+rawName, func(t *testing.T) {
				t.Parallel()
				ctx := i18n.SetLocale(context.Background(), tt.locale)
				edit := &model.DBEdit{
					Actor: model.DBEditActorSystem, Action: model.DBEditActionCreate,
					PersonRevisions: []*model.PersonRevision{{PersonID: 42, Snapshot: model.PersonSnapshot{
						Name: rawName, Status: model.PersonStatusPublished,
					}}},
				}
				var buf strings.Builder
				if err := DBEditList(viewmodel.NewDBEdits(ctx, []*model.DBEdit{edit}, time.UTC, nil, viewmodel.DBRevisionIDNames{})).Render(ctx, &buf); err != nil {
					t.Fatalf("描画に失敗: %v", err)
				}
				body := buf.String()
				for _, expected := range []string{
					`href="/db/people/42/revisions"`,
					">" + tt.name + "</a>",
					`<caption class="sr-only">` + tt.name + ": ",
				} {
					if !strings.Contains(body, expected) {
						t.Errorf("描画結果に%qが無い", expected)
					}
				}
			})
		}
	}
}

// エピソードの版は、どの作品のエピソードかが分かるよう、リンクの名前と表のcaptionに今の親作品の名前を添える。
// 操作の見出しには、編集者へのリンクを1度だけ出す。
func TestDBEditList_EpisodeWithWork(t *testing.T) {
	t.Parallel()
	ctx := i18n.SetLocale(context.Background(), "ja")
	before, after := "第1話", "第2話"
	edit := &model.DBEdit{
		Actor: model.DBEditActorUser, Action: model.DBEditActionUpdate, User: &model.User{ID: 1, Username: "editor"},
		AnimeRevisions: []*model.AnimeRevision{{
			Snapshot: model.AnimeSnapshot{NumberText: &after}, Previous: &model.AnimeSnapshot{NumberText: &before},
			Episode: &model.Episode{ID: 7, WorkID: 3}, Work: &model.Work{ID: 3, Title: "親作品"},
		}},
	}
	var buf strings.Builder
	if err := DBEditList(viewmodel.NewDBEdits(ctx, []*model.DBEdit{edit}, time.UTC, nil, viewmodel.DBRevisionIDNames{})).Render(ctx, &buf); err != nil {
		t.Fatalf("描画に失敗: %v", err)
	}
	body := buf.String()
	for _, expected := range []string{
		`href="/db/episodes/7/revisions"`,
		">第2話 (親作品)</a>",
		`<caption class="sr-only">第2話 (親作品): `,
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("描画結果に%qが無い", expected)
		}
	}
	if got := strings.Count(body, `href="/@editor"`); got != 1 {
		t.Errorf("編集者へのリンクの数 = %d、期待値 = 1", got)
	}
}

// 編集者と操作を1つの文で示し、ユーザーの編集者だけをプロフィールへのリンクにすることを確かめる。
func TestDBRevisionList_EditSentence(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		locale   string
		edit     *model.DBEdit
		expected string
	}{
		{
			name: "ユーザーの更新", locale: "ja",
			edit:     &model.DBEdit{Actor: model.DBEditActorUser, Action: model.DBEditActionUpdate, User: &model.User{ID: 1, Username: "editor"}},
			expected: `<span><a href="/@editor" class="font-medium underline-offset-4 hover:underline">@editor</a>が更新しました</span>`,
		},
		{
			name: "システムの更新", locale: "ja",
			edit:     &model.DBEdit{Actor: model.DBEditActorSystem, Action: model.DBEditActionUpdate},
			expected: `<span><span class="font-medium">システム</span>が更新しました</span>`,
		},
		{
			name: "退会したユーザーの作成", locale: "ja",
			edit:     &model.DBEdit{Actor: model.DBEditActorUser, Action: model.DBEditActionCreate},
			expected: `<span><span class="font-medium">退会したユーザー</span>が作成しました</span>`,
		},
		{
			name: "英語のユーザーの更新", locale: "en",
			edit:     &model.DBEdit{Actor: model.DBEditActorUser, Action: model.DBEditActionUpdate, User: &model.User{ID: 1, Username: "editor"}},
			expected: `<span><a href="/@editor" class="font-medium underline-offset-4 hover:underline">@editor</a> updated</span>`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := i18n.SetLocale(context.Background(), tt.locale)
			revisions := viewmodel.NewDBRevisionsFromAnimeRevisions(ctx, []*model.AnimeRevision{{DBEdit: tt.edit}}, time.UTC, nil, viewmodel.DBRevisionIDNames{})
			edits := viewmodel.NewDBEdits(ctx, []*model.DBEdit{tt.edit}, time.UTC, nil, viewmodel.DBRevisionIDNames{})
			for _, c := range []struct {
				name      string
				component interface {
					Render(context.Context, io.Writer) error
				}
			}{
				{"各リソースの履歴", DBRevisionList(revisions)},
				{"全体の編集履歴", DBEditList(edits)},
			} {
				var buf strings.Builder
				if err := c.component.Render(ctx, &buf); err != nil {
					t.Fatalf("%sの描画に失敗: %v", c.name, err)
				}
				if !strings.Contains(buf.String(), tt.expected) {
					t.Errorf("%sの描画結果に%qが無い", c.name, tt.expected)
				}
			}
		})
	}
}

// 編集者の種類ごとに、文の前にアバター・頭文字・アイコンを示し、どれも支援技術から隠すことを確かめる。
func TestDBRevisionList_EditorAvatar(t *testing.T) {
	t.Parallel()
	ctx := i18n.SetLocale(context.Background(), "ja")
	for _, tt := range []struct {
		name     string
		header   viewmodel.DBEditHeader
		expected string
	}{
		{
			name:     "プロフィール画像を持つユーザー",
			header:   viewmodel.DBEditHeader{EditorKind: viewmodel.DBRevisionEditorUser, EditorUsername: "editor", EditorAvatarURL: "https://example.com/avatar.webp"},
			expected: `<img src="https://example.com/avatar.webp" alt="" width="20" height="20" class="size-5 rounded-full object-cover" loading="lazy">`,
		},
		{
			name:     "プロフィール画像の無いユーザー",
			header:   viewmodel.DBEditHeader{EditorKind: viewmodel.DBRevisionEditorUser, EditorUsername: "editor"},
			expected: `<span aria-hidden="true" class="flex size-5 items-center justify-center rounded-full bg-muted text-xs font-medium">e</span>`,
		},
		{
			name:     "システム",
			header:   viewmodel.DBEditHeader{EditorKind: viewmodel.DBRevisionEditorSystem},
			expected: `<svg aria-hidden="true" focusable="false" class="size-5 text-muted-foreground" xmlns="http://www.w3.org/2000/svg" width="32" height="32" fill="currentColor" viewBox="0 0 256 256"><path d="M200,48H136V16`,
		},
		{
			name:     "退会したユーザー",
			header:   viewmodel.DBEditHeader{EditorKind: viewmodel.DBRevisionEditorDeletedUser},
			expected: `<svg aria-hidden="true" focusable="false" class="size-5 text-muted-foreground" xmlns="http://www.w3.org/2000/svg" width="32" height="32" fill="currentColor" viewBox="0 0 256 256"><path d="M96.26,37A8`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for _, c := range []struct {
				name      string
				component interface {
					Render(context.Context, io.Writer) error
				}
			}{
				{"各リソースの履歴", DBRevisionList([]viewmodel.DBRevision{{DBEditHeader: tt.header}})},
				{"全体の編集履歴", DBEditList([]viewmodel.DBEdit{{DBEditHeader: tt.header}})},
			} {
				var buf strings.Builder
				if err := c.component.Render(ctx, &buf); err != nil {
					t.Fatalf("%sの描画に失敗: %v", c.name, err)
				}
				// アバターは見出しの先頭、文の前に置く。
				heading := `<p class="flex flex-wrap items-center gap-x-2 text-sm">` + tt.expected
				if !strings.Contains(buf.String(), heading) {
					t.Errorf("%sの描画結果の見出しの先頭に%qが無い", c.name, tt.expected)
				}
			}
		})
	}
}
