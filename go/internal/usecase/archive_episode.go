package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// ArchiveEpisodeUsecase archives (unpublishes) an episode from the Annict DB admin screen.
// It sets episodes.unpublished_at, the episode-state source of truth, and dual-writes only the
// derived anime.status = archived in the same transaction.
//
// The usecase enforces committer authorization itself so every entry point gets the same rule,
// in addition to the HTTP route's RequireCommitter middleware. The Rails unpublish records no
// db_activity, so the editor is used for authorization but no activity is attributed to them.
//
// The episode update is conditional on the row still being published and still belonging to
// the work observed by the confirmation-page projection. It returns the anime_id from the row
// it actually archived, so a concurrent mapping change cannot send the status write to the
// former anime. Updating only status also preserves anime content committed after the pre-read.
//
// [Ja] ArchiveEpisodeUsecase は Annict DB 管理画面からエピソードを非公開 (アーカイブ) にする。
// エピソード状態の正本である episodes.unpublished_at を立て、同一トランザクションで導出した
// anime.status = archived だけを両書きする。
//
// HTTP ルートの RequireCommitter middleware に加えて、すべての entry point で同じ規則になる
// よう UseCase 自身も committer 認可を強制する。Rails の非公開は db_activity を記録しないため、
// 編集者は認可に使うが活動履歴の作成者としては記録しない。
//
// episodes の更新は、行が今も公開中であり、確認ページ用の射影が観測した作品に今も属する場合だけ
// 行う。実際に非公開にした行から anime_id を返すため、写像が同時に変わっても以前の anime へ
// status を書かない。status だけを更新することで、事前読み取り後にコミットされた anime の内容も
// 保持する。
type ArchiveEpisodeUsecase struct {
	db          *sql.DB
	episodeRepo *repository.EpisodeRepository
	animeRepo   *repository.AnimeRepository
}

// NewArchiveEpisodeUsecase constructs an ArchiveEpisodeUsecase.
//
// [Ja] NewArchiveEpisodeUsecase は ArchiveEpisodeUsecase を生成する。
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

// ArchiveEpisodeInput identifies the episode to archive and the user authorizing the write.
//
// [Ja] ArchiveEpisodeInput は非公開にするエピソードと、書き込みを認可するユーザーを指定する。
type ArchiveEpisodeInput struct {
	EpisodeID model.EpisodeID
	User      *model.User
}

// ArchiveEpisodeOutput reports the archived episode and the work it belongs to, which the
// caller redirects to.
//
// [Ja] ArchiveEpisodeOutput は非公開にしたエピソードと、その所属作品 (呼び出し元がリダイレクト
// 先にする) を報告する。
type ArchiveEpisodeOutput struct {
	EpisodeID model.EpisodeID
	WorkID    model.WorkID
}

// Execute archives the episode the confirmation page named.
//
// Authorization runs before any read so a direct caller cannot use resource existence as an
// oracle without the same committer permission the HTTP route requires.
//
// [Ja] Execute は確認ページが名指ししたエピソードを非公開にする。
//
// 認可は読み取りより先に行う。直接の呼び出し元も HTTP ルートと同じ committer 権限を持たなければ、
// リソースの存在を判別できないようにするため。
func (uc *ArchiveEpisodeUsecase) Execute(ctx context.Context, input ArchiveEpisodeInput) (*ArchiveEpisodeOutput, error) {
	if input.User == nil || !input.User.IsCommitter() {
		return nil, &model.AppError{
			Code:     model.AppErrCodeForbidden,
			UserMsg:  i18n.T(ctx, "error_forbidden"),
			Metadata: map[string]string{"episode_id": input.EpisodeID.String()},
		}
	}

	// The same loader the confirmation page uses supplies the parent observed by that page and
	// the current lifecycle timestamps. An empty result, or an episode that is no longer
	// published, means the confirmation the editor is acting on is stale. Archive repeats both
	// state and parent conditions in SQL because this read is outside its transaction.
	//
	// [Ja] 確認ページと同じローダーから、そのページが観測した親作品と現在のライフサイクルの
	// タイムスタンプを得る。結果が空の場合、または公開中でなくなっている場合、編集者が操作して
	// いる確認は古い。この読み取りはトランザクション外のため、Archive は状態と親作品の条件を
	// SQL でも繰り返す。
	target, err := uc.episodeRepo.GetForArchiveByID(ctx, input.EpisodeID)
	if err != nil {
		return nil, fmt.Errorf("エピソードの取得に失敗しました: %w", err)
	}
	if target == nil || target.Episode.DerivedStatus() != model.EpisodeStatusPublished {
		return nil, episodeNotFoundError(ctx, input.EpisodeID)
	}

	return uc.archiveEpisode(ctx, target.Episode)
}

// archiveEpisode persists the archive across episodes and, when the updated row is mapped, its
// anime in a single transaction. The episodes write comes first because it owns the definitive
// state and parent guards. Its result carries the current anime mapping; only that anime's
// status is then updated, leaving all content attributes untouched.
//
// [Ja] archiveEpisode は非公開を episodes に、そして実際に更新した行がマッピング済みならその
// anime にも 1 トランザクションで永続化する。episodes への書き込みを先に行うのは、状態と親作品の
// 最終的なガードをそこが持つため。結果が現在の anime の写像を運び、その anime の status だけを
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
			return nil, fmt.Errorf("anime の状態更新に失敗しました: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return &ArchiveEpisodeOutput{EpisodeID: current.ID, WorkID: current.WorkID}, nil
}
