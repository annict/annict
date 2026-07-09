package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/validator"
)

type CreateWorkUsecase struct {
	db                      *sql.DB
	workRepo                *repository.WorkRepository
	animeRepo               *repository.AnimeRepository
	animeClassificationRepo *repository.AnimeClassificationRepository
	satelliteRepos          WorkSatelliteRepos
	validator               *validator.DBWorkCreateValidator
}

func NewCreateWorkUsecase(
	db *sql.DB,
	workRepo *repository.WorkRepository,
	animeRepo *repository.AnimeRepository,
	animeClassificationRepo *repository.AnimeClassificationRepository,
	satelliteRepos WorkSatelliteRepos,
	validator *validator.DBWorkCreateValidator,
) *CreateWorkUsecase {
	return &CreateWorkUsecase{
		db:                      db,
		workRepo:                workRepo,
		animeRepo:               animeRepo,
		animeClassificationRepo: animeClassificationRepo,
		satelliteRepos:          satelliteRepos,
		validator:               validator,
	}
}

// CreateWorkInput carries the form values for creating a work. It is the shared
// WorkFormInput with no extra fields; the type name keeps the create usecase's intent
// explicit.
//
// [Ja] CreateWorkInput は作品作成フォームの入力値を保持する。追加フィールドを持たない
// 共有 WorkFormInput そのもので、型名で作成 UseCase の意図を明示する。
type CreateWorkInput struct {
	WorkFormInput
}

type CreateWorkOutput struct {
	WorkID model.WorkID
}

func (uc *CreateWorkUsecase) Execute(ctx context.Context, input CreateWorkInput) (*CreateWorkOutput, error) {
	if err := uc.validator.Validate(ctx, input.toValidatorInput()); err != nil {
		return nil, err
	}

	params, err := buildWorkFormParams(input.WorkFormInput)
	if err != nil {
		return nil, fmt.Errorf("入力値の変換に失敗: %w", err)
	}

	return uc.createWork(ctx, params)
}

