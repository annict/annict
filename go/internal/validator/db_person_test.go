package validator

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// newTestDBPersonValidatorは共有テストDBに対してバリデーターを組み立てる。名前の重複
// 確認がpeopleを読むため、全ケースでリポジトリが要る。重複確認に見つけさせたい人物を
// テストが用意できるよう、トランザクションも返す。
func newTestDBPersonValidator(t *testing.T) (*DBPersonCreateValidator, *sql.Tx) {
	t.Helper()

	db, tx := testutil.SetupTx(t)
	return NewDBPersonCreateValidator(repository.NewPersonRepository(query.New(db).WithTx(tx))), tx
}

// uniquePersonNameは他のテストがコミットした人物と重ならない名前を返す。
func uniquePersonName(label string) string {
	return label + "-" + uuid.New().String()
}

func TestDBPersonCreateValidatorValidateSuccess(t *testing.T) {
	t.Parallel()

	ieyasu := uniquePersonName("徳川家康")
	nobunaga := uniquePersonName("織田信長")
	nickname := uniquePersonName(`野中"まさ"雄一`)

	tests := []struct {
		name string
		rows string
		want []DBPersonRow
	}{
		{
			name: "名前と読み",
			rows: ieyasu + ",とくがわいえやす",
			want: []DBPersonRow{{Name: ieyasu, NameKana: "とくがわいえやす"}},
		},
		{
			name: "読みを省略した行は読みを空文字列にする",
			rows: ieyasu + "\n" + nobunaga + ",",
			want: []DBPersonRow{{Name: ieyasu}, {Name: nobunaga}},
		},
		{
			name: "前後の空白を除き、空行を読み飛ばす",
			rows: "  " + ieyasu + " , とくがわいえやす \n\n   \n" + nobunaga + ",おだのぶなが",
			want: []DBPersonRow{
				{Name: ieyasu, NameKana: "とくがわいえやす"},
				{Name: nobunaga, NameKana: "おだのぶなが"},
			},
		},
		{
			name: "CRLFとCRの改行を分ける",
			rows: ieyasu + ",とくがわいえやす\r\n" + nobunaga + ",おだのぶなが\r",
			want: []DBPersonRow{
				{Name: ieyasu, NameKana: "とくがわいえやす"},
				{Name: nobunaga, NameKana: "おだのぶなが"},
			},
		},
		{
			name: "ダブルクォートは文字としてそのまま残す",
			rows: nickname + `,"のなか"`,
			want: []DBPersonRow{{Name: nickname, NameKana: `"のなか"`}},
		},
		{
			name: "3列目以降が空の列は許す",
			rows: ieyasu + ",とくがわいえやす,,",
			want: []DBPersonRow{{Name: ieyasu, NameKana: "とくがわいえやす"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v, _ := newTestDBPersonValidator(t)
			ctx := i18n.SetLocale(context.Background(), "ja")

			rows, err := v.Validate(ctx, DBPersonCreateValidatorInput{Rows: tt.rows})
			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}

			if len(rows) != len(tt.want) {
				t.Fatalf("行数 = %d (%+v)、期待値 = %d", len(rows), rows, len(tt.want))
			}
			for i := range tt.want {
				if rows[i] != tt.want[i] {
					t.Errorf("行[%d] = %+v、期待値 = %+v", i, rows[i], tt.want[i])
				}
			}
		})
	}
}

