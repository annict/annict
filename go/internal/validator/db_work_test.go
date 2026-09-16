package validator

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// dbWorkTestTitleはタイトル以外を対象とするケースが送信するタイトル。タイトルの
// 一意性検査はworksを参照するため、他のテストがコミットしない値である必要がある。他の
// テストが残したタイトルと重なると、本ファイルの「エラーなし」のケースがすべて落ちる。
const dbWorkTestTitle = "バリデーターテスト作品"

// newTestDBWorkValidatorは共有テストDBに対してバリデーターを組み立てる。タイトルの
// 一意性検査と話数フォーマットの検査がworksとnumber_formatsを読むため、全ケースで
// リポジトリが要る。SetupTxはテストが挿入した行を他のテストから隠し、終了時にロール
// バックする。これらの検査に見つけさせたい行をテストが用意できるよう、トランザクションも返す。
func newTestDBWorkValidator(t *testing.T) (*DBWorkCreateValidator, *sql.Tx) {
	t.Helper()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	return NewDBWorkCreateValidator(
		repository.NewWorkRepository(queries),
		repository.NewNumberFormatRepository(queries),
	), tx
}

// seedNumberFormatは話数フォーマットを挿入し、そのidを返す。idはシーケンス由来の
// ためテストケースのリテラルに書けず、実行時に差し込む必要がある。number_formats.nameには
// 一意インデックスがあり、並行するテストのトランザクションがテーブルを共有するため、名前は
// テスト名から作る。
func seedNumberFormat(t *testing.T, tx *sql.Tx) model.NumberFormatID {
	t.Helper()

	var id int64
	err := tx.QueryRowContext(
		context.Background(),
		`INSERT INTO number_formats (name, sort_number, created_at, updated_at)
		 VALUES ($1, 0, NOW(), NOW()) RETURNING id`,
		"話数フォーマット_"+t.Name(),
	).Scan(&id)
	if err != nil {
		t.Fatalf("number_formatsの挿入に失敗: %v", err)
	}

	return model.NumberFormatID(id)
}

func TestDBWorkCreateValidatorValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input DBWorkCreateValidatorInput
		// withSeededNumberFormatはinput.NumberFormatIDを、ケースごとに用意した話数
		// フォーマットのidで置き換える。シーケンス採番のidは下のリテラルに書けないため。
		withSeededNumberFormat bool
		wantErrors             bool
		wantFields             []string
	}{
		{
			name: "正常系: 必須フィールドのみ",
			input: DBWorkCreateValidatorInput{
				Title: dbWorkTestTitle,
				Media: "1",
			},
			wantErrors: false,
		},
		{
			name: "正常系: 全フィールド入力",
			input: DBWorkCreateValidatorInput{
				Title:                 dbWorkTestTitle,
				TitleKana:             "てすとあにめ",
				TitleAlter:            "別名",
				TitleEn:               "Test Anime",
				TitleAlterEn:          "Alt Name",
				Media:                 "1",
				SeasonYear:            "2024",
				SeasonName:            "2",
				StartedOn:             "2024-04-01",
				EndedOn:               "2024-06-30",
				OfficialSiteURL:       "https://example.com",
				OfficialSiteURLEn:     "https://example.com/en",
				WikipediaURL:          "https://ja.wikipedia.org/wiki/Test",
				WikipediaURLEn:        "https://en.wikipedia.org/wiki/Test",
				TwitterUsername:       "testanime",
				TwitterHashtag:        "テストアニメ",
				ScTid:                 "12345",
				MalAnimeID:            "54321",
				Synopsis:              "テストのあらすじ",
				SynopsisSource:        "公式サイト",
				SynopsisEn:            "Test synopsis",
				SynopsisSourceEn:      "Official site",
				ManualEpisodesCount:   "12",
				StartEpisodeRawNumber: "1",
				NoEpisodes:            "1",
			},
			withSeededNumberFormat: true,
			wantErrors:             false,
		},
		{
			name: "異常系: タイトルが空",
			input: DBWorkCreateValidatorInput{
				Title: "",
				Media: "1",
			},
			wantErrors: true,
			wantFields: []string{"title"},
		},
		{
			name: "異常系: タイトルがwhitespaceのみ",
			input: DBWorkCreateValidatorInput{
				Title: "   ",
				Media: "1",
			},
			wantErrors: true,
			wantFields: []string{"title"},
		},
		{
			name: "異常系: メディアが空",
			input: DBWorkCreateValidatorInput{
				Title: dbWorkTestTitle,
				Media: "",
			},
			wantErrors: true,
			wantFields: []string{"media"},
		},
		{
			name: "異常系: メディアが不正な値",
			input: DBWorkCreateValidatorInput{
				Title: dbWorkTestTitle,
				Media: "99",
			},
			wantErrors: true,
			wantFields: []string{"media"},
		},
		{
			name: "異常系: タイトルとメディアの両方が空",
			input: DBWorkCreateValidatorInput{
				Title: "",
				Media: "",
			},
			wantErrors: true,
			wantFields: []string{"title", "media"},
		},
		{
			name: "正常系: メディアが0 (その他)",
			input: DBWorkCreateValidatorInput{
				Title: dbWorkTestTitle,
				Media: "0",
			},
			wantErrors: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v, tx := newTestDBWorkValidator(t)

			input := tt.input
			if tt.withSeededNumberFormat {
				input.NumberFormatID = seedNumberFormat(t, tx).String()
			}

			ctx := context.Background()
			_, err := v.Validate(ctx, input)
			ve := model.AsValidationError(err)

			if tt.wantErrors {
				if ve == nil {
					t.Error("エラーが期待されましたが、エラーがありませんでした")
					return
				}

				for _, field := range tt.wantFields {
					if !ve.HasFieldError(field) {
						t.Errorf("フィールド%sのエラーが期待されましたが、見つかりませんでした", field)
					}
				}
			} else {
				if ve != nil {
					t.Errorf("エラーは期待されていませんでしたが、返されました: %+v", ve)
				}
			}
		})
	}
}

