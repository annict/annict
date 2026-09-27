package db_episodes

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/templates"
	"github.com/annict/annict/go/internal/viewmodel"
)

// iconWrapperMarkupは、これらのページで許容しないラッパーのマークアップを表す。
// aria-hiddenの祖先はアイコンをアクセシビリティツリーから外すが、SVG要素を既定で
// フォーカス可能とする実装ではSVGがフォーカス順序に残りうるため。SVGを包む場合に
// 限って照合するのは、一覧の「#」見出しのように、記号のテキストを読み上げから外す
// aria-hiddenのspanは許容するため。
const iconWrapperMarkup = `<span aria-hidden="true"><svg`

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
		// 新しいフォームは入力欄にフォーカスして始まる。カーソルを受け取るエラー要約が
		// 無いため。
		"autofocus",
		"一覧に戻る",
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("レスポンスに%qが含まれていません", expected)
		}
	}

	if strings.Contains(html, `aria-invalid="true"`) {
		t.Error("エラーの無いフォームでaria-invalidが付いてはいけません")
	}

	// 入力欄が説明として指すのはエラーメッセージだけであるため、エラーを集めていない
	// フォームは何も名指ししない。指す先の要素が無いaria-describedbyを出すと、解決しない
	// 説明を欄に持たせることになる。
	if strings.Contains(html, `aria-describedby="rows-`) {
		t.Error("エラーの無いフォームでrows欄にaria-describedbyが付いてはいけません")
	}
}

// TestNew_DecorativeIconsAreHiddenは一括作成ページのアイコンのうち、隣接する可視テキスト
// と意味が重複するものを検証する。警告、ボタン、リンクはテキストですでに意味を伝えるため、
// SVGはアクセシビリティツリーから除外し、ブラウザー依存の別表現を重ねないようにする。一覧
// ページのアイコンはTestIndex_DecorativeIconsAreHiddenが検証する。
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
				t.Errorf("装飾アイコン%qがaria-hiddenの要素内にありません", tt.iconName)
			}
		})
	}
}

// TestNew_WithErrorsは送信が却下された後のページを検証する。メッセージが冒頭に要約され、
// 入力欄は不正と印付けられてそれらを説明として指し、送信された行はtextareaに戻る。
func TestNew_WithErrors(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	formErrors := &viewmodel.FormErrors{
		Fields: map[string][]string{"rows": {
			"1 行目: 話数には数値を入力してください",
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
		"1 行目: 話数には数値を入力してください",
		"2 行目: 表示用話数かタイトルを入力してください",
		`aria-invalid="true"`,
		// メッセージ要素はすべて名指しされ、読み上げから漏れるものが出ないようにする。
		`aria-describedby="rows-error-1 rows-error-2"`,
		`id="rows-error-1"`,
		`id="rows-error-2"`,
		// 送信された行は書き戻され、編集者が入力し直さず手直しできる。
		"#1,いち,はじまり",
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("レスポンスに%qが含まれていません", expected)
		}
	}

	// エラー要約が読み込み時にフォーカスを受け取るため、textareaが同時にautofocusを
	// 主張してはいけない。
	if strings.Count(html, "autofocus") != 1 {
		t.Errorf("autofocusの数 = %d、期待値 = 1 (エラー要約のみ)", strings.Count(html, "autofocus"))
	}
}

func TestNew_ManualCreationRestriction(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	tests := []struct {
		name         string
		data         NewPageData
		wantTitle    string
		wantMessage  string
		wantDisabled bool
	}{
		{
			name: "予定エピソード数到達の編集者",
			data: NewPageData{
				WorkID:         1,
				ManualCreation: viewmodel.DBEpisodeManualCreationEpisodesFilled,
			},
			wantTitle:    "手動登録できません",
			wantMessage:  "新規登録はできません",
			wantDisabled: true,
		},
		{
			name: "予定エピソード数到達の管理者",
			data: NewPageData{
				WorkID:         1,
				IsAdmin:        true,
				ManualCreation: viewmodel.DBEpisodeManualCreationEpisodesFilled,
			},
			wantTitle:    "通常は手動登録できません",
			wantMessage:  "管理者は手動でも登録できますが、予定エピソード数を超える",
			wantDisabled: false,
		},
		{
			name: "放送枠がある編集者",
			data: NewPageData{
				WorkID:         1,
				ManualCreation: viewmodel.DBEpisodeManualCreationSlotsExist,
			},
			wantTitle:    "手動登録できません",
			wantMessage:  "手動によるエピソード登録はできません",
			wantDisabled: true,
		},
		{
			name: "放送枠がある管理者",
			data: NewPageData{
				WorkID:         1,
				IsAdmin:        true,
				ManualCreation: viewmodel.DBEpisodeManualCreationSlotsExist,
			},
			wantTitle:    "通常は手動登録できません",
			wantMessage:  "管理者は手動でも登録できますが、自動生成されるエピソードと重複",
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
			// 編集者向けの見出しは管理者向けの見出しの後方一致になるため、要素全体で
			// 照合する。部分一致だけでは管理者向けの文言でも編集者のケースが通ってしまう。
			if !strings.Contains(html, "<h2>"+tt.wantTitle+"</h2>") {
				t.Errorf("制限の警告の見出しが%qではありません", tt.wantTitle)
			}
			if !strings.Contains(html, tt.wantMessage) {
				t.Errorf("制限理由の警告に%qが含まれていません", tt.wantMessage)
			}
			// 開始タグ全体を確認し、無関係な要素に同じバリアントがあるだけで
			// テストが通ることを防ぐ。
			if !strings.Contains(html, `<div class="alert max-w-2xl" data-variant="warning">`) {
				t.Error("警告がwarningバリアントのアラートとして描画されていません")
			}
			if got := strings.Contains(html, "readonly"); got != tt.wantDisabled {
				t.Errorf("readonly = %v、期待値 = %v", got, tt.wantDisabled)
			}
			if got := strings.Contains(html, "disabled"); got != tt.wantDisabled {
				t.Errorf("disabled = %v、期待値 = %v", got, tt.wantDisabled)
			}
			// 無効化されたフォームは読み込み時にフォーカスを取らない。入力できない欄に
			// 降りると、理由を述べている警告を飛び越えてしまうため。
			wantAutofocus := 0
			if !tt.wantDisabled {
				wantAutofocus = 1
			}
			if got := strings.Count(html, "autofocus"); got != wantAutofocus {
				t.Errorf("autofocusの数 = %d、期待値 = %d", got, wantAutofocus)
			}
		})
	}
}