func TestDBPersonCreateValidatorValidateErrors(t *testing.T) {
	t.Parallel()

	ieyasu := uniquePersonName("徳川家康")
	nobunaga := uniquePersonName("織田信長")

	tests := []struct {
		name string
		rows string
		want []string
	}{
		{
			name: "空の送信",
			rows: "",
			want: []string{"入力してください"},
		},
		{
			name: "空行だけの送信",
			rows: "\n  \n",
			want: []string{"入力してください"},
		},
		{
			name: "名前の無い行",
			rows: ieyasu + "\n,よみだけ",
			want: []string{"2 行目: 名前を入力してください"},
		},
		{
			name: "3列目以降に値のある行",
			rows: ieyasu + ",とくがわいえやす,Tokugawa Ieyasu",
			want: []string{"1 行目「" + ieyasu + "」: 入力できるのは名前と読みの2列までです"},
		},
		{
			name: "入力の中で名前が重複する行は後の行に報告する",
			rows: ieyasu + "\n" + nobunaga + "\n\n" + ieyasu + ",とくがわ",
			want: []string{"4 行目「" + ieyasu + "」: 同じ名前が1 行目にもあります"},
		},
		{
			name: "1行に複数の問題があればすべて報告し、行の順に並べる。名前の無い行には名前を添えない",
			rows: ",,余分な列\n" + ieyasu + ",,余分な列",
			want: []string{
				"1 行目: 名前を入力してください",
				"1 行目: 入力できるのは名前と読みの2列までです",
				"2 行目「" + ieyasu + "」: 入力できるのは名前と読みの2列までです",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v, _ := newTestDBPersonValidator(t)
			ctx := i18n.SetLocale(context.Background(), "ja")

			rows, err := v.Validate(ctx, DBPersonCreateValidatorInput{Rows: tt.rows})
			if rows != nil {
				t.Errorf("正常な行 = %v、期待値 = nil (1行でも不正なら送信全体を失敗させるため)", rows)
			}

			assertPersonRowsErrors(t, err, tt.want)
		})
	}
}

// TestDBPersonCreateValidatorValidateTakenNamesは、既存の人物と名前が重なる行を、
// 一意インデックスと同じくアーカイブ・掲載停止の人物も含めて報告し、削除済みの人物の名前は
// 報告しないことを検証する。掲載停止の人物との重複は、通常の重複と別の文言で報告する。
func TestDBPersonCreateValidatorValidateTakenNames(t *testing.T) {
	t.Parallel()

	v, tx := newTestDBPersonValidator(t)
	ctx := i18n.SetLocale(context.Background(), "ja")

	published := uniquePersonName("公開中の人物")
	archived := uniquePersonName("アーカイブの人物")
	delisted := uniquePersonName("掲載停止の人物")
	deleted := uniquePersonName("削除済みの人物")
	unused := uniquePersonName("未登録の人物")
	testutil.NewPersonBuilder(t, tx).WithName(published).Build()
	testutil.NewPersonBuilder(t, tx).WithName(archived).WithStatus(model.PersonStatusArchived).Build()
	testutil.NewPersonBuilder(t, tx).WithName(delisted).WithStatus(model.PersonStatusDelisted).Build()
	testutil.NewPersonBuilder(t, tx).WithName(deleted).WithStatus(model.PersonStatusDeleted).Build()

	// 既存の人物と重なる名前を入力の中でも重複させ、既存の人物との重複は最初の行にだけ、
	// 入力の中での重複は後の行にだけ報告することも確かめる。
	_, err := v.Validate(ctx, DBPersonCreateValidatorInput{
		Rows: strings.Join([]string{published, archived, unused, delisted, deleted, published}, "\n"),
	})

	assertPersonRowsErrors(t, err, []string{
		"1 行目「" + published + "」: この名前の人物は既に登録されています",
		"2 行目「" + archived + "」: この名前の人物は既に登録されています",
		"4 行目「" + delisted + "」: この人物は掲載を停止しています。関係者に関わる事情のため、停止の詳細はお伝えできません。ご理解いただけますと幸いです。",
		"6 行目「" + published + "」: 同じ名前が1 行目にもあります",
	})
}

// TestDBPersonCreateValidatorValidateTakenNamesAfterMalformedRowは、書式エラーのある行を
// 既存名の照合と入力内の重複判定から外し、後続の正常な行を照合することを検証する。
func TestDBPersonCreateValidatorValidateTakenNamesAfterMalformedRow(t *testing.T) {
	t.Parallel()

	v, tx := newTestDBPersonValidator(t)
	ctx := i18n.SetLocale(context.Background(), "ja")
	name := uniquePersonName("既存の人物")
	testutil.NewPersonBuilder(t, tx).WithName(name).Build()

	_, err := v.Validate(ctx, DBPersonCreateValidatorInput{
		Rows: name + ",よみ,余分な列\n" + name + ",よみ",
	})

	assertPersonRowsErrors(t, err, []string{
		"1 行目「" + name + "」: 入力できるのは名前と読みの2列までです",
		"2 行目「" + name + "」: この名前の人物は既に登録されています",
	})
}