func TestDBWorkCreateValidatorValidate_URL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      DBWorkCreateValidatorInput
		wantErrors bool
		wantFields []string
	}{
		{
			name: "正常系: URLが空 (スキップ)",
			input: DBWorkCreateValidatorInput{
				Title:           dbWorkTestTitle,
				Media:           "1",
				OfficialSiteURL: "",
			},
			wantErrors: false,
		},
		{
			name: "正常系: 有効なhttps URL",
			input: DBWorkCreateValidatorInput{
				Title:           dbWorkTestTitle,
				Media:           "1",
				OfficialSiteURL: "https://example.com",
			},
			wantErrors: false,
		},
		{
			name: "正常系: 有効なhttp URL",
			input: DBWorkCreateValidatorInput{
				Title:           dbWorkTestTitle,
				Media:           "1",
				OfficialSiteURL: "http://example.com",
			},
			wantErrors: false,
		},
		{
			name: "異常系: スキームなしのURL",
			input: DBWorkCreateValidatorInput{
				Title:           dbWorkTestTitle,
				Media:           "1",
				OfficialSiteURL: "example.com",
			},
			wantErrors: true,
			wantFields: []string{"official_site_url"},
		},
		{
			name: "異常系: 不正なURL",
			input: DBWorkCreateValidatorInput{
				Title:           dbWorkTestTitle,
				Media:           "1",
				OfficialSiteURL: "not-a-url",
			},
			wantErrors: true,
			wantFields: []string{"official_site_url"},
		},
		{
			name: "異常系: ftpスキーム",
			input: DBWorkCreateValidatorInput{
				Title:           dbWorkTestTitle,
				Media:           "1",
				OfficialSiteURL: "ftp://example.com/file",
			},
			wantErrors: true,
			wantFields: []string{"official_site_url"},
		},
		{
			name: "異常系: 複数のURL項目でエラー",
			input: DBWorkCreateValidatorInput{
				Title:             dbWorkTestTitle,
				Media:             "1",
				OfficialSiteURL:   "invalid",
				OfficialSiteURLEn: "invalid",
				WikipediaURL:      "invalid",
				WikipediaURLEn:    "invalid",
			},
			wantErrors: true,
			wantFields: []string{"official_site_url", "official_site_url_en", "wikipedia_url", "wikipedia_url_en"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v, _ := newTestDBWorkValidator(t)

			ctx := context.Background()
			_, err := v.Validate(ctx, tt.input)
			ve := model.AsValidationError(err)

			if tt.wantErrors {
				if ve == nil {
					t.Error("エラーが期待されましたが、エラーがありませんでした")
					return
				}

				for _, field := range tt.wantFields {
					if !ve.HasFieldError(field) {
						t.Errorf("フィールド%sのエラーが期待されましたが、見つかりませんでした", field)
					}
				}
			} else {
				if ve != nil {
					t.Errorf("エラーは期待されていませんでしたが、返されました: %+v", ve)
				}
			}
		})
	}
}