// createWork persists a new work across animes / anime_classifications / works and the
// six satellite tables in a single transaction, anchored on animes: it inserts the anime,
// inserts its kind='work' classification, inserts the work, writes works.anime_id back,
// then dual-writes the satellite rows the work sources. works stays the source of truth
// during the migration, so the works writes (Create + UpdateAnimeID) are kept in one block
// that the cutover (phase 17) can remove wholesale.
//
// [Ja] createWork は新規作品を animes / anime_classifications / works と 6 つの別表に
// またがって 1 トランザクションで永続化する。animes を基点に、anime を挿入し、その
// kind='work' 分類を挿入し、works を挿入し、works.anime_id を書き戻し、work が source とする
// 別表行を両書きする。移行期間中は works が正本のため、works への書き込み (Create +
// UpdateAnimeID) は正本切り替え (フェーズ 17) でまるごと外せるよう 1 ブロックにまとめてある。
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

	// Dual-write the six satellite tables the work sources onto the just-created anime.
	// The anime is brand new, so it owns no rows yet and every plan is all-creates; passing
	// the tx-bound repos keeps the writes atomic with the works / anime writes above. The
	// mapping reuses the phase 2 plan* helpers, so a satellite sync right after this create
	// reports Unchanged. work.AnimeID is set here because the anime's id is only known after
	// its insert (mirroring how classificationParams.AnimeID is patched above).
	//
	// [Ja] work が source とする 6 つの別表を、作りたての anime に両書きする。anime は新規で
	// まだ行を持たないため全計画が作成のみになる。tx 束ねリポジトリを渡すことで、上の works /
	// anime の書き込みと原子的になる。写像はフェーズ 2 の plan* ヘルパーを再利用するため、
	// 作成直後の別表同期は Unchanged を報告する。anime の id は挿入後にしか分からないため
	// work.AnimeID はここでセットする (上の classificationParams.AnimeID の書き換えと同じ)。
	work.AnimeID = &anime.ID
	if err := applyWorkSatellitePlans(ctx, uc.satelliteRepos.WithTx(tx), planWorkSatellites(work, workSatelliteExisting{})); err != nil {
		return nil, fmt.Errorf("別表テーブルの両書きに失敗しました: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return &CreateWorkOutput{WorkID: workID}, nil
}

// workFromCreateWorkParams projects a CreateWorkParams onto the *model.Work fields the
// animes / anime_classifications mapping and the satellite-table mapping read, so the
// create path feeds the same animeCreateParamsFromWork / classificationCreateParamsFromWork
// helpers and the same desired* satellite helpers the phase 2 sync uses. Single-sourcing
// the mapping keeps create and sync from drifting, so the sync run right after a create
// reports Unchanged (no spurious UPDATE, no inflated diff metric).
//
// It mirrors the partial-load pattern of workFromAnimeSyncRow and workFromSatelliteSyncRow:
// only the mapped columns are set and the rest of *model.Work stays at its zero value. The
// text columns are NOT NULL with an empty-string default, so they keep the empty string
// (the url columns are mapped to "no row" and the anime text columns to NULL later by the
// helpers); the nullable source columns (sc_tid / mal_anime_id / twitter_* / season_* /
// started_on / ended_on) become pointers exactly as the satellite sync loader reads them
// back. A new work leaves unpublished_at / deleted_at at NULL, so DerivedStatus reports
// published and the anime the sync maps back is published too.
//
// [Ja] workFromCreateWorkParams は CreateWorkParams を、animes / anime_classifications の
// 写像と別表の写像が読む *model.Work フィールドに射影する。これにより create 経路もフェーズ 2
// 同期と同じ animeCreateParamsFromWork / classificationCreateParamsFromWork ヘルパーと同じ
// desired* 別表ヘルパーに通せる。写像の正本を 1 つにすることで create と同期がドリフトせず、
// 作成直後の同期が Unchanged を報告する (無駄な UPDATE も差分メトリクスの水増しも生まない)。
//
// workFromAnimeSyncRow / workFromSatelliteSyncRow の partial-load パターンに倣い、写像対象の
// カラムだけをセットして残りの *model.Work はゼロ値のまま残す。テキストカラムは NOT NULL かつ
// デフォルトが空文字列のため空文字列のまま保持し (url カラムは後段で「行なし」に、anime の
// テキストカラムは NULL にヘルパーが写像する)、NULL 許容のソース列 (sc_tid / mal_anime_id /
// twitter_* / season_* / started_on / ended_on) は別表同期ローダーが読み戻すのと同じくポインタに
// する。新規 work は unpublished_at / deleted_at を NULL のままにするため DerivedStatus は
// published を報告し、同期が写し戻す anime も published になる。
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
		NoEpisodes:            params.NoEpisodes,
		StartEpisodeRawNumber: params.StartEpisodeRawNumber,
		OfficialSiteURL:       params.OfficialSiteURL,
		OfficialSiteURLEn:     params.OfficialSiteURLEn,
		WikipediaURL:          params.WikipediaURL,
		WikipediaURLEn:        params.WikipediaURLEn,
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
	if params.ScTid.Valid {
		scTid := params.ScTid.Int32
		work.ScTid = &scTid
	}
	if params.MalAnimeID.Valid {
		malAnimeID := params.MalAnimeID.Int32
		work.MalAnimeID = &malAnimeID
	}
	if params.TwitterUsername.Valid {
		twitterUsername := params.TwitterUsername.String
		work.TwitterUsername = &twitterUsername
	}
	if params.TwitterHashtag.Valid {
		twitterHashtag := params.TwitterHashtag.String
		work.TwitterHashtag = &twitterHashtag
	}
	if params.SeasonYear.Valid {
		seasonYear := params.SeasonYear.Int32
		work.SeasonYear = &seasonYear
	}
	if params.SeasonName.Valid {
		seasonName := params.SeasonName.Int32
		work.SeasonName = &seasonName
	}
	if params.StartedOn.Valid {
		startedOn := params.StartedOn.Time
		work.StartedOn = &startedOn
	}
	if params.EndedOn.Valid {
		endedOn := params.EndedOn.Time
		work.EndedOn = &endedOn
	}
	return work
}
