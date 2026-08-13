package db_episodes

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates"
	"github.com/annict/annict/go/internal/viewmodel"
)

func decorativeIconMarkup(
	t *testing.T,
	ctx context.Context,
	name string,
	position ...templates.InlineIconPosition,
) string {
	t.Helper()

	component := templates.DecorativeIcon(name)
	if len(position) > 0 {
		component = templates.DecorativeInlineIcon(name, position[0])
	}

	var buf strings.Builder
	if err := component.Render(ctx, &buf); err != nil {
		t.Fatalf("アイコンのレンダリングエラー: %v", err)
	}

	return buf.String()
}

func TestNew_FreshForm(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	data := NewPageData{
		WorkID:    1,
		WorkName:  "テストアニメ",
		CSRFToken: "test-csrf-token",
	}

	var buf strings.Builder
	if err := New(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	expectedContents := []string{
		"テストアニメ",
		`action="/db/works/1/episodes"`,
		`value="test-csrf-token"`,
		// A fresh form starts out focused on the input, since there is no error summary to
		// take the caret.
		//
		// [Ja] 新しいフォームは入力欄にフォーカスして始まる。カーソルを受け取るエラー要約が
		// 無いため。
		"autofocus",
		"一覧に戻る",
		// The input is described by the format guide, so the instructions reach whoever lands
		// on the field rather than only whoever reads the page top to bottom.
		//
		// [Ja] 入力欄は形式の案内を説明として指すため、案内はページを上から読む人だけでなく、
		// 入力欄に降り立った人にも届く。
		`aria-describedby="rows-format-guide"`,
		`id="rows-format-guide"`,
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("レスポンスに %q が含まれていません", expected)
		}
	}

	if strings.Contains(html, `aria-invalid="true"`) {
		t.Error("エラーの無いフォームで aria-invalid が付いてはいけません")
	}
}

// TestNew_DecorativeIconsAreHidden covers the icons of the bulk-create page that repeat
// adjacent visible text. The warning, button and link already carry their meaning in text, so
// the SVGs stay out of the accessibility tree instead of adding a second, browser-dependent
// representation. The list page's own icon is covered by TestIndex_DecorativeIconsAreHidden.
//
// [Ja] TestNew_DecorativeIconsAreHidden は一括作成ページのアイコンのうち、隣接する可視テキスト
// と意味が重複するものを検証する。警告、ボタン、リンクはテキストですでに意味を伝えるため、
// SVG はアクセシビリティツリーから除外し、ブラウザー依存の別表現を重ねないようにする。一覧
// ページのアイコンは TestIndex_DecorativeIconsAreHidden が検証する。
func TestNew_DecorativeIconsAreHidden(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	var newHTML strings.Builder
	if err := New(NewPageData{
		WorkID:         1,
		ManualCreation: viewmodel.DBEpisodeManualCreationSlotsExist,
	}).Render(ctx, &newHTML); err != nil {
		t.Fatalf("新規作成ページのレンダリングエラー: %v", err)
	}

	tests := []struct {
		name     string
		html     string
		iconName string
		inline   bool
	}{
		{name: "作成制限の警告", html: newHTML.String(), iconName: "warning"},
		{name: "送信ボタン", html: newHTML.String(), iconName: "floppy-disk-regular", inline: true},
		{name: "一覧へ戻るリンク", html: newHTML.String(), iconName: "arrow-bend-up-left-regular", inline: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			want := decorativeIconMarkup(t, ctx, tt.iconName)
			if tt.inline {
				want = decorativeIconMarkup(t, ctx, tt.iconName, templates.InlineIconStart)
			}
			if !strings.Contains(tt.html, want) {
				t.Errorf("装飾アイコン %q が aria-hidden の要素内にありません", tt.iconName)
			}
		})
	}
}

// TestNew_WithErrors covers the page after a rejected submit: the messages are summarised at
// the top, the input is marked invalid and describes them, and the submitted lines are back in
// the textarea.
//
// [Ja] TestNew_WithErrors は送信が却下された後のページを検証する。メッセージが冒頭に要約され、
// 入力欄は不正と印付けられてそれらを説明として指し、送信された行は textarea に戻る。
func TestNew_WithErrors(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	formErrors := model.NewValidationError()
	formErrors.AddField("rows", "1 行目: 数値話数には数値を入力してください")
	formErrors.AddField("rows", "2 行目: 表示用話数かタイトルを入力してください")

	data := NewPageData{
		WorkID:     1,
		WorkName:   "テストアニメ",
		CSRFToken:  "test-csrf-token",
		FormErrors: formErrors,
		Rows:       "#1,いち,はじまり\n,,",
	}

	var buf strings.Builder
	if err := New(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	expectedContents := []string{
		`role="alert"`,
		"1 行目: 数値話数には数値を入力してください",
		"2 行目: 表示用話数かタイトルを入力してください",
		`aria-invalid="true"`,
		// The standing description keeps its place ahead of the errors, and every message
		// element is named so none of them goes unannounced.
		//
		// [Ja] 常設の説明はエラーの前に残り、メッセージ要素はすべて名指しされて読み上げから
		// 漏れるものが出ないようにする。
		`aria-describedby="rows-format-guide rows-error-1 rows-error-2"`,
		`id="rows-error-1"`,
		`id="rows-error-2"`,
		// The submitted lines come back so the editor corrects them instead of retyping.
		//
		// [Ja] 送信された行は書き戻され、編集者が入力し直さず手直しできる。
		"#1,いち,はじまり",
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("レスポンスに %q が含まれていません", expected)
		}
	}

	// The error summary takes focus on load, so the textarea must not claim it at the same
	// time.
	//
	// [Ja] エラー要約が読み込み時にフォーカスを受け取るため、textarea が同時に autofocus を
	// 主張してはいけない。
	if strings.Count(html, "autofocus") != 1 {
		t.Errorf("autofocus の数 = %d, want 1 (エラー要約のみ)", strings.Count(html, "autofocus"))
	}
}