func TestDBWorkCreateValidatorValidate_NumericFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      DBWorkCreateValidatorInput
		wantErrors bool
		wantFields []string
	}{
		{
			name: "正常系: sc_tidが空 (スキップ)",
			input: DBWorkCreateValidatorInput{
				Title: dbWorkTestTitle,
				Media: "1",
				ScTid: "",
			},
			wantErrors: false,
		},
		{
			name: "正常系: sc_tidが有効な整数",
			input: DBWorkCreateValidatorInput{
				Title: dbWorkTestTitle,
				Media: "1",
				ScTid: "12345",
			},
			wantErrors: false,
		},
		{
			name: "異常系: sc_tidが整数でない",
			input: DBWorkCreateValidatorInput{
				Title: dbWorkTestTitle,
				Media: "1",
				ScTid: "abc",
			},
			wantErrors: true,
			wantFields: []string{"sc_tid"},
		},
		{
			name: "異常系: sc_tidが小数",
			input: DBWorkCreateValidatorInput{
				Title: dbWorkTestTitle,
				Media: "1",
				ScTid: "12.5",
			},
			wantErrors: true,
			wantFields: []string{"sc_tid"},
		},
		{
			name: "正常系: mal_anime_idが有効な整数",
			input: DBWorkCreateValidatorInput{
				Title:      dbWorkTestTitle,
				Media:      "1",
				MalAnimeID: "54321",
			},
			wantErrors: false,
		},
		{
			name: "異常系: mal_anime_idが整数でない",
			input: DBWorkCreateValidatorInput{
				Title:      dbWorkTestTitle,
				Media:      "1",
				MalAnimeID: "xyz",
			},
			wantErrors: true,
			wantFields: []string{"mal_anime_id"},
		},
		{
			// integerカラムはint32で、それを超える値はここで報告される必要がある。
			// 変換も同じビット幅でパースするため、報告しないと何も言わずに捨てられる。
			name: "異常系: sc_tidがint32の範囲外",
			input: DBWorkCreateValidatorInput{
				Title: dbWorkTestTitle,
				Media: "1",
				ScTid: "2147483648",
			},
			wantErrors: true,
			wantFields: []string{"sc_tid"},
		},
		{
			name: "異常系: mal_anime_idがint32の範囲外",
			input: DBWorkCreateValidatorInput{
				Title:      dbWorkTestTitle,
				Media:      "1",
				MalAnimeID: "2147483648",
			},
			wantErrors: true,
			wantFields: []string{"mal_anime_id"},
		},
		{
			name: "異常系: manual_episodes_countが整数でない",
			input: DBWorkCreateValidatorInput{
				Title:               dbWorkTestTitle,
				Media:               "1",
				ManualEpisodesCount: "たくさん",
			},
			wantErrors: true,
			wantFields: []string{"manual_episodes_count"},
		},
		{
			name: "異常系: manual_episodes_countがint32の範囲外",
			input: DBWorkCreateValidatorInput{
				Title:               dbWorkTestTitle,
				Media:               "1",
				ManualEpisodesCount: "9999999999",
			},
			wantErrors: true,
			wantFields: []string{"manual_episodes_count"},
		},
		{
			name: "異常系: season_yearが整数でない",
			input: DBWorkCreateValidatorInput{
				Title:      dbWorkTestTitle,
				Media:      "1",
				SeasonYear: "令和6年",
			},
			wantErrors: true,
			wantFields: []string{"season_year"},
		},
		{
			name: "異常系: season_nameが整数でない",
			input: DBWorkCreateValidatorInput{
				Title:      dbWorkTestTitle,
				Media:      "1",
				SeasonName: "spring",
			},
			wantErrors: true,
			wantFields: []string{"season_name"},
		},
		{
			name: "正常系: season_nameが冬のenum値",
			input: DBWorkCreateValidatorInput{
				Title:      dbWorkTestTitle,
				Media:      "1",
				SeasonName: "1",
			},
			wantErrors: false,
		},
		{
			name: "正常系: season_nameが秋のenum値",
			input: DBWorkCreateValidatorInput{
				Title:      dbWorkTestTitle,
				Media:      "1",
				SeasonName: "4",
			},
			wantErrors: false,
		},
		{
			name: "異常系: season_nameが許可値の下限未満",
			input: DBWorkCreateValidatorInput{
				Title:      dbWorkTestTitle,
				Media:      "1",
				SeasonName: "0",
			},
			wantErrors: true,
			wantFields: []string{"season_name"},
		},
		{
			name: "異常系: season_nameが許可値の上限超過",
			input: DBWorkCreateValidatorInput{
				Title:      dbWorkTestTitle,
				Media:      "1",
				SeasonName: "5",
			},
			wantErrors: true,
			wantFields: []string{"season_name"},
		},
		{
			name: "異常系: number_format_idが整数でない",
			input: DBWorkCreateValidatorInput{
				Title:          dbWorkTestTitle,
				Media:          "1",
				NumberFormatID: "第n話",
			},
			wantErrors: true,
			wantFields: []string{"number_format_id"},
		},
		{
			// start_episode_raw_numberはdouble precisionカラムに対応するため、
			// 整数フィールドが弾く小数もここでは有効な値になる。
			name: "正常系: start_episode_raw_numberが小数",
			input: DBWorkCreateValidatorInput{
				Title:                 dbWorkTestTitle,
				Media:                 "1",
				StartEpisodeRawNumber: "2.5",
			},
			wantErrors: false,
		},
		{
			name: "異常系: start_episode_raw_numberが数値でない",
			input: DBWorkCreateValidatorInput{
				Title:                 dbWorkTestTitle,
				Media:                 "1",
				StartEpisodeRawNumber: "第1話",
			},
			wantErrors: true,
			wantFields: []string{"start_episode_raw_number"},
		},
		{
			// NaNと無限大はfloat64としてパースでき、double precisionにも格納できる
			// ため、明示的に検査しない限りカラムに入ってしまう。
			name: "異常系: start_episode_raw_numberがNaN",
			input: DBWorkCreateValidatorInput{
				Title:                 dbWorkTestTitle,
				Media:                 "1",
				StartEpisodeRawNumber: "NaN",
			},
			wantErrors: true,
			wantFields: []string{"start_episode_raw_number"},
		},
		{
			name: "異常系: start_episode_raw_numberが無限大",
			input: DBWorkCreateValidatorInput{
				Title:                 dbWorkTestTitle,
				Media:                 "1",
				StartEpisodeRawNumber: "Infinity",
			},
			wantErrors: true,
			wantFields: []string{"start_episode_raw_number"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v, _ := newTestDBWorkValidator(t)

			ctx := context.Background()
			_, err := v.Validate(ctx, tt.input)
			ve := model.AsValidationError(err)

			if tt.wantErrors {
				if ve == nil {
					t.Error("エラーが期待されましたが、エラーがありませんでした")
					return
				}

				for _, field := range tt.wantFields {
					if !ve.HasFieldError(field) {
						t.Errorf("フィールド%sのエラーが期待されましたが、見つかりませんでした", field)
					}
				}
			} else {
				if ve != nil {
					t.Errorf("エラーは期待されていませんでしたが、返されました: %+v", ve)
				}
			}
		})
	}
}

