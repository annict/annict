package db_people

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/viewmodel"
)

// editTestDataは編集のために開いた人物のページデータを返す。どの値がどの入力欄へ届くかを
// テストが検証できるよう、すべての欄を埋めている。
func editTestData() EditPageData {
	return EditPageData{
		Person:    viewmodel.DBPersonEditTarget{ID: 7, Name: "徳川家康"},
		CSRFToken: "test-csrf-token",
		FormOptions: viewmodel.DBPersonFormOptions{
			GenderOptions:     []viewmodel.SelectOption{{Value: "male", Label: "男性"}, {Value: "female", Label: "女性"}},
			BloodTypeOptions:  []viewmodel.SelectOption{{Value: "a", Label: "A型"}, {Value: "o", Label: "O型"}},
			PrefectureOptions: []viewmodel.SelectOption{{Value: "22", Label: "静岡県"}, {Value: "23", Label: "愛知県"}},
		},
		FormInput: viewmodel.DBPersonFormInput{
			Name:              "徳川家康",
			NameKana:          "とくがわいえやす",
			NameEn:            "Tokugawa Ieyasu",
			Nickname:          "狸",
			NicknameEn:        "Tanuki",
			Gender:            "male",
			BloodType:         "o",
			PrefectureID:      "23",
			Birthday:          "1543-01-31",
			Height:            "159",
			URL:               "https://example.com/ieyasu",
			URLEn:             "https://example.com/en/ieyasu",
			WikipediaURL:      "https://ja.wikipedia.org/wiki/ieyasu",
			WikipediaURLEn:    "https://en.wikipedia.org/wiki/ieyasu",
			TwitterUsername:   "ieyasu",
			TwitterUsernameEn: "ieyasu_en",
			UpdatedAt:         "2026-10-15T01:02:03.123456Z",
		},
	}
}

func renderEdit(t *testing.T, data EditPageData) string {
	t.Helper()

	ctx := i18n.SetLocale(context.Background(), "ja")

	var buf strings.Builder
	if err := Edit(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	return buf.String()
}

func TestEdit_StoredValues(t *testing.T) {
	t.Parallel()

	html := renderEdit(t, editTestData())

	expectedContents := []string{
		// 見出しは人物を名指しし、公開側の人物ページを新しいタブで開くリンクを持つ。
		"徳川家康",
		`href="/people/7"`,
		`target="_blank"`,
		`aria-label="人物ページを新しいタブで開く"`,
		// フォームは人物自身を更新する。HTMLフォームがPATCHを送るためのメソッド
		// オーバーライドを通す。
		`action="/db/people/7"`,
		`method="POST"`,
		`name="_method" value="PATCH"`,
		`value="test-csrf-token"`,
		// フォームを開いた時点の版が値と一緒に送信される。
		`name="updated_at" value="2026-10-15T01:02:03.123456Z"`,
		// 保存済みの値はそれぞれの欄に開く。
		`value="とくがわいえやす"`,
		`value="Tokugawa Ieyasu"`,
		`value="狸"`,
		`value="Tanuki"`,
		`<option value="male" selected>男性</option>`,
		`<option value="o" selected>O型</option>`,
		`<option value="23" selected>愛知県</option>`,
		`<option value="22">静岡県</option>`,
		// 入力欄は内容に合った型にする。
		`type="date" id="birthday" name="birthday"`,
		`value="1543-01-31"`,
		`inputmode="numeric"`,
		`value="159"`,
		`type="url"`,
		`value="https://example.com/ieyasu"`,
		`value="https://en.wikipedia.org/wiki/ieyasu"`,
		`value="ieyasu_en"`,
		"必須",
		"更新する",
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("レスポンスに%qが含まれていません", expected)
		}
	}

	// 名前は必須で、編集のために開いたフォームは名前の欄にフォーカスして始まる。
	nameInput := editInputHTML(t, html, "name")
	if !strings.Contains(nameInput, "required") || !strings.Contains(nameInput, "autofocus") {
		t.Errorf("名前の欄 = %s、期待値 = requiredとautofocusを持つ", nameInput)
	}

	// 16項目すべての欄がラベルと対になっている。
	for _, field := range []string{
		"name", "name_kana", "name_en", "nickname", "nickname_en", "gender", "blood_type", "prefecture_id",
		"birthday", "height", "url", "url_en", "wikipedia_url", "wikipedia_url_en", "twitter_username", "twitter_username_en",
	} {
		if !strings.Contains(html, `for="`+field+`"`) || !strings.Contains(html, `id="`+field+`"`) {
			t.Errorf("フィールド%qのラベルまたは入力欄がありません", field)
		}
	}

	if strings.Contains(html, "現在保存されている内容") {
		t.Error("競合していないフォームに保存済みの内容の案内が表示されています")
	}
}

