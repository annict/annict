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

// iconWrapperMarkup is wrapper markup these pages reject. An aria-hidden ancestor removes the
// icon from the accessibility tree but can leave the SVG in the focus order on implementations
// that treat SVG elements as focusable by default.
//
// [Ja] iconWrapperMarkup は、これらのページで許容しないラッパーのマークアップを表す。
// aria-hidden の祖先はアイコンをアクセシビリティツリーから外すが、SVG 要素を既定で
// フォーカス可能とする実装では SVG がフォーカス順序に残りうるため。
const iconWrapperMarkup = `<span aria-hidden="true">`

// decorativeIconMarkup renders one icon the way the pages of this package are expected to
// emit it, so a test states which helper an icon goes through rather than repeating the SVG
// attributes. Passing a position asks for the Basecoat inline form used inside text buttons.
//
// [Ja] decorativeIconMarkup は本パッケージのページが出力するはずの形で 1 つのアイコンを描画
// する。SVG の属性を書き写すのではなく、アイコンがどのヘルパーを通るかをテストが表明できる
// ようにするため。position を渡すとテキスト付きボタン内で使う Basecoat の inline 形式になる。
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

// TestNew_LabelExternalLinks verifies that an external-link icon is rendered next to a label
// whose field has a value.
//
// [Ja] TestNew_LabelExternalLinks は、値が入っているラベルの横に外部リンクアイコンが描画される
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
		// The external link opens in a new tab with tabnabbing protection and carries an
		// accessible name because the icon itself is decorative (aria-hidden).
		//
		// [Ja] 外部リンクは tabnabbing 対策付きで新しいタブで開き、アイコンは装飾 (aria-hidden) なので
		// リンク側にアクセシブルネームを持つ。
		`target="_blank"`,
		`rel="noopener"`,
		`aria-label="公式サイトURL を新しいタブで開く"`,
		// URL fields link to the submitted value itself; the id / username fields derive
		// their service URL via the shared helpers.
		//
		// [Ja] URL 系フィールドは送信値自体を、ID / ユーザー名系は共有ヘルパーで導出した URL をリンクする。
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

// TestNew_NoLabelExternalLinksWhenEmpty verifies that no form label renders an external
// link while the linkable fields are empty.
//
// [Ja] TestNew_NoLabelExternalLinksWhenEmpty は、値が空のあいだフォームのラベルが外部リンクを
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
	// The guideline link under the heading belongs to the page itself and always renders,
	// so the count states that no field label added one of its own.
	//
	// [Ja] 見出しの下のガイドラインリンクはページ自身のもので常に描画されるため、件数で
	// 「フィールドのラベルは自分のリンクを足していない」ことを表明する。
	if got := strings.Count(html, "を新しいタブで開く"); got != 1 {
		t.Errorf("新しいタブで開くリンクの数 = %d, want 1 (見出し下のガイドラインリンクのみ)", got)
	}
}

// TestNew_DecorativeIconsAreHidden covers the new-tab icon beside a form label: the link
// around it already announces where it goes, so the SVG stays out of the accessibility tree
// and out of the focus order instead of adding a second, browser-dependent representation.
//
// [Ja] TestNew_DecorativeIconsAreHidden はフォームのラベル横にある新規タブアイコンを検証する。
// アイコンを囲むリンクが行き先を既に伝えるため、SVG はアクセシビリティツリーとフォーカス順序
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
		t.Error(`装飾アイコン "arrow-square-out-regular" が aria-hidden かつ focusable="false" ではありません`)
	}
	if strings.Contains(html, iconWrapperMarkup) {
		t.Error("装飾アイコンはラッパー要素ではなく SVG 自体で隠すべきです")
	}
}