func TestDBWorkCreateValidatorValidate_PresencePair(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      DBWorkCreateValidatorInput
		wantErrors bool
		wantFields []string
	}{
		{
			name: "正常系: あらすじと出典の両方が空",
			input: DBWorkCreateValidatorInput{
				Title:    dbWorkTestTitle,
				Media:    "1",
				Synopsis: "",
			},
			wantErrors: false,
		},
		{
			name: "正常系: あらすじと出典の両方がある",
			input: DBWorkCreateValidatorInput{
				Title:          dbWorkTestTitle,
				Media:          "1",
				Synopsis:       "テストのあらすじ",
				SynopsisSource: "公式サイト",
			},
			wantErrors: false,
		},
		{
			name: "異常系: あらすじのみで出典がない",
			input: DBWorkCreateValidatorInput{
				Title:          dbWorkTestTitle,
				Media:          "1",
				Synopsis:       "テストのあらすじ",
				SynopsisSource: "",
			},
			wantErrors: true,
			wantFields: []string{"synopsis_source"},
		},
		{
			name: "正常系: 出典のみ (あらすじなし) は許可",
			input: DBWorkCreateValidatorInput{
				Title:          dbWorkTestTitle,
				Media:          "1",
				Synopsis:       "",
				SynopsisSource: "公式サイト",
			},
			wantErrors: false,
		},
		{
			name: "正常系: 英語あらすじと出典の両方がある",
			input: DBWorkCreateValidatorInput{
				Title:            dbWorkTestTitle,
				Media:            "1",
				SynopsisEn:       "Test synopsis",
				SynopsisSourceEn: "Official site",
			},
			wantErrors: false,
		},
		{
			name: "異常系: 英語あらすじのみで出典がない",
			input: DBWorkCreateValidatorInput{
				Title:            dbWorkTestTitle,
				Media:            "1",
				SynopsisEn:       "Test synopsis",
				SynopsisSourceEn: "",
			},
			wantErrors: true,
			wantFields: []string{"synopsis_source_en"},
		},
		{
			name: "異常系: 日英両方のあらすじに出典がない",
			input: DBWorkCreateValidatorInput{
				Title:      dbWorkTestTitle,
				Media:      "1",
				Synopsis:   "テストのあらすじ",
				SynopsisEn: "Test synopsis",
			},
			wantErrors: true,
			wantFields: []string{"synopsis_source", "synopsis_source_en"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v, _ := newTestDBWorkValidator(t)

			ctx := context.Background()
			_, err := v.Validate(ctx, tt.input)
			ve := model.AsValidationError(err)

			if tt.wantErrors {
				if ve == nil {
					t.Error("エラーが期待されましたが、エラーがありませんでした")
					return
				}

				for _, field := range tt.wantFields {
					if !ve.HasFieldError(field) {
						t.Errorf("フィールド%sのエラーが期待されましたが、見つかりませんでした", field)
					}
				}
			} else {
				if ve != nil {
					t.Errorf("エラーは期待されていませんでしたが、返されました: %+v", ve)
				}
			}
		})
	}
}

