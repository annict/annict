package db_works

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/a-h/templ"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/templates"
	"github.com/annict/annict/go/internal/viewmodel"
)

// iconWrapperMarkupは、これらのページで許容しないラッパーのマークアップを表す。
// aria-hiddenの祖先はアイコンをアクセシビリティツリーから外すが、SVG要素を既定で
// フォーカス可能とする実装ではSVGがフォーカス順序に残りうるため。
const iconWrapperMarkup = `<span aria-hidden="true">`

// decorativeIconMarkupは本パッケージのページが出力するはずの形で1つのアイコンを描画
// する。SVGの属性を書き写すのではなく、アイコンがどのヘルパーを通るかをテストが表明できる
// ようにするため。positionを渡すとテキスト付きボタン内で使うBasecoatのinline形式になる。
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

// TestNew_LabelExternalLinksは、値が入っているラベルの横に外部リンクアイコンが描画される
// ことをテストする。
func TestNew_LabelExternalLinks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	data := NewPageData{
		CSRFToken: "test-csrf",
		FormInput: &viewmodel.DBWorkFormInput{
			OfficialSiteURL: "https://example.com",
			WikipediaURL:    "https://ja.wikipedia.org/wiki/x",
			TwitterUsername: "annict_com",
			TwitterHashtag:  "annict",
			ScTid:           "3524",
			MalAnimeID:      "20",
		},
	}

	var buf strings.Builder
	if err := New(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	expectedContents := []string{
		// 外部リンクはtabnabbing対策付きで新しいタブで開き、アイコンは装飾 (aria-hidden) なので
		// リンク側にアクセシブルネームを持つ。
		`target="_blank"`,
		`rel="noopener"`,
		`aria-label="公式サイトURL を新しいタブで開く"`,
		// URL系フィールドは送信値自体を、ID / ユーザー名系は共有ヘルパーで導出したURLをリンクする。
		`href="https://example.com"`,
		`href="https://ja.wikipedia.org/wiki/x"`,
		`href="https://x.com/annict_com"`,
		`href="http://cal.syoboi.jp/tid/3524"`,
		`href="https://myanimelist.net/anime/20"`,
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}
}

// TestNew_NoLabelExternalLinksWhenEmptyは、値が空のあいだフォームのラベルが外部リンクを
// 描画しないことをテストする。
func TestNew_NoLabelExternalLinksWhenEmpty(t *testing.T) {
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

	if !strings.Contains(html, "<form") {
		t.Error("フォームは描画されるべきです")
	}
	// 見出しの下のガイドラインリンクはページ自身のもので常に描画されるため、件数で
	// 「フィールドのラベルは自分のリンクを足していない」ことを表明する。
	if got := strings.Count(html, "を新しいタブで開く"); got != 1 {
		t.Errorf("新しいタブで開くリンクの数 = %d、期待値 = 1 (見出し下のガイドラインリンクのみ)", got)
	}
}

