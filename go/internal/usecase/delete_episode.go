package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// DeleteEpisodeUsecaseはAnnict DB管理画面からエピソードをソフトデリートする。削除に
// ついてのエピソード状態の正本であるepisodes.deleted_atを立て、同一トランザクションで導出した
// anime.status = deletedだけを両書きする。Railsのdestroy_in_batchesと違い削除はソフト
// デリートのみで (ADR 0004: animesは物理削除を持たない)、子リソースへのカスケードは行わない
// (削除状態が可視性を支配する)。statusは本UseCaseが立てたtimestampから導出されるため、直後
// にフェーズ2のリコンシリエーションが走ってもUnchangedとなり、animeがpublishedに差し戻さ
// れる (クロッバー) ことはない。
//
// HTTPルートのRequireAdmin middlewareに加えて、すべてのentry pointで同じ規則になるよう
// UseCase自身もadmin認可を強制する。非公開がcommitterに開かれているのに対し削除がadmin
// 専用なのはADR 0009の権限分離で、RailsのEpisodePolicyが行う分割とも同じ。Railsの削除は
// db_activityを記録しないため、管理者は認可に使うが活動履歴の作成者としては記録しない。
//
// episodesの更新は、行がまだ削除されておらず、トランザクション前の射影が観測した作品に今も属し、
// その作品も未削除の場合だけ行う。親作品のtouchとカウンター更新にも成功した場合だけ結果を返し、
// 親作品が同時に削除された場合はトランザクションをロールバックする。実際に削除した行から
// anime_idを返すため、写像が同時に変わっても以前のanimeへstatusを書かない。statusだけを
// 更新することで、事前読み取り後にコミットされたanimeの内容も保持する。
type DeleteEpisodeUsecase struct {
	db          *sql.DB
	episodeRepo *repository.EpisodeRepository
	animeRepo   *repository.AnimeRepository
}

// NewDeleteEpisodeUsecaseはDeleteEpisodeUsecaseを生成する。
func NewDeleteEpisodeUsecase(
	db *sql.DB,
	episodeRepo *repository.EpisodeRepository,
	animeRepo *repository.AnimeRepository,
) *DeleteEpisodeUsecase {
	return &DeleteEpisodeUsecase{
		db:          db,
		episodeRepo: episodeRepo,
		animeRepo:   animeRepo,
	}
}

// DeleteEpisodeInputは削除するエピソードと、書き込みを認可するユーザーを指定する。
type DeleteEpisodeInput struct {
	EpisodeID model.EpisodeID
	User      *model.User
}

// DeleteEpisodeOutputは削除したエピソードと、その所属作品 (呼び出し元がリダイレクト先に
// する) を報告する。
type DeleteEpisodeOutput struct {
	EpisodeID model.EpisodeID
	WorkID    model.WorkID
}

// Executeは送信が名指ししたエピソードを削除する。
//
// 認可は読み取りより先に行う。直接の呼び出し元もHTTPルートと同じadmin権限を持たなければ、
// リソースの存在を判別できないようにするため。
func (uc *DeleteEpisodeUsecase) Execute(ctx context.Context, input DeleteEpisodeInput) (*DeleteEpisodeOutput, error) {
	if input.User == nil || !input.User.IsAdmin() {
		return nil, &model.AppError{
			Code:     model.AppErrCodeForbidden,
			UserMsg:  i18n.T(ctx, "error_forbidden"),
			Metadata: map[string]string{"episode_id": input.EpisodeID.String()},
		}
	}

	// 削除用の射影から、この送信が名指しするとみなす親作品を得る。結果が空の場合、管理者が
	// 操作した一覧は古い (Railsも削除をEpisode.without_deletedに絞り、外れれば
	// RecordNotFoundを送出する)。この読み取りはトランザクション外のため、Deleteは状態、親作品
	// の同一性、親作品のライフサイクル条件をSQLでも繰り返す。
	target, err := uc.episodeRepo.GetForDeleteByID(ctx, input.EpisodeID)
	if err != nil {
		return nil, fmt.Errorf("エピソードの取得に失敗しました: %w", err)
	}
	if target == nil {
		return nil, episodeNotFoundError(ctx, input.EpisodeID)
	}

	return uc.deleteEpisode(ctx, target)
}

// deleteEpisodeはソフトデリートをepisodesに、そして実際に更新した行がマッピング済みなら
// そのanimeにも1トランザクションで永続化する。episodesへの書き込みを先に行うのは、状態と
// 親作品の最終的なガードをそこが持つため。結果が現在のanimeの写像を運び、そのanimeのstatus
// だけを更新するため、内容属性には触れない。
func (uc *DeleteEpisodeUsecase) deleteEpisode(
	ctx context.Context,
	current *model.Episode,
) (*DeleteEpisodeOutput, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	deleted, err := uc.episodeRepo.WithTx(tx).Delete(ctx, repository.DeleteEpisodeParams{
		ID:     current.ID,
		WorkID: current.WorkID,
	})
	if err != nil {
		return nil, fmt.Errorf("エピソードの削除に失敗しました: %w", err)
	}
	if deleted == nil {
		return nil, episodeNotFoundError(ctx, current.ID)
	}

	if deleted.AnimeID != nil {
		if err := uc.animeRepo.WithTx(tx).UpdateStatus(ctx, *deleted.AnimeID, model.AnimeStatusDeleted); err != nil {
			return nil, fmt.Errorf("animeの状態更新に失敗しました: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return &DeleteEpisodeOutput{EpisodeID: current.ID, WorkID: current.WorkID}, nil
}