// TestEdit_WithErrorsは送信が却下された後のページを検証する。メッセージが冒頭に要約され、
// 却下された欄は不正と印付けられて自身のメッセージを説明として指し、送信された値は欄に戻る。
func TestEdit_WithErrors(t *testing.T) {
	t.Parallel()

	data := editTestData()
	data.FormErrors = &viewmodel.FormErrors{
		Fields: map[string][]string{"height": {"整数で入力してください"}},
	}
	data.FormInput.Height = "高い"

	html := renderEdit(t, data)

	expectedContents := []string{
		"整数で入力してください",
		// 要約はフィールドを可視ラベルで名指しし、その入力欄へリンクする。
		"身長 (cm)",
		`href="#height"`,
		`aria-invalid="true"`,
		`aria-describedby="height-error-1"`,
		`id="height-error-1"`,
		// 却下された値はそのまま返す。
		`value="高い"`,
		// 版は編集者が送ったものをそのまま返す。
		`name="updated_at" value="2026-10-15T01:02:03.123456Z"`,
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("レスポンスに%qが含まれていません", expected)
		}
	}

	// エラー要約が読み込み時にフォーカスを受け取るため、名前の欄はautofocusを手放す。
	if strings.Contains(editInputHTML(t, html, "name"), "autofocus") {
		t.Error("エラー要約があるフォームで名前の欄にautofocusが付いてはいけません")
	}
}

// editInputHTMLはidで指定した入力欄の開始タグを返す。
func editInputHTML(t *testing.T, html string, id string) string {
	t.Helper()

	at := strings.Index(html, `id="`+id+`"`)
	if at < 0 {
		t.Fatalf("%qの入力欄が描画されていません", id)
	}
	start := strings.LastIndex(html[:at], "<input")
	if start < 0 {
		t.Fatalf("%qの入力欄の開始タグが見つかりません", id)
	}
	end := strings.Index(html[start:], ">")
	if end < 0 {
		t.Fatalf("%qの入力欄が閉じられていません", id)
	}

	return html[start : start+end+1]
}

// TestEdit_ConflictShowsDifferingStoredValuesは、競合で返ったページが、送信された値と
// 異なる保存済みの値だけを、選択欄は選択肢のラベルに解決して表示することを検証する。
func TestEdit_ConflictShowsDifferingStoredValues(t *testing.T) {
	t.Parallel()

	data := editTestData()
	stored := data.FormInput
	stored.NameEn = "Ieyasu Tokugawa"
	stored.PrefectureID = "22"
	data.ConflictCurrent = &stored

	html := renderEdit(t, data)

	expectedContents := []string{
		"現在保存されている内容",
		"<dt>名前 (英語)</dt>",
		"Ieyasu Tokugawa",
		"<dt>出身地</dt>",
		// 選択欄の保存済みの値は、コードではなく編集者が目にするラベルで示す。
		">静岡県</dd>",
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("レスポンスに%qが含まれていません", expected)
		}
	}

	// 送信された値と一致する保存済みの値は並べない。
	if strings.Contains(html, "<dt>名前 (かな)</dt>") {
		t.Error("送信された値と一致する保存済みの値が案内に含まれています")
	}
}

// TestEdit_ConflictWithoutFieldChangesは、フォームの項目に違いが無い競合で、違いが無い
// ことを述べることを検証する。
func TestEdit_ConflictWithoutFieldChanges(t *testing.T) {
	t.Parallel()

	data := editTestData()
	stored := data.FormInput
	data.ConflictCurrent = &stored

	html := renderEdit(t, data)

	if !strings.Contains(html, "このフォームで編集できる項目に違いはありません") {
		t.Error("項目に違いが無いことの説明がありません")
	}
}
