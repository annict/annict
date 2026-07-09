package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// UnarchiveWorkUsecase re-publishes (un-archives) a work from the Annict DB admin UI. It
// is the inverse of ArchiveWorkUsecase: anchored on animes, it clears works.unpublished_at
// (the work-state source of truth) and, in the same transaction, dual-writes the derived
// anime.status = published. The status is derived through model.Work.DerivedStatus /
// animeUpdateParamsFromWork from the cleared timestamp, so a phase 2 reconciliation right
// after this re-publish reads the same unpublished_at (now NULL) and reports Unchanged (no
// clobber back to archived).
//
// Authorization (committer) is enforced by the RequireCommitter middleware on the route,
// consistent with the other db_work write endpoints, so this usecase does not re-check it.
//
// [Ja] UnarchiveWorkUsecase は Annict DB 管理画面から作品を再公開 (アーカイブ解除) にする。
// ArchiveWorkUsecase の逆で、animes を基点とし、works.unpublished_at (作品状態の正本) を
// クリアし、同一トランザクションで導出した anime.status = published を両書きする。status は
// クリアした timestamp から model.Work.DerivedStatus / animeUpdateParamsFromWork を通じて
// 導出されるため、この再公開直後にフェーズ 2 のリコンシリエーションが走っても同じ
// unpublished_at (NULL) を読んで Unchanged を報告する (archived への差し戻し = クロッバーが
// 起きない)。
//
// 認可 (committer) はルートの RequireCommitter middleware で強制する。他の db_work 書き込み
// エンドポイントと揃えており、本 UseCase では再チェックしない。
type UnarchiveWorkUsecase struct {
	db        *sql.DB
	workRepo  *repository.WorkRepository
	animeRepo *repository.AnimeRepository
}

func NewUnarchiveWorkUsecase(
	db *sql.DB,
	workRepo *repository.WorkRepository,
	animeRepo *repository.AnimeRepository,
) *UnarchiveWorkUsecase {
	return &UnarchiveWorkUsecase{
		db:        db,
		workRepo:  workRepo,
		animeRepo: animeRepo,
	}
}

type UnarchiveWorkInput struct {
	WorkID model.WorkID
}

type UnarchiveWorkOutput struct {
	WorkID model.WorkID
}

func (uc *UnarchiveWorkUsecase) Execute(ctx context.Context, input UnarchiveWorkInput) (*UnarchiveWorkOutput, error) {
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

	// Only a currently archived work can be re-published, matching the Rails scope
	// Work.without_deleted.unpublished used by Db::WorkPublishingsController#create. An
	// already-published or deleted work is reported as not found (Rails raises
	// RecordNotFound from the scoped find), so a stale re-publish submit turns into a 404
	// rather than clearing unpublished_at on a published work or, worse, deriving
	// anime.status = published from a deleted work.
	//
	// [Ja] 再公開できるのは現在アーカイブ済みの work だけで、これは
	// Db::WorkPublishingsController#create が使う Rails の scope
	// Work.without_deleted.unpublished に一致する。すでに公開中・削除済みの work は not found
	// として扱い (Rails は scoped find で RecordNotFound を送出する)、古い再公開画面からの送信が
	// 公開中の work の unpublished_at クリアや、より悪い「削除済み work から
	// anime.status = published を導出」ではなく 404 になるようにする。
	if current.DerivedStatus() != model.WorkStatusArchived {
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

	return uc.unarchiveWork(ctx, current, existingAnime)
}

// unarchiveWork persists the re-publish across works and, when the work is mapped, its
// anime in a single transaction. Clearing current.UnpublishedAt before building the anime
// params makes animeUpdateParamsFromWork derive status = published through DerivedStatus,
// keeping the works and anime states in step. The params are built before BeginTx so the
// transaction body performs persistence only (write-usecase rule 1).
//
// [Ja] unarchiveWork は再公開を works に、そして work がマッピング済みならその anime にも
// 1 トランザクションで永続化する。anime パラメータを組み立てる前に current.UnpublishedAt を
// クリアすることで、animeUpdateParamsFromWork が DerivedStatus を通じて status = published を
// 導出し、works と anime の状態を揃える。パラメータは BeginTx の前に組み立て、トランザクション
// 本体は永続化のみとする (書き込み UseCase のルール 1)。
func (uc *UnarchiveWorkUsecase) unarchiveWork(ctx context.Context, current *model.Work, existingAnime *model.Anime) (*UnarchiveWorkOutput, error) {
	current.UnpublishedAt = nil

	var animeParams repository.UpdateAnimeParams
	if existingAnime != nil {
		animeParams = animeUpdateParamsFromWork(current, existingAnime)
	}

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := uc.workRepo.WithTx(tx).UpdateUnpublishedAt(ctx, current.ID, nil); err != nil {
		return nil, fmt.Errorf("作品の再公開に失敗しました: %w", err)
	}

	if existingAnime != nil {
		if err := uc.animeRepo.WithTx(tx).Update(ctx, animeParams); err != nil {
			return nil, fmt.Errorf("anime の更新に失敗しました: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return &UnarchiveWorkOutput{WorkID: current.ID}, nil
}

// notFound builds the resource-not-found error the handler maps to a 404.
//
// [Ja] notFound は Handler が 404 に写像するリソース未存在エラーを組み立てる。
func (uc *UnarchiveWorkUsecase) notFound(ctx context.Context, workID model.WorkID) error {
	return &model.AppError{
		Code:     model.AppErrCodeResourceNotFound,
		UserMsg:  i18n.T(ctx, "error_work_not_found"),
		Metadata: map[string]string{"work_id": workID.String()},
	}
}
