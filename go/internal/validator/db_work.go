package validator

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// workLimitedTextMaxLength caps the works columns declared as character varying(510):
// title / official_site_url / wikipedia_url / twitter_username / twitter_hashtag. Without
// the cap, an over-long value passes validation and the INSERT fails with a driver error,
// so the submit ends as a 500 with the input lost. PostgreSQL counts varchar(n) in
// characters, so the check counts runes rather than bytes.
//
// [Ja] workLimitedTextMaxLength は character varying(510) で宣言された works のカラム
// (title / official_site_url / wikipedia_url / twitter_username / twitter_hashtag) の
// 上限。上限が無いと、長すぎる値がバリデーションを通ったあと INSERT がドライバのエラーで
// 失敗し、送信は入力を失ったまま 500 で終わる。PostgreSQL の varchar(n) は文字数で数えるため、
// 検査もバイト数ではなく文字数で行う。
const workLimitedTextMaxLength = 500

// WorkFormDateLayout is the date format the work form's date inputs submit. It is exported
// so the conversion that stores the value parses with the very same layout: a validator
// that accepts a format the conversion rejects would turn a valid submit into a 500.
//
// [Ja] WorkFormDateLayout は作品フォームの日付入力欄が送信する日付形式。値を保存する変換が
// 同一の形式で解釈できるよう公開している。変換が受け付けない形式をバリデーターが許すと、
// 正しい送信が 500 になるため。
const WorkFormDateLayout = "2006-01-02"

// allowedMediaValues lists the media type codes accepted by the create-work form.
// The mapping mirrors the Rails enum on the works.media column
// (0=other, 1=tv, 2=ova, 3=movie, 4=web).
//
// [Ja] allowedMediaValues は作品作成フォームで許可されるメディア種別コードの一覧。
// Rails 版の works.media enum と対応している (0=その他, 1=テレビ, 2=OVA, 3=映画, 4=Web)。
var allowedMediaValues = map[string]bool{
	"0": true,
	"1": true,
	"2": true,
	"3": true,
	"4": true,
}

// allowedSeasonNameValues lists the season codes accepted by the work form. The mapping
// mirrors the Rails Season::NAME_HASH enum on the works.season_name column
// (1=winter, 2=spring, 3=summer, 4=autumn).
//
// [Ja] allowedSeasonNameValues は作品フォームで許可される季節コードの一覧。
// Rails 版の works.season_name の Season::NAME_HASH enum と対応している
// (1=冬, 2=春, 3=夏, 4=秋)。
var allowedSeasonNameValues = map[string]bool{
	"1": true,
	"2": true,
	"3": true,
	"4": true,
}

// DBWorkCreateValidator validates the work form on the Annict DB admin screen. The create
// and edit forms submit the same fields, so both flows run through it.
//
// [Ja] DBWorkCreateValidator は Annict DB 管理画面の作品フォームを検証する。作成フォームと
// 編集フォームは同一のフィールドを送信するため、双方のフローが本バリデーターを通る。
type DBWorkCreateValidator struct {
	workRepo         *repository.WorkRepository
	numberFormatRepo *repository.NumberFormatRepository
}

func NewDBWorkCreateValidator(
	workRepo *repository.WorkRepository,
	numberFormatRepo *repository.NumberFormatRepository,
) *DBWorkCreateValidator {
	return &DBWorkCreateValidator{workRepo: workRepo, numberFormatRepo: numberFormatRepo}
}

type DBWorkCreateValidatorInput struct {
	Title                 string
	TitleKana             string
	TitleAlter            string
	TitleEn               string
	TitleAlterEn          string
	Media                 string
	SeasonYear            string
	SeasonName            string
	StartedOn             string
	EndedOn               string
	OfficialSiteURL       string
	OfficialSiteURLEn     string
	WikipediaURL          string
	WikipediaURLEn        string
	TwitterUsername       string
	TwitterHashtag        string
	ScTid                 string
	MalAnimeID            string
	Synopsis              string
	SynopsisSource        string
	SynopsisEn            string
	SynopsisSourceEn      string
	ManualEpisodesCount   string
	StartEpisodeRawNumber string
	NumberFormatID        string
	NoEpisodes            string

	// ExcludeWorkID is the work being edited, excluded from the title uniqueness check so
	// an update that leaves the title untouched does not collide with itself. It is nil on
	// the create flow, where there is no work to exclude.
	//
	// [Ja] ExcludeWorkID は編集中の work で、タイトルの一意性検査から除外する。タイトルを
	// 変えない更新が自分自身と衝突しないようにするため。除外する work が無い作成フローでは
	// nil になる。
	ExcludeWorkID *model.WorkID
}

