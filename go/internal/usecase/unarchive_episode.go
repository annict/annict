package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// UnarchiveEpisodeUsecase re-publishes (un-archives) an episode from the Annict DB admin screen.
// It is the inverse of ArchiveEpisodeUsecase: it clears episodes.unpublished_at, the
// episode-state source of truth, and dual-writes only the derived anime.status = published in
// the same transaction. Because both directions derive the anime status from the same timestamp,
// a phase 2 reconciliation right after reports Unchanged instead of clobbering the anime back to
// archived.
//
// The usecase enforces committer authorization itself so every entry point gets the same rule,
// in addition to the HTTP route's RequireCommitter middleware. The Rails re-publish records no
// db_activity, so the editor is used for authorization but no activity is attributed to them.
//
// The episode update is conditional on the row still being archived, still belonging to the work
// observed by the pre-transaction projection, and that work still being undeleted. A result is
// returned only when the parent touch and counter update also succeeds, so a concurrent parent
// deletion makes the transaction roll back. It returns the anime_id from the row it actually
// re-published, so a concurrent mapping change cannot send the status write to the
// former anime. Updating only status also preserves anime content committed after the pre-read.
//
// [Ja] UnarchiveEpisodeUsecase は Annict DB 管理画面からエピソードを再公開 (アーカイブ解除)
// にする。ArchiveEpisodeUsecase の逆で、エピソード状態の正本である episodes.unpublished_at を
// クリアし、同一トランザクションで導出した anime.status = published だけを両書きする。両方向とも
// 同じタイムスタンプから anime の status を導出するため、直後にフェーズ 2 のリコンシリエーション
// が走っても Unchanged となり、anime が archived に差し戻される (クロッバー) ことはない。
//
// HTTP ルートの RequireCommitter middleware に加えて、すべての entry point で同じ規則になる
// よう UseCase 自身も committer 認可を強制する。Rails の再公開は db_activity を記録しないため、
// 編集者は認可に使うが活動履歴の作成者としては記録しない。
//
// episodes の更新は、行が今も非公開であり、トランザクション前の射影が観測した作品に今も属し、
// その作品も未削除の場合だけ行う。親作品の touch とカウンター更新にも成功した場合だけ結果を返し、
// 親作品が同時に削除された場合はトランザクションをロールバックする。実際に再公開した行から
// anime_id を返すため、写像が同時に変わっても以前の anime へ status を書かない。status だけを
// 更新することで、事前読み取り後にコミットされた anime の内容も保持する。
type UnarchiveEpisodeUsecase struct {
	db          *sql.DB
	episodeRepo *repository.EpisodeRepository
	animeRepo   *repository.AnimeRepository
}

// NewUnarchiveEpisodeUsecase constructs an UnarchiveEpisodeUsecase.
//
// [Ja] NewUnarchiveEpisodeUsecase は UnarchiveEpisodeUsecase を生成する。
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

// UnarchiveEpisodeInput identifies the episode to re-publish and the user authorizing the write.
//
// [Ja] UnarchiveEpisodeInput は再公開するエピソードと、書き込みを認可するユーザーを指定する。
type UnarchiveEpisodeInput struct {
	EpisodeID model.EpisodeID
	User      *model.User
}

// UnarchiveEpisodeOutput reports the re-published episode and the work it belongs to, which the
// caller redirects to.
//
// [Ja] UnarchiveEpisodeOutput は再公開したエピソードと、その所属作品 (呼び出し元がリダイレクト
// 先にする) を報告する。
type UnarchiveEpisodeOutput struct {
	EpisodeID model.EpisodeID
	WorkID    model.WorkID
}

// Execute re-publishes the episode the submit named.
//
// Authorization runs before any read so a direct caller cannot use resource existence as an
// oracle without the same committer permission the HTTP route requires.
//
// [Ja] Execute は送信が名指ししたエピソードを再公開する。
//
// 認可は読み取りより先に行う。直接の呼び出し元も HTTP ルートと同じ committer 権限を持たなければ、
// リソースの存在を判別できないようにするため。
func (uc *UnarchiveEpisodeUsecase) Execute(ctx context.Context, input UnarchiveEpisodeInput) (*UnarchiveEpisodeOutput, error) {
	if input.User == nil || !input.User.IsCommitter() {
		return nil, &model.AppError{
			Code:     model.AppErrCodeForbidden,
			UserMsg:  i18n.T(ctx, "error_forbidden"),
			Metadata: map[string]string{"episode_id": input.EpisodeID.String()},
		}
	}

	// The loader the archive endpoints share supplies the parent this submit is taken to name
	// and the current lifecycle timestamps. An empty result, or an episode that is not archived,
	// means the list the editor acted on is stale: Rails scopes the re-publish to
	// Episode.without_deleted.unpublished and raises RecordNotFound otherwise. Unarchive repeats
	// the state, parent identity, and parent lifecycle conditions in SQL because this read is
	// outside its transaction.
	//
	// [Ja] 非公開エンドポイント群が共有するローダーから、この送信が名指しするとみなす親作品と
	// 現在のライフサイクルのタイムスタンプを得る。結果が空の場合、または非公開でない場合、編集者
	// が操作した一覧は古い (Rails も再公開を Episode.without_deleted.unpublished に絞り、外れれば
	// RecordNotFound を送出する)。この読み取りはトランザクション外のため、Unarchive は状態、
	// 親作品の同一性、親作品のライフサイクル条件を SQL でも繰り返す。
	target, err := uc.episodeRepo.GetForArchiveByID(ctx, input.EpisodeID)
	if err != nil {
		return nil, fmt.Errorf("エピソードの取得に失敗しました: %w", err)
	}
	if target == nil || target.Episode.DerivedStatus() != model.EpisodeStatusArchived {
		return nil, episodeNotFoundError(ctx, input.EpisodeID)
	}

	return uc.unarchiveEpisode(ctx, target.Episode)
}

// unarchiveEpisode persists the re-publish across episodes and, when the updated row is mapped,
// its anime in a single transaction. The episodes write comes first because it owns the
// definitive state and parent guards. Its result carries the current anime mapping; only that
// anime's status is then updated, leaving all content attributes untouched.
//
// [Ja] unarchiveEpisode は再公開を episodes に、そして実際に更新した行がマッピング済みならその
// anime にも 1 トランザクションで永続化する。episodes への書き込みを先に行うのは、状態と親作品の
// 最終的なガードをそこが持つため。結果が現在の anime の写像を運び、その anime の status だけを
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
			return nil, fmt.Errorf("anime の状態更新に失敗しました: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return &UnarchiveEpisodeOutput{EpisodeID: current.ID, WorkID: current.WorkID}, nil
}
