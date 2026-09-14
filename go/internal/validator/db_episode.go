package validator

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// episodeRowSeparatorは送信された1行を列に分ける文字。Annict DBの一括作成フォームが
// 従来案内してきたカンマで、編集者がAnnictの外に持っている行をそのまま貼り付けられる。
const episodeRowSeparator = ","

// episodeRowColumnCountは1行が持つ列の数 (表示用話数・数値話数・タイトル)。タイトルは
// 最後に取り出して行の残り全体を受け取るため、カンマを含むタイトルも分割で欠けない。
const episodeRowColumnCount = 3

// episodeLimitedTextMaxLengthはcharacter varying(510) で宣言されたepisodesのカラム
// (number / title) の上限。上限が無いと、長すぎる値がバリデーションを通ったあとINSERTが
// ドライバのエラーで失敗し、送信は全行を失ったまま500で終わる。PostgreSQLのvarchar(n) は
// 文字数で数えるため、検査もバイト数ではなく文字数で行う。
const episodeLimitedTextMaxLength = 500

// episodeCreateMaxRowsは1回の一括作成トランザクションの上限。フォームは1シーズン
// 程度の貼り付けを想定しており、100行なら長めのクールも十分扱いつつ、誤って大量に貼り付けた
// ときの行ロックとDB往復回数を制限できる。
const episodeCreateMaxRows = 100

// DBEpisodeCreateValidatorはAnnict DB管理画面の一括作成フォームを検証する。1つの
// textareaに1行1エピソードの形で入力される。
type DBEpisodeCreateValidator struct{}

func NewDBEpisodeCreateValidator() *DBEpisodeCreateValidator {
	return &DBEpisodeCreateValidator{}
}

type DBEpisodeCreateValidatorInput struct {
	Rows string
}

// DBEpisodeRowは送信された1行をパースした結果。各フィールドは値の格納先である
// episodesのNULL許容カラム (number / raw_number / title) に対応しており、空の列はnilの
// まま運ばれてNULLとして書かれる。既存のどの行も持っていない空文字列にはしない。
type DBEpisodeRow struct {
	Number    *string
	RawNumber *float64
	Title     *string
}

// Validateは送信された各行をパースし、入力順に返す。エラーには由来する行番号を付ける。
// 行番号は送信された入力そのものを数えるため、送信者がtextareaで見ている番号と一致する
// (空行は行としては読み飛ばすが、番号は進める)。
//
// 1行でも不正なら送信全体を失敗させる。行の作成は1トランザクションで行い部分的にDBへ
// 届くことは無いため、エラーと一緒に正常な行を返しても呼び出し元が作れない状態を表すことに
// なる。
func (v *DBEpisodeCreateValidator) Validate(ctx context.Context, input DBEpisodeCreateValidatorInput) ([]DBEpisodeRow, error) {
	ve := model.NewValidationError()

	lines := splitEpisodeRowLines(input.Rows)
	nonBlankRowCount := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonBlankRowCount++
		}
	}
	if nonBlankRowCount > episodeCreateMaxRows {
		ve.AddField("rows", i18n.T(ctx, "validation_episode_rows_too_many", map[string]any{
			"MaxRows": episodeCreateMaxRows,
		}))
		return nil, ve
	}

	rows := make([]DBEpisodeRow, 0)
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		row, messages := parseEpisodeRow(ctx, line)
		for _, message := range messages {
			ve.AddField("rows", i18n.T(ctx, "validation_episode_row_error", map[string]any{
				"Line":    i + 1,
				"Message": message,
			}))
		}
		if len(messages) == 0 {
			rows = append(rows, row)
		}
	}

	if len(rows) == 0 && !ve.HasErrors() {
		ve.AddField("rows", i18n.T(ctx, "validation_required"))
	}

	if ve.HasErrors() {
		return nil, ve
	}

	return rows, nil
}

