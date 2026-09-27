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

// CreateWorkInputは作品作成フォームの入力値を保持する。追加フィールドを持たない
// 共有WorkFormInputそのもので、型名で作成UseCaseの意図を明示する。
type CreateWorkInput struct {
	WorkFormInput
}

type CreateWorkOutput struct {
	WorkID model.WorkID
}

func (uc *CreateWorkUsecase) Execute(ctx context.Context, input CreateWorkInput) (*CreateWorkOutput, error) {
	// 作成フローは版を示さない。送信が古くなり得る保存済みの行が無いため、バリデーターは
	// 版を返さず、捨てている戻り値は取りこぼしではない。
	if _, err := uc.validator.Validate(ctx, input.toValidatorInput(nil, nil)); err != nil {
		return nil, err
	}

	params, err := buildWorkFormParams(input.WorkFormInput)
	if err != nil {
		return nil, fmt.Errorf("入力値の変換に失敗: %w", err)
	}

	return uc.createWork(ctx, params)
}

// createWorkは新規作品をanimes / anime_classifications / worksと6つの別表に
// またがって1トランザクションで永続化する。animesを基点に、animeを挿入し、その
// kind='work' 分類を挿入し、worksを挿入し、works.anime_idを書き戻し、workがsourceとする
// 別表行を両書きする。移行期間中はworksが正本のため、worksへの書き込み (Create +
// UpdateAnimeID) は正本切り替え (フェーズ17) でまるごと外せるよう1ブロックにまとめてある。
func (uc *CreateWorkUsecase) createWork(ctx context.Context, params repository.CreateWorkParams) (*CreateWorkOutput, error) {
	// createパラメータを *model.Workに射影し、フェーズ2同期の写像ヘルパーを
	// 再利用して、work -> anime / 分類 の写像の正本を1つに保つ。
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

	// works.anime_idはanimes(id) へのFKなのでanimeを先に作る必要がある。
	// 書き込み順はanime -> classification -> works -> anime_id書き戻し。
	anime, err := animeRepo.Create(ctx, animeParams)
	if err != nil {
		return nil, fmt.Errorf("animeの作成に失敗しました: %w", err)
	}

	classificationParams.AnimeID = anime.ID
	if _, err := classificationRepo.Create(ctx, classificationParams); err != nil {
		return nil, fmt.Errorf("anime_classificationの作成に失敗しました: %w", err)
	}

	workID, err := workRepo.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("作品の作成に失敗しました: %w", err)
	}

	if err := workRepo.UpdateAnimeID(ctx, workID, anime.ID); err != nil {
		return nil, fmt.Errorf("works.anime_idの書き戻しに失敗しました: %w", err)
	}

	// workがsourceとする6つの別表を、作りたてのanimeに両書きする。animeは新規で
	// まだ行を持たないため全計画が作成のみになる。tx束ねリポジトリを渡すことで、上のworks /
	// animeの書き込みと原子的になる。写像はフェーズ2のplan* ヘルパーを再利用するため、
	// 作成直後の別表同期はUnchangedを報告する。animeのidは挿入後にしか分からないため
	// work.AnimeIDはここでセットする (上のclassificationParams.AnimeIDの書き換えと同じ)。
	work.AnimeID = &anime.ID
	if err := applyWorkSatellitePlans(ctx, uc.satelliteRepos.WithTx(tx), planWorkSatellites(work, workSatelliteExisting{})); err != nil {
		return nil, fmt.Errorf("別表テーブルの両書きに失敗しました: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return &CreateWorkOutput{WorkID: workID}, nil
}

// workFromCreateWorkParamsはCreateWorkParamsを、animes / anime_classificationsの
// 写像と別表の写像が読む *model.Workフィールドに射影する。これによりcreate経路もフェーズ2
// 同期と同じanimeCreateParamsFromWork / classificationCreateParamsFromWorkヘルパーと同じ
// desired* 別表ヘルパーに通せる。写像の正本を1つにすることでcreateと同期がドリフトせず、
// 作成直後の同期がUnchangedを報告する (無駄なUPDATEも差分メトリクスの水増しも生まない)。
//
// workFromAnimeSyncRow / workFromSatelliteSyncRowのpartial-loadパターンに倣い、写像対象の
// カラムだけをセットして残りの *model.Workはゼロ値のまま残す。テキストカラムはNOT NULLかつ
// デフォルトが空文字列のため空文字列のまま保持し (urlカラムは後段で「行なし」に、animeの
// テキストカラムはNULLにヘルパーが写像する)、NULL許容のソース列 (sc_tid / mal_anime_id /
// twitter_* / season_* / started_on / ended_on) は別表同期ローダーが読み戻すのと同じくポインタに
// する。新規workはunpublished_at / deleted_atをNULLのままにするためDerivedStatusは
// publishedを報告し、同期が写し戻すanimeもpublishedになる。
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
