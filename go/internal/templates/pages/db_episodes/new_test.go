package db_episodes

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/templates"
	"github.com/annict/annict/go/internal/viewmodel"
)

// iconWrapperMarkup is wrapper markup these pages reject. An aria-hidden ancestor removes the
// icon from the accessibility tree but can leave the SVG in the focus order on implementations
// that treat SVG elements as focusable by default.
//
// [Ja] iconWrapperMarkup は、これらのページで許容しないラッパーのマークアップを表す。
// aria-hidden の祖先はアイコンをアクセシビリティツリーから外すが、SVG 要素を既定で
// フォーカス可能とする実装では SVG がフォーカス順序に残りうるため。
const iconWrapperMarkup = `<span aria-hidden="true">`

func decorativeIconMarkup(
	t *testing.T,
	ctx context.Context,
	name string,
	class string,
	position ...templates.InlineIconPosition,
) string {
	t.Helper()

	component := templates.DecorativeIcon(name, class)
	if len(position) > 0 {
		component = templates.DecorativeInlineIcon(name, position[0], class)
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
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("レスポンスに %q が含まれていません", expected)
		}
	}

	if strings.Contains(html, `aria-invalid="true"`) {
		t.Error("エラーの無いフォームで aria-invalid が付いてはいけません")
	}

	// The input describes the error messages alone, so a form that collected none names
	// nothing: an aria-describedby pointing at no element would leave the field with a
	// description that never resolves.
	//
	// [Ja] 入力欄が説明として指すのはエラーメッセージだけであるため、エラーを集めていない
	// フォームは何も名指ししない。指す先の要素が無い aria-describedby を出すと、解決しない
	// 説明を欄に持たせることになる。
	if strings.Contains(html, `aria-describedby="rows-`) {
		t.Error("エラーの無いフォームで rows 欄に aria-describedby が付いてはいけません")
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

			want := decorativeIconMarkup(t, ctx, tt.iconName, "")
			if tt.inline {
				want = decorativeIconMarkup(t, ctx, tt.iconName, "", templates.InlineIconStart)
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

	formErrors := &viewmodel.FormErrors{
		Fields: map[string][]string{"rows": {
			"1 行目: 数値話数には数値を入力してください",
			"2 行目: 表示用話数かタイトルを入力してください",
		}},
	}

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
		// Every message element is named, so none of them goes unannounced.
		//
		// [Ja] メッセージ要素はすべて名指しされ、読み上げから漏れるものが出ないようにする。
		`aria-describedby="rows-error-1 rows-error-2"`,
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
			// Checking the whole opening tag ties the warning variant to the alert element instead of
			// accepting the same variant on an unrelated element.
			//
			// [Ja] 開始タグ全体を確認し、無関係な要素に同じバリアントがあるだけで
			// テストが通ることを防ぐ。
			if !strings.Contains(html, `<div class="alert max-w-2xl" data-variant="warning">`) {
				t.Error("警告が warning バリアントのアラートとして描画されていません")
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

	formErrors := &viewmodel.FormErrors{
		Global: []string{"話数分のエピソードがすでに登録されているため、エピソードを登録できません"},
	}

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

// TestNew_FieldsUseBasecoatGroups covers the wrappers of the bulk-create form. The Basecoat
// Field component is a role="group" element. The class provides styling but no grouping
// semantics, so each wrapper keeps the role alongside the class.
//
// The textarea takes no keyboard hint: one line is one episode, so its Enter key inserts a
// line break, and labelling that key "next" or "done" would promise a move the key never
// makes.
//
// [Ja] TestNew_FieldsUseBasecoatGroups は一括作成フォームのラッパーを検証する。Basecoat の
// Field コンポーネントは role="group" の要素である。class はスタイルだけを与え、グループの
// セマンティクスを持たないため、各ラッパーで role を class と合わせて保持する。
//
// textarea はキーボードヒントを持たない。1 行が 1 エピソードのため Enter は改行を入れる
// キーであり、そこに "next" や "done" の札を付けるとキーが行わない移動を約束することになる。
func TestNew_FieldsUseBasecoatGroups(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	var buf strings.Builder
	if err := New(NewPageData{WorkID: 1, WorkName: "テストアニメ", CSRFToken: "test-csrf-token"}).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	if got := strings.Count(html, `role="group" class="field"`); got != 2 {
		t.Errorf("Basecoat の field group = %d 個, want 2 個", got)
	}

	if strings.Contains(html, "enterkeyhint") {
		t.Error("行入力の textarea は Enter で改行を入れるため enterkeyhint を持つべきではありません")
	}
}

// TestNew_GuidelineLink verifies that the bulk-create page offers the bulk registration
// guideline below its heading. The page states the line format nowhere itself, so this link is
// what an editor follows to learn the column order and the shapes a partly filled line takes.
//
// [Ja] TestNew_GuidelineLink は、一括作成ページが見出しの下にエピソードの一括登録ガイドライン
// への導線を持つことを検証する。ページ自身は行の形式をどこにも述べないため、列の順序や一部
// だけ入力した行の形を編集者が知るには、このリンクを辿ることになる。
func TestNew_GuidelineLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		locale    string
		label     string
		ariaLabel string
	}{
		{
			name:      "日本語",
			locale:    "ja",
			label:     "エピソードの一括登録ガイドライン",
			ariaLabel: "エピソードの一括登録ガイドライン を新しいタブで開く",
		},
		{
			name:      "英語",
			locale:    "en",
			label:     "Bulk episode registration guidelines",
			ariaLabel: "Open Bulk episode registration guidelines in a new tab",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), tt.locale)

			var buf strings.Builder
			if err := New(NewPageData{WorkID: 1, WorkName: "テストアニメ"}).Render(ctx, &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			html := buf.String()

			for _, expected := range []string{
				// The link names where it goes and points at the bulk registration guideline,
				// opening it in a new tab with tabnabbing protection so the lines the editor is
				// entering stay in the form.
				//
				// [Ja] リンクは行き先を名乗り、エピソードの一括登録ガイドラインを指す。編集者が
				// 入力中の行をフォームに残せるよう、tabnabbing 対策付きで新しいタブに開く。
				">" + tt.label + "<",
				`href="` + viewmodel.HelpEpisodeBulkCreateURL() + `"`,
				`aria-label="` + tt.ariaLabel + `"`,
				`target="_blank"`,
				`rel="noopener"`,
			} {
				if !strings.Contains(html, expected) {
					t.Errorf("期待する文字列が含まれていません: %q", expected)
				}
			}
		})
	}
}
