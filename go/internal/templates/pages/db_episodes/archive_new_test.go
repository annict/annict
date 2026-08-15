package db_episodes

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
)

// archiveNewTestData returns the page data of an episode opened for archive confirmation.
//
// [Ja] archiveNewTestData は非公開の確認のために開いたエピソードのページデータを返す。
func archiveNewTestData() ArchiveNewPageData {
	return ArchiveNewPageData{
		EpisodeID:   5,
		EpisodeName: "第2話「もう、お婿にいけません」",
		WorkID:      1,
		WorkName:    "テストアニメ",
		NoEpisodes:  false,
		CSRFToken:   "test-csrf-token",
	}
}

// TestArchiveNew verifies the confirmation page names the work in its heading and the episode
// in its confirmation message, keeps the shared work subnav, and posts to the episode's archive
// endpoint with a CSRF token.
//
// [Ja] TestArchiveNew は確認ページが見出しで作品を、確認メッセージでエピソードを名指しし、共有の
// 作品サブナビを保ち、CSRF トークン付きでエピソードの非公開エンドポイントへ POST することを
// 検証する。
func TestArchiveNew(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	var buf strings.Builder
	if err := ArchiveNew(archiveNewTestData()).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	html := buf.String()

	expectedContents := []string{
		// The heading names the work, as the other episode pages do; the episode is named by
		// the confirmation message below it.
		//
		// [Ja] 見出しは他のエピソードページと同じく作品を名指しする。エピソードはその下の
		// 確認メッセージが名指しする。
		">テストアニメ</h1>",
		"第2話「もう、お婿にいけません」を非公開にしますか？",
		// The page keeps the shared subnav, so the confirmation is not a dead end: the other
		// sections of the work stay reachable from it.
		//
		// [Ja] ページは共有サブナビを保つため、確認は行き止まりにならない。作品の他のセクション
		// にもここから到達できる。
		`href="/db/works/1/episodes"`,
		`aria-current="page"`,
		// The confirmation sits in the same card container as the other /db pages.
		//
		// [Ja] 確認内容は他の /db 画面と同じカードコンテナに載る。
		`class="card`,
		"<form",
		`method="POST"`,
		`action="/db/episodes/5/archive"`,
		`name="csrf_token"`,
		`value="test-csrf-token"`,
		// The execute button carries the warning variant the work archive confirmation uses,
		// and the cancel link the outline variant.
		//
		// [Ja] 実行ボタンは作品の非公開確認と同じ warning、キャンセルリンクは outline の
		// バリアントを持つ。
		`class="btn rounded-full" data-variant="warning"`,
		`class="btn rounded-full" data-variant="outline"`,
		// The header renders the sidebar toggle, wired to the sidebar at every viewport size.
		//
		// [Ja] ヘッダーはサイドバートグルを描画する。サイドバーに結線され、全画面幅で
		// 利用できる。
		`data-sidebar-toggle="db-sidebar"`,
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("レンダリング結果に %q が含まれていません", expected)
		}
	}
}

// TestArchiveNew_HeadingFallsBackWithoutWorkName verifies a work with no display name leaves the
// page a heading of its own rather than an empty <h1>, and that the episode is still named in
// the confirmation message.
//
// [Ja] TestArchiveNew_HeadingFallsBackWithoutWorkName は、表示名の無い作品でもページが空の <h1>
// ではなく自前の見出しを持ち、確認メッセージがエピソードを名指しし続けることを検証する。
func TestArchiveNew_HeadingFallsBackWithoutWorkName(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	data := archiveNewTestData()
	data.WorkName = ""

	var buf strings.Builder
	if err := ArchiveNew(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	html := buf.String()

	if !strings.Contains(html, ">エピソードを非公開にする</h1>") {
		t.Error("表示名の無い作品で汎用の見出しが描画されていません")
	}
	if !strings.Contains(html, "第2話「もう、お婿にいけません」を非公開にしますか？") {
		t.Error("確認メッセージがエピソードを名指ししていません")
	}
}

// TestArchiveNew_EscapesEpisodeName verifies the episode name is HTML-escaped: it comes from
// editor-entered columns, so it must not be able to inject markup into the confirmation.
//
// [Ja] TestArchiveNew_EscapesEpisodeName はエピソード名が HTML エスケープされることを検証する。
// 編集者が入力するカラム由来のため、確認にマークアップを差し込めてはならない。
func TestArchiveNew_EscapesEpisodeName(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	data := archiveNewTestData()
	data.EpisodeName = `<script>alert("x")</script>`

	var buf strings.Builder
	if err := ArchiveNew(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	if strings.Contains(buf.String(), "<script>") {
		t.Error("エピソード名がエスケープされずに描画されました")
	}
}