func TestNew_ManualCreationRestriction(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	tests := []struct {
		name         string
		data         NewPageData
		wantMessage  string
		wantDisabled bool
	}{
		{
			name: "予定話数到達の編集者",
			data: NewPageData{
				WorkID:         1,
				ManualCreation: viewmodel.DBEpisodeManualCreationEpisodesFilled,
			},
			wantMessage:  "話数分のエピソードがすでに登録",
			wantDisabled: true,
		},
		{
			name: "放送枠がある管理者",
			data: NewPageData{
				WorkID:         1,
				IsAdmin:        true,
				ManualCreation: viewmodel.DBEpisodeManualCreationSlotsExist,
			},
			wantMessage:  "放送予定の情報を使って自動的に生成",
			wantDisabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf strings.Builder
			if err := New(tt.data).Render(ctx, &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}
			html := buf.String()
			if !strings.Contains(html, "手動登録できません") || !strings.Contains(html, tt.wantMessage) {
				t.Errorf("制限理由の警告が表示されていません")
			}
			if got := strings.Contains(html, "readonly"); got != tt.wantDisabled {
				t.Errorf("readonly = %v, want %v", got, tt.wantDisabled)
			}
			if got := strings.Contains(html, "disabled"); got != tt.wantDisabled {
				t.Errorf("disabled = %v, want %v", got, tt.wantDisabled)
			}
			// A disabled form takes no focus on load: landing in a field that cannot be typed
			// into would skip past the warning stating why.
			//
			// [Ja] 無効化されたフォームは読み込み時にフォーカスを取らない。入力できない欄に
			// 降りると、理由を述べている警告を飛び越えてしまうため。
			wantAutofocus := 0
			if !tt.wantDisabled {
				wantAutofocus = 1
			}
			if got := strings.Count(html, "autofocus"); got != wantAutofocus {
				t.Errorf("autofocus の数 = %d, want %d", got, wantAutofocus)
			}
		})
	}
}

// TestNew_RestrictionReportedOnceAfterRejectedSubmit covers a submit refused by the work's
// state: the error summary states the reason and takes focus, so the standing warning steps
// aside instead of stating it a second time. The form stays disabled either way.
//
// [Ja] TestNew_RestrictionReportedOnceAfterRejectedSubmit は作品の状態で却下された送信を検証
// する。エラー要約が理由を述べてフォーカスを受け取るため、常設の警告は退いて二重に述べない。
// フォームはどちらの場合も無効のままになる。
func TestNew_RestrictionReportedOnceAfterRejectedSubmit(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	formErrors := model.NewValidationError()
	formErrors.AddGlobal("話数分のエピソードがすでに登録されているため、エピソードを登録できません")

	var buf strings.Builder
	data := NewPageData{
		WorkID:         1,
		ManualCreation: viewmodel.DBEpisodeManualCreationEpisodesFilled,
		FormErrors:     formErrors,
		Rows:           "#2,2,つづき",
	}
	if err := New(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	if strings.Contains(html, "手動登録できません") {
		t.Error("エラー要約が理由を述べているときに常設の警告も描画されています")
	}
	if count := strings.Count(html, "話数分のエピソードがすでに登録"); count != 1 {
		t.Errorf("制限の理由の出現回数 = %d, want 1", count)
	}
	// The lines themselves are fine, so the textarea is not marked invalid.
	//
	// [Ja] 送信された行自体に問題は無いため、textarea は不正と印付けない。
	if strings.Contains(html, `aria-invalid="true"`) {
		t.Error("フォーム全体のエラーで textarea が不正と印付けられています")
	}
	if !strings.Contains(html, "readonly") || !strings.Contains(html, "disabled") {
		t.Error("制限された作品のフォームが無効化されていません")
	}
	// The error summary is the only autofocus candidate, so it receives focus and announces the
	// global error when the server-rendered page loads.
	//
	// [Ja] エラー要約だけを autofocus の候補にし、サーバー描画されたページの読み込み時に
	// フォーカスを受け取ってグローバルエラーを通知できるようにする。
	if count := strings.Count(html, "autofocus"); count != 1 {
		t.Errorf("autofocus の数 = %d, want 1 (エラー要約のみ)", count)
	}
}

// TestNew_HeadingFallsBackToPageTitle covers a work with no name to show: the page still needs
// a heading, so it falls back to the generic page title rather than rendering an empty one.
//
// [Ja] TestNew_HeadingFallsBackToPageTitle は表示できる名前が無い作品を検証する。ページには
// 見出しが要るため、空の見出しを描画せず汎用のページタイトルにフォールバックする。
func TestNew_HeadingFallsBackToPageTitle(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	var buf strings.Builder
	if err := New(NewPageData{WorkID: 1}).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	if !strings.Contains(buf.String(), "エピソード登録") {
		t.Error("名前の無い作品では汎用のページタイトルが見出しになるべきです")
	}
}
