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

// workLimitedTextMaxLengthはcharacter varying(510) で宣言されたworksのカラム
// (title / official_site_url / wikipedia_url / twitter_username / twitter_hashtag) の
// 上限。上限が無いと、長すぎる値がバリデーションを通ったあとINSERTがドライバのエラーで
// 失敗し、送信は入力を失ったまま500で終わる。PostgreSQLのvarchar(n) は文字数で数えるため、
// 検査もバイト数ではなく文字数で行う。
const workLimitedTextMaxLength = 500

// WorkFormDateLayoutは作品フォームの日付入力欄が送信する日付形式。値を保存する変換が
// 同一の形式で解釈できるよう公開している。変換が受け付けない形式をバリデーターが許すと、
// 正しい送信が500になるため。
const WorkFormDateLayout = "2006-01-02"

// allowedMediaValuesは作品作成フォームで許可されるメディア種別コードの一覧。
// Rails版のworks.media enumと対応している (0=その他, 1=テレビ, 2=OVA, 3=映画, 4=Web)。
var allowedMediaValues = map[string]bool{
	"0": true,
	"1": true,
	"2": true,
	"3": true,
	"4": true,
}

// allowedSeasonNameValuesは作品フォームで許可される季節コードの一覧。
// Rails版のworks.season_nameのSeason::NAME_HASH enumと対応している
// (1=冬, 2=春, 3=夏, 4=秋)。
var allowedSeasonNameValues = map[string]bool{
	"1": true,
	"2": true,
	"3": true,
	"4": true,
}

// DBWorkCreateValidatorはAnnict DB管理画面の作品フォームを検証する。作成フォームと
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

	// ExcludeWorkIDは編集中のworkで、タイトルの一意性検査から除外する。タイトルを
	// 変えない更新が自分自身と衝突しないようにするため。除外するworkが無い作成フローでは
	// nilになる。
	ExcludeWorkID *model.WorkID

	// UpdatedAtは編集フォームを開いた時点の版で、hiddenフィールドが運ぶ形のまま。
	// 送信が古くなり得る保存済みの行が無い作成フローではnilになり、nilでない空文字列は版を
	// まったく示さなかった送信を意味し、拒否する。フローに依存する各フィールドは、
	// ExcludeWorkIDから推論させず自らの不在を表明する。一方を他方の代理として読まないため。
	UpdatedAt *string
}

// Validateは送信された作品フォームを検証し、送信が前提とする版を返す。数値・日付の
// フィールドはbuildWorkFormParamsが変換する範囲そのもので検査するため、ここを通った値は
// 保存できる。本メソッドが通して変換が落とす値は、エラーも出ないままNULLで保存され、カラムに
// 収まらない値はINSERTが失敗して500になるため。
//
// 返す版は、版を示さない作成フローと、編集フローのFormNullVersionのセンチネル (更新側は
// updated_at IS NULLで照合する) ではnilになる。hiddenの版の欠落・形式不正は先出しせず、
// 他の不正なフォーム値と一緒に報告する。正しい形式の版が古いかどうかは、バリデーション成功後の
// 更新条件で原子的に検査する。
//
// 形式チェックを先に行い、すべて通ったときだけDBを要するチェック (タイトルの一意性と
// 話数フォーマットの存在) を実行する。形式チェックが既に落ちている間はクエリを飛ばし、
// 送信者が対処できる問題だけを伝える。
func (v *DBWorkCreateValidator) Validate(ctx context.Context, input DBWorkCreateValidatorInput) (*time.Time, error) {
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

	// 版は編集できるフィールドではないため、欠落や形式不正はフォーム全体に対して述べる。
	// 編集者が直せる入力は無く、ページを開き直すしかないため。
	var version *time.Time
	if input.UpdatedAt != nil {
		parsed, ok := parseFormVersion(*input.UpdatedAt)
		if !ok {
			ve.AddGlobal(i18n.T(ctx, "validation_version_missing"))
		}
		version = parsed
	}

	if ve.HasErrors() {
		return nil, ve
	}

	taken, err := v.workRepo.ExistsKeptByTitle(ctx, title, input.ExcludeWorkID)
	if err != nil {
		return nil, fmt.Errorf("作品タイトルの重複確認に失敗: %w", err)
	}
	if taken {
		ve.AddField("title", i18n.T(ctx, "validation_work_title_already_taken"))
	}

	if err := v.validateNumberFormatExists(ctx, ve, input.NumberFormatID); err != nil {
		return nil, err
	}

	if ve.HasErrors() {
		return nil, ve
	}

	return version, nil
}