// TestNew_SidebarToggle verifies the new form renders the sidebar toggle in its header.
//
// [Ja] TestNew_SidebarToggle は新規フォームがヘッダーにサイドバートグルを描画する
// ことを検証する。
func TestNew_SidebarToggle(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	if err := New(NewPageData{FormInput: &viewmodel.DBWorkFormInput{}}).Render(context.Background(), &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	html := buf.String()

	// The toggle is wired to the sidebar at every viewport size.
	//
	// [Ja] トグルはサイドバーに結線され、全画面幅で利用できる。
	for _, expected := range []string{`data-sidebar-toggle="db-sidebar"`} {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}
}

// TestNew_RendersTitleRowAndCard verifies that the new page is built like the work list:
// a title row carrying the heading and the back-to-list action, with the form on a card.
//
// [Ja] TestNew_RendersTitleRowAndCard は、新規画面が作品一覧と同じ組み方であることを検証する。
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
		// The heading comes from the shared MainTitle component, and the action beside it
		// links back to the work list.
		//
		// [Ja] 見出しは共有の MainTitle コンポーネントが描画し、その横の操作は作品一覧へ戻る。
		"<h1",
		"作品登録",
		`href="/db/works"`,
		"一覧に戻る",
		// The form sits on a content card, like the list and filter form do.
		//
		// [Ja] フォームは一覧やフィルタフォームと同じくコンテンツカードに載る。
		`class="card`,
		"<form",
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}
}

// TestManualEpisodesCountLabel verifies that the create and edit forms name their shared
// manual_episodes_count field as the work's expected total in both supported locales.
//
// [Ja] TestManualEpisodesCountLabel は、登録・編集フォームで共通の manual_episodes_count 欄を、
// 対応する両ロケールで作品の予定総話数として名付けることを検証する。
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
		{code: "ja", want: "予定総話数"},
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
					t.Errorf("出力に %q が含まれていません", want)
				}
			})
		}
	}
}

// validatedFields lists every field that validator.DBWorkCreateValidator can attach an
// error to. Both forms have to render a message for each of them: an error with nowhere
// to go leaves the submit failing at 422 with no visible reason. The validator names its
// fields inline in Validate, so there is no list to import: this copy has to grow with it.
//
// [Ja] validatedFields は validator.DBWorkCreateValidator がエラーを付けうるフィールドの
// 一覧。両フォームはそのすべてにメッセージを描画する必要がある。行き場の無いエラーがあると、
// 422 で送信が失敗した理由が画面に出ないまま終わるため。バリデーターはフィールド名を Validate
// 内に直書きしており import できる一覧が無いため、この写しをバリデーター側の追加に合わせて更新する。
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

// fieldsWithStandingDescription maps the fields whose input is described by something else
// besides its error message to that element's id: the @ / # prefix shown inside their input
// group carries the instruction that the symbol is already typed. Their aria-describedby has
// to keep naming the prefix once an error appears, so the expected value is not the error id
// alone.
//
// [Ja] fieldsWithStandingDescription は、エラーメッセージ以外の要素にも説明されている入力欄を
// その要素の id に対応付ける。input group 内に表示する @ / # の接頭辞が「記号は入力済み」と
// いう指示を担っている。エラー時も aria-describedby は接頭辞を指し続ける必要があるため、
// 期待値はエラーの id だけにはならない。
var fieldsWithStandingDescription = map[string]string{
	"twitter_username": "twitter_username-prefix",
	"twitter_hashtag":  "twitter_hashtag-prefix",
}

// TestFieldErrorsAreAssociatedWithInputs verifies that a field error is announced on the
// input it belongs to: the control is marked invalid and points at the element holding the
// message, so the message reaches people who never see it next to the field.
//
// [Ja] TestFieldErrorsAreAssociatedWithInputs は、フィールドのエラーが対象の入力欄で伝わる
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

			// The message is announced as an alert, and the field carries the invalid
			// state so its label reads as part of the error too.
			//
			// [Ja] メッセージは alert として通知され、フィールド自身も不正な状態を持つため
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
						t.Errorf("%s のエラーが描画されていません: %q", field, expected)
					}
				}
			}
		})
	}
}

