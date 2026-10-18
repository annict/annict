package validator

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// personRowSeparatorは送信された1行を列に分ける文字。Rails版の一括登録フォームが
// 案内してきたカンマで、編集者がAnnictの外に持っている行をそのまま貼り付けられる。
const personRowSeparator = ","

// personRowColumnCountは1行が持てる列の数 (名前・読み)。
const personRowColumnCount = 2

// personCreateMaxRowsは1回の一括登録トランザクションの上限。エピソードの一括作成
// (episodeCreateMaxRows) と同じく、誤って大量に貼り付けたときの行ロックとDB往復回数を
// 制限する。
const personCreateMaxRows = 100

// DBPersonCreateValidatorはAnnict DB管理画面の人物の一括登録フォームを検証する。1つの
// textareaに1行1人の形で入力される。
type DBPersonCreateValidator struct {
	personRepo *repository.PersonRepository
}

func NewDBPersonCreateValidator(personRepo *repository.PersonRepository) *DBPersonCreateValidator {
	return &DBPersonCreateValidator{personRepo: personRepo}
}

type DBPersonCreateValidatorInput struct {
	Rows string
}

// DBPersonRowは送信された1行をパースした結果。読みはpeople.name_kanaがNOT NULLのため、
// 空のときも空文字列のまま運ぶ (Rails版と同じ)。
type DBPersonRow struct {
	Name     string
	NameKana string
}

// parsedPersonRowは検証中の1行。行番号とその行で失敗したチェックのメッセージを持つ。
type parsedPersonRow struct {
	line     int
	row      DBPersonRow
	messages []string
}

// Validateは送信された各行をパースし、入力順に返す。エラーには由来する行番号を付ける。
// 行番号は送信された入力そのものを数えるため、送信者がtextareaで見ている番号と一致する
// (空行は行としては読み飛ばすが、番号は進める)。
//
// 1行でも不正なら送信全体を失敗させる (DBEpisodeCreateValidatorと同じ)。名前の重複は、
// 入力の中での重複と既存の人物との重複の両方を確かめる。書式の誤りと名前の重複を一度に
// 報告できるよう、既存の人物との照合は書式を通った行について行い、メッセージは行の順に並べる。
func (v *DBPersonCreateValidator) Validate(ctx context.Context, input DBPersonCreateValidatorInput) ([]DBPersonRow, error) {
	ve := model.NewValidationError()

	lines := splitFormRowLines(input.Rows)
	nonBlankRowCount := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonBlankRowCount++
		}
	}
	if nonBlankRowCount > personCreateMaxRows {
		ve.AddField("rows", i18n.T(ctx, "validation_person_rows_too_many", map[string]any{
			"MaxRows": personCreateMaxRows,
		}))
		return nil, ve
	}

	parsedRows := make([]*parsedPersonRow, 0, nonBlankRowCount)
	firstLineByName := make(map[string]int, nonBlankRowCount)
	names := make([]string, 0, nonBlankRowCount)
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		row, messages := parsePersonRow(ctx, line)
		parsed := &parsedPersonRow{line: i + 1, row: row, messages: messages}
		parsedRows = append(parsedRows, parsed)

		// 書式が不正な行は既存名との照合対象にせず、後続の正常な行の重複判定も占有しない。
		if len(messages) > 0 {
			continue
		}
		if firstLine, ok := firstLineByName[row.Name]; ok {
			parsed.messages = append(parsed.messages, i18n.T(ctx, "validation_person_row_name_duplicated", map[string]any{
				"Line": firstLine,
			}))
			continue
		}
		firstLineByName[row.Name] = parsed.line
		names = append(names, row.Name)
	}

	if len(parsedRows) == 0 {
		ve.AddField("rows", i18n.T(ctx, "validation_required"))
		return nil, ve
	}

	holders, err := v.personRepo.ListNameHolders(ctx, names)
	if err != nil {
		return nil, fmt.Errorf("人物名の重複確認に失敗: %w", err)
	}
	holderStatusByName := make(map[string]model.PersonStatus, len(holders))
	for _, holder := range holders {
		holderStatusByName[holder.Name] = holder.Status
	}

	rows := make([]DBPersonRow, 0, len(parsedRows))
	for _, parsed := range parsedRows {
		// 既存の人物と重なる名前は、その名前が最初に現れた行にだけ報告する。2回目以降の行は
		// 入力の中での重複として報告済みのため。
		if status, taken := holderStatusByName[parsed.row.Name]; taken && firstLineByName[parsed.row.Name] == parsed.line {
			parsed.messages = append(parsed.messages, personNameTakenMessage(ctx, status, "validation_person_row_name_already_taken"))
		}

		for _, message := range parsed.messages {
			ve.AddField("rows", i18n.T(ctx, "validation_person_row_error", map[string]any{
				"Line":    parsed.line,
				"Message": message,
			}))
		}
		rows = append(rows, parsed.row)
	}

	if ve.HasErrors() {
		return nil, ve
	}

	return rows, nil
}