// Validate checks the submitted work form. The numeric and date fields are checked against
// the exact ranges buildWorkFormParams converts them with, so a value that gets here can be
// stored: a value this method accepts but the conversion drops would be saved as NULL with
// no error shown, and one the column cannot hold would fail the INSERT with a 500.
//
// The format checks run first and, when they all pass, the checks that need the database
// run: the title uniqueness and the existence of the number format. Skipping the queries
// while a format check has already failed keeps the message on the problem the submitter
// can act on.
//
// [Ja] Validate は送信された作品フォームを検証する。数値・日付のフィールドは
// buildWorkFormParams が変換する範囲そのもので検査するため、ここを通った値は保存できる。
// 本メソッドが通して変換が落とす値は、エラーも出ないまま NULL で保存され、カラムに収まらない
// 値は INSERT が失敗して 500 になるため。
//
// 形式チェックを先に行い、すべて通ったときだけ DB を要するチェック (タイトルの一意性と
// 話数フォーマットの存在) を実行する。形式チェックが既に落ちている間はクエリを飛ばし、
// 送信者が対処できる問題だけを伝える。
func (v *DBWorkCreateValidator) Validate(ctx context.Context, input DBWorkCreateValidatorInput) error {
	ve := model.NewValidationError()

	title := strings.TrimSpace(input.Title)
	if title == "" {
		ve.AddField("title", i18n.T(ctx, "validation_required"))
	}
	validateMaxLength(ctx, ve, "title", title)

	if strings.TrimSpace(input.Media) == "" {
		ve.AddField("media", i18n.T(ctx, "validation_required"))
	} else if !allowedMediaValues[input.Media] {
		ve.AddField("media", i18n.T(ctx, "validation_media_invalid"))
	}

	validateOptionalInt32(ctx, ve, "season_year", input.SeasonYear)
	validateOptionalSeasonName(ctx, ve, input.SeasonName)
	validateOptionalDate(ctx, ve, "started_on", input.StartedOn)
	validateOptionalDate(ctx, ve, "ended_on", input.EndedOn)

	validateOptionalURL(ctx, ve, "official_site_url", input.OfficialSiteURL)
	validateOptionalURL(ctx, ve, "official_site_url_en", input.OfficialSiteURLEn)
	validateOptionalURL(ctx, ve, "wikipedia_url", input.WikipediaURL)
	validateOptionalURL(ctx, ve, "wikipedia_url_en", input.WikipediaURLEn)
	validateMaxLength(ctx, ve, "official_site_url", strings.TrimSpace(input.OfficialSiteURL))
	validateMaxLength(ctx, ve, "wikipedia_url", strings.TrimSpace(input.WikipediaURL))
	validateMaxLength(ctx, ve, "twitter_username", strings.TrimSpace(input.TwitterUsername))
	validateMaxLength(ctx, ve, "twitter_hashtag", strings.TrimSpace(input.TwitterHashtag))

	validateOptionalInt32(ctx, ve, "sc_tid", input.ScTid)
	validateOptionalInt32(ctx, ve, "mal_anime_id", input.MalAnimeID)

	validatePresencePair(ctx, ve, "synopsis_source", input.Synopsis, input.SynopsisSource, "validation_synopsis_source_required")
	validatePresencePair(ctx, ve, "synopsis_source_en", input.SynopsisEn, input.SynopsisSourceEn, "validation_synopsis_source_en_required")

	validateOptionalInt32(ctx, ve, "manual_episodes_count", input.ManualEpisodesCount)
	validateOptionalFloat(ctx, ve, "start_episode_raw_number", input.StartEpisodeRawNumber)
	validateOptionalInt64(ctx, ve, "number_format_id", input.NumberFormatID)

	if ve.HasErrors() {
		return ve
	}

	taken, err := v.workRepo.ExistsKeptByTitle(ctx, title, input.ExcludeWorkID)
	if err != nil {
		return fmt.Errorf("作品タイトルの重複確認に失敗: %w", err)
	}
	if taken {
		ve.AddField("title", i18n.T(ctx, "validation_work_title_already_taken"))
	}

	if err := v.validateNumberFormatExists(ctx, ve, input.NumberFormatID); err != nil {
		return err
	}

	if ve.HasErrors() {
		return ve
	}

	return nil
}