// TestFieldErrorsAreSummarisedAtTheTopOfTheForm verifies that a failed submit is summarised
// above the form and that every error links to the input it belongs to. The forms run to
// around 25 fields, so an error near the bottom is off screen when the page comes back:
// without the summary the only way to find it is to scroll through the whole form.
//
// [Ja] TestFieldErrorsAreSummarisedAtTheTopOfTheForm は、送信の失敗がフォームの上に要約され、
// 各エラーが対象の入力欄へリンクされることを検証する。フォームは約 25 フィールドあり、下の方の
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

			// Every field the validator can fail is reachable from the summary, and each entry
			// names the field with the label the form shows for it.
			//
			// [Ja] バリデーターが落としうるすべてのフィールドへ要約から到達でき、各項目は
			// フォームが表示しているラベルでフィールドを名指しする。
			for _, field := range validatedFields {
				if !strings.Contains(html, fmt.Sprintf(`<a href="#%s"`, field)) {
					t.Errorf("%s のエラーが要約からリンクされていません", field)
				}
			}

			// The summary is the first thing focused after a failed submit, so the title field
			// must give up the autofocus it holds on an untouched form.
			//
			// [Ja] 送信が失敗した後に最初にフォーカスされるのは要約であるため、手つかずの
			// フォームで autofocus を持つタイトル欄はそれを譲る必要がある。
			if strings.Count(html, "autofocus") != 1 {
				t.Errorf("autofocus は要約だけが持つべきです\nHTML: %s", html)
			}
		})
	}
}

// TestFormWithoutErrorsFocusesTheTitle verifies that a form opened without errors puts the
// caret in the title field: with no summary to show, there is nothing else to focus.
//
// [Ja] TestFormWithoutErrorsFocusesTheTitle は、エラー無しで開いたフォームがタイトル欄に
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
				t.Errorf("タイトル欄が autofocus を持つべきです\nHTML: %s", html)
			}

			if strings.Contains(html, "入力内容にエラーがあります") {
				t.Errorf("エラーが無いときに要約を出してはいけません\nHTML: %s", html)
			}
		})
	}
}

// TestRequiredFieldsAreMarkedInWords verifies that the required fields carry the shared
// marker in both forms, and that no bare asterisk is left behind: the form has no legend
// explaining one, so the requirement has to be stated in words.
//
// [Ja] TestRequiredFieldsAreMarkedInWords は、両フォームの必須フィールドが共有の印を持ち、
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

			// Both required fields (title and media) are marked, and no other field is.
			//
			// [Ja] 必須フィールド (タイトルとメディア) の 2 つが印を持ち、他のフィールドは持たない。
			if got := strings.Count(html, "必須"); got != 2 {
				t.Errorf("必須の印の数 = %d, want 2", got)
			}
			if strings.Contains(html, `<span class="text-destructive">*</span>`) {
				t.Error("必須は素のアスタリスクではなく言葉で示すべきです")
			}
		})
	}
}

// TestSocialFieldsCarryPrefixAffix verifies that the X username and hashtag inputs show the
// sign they do not want typed as an affix inside the input, and that the affix reaches
// assistive technology through the input's description rather than being purely visual.
//
// [Ja] TestSocialFieldsCarryPrefixAffix は、X ユーザー名とハッシュタグの入力欄が、入力して
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

			// The affix replaces the instruction that used to live in the placeholder, and the
			// input-group shell draws the border, so the input keeps neither.
			//
			// [Ja] 接頭辞は placeholder にあった指示を置き換えるものであり、枠線は input-group の
			// シェルが描くため、入力欄はそのどちらも持たない。
			if strings.Contains(html, `name="twitter_username" class="input"`) {
				t.Error("input-group 内の入力欄は input クラスを持つべきではありません")
			}
			for _, unexpected := range []string{"@なし", "#なし"} {
				if strings.Contains(html, unexpected) {
					t.Errorf("接頭辞を表示するので placeholder の指示は残すべきではありません: %q", unexpected)
				}
			}
		})
	}
}

