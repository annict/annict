package validator

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// episodeRowPtr builds the pointer fields of an expected row. The parsed row distinguishes
// an empty column (nil) from a filled one, so the expectations have to be pointers too.
//
// [Ja] episodeRowPtr は期待する行のポインタフィールドを組み立てる。パース結果は空の列 (nil)
// と入力された列を区別するため、期待値の側もポインタで書く必要がある。
func episodeRowPtr[T any](value T) *T {
	return &value
}

// formatEpisodeRow renders a parsed row for failure messages, spelling nil out instead of
// printing the pointer addresses %+v would show.
//
// [Ja] formatEpisodeRow は失敗メッセージ用にパース結果を文字列化する。%+v が出すポインタの
// アドレスではなく nil を明示する。
func formatEpisodeRow(row DBEpisodeRow) string {
	number := "nil"
	if row.Number != nil {
		number = fmt.Sprintf("%q", *row.Number)
	}
	rawNumber := "nil"
	if row.RawNumber != nil {
		rawNumber = fmt.Sprintf("%v", *row.RawNumber)
	}
	title := "nil"
	if row.Title != nil {
		title = fmt.Sprintf("%q", *row.Title)
	}

	return fmt.Sprintf("{Number: %s, RawNumber: %s, Title: %s}", number, rawNumber, title)
}

func assertEpisodeRowsEqual(t *testing.T, got, want []DBEpisodeRow) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("行数が想定と異なる: got %d, want %d", len(got), len(want))
	}

	for i := range want {
		if formatEpisodeRow(got[i]) != formatEpisodeRow(want[i]) {
			t.Errorf("%d 番目の行が想定と異なる: got %s, want %s", i, formatEpisodeRow(got[i]), formatEpisodeRow(want[i]))
		}
	}
}

func TestDBEpisodeCreateValidatorValidateSuccess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input DBEpisodeCreateValidatorInput
		want  []DBEpisodeRow
	}{
		{
			name:  "正常系: 3 列そろった 1 行",
			input: DBEpisodeCreateValidatorInput{Rows: "#1,1,教えてティーチャー"},
			want: []DBEpisodeRow{
				{Number: episodeRowPtr("#1"), RawNumber: episodeRowPtr(1.0), Title: episodeRowPtr("教えてティーチャー")},
			},
		},
		{
			name:  "正常系: 複数行を入力順に返す",
			input: DBEpisodeCreateValidatorInput{Rows: "#1,1,教えてティーチャー\n#2,2,もう、お婿にいけません"},
			want: []DBEpisodeRow{
				{Number: episodeRowPtr("#1"), RawNumber: episodeRowPtr(1.0), Title: episodeRowPtr("教えてティーチャー")},
				{Number: episodeRowPtr("#2"), RawNumber: episodeRowPtr(2.0), Title: episodeRowPtr("もう、お婿にいけません")},
			},
		},
		{
			name:  "正常系: フォーム送信の CRLF 改行",
			input: DBEpisodeCreateValidatorInput{Rows: "#1,1,教えてティーチャー\r\n#2,2,まずいよ☆先生\r\n"},
			want: []DBEpisodeRow{
				{Number: episodeRowPtr("#1"), RawNumber: episodeRowPtr(1.0), Title: episodeRowPtr("教えてティーチャー")},
				{Number: episodeRowPtr("#2"), RawNumber: episodeRowPtr(2.0), Title: episodeRowPtr("まずいよ☆先生")},
			},
		},
		{
			name:  "正常系: 空行を読み飛ばす",
			input: DBEpisodeCreateValidatorInput{Rows: "\n#1,1,教えてティーチャー\n   \n#2,2,まずいよ☆先生\n"},
			want: []DBEpisodeRow{
				{Number: episodeRowPtr("#1"), RawNumber: episodeRowPtr(1.0), Title: episodeRowPtr("教えてティーチャー")},
				{Number: episodeRowPtr("#2"), RawNumber: episodeRowPtr(2.0), Title: episodeRowPtr("まずいよ☆先生")},
			},
		},
		{
			name:  "正常系: 各列の前後の空白を取り除く",
			input: DBEpisodeCreateValidatorInput{Rows: " #1 , 1 , 教えてティーチャー "},
			want: []DBEpisodeRow{
				{Number: episodeRowPtr("#1"), RawNumber: episodeRowPtr(1.0), Title: episodeRowPtr("教えてティーチャー")},
			},
		},
		{
			name:  "正常系: タイトルに含まれるカンマを保つ",
			input: DBEpisodeCreateValidatorInput{Rows: "#1,1,教えて、ティーチャー,先生"},
			want: []DBEpisodeRow{
				{Number: episodeRowPtr("#1"), RawNumber: episodeRowPtr(1.0), Title: episodeRowPtr("教えて、ティーチャー,先生")},
			},
		},
		{
			name:  "正常系: タイトルのみ (話数の列は空)",
			input: DBEpisodeCreateValidatorInput{Rows: ",,双子が3人?"},
			want: []DBEpisodeRow{
				{Number: nil, RawNumber: nil, Title: episodeRowPtr("双子が3人?")},
			},
		},
		{
			name:  "正常系: 表示用話数のみ (タイトルの列は空)",
			input: DBEpisodeCreateValidatorInput{Rows: "#1,1,"},
			want: []DBEpisodeRow{
				{Number: episodeRowPtr("#1"), RawNumber: episodeRowPtr(1.0), Title: nil},
			},
		},
		{
			name:  "正常系: 末尾の列を省略した行",
			input: DBEpisodeCreateValidatorInput{Rows: "#1"},
			want: []DBEpisodeRow{
				{Number: episodeRowPtr("#1"), RawNumber: nil, Title: nil},
			},
		},
		{
			name:  "正常系: 小数の数値話数",
			input: DBEpisodeCreateValidatorInput{Rows: "#5.5,5.5,総集編"},
			want: []DBEpisodeRow{
				{Number: episodeRowPtr("#5.5"), RawNumber: episodeRowPtr(5.5), Title: episodeRowPtr("総集編")},
			},
		},
		{
			name:  "正常系: 上限ちょうどの長さの表示用話数とタイトル",
			input: DBEpisodeCreateValidatorInput{Rows: strings.Repeat("あ", 500) + ",1," + strings.Repeat("い", 500)},
			want: []DBEpisodeRow{
				{
					Number:    episodeRowPtr(strings.Repeat("あ", 500)),
					RawNumber: episodeRowPtr(1.0),
					Title:     episodeRowPtr(strings.Repeat("い", 500)),
				},
			},
		},
	}

	v := NewDBEpisodeCreateValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), "ja")

			rows, err := v.Validate(ctx, tt.input)
			if err != nil {
				t.Fatalf("想定外のエラー: %v", err)
			}

			assertEpisodeRowsEqual(t, rows, tt.want)
		})
	}
}