func TestDBPersonCreateValidatorValidateRowLimit(t *testing.T) {
	t.Parallel()

	v, _ := newTestDBPersonValidator(t)
	ctx := i18n.SetLocale(context.Background(), "ja")

	lines := make([]string, 0, personCreateMaxRows+1)
	prefix := uniquePersonName("上限テストの人物")
	for i := range personCreateMaxRows {
		lines = append(lines, fmt.Sprintf("%s-%d", prefix, i))
	}

	// 上限ちょうどは受け付ける。空行は数えない。
	rows, err := v.Validate(ctx, DBPersonCreateValidatorInput{Rows: strings.Join(lines, "\n\n")})
	if err != nil {
		t.Fatalf("上限ちょうどの送信でエラー: %v", err)
	}
	if len(rows) != personCreateMaxRows {
		t.Errorf("行数 = %d、期待値 = %d", len(rows), personCreateMaxRows)
	}

	lines = append(lines, prefix+"-over")
	_, err = v.Validate(ctx, DBPersonCreateValidatorInput{Rows: strings.Join(lines, "\n")})
	assertPersonRowsErrors(t, err, []string{"一度に登録する人物は100件以内にしてください"})
}

func TestDBPersonCreateValidatorValidateLocalizesMessages(t *testing.T) {
	t.Parallel()

	ieyasu := uniquePersonName("徳川家康")

	tests := []struct {
		locale string
		want   []string
	}{
		{
			locale: "ja",
			want: []string{
				"1 行目「" + ieyasu + "」: 入力できるのは名前と読みの2列までです",
				"2 行目: 名前を入力してください",
			},
		},
		{
			locale: "en",
			want: []string{
				`Line 1 ("` + ieyasu + `"): Only the name and its reading (2 columns) can be entered`,
				"Line 2: Enter a name",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.locale, func(t *testing.T) {
			t.Parallel()

			v, _ := newTestDBPersonValidator(t)
			ctx := i18n.SetLocale(context.Background(), tt.locale)

			_, err := v.Validate(ctx, DBPersonCreateValidatorInput{Rows: ieyasu + ",よみ,余分な列\n,よみだけ"})

			assertPersonRowsErrors(t, err, tt.want)
		})
	}
}

func assertPersonRowsErrors(t *testing.T, err error, want []string) {
	t.Helper()

	ve := model.AsValidationError(err)
	if ve == nil {
		t.Fatalf("*model.ValidationErrorを期待したが得られなかった: %v", err)
	}

	messages := ve.GetFieldErrors("rows")
	if len(messages) != len(want) {
		t.Fatalf("エラー = %q、期待値 = %q", messages, want)
	}
	for i := range want {
		if messages[i] != want[i] {
			t.Errorf("エラー[%d] = %q、期待値 = %q", i, messages[i], want[i])
		}
	}
}

// newTestDBPersonUpdateValidatorは共有テストDBに対して編集のバリデーターを組み立てる。
// 名前の重複確認と出身地の存在確認のための行をテストが用意できるよう、トランザクションも返す。
func newTestDBPersonUpdateValidator(t *testing.T) (*DBPersonUpdateValidator, *sql.Tx) {
	t.Helper()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	return NewDBPersonUpdateValidator(
		repository.NewPersonRepository(queries),
		repository.NewPrefectureRepository(queries),
	), tx
}

// personTestFormVersionは編集の送信が運ぶ版。
const personTestFormVersion = "2026-10-15T01:02:03.123456Z"

// dbPersonUpdateInputは名前と版だけを持つ、受理される最小の送信を返す。
func dbPersonUpdateInput(name string) DBPersonUpdateValidatorInput {
	return DBPersonUpdateValidatorInput{
		Name:      name,
		UpdatedAt: personTestFormVersion,
	}
}

func TestDBPersonUpdateValidatorValidateFormatSuccess(t *testing.T) {
	t.Parallel()

	v, _ := newTestDBPersonUpdateValidator(t)
	ctx := context.Background()

	t.Run("すべての項目を型に変換する", func(t *testing.T) {
		input := DBPersonUpdateValidatorInput{
			Name:              "  名前  ",
			NameKana:          " よみ ",
			NameEn:            "Name",
			Nickname:          " あだな ",
			NicknameEn:        "Nick",
			Gender:            "male",
			BloodType:         "o",
			PrefectureID:      "13",
			Birthday:          "1985-12-31",
			Height:            "172",
			URL:               "https://example.com",
			URLEn:             "https://example.com/en",
			WikipediaURL:      "https://ja.wikipedia.org/wiki/a",
			WikipediaURLEn:    "https://en.wikipedia.org/wiki/a",
			TwitterUsername:   "x_user",
			TwitterUsernameEn: "x_user_en",
			UpdatedAt:         personTestFormVersion,
		}

		got, err := v.ValidateFormat(ctx, input)
		if err != nil {
			t.Fatalf("ValidateFormat()のエラー = %v", err)
		}

		if got.Name != "名前" || got.NameKana != "よみ" || got.NameEn != "Name" {
			t.Errorf("名前 = %q / %q / %q", got.Name, got.NameKana, got.NameEn)
		}
		if got.Nickname == nil || *got.Nickname != "あだな" || got.NicknameEn != "Nick" {
			t.Errorf("ニックネーム = %v / %q", got.Nickname, got.NicknameEn)
		}
		if got.Gender == nil || *got.Gender != model.PersonGenderMale {
			t.Errorf("Gender = %v、期待値 = male", got.Gender)
		}
		if got.BloodType == nil || *got.BloodType != model.PersonBloodTypeO {
			t.Errorf("BloodType = %v、期待値 = o", got.BloodType)
		}
		if got.PrefectureID == nil || *got.PrefectureID != 13 {
			t.Errorf("PrefectureID = %v、期待値 = 13", got.PrefectureID)
		}
		if got.Birthday == nil || !got.Birthday.Equal(time.Date(1985, 12, 31, 0, 0, 0, 0, time.UTC)) {
			t.Errorf("Birthday = %v、期待値 = 1985-12-31", got.Birthday)
		}
		if got.Height == nil || *got.Height != 172 {
			t.Errorf("Height = %v、期待値 = 172", got.Height)
		}
		if got.URL == nil || *got.URL != "https://example.com" || got.URLEn != "https://example.com/en" {
			t.Errorf("URL = %v / %q", got.URL, got.URLEn)
		}
		if got.TwitterUsername == nil || *got.TwitterUsername != "x_user" || got.TwitterUsernameEn != "x_user_en" {
			t.Errorf("TwitterUsername = %v / %q", got.TwitterUsername, got.TwitterUsernameEn)
		}
		wantVersion := time.Date(2026, 10, 15, 1, 2, 3, 123456000, time.UTC)
		if !got.UpdatedAt.Equal(wantVersion) {
			t.Errorf("UpdatedAt = %v、期待値 = %v", got.UpdatedAt, wantVersion)
		}
	})

	t.Run("空欄はNULL許容のカラムでnil、NOT NULLのカラムで空文字列にする", func(t *testing.T) {
		got, err := v.ValidateFormat(ctx, dbPersonUpdateInput("名前"))
		if err != nil {
			t.Fatalf("ValidateFormat()のエラー = %v", err)
		}

		if got.Nickname != nil || got.Gender != nil || got.BloodType != nil || got.PrefectureID != nil ||
			got.Birthday != nil || got.Height != nil || got.URL != nil || got.WikipediaURL != nil || got.TwitterUsername != nil {
			t.Errorf("NULL許容のカラムにnilでない値があります: %+v", got)
		}
		if got.NameKana != "" || got.NameEn != "" || got.NicknameEn != "" || got.URLEn != "" ||
			got.WikipediaURLEn != "" || got.TwitterUsernameEn != "" {
			t.Errorf("NOT NULLのカラムに空文字列でない値があります: %+v", got)
		}
	})
}

func TestDBPersonUpdateValidatorValidateFormatErrors(t *testing.T) {
	t.Parallel()

	v, _ := newTestDBPersonUpdateValidator(t)
	ctx := i18n.SetLocale(context.Background(), "ja")

	tests := []struct {
		name       string
		mutate     func(*DBPersonUpdateValidatorInput)
		wantField  string
		wantGlobal bool
		wantMsg    string
	}{
		{
			name:      "異常系: 名前が空",
			mutate:    func(in *DBPersonUpdateValidatorInput) { in.Name = "  " },
			wantField: "name",
			wantMsg:   "入力してください",
		},
		{
			name:      "異常系: 性別が許可リストに無い",
			mutate:    func(in *DBPersonUpdateValidatorInput) { in.Gender = "other" },
			wantField: "gender",
			wantMsg:   "正しい性別を選択してください",
		},
		{
			name:      "異常系: 血液型が許可リストに無い",
			mutate:    func(in *DBPersonUpdateValidatorInput) { in.BloodType = "c" },
			wantField: "blood_type",
			wantMsg:   "正しい血液型を選択してください",
		},
		{
			name:      "異常系: 出身地が整数でない",
			mutate:    func(in *DBPersonUpdateValidatorInput) { in.PrefectureID = "tokyo" },
			wantField: "prefecture_id",
		},
		{
			// Rails版は誕生日の解析に失敗すると500になっていた。
			name:      "異常系: 誕生日が日付でない",
			mutate:    func(in *DBPersonUpdateValidatorInput) { in.Birthday = "1990-02-30" },
			wantField: "birthday",
			wantMsg:   "正しい日付を入力してください",
		},
		{
			name:      "異常系: 身長が整数でない",
			mutate:    func(in *DBPersonUpdateValidatorInput) { in.Height = "170.5" },
			wantField: "height",
			wantMsg:   "整数で入力してください",
		},
		{
			name:      "異常系: 身長がint32を超える",
			mutate:    func(in *DBPersonUpdateValidatorInput) { in.Height = "2147483648" },
			wantField: "height",
		},
		{
			name:      "異常系: URLがホストを持たない",
			mutate:    func(in *DBPersonUpdateValidatorInput) { in.URL = "https://" },
			wantField: "url",
			wantMsg:   "URLの形式が正しくありません",
		},
		{
			name:      "異常系: 英語のURLが形式不正",
			mutate:    func(in *DBPersonUpdateValidatorInput) { in.URLEn = "example.com" },
			wantField: "url_en",
		},
		{
			name:      "異常系: WikipediaのURLが形式不正",
			mutate:    func(in *DBPersonUpdateValidatorInput) { in.WikipediaURL = "wikipedia" },
			wantField: "wikipedia_url",
		},
		{
			name:      "異常系: 英語のWikipediaのURLが形式不正",
			mutate:    func(in *DBPersonUpdateValidatorInput) { in.WikipediaURLEn = "wikipedia" },
			wantField: "wikipedia_url_en",
		},
		{
			name:       "異常系: 版が空",
			mutate:     func(in *DBPersonUpdateValidatorInput) { in.UpdatedAt = "" },
			wantGlobal: true,
		},
		{
			// people.updated_atはNOT NULLのため、NULLの版を名乗る送信はありえない。
			name:       "異常系: 版がNULLのセンチネル",
			mutate:     func(in *DBPersonUpdateValidatorInput) { in.UpdatedAt = FormNullVersion },
			wantGlobal: true,
		},
		{
			name:       "異常系: 版が往復書式でない",
			mutate:     func(in *DBPersonUpdateValidatorInput) { in.UpdatedAt = "2026-10-15 01:02:03" },
			wantGlobal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := dbPersonUpdateInput("名前")
			tt.mutate(&input)

			got, err := v.ValidateFormat(ctx, input)
			if got != nil {
				t.Errorf("ValidateFormat() = %+v、期待値 = nil", got)
			}
			assertPersonUpdateValidationError(t, err, tt.wantField, tt.wantMsg, tt.wantGlobal)
		})
	}
}

