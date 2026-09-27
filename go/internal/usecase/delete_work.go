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

// DeleteWorkUsecaseはAnnict DB管理画面から作品をソフトデリートする。アーカイブ /
// 更新UseCaseと同じくanimesを基点とし、works.deleted_at (作品状態の正本) を立て、同一
// トランザクションで導出したanime.status = deletedを両書きする。削除はソフトデリートのみ
// (ADR 0004: animesは物理削除を持たない) で、子リソースへのカスケードは行わない (親の
// ソフトデリート状態が可視性を支配する)。statusは設定したtimestampから
// model.Work.DerivedStatus / animeUpdateParamsFromWorkを通じて導出されるため、この削除
// 直後にフェーズ2のリコンシリエーションが走っても同じdeleted_atを読んでUnchangedを
// 報告する (publishedへの差し戻し = クロッバーが起きない)。
//
// 認可 (admin) は読み取りより先に本UseCaseで行い、ルートのRequireAdmin middlewareも同じ
// リクエストを手前で拒否する。UseCase側に検査を残すことで、そのルート以外から到達した
// 呼び出し元にも同じ権限を要求する。
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

// DeleteWorkInputは削除する作品と、書き込みを認可するユーザーを指定する。
type DeleteWorkInput struct {
	User   *model.User
	WorkID model.WorkID
}

type DeleteWorkOutput struct {
	WorkID model.WorkID
}

func (uc *DeleteWorkUsecase) Execute(ctx context.Context, input DeleteWorkInput) (*DeleteWorkOutput, error) {
	if input.User == nil || !input.User.IsAdmin() {
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

	// 削除できるのは未削除のworkだけで、これはDb::WorksController#destroyが使う
	// Railsのscope Work.without_deletedに一致する。公開中・アーカイブ済みのworkは削除
	// 可能で、すでにソフトデリート済みのworkはnot foundとして扱う (Railsはscoped findで
	// RecordNotFoundを送出する)。古い削除送信がdeleted_atの再スタンプではなく404になる
	// ようにする。
	if current.DerivedStatus() == model.WorkStatusDeleted {
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

	return uc.deleteWork(ctx, current, existingAnime, time.Now())
}

// deleteWorkはソフトデリートをworksに、そしてworkがマッピング済みならそのanimeにも
// 1トランザクションで永続化する。animeパラメータを組み立てる前にcurrent.DeletedAtを
// セットすることで、animeUpdateParamsFromWorkがDerivedStatusを通じてstatus = deletedを
// 導出し (deleted_atがunpublished_atより優先される)、worksとanimeの状態を揃える。
// パラメータはBeginTxの前に組み立て、トランザクション本体は永続化のみとする
// (書き込みUseCaseのルール1)。
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
			return nil, fmt.Errorf("animeの更新に失敗しました: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return &DeleteWorkOutput{WorkID: current.ID}, nil
}

// notFoundはHandlerが404に写像するリソース未存在エラーを組み立てる。
func (uc *DeleteWorkUsecase) notFound(ctx context.Context, workID model.WorkID) error {
	return &model.AppError{
		Code:     model.AppErrCodeResourceNotFound,
		UserMsg:  i18n.T(ctx, "error_work_not_found"),
		Metadata: map[string]string{"work_id": workID.String()},
	}
}

// forbiddenはHandlerが403に写像する権限エラーを組み立てる。
func (uc *DeleteWorkUsecase) forbidden(ctx context.Context, workID model.WorkID) error {
	return &model.AppError{
		Code:     model.AppErrCodeForbidden,
		UserMsg:  i18n.T(ctx, "error_forbidden"),
		Metadata: map[string]string{"work_id": workID.String()},
	}
}
