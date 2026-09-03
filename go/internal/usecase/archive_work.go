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

// ArchiveWorkUsecase archives (unpublishes) a work from the Annict DB admin UI. Like the
// create / update usecases it is anchored on animes: it sets works.unpublished_at (the
// work-state source of truth) and, in the same transaction, dual-writes the derived
// anime.status = archived. The status is derived through model.Work.DerivedStatus /
// animeUpdateParamsFromWork from the timestamp we set, so a phase 2 reconciliation right
// after this archive reads the same unpublished_at and reports Unchanged (no clobber back
// to published).
//
// Authorization (committer) runs in this usecase before any read, and the RequireCommitter
// middleware on the route rejects the same request earlier. Keeping the check here means a
// caller reaching the usecase outside that route needs the same permission.
//
// [Ja] ArchiveWorkUsecase は Annict DB 管理画面から作品を非公開 (アーカイブ) にする。
// 作成 / 更新 UseCase と同じく animes を基点とし、works.unpublished_at (作品状態の正本) を
// 立て、同一トランザクションで導出した anime.status = archived を両書きする。status は設定
// した timestamp から model.Work.DerivedStatus / animeUpdateParamsFromWork を通じて導出
// されるため、このアーカイブ直後にフェーズ 2 のリコンシリエーションが走っても同じ
// unpublished_at を読んで Unchanged を報告する (published への差し戻し = クロッバーが起き
// ない)。
//
// 認可 (committer) は読み取りより先に本 UseCase で行い、ルートの RequireCommitter middleware も
// 同じリクエストを手前で拒否する。UseCase 側に検査を残すことで、そのルート以外から到達した
// 呼び出し元にも同じ権限を要求する。
type ArchiveWorkUsecase struct {
	db        *sql.DB
	workRepo  *repository.WorkRepository
	animeRepo *repository.AnimeRepository
}

func NewArchiveWorkUsecase(
	db *sql.DB,
	workRepo *repository.WorkRepository,
	animeRepo *repository.AnimeRepository,
) *ArchiveWorkUsecase {
	return &ArchiveWorkUsecase{
		db:        db,
		workRepo:  workRepo,
		animeRepo: animeRepo,
	}
}

// ArchiveWorkInput identifies the work to archive and the user authorizing the write.
//
// [Ja] ArchiveWorkInput はアーカイブする作品と、書き込みを認可するユーザーを指定する。
type ArchiveWorkInput struct {
	User   *model.User
	WorkID model.WorkID
}

type ArchiveWorkOutput struct {
	WorkID model.WorkID
}

func (uc *ArchiveWorkUsecase) Execute(ctx context.Context, input ArchiveWorkInput) (*ArchiveWorkOutput, error) {
	if input.User == nil || !input.User.IsCommitter() {
		return nil, uc.forbidden(ctx, input.WorkID)
	}

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

	// Only a currently published work can be archived, matching the Rails scope
	// Work.without_deleted.published used by Db::WorkPublishingsController#destroy. An
	// already-archived or deleted work is reported as not found (Rails raises
	// RecordNotFound from the scoped find), so a stale confirmation submit turns into a 404
	// rather than re-stamping unpublished_at or, worse, deriving anime.status = deleted from
	// a deleted work.
	//
	// [Ja] アーカイブできるのは現在公開中の work だけで、これは Db::WorkPublishingsController
	// #destroy が使う Rails の scope Work.without_deleted.published に一致する。すでにアーカイブ
	// 済み・削除済みの work は not found として扱い (Rails は scoped find で RecordNotFound を
	// 送出する)、古い確認画面からの送信が unpublished_at の再スタンプや、より悪い「削除済み
	// work から anime.status = deleted を導出」ではなく 404 になるようにする。
	if current.DerivedStatus() != model.WorkStatusPublished {
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

	return uc.archiveWork(ctx, current, existingAnime, time.Now())
}

// archiveWork persists the archive across works and, when the work is mapped, its anime in
// a single transaction. Setting current.UnpublishedAt before building the anime params
// makes animeUpdateParamsFromWork derive status = archived through DerivedStatus, keeping
// the works and anime states in step. The params are built before BeginTx so the
// transaction body performs persistence only (write-usecase rule 1).
//
// [Ja] archiveWork はアーカイブを works に、そして work がマッピング済みならその anime にも
// 1 トランザクションで永続化する。anime パラメータを組み立てる前に current.UnpublishedAt を
// セットすることで、animeUpdateParamsFromWork が DerivedStatus を通じて status = archived を
// 導出し、works と anime の状態を揃える。パラメータは BeginTx の前に組み立て、トランザクション
// 本体は永続化のみとする (書き込み UseCase のルール 1)。
func (uc *ArchiveWorkUsecase) archiveWork(ctx context.Context, current *model.Work, existingAnime *model.Anime, now time.Time) (*ArchiveWorkOutput, error) {
	current.UnpublishedAt = &now

	var animeParams repository.UpdateAnimeParams
	if existingAnime != nil {
		animeParams = animeUpdateParamsFromWork(current, existingAnime)
	}

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := uc.workRepo.WithTx(tx).UpdateUnpublishedAt(ctx, current.ID, &now); err != nil {
		return nil, fmt.Errorf("作品の非公開に失敗しました: %w", err)
	}

	if existingAnime != nil {
		if err := uc.animeRepo.WithTx(tx).Update(ctx, animeParams); err != nil {
			return nil, fmt.Errorf("anime の更新に失敗しました: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return &ArchiveWorkOutput{WorkID: current.ID}, nil
}

// notFound builds the resource-not-found error the handler maps to a 404.
//
// [Ja] notFound は Handler が 404 に写像するリソース未存在エラーを組み立てる。
func (uc *ArchiveWorkUsecase) notFound(ctx context.Context, workID model.WorkID) error {
	return &model.AppError{
		Code:     model.AppErrCodeResourceNotFound,
		UserMsg:  i18n.T(ctx, "error_work_not_found"),
		Metadata: map[string]string{"work_id": workID.String()},
	}
}

// forbidden builds the permission error the handler maps to a 403.
//
// [Ja] forbidden は Handler が 403 に写像する権限エラーを組み立てる。
func (uc *ArchiveWorkUsecase) forbidden(ctx context.Context, workID model.WorkID) error {
	return &model.AppError{
		Code:     model.AppErrCodeForbidden,
		UserMsg:  i18n.T(ctx, "error_forbidden"),
		Metadata: map[string]string{"work_id": workID.String()},
	}
}
