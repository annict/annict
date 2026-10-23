package db_people

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/viewmodel"
)

// TestDelistingNewは掲載停止確認ページが人物の名前を見出しと確認メッセージに描画し、保存済みの
// 対応記録を説明付きの入力欄の初期値にして、CSRFトークンと読み手の来た一覧を持ち回って掲載停止の
// エンドポイントへ送信することを検証する。
func TestDelistingNew(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	var buf strings.Builder
	if err := DelistingNew(DelistingNewPageData{
		CSRFToken:     "test-csrf",
		Person:        viewmodel.DBPersonDelistingTarget{ID: 42, Name: "確認対象の人物", DelistMessage: "以前の対応記録"},
		DelistMessage: "以前の対応記録",
		ReturnTo:      "/db/search?q=test",
	}).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	html := buf.String()

	for _, expected := range []string{
		`action="/db/people/42/delisting"`,
		`method="POST"`,
		`name="csrf_token" value="test-csrf"`,
		">確認対象の人物</h1>",
		"「確認対象の人物」の掲載を停止しますか？",
		`<label for="delist_message" class="label">`,
		`aria-describedby="delist_message-description"`,
		`<p id="delist_message-description"`,
		">以前の対応記録</textarea>",
		`class="btn rounded-full" data-variant="destructive"`,
		`<a href="/db/search?q=test"`,
		`name="return_to" value="/db/search?q=test"`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}

	// 掲載停止はPOSTのまま送るため、メソッドオーバーライドを持たない。
	if strings.Contains(html, `name="_method"`) {
		t.Error("掲載停止のフォームがメソッドオーバーライドを持っています")
	}
	if strings.Contains(html, "保存済みの対応記録") {
		t.Error("競合していない確認画面が保存済みの対応記録を示しています")
	}
}

// TestDelistingNew_Conflictは、競合で再描画した確認画面が送信された対応記録を入力欄に残し、それと
// 異なる保存済みの対応記録だけを入力欄とは別に示すことを検証する。
func TestDelistingNew_Conflict(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		saved         string
		submitted     string
		wantSaved     string
		wantSavedShow bool
	}{
		{name: "保存済みの記録が送信と異なる", saved: "別の管理者の記録", submitted: "送信した記録", wantSaved: ">別の管理者の記録</p>", wantSavedShow: true},
		{name: "保存済みの記録が無い", submitted: "送信した記録", wantSaved: "保存済みの対応記録はありません。", wantSavedShow: true},
		{name: "前後の空白を除くと送信と同じ", saved: "送信した記録", submitted: " 送信した記録\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), "ja")
			var buf strings.Builder
			if err := DelistingNew(DelistingNewPageData{
				Person:        viewmodel.DBPersonDelistingTarget{ID: 42, Name: "確認対象の人物", DelistMessage: tt.saved},
				DelistMessage: tt.submitted,
				Conflict:      true,
			}).Render(ctx, &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}
			html := buf.String()

			if !strings.Contains(html, ">"+tt.submitted+"</textarea>") {
				t.Errorf("入力欄に送信した対応記録 %q が残っていません: %s", tt.submitted, html)
			}
			if got := strings.Contains(html, `aria-labelledby="saved_delist_message-heading"`); got != tt.wantSavedShow {
				t.Errorf("保存済みの対応記録の表示 = %v、期待値 = %v", got, tt.wantSavedShow)
			}
			if tt.wantSaved != "" && !strings.Contains(html, tt.wantSaved) {
				t.Errorf("レスポンスに含まれていない文字列 = %q", tt.wantSaved)
			}
		})
	}
}