// TestNew_DecorativeIconsAreHiddenはフォームのラベル横にある新規タブアイコンを検証する。
// アイコンを囲むリンクが行き先を既に伝えるため、SVGはアクセシビリティツリーとフォーカス順序
// から除外し、ブラウザー依存の別表現を重ねないようにする。
func TestNew_DecorativeIconsAreHidden(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	data := NewPageData{
		FormInput: &viewmodel.DBWorkFormInput{OfficialSiteURL: "https://example.com"},
	}

	var buf strings.Builder
	if err := New(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	want := decorativeIconMarkup(t, ctx, "arrow-square-out-regular", "w-[18px] h-[18px]")
	if !strings.Contains(html, want) {
		t.Error(`装飾アイコン "arrow-square-out-regular" がaria-hiddenかつfocusable="false" ではありません`)
	}
	if strings.Contains(html, iconWrapperMarkup) {
		t.Error("装飾アイコンはラッパー要素ではなくSVG自体で隠すべきです")
	}
}

// TestNew_SidebarToggleは新規フォームがヘッダーにサイドバートグルを描画する
// ことを検証する。
func TestNew_SidebarToggle(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	if err := New(NewPageData{FormInput: &viewmodel.DBWorkFormInput{}}).Render(context.Background(), &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	html := buf.String()

	// トグルはサイドバーに結線され、全画面幅で利用できる。
	for _, expected := range []string{`data-sidebar-toggle="db-sidebar"`} {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}
}

// TestNew_RendersTitleRowAndCardは、新規画面が作品一覧と同じ組み方であることを検証する。
// 見出しと一覧へ戻る操作を持つタイトル行を置き、フォームはカードに載せる。
func TestNew_RendersTitleRowAndCard(t *testing.T) {
	t.Parallel()

	data := NewPageData{
		CSRFToken: "test-csrf",
		FormInput: &viewmodel.DBWorkFormInput{},
	}

	var buf strings.Builder
	if err := New(data).Render(context.Background(), &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	for _, expected := range []string{
		// 見出しは共有のMainTitleコンポーネントが描画し、その横の操作は作品一覧へ戻る。
		"<h1",
		"作品登録",
		`href="/db/works"`,
		"一覧に戻る",
		// フォームは一覧やフィルタフォームと同じくコンテンツカードに載る。
		`class="card`,
		"<form",
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}
}

// TestManualEpisodesCountLabelは、登録・編集フォームで共通のmanual_episodes_count欄を、
// 対応する両ロケールで作品の予定エピソード数として名付けることを検証する。
func TestManualEpisodesCountLabel(t *testing.T) {
	t.Parallel()

	pages := []struct {
		name      string
		component func() templ.Component
	}{
		{
			name: "新規",
			component: func() templ.Component {
				return New(NewPageData{
					FormInput: &viewmodel.DBWorkFormInput{},
				})
			},
		},
		{
			name: "編集",
			component: func() templ.Component {
				return Edit(EditPageData{
					WorkID:    1,
					WorkTitle: "編集対象アニメ",
					FormInput: &viewmodel.DBWorkFormInput{},
				})
			},
		},
	}
	locales := []struct {
		code string
		want string
	}{
		{code: "ja", want: "予定エピソード数"},
		{code: "en", want: "Expected Episodes"},
	}

	for _, page := range pages {
		for _, locale := range locales {
			t.Run(page.name+"/"+locale.code, func(t *testing.T) {
				t.Parallel()

				ctx := i18n.SetLocale(context.Background(), locale.code)
				var buf strings.Builder
				if err := page.component().Render(ctx, &buf); err != nil {
					t.Fatalf("レンダリングエラー: %v", err)
				}

				want := `<label for="manual_episodes_count" class="label">` + locale.want + "</label>"
				if !strings.Contains(buf.String(), want) {
					t.Errorf("出力に%qが含まれていません", want)
				}
			})
		}
	}
}

// validatedFieldsはvalidator.DBWorkCreateValidatorがエラーを付けうるフィールドの
// 一覧。両フォームはそのすべてにメッセージを描画する必要がある。行き場の無いエラーがあると、
// 422で送信が失敗した理由が画面に出ないまま終わるため。バリデーターはフィールド名をValidate
// 内に直書きしておりimportできる一覧が無いため、この写しをバリデーター側の追加に合わせて更新する。
var validatedFields = []string{
	"title",
	"media",
	"season_year",
	"season_name",
	"started_on",
	"ended_on",
	"official_site_url",
	"official_site_url_en",
	"wikipedia_url",
	"wikipedia_url_en",
	"twitter_username",
	"twitter_hashtag",
	"sc_tid",
	"mal_anime_id",
	"synopsis_source",
	"synopsis_source_en",
	"manual_episodes_count",
	"start_episode_raw_number",
	"number_format_id",
}

// fieldsWithStandingDescriptionは、エラーメッセージ以外の要素にも説明されている入力欄を
// その要素のidに対応付ける。input group内に表示する @ / # の接頭辞が「記号は入力済み」と
// いう指示を担っている。エラー時もaria-describedbyは接頭辞を指し続ける必要があるため、
// 期待値はエラーのidだけにはならない。
var fieldsWithStandingDescription = map[string]string{
	"twitter_username": "twitter_username-prefix",
	"twitter_hashtag":  "twitter_hashtag-prefix",
}

// TestFieldErrorsAreAssociatedWithInputsは、フィールドのエラーが対象の入力欄で伝わる
// ことを検証する。コントロールを不正としてマークし、メッセージを持つ要素を指すことで、
// 欄の横の表示を見られない利用者にもメッセージが届く。
func TestFieldErrorsAreAssociatedWithInputs(t *testing.T) {
	t.Parallel()

	formErrors := &viewmodel.FormErrors{Fields: map[string][]string{}}
	for _, field := range validatedFields {
		formErrors.Fields[field] = []string{field + " のエラーメッセージ"}
	}

	tests := []struct {
		name      string
		component templ.Component
	}{
		{
			name: "新規",
			component: New(NewPageData{
				FormErrors: formErrors,
				FormInput:  &viewmodel.DBWorkFormInput{},
			}),
		},
		{
			name: "編集",
			component: Edit(EditPageData{
				WorkID:     1,
				FormErrors: formErrors,
				FormInput:  &viewmodel.DBWorkFormInput{},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf strings.Builder
			if err := tt.component.Render(context.Background(), &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			html := buf.String()

			// メッセージはalertとして通知され、フィールド自身も不正な状態を持つため
			// ラベルまで含めてエラーとして読める。
			for _, expected := range []string{`role="alert"`, "data-invalid", `aria-invalid="true"`} {
				if !strings.Contains(html, expected) {
					t.Errorf("期待する文字列が含まれていません: %q", expected)
				}
			}

			for _, field := range validatedFields {
				describedBy := fmt.Sprintf(`aria-describedby="%s-error-1"`, field)
				if prefixID, ok := fieldsWithStandingDescription[field]; ok {
					describedBy = fmt.Sprintf(`aria-describedby="%s %s-error-1"`, prefixID, field)
				}

				for _, expected := range []string{
					describedBy,
					fmt.Sprintf(`id="%s-error-1"`, field),
					field + " のエラーメッセージ",
				} {
					if !strings.Contains(html, expected) {
						t.Errorf("%sのエラーが描画されていません: %q", field, expected)
					}
				}
			}
		})
	}
}

// TestFieldErrorsAreSummarisedAtTheTopOfTheFormは、送信の失敗がフォームの上に要約され、
// 各エラーが対象の入力欄へリンクされることを検証する。フォームは約25フィールドあり、下の方の
// エラーはページが返ってきた時点で画面外にある。要約が無いと、それを見つける手段はフォームを
// 端からスクロールすることだけになる。
func TestFieldErrorsAreSummarisedAtTheTopOfTheForm(t *testing.T) {
	t.Parallel()

	formErrors := &viewmodel.FormErrors{Fields: map[string][]string{}}
	for _, field := range validatedFields {
		formErrors.Fields[field] = []string{field + " のエラーメッセージ"}
	}

	tests := []struct {
		name      string
		component templ.Component
	}{
		{
			name: "新規",
			component: New(NewPageData{
				FormErrors: formErrors,
				FormInput:  &viewmodel.DBWorkFormInput{},
			}),
		},
		{
			name: "編集",
			component: Edit(EditPageData{
				WorkID:     1,
				FormErrors: formErrors,
				FormInput:  &viewmodel.DBWorkFormInput{},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), "ja")

			var buf strings.Builder
			if err := tt.component.Render(ctx, &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			html := buf.String()

			if !strings.Contains(html, "入力内容にエラーがあります") {
				t.Errorf("エラー要約が描画されていません\nHTML: %s", html)
			}

			// バリデーターが落としうるすべてのフィールドへ要約から到達でき、各項目は
			// フォームが表示しているラベルでフィールドを名指しする。
			for _, field := range validatedFields {
				if !strings.Contains(html, fmt.Sprintf(`<a href="#%s"`, field)) {
					t.Errorf("%sのエラーが要約からリンクされていません", field)
				}
			}

			// 送信が失敗した後に最初にフォーカスされるのは要約であるため、手つかずの
			// フォームでautofocusを持つタイトル欄はそれを譲る必要がある。
			if strings.Count(html, "autofocus") != 1 {
				t.Errorf("autofocusは要約だけが持つべきです\nHTML: %s", html)
			}
		})
	}
}

// TestFormWithoutErrorsFocusesTheTitleは、エラー無しで開いたフォームがタイトル欄に
// カーソルを置くことを検証する。表示する要約が無いため、ほかにフォーカスする先が無い。
func TestFormWithoutErrorsFocusesTheTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		component templ.Component
	}{
		{
			name:      "新規",
			component: New(NewPageData{FormInput: &viewmodel.DBWorkFormInput{}}),
		},
		{
			name:      "編集",
			component: Edit(EditPageData{WorkID: 1, FormInput: &viewmodel.DBWorkFormInput{}}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), "ja")

			var buf strings.Builder
			if err := tt.component.Render(ctx, &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			html := buf.String()

			if !strings.Contains(html, "autofocus") {
				t.Errorf("タイトル欄がautofocusを持つべきです\nHTML: %s", html)
			}

			if strings.Contains(html, "入力内容にエラーがあります") {
				t.Errorf("エラーが無いときに要約を出してはいけません\nHTML: %s", html)
			}
		})
	}
}

// TestRequiredFieldsAreMarkedInWordsは、両フォームの必須フィールドが共有の印を持ち、
// 素のアスタリスクが残っていないことを検証する。フォームにはアスタリスクを説明する凡例が
// 無いため、必須であることは言葉で示す必要がある。
func TestRequiredFieldsAreMarkedInWords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		component templ.Component
	}{
		{
			name:      "新規",
			component: New(NewPageData{FormInput: &viewmodel.DBWorkFormInput{}}),
		},
		{
			name:      "編集",
			component: Edit(EditPageData{WorkID: 1, FormInput: &viewmodel.DBWorkFormInput{}}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf strings.Builder
			if err := tt.component.Render(context.Background(), &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			html := buf.String()

			// 必須フィールド (タイトルとメディア) の2つが印を持ち、他のフィールドは持たない。
			if got := strings.Count(html, "必須"); got != 2 {
				t.Errorf("必須の印の数 = %d、期待値 = 2", got)
			}
			if strings.Contains(html, `<span class="text-destructive">*</span>`) {
				t.Error("必須は素のアスタリスクではなく言葉で示すべきです")
			}
		})
	}
}

// TestSocialFieldsCarryPrefixAffixは、Xユーザー名とハッシュタグの入力欄が、入力して
// ほしくない記号を入力欄の中に接頭辞として表示し、その接頭辞が見た目だけで終わらず入力欄の
// 説明として支援技術にも届くことを検証する。
func TestSocialFieldsCarryPrefixAffix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		component templ.Component
	}{
		{
			name:      "新規",
			component: New(NewPageData{FormInput: &viewmodel.DBWorkFormInput{}}),
		},
		{
			name:      "編集",
			component: Edit(EditPageData{WorkID: 1, FormInput: &viewmodel.DBWorkFormInput{}}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf strings.Builder
			if err := tt.component.Render(context.Background(), &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			html := buf.String()

			for _, expected := range []string{
				`<div class="input-group">`,
				`<span id="twitter_username-prefix" data-align="start">@</span>`,
				`<span id="twitter_hashtag-prefix" data-align="start">#</span>`,
				`aria-describedby="twitter_username-prefix"`,
				`aria-describedby="twitter_hashtag-prefix"`,
			} {
				if !strings.Contains(html, expected) {
					t.Errorf("期待する文字列が含まれていません: %q", expected)
				}
			}

			// 接頭辞はplaceholderにあった指示を置き換えるものであり、枠線はinput-groupの
			// シェルが描くため、入力欄はそのどちらも持たない。
			if strings.Contains(html, `name="twitter_username" class="input"`) {
				t.Error("input-group内の入力欄はinputクラスを持つべきではありません")
			}
			for _, unexpected := range []string{"@なし", "#なし"} {
				if strings.Contains(html, unexpected) {
					t.Errorf("接頭辞を表示するのでplaceholderの指示は残すべきではありません: %q", unexpected)
				}
			}
		})
	}
}

