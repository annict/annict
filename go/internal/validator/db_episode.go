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

// episodeRowSeparator is the character that separates the columns of one submitted line. It
// is the comma the Annict DB bulk-create form has always been documented with, so the rows
// an editor keeps outside Annict paste in unchanged.
//
// [Ja] episodeRowSeparator は送信された 1 行を列に分ける文字。Annict DB の一括作成フォームが
// 従来案内してきたカンマで、編集者が Annict の外に持っている行をそのまま貼り付けられる。
const episodeRowSeparator = ","

// episodeRowColumnCount is the number of columns one line carries: the display number, the
// numeric number, and the title. The title is taken last and absorbs the rest of the line,
// so a title that itself contains a comma survives the split.
//
// [Ja] episodeRowColumnCount は 1 行が持つ列の数 (表示用話数・数値話数・タイトル)。タイトルは
// 最後に取り出して行の残り全体を受け取るため、カンマを含むタイトルも分割で欠けない。
const episodeRowColumnCount = 3

// episodeLimitedTextMaxLength caps the episodes columns declared as character varying(510):
// number and title. Without the cap, an over-long value passes validation and the INSERT
// fails with a driver error, so the submit ends as a 500 with every row lost. PostgreSQL
// counts varchar(n) in characters, so the check counts runes rather than bytes.
//
// [Ja] episodeLimitedTextMaxLength は character varying(510) で宣言された episodes のカラム
// (number / title) の上限。上限が無いと、長すぎる値がバリデーションを通ったあと INSERT が
// ドライバのエラーで失敗し、送信は全行を失ったまま 500 で終わる。PostgreSQL の varchar(n) は
// 文字数で数えるため、検査もバイト数ではなく文字数で行う。
const episodeLimitedTextMaxLength = 500

// episodeCreateMaxRows caps one bulk-create transaction. The form is intended for a season
// sized paste; one hundred rows still covers unusually long cours while bounding the row
// locks and database round trips caused by an accidental large paste.
//
// [Ja] episodeCreateMaxRows は 1 回の一括作成トランザクションの上限。フォームは 1 シーズン
// 程度の貼り付けを想定しており、100 行なら長めのクールも十分扱いつつ、誤って大量に貼り付けた
// ときの行ロックと DB 往復回数を制限できる。
const episodeCreateMaxRows = 100

// DBEpisodeCreateValidator validates the bulk-create form on the Annict DB admin screen,
// where a single textarea carries one episode per line.
//
// [Ja] DBEpisodeCreateValidator は Annict DB 管理画面の一括作成フォームを検証する。1 つの
// textarea に 1 行 1 エピソードの形で入力される。
type DBEpisodeCreateValidator struct{}

func NewDBEpisodeCreateValidator() *DBEpisodeCreateValidator {
	return &DBEpisodeCreateValidator{}
}

type DBEpisodeCreateValidatorInput struct {
	Rows string
}

// DBEpisodeRow is one submitted line after parsing. The fields mirror the nullable episodes
// columns the values are stored in (number / raw_number / title), so a column left empty
// travels as nil and is written as NULL instead of as an empty string that no existing row
// carries.
//
// [Ja] DBEpisodeRow は送信された 1 行をパースした結果。各フィールドは値の格納先である
// episodes の NULL 許容カラム (number / raw_number / title) に対応しており、空の列は nil の
// まま運ばれて NULL として書かれる。既存のどの行も持っていない空文字列にはしない。
type DBEpisodeRow struct {
	Number    *string
	RawNumber *float64
	Title     *string
}

// Validate parses the submitted lines and returns them in input order. Errors carry the
// line number they came from, counted over the raw input so it matches what the submitter
// sees in the textarea: blank lines are skipped as rows but still advance the count.
//
// A single bad line fails the whole submit. The rows are created in one transaction
// (nothing partial reaches the database), so returning the good rows alongside the errors
// would describe a state the caller can never produce.
//
// [Ja] Validate は送信された各行をパースし、入力順に返す。エラーには由来する行番号を付ける。
// 行番号は送信された入力そのものを数えるため、送信者が textarea で見ている番号と一致する
// (空行は行としては読み飛ばすが、番号は進める)。
//
// 1 行でも不正なら送信全体を失敗させる。行の作成は 1 トランザクションで行い部分的に DB へ
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

