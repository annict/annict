package db_episodes

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/templates"
	"github.com/annict/annict/go/internal/viewmodel"
)

// editTestData returns the page data of an episode opened for editing, with every field
// filled in so a test can assert which value reaches which input.
//
// [Ja] editTestData は編集のために開いたエピソードのページデータを返す。どの値がどの入力欄へ
// 届くかをテストが検証できるよう、すべての欄を埋めている。
func editTestData() EditPageData {
	return EditPageData{
		EpisodeID: 5,
		WorkID:    1,
		WorkName:  "テストアニメ",
		CSRFToken: "test-csrf-token",
		FormInput: viewmodel.DBEpisodeFormInput{
			Number:     "第2話",
			RawNumber:  "2",
			SortNumber: "200",
			Title:      "もう、お婿にいけません",
			TitleEn:    "No Longer Marriageable",
			UpdatedAt:  "2026-08-12T09:30:15.123456Z",
		},
	}
}

func TestEdit_StoredValues(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	var buf strings.Builder
	if err := Edit(editTestData()).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	expectedContents := []string{
		// The heading names the work, and the shared subnav links back to its form.
		//
		// [Ja] 見出しは作品を名指しし、共有のサブナビはそのフォームへ戻るリンクを持つ。
		"テストアニメ",
		`href="/db/works/1/edit"`,
		// The form updates the episode itself, through the method override an HTML form
		// needs to send a PATCH.
		//
		// [Ja] フォームはエピソード自身を更新する。HTML フォームが PATCH を送るための
		// メソッドオーバーライドを通す。
		`action="/db/episodes/5"`,
		`method="POST"`,
		`name="_method" value="PATCH"`,
		`value="test-csrf-token"`,
		// The version the form was opened against is submitted alongside the values, so a
		// stale submit can be rejected rather than silently overwriting.
		//
		// [Ja] フォームを開いた時点の版が値と一緒に送信される。古い送信を黙って上書きせずに
		// 却下できるようにするため。
		`name="updated_at" value="2026-08-12T09:30:15.123456Z"`,
		// Every stored value opens in its own field.
		//
		// [Ja] 保存済みの値はそれぞれの欄に開く。
		`id="number"`,
		`value="第2話"`,
		`id="raw_number"`,
		`value="2"`,
		`id="sort_number"`,
		`value="200"`,
		// The two number fields ask for a numeric keypad on touch keyboards while staying
		// text inputs, so a rejected submit can be re-rendered with whatever was typed.
		//
		// [Ja] 2 つの話数の欄は、text の入力欄のままタッチキーボードに数字キーパッドを
		// 要求する。却下された送信を、入力された内容のまま再描画できるようにするため。
		`inputmode="decimal"`,
		`inputmode="numeric"`,
		`id="title"`,
		"もう、お婿にいけません",
		`id="title_en"`,
		"No Longer Marriageable",
		// A form opened for editing starts out focused on the first field, since there is no
		// error summary to take the caret.
		//
		// [Ja] 編集のために開いたフォームは先頭の欄にフォーカスして始まる。カーソルを受け取る
		// エラー要約が無いため。
		"autofocus",
		"一覧に戻る",
		"更新する",
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("レスポンスに %q が含まれていません", expected)
		}
	}

	if strings.Contains(html, `aria-invalid="true"`) {
		t.Error("エラーの無いフォームで aria-invalid が付いてはいけません")
	}

	if got := strings.Count(html, `role="group" class="field"`); got != 6 {
		t.Errorf("Basecoat の field group = %d 個, want 6 個", got)
	}

	keyboardHints := []struct {
		field string
		want  string
	}{
		{field: "number", want: `enterkeyhint="next"`},
		{field: "raw_number", want: `enterkeyhint="next"`},
		{field: "sort_number", want: `enterkeyhint="next"`},
		{field: "title", want: `enterkeyhint="next"`},
		{field: "title_en", want: `enterkeyhint="done"`},
	}
	for _, hint := range keyboardHints {
		if input := editInputHTML(t, html, hint.field); !strings.Contains(input, hint.want) {
			t.Errorf("%q の入力欄に %s がありません: %s", hint.field, hint.want, input)
		}
	}
}