// TestDBWorkCreateValidatorValidate_MaxLengthはcharacter varying(510) で宣言された
// worksのカラムを対象とする。ここで上限を検査する前は、長すぎる値がINSERTまで届き、
// フォームの入力を丸ごと失ったまま500で返っていた。
func TestDBWorkCreateValidatorValidate_MaxLength(t *testing.T) {
	t.Parallel()

	// 形式ではなく長さのケースにするため、有効なURLのまま上限を超える長さにする。
	longURL := func(length int) string {
		const prefix = "https://example.com/"
		return prefix + strings.Repeat("a", length-len(prefix))
	}

	tests := []struct {
		name       string
		input      DBWorkCreateValidatorInput
		wantErrors bool
		wantFields []string
	}{
		{
			// 全角500文字は1500バイトある。varchar(n) の数え方である文字数ではなく
			// バイト数で上限を測っていると、このケースが落ちる。
			name: "正常系: タイトルが上限ちょうど (全角500文字)",
			input: DBWorkCreateValidatorInput{
				Title: strings.Repeat("あ", 500),
				Media: "1",
			},
			wantErrors: false,
		},
		{
			name: "異常系: タイトルが上限超過",
			input: DBWorkCreateValidatorInput{
				Title: strings.Repeat("あ", 501),
				Media: "1",
			},
			wantErrors: true,
			wantFields: []string{"title"},
		},
		{
			name: "異常系: 公式サイトURLが上限超過",
			input: DBWorkCreateValidatorInput{
				Title:           dbWorkTestTitle,
				Media:           "1",
				OfficialSiteURL: longURL(501),
			},
			wantErrors: true,
			wantFields: []string{"official_site_url"},
		},
		{
			name: "異常系: WikipediaのURLが上限超過",
			input: DBWorkCreateValidatorInput{
				Title:        dbWorkTestTitle,
				Media:        "1",
				WikipediaURL: longURL(501),
			},
			wantErrors: true,
			wantFields: []string{"wikipedia_url"},
		},
		{
			name: "異常系: Xのユーザー名が上限超過",
			input: DBWorkCreateValidatorInput{
				Title:           dbWorkTestTitle,
				Media:           "1",
				TwitterUsername: strings.Repeat("a", 501),
			},
			wantErrors: true,
			wantFields: []string{"twitter_username"},
		},
		{
			name: "異常系: ハッシュタグが上限超過",
			input: DBWorkCreateValidatorInput{
				Title:          dbWorkTestTitle,
				Media:          "1",
				TwitterHashtag: strings.Repeat("a", 501),
			},
			wantErrors: true,
			wantFields: []string{"twitter_hashtag"},
		},
		{
			// 英語版のURLカラムはスキーマ上の長さ制限が無いため、長い値は、カラムに
			// 存在しない上限で弾かずに受け入れる。
			name: "正常系: 英語版URLは長さ制限なし",
			input: DBWorkCreateValidatorInput{
				Title:             dbWorkTestTitle,
				Media:             "1",
				OfficialSiteURLEn: longURL(600),
				WikipediaURLEn:    longURL(600),
			},
			wantErrors: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v, _ := newTestDBWorkValidator(t)

			_, err := v.Validate(context.Background(), tt.input)
			ve := model.AsValidationError(err)

			if tt.wantErrors {
				if ve == nil {
					t.Fatal("エラーが期待されましたが、エラーがありませんでした")
				}

				for _, field := range tt.wantFields {
					if !ve.HasFieldError(field) {
						t.Errorf("フィールド%sのエラーが期待されましたが、見つかりませんでした", field)
					}
				}
			} else if ve != nil {
				t.Errorf("エラーは期待されていませんでしたが、返されました: %+v", ve)
			}
		})
	}
}