// validateNumberFormatExists rejects a number_format_id that names no row in
// number_formats. works.number_format_id and anime_classifications.number_format_id are
// both foreign keys to that table, so a value that only passed the integer check fails the
// INSERT and ends the submit as a 500 with the input lost. The select offers the registered
// formats only, so a value that gets here was altered after the page was rendered.
//
// [Ja] validateNumberFormatExists は number_formats のどの行も指さない number_format_id を
// 弾く。works.number_format_id と anime_classifications.number_format_id はいずれも同表への
// 外部キーのため、整数チェックだけを通った値は INSERT で失敗し、送信は入力を失ったまま
// 500 で終わる。select は登録済みのフォーマットしか提示しないため、ここに届く値は
// ページ描画後に改変されたものである。
func (v *DBWorkCreateValidator) validateNumberFormatExists(ctx context.Context, ve *model.ValidationError, value string) error {
	id, ok := parseOptionalInt64(value)
	if !ok {
		return nil
	}

	exists, err := v.numberFormatRepo.ExistsByID(ctx, model.NumberFormatID(id))
	if err != nil {
		return fmt.Errorf("話数フォーマットの存在確認に失敗: %w", err)
	}
	if !exists {
		ve.AddField("number_format_id", i18n.T(ctx, "validation_number_format_invalid"))
	}

	return nil
}

// parseOptionalInt64 returns the parsed value of an optional integer field, and false when
// the field is empty or unparsable. Callers are the checks that need the database, which
// run only once every format check has passed, so false here means the field was left
// blank and there is nothing to look up.
//
// [Ja] parseOptionalInt64 は任意入力の整数フィールドをパースした値を返し、空値または解釈
// できない場合は false を返す。呼び出し元は DB を要するチェックで、形式チェックがすべて
// 通った後にしか走らないため、ここでの false は「未入力なので照会するものが無い」ことを
// 意味する。
func parseOptionalInt64(value string) (int64, bool) {
	if value == "" {
		return 0, false
	}

	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, false
	}

	return v, true
}

func validateOptionalURL(ctx context.Context, ve *model.ValidationError, field, value string) {
	if value == "" {
		return
	}
	u, err := url.ParseRequestURI(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		ve.AddField(field, i18n.T(ctx, "validation_url_invalid"))
	}
}

// validateMaxLength rejects a value longer than the works varchar(510) columns accept.
// It counts runes because PostgreSQL measures varchar(n) in characters; counting bytes
// would reject Japanese titles at a third of the real limit.
//
// [Ja] validateMaxLength は works の varchar(510) カラムに収まらない長さの値を弾く。
// PostgreSQL の varchar(n) は文字数で数えるため文字数で数える。バイト数で数えると日本語の
// タイトルを実際の上限の 1/3 で弾いてしまう。
func validateMaxLength(ctx context.Context, ve *model.ValidationError, field, value string) {
	if utf8.RuneCountInString(value) > workLimitedTextMaxLength {
		ve.AddField(field, i18n.T(ctx, "validation_too_long", map[string]any{"MaxLength": workLimitedTextMaxLength}))
	}
}