// parsePersonRowは1行を列に分けて検査し、パース結果と失敗したチェックのメッセージを
// 返す。
//
// ダブルクォートは列を囲む記号として扱わず、文字としてそのまま残す (Rails版の
// ResourceRows#parsed_rowsと同じ)。既存の人物名には「野中"まさ"雄一」のように愛称を
// ダブルクォートで囲んだものがあり、CSVの引用符として解釈すると、これらをそのまま貼り付けた
// 行が書式の誤りになるか、別の名前に変わってしまうため。
func parsePersonRow(ctx context.Context, line string) (DBPersonRow, []string) {
	columns := strings.Split(line, personRowSeparator)
	for i := range columns {
		columns[i] = strings.TrimSpace(columns[i])
	}

	var messages []string

	name := columns[0]
	if name == "" {
		messages = append(messages, i18n.T(ctx, "validation_person_row_name_required"))
	}

	// 3列目以降に値があれば、黙って捨てずに誤りとして返す。Rails版は捨てていたが、
	// 英語名などを書き足した行が、書き足した値を失ったまま登録されてしまうため。末尾の
	// 余分なカンマのような空の列は許す。
	for _, column := range columns[min(len(columns), personRowColumnCount):] {
		if column != "" {
			messages = append(messages, i18n.T(ctx, "validation_person_row_too_many_columns", map[string]any{
				"MaxColumns": personRowColumnCount,
			}))
			break
		}
	}

	nameKana := ""
	if len(columns) > 1 {
		nameKana = columns[1]
	}

	return DBPersonRow{Name: name, NameKana: nameKana}, messages
}

// DBPersonUpdateValidatorはAnnict DB管理画面の人物編集フォームの1回の送信を検証する。
type DBPersonUpdateValidator struct {
	personRepo     *repository.PersonRepository
	prefectureRepo *repository.PrefectureRepository
}

func NewDBPersonUpdateValidator(
	personRepo *repository.PersonRepository,
	prefectureRepo *repository.PrefectureRepository,
) *DBPersonUpdateValidator {
	return &DBPersonUpdateValidator{personRepo: personRepo, prefectureRepo: prefectureRepo}
}

// DBPersonUpdateValidatorInputは届いたままの送信フォーム。hiddenが運ぶ版を含め、
// 送信された値はすべて文字列。
type DBPersonUpdateValidatorInput struct {
	Name              string
	NameKana          string
	NameEn            string
	Nickname          string
	NicknameEn        string
	Gender            string
	BloodType         string
	PrefectureID      string
	Birthday          string
	Height            string
	URL               string
	URLEn             string
	WikipediaURL      string
	WikipediaURLEn    string
	TwitterUsername   string
	TwitterUsernameEn string
	UpdatedAt         string
}

// DBPersonUpdateValidateOutputは受理された送信を、peopleの各カラムが保持する形へ変換した
// もの。NULL許容のカラムはポインタで運び、編集者が空にした欄をNULLとして書けるようにする。
// NOT NULLの文字列カラム (英語の各項目と読み) は空文字列のまま運ぶ。
type DBPersonUpdateValidateOutput struct {
	Name              string
	NameKana          string
	NameEn            string
	Nickname          *string
	NicknameEn        string
	Gender            *model.PersonGender
	BloodType         *model.PersonBloodType
	PrefectureID      *model.PrefectureID
	Birthday          *time.Time
	Height            *int32
	URL               *string
	URLEn             string
	WikipediaURL      *string
	WikipediaURLEn    string
	TwitterUsername   *string
	TwitterUsernameEn string
	// UpdatedAtは送信が前提とする版。people.updated_atはNOT NULLのため、NULLの版
	// (FormNullVersion) は受け付けない。
	UpdatedAt time.Time
}

