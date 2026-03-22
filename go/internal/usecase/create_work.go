package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/repository"
)

// CreateWorkUsecase は作品作成のユースケース
type CreateWorkUsecase struct {
	db       *sql.DB
	workRepo *repository.WorkRepository
}

// NewCreateWorkUsecase はCreateWorkUsecaseを作成します
func NewCreateWorkUsecase(
	db *sql.DB,
	workRepo *repository.WorkRepository,
) *CreateWorkUsecase {
	return &CreateWorkUsecase{
		db:       db,
		workRepo: workRepo,
	}
}

// CreateWorkInput は作品作成の入力データ
type CreateWorkInput struct {
	Title                 string
	TitleKana             string
	TitleAlter            string
	TitleEn               string
	TitleAlterEn          string
	Media                 int32
	SeasonYear            sql.NullInt32
	SeasonName            sql.NullInt32
	StartedOn             sql.NullTime
	EndedOn               sql.NullTime
	OfficialSiteURL       string
	OfficialSiteURLEn     string
	WikipediaURL          string
	WikipediaURLEn        string
	TwitterUsername       sql.NullString
	TwitterHashtag        sql.NullString
	ScTid                 sql.NullInt32
	MalAnimeID            sql.NullInt32
	Synopsis              string
	SynopsisSource        string
	SynopsisEn            string
	SynopsisSourceEn      string
	ManualEpisodesCount   sql.NullInt32
	StartEpisodeRawNumber float64
	NumberFormatID        sql.NullInt64
	NoEpisodes            bool
}

// CreateWorkResult は作品作成の結果
type CreateWorkResult struct {
	WorkID int64
}

// Execute は作品を作成します
func (uc *CreateWorkUsecase) Execute(ctx context.Context, input CreateWorkInput) (*CreateWorkResult, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	workRepo := uc.workRepo.WithTx(tx)

	workID, err := workRepo.Create(ctx, repository.CreateWorkParams{
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
	})
	if err != nil {
		return nil, fmt.Errorf("作品の作成に失敗しました: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return &CreateWorkResult{WorkID: workID}, nil
}