// TestNew_OmitsErrorAssociationWithoutErrors verifies that a form without errors marks no
// input invalid, so assistive technology is not told about errors that are not there.
//
// [Ja] TestNew_OmitsErrorAssociationWithoutErrors は、エラーの無いフォームがどの入力欄も
// 不正としてマークしないことを検証する。存在しないエラーを支援技術へ伝えないため。
func TestNew_OmitsErrorAssociationWithoutErrors(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	if err := New(NewPageData{FormInput: &viewmodel.DBWorkFormInput{}}).Render(context.Background(), &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	// Only the error association is checked: the prefix affixes describe their inputs
	// regardless of validity, so aria-describedby itself is expected on a clean form.
	//
	// [Ja] 検証対象はエラーの関連付けだけにする。接頭辞は入力の妥当性に関わらず入力欄を説明
	// するため、エラーの無いフォームにも aria-describedby は現れる。
	for _, unexpected := range []string{`aria-invalid`, `-error-1"`, "data-invalid"} {
		if strings.Contains(html, unexpected) {
			t.Errorf("エラーが無いとき描画されてはいけません: %q", unexpected)
		}
	}
}

// TestEdit_SidebarToggle verifies the edit form renders the sidebar toggle in its header.
//
// [Ja] TestEdit_SidebarToggle は編集フォームがヘッダーにサイドバートグルを描画する
// ことを検証する。
func TestEdit_SidebarToggle(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	if err := Edit(EditPageData{WorkID: 1, FormInput: &viewmodel.DBWorkFormInput{}}).Render(context.Background(), &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	html := buf.String()

	// The toggle is wired to the sidebar at every viewport size.
	//
	// [Ja] トグルはサイドバーに結線され、全画面幅で利用できる。
	for _, expected := range []string{`data-sidebar-toggle="db-sidebar"`} {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}
}

// TestEdit_HeadingAndWorkPageLink verifies that the edit page titles itself with the work
// being edited and offers the work's page on the public site as its action.
//
// [Ja] TestEdit_HeadingAndWorkPageLink は、編集画面が編集対象の作品でページの見出しを付け、
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
		// The action points at the public work page and opens it in a new tab with
		// tabnabbing protection, announcing that in its accessible name.
		//
		// [Ja] 操作は公開サイトの作品ページを指し、tabnabbing 対策付きで新しいタブに開く。
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

// TestEdit_HeadingFallsBackToPageTitle verifies that the heading falls back to the generic
// page title when the work has no display name, which is what the handler passes for a work
// with a blank title and for a submit that cleared the title field.
//
// [Ja] TestEdit_HeadingFallsBackToPageTitle は、作品に表示名が無いとき見出しが汎用の
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

// TestEdit_RendersSubnav verifies that the edit page wires the shared work subnav above
// the form and passes the work id through to the links.
//
// [Ja] TestEdit_RendersSubnav は、編集画面がフォームの上に共有の作品サブナビを配線し、
// 作品 ID をリンクまで通すことを検証する。
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

// TestEdit_SubnavOmitsEpisodeItemsWhenNoEpisodes verifies that the edit page forwards the
// no_episodes form value to the subnav so episode-derived entries are hidden.
//
// [Ja] TestEdit_SubnavOmitsEpisodeItemsWhenNoEpisodes は、編集画面が no_episodes の
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

// TestNew_OmitsSubnav verifies that the new page renders no subnav, since there is no work
// yet to navigate the sub-resources of.
//
// [Ja] TestNew_OmitsSubnav は、新規画面ではサブナビを描画しないことを検証する。まだ
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

// TestEdit_LabelExternalLinks verifies that the edit form renders external links through
// the shared sub-template too, confirming it is wired the same way as the new form.
//
// [Ja] TestEdit_LabelExternalLinks は、編集フォームでも共有サブテンプレート経由で外部リンクが
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

// TestEdit_DecorativeIconsAreHidden covers the two new-tab icons of the edit page. Both repeat
// what the link around them already announces, so neither reaches the accessibility tree or the
// focus order. The work-page link is a Basecoat text button, so its icon also declares the
// inline-end position that lets Basecoat apply the button's icon-aware spacing; the icon beside
// a form label sits in a plain link and takes no position.
//
// [Ja] TestEdit_DecorativeIconsAreHidden は編集ページの 2 つの新規タブアイコンを検証する。
// どちらも囲むリンクが既に伝えている内容を繰り返すため、アクセシビリティツリーにもフォーカス
// 順序にも出ない。作品ページへのリンクは Basecoat のテキスト付きボタンのため、そのアイコンは
// Basecoat がボタンのアイコン用間隔を適用できる inline-end の位置も宣言する。フォームのラベル
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
			wantHint: `inline-end の装飾アイコン`,
		},
		{
			name:     "フォームのラベル横のリンク",
			want:     decorativeIconMarkup(t, ctx, "arrow-square-out-regular", "w-[18px] h-[18px]"),
			wantHint: `位置を持たない装飾アイコン`,
		},
	}
	for _, tt := range tests {
		if !strings.Contains(html, tt.want) {
			t.Errorf("%s: %s が含まれていません", tt.name, tt.wantHint)
		}
	}

	if strings.Contains(html, iconWrapperMarkup) {
		t.Error("装飾アイコンはラッパー要素ではなく SVG 自体で隠すべきです")
	}
}

