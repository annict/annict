package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
)

// DeleteEpisodeUsecase soft-deletes an episode from the Annict DB admin screen. It sets
// episodes.deleted_at, the episode-state source of truth for deletion, and dual-writes only the
// derived anime.status = deleted in the same transaction. The delete is a soft delete only
// (ADR 0004: animes has no physical delete), unlike the Rails destroy_in_batches, so no child
// resources are cascaded: the deleted state governs visibility. Because the status is derived
// from the timestamp this usecase sets, a phase 2 reconciliation right after reports Unchanged
// instead of clobbering the anime back to published.
//
// The usecase enforces admin authorization itself so every entry point gets the same rule, in
// addition to the HTTP route's RequireAdmin middleware. Deleting is admin-only while archiving
// is open to committers (ADR 0009), which is also the split the Rails EpisodePolicy makes. The
// Rails delete records no db_activity, so the administrator is used for authorization but no
// activity is attributed to them.
//
// The episode update is conditional on the row not being deleted already, still belonging to the
// work observed by the pre-transaction projection, and that work still being undeleted. A result
// is returned only when the parent touch and counter update also succeeds, so a concurrent parent
// deletion makes the transaction roll back. It returns the anime_id from the row it actually
// deleted, so a concurrent mapping change cannot send the status write to the former anime.
// Updating only status also preserves anime content committed after the pre-read.
//
// [Ja] DeleteEpisodeUsecase は Annict DB 管理画面からエピソードをソフトデリートする。削除に
// ついてのエピソード状態の正本である episodes.deleted_at を立て、同一トランザクションで導出した
// anime.status = deleted だけを両書きする。Rails の destroy_in_batches と違い削除はソフト
// デリートのみで (ADR 0004: animes は物理削除を持たない)、子リソースへのカスケードは行わない
// (削除状態が可視性を支配する)。status は本 UseCase が立てた timestamp から導出されるため、直後
// にフェーズ 2 のリコンシリエーションが走っても Unchanged となり、anime が published に差し戻さ
// れる (クロッバー) ことはない。
//
// HTTP ルートの RequireAdmin middleware に加えて、すべての entry point で同じ規則になるよう
// UseCase 自身も admin 認可を強制する。非公開が committer に開かれているのに対し削除が admin
// 専用なのは ADR 0009 の権限分離で、Rails の EpisodePolicy が行う分割とも同じ。Rails の削除は
// db_activity を記録しないため、管理者は認可に使うが活動履歴の作成者としては記録しない。
//
// episodes の更新は、行がまだ削除されておらず、トランザクション前の射影が観測した作品に今も属し、
// その作品も未削除の場合だけ行う。親作品の touch とカウンター更新にも成功した場合だけ結果を返し、
// 親作品が同時に削除された場合はトランザクションをロールバックする。実際に削除した行から
// anime_id を返すため、写像が同時に変わっても以前の anime へ status を書かない。status だけを
// 更新することで、事前読み取り後にコミットされた anime の内容も保持する。
type DeleteEpisodeUsecase struct {
	db          *sql.DB
	episodeRepo *repository.EpisodeRepository
	animeRepo   *repository.AnimeRepository
}

// NewDeleteEpisodeUsecase constructs a DeleteEpisodeUsecase.
//
// [Ja] NewDeleteEpisodeUsecase は DeleteEpisodeUsecase を生成する。
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

// DeleteEpisodeInput identifies the episode to delete and the user authorizing the write.
//
// [Ja] DeleteEpisodeInput は削除するエピソードと、書き込みを認可するユーザーを指定する。
type DeleteEpisodeInput struct {
	EpisodeID model.EpisodeID
	User      *model.User
}

// DeleteEpisodeOutput reports the deleted episode and the work it belonged to, which the caller
// redirects to.
//
// [Ja] DeleteEpisodeOutput は削除したエピソードと、その所属作品 (呼び出し元がリダイレクト先に
// する) を報告する。
type DeleteEpisodeOutput struct {
	EpisodeID model.EpisodeID
	WorkID    model.WorkID
}

// Execute deletes the episode the submit named.
//
// Authorization runs before any read so a direct caller cannot use resource existence as an
// oracle without the same admin permission the HTTP route requires.
//
// [Ja] Execute は送信が名指ししたエピソードを削除する。
//
// 認可は読み取りより先に行う。直接の呼び出し元も HTTP ルートと同じ admin 権限を持たなければ、
// リソースの存在を判別できないようにするため。
func (uc *DeleteEpisodeUsecase) Execute(ctx context.Context, input DeleteEpisodeInput) (*DeleteEpisodeOutput, error) {
	if input.User == nil || !input.User.IsAdmin() {
		return nil, &model.AppError{
			Code:     model.AppErrCodeForbidden,
			UserMsg:  i18n.T(ctx, "error_forbidden"),
			Metadata: map[string]string{"episode_id": input.EpisodeID.String()},
		}
	}

	// The delete projection supplies the parent this submit is taken to name. An empty result
	// means the list the administrator acted on is stale: Rails scopes the delete to
	// Episode.without_deleted and raises RecordNotFound otherwise. Delete repeats the state,
	// parent identity, and parent lifecycle conditions in SQL because this read is outside its
	// transaction.
	//
	// [Ja] 削除用の射影から、この送信が名指しするとみなす親作品を得る。結果が空の場合、管理者が
	// 操作した一覧は古い (Rails も削除を Episode.without_deleted に絞り、外れれば
	// RecordNotFound を送出する)。この読み取りはトランザクション外のため、Delete は状態、親作品
	// の同一性、親作品のライフサイクル条件を SQL でも繰り返す。
	target, err := uc.episodeRepo.GetForDeleteByID(ctx, input.EpisodeID)
	if err != nil {
		return nil, fmt.Errorf("エピソードの取得に失敗しました: %w", err)
	}
	if target == nil {
		return nil, episodeNotFoundError(ctx, input.EpisodeID)
	}

	return uc.deleteEpisode(ctx, target)
}

// deleteEpisode persists the soft delete across episodes and, when the updated row is mapped, its
// anime in a single transaction. The episodes write comes first because it owns the definitive
// state and parent guards. Its result carries the current anime mapping; only that anime's status
// is then updated, leaving all content attributes untouched.
//
// [Ja] deleteEpisode はソフトデリートを episodes に、そして実際に更新した行がマッピング済みなら
// その anime にも 1 トランザクションで永続化する。episodes への書き込みを先に行うのは、状態と
// 親作品の最終的なガードをそこが持つため。結果が現在の anime の写像を運び、その anime の status
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
			return nil, fmt.Errorf("anime の状態更新に失敗しました: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return &DeleteEpisodeOutput{EpisodeID: current.ID, WorkID: current.WorkID}, nil
}
