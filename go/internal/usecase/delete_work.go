package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// DeleteWorkUsecase soft-deletes a work from the Annict DB admin UI. Like the archive /
// update usecases it is anchored on animes: it sets works.deleted_at (the work-state source
// of truth) and, in the same transaction, dual-writes the derived anime.status = deleted.
// The delete is a soft delete only (ADR 0004: animes has no physical delete); no child
// resources are cascaded, since the parent's soft-deleted state governs visibility. The
// status is derived through model.Work.DerivedStatus / animeUpdateParamsFromWork from the
// timestamp we set, so a phase 2 reconciliation right after this delete reads the same
// deleted_at and reports Unchanged (no clobber back to published).
//
// Authorization (admin) is enforced by the RequireAdmin middleware on the route, consistent
// with how the other db_work write endpoints gate roles at the middleware, so this usecase
// does not re-check it.
//
// [Ja] DeleteWorkUsecase は Annict DB 管理画面から作品をソフトデリートする。アーカイブ /
// 更新 UseCase と同じく animes を基点とし、works.deleted_at (作品状態の正本) を立て、同一
// トランザクションで導出した anime.status = deleted を両書きする。削除はソフトデリートのみ
// (ADR 0004: animes は物理削除を持たない) で、子リソースへのカスケードは行わない (親の
// ソフトデリート状態が可視性を支配する)。status は設定した timestamp から
// model.Work.DerivedStatus / animeUpdateParamsFromWork を通じて導出されるため、この削除
// 直後にフェーズ 2 のリコンシリエーションが走っても同じ deleted_at を読んで Unchanged を
// 報告する (published への差し戻し = クロッバーが起きない)。
//
// 認可 (admin) はルートの RequireAdmin middleware で強制する。他の db_work 書き込み
// エンドポイントが middleware でロールをゲートするのと揃えており、本 UseCase では再チェック
// しない。
type DeleteWorkUsecase struct {
	db        *sql.DB
	workRepo  *repository.WorkRepository
	animeRepo *repository.AnimeRepository
}

func NewDeleteWorkUsecase(
	db *sql.DB,
	workRepo *repository.WorkRepository,
	animeRepo *repository.AnimeRepository,
) *DeleteWorkUsecase {
	return &DeleteWorkUsecase{
		db:        db,
		workRepo:  workRepo,
		animeRepo: animeRepo,
	}
}

type DeleteWorkInput struct {
	WorkID model.WorkID
}

type DeleteWorkOutput struct {
	WorkID model.WorkID
}

