package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/validator"
)

// UpdateWorkUsecaseはAnnict DB管理画面の編集フォームから作品を更新する。
// CreateWorkUsecaseと同じくanimesを基点とし、works行 (移行期間中は正本のまま) を
// 更新しつつ、同一トランザクションでマッピング済みのanime / 分類にも両書きする。
// 編集フォームは作成フォームと同一のフィールドを検証するためDBWorkCreateValidatorを
// 再利用する。
type UpdateWorkUsecase struct {
	db                      *sql.DB
	workRepo                *repository.WorkRepository
	animeRepo               *repository.AnimeRepository
	animeClassificationRepo *repository.AnimeClassificationRepository
	satelliteRepos          WorkSatelliteRepos
	validator               *validator.DBWorkCreateValidator
}

func NewUpdateWorkUsecase(
	db *sql.DB,
	workRepo *repository.WorkRepository,
	animeRepo *repository.AnimeRepository,
	animeClassificationRepo *repository.AnimeClassificationRepository,
	satelliteRepos WorkSatelliteRepos,
	validator *validator.DBWorkCreateValidator,
) *UpdateWorkUsecase {
	return &UpdateWorkUsecase{
		db:                      db,
		workRepo:                workRepo,
		animeRepo:               animeRepo,
		animeClassificationRepo: animeClassificationRepo,
		satelliteRepos:          satelliteRepos,
		validator:               validator,
	}
}

// UpdateWorkInputは対象のwork IDと共有の作品フォーム入力値を保持する。WorkFormInput
// を埋め込み、フィールド集合・バリデーター入力・文字列→型変換を作成と共有する。
type UpdateWorkInput struct {
	WorkID model.WorkID
	// UpdatedAtはフォームを開いた時点の版で、hiddenフィールドが運ぶ形のまま。
	// WorkFormInputの中ではなく外に置くのは、作成フローには示すべき版が無いため。非競合の
	// 却下では送信された版を保ち、競合時は保存済みの内容を示してからその版へ載せ替える。
	UpdatedAt string
	WorkFormInput
}

type UpdateWorkOutput struct {
	WorkID model.WorkID
}

func (uc *UpdateWorkUsecase) Execute(ctx context.Context, input UpdateWorkInput) (*UpdateWorkOutput, error) {
	version, err := uc.validator.Validate(ctx, input.toValidatorInput(&input.WorkID, &input.UpdatedAt))
	if err != nil {
		return nil, err
	}

	// 編集対象のworkをanime同期の射影で読み込む。これはworks.anime_idの
	// マッピングと、編集フォームが触れないanime写像カラム (title_ro / archive_message /
	// anime.statusの導出源である作品状態のsource unpublished_at / deleted_at) を持ち、
	// いずれもマッピング済みanimeをそれらのカラムを潰さずに更新するために要る。結果が
	// 空の場合は編集GETとこのPATCHの間にworkが削除された。
	works, err := uc.workRepo.ListForAnimeSyncByIDs(ctx, []model.WorkID{input.WorkID})
	if err != nil {
		return nil, fmt.Errorf("作品の取得に失敗しました: %w", err)
	}
	if len(works) == 0 {
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_work_not_found"),
			Metadata: map[string]string{"work_id": input.WorkID.String()},
		}
	}
	current := works[0]

	// animeUpdateParamsFromWorkがanimes由来でないカラム (release_statusなど) を
	// 引き継げるよう既存animeを読み込む。nilはworkが未だanimeにマッピングされて
	// いないことを表し、その場合animesへの両書きはスキップする (updateWorkを参照)。
	var existingAnime *model.Anime
	var existingSatellites workSatelliteExisting
	if current.AnimeID != nil {
		existingAnime, err = uc.animeRepo.GetByID(ctx, *current.AnimeID)
		if err != nil {
			return nil, fmt.Errorf("animeの取得に失敗しました: %w", err)
		}
		// workがマッピング済みのときは、更新値と突合できるようanimeの既存別表行を
		// トランザクションの前に読む (書き込みUseCaseのルールでI/Oをトランザクション本体の
		// 外に出す)。anime_idが存在しないanimeを指す場合GetByIDはnilを返し、その場合は
		// NULLのanime_idと同じく読み込みとanime / 別表の両書きをスキップする。
		if existingAnime != nil {
			existingSatellites, err = readWorkSatelliteExisting(ctx, uc.satelliteRepos, existingAnime.ID)
			if err != nil {
				return nil, fmt.Errorf("既存別表行の取得に失敗しました: %w", err)
			}
		}
	}

	params, err := buildUpdateWorkParams(input, version)
	if err != nil {
		return nil, fmt.Errorf("入力値の変換に失敗: %w", err)
	}

	return uc.updateWork(ctx, params, current, existingAnime, existingSatellites)
}