func TestDBEpisodeCreateValidatorValidateRowLimit(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	v := NewDBEpisodeCreateValidator()

	rowsAtLimit := strings.TrimSuffix(strings.Repeat("#1,1,はじまり\n", episodeCreateMaxRows), "\n")
	rows, err := v.Validate(ctx, DBEpisodeCreateValidatorInput{Rows: rowsAtLimit})
	if err != nil {
		t.Fatalf("上限ちょうどの Validate() error = %v", err)
	}
	if len(rows) != episodeCreateMaxRows {
		t.Errorf("上限ちょうどの行数 = %d, want %d", len(rows), episodeCreateMaxRows)
	}

	rowsOverLimit := rowsAtLimit + "\n#101,101,おわり"
	_, err = v.Validate(ctx, DBEpisodeCreateValidatorInput{Rows: rowsOverLimit})
	ve := model.AsValidationError(err)
	if ve == nil {
		t.Fatalf("上限超過の error = %v, want *model.ValidationError", err)
	}
	messages := ve.GetFieldErrors("rows")
	if len(messages) != 1 || !strings.Contains(messages[0], "100件以内") {
		t.Errorf("上限超過のエラー = %v, want 100件以内", messages)
	}
}

func TestDBEpisodeCreateValidatorValidateErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input DBEpisodeCreateValidatorInput
		// wantMessages are the substrings each reported error has to contain, in the order
		// the errors are reported.
		//
		// [Ja] wantMessages は報告された各エラーが含むべき部分文字列を、報告順に並べたもの。
		wantMessages []string
	}{
		{
			name:         "異常系: 未入力",
			input:        DBEpisodeCreateValidatorInput{Rows: ""},
			wantMessages: []string{"入力してください"},
		},
		{
			name:         "異常系: 空行だけの入力",
			input:        DBEpisodeCreateValidatorInput{Rows: "\n   \n\r\n"},
			wantMessages: []string{"入力してください"},
		},
		{
			name:         "異常系: 数値話数が数値ではない",
			input:        DBEpisodeCreateValidatorInput{Rows: "#1,いち,教えてティーチャー"},
			wantMessages: []string{"1 行目: 数値話数は数値で入力してください"},
		},
		{
			name:         "異常系: 数値話数が NaN",
			input:        DBEpisodeCreateValidatorInput{Rows: "#1,NaN,教えてティーチャー"},
			wantMessages: []string{"1 行目: 数値話数は数値で入力してください"},
		},
		{
			name:         "異常系: 数値話数が無限大",
			input:        DBEpisodeCreateValidatorInput{Rows: "#1,Inf,教えてティーチャー"},
			wantMessages: []string{"1 行目: 数値話数は数値で入力してください"},
		},
		{
			name:         "異常系: 表示用話数もタイトルも空",
			input:        DBEpisodeCreateValidatorInput{Rows: ",1,"},
			wantMessages: []string{"1 行目: 表示用話数かタイトルのいずれかを入力してください"},
		},
		{
			name:         "異常系: 表示用話数が長すぎる",
			input:        DBEpisodeCreateValidatorInput{Rows: strings.Repeat("あ", 501) + ",1,教えてティーチャー"},
			wantMessages: []string{"1 行目: 表示用話数は500文字以内で入力してください"},
		},
		{
			name:         "異常系: タイトルが長すぎる",
			input:        DBEpisodeCreateValidatorInput{Rows: "#1,1," + strings.Repeat("あ", 501)},
			wantMessages: []string{"1 行目: タイトルは500文字以内で入力してください"},
		},
		{
			name:  "異常系: 1 行の複数の問題をまとめて報告する",
			input: DBEpisodeCreateValidatorInput{Rows: "#1,いち," + strings.Repeat("あ", 501)},
			wantMessages: []string{
				"1 行目: タイトルは500文字以内で入力してください",
				"1 行目: 数値話数は数値で入力してください",
			},
		},
		{
			name:  "異常系: 空行を挟んでも行番号が入力どおりになる",
			input: DBEpisodeCreateValidatorInput{Rows: "#1,1,教えてティーチャー\n\n#3,さん,まずいよ☆先生\n#4,よん,もう、お婿にいけません"},
			wantMessages: []string{
				"3 行目: 数値話数は数値で入力してください",
				"4 行目: 数値話数は数値で入力してください",
			},
		},
	}

	v := NewDBEpisodeCreateValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), "ja")

			rows, err := v.Validate(ctx, tt.input)
			if rows != nil {
				t.Errorf("エラー時は行を返さないはず: got %v", rows)
			}

			ve := model.AsValidationError(err)
			if ve == nil {
				t.Fatalf("*model.ValidationError を期待したが得られなかった: %v", err)
			}

			messages := ve.GetFieldErrors("rows")
			if len(messages) != len(tt.wantMessages) {
				t.Fatalf("エラー件数が想定と異なる: got %d (%v), want %d", len(messages), messages, len(tt.wantMessages))
			}
			for i, want := range tt.wantMessages {
				if !strings.Contains(messages[i], want) {
					t.Errorf("%d 番目のエラーメッセージが想定と異なる: got %q, want to contain %q", i, messages[i], want)
				}
			}
		})
	}
}