// validateOptionalInt32 accepts an empty value and otherwise requires an integer that fits
// the works integer columns. The bit size matches the conversion in buildWorkFormParams,
// so a value beyond int32 is reported here instead of being silently dropped.
//
// [Ja] validateOptionalInt32 は空値を許し、それ以外は works の integer カラムに収まる整数を
// 要求する。ビット幅は buildWorkFormParams の変換と揃えてあり、int32 を超える値は黙って
// 捨てられずここで報告される。
func validateOptionalInt32(ctx context.Context, ve *model.ValidationError, field, value string) {
	if value == "" {
		return
	}
	if _, err := strconv.ParseInt(value, 10, 32); err != nil {
		ve.AddField(field, i18n.T(ctx, "validation_integer_invalid"))
	}
}

// validateOptionalSeasonName accepts an empty value or one of the codes in
// allowedSeasonNameValues. Checking the allowed set as well as integer syntax prevents a
// crafted form submission from storing a season value that Rails rejects and the UI cannot
// represent; works.season_name carries no CHECK constraint to fall back on.
//
// [Ja] validateOptionalSeasonName は空値、または allowedSeasonNameValues にあるコードを
// 許可する。整数形式だけでなく許可値も検査し、Rails が拒否して UI でも表現できない季節を
// 改変されたフォームから保存できないようにする。works.season_name には拠り所となる CHECK
// 制約が無いため。
func validateOptionalSeasonName(ctx context.Context, ve *model.ValidationError, value string) {
	if value == "" {
		return
	}
	if _, err := strconv.ParseInt(value, 10, 32); err != nil {
		ve.AddField("season_name", i18n.T(ctx, "validation_integer_invalid"))
		return
	}
	if !allowedSeasonNameValues[value] {
		ve.AddField("season_name", i18n.T(ctx, "validation_season_name_invalid"))
	}
}

// validateOptionalInt64 is validateOptionalInt32 for the bigint columns (number_format_id).
//
// [Ja] validateOptionalInt64 は bigint カラム (number_format_id) 向けの
// validateOptionalInt32。
func validateOptionalInt64(ctx context.Context, ve *model.ValidationError, field, value string) {
	if value == "" {
		return
	}
	if _, err := strconv.ParseInt(value, 10, 64); err != nil {
		ve.AddField(field, i18n.T(ctx, "validation_integer_invalid"))
	}
}

// validateOptionalFloat accepts an empty value and otherwise requires a finite number,
// matching the double precision column behind start_episode_raw_number. NaN and the
// infinities are rejected even though they parse as floats: double precision stores them,
// so accepting them would put a value into the column that every later episode-number
// calculation reads as meaningless.
//
// [Ja] validateOptionalFloat は空値を許し、それ以外は有限の数値を要求する。
// start_episode_raw_number の double precision カラムに合わせている。NaN と無限大は float
// としてパースできるが弾く。double precision はこれらを格納できるため、受け入れると以降の
// 話数計算がすべて意味を成さなくなる値がカラムに入ってしまう。
func validateOptionalFloat(ctx context.Context, ve *model.ValidationError, field, value string) {
	if value == "" {
		return
	}
	v, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		ve.AddField(field, i18n.T(ctx, "validation_number_invalid"))
	}
}

// validateOptionalDate accepts an empty value and otherwise requires the date format the
// date inputs submit.
//
// [Ja] validateOptionalDate は空値を許し、それ以外は日付入力欄が送信する日付形式を要求する。
func validateOptionalDate(ctx context.Context, ve *model.ValidationError, field, value string) {
	if value == "" {
		return
	}
	if _, err := time.Parse(WorkFormDateLayout, value); err != nil {
		ve.AddField(field, i18n.T(ctx, "validation_date_invalid"))
	}
}

// validatePresencePair requires the source field whenever the content field is filled in.
// Used for paired inputs like a synopsis and its citation, where filling one half
// without the other would leave a half-completed record.
//
// [Ja] validatePresencePair は対になる 2 フィールドのうち、content が入力されているときに
// source も必須にする。あらすじと出典のように対で意味を持つ入力で使い、片方だけ埋まった
// 中途半端なレコードを防ぐ。
func validatePresencePair(ctx context.Context, ve *model.ValidationError, sourceField, content, source, errKey string) {
	if strings.TrimSpace(content) != "" && strings.TrimSpace(source) == "" {
		ve.AddField(sourceField, i18n.T(ctx, errKey))
	}
}
