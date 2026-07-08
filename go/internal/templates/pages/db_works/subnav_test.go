package db_works

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/templates"
	"github.com/annict/annict/go/internal/viewmodel"
)

// TestSubnav_RendersWorkSubResourceLinks verifies the vertical subnav links to every
// work sub-resource, labels them, exposes a navigation landmark, and marks the current
// (work edit) entry with aria-current.
//
// [Ja] TestSubnav_RendersWorkSubResourceLinks は、縦型サブナビが作品の各サブリソースへ
// リンクし、ラベルを付け、ナビゲーションランドマークを持ち、現在ページ (作品編集) の項目に
// aria-current を付けることをテストする。
func TestSubnav_RendersWorkSubResourceLinks(t *testing.T) {
	t.Parallel()

	ctx := templates.SetCurrentPath(context.Background(), "/db/works/1/edit")

	var buf strings.Builder
	if err := subnav(viewmodel.WorkID(1), false).Render(ctx, &buf); err != nil {
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
		// The work edit page is the current page, so its entry is marked.
		//
		// [Ja] 作品編集ページが現在ページのため、その項目に印が付く。
		`aria-current="page"`,
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}
}

// TestSubnav_OmitsEpisodeItemsWhenNoEpisodes verifies that the episode-derived entries
// (episodes, broadcast slots) are hidden when the work has no episodes, mirroring Rails.
//
// [Ja] TestSubnav_OmitsEpisodeItemsWhenNoEpisodes は、作品が「エピソード無し」のとき
// エピソード由来の項目 (エピソード・放送予定) が隠れることをテストする (Rails に合わせている)。
func TestSubnav_OmitsEpisodeItemsWhenNoEpisodes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var buf strings.Builder
	if err := subnav(viewmodel.WorkID(1), true).Render(ctx, &buf); err != nil {
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

// TestEdit_RendersSubnav verifies that the edit page wires the vertical subnav beside
// the form and passes the work id through to the links.
//
// [Ja] TestEdit_RendersSubnav は、編集画面がフォーム横に縦型サブナビを配線し、作品 ID を
// リンクまで通すことをテストする。
func TestEdit_RendersSubnav(t *testing.T) {
	t.Parallel()

	ctx := templates.SetCurrentPath(context.Background(), "/db/works/1/edit")
	data := EditPageData{
		CSRFToken: "test-csrf",
		WorkID:    1,
		FormInput: &viewmodel.DBWorkFormInput{},
	}

	var buf strings.Builder
	if err := Edit(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	for _, expected := range []string{
		`aria-label="作品ナビゲーション"`,
		`href="/db/works/1/episodes"`,
		`href="/db/works/1/casts"`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("編集画面にサブナビが描画されるべきです: %q", expected)
		}
	}
}

// TestEdit_SubnavOmitsEpisodeItemsWhenNoEpisodes verifies that the edit page forwards
// the no_episodes form value to the subnav so episode-derived entries are hidden.
//
// [Ja] TestEdit_SubnavOmitsEpisodeItemsWhenNoEpisodes は、編集画面が no_episodes の
// フォーム値をサブナビに渡し、エピソード由来の項目が隠れることをテストする。
func TestEdit_SubnavOmitsEpisodeItemsWhenNoEpisodes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	data := EditPageData{
		CSRFToken: "test-csrf",
		WorkID:    1,
		FormInput: &viewmodel.DBWorkFormInput{NoEpisodes: "1"},
	}

	var buf strings.Builder
	if err := Edit(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	if strings.Contains(html, "/db/works/1/episodes") {
		t.Error("エピソード無しのときエピソード項目は描画されてはいけません")
	}
	if strings.Contains(html, "/db/works/1/slots") {
		t.Error("エピソード無しのとき放送予定項目は描画されてはいけません")
	}
}

// TestNew_OmitsSubnav verifies that the new page renders no subnav, since there is no
// work yet to navigate the sub-resources of.
//
// [Ja] TestNew_OmitsSubnav は、新規画面ではサブナビを描画しないことをテストする。まだ
// サブリソースをたどる対象の作品が無いため。
func TestNew_OmitsSubnav(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	data := NewPageData{
		CSRFToken: "test-csrf",
		FormInput: &viewmodel.DBWorkFormInput{},
	}

	var buf strings.Builder
	if err := New(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	if strings.Contains(html, `aria-label="作品ナビゲーション"`) {
		t.Error("新規画面にサブナビは描画されてはいけません")
	}
	if strings.Contains(html, "/episodes") {
		t.Error("新規画面に作品サブリソースへのリンクは描画されてはいけません")
	}
}