// TestDBWorkCreateValidatorValidate_DateFieldsは日付フィールドを対象とする。以前は
// 解釈できない日付がカラムへ渡る途中で捨てられ、送信者が入力した日付が黙って欠けたまま
// 作品が保存されていた。
func TestDBWorkCreateValidatorValidate_DateFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      DBWorkCreateValidatorInput
		wantErrors bool
		wantFields []string
	}{
		{
			name: "正常系: 開始日・終了日が空",
			input: DBWorkCreateValidatorInput{
				Title: dbWorkTestTitle,
				Media: "1",
			},
			wantErrors: false,
		},
		{
			name: "正常系: 有効な日付",
			input: DBWorkCreateValidatorInput{
				Title:     dbWorkTestTitle,
				Media:     "1",
				StartedOn: "2024-04-01",
				EndedOn:   "2024-06-30",
			},
			wantErrors: false,
		},
		{
			name: "異常系: 開始日が日付として解釈できない",
			input: DBWorkCreateValidatorInput{
				Title:     dbWorkTestTitle,
				Media:     "1",
				StartedOn: "2024/04/01",
			},
			wantErrors: true,
			wantFields: []string{"started_on"},
		},
		{
			name: "異常系: 存在しない日付",
			input: DBWorkCreateValidatorInput{
				Title:   dbWorkTestTitle,
				Media:   "1",
				EndedOn: "2024-02-30",
			},
			wantErrors: true,
			wantFields: []string{"ended_on"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v, _ := newTestDBWorkValidator(t)

			_, err := v.Validate(context.Background(), tt.input)
			ve := model.AsValidationError(err)

			if tt.wantErrors {
				if ve == nil {
					t.Fatal("エラーが期待されましたが、エラーがありませんでした")
				}

				for _, field := range tt.wantFields {
					if !ve.HasFieldError(field) {
						t.Errorf("フィールド%sのエラーが期待されましたが、見つかりませんでした", field)
					}
				}
			} else if ve != nil {
				t.Errorf("エラーは期待されていませんでしたが、返されました: %+v", ve)
			}
		})
	}
}