func TestDBPersonUpdateValidatorValidateDatabase(t *testing.T) {
	t.Parallel()

	v, tx := newTestDBPersonUpdateValidator(t)
	ctx := i18n.SetLocale(context.Background(), "ja")
	prefectureID := testutil.CreateTestPrefecture(t, tx, "京都府")
	selfName := uniquePersonName("編集中の人物")
	selfID := testutil.NewPersonBuilder(t, tx).WithName(selfName).Build()
	takenName := uniquePersonName("別の人物")
	testutil.NewPersonBuilder(t, tx).WithName(takenName).Build()
	delistedName := uniquePersonName("掲載停止の人物")
	testutil.NewPersonBuilder(t, tx).WithName(delistedName).WithStatus(model.PersonStatusDelisted).Build()

	t.Run("名前を変えない送信は自分自身と重複しない", func(t *testing.T) {
		fields := &DBPersonUpdateValidateOutput{Name: selfName, PrefectureID: &prefectureID}

		got, err := v.ValidateDatabase(ctx, selfID, fields)
		if err != nil {
			t.Fatalf("ValidateDatabase()のエラー = %v", err)
		}
		if got != fields {
			t.Errorf("ValidateDatabase() = %+v、期待値 = 受け取った値をそのまま返す", got)
		}
	})

	missingPrefectureID := model.PrefectureID(999999999)
	tests := []struct {
		name      string
		fields    DBPersonUpdateValidateOutput
		wantField string
		wantMsg   string
	}{
		{
			name:      "異常系: 名前が別の人物と重複",
			fields:    DBPersonUpdateValidateOutput{Name: takenName},
			wantField: "name",
			wantMsg:   "この名前の人物は既に登録されています",
		},
		{
			name:      "異常系: 名前が掲載停止の人物と重複",
			fields:    DBPersonUpdateValidateOutput{Name: delistedName},
			wantField: "name",
			wantMsg:   "この人物は掲載を停止しています。関係者に関わる事情のため、停止の詳細はお伝えできません。ご理解いただけますと幸いです。",
		},
		{
			name:      "異常系: 出身地が登録されていない",
			fields:    DBPersonUpdateValidateOutput{Name: selfName, PrefectureID: &missingPrefectureID},
			wantField: "prefecture_id",
			wantMsg:   "正しい出身地を選択してください",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := tt.fields
			got, err := v.ValidateDatabase(ctx, selfID, &fields)
			if got != nil {
				t.Errorf("ValidateDatabase() = %+v、期待値 = nil", got)
			}
			assertPersonUpdateValidationError(t, err, tt.wantField, tt.wantMsg, false)
		})
	}
}