// TestDBEpisodeCreateValidatorValidateRejectsWholeSubmit fixes that a single bad line fails
// the whole submit. The caller creates the rows in one transaction, so returning the good
// rows alongside the errors would describe a state it can never produce.
//
// [Ja] TestDBEpisodeCreateValidatorValidateRejectsWholeSubmit は 1 行でも不正なら送信全体が
// 失敗することを固定する。呼び出し元は行を 1 トランザクションで作成するため、エラーと一緒に
// 正常な行を返すと呼び出し元が作れない状態を表すことになる。
func TestDBEpisodeCreateValidatorValidateRejectsWholeSubmit(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	v := NewDBEpisodeCreateValidator()
	rows, err := v.Validate(ctx, DBEpisodeCreateValidatorInput{
		Rows: "#1,1,教えてティーチャー\n#2,に,まずいよ☆先生",
	})

	if rows != nil {
		t.Errorf("正常な行も返さないはず: got %v", rows)
	}
	if ve := model.AsValidationError(err); ve == nil {
		t.Fatalf("*model.ValidationError を期待したが得られなかった: %v", err)
	}
}

func TestDBEpisodeCreateValidatorValidateLocalizesMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		locale string
		want   string
	}{
		{locale: "ja", want: "2 行目: 数値話数は数値で入力してください"},
		{locale: "en", want: "Line 2: Please enter a number for the numeric number"},
	}

	v := NewDBEpisodeCreateValidator()

	for _, tt := range tests {
		t.Run(tt.locale, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), tt.locale)

			_, err := v.Validate(ctx, DBEpisodeCreateValidatorInput{
				Rows: "#1,1,教えてティーチャー\n#2,に,まずいよ☆先生",
			})

			ve := model.AsValidationError(err)
			if ve == nil {
				t.Fatalf("*model.ValidationError を期待したが得られなかった: %v", err)
			}

			messages := ve.GetFieldErrors("rows")
			if len(messages) != 1 {
				t.Fatalf("エラー件数が想定と異なる: got %d (%v), want 1", len(messages), messages)
			}
			if messages[0] != tt.want {
				t.Errorf("エラーメッセージが想定と異なる: got %q, want %q", messages[0], tt.want)
			}
		})
	}
}

