package usecase

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/validator"
)

// WorkFormInput carries the Annict DB work form values shared by the create and update
// flows. Every field is a string because it comes straight from the HTML form; the
// typed conversion lives in buildWorkFormParams and the validation in toValidatorInput,
// so create and update stay single-sourced and cannot drift when a column is added.
//
// [Ja] WorkFormInput は Annict DB 作品フォームの入力値を、作成と更新の両フローで共有する。
// 各フィールドは HTML フォーム由来のため文字列で持つ。型変換は buildWorkFormParams、検証は
// toValidatorInput に集約し、作成と更新の写像・検証の正本を 1 つに保つ。カラム追加時に
// 片方だけ追従してドリフトする事故を防ぐ。
type WorkFormInput struct {
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
}

// toValidatorInput projects the form values onto the validator input. Create and update
// validate the identical set of fields, so both go through this single mapping.
// excludeWorkID is the only difference between the two flows: update passes the work being
// edited so the title uniqueness check does not match it against itself, create passes nil.
// Taking it as an argument makes both call sites state which flow they are.
//
// [Ja] toValidatorInput はフォーム値をバリデーター入力に射影する。作成と更新は同一の
// フィールド集合を検証するため、双方ともこの単一の写像を通す。両フローの唯一の違いが
// excludeWorkID で、更新は編集中の work を渡してタイトルの一意性検査が自分自身に一致しない
// ようにし、作成は nil を渡す。引数で受け取ることで、両方の呼び出し側がどちらのフローかを
// 明示する。
func (in WorkFormInput) toValidatorInput(excludeWorkID *model.WorkID) validator.DBWorkCreateValidatorInput {
	return validator.DBWorkCreateValidatorInput{
		ExcludeWorkID:         excludeWorkID,
		Title:                 in.Title,
		TitleKana:             in.TitleKana,
		TitleAlter:            in.TitleAlter,
		TitleEn:               in.TitleEn,
		TitleAlterEn:          in.TitleAlterEn,
		Media:                 in.Media,
		SeasonYear:            in.SeasonYear,
		SeasonName:            in.SeasonName,
		StartedOn:             in.StartedOn,
		EndedOn:               in.EndedOn,
		OfficialSiteURL:       in.OfficialSiteURL,
		OfficialSiteURLEn:     in.OfficialSiteURLEn,
		WikipediaURL:          in.WikipediaURL,
		WikipediaURLEn:        in.WikipediaURLEn,
		TwitterUsername:       in.TwitterUsername,
		TwitterHashtag:        in.TwitterHashtag,
		ScTid:                 in.ScTid,
		MalAnimeID:            in.MalAnimeID,
		Synopsis:              in.Synopsis,
		SynopsisSource:        in.SynopsisSource,
		SynopsisEn:            in.SynopsisEn,
		SynopsisSourceEn:      in.SynopsisSourceEn,
		ManualEpisodesCount:   in.ManualEpisodesCount,
		StartEpisodeRawNumber: in.StartEpisodeRawNumber,
		NumberFormatID:        in.NumberFormatID,
		NoEpisodes:            in.NoEpisodes,
	}
}