// workFormFieldGroups is the number of Field wrappers each work form renders. Counting them
// keeps a field added later from being written without the group, which the class alone does
// not reveal.
//
// [Ja] workFormFieldGroups は各作品フォームが描画する Field ラッパーの数。数を固定すること
// で、後から追加した欄がグループ無しで書かれることを防ぐ (クラスだけでは気付けないため)。
const workFormFieldGroups = 27

// workKeyboardHints lists the attributes every typed field of the work forms carries, in the
// order the forms show them. The last typed field asks for "done" because nothing after it is
// typed into: the remaining controls are a select and a checkbox, so the touch keyboard can
// close rather than offer to move on.
//
// The fields backed by a numeric column ask for a keypad while staying text inputs, which is
// what lets a rejected submit come back with whatever was typed (see the comment on the form).
//
// [Ja] workKeyboardHints は作品フォームの入力する欄が持つ属性を、フォームが表示する順で並べ
// る。最後の欄が "done" を求めるのは、その後ろに入力する欄が無いため (残るのはセレクトと
// チェックボックスで、タッチキーボードは次へ進むのではなく閉じてよい)。
//
// 数値カラムに対応する欄は text の入力欄のままキーパッドを要求する。これにより却下された送信
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

// workUntypedControls lists the controls of the work forms that take no typing. A keyboard
// hint on them would label a key the reader never reaches: a select and a checkbox open no
// keyboard, a date field opens a picker, and Enter in a textarea inserts a line break.
//
// [Ja] workUntypedControls は作品フォームのうち文字を入力しないコントロールを並べる。これらに
// キーボードヒントを付けても、読み手が触れないキーに札を付けることになる (セレクトと
// チェックボックスはキーボードを開かず、日付欄はピッカーを開き、textarea の Enter は改行を
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

// workControlHTML returns the markup of the opening tag of the form control carrying the given
// id, whichever element it is.
//
// [Ja] workControlHTML は指定した id を持つフォームコントロールの開始タグのマークアップを返す
// (要素の種類を問わない)。
func workControlHTML(t *testing.T, html string, id string) string {
	t.Helper()

	at := strings.Index(html, `id="`+id+`"`)
	if at < 0 {
		t.Fatalf("%q のコントロールが描画されていません", id)
	}
	start := strings.LastIndex(html[:at], "<")
	if start < 0 {
		t.Fatalf("%q のコントロールの開始タグが見つかりません", id)
	}
	end := strings.Index(html[at:], ">")
	if end < 0 {
		t.Fatalf("%q のコントロールが閉じられていません", id)
	}

	return html[start : at+end+1]
}

// assertWorkFormFields checks the field structure and the keyboard hints of one work form.
// Both forms show the same fields, so the expectations live in one place and each page states
// that it meets them.
//
// [Ja] assertWorkFormFields は作品フォーム 1 つの欄の構造とキーボードヒントを検証する。
// 両フォームは同じ欄を表示するため、期待値は 1 箇所に置き、各ページはそれを満たすことを表明
// する。
func assertWorkFormFields(t *testing.T, html string) {
	t.Helper()

	if got := strings.Count(html, `role="group" class="field"`); got != workFormFieldGroups {
		t.Errorf("Basecoat の field group = %d 個, want %d 個", got, workFormFieldGroups)
	}

	for _, hint := range workKeyboardHints {
		control := workControlHTML(t, html, hint.field)
		for _, attr := range hint.attrs {
			if !strings.Contains(control, attr) {
				t.Errorf("%q の入力欄に %s がありません: %s", hint.field, attr, control)
			}
		}
	}

	for _, id := range workUntypedControls {
		if control := workControlHTML(t, html, id); strings.Contains(control, "enterkeyhint") {
			t.Errorf("%q は文字を入力しないため enterkeyhint を持つべきではありません: %s", id, control)
		}
	}
}