// updateWorkは更新をworksに、そしてworkが既にマッピング済みなら そのanime /
// anime_classificationと6つの別表にも1トランザクションで永続化する。移行期間中はworksが
// 正本のため、animeと別表への書き込みは正本切り替え (フェーズ17) でまるごと外せるよう1
// ブロックにまとめてある。
//
// workが未マッピング (existingAnime == nil) の場合はworksだけを更新し、animeの作成は
// フェーズ2の同期バッチに委ねる。同期は裁定者であり既存workに対するanimeの唯一の
// 作成者なので、ここでは作成しない。workはanime_id NULLのまま既にバッチから見えており、
// ここで並行してanimeを作ると1つのworkに2つのanimeができる競合になる。
func (uc *UpdateWorkUsecase) updateWork(ctx context.Context, params repository.UpdateWorkParams, current *model.Work, existingAnime *model.Anime, existingSatellites workSatelliteExisting) (*UpdateWorkOutput, error) {
	// anime / 分類の写像と別表のリコンサイル計画はトランザクションを開く前に組み立てる
	// (createWorkと対称、トランザクション内は永続化のみとするルール1)。送信されたparamsを
	// *model.Workに射影し (編集フォームが触れないカラムは保持)、フェーズ2同期の写像ヘルパーを
	// 再利用してwork -> anime / 分類 / 別表 の写像の正本を1つに保つ。これにより更新直後の
	// 同期はUnchangedを報告する。別表のあるべき行導出が行をマッピング済みanimeに紐付けられる
	// ようwork.AnimeIDをセットする。
	var animeParams repository.UpdateAnimeParams
	var classificationParams repository.UpdateAnimeClassificationParams
	var satellitePlans workSatellitePlans
	if existingAnime != nil {
		work := workFromUpdateWorkParams(params, current)
		work.AnimeID = &existingAnime.ID
		animeParams = animeUpdateParamsFromWork(work, existingAnime)
		classificationParams = classificationUpdateParamsFromWork(work, existingAnime.ID)
		satellitePlans = planWorkSatellites(work, existingSatellites)
	}

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	updated, err := uc.workRepo.WithTx(tx).Update(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("作品の更新に失敗しました: %w", err)
	}
	// 送信が名乗った版にどの行も一致しなかった。編集フォームを開いてから本送信までの間に、
	// 他者がその作品を書いたということ。送信は適用せず却下する。どちらの値を残すかを決めるのは
	// 編集者であり、呼び出し側が両方を示すため。
	if !updated {
		return nil, &model.AppError{
			Code:     model.AppErrCodeConflict,
			UserMsg:  i18n.T(ctx, "validation_version_conflict"),
			Metadata: map[string]string{"work_id": params.ID.String()},
		}
	}

	if existingAnime != nil {
		if err := uc.animeRepo.WithTx(tx).Update(ctx, animeParams); err != nil {
			return nil, fmt.Errorf("animeの更新に失敗しました: %w", err)
		}
		if err := uc.animeClassificationRepo.WithTx(tx).UpdateByAnimeID(ctx, classificationParams); err != nil {
			return nil, fmt.Errorf("anime_classificationの更新に失敗しました: %w", err)
		}
		if err := applyWorkSatellitePlans(ctx, uc.satelliteRepos.WithTx(tx), satellitePlans); err != nil {
			return nil, fmt.Errorf("別表テーブルの両書きに失敗しました: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return &UpdateWorkOutput{WorkID: params.ID}, nil
}

// workFromUpdateWorkParamsはUpdateWorkParamsを、workFromCreateWorkParamsを
// フォームカラムに再利用して *model.Workに射影し、編集フォームが送信しないanime写像
// カラムを現在のworks行から引き継ぐ: title_ro、およびanime.statusを
// animeUpdateParamsFromWorkが導出する作品状態のsource (unpublished_at / deleted_at)。
// 状態タイムスタンプを引き継ぐことが、内容編集でアーカイブ済み / 削除済みのanimeを
// publishedに戻してしまうのを防ぐ。更新はこれらのカラムを変えないため、更新後のanimeが
// 更新後のworks行を写し、更新直後の同期はUnchangedを報告する。
func workFromUpdateWorkParams(params repository.UpdateWorkParams, current *model.Work) *model.Work {
	work := workFromCreateWorkParams(params.CreateWorkParams)
	work.TitleRo = current.TitleRo
	work.UnpublishedAt = current.UnpublishedAt
	work.DeletedAt = current.DeletedAt
	return work
}

// buildUpdateWorkParamsは編集フォーム入力をUpdateWorkParamsに変換する。共有の
// buildWorkFormParamsが共通のworksカラムを生成し、本関数が対象IDと、更新が照合する版を
// 足す。版は入力から読み直さずバリデーターがパースしたものを受け取る。検査された値がそのまま
// UPDATEに届くようにするため。
func buildUpdateWorkParams(input UpdateWorkInput, version *time.Time) (repository.UpdateWorkParams, error) {
	common, err := buildWorkFormParams(input.WorkFormInput)
	if err != nil {
		return repository.UpdateWorkParams{}, err
	}
	return repository.UpdateWorkParams{
		ID:               input.WorkID,
		Version:          version,
		CreateWorkParams: common,
	}, nil
}
