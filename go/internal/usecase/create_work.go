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

type CreateWorkUsecase struct {
	db                      *sql.DB
	workRepo                *repository.WorkRepository
	animeRepo               *repository.AnimeRepository
	animeClassificationRepo *repository.AnimeClassificationRepository
	validator               *validator.DbWorkCreateValidator
}

func NewCreateWorkUsecase(
	db *sql.DB,
	workRepo *repository.WorkRepository,
	animeRepo *repository.AnimeRepository,
	animeClassificationRepo *repository.AnimeClassificationRepository,
	validator *validator.DbWorkCreateValidator,
) *CreateWorkUsecase {
	return &CreateWorkUsecase{
		db:                      db,
		workRepo:                workRepo,
		animeRepo:               animeRepo,
		animeClassificationRepo: animeClassificationRepo,
		validator:               validator,
	}
}

// CreateWorkInput carries the form values for creating a work. All fields are typed
// as string because they come straight from the HTML form submission and are
// type-converted later in buildCreateWorkParams.
//
// [Ja] CreateWorkInput は作品作成フォームの入力値を保持する。HTML フォーム由来のため
// 全フィールドを文字列として持ち、後段の buildCreateWorkParams で型変換する。
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

type CreateWorkOutput struct {
	WorkID model.WorkID
}

func (uc *CreateWorkUsecase) Execute(ctx context.Context, input CreateWorkInput) (*CreateWorkOutput, error) {
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

	params, err := buildCreateWorkParams(input)
	if err != nil {
		return nil, fmt.Errorf("入力値の変換に失敗: %w", err)
	}

	return uc.createWork(ctx, params)
}

// createWork persists a new work across animes / anime_classifications / works in a
// single transaction, anchored on animes: it inserts the anime, inserts its kind='work'
// classification, inserts the work, then writes works.anime_id back. works stays the
// source of truth during the migration, so the works writes (Create + UpdateAnimeID) are
// kept in one block that the cutover (phase 17) can remove wholesale.
//
// [Ja] createWork は新規作品を animes / anime_classifications / works にまたがって 1
// トランザクションで永続化する。animes を基点に、anime を挿入し、その kind='work' 分類を
// 挿入し、works を挿入し、works.anime_id を書き戻す。移行期間中は works が正本のため、
// works への書き込み (Create + UpdateAnimeID) は正本切り替え (フェーズ 17) でまるごと
// 外せるよう 1 ブロックにまとめてある。
func (uc *CreateWorkUsecase) createWork(ctx context.Context, params repository.CreateWorkParams) (*CreateWorkOutput, error) {
	// Project the create params onto a *model.Work and reuse the phase 2 sync mapping
	// helpers, keeping the work -> anime / classification mapping single-sourced.
	//
	// [Ja] create パラメータを *model.Work に射影し、フェーズ 2 同期の写像ヘルパーを
	// 再利用して、work -> anime / 分類 の写像の正本を 1 つに保つ。
	work := workFromCreateWorkParams(params)
	animeParams := animeCreateParamsFromWork(work)
	classificationParams := classificationCreateParamsFromWork(work, 0)

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	animeRepo := uc.animeRepo.WithTx(tx)
	classificationRepo := uc.animeClassificationRepo.WithTx(tx)
	workRepo := uc.workRepo.WithTx(tx)

	// works.anime_id is an FK to animes(id), so the anime must exist first; write order
	// is anime -> classification -> works -> anime_id write-back.
	//
	// [Ja] works.anime_id は animes(id) への FK なので anime を先に作る必要がある。
	// 書き込み順は anime -> classification -> works -> anime_id 書き戻し。
	anime, err := animeRepo.Create(ctx, animeParams)
	if err != nil {
		return nil, fmt.Errorf("anime の作成に失敗しました: %w", err)
	}

	classificationParams.AnimeID = anime.ID
	if _, err := classificationRepo.Create(ctx, classificationParams); err != nil {
		return nil, fmt.Errorf("anime_classification の作成に失敗しました: %w", err)
	}

	workID, err := workRepo.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("作品の作成に失敗しました: %w", err)
	}

	if err := workRepo.UpdateAnimeID(ctx, workID, anime.ID); err != nil {
		return nil, fmt.Errorf("works.anime_id の書き戻しに失敗しました: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return &CreateWorkOutput{WorkID: workID}, nil
}

// workFromCreateWorkParams projects a CreateWorkParams onto the *model.Work fields the
// animes / anime_classifications mapping reads, so the create path feeds the same
// animeCreateParamsFromWork / classificationCreateParamsFromWork helpers the phase 2 sync
// uses. Single-sourcing the mapping keeps create and sync from drifting, so the sync run
// right after a create reports Unchanged (no spurious UPDATE, no inflated diff metric).
//
// It mirrors workFromAnimeSyncRow's partial-load pattern: only the mapped columns are
// set and the rest of *model.Work stays at its zero value. Those text columns are
// NOT NULL with an empty-string default, so they keep the empty string (mapped to NULL
// later by the helpers), and a new work is always status='published' (the works.status
// column default), matching the value the sync loader reads back.
//
// [Ja] workFromCreateWorkParams は CreateWorkParams を、animes / anime_classifications の
// 写像が読む *model.Work フィールドに射影する。これにより create 経路もフェーズ 2 同期と
// 同じ animeCreateParamsFromWork / classificationCreateParamsFromWork ヘルパーに通せる。
// 写像の正本を 1 つにすることで create と同期がドリフトせず、作成直後の同期が Unchanged を
// 報告する (無駄な UPDATE も差分メトリクスの水増しも生まない)。
//
// workFromAnimeSyncRow の partial-load パターンに倣い、写像対象のカラムだけをセットして
// 残りの *model.Work はゼロ値のまま残す。これらのテキストカラムは NOT NULL かつデフォルトが
// 空文字列のため、空文字列のまま保持され (ヘルパーが後段で NULL に写像する)、新規 work は
// 常に status='published' (works.status カラムの既定値) とし、同期ローダーが読み戻す値に
// 一致させる。
func workFromCreateWorkParams(params repository.CreateWorkParams) *model.Work {
	work := &model.Work{
		Title:                 params.Title,
		TitleEn:               params.TitleEn,
		TitleAlter:            params.TitleAlter,
		TitleAlterEn:          params.TitleAlterEn,
		Media:                 params.Media,
		Synopsis:              params.Synopsis,
		SynopsisEn:            params.SynopsisEn,
		SynopsisSource:        params.SynopsisSource,
		SynopsisSourceEn:      params.SynopsisSourceEn,
		Status:                model.WorkStatusPublished,
		NoEpisodes:            params.NoEpisodes,
		StartEpisodeRawNumber: params.StartEpisodeRawNumber,
	}
	if params.TitleKana != "" {
		titleKana := params.TitleKana
		work.TitleKana = &titleKana
	}
	if params.ManualEpisodesCount.Valid {
		manualEpisodesCount := params.ManualEpisodesCount.Int32
		work.ManualEpisodesCount = &manualEpisodesCount
	}
	if params.NumberFormatID.Valid {
		numberFormatID := model.NumberFormatID(params.NumberFormatID.Int64)
		work.NumberFormatID = &numberFormatID
	}
	return work
}

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