// assertPersonUpdateValidationErrorは、errが期待するフィールドやフォーム全体のエラーを含む
// *model.ValidationErrorであることを確かめる。
func assertPersonUpdateValidationError(t *testing.T, err error, wantField, wantMsg string, wantGlobal bool) {
	t.Helper()

	ve := model.AsValidationError(err)
	if ve == nil {
		t.Fatalf("エラー = %v、期待値 = *model.ValidationError", err)
	}
	if wantField != "" && !ve.HasFieldError(wantField) {
		t.Errorf("フィールド%qのエラーがありません: %+v", wantField, ve)
	}
	if wantMsg != "" && !slices.Contains(ve.Fields[wantField], wantMsg) {
		t.Errorf("フィールド%qのエラー = %v、期待値に%qを含む", wantField, ve.Fields[wantField], wantMsg)
	}
	if wantGlobal && len(ve.Global) == 0 {
		t.Errorf("グローバルエラーがありません: %+v", ve)
	}
}

// TestDBPersonUpdateValidatorValidateFormatReportsEveryProblemは複数の形式の問題を同時に
// 含む送信を検証する。すべてが報告されるため、編集者は一度に直せる。
func TestDBPersonUpdateValidatorValidateFormatReportsEveryProblem(t *testing.T) {
	t.Parallel()

	v, _ := newTestDBPersonUpdateValidator(t)
	ctx := i18n.SetLocale(context.Background(), "ja")

	input := DBPersonUpdateValidatorInput{
		Name:      "",
		Gender:    "other",
		Birthday:  "昨日",
		Height:    "高い",
		URL:       "ftp://example.com",
		UpdatedAt: "",
	}

	_, err := v.ValidateFormat(ctx, input)
	ve := model.AsValidationError(err)
	if ve == nil {
		t.Fatalf("ValidateFormat()のエラー = %v、期待値 = *model.ValidationError", err)
	}
	for _, field := range []string{"name", "gender", "birthday", "height", "url"} {
		if !ve.HasFieldError(field) {
			t.Errorf("フィールド%qのエラーがありません: %+v", field, ve)
		}
	}
	if len(ve.Global) == 0 {
		t.Errorf("版のグローバルエラーがありません: %+v", ve)
	}
}
