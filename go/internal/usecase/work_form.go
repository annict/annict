package usecase

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

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
//
// [Ja] toValidatorInput はフォーム値をバリデーター入力に射影する。作成と更新は同一の
// フィールド集合を検証するため、双方ともこの単一の写像を通す。
func (in WorkFormInput) toValidatorInput() validator.DBWorkCreateValidatorInput {
	return validator.DBWorkCreateValidatorInput{
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
// the update flow adds the target ID (see buildUpdateWorkParams). Numeric / date fields
// that fail to parse are left unset, mirroring the create form: the validator rejects
// malformed required fields upstream, and the optional ones fall back to NULL.
//
// [Ja] buildWorkFormParams は文字列のフォーム値を、作成と更新で共有する works の型付き
// カラムに変換する。repository.CreateWorkParams が共通の形そのもので、更新フローは対象 ID を
// 足す (buildUpdateWorkParams を参照)。数値・日付のパースに失敗したフィールドは未設定のまま
// にする (作成フォームと同じ挙動)。必須項目の不正は上流のバリデーターが弾き、任意項目は
// NULL に落ちる。
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

	if input.SeasonYear != "" {
		v, err := strconv.ParseInt(input.SeasonYear, 10, 32)
		if err == nil {
			params.SeasonYear = sql.NullInt32{Int32: int32(v), Valid: true}
		}
	}

	if input.SeasonName != "" {
		v, err := strconv.ParseInt(input.SeasonName, 10, 32)
		if err == nil {
			params.SeasonName = sql.NullInt32{Int32: int32(v), Valid: true}
		}
	}

	if input.StartedOn != "" {
		t, err := time.Parse("2006-01-02", input.StartedOn)
		if err == nil {
			params.StartedOn = sql.NullTime{Time: t, Valid: true}
		}
	}

	if input.EndedOn != "" {
		t, err := time.Parse("2006-01-02", input.EndedOn)
		if err == nil {
			params.EndedOn = sql.NullTime{Time: t, Valid: true}
		}
	}

	if input.TwitterUsername != "" {
		params.TwitterUsername = sql.NullString{String: strings.TrimSpace(input.TwitterUsername), Valid: true}
	}

	if input.TwitterHashtag != "" {
		params.TwitterHashtag = sql.NullString{String: strings.TrimSpace(input.TwitterHashtag), Valid: true}
	}

	if input.ScTid != "" {
		v, err := strconv.ParseInt(input.ScTid, 10, 32)
		if err == nil {
			params.ScTid = sql.NullInt32{Int32: int32(v), Valid: true}
		}
	}

	if input.MalAnimeID != "" {
		v, err := strconv.ParseInt(input.MalAnimeID, 10, 32)
		if err == nil {
			params.MalAnimeID = sql.NullInt32{Int32: int32(v), Valid: true}
		}
	}

	if input.ManualEpisodesCount != "" {
		v, err := strconv.ParseInt(input.ManualEpisodesCount, 10, 32)
		if err == nil {
			params.ManualEpisodesCount = sql.NullInt32{Int32: int32(v), Valid: true}
		}
	}

	if input.StartEpisodeRawNumber != "" {
		v, err := strconv.ParseFloat(input.StartEpisodeRawNumber, 64)
		if err == nil {
			params.StartEpisodeRawNumber = v
		}
	}

	if input.NumberFormatID != "" {
		v, err := strconv.ParseInt(input.NumberFormatID, 10, 64)
		if err == nil {
			params.NumberFormatID = sql.NullInt64{Int64: v, Valid: true}
		}
	}

	return params, nil
}