// splitEpisodeRowLines splits the submitted textarea into lines. A form submission encodes
// textarea line breaks as CRLF, and a lone CR can arrive from pasted content, so both are
// folded to LF first; splitting on LF alone would otherwise leave the CR inside the last
// column of every line.
//
// [Ja] splitEpisodeRowLines は送信された textarea を行に分割する。フォーム送信では textarea
// の改行が CRLF で符号化され、貼り付けた内容から単独の CR が届くこともあるため、まず双方を
// LF に畳む。LF だけで分割すると、各行の最後の列に CR が残ってしまう。
func splitEpisodeRowLines(rows string) []string {
	normalized := strings.ReplaceAll(rows, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")

	return strings.Split(normalized, "\n")
}

// parseEpisodeRow splits one line into its columns and checks them, returning the parsed
// row together with the messages for whatever failed. All checks run so that a line with
// two problems reports both instead of hiding one behind the other.
//
// [Ja] parseEpisodeRow は 1 行を列に分けて検査し、パース結果と失敗したチェックのメッセージを
// 返す。すべてのチェックを実行するため、2 箇所に問題がある行は片方がもう片方の陰に隠れず
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

// episodeRowColumn returns the column at the given position with the surrounding whitespace
// removed, and an empty string when the line stops before that position. A line may leave
// the trailing columns off entirely (a number with no title is submitted as "#1,1"), so a
// missing column means the same thing as one left empty.
//
// [Ja] episodeRowColumn は指定位置の列を前後の空白を取り除いて返し、行がその位置まで無い場合
// は空文字列を返す。行は末尾の列をまるごと省略できる (タイトルの無い話数は "#1,1" として
// 送信される) ため、列が無いことは列が空であることと同じ意味になる。
func episodeRowColumn(columns []string, index int) string {
	if index >= len(columns) {
		return ""
	}

	return strings.TrimSpace(columns[index])
}

// parseEpisodeRowRawNumber returns the numeric number of a row, and false when the column
// holds something that is not a number. An empty column is accepted and yields nil: the
// numeric number is optional for episodes that are not managed by number at all.
//
// NaN and the infinities are rejected even though they parse as floats. episodes.raw_number
// is a double precision column that stores them, so accepting them would put a value into
// the column that every later episode-number calculation reads as meaningless.
//
// [Ja] parseEpisodeRowRawNumber は行の数値話数を返し、数値でないものが入っている場合は false
// を返す。空の列は許可して nil を返す。数字で管理されていないエピソードのために数値話数は
// 任意入力であるため。
//
// NaN と無限大は float としてパースできるが弾く。episodes.raw_number は double precision
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

// DBEpisodeNullVersion is the explicit version an edit form carries when the stored
// updated_at is NULL. The column is nullable on the shared database, so "no timestamp" is a
// state the form has to be able to state; an empty request field means instead that no version
// was stated at all and is refused.
//
// The form renders the value and this package reads it back, so the literal lives here, on the
// side that decides whether a submitted version is acceptable.
//
// [Ja] DBEpisodeNullVersion は、保存済みの updated_at が NULL のとき編集フォームが運ぶ明示的な
// 版。共有 DB では同カラムが NULL 許容のため、「タイムスタンプが無い」ことをフォームが表明できる
// 必要がある。リクエストの値が空であることは、そもそも版が示されていないことを意味し、拒否する。
//
// フォームが値を描画し、本パッケージが読み戻すため、リテラルは、送信された版を受け入れるかを
// 判断する側であるここに置く。
const DBEpisodeNullVersion = "null"

// DBEpisodeVersionLayout is the format the version travels in. It keeps the sub-second digits
// that separate two writes made within the same second, and parses back to the instant it came
// from so the update can match it against the stored column.
//
// [Ja] DBEpisodeVersionLayout は版を往復させる書式。同一秒内の 2 つの書き込みを区別する秒未満の
// 桁を保ち、元の時刻へパースし直せるため、更新側は保存済みのカラムと照合できる。
const DBEpisodeVersionLayout = time.RFC3339Nano

// DBEpisodeUpdateValidator validates one submit of the episode edit form on the Annict DB
// admin screen.
//
// [Ja] DBEpisodeUpdateValidator は Annict DB 管理画面のエピソード編集フォームの 1 回の送信を
// 検証する。
type DBEpisodeUpdateValidator struct{}

func NewDBEpisodeUpdateValidator() *DBEpisodeUpdateValidator {
	return &DBEpisodeUpdateValidator{}
}

// DBEpisodeUpdateValidatorInput is the submitted form as it arrives: every field a string,
// including the version the hidden field carries.
//
// [Ja] DBEpisodeUpdateValidatorInput は届いたままの送信フォーム。hidden が運ぶ版を含め、
// すべてのフィールドが文字列。
type DBEpisodeUpdateValidatorInput struct {
	Number     string
	RawNumber  string
	SortNumber string
	Title      string
	TitleEn    string
	UpdatedAt  string
}

// DBEpisodeUpdateValidateOutput is the accepted submit converted to what the episodes columns
// hold. The optional text and number columns travel as pointers so a field the editor cleared
// is written as NULL rather than as an empty string no existing row carries.
//
// UpdatedAt is the version the submit was made against; nil is the DBEpisodeNullVersion
// sentinel, which the update matches with updated_at IS NULL.
//
// [Ja] DBEpisodeUpdateValidateOutput は受理された送信を、episodes の各カラムが保持する形へ
// 変換したもの。任意入力のテキストと数値のカラムはポインタで運び、編集者が消したフィールドを、
// 既存のどの行も持たない空文字列ではなく NULL として書けるようにする。
//
// UpdatedAt は送信が前提とする版。nil は DBEpisodeNullVersion のセンチネルで、更新側は
// updated_at IS NULL で照合する。
type DBEpisodeUpdateValidateOutput struct {
	Number     *string
	RawNumber  *float64
	Title      *string
	TitleEn    string
	SortNumber int32
	UpdatedAt  *time.Time
}

// Validate checks the submitted episode edit form and returns it converted to the stored
// column types. The numeric fields are checked against the exact ranges the conversion uses,
// so a value that gets here can be stored: one this method accepts but the conversion drops
// would be saved as NULL with no error shown, and one the column cannot hold would fail the
// UPDATE with a 500.
//
// The version is checked with the rest rather than before it, so a submit that is both stale
// and malformed reports everything the editor has to deal with at once.
//
// [Ja] Validate は送信されたエピソード編集フォームを検証し、保存先のカラムの型へ変換して返す。
// 数値フィールドは変換が使う範囲そのもので検査するため、ここを通った値は保存できる。本メソッドが
// 通して変換が落とす値は、エラーも出ないまま NULL で保存され、カラムに収まらない値は UPDATE が
// 失敗して 500 になるため。
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

	// The version is not an editable field, so a missing or malformed one is stated for the
	// form as a whole: there is no input for the editor to correct, only the page to reopen.
	//
	// [Ja] 版は編集できるフィールドではないため、欠落や形式不正はフォーム全体に対して述べる。
	// 編集者が直せる入力は無く、ページを開き直すしかないため。
	version, versionOK := parseDBEpisodeVersion(input.UpdatedAt)
	if !versionOK {
		ve.AddGlobal(i18n.T(ctx, "validation_version_missing"))
	}

	if ve.HasErrors() {
		return nil, ve
	}

	parsedSortNumber, err := strconv.ParseInt(sortNumber, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("並び順の変換に失敗: %w", err)
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

// parseDBEpisodeVersion reads the version a submit states, reporting false when the field is
// empty or holds neither the null sentinel nor a timestamp in the round-trip format. A nil
// time with true is the sentinel, which stands for a stored updated_at of NULL.
//
// [Ja] parseDBEpisodeVersion は送信が示す版を読み取り、フィールドが空か、null のセンチネルでも
// 往復書式のタイムスタンプでもない場合に false を返す。nil の時刻と true の組はセンチネルで、
// 保存済みの updated_at が NULL であることを表す。
func parseDBEpisodeVersion(value string) (*time.Time, bool) {
	switch value {
	case "":
		return nil, false
	case DBEpisodeNullVersion:
		return nil, true
	}

	parsed, err := time.Parse(DBEpisodeVersionLayout, value)
	if err != nil {
		return nil, false
	}

	return &parsed, true
}

// validateEpisodeMaxLength rejects a value longer than the episodes varchar(510) columns
// accept. It counts runes because PostgreSQL measures varchar(n) in characters; counting bytes
// would reject Japanese titles at a third of the real limit.
//
// [Ja] validateEpisodeMaxLength は episodes の varchar(510) カラムに収まらない長さの値を弾く。
// PostgreSQL の varchar(n) は文字数で数えるため文字数で数える。バイト数で数えると日本語の
// タイトルを実際の上限の 1/3 で弾いてしまう。
func validateEpisodeMaxLength(ctx context.Context, ve *model.ValidationError, field, value string) {
	if utf8.RuneCountInString(value) > episodeLimitedTextMaxLength {
		ve.AddField(field, i18n.T(ctx, "validation_too_long", map[string]any{
			"MaxLength": episodeLimitedTextMaxLength,
		}))
	}
}

// parseEpisodeOptionalFloat converts an already validated numeric number, returning nil for an
// empty field so the column is written as NULL.
//
// [Ja] parseEpisodeOptionalFloat は検証済みの数値話数を変換する。空のフィールドでは nil を返し、
// カラムが NULL として書かれるようにする。
func parseEpisodeOptionalFloat(value string) *float64 {
	parsed, _ := parseEpisodeRowRawNumber(value)
	return parsed
}

// optionalEpisodeRowText returns nil for an empty column so the value is stored as NULL.
//
// [Ja] optionalEpisodeRowText は空の列に対して nil を返し、値が NULL として格納されるように
// する。
func optionalEpisodeRowText(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}