// validateNumberFormatExistsはnumber_formatsのどの行も指さないnumber_format_idを
// 弾く。works.number_format_idとanime_classifications.number_format_idはいずれも同表への
// 外部キーのため、整数チェックだけを通った値はINSERTで失敗し、送信は入力を失ったまま
// 500で終わる。selectは登録済みのフォーマットしか提示しないため、ここに届く値は
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

// parseOptionalInt64は任意入力の整数フィールドをパースした値を返し、空値または解釈
// できない場合はfalseを返す。呼び出し元はDBを要するチェックで、形式チェックがすべて
// 通った後にしか走らないため、ここでのfalseは「未入力なので照会するものが無い」ことを
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

// validateMaxLengthはworksのvarchar(510) カラムに収まらない長さの値を弾く。
// PostgreSQLのvarchar(n) は文字数で数えるため文字数で数える。バイト数で数えると日本語の
// タイトルを実際の上限の1/3で弾いてしまう。
func validateMaxLength(ctx context.Context, ve *model.ValidationError, field, value string) {
	if utf8.RuneCountInString(value) > workLimitedTextMaxLength {
		ve.AddField(field, i18n.T(ctx, "validation_too_long", map[string]any{"MaxLength": workLimitedTextMaxLength}))
	}
}

// validateOptionalInt32は空値を許し、それ以外はworksのintegerカラムに収まる整数を
// 要求する。ビット幅はbuildWorkFormParamsの変換と揃えてあり、int32を超える値は黙って
// 捨てられずここで報告される。
func validateOptionalInt32(ctx context.Context, ve *model.ValidationError, field, value string) {
	if value == "" {
		return
	}
	if _, err := strconv.ParseInt(value, 10, 32); err != nil {
		ve.AddField(field, i18n.T(ctx, "validation_integer_invalid"))
	}
}

// validateOptionalSeasonNameは空値、またはallowedSeasonNameValuesにあるコードを
// 許可する。整数形式だけでなく許可値も検査し、Railsが拒否してUIでも表現できない季節を
// 改変されたフォームから保存できないようにする。works.season_nameには拠り所となるCHECK
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

// validateOptionalInt64はbigintカラム (number_format_id) 向けの
// validateOptionalInt32。
func validateOptionalInt64(ctx context.Context, ve *model.ValidationError, field, value string) {
	if value == "" {
		return
	}
	if _, err := strconv.ParseInt(value, 10, 64); err != nil {
		ve.AddField(field, i18n.T(ctx, "validation_integer_invalid"))
	}
}

// validateOptionalFloatは空値を許し、それ以外は有限の数値を要求する。
// start_episode_raw_numberのdouble precisionカラムに合わせている。NaNと無限大はfloat
// としてパースできるが弾く。double precisionはこれらを格納できるため、受け入れると以降の
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

// validateOptionalDateは空値を許し、それ以外は日付入力欄が送信する日付形式を要求する。
func validateOptionalDate(ctx context.Context, ve *model.ValidationError, field, value string) {
	if value == "" {
		return
	}
	if _, err := time.Parse(WorkFormDateLayout, value); err != nil {
		ve.AddField(field, i18n.T(ctx, "validation_date_invalid"))
	}
}

// validatePresencePairは対になる2フィールドのうち、contentが入力されているときに
// sourceも必須にする。あらすじと出典のように対で意味を持つ入力で使い、片方だけ埋まった
// 中途半端なレコードを防ぐ。
func validatePresencePair(ctx context.Context, ve *model.ValidationError, sourceField, content, source, errKey string) {
	if strings.TrimSpace(content) != "" && strings.TrimSpace(source) == "" {
		ve.AddField(sourceField, i18n.T(ctx, errKey))
	}
}