// TestNew_RestrictionReportedOnceAfterRejectedSubmitは作品の状態で却下された送信を検証
// する。エラー要約が理由を述べてフォーカスを受け取るため、常設の警告は退いて二重に述べない。
// フォームはどちらの場合も無効のままになる。
func TestNew_RestrictionReportedOnceAfterRejectedSubmit(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	formErrors := &viewmodel.FormErrors{
		Global: []string{"予定エピソード数分のエピソードがすでに登録されているため、エピソードを登録できません"},
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
	if count := strings.Count(html, "予定エピソード数分のエピソードがすでに登録"); count != 1 {
		t.Errorf("制限の理由の出現回数 = %d、期待値 = 1", count)
	}
	// 送信された行自体に問題は無いため、textareaは不正と印付けない。
	if strings.Contains(html, `aria-invalid="true"`) {
		t.Error("フォーム全体のエラーでtextareaが不正と印付けられています")
	}
	if !strings.Contains(html, "readonly") || !strings.Contains(html, "disabled") {
		t.Error("制限された作品のフォームが無効化されていません")
	}
	// エラー要約だけをautofocusの候補にし、サーバー描画されたページの読み込み時に
	// フォーカスを受け取ってグローバルエラーを通知できるようにする。
	if count := strings.Count(html, "autofocus"); count != 1 {
		t.Errorf("autofocusの数 = %d、期待値 = 1 (エラー要約のみ)", count)
	}
}

// TestNew_HeadingFallsBackToPageTitleは表示できる名前が無い作品を検証する。ページには
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

// TestNew_FieldsUseBasecoatGroupsは一括作成フォームのラッパーを検証する。Basecoatの
// Fieldコンポーネントはrole="group" の要素である。classはスタイルだけを与え、グループの
// セマンティクスを持たないため、各ラッパーでroleをclassと合わせて保持する。
//
// textareaはキーボードヒントを持たない。1行が1エピソードのためEnterは改行を入れる
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
		t.Errorf("Basecoatのfield group = %d個、期待値 = 2個", got)
	}

	if strings.Contains(html, "enterkeyhint") {
		t.Error("行入力のtextareaはEnterで改行を入れるためenterkeyhintを持つべきではありません")
	}
}

// TestNew_GuidelineLinkは、一括作成ページが見出しの下にエピソードの一括登録ガイドライン
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
				// リンクは行き先を名乗り、エピソードの一括登録ガイドラインを指す。編集者が
				// 入力中の行をフォームに残せるよう、tabnabbing対策付きで新しいタブに開く。
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
