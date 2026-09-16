package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// UnarchiveEpisodeUsecaseはAnnict DB管理画面からエピソードを再公開 (アーカイブ解除)
// にする。ArchiveEpisodeUsecaseの逆で、エピソード状態の正本であるepisodes.unpublished_atを
// クリアし、同一トランザクションで導出したanime.status = publishedだけを両書きする。両方向とも
// 同じタイムスタンプからanimeのstatusを導出するため、直後にフェーズ2のリコンシリエーション
// が走ってもUnchangedとなり、animeがarchivedに差し戻される (クロッバー) ことはない。
//
// HTTPルートのRequireCommitter middlewareに加えて、すべてのentry pointで同じ規則になる
// ようUseCase自身もcommitter認可を強制する。Railsの再公開はdb_activityを記録しないため、
// 編集者は認可に使うが活動履歴の作成者としては記録しない。
//
// episodesの更新は、行が今も非公開であり、トランザクション前の射影が観測した作品に今も属し、
// その作品も未削除の場合だけ行う。親作品のtouchとカウンター更新にも成功した場合だけ結果を返し、
// 親作品が同時に削除された場合はトランザクションをロールバックする。実際に再公開した行から
// anime_idを返すため、写像が同時に変わっても以前のanimeへstatusを書かない。statusだけを
// 更新することで、事前読み取り後にコミットされたanimeの内容も保持する。
type UnarchiveEpisodeUsecase struct {
	db          *sql.DB
	episodeRepo *repository.EpisodeRepository
	animeRepo   *repository.AnimeRepository
}

// NewUnarchiveEpisodeUsecaseはUnarchiveEpisodeUsecaseを生成する。
func NewUnarchiveEpisodeUsecase(
	db *sql.DB,
	episodeRepo *repository.EpisodeRepository,
	animeRepo *repository.AnimeRepository,
) *UnarchiveEpisodeUsecase {
	return &UnarchiveEpisodeUsecase{
		db:          db,
		episodeRepo: episodeRepo,
		animeRepo:   animeRepo,
	}
}

// UnarchiveEpisodeInputは再公開するエピソードと、書き込みを認可するユーザーを指定する。
type UnarchiveEpisodeInput struct {
	EpisodeID model.EpisodeID
	User      *model.User
}

// UnarchiveEpisodeOutputは再公開したエピソードと、その所属作品 (呼び出し元がリダイレクト
// 先にする) を報告する。
type UnarchiveEpisodeOutput struct {
	EpisodeID model.EpisodeID
	WorkID    model.WorkID
}

// Executeは送信が名指ししたエピソードを再公開する。
//
// 認可は読み取りより先に行う。直接の呼び出し元もHTTPルートと同じcommitter権限を持たなければ、
// リソースの存在を判別できないようにするため。
func (uc *UnarchiveEpisodeUsecase) Execute(ctx context.Context, input UnarchiveEpisodeInput) (*UnarchiveEpisodeOutput, error) {
	if input.User == nil || !input.User.IsCommitter() {
		return nil, &model.AppError{
			Code:     model.AppErrCodeForbidden,
			UserMsg:  i18n.T(ctx, "error_forbidden"),
			Metadata: map[string]string{"episode_id": input.EpisodeID.String()},
		}
	}

	// 非公開エンドポイント群が共有するローダーから、この送信が名指しするとみなす親作品と
	// 現在のライフサイクルのタイムスタンプを得る。結果が空の場合、または非公開でない場合、編集者
	// が操作した一覧は古い (Railsも再公開をEpisode.without_deleted.unpublishedに絞り、外れれば
	// RecordNotFoundを送出する)。この読み取りはトランザクション外のため、Unarchiveは状態、
	// 親作品の同一性、親作品のライフサイクル条件をSQLでも繰り返す。
	target, err := uc.episodeRepo.GetForArchiveByID(ctx, input.EpisodeID)
	if err != nil {
		return nil, fmt.Errorf("エピソードの取得に失敗しました: %w", err)
	}
	if target == nil || target.Episode.DerivedStatus() != model.EpisodeStatusArchived {
		return nil, episodeNotFoundError(ctx, input.EpisodeID)
	}

	return uc.unarchiveEpisode(ctx, target.Episode)
}

// unarchiveEpisodeは再公開をepisodesに、そして実際に更新した行がマッピング済みならその
// animeにも1トランザクションで永続化する。episodesへの書き込みを先に行うのは、状態と親作品の
// 最終的なガードをそこが持つため。結果が現在のanimeの写像を運び、そのanimeのstatusだけを
// 更新するため、内容属性には触れない。
func (uc *UnarchiveEpisodeUsecase) unarchiveEpisode(
	ctx context.Context,
	current *model.Episode,
) (*UnarchiveEpisodeOutput, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	unarchived, err := uc.episodeRepo.WithTx(tx).Unarchive(ctx, repository.UnarchiveEpisodeParams{
		ID:     current.ID,
		WorkID: current.WorkID,
	})
	if err != nil {
		return nil, fmt.Errorf("エピソードの再公開に失敗しました: %w", err)
	}
	if unarchived == nil {
		return nil, episodeNotFoundError(ctx, current.ID)
	}

	if unarchived.AnimeID != nil {
		if err := uc.animeRepo.WithTx(tx).UpdateStatus(ctx, *unarchived.AnimeID, model.AnimeStatusPublished); err != nil {
			return nil, fmt.Errorf("animeの状態更新に失敗しました: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return &UnarchiveEpisodeOutput{EpisodeID: current.ID, WorkID: current.WorkID}, nil
}
