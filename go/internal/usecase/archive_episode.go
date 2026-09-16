package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// ArchiveEpisodeUsecaseはAnnict DB管理画面からエピソードを非公開 (アーカイブ) にする。
// エピソード状態の正本であるepisodes.unpublished_atを立て、同一トランザクションで導出した
// anime.status = archivedだけを両書きする。
//
// HTTPルートのRequireCommitter middlewareに加えて、すべてのentry pointで同じ規則になる
// ようUseCase自身もcommitter認可を強制する。Railsの非公開はdb_activityを記録しないため、
// 編集者は認可に使うが活動履歴の作成者としては記録しない。
//
// episodesの更新は、行が今も公開中であり、確認ページ用の射影が観測した作品に今も属し、その作品
// も未削除の場合だけ行う。親作品のtouchとカウンター更新にも成功した場合だけ結果を返し、親作品が
// 同時に削除された場合はトランザクションをロールバックする。実際に非公開にした行からanime_idを
// 返すため、写像が同時に変わっても以前のanimeへstatusを書かない。statusだけを更新することで、
// 事前読み取り後にコミットされたanimeの内容も保持する。
type ArchiveEpisodeUsecase struct {
	db          *sql.DB
	episodeRepo *repository.EpisodeRepository
	animeRepo   *repository.AnimeRepository
}

// NewArchiveEpisodeUsecaseはArchiveEpisodeUsecaseを生成する。
func NewArchiveEpisodeUsecase(
	db *sql.DB,
	episodeRepo *repository.EpisodeRepository,
	animeRepo *repository.AnimeRepository,
) *ArchiveEpisodeUsecase {
	return &ArchiveEpisodeUsecase{
		db:          db,
		episodeRepo: episodeRepo,
		animeRepo:   animeRepo,
	}
}

// ArchiveEpisodeInputは非公開にするエピソードと、書き込みを認可するユーザーを指定する。
type ArchiveEpisodeInput struct {
	EpisodeID model.EpisodeID
	User      *model.User
}

// ArchiveEpisodeOutputは非公開にしたエピソードと、その所属作品 (呼び出し元がリダイレクト
// 先にする) を報告する。
type ArchiveEpisodeOutput struct {
	EpisodeID model.EpisodeID
	WorkID    model.WorkID
}

// Executeは確認ページが名指ししたエピソードを非公開にする。
//
// 認可は読み取りより先に行う。直接の呼び出し元もHTTPルートと同じcommitter権限を持たなければ、
// リソースの存在を判別できないようにするため。
func (uc *ArchiveEpisodeUsecase) Execute(ctx context.Context, input ArchiveEpisodeInput) (*ArchiveEpisodeOutput, error) {
	if input.User == nil || !input.User.IsCommitter() {
		return nil, &model.AppError{
			Code:     model.AppErrCodeForbidden,
			UserMsg:  i18n.T(ctx, "error_forbidden"),
			Metadata: map[string]string{"episode_id": input.EpisodeID.String()},
		}
	}

	// 確認ページと同じローダーから、そのページが観測した親作品と現在のライフサイクルの
	// タイムスタンプを得る。結果が空の場合、または公開中でなくなっている場合、編集者が操作して
	// いる確認は古い。この読み取りはトランザクション外のため、Archiveは状態、親作品の同一性、
	// 親作品のライフサイクル条件をSQLでも繰り返す。
	target, err := uc.episodeRepo.GetForArchiveByID(ctx, input.EpisodeID)
	if err != nil {
		return nil, fmt.Errorf("エピソードの取得に失敗しました: %w", err)
	}
	if target == nil || target.Episode.DerivedStatus() != model.EpisodeStatusPublished {
		return nil, episodeNotFoundError(ctx, input.EpisodeID)
	}

	return uc.archiveEpisode(ctx, target.Episode)
}

// archiveEpisodeは非公開をepisodesに、そして実際に更新した行がマッピング済みならその
// animeにも1トランザクションで永続化する。episodesへの書き込みを先に行うのは、状態と親作品の
// 最終的なガードをそこが持つため。結果が現在のanimeの写像を運び、そのanimeのstatusだけを
// 更新するため、内容属性には触れない。
func (uc *ArchiveEpisodeUsecase) archiveEpisode(
	ctx context.Context,
	current *model.Episode,
) (*ArchiveEpisodeOutput, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	archived, err := uc.episodeRepo.WithTx(tx).Archive(ctx, repository.ArchiveEpisodeParams{
		ID:     current.ID,
		WorkID: current.WorkID,
	})
	if err != nil {
		return nil, fmt.Errorf("エピソードの非公開に失敗しました: %w", err)
	}
	if archived == nil {
		return nil, episodeNotFoundError(ctx, current.ID)
	}

	if archived.AnimeID != nil {
		if err := uc.animeRepo.WithTx(tx).UpdateStatus(ctx, *archived.AnimeID, model.AnimeStatusArchived); err != nil {
			return nil, fmt.Errorf("animeの状態更新に失敗しました: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return &ArchiveEpisodeOutput{EpisodeID: current.ID, WorkID: current.WorkID}, nil
}