func (uc *DeleteWorkUsecase) Execute(ctx context.Context, input DeleteWorkInput) (*DeleteWorkOutput, error) {
	// Load the work via the anime-sync projection: it carries works.anime_id and the
	// anime-mapped columns animeUpdateParamsFromWork needs (title_ro / archive_message /
	// the work-state source), so the derived anime write mirrors the works row. An empty
	// result means the work does not exist.
	//
	// [Ja] work を anime 同期の射影で読み込む。これは works.anime_id と、
	// animeUpdateParamsFromWork が必要とする anime 写像カラム (title_ro / archive_message /
	// 作品状態の source) を持ち、導出する anime の書き込みが works 行を写すようにする。
	// 結果が空なら work は存在しない。
	works, err := uc.workRepo.ListForAnimeSyncByIDs(ctx, []model.WorkID{input.WorkID})
	if err != nil {
		return nil, fmt.Errorf("作品の取得に失敗しました: %w", err)
	}
	if len(works) == 0 {
		return nil, uc.notFound(ctx, input.WorkID)
	}
	current := works[0]

	// Only a not-yet-deleted work can be deleted, matching the Rails scope
	// Work.without_deleted used by Db::WorksController#destroy. A published or archived
	// work is deletable; an already soft-deleted work is reported as not found (Rails raises
	// RecordNotFound from the scoped find), so a stale delete submit turns into a 404 rather
	// than re-stamping deleted_at.
	//
	// [Ja] 削除できるのは未削除の work だけで、これは Db::WorksController#destroy が使う
	// Rails の scope Work.without_deleted に一致する。公開中・アーカイブ済みの work は削除
	// 可能で、すでにソフトデリート済みの work は not found として扱う (Rails は scoped find で
	// RecordNotFound を送出する)。古い削除送信が deleted_at の再スタンプではなく 404 になる
	// ようにする。
	if current.DerivedStatus() == model.WorkStatusDeleted {
		return nil, uc.notFound(ctx, input.WorkID)
	}

	// Load the mapped anime so animeUpdateParamsFromWork can carry over the columns animes
	// does not source from works (release_status / archive_message etc.). A nil result
	// means the work is not yet mapped to an anime; the dual-write to animes is then
	// skipped and the phase 2 sync creates the anime later with the derived status.
	//
	// [Ja] animeUpdateParamsFromWork が animes 由来でないカラム (release_status /
	// archive_message など) を引き継げるようマッピング済み anime を読み込む。nil は work が
	// 未だ anime にマッピングされていないことを表し、その場合 animes への両書きはスキップし、
	// フェーズ 2 の同期が後で導出済み status の anime を作成する。
	var existingAnime *model.Anime
	if current.AnimeID != nil {
		existingAnime, err = uc.animeRepo.GetByID(ctx, *current.AnimeID)
		if err != nil {
			return nil, fmt.Errorf("anime の取得に失敗しました: %w", err)
		}
	}

	return uc.deleteWork(ctx, current, existingAnime, time.Now())
}

// deleteWork persists the soft delete across works and, when the work is mapped, its anime
// in a single transaction. Setting current.DeletedAt before building the anime params makes
// animeUpdateParamsFromWork derive status = deleted through DerivedStatus (deleted_at wins
// over unpublished_at), keeping the works and anime states in step. The params are built
// before BeginTx so the transaction body performs persistence only (write-usecase rule 1).
//
// [Ja] deleteWork はソフトデリートを works に、そして work がマッピング済みならその anime にも
// 1 トランザクションで永続化する。anime パラメータを組み立てる前に current.DeletedAt を
// セットすることで、animeUpdateParamsFromWork が DerivedStatus を通じて status = deleted を
// 導出し (deleted_at が unpublished_at より優先される)、works と anime の状態を揃える。
// パラメータは BeginTx の前に組み立て、トランザクション本体は永続化のみとする
// (書き込み UseCase のルール 1)。
func (uc *DeleteWorkUsecase) deleteWork(ctx context.Context, current *model.Work, existingAnime *model.Anime, now time.Time) (*DeleteWorkOutput, error) {
	current.DeletedAt = &now

	var animeParams repository.UpdateAnimeParams
	if existingAnime != nil {
		animeParams = animeUpdateParamsFromWork(current, existingAnime)
	}

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := uc.workRepo.WithTx(tx).UpdateDeletedAt(ctx, current.ID, &now); err != nil {
		return nil, fmt.Errorf("作品の削除に失敗しました: %w", err)
	}

	if existingAnime != nil {
		if err := uc.animeRepo.WithTx(tx).Update(ctx, animeParams); err != nil {
			return nil, fmt.Errorf("anime の更新に失敗しました: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return &DeleteWorkOutput{WorkID: current.ID}, nil
}

// notFound builds the resource-not-found error the handler maps to a 404.
//
// [Ja] notFound は Handler が 404 に写像するリソース未存在エラーを組み立てる。
func (uc *DeleteWorkUsecase) notFound(ctx context.Context, workID model.WorkID) error {
	return &model.AppError{
		Code:     model.AppErrCodeResourceNotFound,
		UserMsg:  i18n.T(ctx, "error_work_not_found"),
		Metadata: map[string]string{"work_id": workID.String()},
	}
}