// TestEdit_NumberFieldsStateTheirConventions covers the hints of the three number fields. Each
// follows a convention the value alone does not reveal and validation cannot check (the work's
// own wording, digits only, steps of 100), and the Rails page states all three in its sidebar,
// so the fields have to carry them here and name them from aria-describedby.
//
// [Ja] TestEdit_NumberFieldsStateTheirConventions は 3 つの話数の欄のヒントを検証する。いずれも
// 値そのものからは読み取れず、バリデーションでも検査できない作法 (作品ごとの表記・数字のみ・
// 100 刻み) を持ち、Rails のページはその 3 つをサイドバーで述べている。本ページでは欄が作法を
// 携え、aria-describedby から名指しする必要がある。
func TestEdit_NumberFieldsStateTheirConventions(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	var buf strings.Builder
	if err := Edit(editTestData()).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	hints := []struct {
		field  string
		hintID string
		text   string
	}{
		{field: "number", hintID: editNumberHintID, text: "その作品で使われている表記"},
		{field: "raw_number", hintID: editRawNumberHintID, text: "文字や記号を取り除いた数字"},
		{field: "sort_number", hintID: editSortNumberHintID, text: "話数 × 100"},
	}
	for _, hint := range hints {
		if !strings.Contains(html, `id="`+hint.hintID+`"`) {
			t.Errorf("%q のヒントが描画されていません", hint.field)
		}
		if !strings.Contains(html, hint.text) {
			t.Errorf("%q のヒントに %q が含まれていません", hint.field, hint.text)
		}

		want := `aria-describedby="` + hint.hintID + `"`
		if input := editInputHTML(t, html, hint.field); !strings.Contains(input, want) {
			t.Errorf("%q の入力欄が自身のヒントを指していません: %s", hint.field, input)
		}
	}
}

// TestEdit_MarksEpisodesSubnavEntry covers the subnav on a page the request path cannot place:
// the edit form is keyed by the episode, so without the page naming its section no entry would
// be marked and the reader would lose where they are.
//
// [Ja] TestEdit_MarksEpisodesSubnavEntry は、リクエストパスからは位置を決められないページの
// サブナビを検証する。編集フォームはエピソード基点のため、ページが所属する項目を名指ししないと
// どの項目にも印が付かず、読み手は現在地を見失う。
func TestEdit_MarksEpisodesSubnavEntry(t *testing.T) {
	t.Parallel()

	ctx := templates.SetCurrentPath(i18n.SetLocale(context.Background(), "ja"), "/db/episodes/5/edit")

	var buf strings.Builder
	if err := Edit(editTestData()).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	// The page links to the episode list twice (from the heading's action and from the
	// subnav), so the search starts at the nav to reach the subnav entry rather than the
	// back link.
	//
	// [Ja] ページはエピソード一覧へ 2 回リンクする (見出しの操作とサブナビ) ため、戻る
	// リンクではなくサブナビの項目に届くよう nav から探し始める。
	html := buf.String()
	navAt := strings.Index(html, "<nav ")
	if navAt < 0 {
		t.Fatal("サブナビが描画されていません")
	}

	start := strings.Index(html[navAt:], `<a href="/db/works/1/episodes"`)
	if start < 0 {
		t.Fatal("エピソード一覧へのサブナビ項目が描画されていません")
	}
	item := html[navAt+start:]
	end := strings.Index(item, "</a>")
	if end < 0 {
		t.Fatal("エピソード一覧へのサブナビ項目が閉じられていません")
	}
	if !strings.Contains(item[:end], `aria-current="page"`) {
		t.Error("エピソード編集ページでエピソードの項目に aria-current が付いていません")
	}
}

