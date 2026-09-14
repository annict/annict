package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// UnarchiveWorkUsecaseはAnnict DB管理画面から作品を再公開 (アーカイブ解除) にする。
// ArchiveWorkUsecaseの逆で、animesを基点とし、works.unpublished_at (作品状態の正本) を
// クリアし、同一トランザクションで導出したanime.status = publishedを両書きする。statusは
// クリアしたtimestampからmodel.Work.DerivedStatus / animeUpdateParamsFromWorkを通じて
// 導出されるため、この再公開直後にフェーズ2のリコンシリエーションが走っても同じ
// unpublished_at (NULL) を読んでUnchangedを報告する (archivedへの差し戻し = クロッバーが
// 起きない)。
//
// 認可 (committer) は読み取りより先に本UseCaseで行い、ルートのRequireCommitter middlewareも
// 同じリクエストを手前で拒否する。UseCase側に検査を残すことで、そのルート以外から到達した
// 呼び出し元にも同じ権限を要求する。
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

// UnarchiveWorkInputは再公開する作品と、書き込みを認可するユーザーを指定する。
type UnarchiveWorkInput struct {
	User   *model.User
	WorkID model.WorkID
}

type UnarchiveWorkOutput struct {
	WorkID model.WorkID
}

func (uc *UnarchiveWorkUsecase) Execute(ctx context.Context, input UnarchiveWorkInput) (*UnarchiveWorkOutput, error) {
	if input.User == nil || !input.User.IsCommitter() {
		return nil, uc.forbidden(ctx, input.WorkID)
	}

	// workをanime同期の射影で読み込む。これはworks.anime_idと、
	// animeUpdateParamsFromWorkが必要とするanime写像カラム (title_ro / archive_message /
	// 作品状態のsource) を持ち、導出するanimeの書き込みがworks行を写すようにする。
	// 結果が空ならworkは存在しない。
	works, err := uc.workRepo.ListForAnimeSyncByIDs(ctx, []model.WorkID{input.WorkID})
	if err != nil {
		return nil, fmt.Errorf("作品の取得に失敗しました: %w", err)
	}
	if len(works) == 0 {
		return nil, uc.notFound(ctx, input.WorkID)
	}
	current := works[0]

	// 再公開できるのは現在アーカイブ済みのworkだけで、これは
	// Db::WorkPublishingsController#createが使うRailsのscope
	// Work.without_deleted.unpublishedに一致する。すでに公開中・削除済みのworkはnot found
	// として扱い (Railsはscoped findでRecordNotFoundを送出する)、古い再公開画面からの送信が
	// 公開中のworkのunpublished_atクリアや、より悪い「削除済みworkから
	// anime.status = publishedを導出」ではなく404になるようにする。
	if current.DerivedStatus() != model.WorkStatusArchived {
		return nil, uc.notFound(ctx, input.WorkID)
	}

	// animeUpdateParamsFromWorkがanimes由来でないカラム (release_status /
	// archive_messageなど) を引き継げるようマッピング済みanimeを読み込む。nilはworkが
	// 未だanimeにマッピングされていないことを表し、その場合animesへの両書きはスキップし、
	// フェーズ2の同期が後で導出済みstatusのanimeを作成する。
	var existingAnime *model.Anime
	if current.AnimeID != nil {
		existingAnime, err = uc.animeRepo.GetByID(ctx, *current.AnimeID)
		if err != nil {
			return nil, fmt.Errorf("animeの取得に失敗しました: %w", err)
		}
	}

	return uc.unarchiveWork(ctx, current, existingAnime)
}

// unarchiveWorkは再公開をworksに、そしてworkがマッピング済みならそのanimeにも
// 1トランザクションで永続化する。animeパラメータを組み立てる前にcurrent.UnpublishedAtを
// クリアすることで、animeUpdateParamsFromWorkがDerivedStatusを通じてstatus = publishedを
// 導出し、worksとanimeの状態を揃える。パラメータはBeginTxの前に組み立て、トランザクション
// 本体は永続化のみとする (書き込みUseCaseのルール1)。
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
			return nil, fmt.Errorf("animeの更新に失敗しました: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return &UnarchiveWorkOutput{WorkID: current.ID}, nil
}

// notFoundはHandlerが404に写像するリソース未存在エラーを組み立てる。
func (uc *UnarchiveWorkUsecase) notFound(ctx context.Context, workID model.WorkID) error {
	return &model.AppError{
		Code:     model.AppErrCodeResourceNotFound,
		UserMsg:  i18n.T(ctx, "error_work_not_found"),
		Metadata: map[string]string{"work_id": workID.String()},
	}
}

// forbiddenはHandlerが403に写像する権限エラーを組み立てる。
func (uc *UnarchiveWorkUsecase) forbidden(ctx context.Context, workID model.WorkID) error {
	return &model.AppError{
		Code:     model.AppErrCodeForbidden,
		UserMsg:  i18n.T(ctx, "error_forbidden"),
		Metadata: map[string]string{"work_id": workID.String()},
	}
}