// splitEpisodeRowLinesは送信されたtextareaを行に分割する。フォーム送信ではtextarea
// の改行がCRLFで符号化され、貼り付けた内容から単独のCRが届くこともあるため、まず双方を
// LFに畳む。LFだけで分割すると、各行の最後の列にCRが残ってしまう。
func splitEpisodeRowLines(rows string) []string {
	normalized := strings.ReplaceAll(rows, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")

	return strings.Split(normalized, "\n")
}

// parseEpisodeRowは1行を列に分けて検査し、パース結果と失敗したチェックのメッセージを
// 返す。すべてのチェックを実行するため、2箇所に問題がある行は片方がもう片方の陰に隠れず
// 両方報告される。
func parseEpisodeRow(ctx context.Context, line string) (DBEpisodeRow, []string) {
	columns := strings.SplitN(line, episodeRowSeparator, episodeRowColumnCount)
	number := episodeRowColumn(columns, 0)
	rawNumber := episodeRowColumn(columns, 1)
	title := episodeRowColumn(columns, 2)

	var messages []string

	if number == "" && title == "" {
		messages = append(messages, i18n.T(ctx, "validation_episode_row_number_or_title_required"))
	}
	if utf8.RuneCountInString(number) > episodeLimitedTextMaxLength {
		messages = append(messages, i18n.T(ctx, "validation_episode_row_number_too_long", map[string]any{
			"MaxLength": episodeLimitedTextMaxLength,
		}))
	}
	if utf8.RuneCountInString(title) > episodeLimitedTextMaxLength {
		messages = append(messages, i18n.T(ctx, "validation_episode_row_title_too_long", map[string]any{
			"MaxLength": episodeLimitedTextMaxLength,
		}))
	}

	parsedRawNumber, ok := parseEpisodeRowRawNumber(rawNumber)
	if !ok {
		messages = append(messages, i18n.T(ctx, "validation_episode_row_raw_number_invalid"))
	}

	return DBEpisodeRow{
		Number:    optionalEpisodeRowText(number),
		RawNumber: parsedRawNumber,
		Title:     optionalEpisodeRowText(title),
	}, messages
}

// episodeRowColumnは指定位置の列を前後の空白を取り除いて返し、行がその位置まで無い場合
// は空文字列を返す。行は末尾の列をまるごと省略できる (タイトルの無い話数は "#1,1" として
// 送信される) ため、列が無いことは列が空であることと同じ意味になる。
func episodeRowColumn(columns []string, index int) string {
	if index >= len(columns) {
		return ""
	}

	return strings.TrimSpace(columns[index])
}

// parseEpisodeRowRawNumberは行の数値話数を返し、数値でないものが入っている場合はfalse
// を返す。空の列は許可してnilを返す。数字で管理されていないエピソードのために数値話数は
// 任意入力であるため。
//
// NaNと無限大はfloatとしてパースできるが弾く。episodes.raw_numberはdouble precision
// カラムでこれらを格納できるため、受け入れると以降の話数計算がすべて意味を成さなくなる値が
// カラムに入ってしまう。
func parseEpisodeRowRawNumber(value string) (*float64, bool) {
	if value == "" {
		return nil, true
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return nil, false
	}

	return &parsed, true
}

// DBEpisodeUpdateValidatorはAnnict DB管理画面のエピソード編集フォームの1回の送信を
// 検証する。
type DBEpisodeUpdateValidator struct{}

func NewDBEpisodeUpdateValidator() *DBEpisodeUpdateValidator {
	return &DBEpisodeUpdateValidator{}
}

// DBEpisodeUpdateValidatorInputは届いたままの送信フォーム。hiddenが運ぶ版を含め、
// すべてのフィールドが文字列。
type DBEpisodeUpdateValidatorInput struct {
	Number     string
	RawNumber  string
	SortNumber string
	Title      string
	TitleEn    string
	UpdatedAt  string
}

// DBEpisodeUpdateValidateOutputは受理された送信を、episodesの各カラムが保持する形へ
// 変換したもの。任意入力のテキストと数値のカラムはポインタで運び、編集者が消したフィールドを、
// 既存のどの行も持たない空文字列ではなくNULLとして書けるようにする。
//
// UpdatedAtは送信が前提とする版。nilはFormNullVersionのセンチネルで、更新側は
// updated_at IS NULLで照合する。
type DBEpisodeUpdateValidateOutput struct {
	Number     *string
	RawNumber  *float64
	Title      *string
	TitleEn    string
	SortNumber int32
	UpdatedAt  *time.Time
}

// Validateは送信されたエピソード編集フォームを検証し、保存先のカラムの型へ変換して返す。
// 数値フィールドは変換が使う範囲そのもので検査するため、ここを通った値は保存できる。本メソッドが
// 通して変換が落とす値は、エラーも出ないままNULLで保存され、カラムに収まらない値はUPDATEが
// 失敗して500になるため。
//
// 版の検査は先出しせず他の項目と一緒に行う。古い版でありかつ形式も不正な送信で、編集者が対処
// すべきことが一度に揃って報告されるようにする。
func (v *DBEpisodeUpdateValidator) Validate(ctx context.Context, input DBEpisodeUpdateValidatorInput) (*DBEpisodeUpdateValidateOutput, error) {
	ve := model.NewValidationError()

	number := strings.TrimSpace(input.Number)
	title := strings.TrimSpace(input.Title)
	titleEn := strings.TrimSpace(input.TitleEn)
	sortNumber := strings.TrimSpace(input.SortNumber)
	rawNumber := strings.TrimSpace(input.RawNumber)

	validateEpisodeMaxLength(ctx, ve, "number", number)
	validateEpisodeMaxLength(ctx, ve, "title", title)
	validateOptionalFloat(ctx, ve, "raw_number", rawNumber)

	if sortNumber == "" {
		ve.AddField("sort_number", i18n.T(ctx, "validation_required"))
	} else {
		validateOptionalInt32(ctx, ve, "sort_number", sortNumber)
	}

	// 版は編集できるフィールドではないため、欠落や形式不正はフォーム全体に対して述べる。
	// 編集者が直せる入力は無く、ページを開き直すしかないため。
	version, versionOK := parseFormVersion(input.UpdatedAt)
	if !versionOK {
		ve.AddGlobal(i18n.T(ctx, "validation_version_missing"))
	}

	if ve.HasErrors() {
		return nil, ve
	}

	parsedSortNumber, err := strconv.ParseInt(sortNumber, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("ソート番号の変換に失敗: %w", err)
	}

	return &DBEpisodeUpdateValidateOutput{
		Number:     optionalEpisodeRowText(number),
		RawNumber:  parseEpisodeOptionalFloat(rawNumber),
		Title:      optionalEpisodeRowText(title),
		TitleEn:    titleEn,
		SortNumber: int32(parsedSortNumber),
		UpdatedAt:  version,
	}, nil
}

// validateEpisodeMaxLengthはepisodesのvarchar(510) カラムに収まらない長さの値を弾く。
// PostgreSQLのvarchar(n) は文字数で数えるため文字数で数える。バイト数で数えると日本語の
// タイトルを実際の上限の1/3で弾いてしまう。
func validateEpisodeMaxLength(ctx context.Context, ve *model.ValidationError, field, value string) {
	if utf8.RuneCountInString(value) > episodeLimitedTextMaxLength {
		ve.AddField(field, i18n.T(ctx, "validation_too_long", map[string]any{
			"MaxLength": episodeLimitedTextMaxLength,
		}))
	}
}

// parseEpisodeOptionalFloatは検証済みの数値話数を変換する。空のフィールドではnilを返し、
// カラムがNULLとして書かれるようにする。
func parseEpisodeOptionalFloat(value string) *float64 {
	parsed, _ := parseEpisodeRowRawNumber(value)
	return parsed
}

// optionalEpisodeRowTextは空の列に対してnilを返し、値がNULLとして格納されるように
// する。
func optionalEpisodeRowText(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}