// TestNew_FieldsCarryGroupsAndKeyboardHints covers the create form's field wrappers and the
// keyboard hints of its inputs.
//
// [Ja] TestNew_FieldsCarryGroupsAndKeyboardHints は新規作成フォームの欄のラッパーと入力欄の
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

// TestEdit_FieldsCarryGroupsAndKeyboardHints covers the edit form's field wrappers and the
// keyboard hints of its inputs.
//
// [Ja] TestEdit_FieldsCarryGroupsAndKeyboardHints は編集フォームの欄のラッパーと入力欄の
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

// TestEdit_CarriesVersion verifies the edit form ships the version it was opened against, which
// the update matches so a submit made from a stale read is refused instead of overwriting
// whoever wrote in between.
//
// [Ja] TestEdit_CarriesVersion は、編集フォームが開いた時点の版を送り出すことを検証する。更新側は
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
		t.Error("フォームが hidden の版を運んでいません")
	}
}

// TestEdit_ConflictNoticeListsChangedStoredValues verifies the conflict notice states the stored
// values a second submit would overwrite. Only the fields that differ are listed, and the values
// are the ones the form displays rather than the codes the selects and the checkbox store.
//
// [Ja] TestEdit_ConflictNoticeListsChangedStoredValues は、競合の案内が、2 回目の送信で上書き
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
		// The stored media is named by the label the select shows, not by its enum code.
		//
		// [Ja] 保存済みのメディアは enum のコードではなく、選択欄が表示するラベルで名指しする。
		"<dt>メディア</dt>",
		"映画",
		// The checkbox states whether it is on rather than showing the "1" it stores.
		//
		// [Ja] チェックボックスは保持する "1" ではなく、入っているかどうかを述べる。
		"<dt>エピソードなし</dt>",
		"オン",
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}

	// A field both sides agree on is not listed: the notice names what a second submit would
	// overwrite, and nothing about that field would change.
	//
	// [Ja] 両者が一致するフィールドは並べない。案内は 2 回目の送信が上書きするものを名指しする
	// ものであり、そのフィールドは何も変わらないため。
	if strings.Contains(html, "<dt>英語タイトル</dt>") {
		t.Error("一致するフィールドが競合の案内に並んでいます")
	}
}

// TestEdit_ConflictNoticeWithoutFieldChanges verifies the notice says so when the stored row
// differs from the submit in nothing this form writes, rather than presenting an empty list.
//
// [Ja] TestEdit_ConflictNoticeWithoutFieldChanges は、保存済みの行が本フォームの書き込む
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

// TestEdit_OmitsConflictNoticeWithoutConflict verifies a form opened for editing, and one
// re-rendered for a submit refused for any other reason, carries no conflict notice.
//
// [Ja] TestEdit_OmitsConflictNoticeWithoutConflict は、編集のために開いたフォームと、他の理由で
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

// TestGuidelineLinkBelowHeading verifies that both work forms offer the work editing
// guideline below their heading: an editor who opens a form reaches the guideline from the
// screen itself instead of having to find the help pages on their own.
//
// [Ja] TestGuidelineLinkBelowHeading は、両方の作品フォームが見出しの下に作品の編集ガイド
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
				// The link names where it goes and points at the work editing guideline,
				// opening it in a new tab with tabnabbing protection so the form the editor
				// is filling in stays open.
				//
				// [Ja] リンクは行き先を名乗り、作品の編集ガイドラインを指す。編集者が入力中の
				// フォームを開いたままにできるよう、tabnabbing 対策付きで新しいタブに開く。
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