// dbEpisodeUpdateInput returns a submit that passes every check, so each case below states only
// the field it is about.
//
// [Ja] dbEpisodeUpdateInput はすべての検査を通る送信を返す。以降の各ケースが、対象のフィールド
// だけを述べられるようにするため。
func dbEpisodeUpdateInput() DBEpisodeUpdateValidatorInput {
	return DBEpisodeUpdateValidatorInput{
		Number:     "第2話",
		RawNumber:  "2.5",
		SortNumber: "200",
		Title:      "もう、お婿にいけません",
		TitleEn:    "No Longer Marriageable",
		UpdatedAt:  "2026-08-14T12:34:56.123456789Z",
	}
}

func TestDBEpisodeUpdateValidatorValidateSuccess(t *testing.T) {
	t.Parallel()

	v := NewDBEpisodeUpdateValidator()
	ctx := i18n.SetLocale(context.Background(), "ja")

	t.Run("正常系: 送信された値を保存先のカラムの型へ変換する", func(t *testing.T) {
		t.Parallel()

		got, err := v.Validate(ctx, dbEpisodeUpdateInput())
		if err != nil {
			t.Fatalf("Validate() error = %v", err)
		}
		if got.Number == nil || *got.Number != "第2話" {
			t.Errorf("Number = %v, want %q", got.Number, "第2話")
		}
		if got.RawNumber == nil || *got.RawNumber != 2.5 {
			t.Errorf("RawNumber = %v, want 2.5", got.RawNumber)
		}
		if got.Title == nil || *got.Title != "もう、お婿にいけません" {
			t.Errorf("Title = %v, want %q", got.Title, "もう、お婿にいけません")
		}
		if got.TitleEn != "No Longer Marriageable" {
			t.Errorf("TitleEn = %q, want %q", got.TitleEn, "No Longer Marriageable")
		}
		if got.SortNumber != 200 {
			t.Errorf("SortNumber = %d, want 200", got.SortNumber)
		}
		if got.UpdatedAt == nil {
			t.Fatal("UpdatedAt = nil, want 送信された版")
		}
		if formatted := got.UpdatedAt.UTC().Format(DBEpisodeVersionLayout); formatted != "2026-08-14T12:34:56.123456789Z" {
			t.Errorf("UpdatedAt = %q, want %q", formatted, "2026-08-14T12:34:56.123456789Z")
		}
	})

	// A field the editor cleared has to reach the column as NULL: an empty string is a value no
	// existing row carries, and the list would render it as a filled-in but blank number.
	//
	// [Ja] 編集者が消したフィールドは NULL としてカラムに届く必要がある。空文字列は既存のどの行も
	// 持たない値で、一覧では「入力されているが空」の話数として描画されてしまう。
	t.Run("正常系: 空にした任意入力は nil になる", func(t *testing.T) {
		t.Parallel()

		input := dbEpisodeUpdateInput()
		input.Number = "  "
		input.RawNumber = ""
		input.Title = ""

		got, err := v.Validate(ctx, input)
		if err != nil {
			t.Fatalf("Validate() error = %v", err)
		}
		if got.Number != nil || got.RawNumber != nil || got.Title != nil {
			t.Errorf("(Number, RawNumber, Title) = (%v, %v, %v), want (nil, nil, nil)", got.Number, got.RawNumber, got.Title)
		}
	})

	// The NULL sentinel is an explicit version, so it has to be accepted and reach the update as
	// "match a row whose updated_at is NULL".
	//
	// [Ja] NULL のセンチネルは明示的な版のため、受理して「updated_at が NULL の行に一致させる」と
	// して更新側へ届く必要がある。
	t.Run("正常系: null のセンチネルは版なしとして受理される", func(t *testing.T) {
		t.Parallel()

		input := dbEpisodeUpdateInput()
		input.UpdatedAt = DBEpisodeNullVersion

		got, err := v.Validate(ctx, input)
		if err != nil {
			t.Fatalf("Validate() error = %v", err)
		}
		if got.UpdatedAt != nil {
			t.Errorf("UpdatedAt = %v, want nil", got.UpdatedAt)
		}
	})
}