// TestDBWorkCreateValidatorValidate_NumberFormatExistenceはnumber_formatsの行を指す
// number_format_idを対象とする。works.number_format_idと
// anime_classifications.number_format_idはいずれも同表への外部キーのため、存在しないidは
// INSERTで失敗し、送信は入力を失ったまま500で終わる。
func TestDBWorkCreateValidatorValidate_NumberFormatExistence(t *testing.T) {
	t.Parallel()

	t.Run("正常系: 話数フォーマットが空", func(t *testing.T) {
		t.Parallel()

		v, _ := newTestDBWorkValidator(t)

		_, err := v.Validate(context.Background(), DBWorkCreateValidatorInput{
			Title:          dbWorkTestTitle,
			Media:          "1",
			NumberFormatID: "",
		})
		if err != nil {
			t.Errorf("エラーは期待されていませんでしたが、返されました: %v", err)
		}
	})

	t.Run("正常系: 登録済みの話数フォーマット", func(t *testing.T) {
		t.Parallel()

		v, tx := newTestDBWorkValidator(t)

		_, err := v.Validate(context.Background(), DBWorkCreateValidatorInput{
			Title:          dbWorkTestTitle,
			Media:          "1",
			NumberFormatID: seedNumberFormat(t, tx).String(),
		})
		if err != nil {
			t.Errorf("エラーは期待されていませんでしたが、返されました: %v", err)
		}
	})

	t.Run("異常系: 存在しない話数フォーマット", func(t *testing.T) {
		t.Parallel()

		v, _ := newTestDBWorkValidator(t)

		// number_formatsのシーケンスが採番する範囲を大きく超えたid。他のテストが
		// いくつフォーマットを登録しても、このidの行は存在し得ない。
		const missingNumberFormatID = "9000000000000000000"

		_, err := v.Validate(context.Background(), DBWorkCreateValidatorInput{
			Title:          dbWorkTestTitle,
			Media:          "1",
			NumberFormatID: missingNumberFormatID,
		})

		ve := model.AsValidationError(err)
		if ve == nil {
			t.Fatalf("エラーが期待されましたが、エラーがありませんでした (err=%v)", err)
		}
		if !ve.HasFieldError("number_format_id") {
			t.Errorf("フィールドnumber_format_idのエラーが期待されましたが、見つかりませんでした: %+v", ve)
		}
	})
}

// TestDBWorkCreateValidatorValidate_TitleUniquenessはタイトルの一意性を対象とする。
// RailsのWorkのバリデーション (only_keptにスコープしたuniqueness) に対応する。非公開・
// 削除済みの作品はそのスコープの外にあり、タイトルは新しい作品のために空いたままになる。
func TestDBWorkCreateValidatorValidate_TitleUniqueness(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		// seedは既にDBにあるworkを作り、編集フォームと同じように除外できるよう
		// そのidを返す。
		seed          func(t *testing.T, tx *sql.Tx, title string) model.WorkID
		excludeSeeded bool
		wantErrors    bool
	}{
		{
			name: "異常系: 公開中の作品と重複",
			seed: func(t *testing.T, tx *sql.Tx, title string) model.WorkID {
				return testutil.NewWorkBuilder(t, tx).WithTitle(title).Build()
			},
			wantErrors: true,
		},
		{
			name: "正常系: 重複相手が自分自身 (編集で除外される)",
			seed: func(t *testing.T, tx *sql.Tx, title string) model.WorkID {
				return testutil.NewWorkBuilder(t, tx).WithTitle(title).Build()
			},
			excludeSeeded: true,
			wantErrors:    false,
		},
		{
			name: "正常系: 重複相手が非公開",
			seed: func(t *testing.T, tx *sql.Tx, title string) model.WorkID {
				return testutil.NewWorkBuilder(t, tx).
					WithTitle(title).
					WithUnpublishedAt(time.Now()).
					Build()
			},
			wantErrors: false,
		},
		{
			name: "正常系: 重複相手が削除済み",
			seed: func(t *testing.T, tx *sql.Tx, title string) model.WorkID {
				return testutil.NewWorkBuilder(t, tx).
					WithTitle(title).
					WithDeletedAt(time.Now()).
					Build()
			},
			wantErrors: false,
		},
		{
			name: "正常系: 同じタイトルの作品が無い",
			seed: func(t *testing.T, tx *sql.Tx, title string) model.WorkID {
				return testutil.NewWorkBuilder(t, tx).WithTitle(title + "の続編").Build()
			},
			wantErrors: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v, tx := newTestDBWorkValidator(t)

			// 用意する行は本テストのトランザクション内にあるが、タイトルはケースごとに
			// ユニークである必要がある。検査はテーブル全体を読み、他のテストがコミットした
			// 行も対象になるため。
			title := "一意性テスト作品_" + t.Name()
			seededID := tt.seed(t, tx, title)

			input := DBWorkCreateValidatorInput{Title: title, Media: "1"}
			if tt.excludeSeeded {
				input.ExcludeWorkID = &seededID
			}

			_, err := v.Validate(context.Background(), input)
			ve := model.AsValidationError(err)

			if tt.wantErrors {
				if ve == nil {
					t.Fatalf("エラーが期待されましたが、エラーがありませんでした (err=%v)", err)
				}
				if !ve.HasFieldError("title") {
					t.Error("フィールドtitleのエラーが期待されましたが、見つかりませんでした")
				}
			} else if err != nil {
				t.Errorf("エラーは期待されていませんでしたが、返されました: %v", err)
			}
		})
	}
}

