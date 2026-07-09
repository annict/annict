package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/validator"
)

// UpdateWorkUsecase updates a work from the Annict DB admin edit form. Like
// CreateWorkUsecase it is anchored on animes: it updates the works row (still the
// source of truth during the migration) and dual-writes the mapped anime /
// classification in the same transaction. It reuses DBWorkCreateValidator because the
// edit form validates the identical set of fields.
//
// [Ja] UpdateWorkUsecase は Annict DB 管理画面の編集フォームから作品を更新する。
// CreateWorkUsecase と同じく animes を基点とし、works 行 (移行期間中は正本のまま) を
// 更新しつつ、同一トランザクションでマッピング済みの anime / 分類にも両書きする。
// 編集フォームは作成フォームと同一のフィールドを検証するため DBWorkCreateValidator を
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

// UpdateWorkInput carries the target work ID plus the shared work form values. It embeds
// WorkFormInput so create and update share the field set, the validator input, and the
// string->typed conversion.
//
// [Ja] UpdateWorkInput は対象の work ID と共有の作品フォーム入力値を保持する。WorkFormInput
// を埋め込み、フィールド集合・バリデーター入力・文字列→型変換を作成と共有する。
type UpdateWorkInput struct {
	WorkID model.WorkID
	WorkFormInput
}

type UpdateWorkOutput struct {
	WorkID model.WorkID
}

func (uc *UpdateWorkUsecase) Execute(ctx context.Context, input UpdateWorkInput) (*UpdateWorkOutput, error) {
	if err := uc.validator.Validate(ctx, input.toValidatorInput()); err != nil {
		return nil, err
	}

	// Load the work being edited via the anime-sync projection: it carries the
	// works.anime_id mapping and the anime-mapped columns the edit form does not touch
	// (title_ro / archive_message / the work-state source unpublished_at / deleted_at,
	// from which anime.status is derived), both needed to update the mapped anime without
	// clobbering those columns. An empty result means the work was deleted between the
	// edit GET and this PATCH.
	//
	// [Ja] 編集対象の work を anime 同期の射影で読み込む。これは works.anime_id の
	// マッピングと、編集フォームが触れない anime 写像カラム (title_ro / archive_message /
	// anime.status の導出源である作品状態の source unpublished_at / deleted_at) を持ち、
	// いずれもマッピング済み anime をそれらのカラムを潰さずに更新するために要る。結果が
	// 空の場合は編集 GET とこの PATCH の間に work が削除された。
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

	// Load the existing anime so animeUpdateParamsFromWork can carry over the columns
	// animes does not source from works (release_status etc.). A nil result means the
	// work is not yet mapped to an anime; the dual-write to animes is then skipped (see
	// updateWork).
	//
	// [Ja] animeUpdateParamsFromWork が animes 由来でないカラム (release_status など) を
	// 引き継げるよう既存 anime を読み込む。nil は work が未だ anime にマッピングされて
	// いないことを表し、その場合 animes への両書きはスキップする (updateWork を参照)。
	var existingAnime *model.Anime
	var existingSatellites workSatelliteExisting
	if current.AnimeID != nil {
		existingAnime, err = uc.animeRepo.GetByID(ctx, *current.AnimeID)
		if err != nil {
			return nil, fmt.Errorf("anime の取得に失敗しました: %w", err)
		}
		// When the work is mapped, read the anime's existing satellite rows before the
		// transaction (the write-usecase rule keeps I/O out of the transaction body) so
		// updateWork can reconcile them against the submitted values. GetByID returns nil
		// for an anime_id pointing at a missing anime; that case skips the read and the
		// anime / satellite dual-write, same as a NULL anime_id.
		//
		// [Ja] work がマッピング済みのときは、更新値と突合できるよう anime の既存別表行を
		// トランザクションの前に読む (書き込み UseCase のルールで I/O をトランザクション本体の
		// 外に出す)。anime_id が存在しない anime を指す場合 GetByID は nil を返し、その場合は
		// NULL の anime_id と同じく読み込みと anime / 別表の両書きをスキップする。
		if existingAnime != nil {
			existingSatellites, err = readWorkSatelliteExisting(ctx, uc.satelliteRepos, existingAnime.ID)
			if err != nil {
				return nil, fmt.Errorf("既存別表行の取得に失敗しました: %w", err)
			}
		}
	}

	params, err := buildUpdateWorkParams(input)
	if err != nil {
		return nil, fmt.Errorf("入力値の変換に失敗: %w", err)
	}

	return uc.updateWork(ctx, params, current, existingAnime, existingSatellites)
}