// TestNew_OmitsErrorAssociationWithoutErrorsは、エラーの無いフォームがどの入力欄も
// 不正としてマークしないことを検証する。存在しないエラーを支援技術へ伝えないため。
func TestNew_OmitsErrorAssociationWithoutErrors(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	if err := New(NewPageData{FormInput: &viewmodel.DBWorkFormInput{}}).Render(context.Background(), &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	// 検証対象はエラーの関連付けだけにする。接頭辞は入力の妥当性に関わらず入力欄を説明
	// するため、エラーの無いフォームにもaria-describedbyは現れる。
	for _, unexpected := range []string{`aria-invalid`, `-error-1"`, "data-invalid"} {
		if strings.Contains(html, unexpected) {
			t.Errorf("エラーが無いとき描画されてはいけません: %q", unexpected)
		}
	}
}

// TestEdit_SidebarToggleは編集フォームがヘッダーにサイドバートグルを描画する
// ことを検証する。
func TestEdit_SidebarToggle(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	if err := Edit(EditPageData{WorkID: 1, FormInput: &viewmodel.DBWorkFormInput{}}).Render(context.Background(), &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	html := buf.String()

	// トグルはサイドバーに結線され、全画面幅で利用できる。
	for _, expected := range []string{`data-sidebar-toggle="db-sidebar"`} {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}
}

// TestEdit_HeadingAndWorkPageLinkは、編集画面が編集対象の作品でページの見出しを付け、
// 操作として公開サイト側の作品ページを提供することを検証する。
func TestEdit_HeadingAndWorkPageLink(t *testing.T) {
	t.Parallel()

	data := EditPageData{
		WorkID:    1,
		WorkTitle: "編集対象アニメ",
		FormInput: &viewmodel.DBWorkFormInput{},
	}

	var buf strings.Builder
	if err := Edit(data).Render(context.Background(), &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	for _, expected := range []string{
		"<h1",
		"編集対象アニメ",
		// 操作は公開サイトの作品ページを指し、tabnabbing対策付きで新しいタブに開く。
		// そのことはアクセシブルネームで伝える。
		`href="/works/1"`,
		`target="_blank"`,
		`rel="noopener"`,
		`aria-label="作品ページを新しいタブで開く"`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}

	if strings.Contains(html, `href="/db/works"`) {
		t.Error("見出しの操作は一覧ではなく作品ページを指すべきです")
	}
}

// TestEdit_HeadingFallsBackToPageTitleは、作品に表示名が無いとき見出しが汎用の
// ページタイトルにフォールバックすることを検証する (タイトルが空の作品と、タイトル欄を空に
// した送信で、ハンドラーが渡す値がこれにあたる)。
func TestEdit_HeadingFallsBackToPageTitle(t *testing.T) {
	t.Parallel()

	data := EditPageData{
		WorkID:    1,
		WorkTitle: "",
		FormInput: &viewmodel.DBWorkFormInput{},
	}

	var buf strings.Builder
	if err := Edit(data).Render(context.Background(), &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	if !strings.Contains(buf.String(), "作品編集") {
		t.Error("作品に表示名が無いとき見出しはページタイトルを表示するべきです")
	}
}

// TestEdit_RendersSubnavは、編集画面がフォームの上に共有の作品サブナビを配線し、
// 作品IDをリンクまで通すことを検証する。
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

// TestEdit_SubnavOmitsEpisodeItemsWhenNoEpisodesは、編集画面がno_episodesの
// フォーム値をサブナビに渡し、エピソード由来の項目が隠れることを検証する。
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

// TestNew_OmitsSubnavは、新規画面ではサブナビを描画しないことを検証する。まだ
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

// TestEdit_LabelExternalLinksは、編集フォームでも共有サブテンプレート経由で外部リンクが
// 描画されることをテストする (新規フォームと同じ配線であることの担保)。
func TestEdit_LabelExternalLinks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	data := EditPageData{
		CSRFToken: "test-csrf",
		WorkID:    1,
		FormInput: &viewmodel.DBWorkFormInput{
			OfficialSiteURL: "https://example.com",
			ScTid:           "3524",
		},
	}

	var buf strings.Builder
	if err := Edit(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	expectedContents := []string{
		`aria-label="公式サイトURL を新しいタブで開く"`,
		`href="https://example.com"`,
		`href="http://cal.syoboi.jp/tid/3524"`,
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}
}

// TestEdit_DecorativeIconsAreHiddenは編集ページの2つの新規タブアイコンを検証する。
// どちらも囲むリンクが既に伝えている内容を繰り返すため、アクセシビリティツリーにもフォーカス
// 順序にも出ない。作品ページへのリンクはBasecoatのテキスト付きボタンのため、そのアイコンは
// Basecoatがボタンのアイコン用間隔を適用できるinline-endの位置も宣言する。フォームのラベル
// 横のアイコンは通常のリンク内にあるため位置を持たない。
func TestEdit_DecorativeIconsAreHidden(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	data := EditPageData{
		WorkID:    1,
		WorkTitle: "編集対象アニメ",
		FormInput: &viewmodel.DBWorkFormInput{OfficialSiteURL: "https://example.com"},
	}

	var buf strings.Builder
	if err := Edit(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	tests := []struct {
		name     string
		want     string
		wantHint string
	}{
		{
			name:     "作品ページへのリンク",
			want:     decorativeIconMarkup(t, ctx, "arrow-square-out-regular", "w-[18px] h-[18px]", templates.InlineIconEnd),
			wantHint: `inline-endの装飾アイコン`,
		},
		{
			name:     "フォームのラベル横のリンク",
			want:     decorativeIconMarkup(t, ctx, "arrow-square-out-regular", "w-[18px] h-[18px]"),
			wantHint: `位置を持たない装飾アイコン`,
		},
	}
	for _, tt := range tests {
		if !strings.Contains(html, tt.want) {
			t.Errorf("%s: %sが含まれていません", tt.name, tt.wantHint)
		}
	}

	if strings.Contains(html, iconWrapperMarkup) {
		t.Error("装飾アイコンはラッパー要素ではなくSVG自体で隠すべきです")
	}
}

// workFormFieldGroupsは各作品フォームが描画するFieldラッパーの数。数を固定すること
// で、後から追加した欄がグループ無しで書かれることを防ぐ (クラスだけでは気付けないため)。
const workFormFieldGroups = 27

// workKeyboardHintsは作品フォームの入力する欄が持つ属性を、フォームが表示する順で並べ
// る。最後の欄が "done" を求めるのは、その後ろに入力する欄が無いため (残るのはセレクトと
// チェックボックスで、タッチキーボードは次へ進むのではなく閉じてよい)。
//
// 数値カラムに対応する欄はtextの入力欄のままキーパッドを要求する。これにより却下された送信
// が入力された内容のまま戻る (理由はフォーム側のコメントを参照)。
var workKeyboardHints = []struct {
	field string
	attrs []string
}{
	{field: "title", attrs: []string{`enterkeyhint="next"`}},
	{field: "title_kana", attrs: []string{`enterkeyhint="next"`}},
	{field: "title_alter", attrs: []string{`enterkeyhint="next"`}},
	{field: "title_en", attrs: []string{`enterkeyhint="next"`}},
	{field: "title_alter_en", attrs: []string{`enterkeyhint="next"`}},
	{field: "official_site_url", attrs: []string{`enterkeyhint="next"`}},
	{field: "official_site_url_en", attrs: []string{`enterkeyhint="next"`}},
	{field: "wikipedia_url", attrs: []string{`enterkeyhint="next"`}},
	{field: "wikipedia_url_en", attrs: []string{`enterkeyhint="next"`}},
	{field: "twitter_username", attrs: []string{`enterkeyhint="next"`}},
	{field: "twitter_hashtag", attrs: []string{`enterkeyhint="next"`}},
	{field: "sc_tid", attrs: []string{`type="text"`, `inputmode="numeric"`, `enterkeyhint="next"`}},
	{field: "mal_anime_id", attrs: []string{`type="text"`, `inputmode="numeric"`, `enterkeyhint="next"`}},
	{field: "synopsis_source", attrs: []string{`enterkeyhint="next"`}},
	{field: "synopsis_source_en", attrs: []string{`enterkeyhint="next"`}},
	{field: "manual_episodes_count", attrs: []string{`type="text"`, `inputmode="numeric"`, `enterkeyhint="next"`}},
	{field: "start_episode_raw_number", attrs: []string{`type="text"`, `inputmode="decimal"`, `enterkeyhint="done"`}},
}

// workUntypedControlsは作品フォームのうち文字を入力しないコントロールを並べる。これらに
// キーボードヒントを付けても、読み手が触れないキーに札を付けることになる (セレクトと
// チェックボックスはキーボードを開かず、日付欄はピッカーを開き、textareaのEnterは改行を
// 入れるため)。
var workUntypedControls = []string{
	"media",
	"season_year",
	"season_name",
	"started_on",
	"ended_on",
	"synopsis",
	"synopsis_en",
	"number_format_id",
	"no_episodes",
}

// workControlHTMLは指定したidを持つフォームコントロールの開始タグのマークアップを返す
// (要素の種類を問わない)。
func workControlHTML(t *testing.T, html string, id string) string {
	t.Helper()

	at := strings.Index(html, `id="`+id+`"`)
	if at < 0 {
		t.Fatalf("%qのコントロールが描画されていません", id)
	}
	start := strings.LastIndex(html[:at], "<")
	if start < 0 {
		t.Fatalf("%qのコントロールの開始タグが見つかりません", id)
	}
	end := strings.Index(html[at:], ">")
	if end < 0 {
		t.Fatalf("%qのコントロールが閉じられていません", id)
	}

	return html[start : at+end+1]
}

// assertWorkFormFieldsは作品フォーム1つの欄の構造とキーボードヒントを検証する。
// 両フォームは同じ欄を表示するため、期待値は1箇所に置き、各ページはそれを満たすことを表明
// する。
func assertWorkFormFields(t *testing.T, html string) {
	t.Helper()

	if got := strings.Count(html, `role="group" class="field"`); got != workFormFieldGroups {
		t.Errorf("Basecoatのfield group = %d個、期待値 = %d個", got, workFormFieldGroups)
	}

	for _, hint := range workKeyboardHints {
		control := workControlHTML(t, html, hint.field)
		for _, attr := range hint.attrs {
			if !strings.Contains(control, attr) {
				t.Errorf("%qの入力欄に%sがありません: %s", hint.field, attr, control)
			}
		}
	}

	for _, id := range workUntypedControls {
		if control := workControlHTML(t, html, id); strings.Contains(control, "enterkeyhint") {
			t.Errorf("%qは文字を入力しないためenterkeyhintを持つべきではありません: %s", id, control)
		}
	}
}

// TestNew_FieldsCarryGroupsAndKeyboardHintsは新規作成フォームの欄のラッパーと入力欄の
// キーボードヒントを検証する。
func TestNew_FieldsCarryGroupsAndKeyboardHints(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	var buf strings.Builder
	if err := New(NewPageData{FormInput: &viewmodel.DBWorkFormInput{}}).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	assertWorkFormFields(t, buf.String())
}

// TestEdit_FieldsCarryGroupsAndKeyboardHintsは編集フォームの欄のラッパーと入力欄の
// キーボードヒントを検証する。
func TestEdit_FieldsCarryGroupsAndKeyboardHints(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	var buf strings.Builder
	if err := Edit(EditPageData{WorkID: 1, WorkTitle: "編集対象アニメ", FormInput: &viewmodel.DBWorkFormInput{}}).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	assertWorkFormFields(t, buf.String())
}

// TestEdit_CarriesVersionは、編集フォームが開いた時点の版を送り出すことを検証する。更新側は
// これを照合し、古い読み取りからの送信を、間に書いた人の変更を上書きせずに却下する。
func TestEdit_CarriesVersion(t *testing.T) {
	t.Parallel()

	data := EditPageData{
		WorkID:    1,
		FormInput: &viewmodel.DBWorkFormInput{UpdatedAt: "2026-08-17T01:02:03.456789Z"},
	}

	var buf strings.Builder
	if err := Edit(data).Render(context.Background(), &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	if html := buf.String(); !strings.Contains(html, `<input type="hidden" name="updated_at" value="2026-08-17T01:02:03.456789Z">`) {
		t.Error("フォームがhiddenの版を運んでいません")
	}
}

// TestEdit_ConflictNoticeListsChangedStoredValuesは、競合の案内が、2回目の送信で上書き
// される保存済みの値を述べることを検証する。並ぶのは異なるフィールドだけで、値は選択欄や
// チェックボックスが保持するコードではなく、フォームが表示する形にする。
func TestEdit_ConflictNoticeListsChangedStoredValues(t *testing.T) {
	t.Parallel()

	data := EditPageData{
		WorkID: 1,
		FormOptions: viewmodel.DBWorkFormOptions{
			MediaOptions: []viewmodel.SelectOption{
				{Value: "1", Label: "TV"},
				{Value: "3", Label: "映画"},
			},
		},
		FormInput: &viewmodel.DBWorkFormInput{
			Title:      "送信されたタイトル",
			TitleEn:    "同じ英語タイトル",
			Media:      "1",
			NoEpisodes: "",
			UpdatedAt:  "2026-08-17T01:02:03.456789Z",
		},
		ConflictCurrent: &viewmodel.DBWorkFormInput{
			Title:      "保存済みのタイトル",
			TitleEn:    "同じ英語タイトル",
			Media:      "3",
			NoEpisodes: "1",
			UpdatedAt:  "2026-08-17T01:02:03.456789Z",
		},
	}

	var buf strings.Builder
	if err := Edit(data).Render(context.Background(), &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	html := buf.String()

	for _, expected := range []string{
		"現在保存されている内容",
		"<dt>タイトル</dt>",
		"保存済みのタイトル",
		// 保存済みのメディアはenumのコードではなく、選択欄が表示するラベルで名指しする。
		"<dt>メディア</dt>",
		"映画",
		// チェックボックスは保持する "1" ではなく、入っているかどうかを述べる。
		"<dt>エピソードなし</dt>",
		"オン",
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}

	// 両者が一致するフィールドは並べない。案内は2回目の送信が上書きするものを名指しする
	// ものであり、そのフィールドは何も変わらないため。
	if strings.Contains(html, "<dt>英語タイトル</dt>") {
		t.Error("一致するフィールドが競合の案内に並んでいます")
	}
}

// TestEdit_ConflictNoticeWithoutFieldChangesは、保存済みの行が本フォームの書き込む
// どのフィールドでも送信と異ならないとき、案内が空の一覧ではなくそのことを述べるのを検証する。
func TestEdit_ConflictNoticeWithoutFieldChanges(t *testing.T) {
	t.Parallel()

	submitted := viewmodel.DBWorkFormInput{Title: "同じタイトル", Media: "1"}
	stored := submitted
	data := EditPageData{
		WorkID:          1,
		FormInput:       &submitted,
		ConflictCurrent: &stored,
	}

	var buf strings.Builder
	if err := Edit(data).Render(context.Background(), &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	html := buf.String()

	if !strings.Contains(html, "このフォームで編集できる項目に違いはありません") {
		t.Error("違いが無いことを述べる案内が含まれていません")
	}
	if strings.Contains(html, "<dl") {
		t.Error("違いが無いのに保存済みの値の一覧が描画されています")
	}
}

// TestEdit_OmitsConflictNoticeWithoutConflictは、編集のために開いたフォームと、他の理由で
// 却下された送信の再描画のいずれにも競合の案内が出ないことを検証する。
func TestEdit_OmitsConflictNoticeWithoutConflict(t *testing.T) {
	t.Parallel()

	data := EditPageData{
		WorkID:    1,
		FormInput: &viewmodel.DBWorkFormInput{Title: "編集中のタイトル"},
	}

	var buf strings.Builder
	if err := Edit(data).Render(context.Background(), &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	if html := buf.String(); strings.Contains(html, "現在保存されている内容") {
		t.Error("競合していないのに競合の案内が描画されています")
	}
}

// TestGuidelineLinkBelowHeadingは、両方の作品フォームが見出しの下に作品の編集ガイド
// ラインへの導線を持つことを検証する。フォームを開いた編集者が、自力でヘルプページを探さずに
// 画面からガイドラインへ辿れるようにするため。
func TestGuidelineLinkBelowHeading(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		component templ.Component
	}{
		{
			name:      "新規",
			component: New(NewPageData{FormInput: &viewmodel.DBWorkFormInput{}}),
		},
		{
			name:      "編集",
			component: Edit(EditPageData{WorkID: 1, FormInput: &viewmodel.DBWorkFormInput{}}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf strings.Builder
			if err := tt.component.Render(context.Background(), &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			html := buf.String()

			for _, expected := range []string{
				// リンクは行き先を名乗り、作品の編集ガイドラインを指す。編集者が入力中の
				// フォームを開いたままにできるよう、tabnabbing対策付きで新しいタブに開く。
				"作品の編集ガイドライン",
				`href="` + viewmodel.HelpWorkEditingURL() + `"`,
				`aria-label="作品の編集ガイドライン を新しいタブで開く"`,
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