// TestDBWorkCreateValidatorValidate_Versionは編集フォームが運ぶ版を対象とする。作成フローは
// 版を示さず、編集フローは示す必要がある。何を示すかが、更新がどの保存済みの行に書けるかを決める。
func TestDBWorkCreateValidatorValidate_Version(t *testing.T) {
	t.Parallel()

	stored := time.Date(2026, 8, 17, 1, 2, 3, 456789000, time.UTC)

	tests := []struct {
		name string
		// updatedAtは届いたままのhiddenフィールド。nilのポインタは作成フローで、版を
		// 示すべき保存済みの行が無い。
		updatedAt   *string
		wantErr     bool
		wantVersion *time.Time
	}{
		{
			name:      "正常系: 作成フローは版を示さない",
			updatedAt: nil,
		},
		{
			name:        "正常系: 保存済みの版をそのまま往復させる",
			updatedAt:   ptrString(stored.Format(FormVersionLayout)),
			wantVersion: &stored,
		},
		{
			name:      "正常系: updated_atがNULLの行はセンチネルで表す",
			updatedAt: ptrString(FormNullVersion),
		},
		{
			name:      "異常系: 版が空 (hiddenフィールドを伴わない送信)",
			updatedAt: ptrString(""),
			wantErr:   true,
		},
		{
			name:      "異常系: 版が往復書式でない",
			updatedAt: ptrString("2026-08-17 01:02:03"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v, _ := newTestDBWorkValidator(t)

			version, err := v.Validate(context.Background(), DBWorkCreateValidatorInput{
				Title:     "版テスト作品_" + t.Name(),
				Media:     "1",
				UpdatedAt: tt.updatedAt,
			})

			if tt.wantErr {
				ve := model.AsValidationError(err)
				if ve == nil {
					t.Fatalf("エラーが期待されましたが、エラーがありませんでした (err=%v)", err)
				}
				// 版は編集できるフィールドではないため、編集者が直せる入力欄に対してでは
				// なくフォーム全体に対して報告する。
				if len(ve.Global) == 0 {
					t.Error("フォーム全体のエラーが期待されましたが、見つかりませんでした")
				}
				if version != nil {
					t.Errorf("version = %v、期待値 = nil", version)
				}
				return
			}

			if err != nil {
				t.Fatalf("エラーは期待されていませんでしたが、返されました: %v", err)
			}
			switch {
			case tt.wantVersion == nil && version != nil:
				t.Errorf("version = %v、期待値 = nil", version)
			case tt.wantVersion != nil && version == nil:
				t.Errorf("version = nil、期待値 = %v", tt.wantVersion)
			case tt.wantVersion != nil && !version.Equal(*tt.wantVersion):
				t.Errorf("version = %v、期待値 = %v", version, tt.wantVersion)
			}
		})
	}
}

// ptrStringはvalueへのポインタを返す。不在に意味がある任意フィールド用。
func ptrString(value string) *string {
	return &value
}