// TestEdit_EmptyValues covers an episode whose optional editable columns are unset: the fields
// open empty rather than with a placeholder the editor would have to clear before typing. A
// NULL updated_at remains an explicit version so the update can distinguish it from an absent
// precondition.
//
// [Ja] TestEdit_EmptyValues は編集できる任意カラムが未設定のエピソードを検証する。各欄は、
// 編集者が入力の前に消す必要のあるプレースホルダーではなく空で開く。NULL の updated_at は
// 明示的な版のままにし、更新側が前提条件の欠落と区別できるようにする。
func TestEdit_EmptyValues(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	data := editTestData()
	data.FormInput = viewmodel.DBEpisodeFormInput{
		SortNumber: "100",
		UpdatedAt:  viewmodel.FormNullVersion,
	}

	var buf strings.Builder
	if err := Edit(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()
	for _, field := range []string{"number", "raw_number", "title", "title_en"} {
		if input := editInputHTML(t, html, field); !strings.Contains(input, `value=""`) {
			t.Errorf("%q の欄が空で描画されていません: %s", field, input)
		}
	}
	if !strings.Contains(html, `name="updated_at" value="null"`) {
		t.Error("NULL の updated_at が明示的な版として描画されていません")
	}
}

// editInputHTML returns the markup of the form input carrying the given id.
//
// [Ja] editInputHTML は指定した id を持つフォーム入力欄のマークアップを返す。
func editInputHTML(t *testing.T, html string, id string) string {
	t.Helper()

	at := strings.Index(html, `id="`+id+`"`)
	if at < 0 {
		t.Fatalf("%q の入力欄が描画されていません", id)
	}
	start := strings.LastIndex(html[:at], "<input")
	if start < 0 {
		t.Fatalf("%q の入力欄の開始タグが見つかりません", id)
	}
	end := strings.Index(html[start:], ">")
	if end < 0 {
		t.Fatalf("%q の入力欄が閉じられていません", id)
	}

	return html[start : start+end+1]
}

// TestEdit_WithErrors covers the page after a rejected submit: the messages are summarised at
// the top, each rejected field is marked invalid and describes its own messages, and the
// submitted values are back in the fields.
//
// [Ja] TestEdit_WithErrors は送信が却下された後のページを検証する。メッセージが冒頭に要約され、
// 却下された各欄は不正と印付けられて自身のメッセージを説明として指し、送信された値は各欄に
// 戻る。
func TestEdit_WithErrors(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	formErrors := &viewmodel.FormErrors{
		Fields: map[string][]string{"sort_number": {"整数を入力してください"}},
	}

	data := editTestData()
	data.FormErrors = formErrors
	data.FormInput.SortNumber = "abc"

	var buf strings.Builder
	if err := Edit(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	expectedContents := []string{
		"整数を入力してください",
		// The summary names the field by its visible label and links to its input.
		//
		// [Ja] 要約はフィールドを可視ラベルで名指しし、その入力欄へリンクする。
		"並び順",
		`href="#sort_number"`,
		`aria-invalid="true"`,
		// The standing hint stays in describedby and comes before the error, so the
		// convention the value has to follow is not taken away while it is corrected.
		//
		// [Ja] 常設のヒントは describedby に残り、エラーより先に来る。値が従うべき作法を、
		// 直している最中に取り上げないため。
		`aria-describedby="sort_number-hint sort_number-error-1"`,
		`id="sort_number-error-1"`,
		// The rejected value is echoed back so the submit is corrected rather than retyped.
		//
		// [Ja] 却下された値はそのまま返し、送信は入力し直しではなく手直しで済むようにする。
		`value="abc"`,
		// The version is echoed back as the editor submitted it. Replacing it with the
		// server's current one would let the corrected submit overwrite a change written
		// while the first submit was being fixed.
		//
		// [Ja] 版は編集者が送ったものをそのまま返す。サーバーの現在値に差し替えると、最初の
		// 送信を直している間に書かれた変更を、手直し後の送信が上書きしてしまう。
		`name="updated_at" value="2026-08-12T09:30:15.123456Z"`,
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("レスポンスに %q が含まれていません\nHTML: %s", expected, html)
		}
	}

	// The error summary takes focus on load, so the first field gives up its autofocus and
	// the caret is not pulled past the summary that states why the submit failed.
	//
	// [Ja] エラー要約が読み込み時にフォーカスを受け取るため、先頭の欄は autofocus を手放す。
	// 送信の失敗理由を述べる要約を飛び越えてカーソルが移らないようにするため。
	if strings.Contains(editInputHTML(t, html, "number"), "autofocus") {
		t.Error("エラー要約があるフォームで先頭の欄に autofocus が付いてはいけません")
	}
}

// TestEdit_FallsBackToGenericHeading covers an episode whose work has no name to show: the
// heading falls back to the page name instead of rendering an empty <h1>.
//
// [Ja] TestEdit_FallsBackToGenericHeading は、表示できる名前が無い作品のエピソードを検証する。
// 見出しは空の <h1> を描画せず、ページ名へフォールバックする。
func TestEdit_FallsBackToGenericHeading(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	data := editTestData()
	data.WorkName = ""

	var buf strings.Builder
	if err := Edit(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	if !strings.Contains(buf.String(), "エピソードを編集") {
		t.Error("見出しが汎用のページタイトルへフォールバックしていません")
	}
}

// TestEdit_DecorativeIconsAreHidden covers the icons of the edit page that repeat adjacent
// visible text. The button and link already carry their meaning in text, so the SVGs stay out
// of the accessibility tree instead of adding a second, browser-dependent representation.
//
// [Ja] TestEdit_DecorativeIconsAreHidden は編集ページのアイコンのうち、隣接する可視テキストと
// 意味が重複するものを検証する。ボタンとリンクはテキストですでに意味を伝えるため、SVG は
// アクセシビリティツリーから除外し、ブラウザー依存の別表現を重ねないようにする。
func TestEdit_DecorativeIconsAreHidden(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	var buf strings.Builder
	if err := Edit(editTestData()).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	html := buf.String()

	for _, iconName := range []string{"floppy-disk-regular", "arrow-bend-up-left-regular"} {
		want := decorativeIconMarkup(t, ctx, iconName, "", templates.InlineIconStart)
		if !strings.Contains(html, want) {
			t.Errorf("装飾アイコン %q が aria-hidden の要素内にありません", iconName)
		}
	}
}

// TestEdit_ConflictShowsStoredValues covers a submit refused because someone else wrote the
// episode first: the stored values are shown beside the submitted ones so the editor can see
// what they would replace, and the options are stated in words rather than left to be guessed.
//
// [Ja] TestEdit_ConflictShowsStoredValues は、他者が先にエピソードを書いたために却下された送信を
// 検証する。保存済みの値が送信された値と並んで表示され、編集者が何を置き換えることになるのかを
// 見られる。取りうる対処も、推測に委ねず言葉で述べる。
func TestEdit_ConflictShowsStoredValues(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	data := editTestData()
	data.FormInput.Title = "後から届いたタイトル"
	data.FormErrors = &viewmodel.FormErrors{
		Global: []string{"このデータは他の編集者によって更新されました"},
	}
	data.ConflictCurrent = &viewmodel.DBEpisodeFormInput{
		Number:     "第3話",
		SortNumber: "300",
		Title:      "先に保存したタイトル",
		UpdatedAt:  "2026-08-12T10:00:00Z",
	}

	var buf strings.Builder
	if err := Edit(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()
	expectedContents := []string{
		"このデータは他の編集者によって更新されました",
		templates.T(ctx, "db_episodes_edit_conflict_heading"),
		templates.T(ctx, "db_episodes_edit_conflict_help"),
		"先に保存したタイトル",
		"第3話",
		// The submitted values stay in the form: the editor decides which of the two to keep,
		// and losing their input would leave only one of them to choose from.
		//
		// [Ja] 送信された値はフォームに残る。どちらを残すかを決めるのは編集者であり、入力が
		// 失われると選べる側が片方しか残らないため。
		"後から届いたタイトル",
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("レスポンスに %q が含まれていません", expected)
		}
	}

	// The stored values the notice lists include columns the episode leaves unset, which read
	// as "nothing recorded" rather than as a rendering gap.
	//
	// [Ja] 案内が並べる保存済みの値には、そのエピソードが未設定にしているカラムも含まれる。
	// これらは描画漏れではなく「未登録」と読める形にする。
	if !strings.Contains(html, missingValuePlaceholder) {
		t.Error("未設定の保存済みの値がプレースホルダーで描画されていません")
	}
}

// TestEdit_WithoutConflictOmitsStoredValues keeps the notice out of the ordinary edit page: a
// form opened for editing already shows the stored values in its own inputs, and repeating them
// would read as a second, competing set.
//
// [Ja] TestEdit_WithoutConflictOmitsStoredValues は通常の編集ページから案内を外していることを
// 検証する。編集のために開いたフォームは保存済みの値を各欄に既に表示しており、それを繰り返すと
// 対立するもう 1 組の値のように読めてしまうため。
func TestEdit_WithoutConflictOmitsStoredValues(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	var buf strings.Builder
	if err := Edit(editTestData()).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	if strings.Contains(buf.String(), templates.T(ctx, "db_episodes_edit_conflict_heading")) {
		t.Error("競合していないのに保存済みの内容の案内が描画されています")
	}
}