func TestDBEpisodeUpdateValidatorValidateErrors(t *testing.T) {
	t.Parallel()

	v := NewDBEpisodeUpdateValidator()
	ctx := i18n.SetLocale(context.Background(), "ja")

	tests := []struct {
		name       string
		mutate     func(*DBEpisodeUpdateValidatorInput)
		wantField  string
		wantGlobal bool
	}{
		{
			name:      "異常系: 並び順が空",
			mutate:    func(in *DBEpisodeUpdateValidatorInput) { in.SortNumber = " " },
			wantField: "sort_number",
		},
		{
			name:      "異常系: 並び順が整数でない",
			mutate:    func(in *DBEpisodeUpdateValidatorInput) { in.SortNumber = "200.5" },
			wantField: "sort_number",
		},
		{
			name:      "異常系: 並び順が int32 を超える",
			mutate:    func(in *DBEpisodeUpdateValidatorInput) { in.SortNumber = "2147483648" },
			wantField: "sort_number",
		},
		{
			name:      "異常系: 数値話数が数値でない",
			mutate:    func(in *DBEpisodeUpdateValidatorInput) { in.RawNumber = "いち" },
			wantField: "raw_number",
		},
		{
			name:      "異常系: 数値話数が NaN",
			mutate:    func(in *DBEpisodeUpdateValidatorInput) { in.RawNumber = "NaN" },
			wantField: "raw_number",
		},
		{
			name:      "異常系: 表示用話数が長すぎる",
			mutate:    func(in *DBEpisodeUpdateValidatorInput) { in.Number = strings.Repeat("あ", 501) },
			wantField: "number",
		},
		{
			name:      "異常系: タイトルが長すぎる",
			mutate:    func(in *DBEpisodeUpdateValidatorInput) { in.Title = strings.Repeat("あ", 501) },
			wantField: "title",
		},
		{
			// An empty version means the submit stated none at all, which is not the same as
			// the NULL sentinel: accepting it would let a crafted request skip the check that
			// stops one editor from overwriting another.
			//
			// [Ja] 空の版は、送信が版をまったく示していないことを意味し、NULL のセンチネルとは
			// 別物である。受理すると、ある編集者が別の編集者を上書きするのを止める検査を、
			// 改変されたリクエストが素通りできてしまう。
			name:       "異常系: 版が空",
			mutate:     func(in *DBEpisodeUpdateValidatorInput) { in.UpdatedAt = "" },
			wantGlobal: true,
		},
		{
			name:       "異常系: 版が往復書式でない",
			mutate:     func(in *DBEpisodeUpdateValidatorInput) { in.UpdatedAt = "2026-08-14 12:34:56" },
			wantGlobal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			input := dbEpisodeUpdateInput()
			tt.mutate(&input)

			got, err := v.Validate(ctx, input)
			if got != nil {
				t.Errorf("Validate() = %+v, want nil", got)
			}
			ve := model.AsValidationError(err)
			if ve == nil {
				t.Fatalf("Validate() error = %v, want *model.ValidationError", err)
			}
			if tt.wantField != "" && !ve.HasFieldError(tt.wantField) {
				t.Errorf("フィールド %q のエラーがありません: %+v", tt.wantField, ve)
			}
			if tt.wantGlobal && len(ve.Global) == 0 {
				t.Errorf("グローバルエラーがありません: %+v", ve)
			}
		})
	}
}

// TestDBEpisodeUpdateValidatorValidateReportsEveryProblem covers a submit with several problems
// at once: every one is reported so the editor fixes them in a single pass instead of
// discovering them one submit at a time.
//
// [Ja] TestDBEpisodeUpdateValidatorValidateReportsEveryProblem は複数の問題を同時に含む送信を
// 検証する。すべてが報告されるため、編集者は送信するたびに 1 つずつ気付くのではなく一度に直せる。
func TestDBEpisodeUpdateValidatorValidateReportsEveryProblem(t *testing.T) {
	t.Parallel()

	v := NewDBEpisodeUpdateValidator()
	ctx := i18n.SetLocale(context.Background(), "ja")

	input := dbEpisodeUpdateInput()
	input.SortNumber = ""
	input.RawNumber = "いち"
	input.UpdatedAt = ""

	_, err := v.Validate(ctx, input)
	ve := model.AsValidationError(err)
	if ve == nil {
		t.Fatalf("Validate() error = %v, want *model.ValidationError", err)
	}
	if !ve.HasFieldError("sort_number") || !ve.HasFieldError("raw_number") || len(ve.Global) == 0 {
		t.Errorf("報告されたエラーが不足しています: %+v", ve)
	}
}
