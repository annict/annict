package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/validator"
)

// CreateWorkUsecase は作品作成のユースケース
type CreateWorkUsecase struct {
	db        *sql.DB
	workRepo  *repository.WorkRepository
	validator *validator.DbWorkCreateValidator
}

// NewCreateWorkUsecase はCreateWorkUsecaseを作成します
func NewCreateWorkUsecase(
	db *sql.DB,
	workRepo *repository.WorkRepository,
	validator *validator.DbWorkCreateValidator,
) *CreateWorkUsecase {
	return &CreateWorkUsecase{
		db:        db,
		workRepo:  workRepo,
		validator: validator,
	}
}

// CreateWorkInput は作品作成の入力データ（フォームの文字列値）
type CreateWorkInput struct {
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

// CreateWorkOutput は作品作成の結果
type CreateWorkOutput struct {
	WorkID model.WorkID
}

// Execute はバリデーション・型変換・作品作成を行います
func (uc *CreateWorkUsecase) Execute(ctx context.Context, input CreateWorkInput) (*CreateWorkOutput, error) {
	// 1. バリデーション
	if err := uc.validator.Validate(ctx, validator.DbWorkCreateValidatorInput{
		Title:                 input.Title,
		TitleKana:             input.TitleKana,
		TitleAlter:            input.TitleAlter,
		TitleEn:               input.TitleEn,
		TitleAlterEn:          input.TitleAlterEn,
		Media:                 input.Media,
		SeasonYear:            input.SeasonYear,
		SeasonName:            input.SeasonName,
		StartedOn:             input.StartedOn,
		EndedOn:               input.EndedOn,
		OfficialSiteURL:       input.OfficialSiteURL,
		OfficialSiteURLEn:     input.OfficialSiteURLEn,
		WikipediaURL:          input.WikipediaURL,
		WikipediaURLEn:        input.WikipediaURLEn,
		TwitterUsername:       input.TwitterUsername,
		TwitterHashtag:        input.TwitterHashtag,
		ScTid:                 input.ScTid,
		MalAnimeID:            input.MalAnimeID,
		Synopsis:              input.Synopsis,
		SynopsisSource:        input.SynopsisSource,
		SynopsisEn:            input.SynopsisEn,
		SynopsisSourceEn:      input.SynopsisSourceEn,
		ManualEpisodesCount:   input.ManualEpisodesCount,
		StartEpisodeRawNumber: input.StartEpisodeRawNumber,
		NumberFormatID:        input.NumberFormatID,
		NoEpisodes:            input.NoEpisodes,
	}); err != nil {
		return nil, err
	}

	// 2. フォーム値を型変換
	params, err := buildCreateWorkParams(input)
	if err != nil {
		return nil, fmt.Errorf("入力値の変換に失敗: %w", err)
	}

	// 3. トランザクション内で作品を作成
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクションの開始に失敗しま���た: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	workRepo := uc.workRepo.WithTx(tx)

	workID, err := workRepo.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("作品の作成に失敗しました: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return &CreateWorkOutput{WorkID: workID}, nil
}

// buildCreateWorkParams はフォーム入力値をリポジトリのパラメータに変換します
func buildCreateWorkParams(input CreateWorkInput) (repository.CreateWorkParams, error) {
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
