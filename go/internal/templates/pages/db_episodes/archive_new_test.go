package db_episodes

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
)

// archiveNewTestDataは非公開の確認のために開いたエピソードのページデータを返す。
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

// TestArchiveNewは確認ページが見出しで作品を、確認メッセージでエピソードを名指しし、共有の
// 作品サブナビを保ち、CSRFトークン付きでエピソードの非公開エンドポイントへPOSTすることを
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
		// 見出しは他のエピソードページと同じく作品を名指しする。エピソードはその下の
		// 確認メッセージが名指しする。
		">テストアニメ</h1>",
		"第2話「もう、お婿にいけません」を非公開にしますか？",
		// ページは共有サブナビを保つため、確認は行き止まりにならない。作品の他のセクション
		// にもここから到達できる。
		`href="/db/works/1/episodes"`,
		`aria-current="page"`,
		// 確認内容は他の /db画面と同じカードコンテナに載る。
		`class="card`,
		"<form",
		`method="POST"`,
		`action="/db/episodes/5/archive"`,
		`name="csrf_token"`,
		`value="test-csrf-token"`,
		// 実行ボタンは作品の非公開確認と同じwarning、キャンセルリンクはoutlineの
		// バリアントを持つ。
		`class="btn rounded-full" data-variant="warning"`,
		`class="btn rounded-full" data-variant="outline"`,
		// ヘッダーはサイドバートグルを描画する。サイドバーに結線され、全画面幅で
		// 利用できる。
		`data-sidebar-toggle="db-sidebar"`,
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("レンダリング結果に%qが含まれていません", expected)
		}
	}
}

// TestArchiveNew_HeadingFallsBackWithoutWorkNameは、表示名の無い作品でもページが空の <h1>
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

	if !strings.Contains(html, ">エピソード非公開</h1>") {
		t.Error("表示名の無い作品で汎用の見出しが描画されていません")
	}
	if !strings.Contains(html, "第2話「もう、お婿にいけません」を非公開にしますか？") {
		t.Error("確認メッセージがエピソードを名指ししていません")
	}
}

// TestArchiveNew_EscapesEpisodeNameはエピソード名がHTMLエスケープされることを検証する。
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