// updateWork persists the update across works and, when the work is already mapped,
// its anime / anime_classification and the six satellite tables in a single transaction.
// works stays the source of truth during the migration, so the anime and satellite writes
// are kept in one block that the cutover (phase 17) can remove wholesale.
//
// When the work is not yet mapped (existingAnime == nil), only works is updated and the
// anime is left for the phase 2 sync batch to create. The sync is the arbiter and the
// sole creator of animes for existing works, so the usecase does not create one here:
// the work is already visible to the batch with anime_id NULL, and creating an anime
// concurrently would race the batch into two animes for one work.
//
// [Ja] updateWork は更新を works に、そして work が既にマッピング済みなら その anime /
// anime_classification と 6 つの別表にも 1 トランザクションで永続化する。移行期間中は works が
// 正本のため、anime と別表への書き込みは正本切り替え (フェーズ 17) でまるごと外せるよう 1
// ブロックにまとめてある。
//
// work が未マッピング (existingAnime == nil) の場合は works だけを更新し、anime の作成は
// フェーズ 2 の同期バッチに委ねる。同期は裁定者であり既存 work に対する anime の唯一の
// 作成者なので、ここでは作成しない。work は anime_id NULL のまま既にバッチから見えており、
// ここで並行して anime を作ると 1 つの work に 2 つの anime ができる競合になる。
func (uc *UpdateWorkUsecase) updateWork(ctx context.Context, params repository.UpdateWorkParams, current *model.Work, existingAnime *model.Anime, existingSatellites workSatelliteExisting) (*UpdateWorkOutput, error) {
	// Build the anime / classification mapping and the satellite reconcile plans before
	// opening the transaction, mirroring createWork and keeping the transaction body to
	// persistence only (usecase rule 1). Projecting the submitted params onto a *model.Work
	// (preserving the columns the edit form does not touch) and reusing the phase 2 sync
	// mapping helpers keeps the work -> anime / classification / satellite mapping
	// single-sourced, so a sync run right after this update reports Unchanged. work.AnimeID
	// is set so the satellite desired-row derivation anchors the rows on the mapped anime.
	//
	// [Ja] anime / 分類の写像と別表のリコンサイル計画はトランザクションを開く前に組み立てる
	// (createWork と対称、トランザクション内は永続化のみとするルール 1)。送信された params を
	// *model.Work に射影し (編集フォームが触れないカラムは保持)、フェーズ 2 同期の写像ヘルパーを
	// 再利用して work -> anime / 分類 / 別表 の写像の正本を 1 つに保つ。これにより更新直後の
	// 同期は Unchanged を報告する。別表のあるべき行導出が行をマッピング済み anime に紐付けられる
	// よう work.AnimeID をセットする。
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

	if err := uc.workRepo.WithTx(tx).Update(ctx, params); err != nil {
		return nil, fmt.Errorf("作品の更新に失敗しました: %w", err)
	}

	if existingAnime != nil {
		if err := uc.animeRepo.WithTx(tx).Update(ctx, animeParams); err != nil {
			return nil, fmt.Errorf("anime の更新に失敗しました: %w", err)
		}
		if err := uc.animeClassificationRepo.WithTx(tx).UpdateByAnimeID(ctx, classificationParams); err != nil {
			return nil, fmt.Errorf("anime_classification の更新に失敗しました: %w", err)
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

// workFromUpdateWorkParams projects an UpdateWorkParams onto a *model.Work by reusing
// workFromCreateWorkParams for the form columns, then carrying over the anime-mapped
// columns the edit form does not submit from the current works row: title_ro and the
// work-state source (unpublished_at / deleted_at), from which animeUpdateParamsFromWork
// derives anime.status. Carrying the state timestamps over is
// what keeps a content edit from clobbering an archived / deleted anime back to published.
// The update never changes those columns, so the updated anime mirrors the post-update
// works row and the sync right after the update reports Unchanged.
//
// [Ja] workFromUpdateWorkParams は UpdateWorkParams を、workFromCreateWorkParams を
// フォームカラムに再利用して *model.Work に射影し、編集フォームが送信しない anime 写像
// カラムを現在の works 行から引き継ぐ: title_ro、および anime.status を
// animeUpdateParamsFromWork が導出する作品状態の source (unpublished_at / deleted_at)。
// 状態タイムスタンプを引き継ぐことが、内容編集でアーカイブ済み / 削除済みの anime を
// published に戻してしまうのを防ぐ。更新はこれらのカラムを変えないため、更新後の anime が
// 更新後の works 行を写し、更新直後の同期は Unchanged を報告する。
func workFromUpdateWorkParams(params repository.UpdateWorkParams, current *model.Work) *model.Work {
	work := workFromCreateWorkParams(params.CreateWorkParams)
	work.TitleRo = current.TitleRo
	work.UnpublishedAt = current.UnpublishedAt
	work.DeletedAt = current.DeletedAt
	return work
}

// buildUpdateWorkParams converts the edit form input into UpdateWorkParams: the shared
// buildWorkFormParams produces the common works columns and this adds the target ID.
//
// [Ja] buildUpdateWorkParams は編集フォーム入力を UpdateWorkParams に変換する。共有の
// buildWorkFormParams が共通の works カラムを生成し、本関数が対象 ID を足す。
func buildUpdateWorkParams(input UpdateWorkInput) (repository.UpdateWorkParams, error) {
	common, err := buildWorkFormParams(input.WorkFormInput)
	if err != nil {
		return repository.UpdateWorkParams{}, err
	}
	return repository.UpdateWorkParams{
		ID:               input.WorkID,
		CreateWorkParams: common,
	}, nil
}