// buildWorkFormParams converts the string form values into the typed works columns
// shared by create and update. repository.CreateWorkParams is exactly the common shape;
// the update flow adds the target ID (see buildUpdateWorkParams).
//
// Every parse failure is returned as an error. The validator accepts exactly the values
// converted here, so a failure means the two drifted apart; failing loudly turns that into
// a logged 500 instead of a save that quietly stores NULL for what the submitter typed.
//
// [Ja] buildWorkFormParams は文字列のフォーム値を、作成と更新で共有する works の型付き
// カラムに変換する。repository.CreateWorkParams が共通の形そのもので、更新フローは対象 ID を
// 足す (buildUpdateWorkParams を参照)。
//
// パースの失敗はすべてエラーとして返す。バリデーターはここで変換する値をそのまま許容するため、
// 失敗は両者がドリフトしたことを意味する。明示的に失敗させることで、送信者が入力した値を
// 黙って NULL で保存する代わりに、ログの残る 500 になる。
func buildWorkFormParams(input WorkFormInput) (repository.CreateWorkParams, error) {
	media, err := strconv.ParseInt(input.Media, 10, 32)
	if err != nil {
		return repository.CreateWorkParams{}, fmt.Errorf("メディア値の変換に失敗: %w", err)
	}

	params := repository.CreateWorkParams{
		Title:                 strings.TrimSpace(input.Title),
		TitleKana:             strings.TrimSpace(input.TitleKana),
		TitleAlter:            strings.TrimSpace(input.TitleAlter),
		TitleEn:               strings.TrimSpace(input.TitleEn),
		TitleAlterEn:          strings.TrimSpace(input.TitleAlterEn),
		Media:                 int32(media),
		OfficialSiteURL:       strings.TrimSpace(input.OfficialSiteURL),
		OfficialSiteURLEn:     strings.TrimSpace(input.OfficialSiteURLEn),
		WikipediaURL:          strings.TrimSpace(input.WikipediaURL),
		WikipediaURLEn:        strings.TrimSpace(input.WikipediaURLEn),
		Synopsis:              strings.TrimSpace(input.Synopsis),
		SynopsisSource:        strings.TrimSpace(input.SynopsisSource),
		SynopsisEn:            strings.TrimSpace(input.SynopsisEn),
		SynopsisSourceEn:      strings.TrimSpace(input.SynopsisSourceEn),
		NoEpisodes:            input.NoEpisodes == "1",
		StartEpisodeRawNumber: 1.0,
	}

	if params.SeasonYear, err = parseOptionalInt32(input.SeasonYear); err != nil {
		return repository.CreateWorkParams{}, fmt.Errorf("リリース時期 (年) の変換に失敗: %w", err)
	}

	if params.SeasonName, err = parseOptionalInt32(input.SeasonName); err != nil {
		return repository.CreateWorkParams{}, fmt.Errorf("リリース時期 (季節) の変換に失敗: %w", err)
	}

	if params.StartedOn, err = parseOptionalDate(input.StartedOn); err != nil {
		return repository.CreateWorkParams{}, fmt.Errorf("開始日の変換に失敗: %w", err)
	}

	if params.EndedOn, err = parseOptionalDate(input.EndedOn); err != nil {
		return repository.CreateWorkParams{}, fmt.Errorf("終了日の変換に失敗: %w", err)
	}

	if input.TwitterUsername != "" {
		params.TwitterUsername = sql.NullString{String: strings.TrimSpace(input.TwitterUsername), Valid: true}
	}

	if input.TwitterHashtag != "" {
		params.TwitterHashtag = sql.NullString{String: strings.TrimSpace(input.TwitterHashtag), Valid: true}
	}

	if params.ScTid, err = parseOptionalInt32(input.ScTid); err != nil {
		return repository.CreateWorkParams{}, fmt.Errorf("しょぼいカレンダー TID の変換に失敗: %w", err)
	}

	if params.MalAnimeID, err = parseOptionalInt32(input.MalAnimeID); err != nil {
		return repository.CreateWorkParams{}, fmt.Errorf("MyAnimeList ID の変換に失敗: %w", err)
	}

	if params.ManualEpisodesCount, err = parseOptionalInt32(input.ManualEpisodesCount); err != nil {
		return repository.CreateWorkParams{}, fmt.Errorf("エピソード数の変換に失敗: %w", err)
	}

	if input.StartEpisodeRawNumber != "" {
		v, err := strconv.ParseFloat(input.StartEpisodeRawNumber, 64)
		if err != nil {
			return repository.CreateWorkParams{}, fmt.Errorf("開始話数の変換に失敗: %w", err)
		}
		params.StartEpisodeRawNumber = v
	}

	if input.NumberFormatID != "" {
		v, err := strconv.ParseInt(input.NumberFormatID, 10, 64)
		if err != nil {
			return repository.CreateWorkParams{}, fmt.Errorf("話数フォーマットの変換に失敗: %w", err)
		}
		params.NumberFormatID = sql.NullInt64{Int64: v, Valid: true}
	}

	return params, nil
}

// parseOptionalInt32 converts an optional form value into the works integer columns,
// mapping an empty value to NULL. The bit size is what the validator checks against, so
// the two accept the same range.
//
// [Ja] parseOptionalInt32 は任意入力のフォーム値を works の integer カラムに変換し、空値を
// NULL に写像する。ビット幅はバリデーターが検査するものと同じで、両者は同一の範囲を許容する。
func parseOptionalInt32(value string) (sql.NullInt32, error) {
	if value == "" {
		return sql.NullInt32{}, nil
	}
	v, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return sql.NullInt32{}, err
	}
	return sql.NullInt32{Int32: int32(v), Valid: true}, nil
}

// parseOptionalDate converts an optional date form value into a works date column, mapping
// an empty value to NULL.
//
// [Ja] parseOptionalDate は任意入力の日付フォーム値を works の date カラムに変換し、空値を
// NULL に写像する。
func parseOptionalDate(value string) (sql.NullTime, error) {
	if value == "" {
		return sql.NullTime{}, nil
	}
	t, err := time.Parse(validator.WorkFormDateLayout, value)
	if err != nil {
		return sql.NullTime{}, err
	}
	return sql.NullTime{Time: t, Valid: true}, nil
}