// ValidateFormatはDBを読まずに形式を検証する。版の欠落・形式不正も他の項目と
// 一緒に報告する。UseCaseはこの後、名前の重複確認より先に保存済みの版を照合する。
func (v *DBPersonUpdateValidator) ValidateFormat(ctx context.Context, input DBPersonUpdateValidatorInput) (*DBPersonUpdateValidateOutput, error) {
	ve := model.NewValidationError()

	name := strings.TrimSpace(input.Name)
	gender := strings.TrimSpace(input.Gender)
	bloodType := strings.TrimSpace(input.BloodType)
	prefectureID := strings.TrimSpace(input.PrefectureID)
	birthday := strings.TrimSpace(input.Birthday)
	height := strings.TrimSpace(input.Height)
	url := strings.TrimSpace(input.URL)
	urlEn := strings.TrimSpace(input.URLEn)
	wikipediaURL := strings.TrimSpace(input.WikipediaURL)
	wikipediaURLEn := strings.TrimSpace(input.WikipediaURLEn)

	if name == "" {
		ve.AddField("name", i18n.T(ctx, "validation_required"))
	}
	if gender != "" && !model.PersonGender(gender).IsValid() {
		ve.AddField("gender", i18n.T(ctx, "validation_person_gender_invalid"))
	}
	if bloodType != "" && !model.PersonBloodType(bloodType).IsValid() {
		ve.AddField("blood_type", i18n.T(ctx, "validation_person_blood_type_invalid"))
	}
	validateOptionalInt64(ctx, ve, "prefecture_id", prefectureID)
	validateOptionalDate(ctx, ve, "birthday", birthday)
	validateOptionalInt32(ctx, ve, "height", height)
	validateOptionalURL(ctx, ve, "url", url)
	validateOptionalURL(ctx, ve, "url_en", urlEn)
	validateOptionalURL(ctx, ve, "wikipedia_url", wikipediaURL)
	validateOptionalURL(ctx, ve, "wikipedia_url_en", wikipediaURLEn)

	// 版は編集できるフィールドではないため、欠落や形式不正はフォーム全体に対して述べる。
	version, versionOK := parseFormVersion(input.UpdatedAt)
	if !versionOK || version == nil {
		ve.AddGlobal(i18n.T(ctx, "validation_version_missing"))
	}

	if ve.HasErrors() {
		return nil, ve
	}

	output := &DBPersonUpdateValidateOutput{
		Name:              name,
		NameKana:          strings.TrimSpace(input.NameKana),
		NameEn:            strings.TrimSpace(input.NameEn),
		Nickname:          optionalPersonText(input.Nickname),
		NicknameEn:        strings.TrimSpace(input.NicknameEn),
		URL:               optionalPersonText(url),
		URLEn:             urlEn,
		WikipediaURL:      optionalPersonText(wikipediaURL),
		WikipediaURLEn:    wikipediaURLEn,
		TwitterUsername:   optionalPersonText(input.TwitterUsername),
		TwitterUsernameEn: strings.TrimSpace(input.TwitterUsernameEn),
		UpdatedAt:         *version,
	}
	if gender != "" {
		value := model.PersonGender(gender)
		output.Gender = &value
	}
	if bloodType != "" {
		value := model.PersonBloodType(bloodType)
		output.BloodType = &value
	}
	if id, ok := parseOptionalInt64(prefectureID); ok {
		value := model.PrefectureID(id)
		output.PrefectureID = &value
	}
	if birthday != "" {
		parsed, err := time.Parse(PersonFormDateLayout, birthday)
		if err != nil {
			return nil, fmt.Errorf("誕生日の変換に失敗: %w", err)
		}
		output.Birthday = &parsed
	}
	if height != "" {
		parsed, err := strconv.ParseInt(height, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("身長の変換に失敗: %w", err)
		}
		value := int32(parsed)
		output.Height = &value
	}

	return output, nil
}

// ValidateDatabaseは形式を通った値について、DBを要する名前と出身地の検証を行う。personIDは
// 編集中の人物で、名前の一意性の検査から除外する。名前を変えない更新が自分自身と衝突しない
// ようにするため。
func (v *DBPersonUpdateValidator) ValidateDatabase(ctx context.Context, personID model.PersonID, output *DBPersonUpdateValidateOutput) (*DBPersonUpdateValidateOutput, error) {
	ve := model.NewValidationError()
	holder, err := v.personRepo.GetNameHolderExcludingID(ctx, output.Name, personID)
	if err != nil {
		return nil, fmt.Errorf("人物名の重複確認に失敗: %w", err)
	}
	if holder != nil {
		ve.AddField("name", personNameTakenMessage(ctx, holder.Status, "validation_person_name_already_taken"))
	}
	if output.PrefectureID != nil {
		exists, err := v.prefectureRepo.ExistsByID(ctx, *output.PrefectureID)
		if err != nil {
			return nil, fmt.Errorf("出身地の存在確認に失敗: %w", err)
		}
		if !exists {
			ve.AddField("prefecture_id", i18n.T(ctx, "validation_person_prefecture_invalid"))
		}
	}
	if ve.HasErrors() {
		return nil, ve
	}
	return output, nil
}

// personNameTakenMessageは、名前がholderStatusの人物と重複したときのエラーメッセージを返す。
// 掲載停止 (delisted) の人物は一覧に出ないため、通常の重複のメッセージでは編集者が理由を
// 掴めず登録を試し続ける。掲載停止であることだけを伝え、申出の有無などの経緯は伏せる。
func personNameTakenMessage(ctx context.Context, holderStatus model.PersonStatus, takenMessageID string) string {
	if holderStatus == model.PersonStatusDelisted {
		return i18n.T(ctx, "validation_person_name_delisted")
	}
	return i18n.T(ctx, takenMessageID)
}

// PersonFormDateLayoutは人物編集フォームの誕生日の入力欄が送信する日付形式。作品フォームの
// 日付入力欄と同じく、ブラウザーのtype="date" が送る形式。
const PersonFormDateLayout = WorkFormDateLayout

// optionalPersonTextは前後の空白を除いた値を返し、空の欄に対してはnilを返す。NULL許容の
// カラムに、既存の行がNULLとして持つ「未設定」を空文字列で書かないようにするため。
func optionalPersonText(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
