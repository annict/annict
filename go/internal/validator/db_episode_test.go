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
