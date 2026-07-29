package db_works

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/a-h/templ"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates"
	"github.com/annict/annict/go/internal/viewmodel"
)

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

// TestNew_NoLabelExternalLinksWhenEmpty verifies that no external link is rendered when
// the linkable fields are empty.
//
// [Ja] TestNew_NoLabelExternalLinksWhenEmpty は、値が空のとき外部リンクが描画されないことを
// テストする。
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
	if strings.Contains(html, "を新しいタブで開く") {
		t.Error("値が空のとき外部リンクは描画されてはいけません")
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

	formErrors := model.NewValidationError()
	for _, field := range validatedFields {
		formErrors.AddField(field, field+" のエラーメッセージ")
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

	formErrors := model.NewValidationError()
	for _, field := range validatedFields {
		formErrors.AddField(field, field+" のエラーメッセージ")
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
// page title when the work title is blank, which is what the form re-rendered for a blank
// title receives.
//
// [Ja] TestEdit_HeadingFallsBackToPageTitle は、作品タイトルが空のとき見出しが汎用の
// ページタイトルにフォールバックすることを検証する (タイトルを空にして再描画されたフォームが
// 受け取る値がこれにあたる)。
func TestEdit_HeadingFallsBackToPageTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		workTitle string
	}{
		{
			name:      "空文字",
			workTitle: "",
		},
		{
			name:      "空白文字のみ",
			workTitle: " \t\n ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := EditPageData{
				WorkID:    1,
				WorkTitle: tt.workTitle,
				FormInput: &viewmodel.DBWorkFormInput{},
			}

			var buf strings.Builder
			if err := Edit(data).Render(context.Background(), &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			if !strings.Contains(buf.String(), "作品を編集") {
				t.Error("作品タイトルが空のとき見出しはページタイトルを表示するべきです")
			}
		})
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
