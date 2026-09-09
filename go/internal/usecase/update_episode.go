package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/validator"
)

const (
	defaultUpdateEpisodeLockRetryLimit     = 5
	defaultUpdateEpisodeLockRetryBaseDelay = 10 * time.Millisecond
)

// UpdateEpisodeUsecase applies one submit of the Annict DB episode edit form. Like the bulk
// create it is anchored on animes: the mapped anime and its kind='episode' classification are
// dual-written alongside the episodes row in the same transaction, and episodes stays the
// source of truth during the migration.
//
// [Ja] UpdateEpisodeUsecase は Annict DB のエピソード編集フォームの 1 回の送信を適用する。
// 一括作成と同じく animes を基点とし、マッピング済みの anime とその kind='episode' の分類を
// episodes の行と同一トランザクションで両書きする。移行期間中の正本は episodes 側。
type UpdateEpisodeUsecase struct {
	db                      *sql.DB
	episodeRepo             *repository.EpisodeRepository
	animeRepo               *repository.AnimeRepository
	animeClassificationRepo *repository.AnimeClassificationRepository
	validator               *validator.DBEpisodeUpdateValidator
	// lockRetryLimit and lockRetryBaseDelay bound the whole-transaction retry a NOWAIT lock
	// miss triggers. They are fields rather than constants so a test can shorten the wait
	// without every case that reaches this use case paying for the real backoff.
	//
	// [Ja] lockRetryLimit と lockRetryBaseDelay は、NOWAIT のロック取得失敗が引き起こす
	// トランザクション全体の再試行を制限する。定数ではなくフィールドにしているのは、この
	// UseCase に到達する全ケースが実際の backoff を払わずに、テストから待ち時間を縮められる
	// ようにするため。
	lockRetryLimit     int
	lockRetryBaseDelay time.Duration
}

// NewUpdateEpisodeUsecase constructs an UpdateEpisodeUsecase.
//
// [Ja] NewUpdateEpisodeUsecase は UpdateEpisodeUsecase を生成する。
func NewUpdateEpisodeUsecase(
	db *sql.DB,
	episodeRepo *repository.EpisodeRepository,
	animeRepo *repository.AnimeRepository,
	animeClassificationRepo *repository.AnimeClassificationRepository,
	validator *validator.DBEpisodeUpdateValidator,
) *UpdateEpisodeUsecase {
	return &UpdateEpisodeUsecase{
		db:                      db,
		episodeRepo:             episodeRepo,
		animeRepo:               animeRepo,
		animeClassificationRepo: animeClassificationRepo,
		validator:               validator,
		lockRetryLimit:          defaultUpdateEpisodeLockRetryLimit,
		lockRetryBaseDelay:      defaultUpdateEpisodeLockRetryBaseDelay,
	}
}

// UpdateEpisodeInput carries the submitted edit form. The values are strings because that is
// what the form sends, and UpdatedAt is the version it was opened against.
//
// [Ja] UpdateEpisodeInput は送信された編集フォームを保持する。値が文字列なのはフォームが送信
// するものが文字列であるため。UpdatedAt はフォームを開いた時点の版。
type UpdateEpisodeInput struct {
	EpisodeID model.EpisodeID
	// User is the editor the change is attributed to, and the source of the role that decides
	// whether the submit is accepted at all. The user travels whole rather than as a
	// pre-resolved flag so every entry point reaching this use case gets the same reading of
	// the role.
	//
	// [Ja] User は変更の帰属先となる編集者であり、送信を受け付けるかを決めるロールの source
	// でもある。解決済みのフラグではなくユーザーのまま運ぶことで、この UseCase に到達するどの
	// 経路でもロールの解釈が揃う。
	User       *model.User
	Number     string
	RawNumber  string
	SortNumber string
	Title      string
	TitleEn    string
	UpdatedAt  string
}

// UpdateEpisodeOutput reports the updated episode and the work it belongs to, which the caller
// redirects to.
//
// [Ja] UpdateEpisodeOutput は更新されたエピソードと、その所属作品 (呼び出し元がリダイレクト
// 先にする) を報告する。
type UpdateEpisodeOutput struct {
	EpisodeID model.EpisodeID
	WorkID    model.WorkID
}

// Execute applies the submitted values to the episode.
//
// Authorization runs first, then validation, and the episode is loaded last: the load only
// supplies what the dual-write needs beyond the submitted values, so a rejected submit does not
// pay for it.
//
// [Ja] Execute は送信された値をエピソードに適用する。
//
// 最初に認可、次にバリデーションを行い、エピソードの読み込みは最後に行う。読み込みは送信された
// 値に加えて両書きが必要とするものだけを供給するため、却下される送信でその費用を払わない。
func (uc *UpdateEpisodeUsecase) Execute(ctx context.Context, input UpdateEpisodeInput) (*UpdateEpisodeOutput, error) {
	// The submit is restricted to committers and attributed to its author, so a call carrying
	// no user or an ordinary user is refused before anything is read. The web route is gated
	// by the committer middleware and never gets here, but the check keeps the rule with the
	// use case rather than with one caller.
	//
	// [Ja] 送信は committer に限られ、変更者に帰属するため、ユーザーを伴わない呼び出しと
	// 一般ユーザーの呼び出しは何かを読む前に拒否する。web のルートは committer の
	// ミドルウェアで塞がれていてここには来ないが、この確認によりルールが呼び出し元の 1 つでは
	// なく UseCase 側に残る。
	if input.User == nil || !input.User.IsCommitter() {
		return nil, &model.AppError{
			Code:     model.AppErrCodeForbidden,
			UserMsg:  i18n.T(ctx, "error_forbidden"),
			Metadata: map[string]string{"episode_id": input.EpisodeID.String()},
		}
	}

	fields, err := uc.validator.Validate(ctx, validator.DBEpisodeUpdateValidatorInput{
		Number:     input.Number,
		RawNumber:  input.RawNumber,
		SortNumber: input.SortNumber,
		Title:      input.Title,
		TitleEn:    input.TitleEn,
		UpdatedAt:  input.UpdatedAt,
	})
	if err != nil {
		return nil, err
	}

	current, err := uc.episodeRepo.GetForUpdateByID(ctx, input.EpisodeID)
	if err != nil {
		return nil, fmt.Errorf("エピソードの取得に失敗しました: %w", err)
	}
	// An empty result means the episode cannot be edited: it never existed, it was deleted, or
	// its work was, between the edit GET and this submit.
	//
	// [Ja] 結果が空の場合、そのエピソードは編集できない。存在しなかったか、編集の GET と本送信の
	// 間にエピソードまたはその作品が削除された。
	if current == nil {
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_episode_not_found"),
			Metadata: map[string]string{"episode_id": input.EpisodeID.String()},
		}
	}

	// Load the mapped anime only when both sides of the episode mapping are complete, so the
	// dual-write can carry over the columns animes does not source from episodes
	// (archive_message etc.), mirroring UpdateWorkUsecase. A nil parent mapping, a nil episode
	// mapping, or an episode anime_id that points at a missing anime all skip the anime
	// dual-write and leave the phase 2 sync to catch up after the parent is mapped.
	//
	// [Ja] エピソード側と親側の写像がともに完成している場合だけ、両書きが animes 由来でない
	// カラム (archive_message など) を引き継げるよう、マッピング済みの anime を読み込む
	// (UpdateWorkUsecase と対称)。親の写像が nil、エピソードの写像が nil、または episode の
	// anime_id が存在しない anime を指す場合はいずれも anime への両書きをスキップし、親の
	// マッピング後にフェーズ 2 の同期が追いつくようにする。
	var existingAnime *model.Anime
	if current.AnimeID != nil && current.ParentAnimeID != nil {
		existingAnime, err = uc.animeRepo.GetByID(ctx, *current.AnimeID)
		if err != nil {
			return nil, fmt.Errorf("anime の取得に失敗しました: %w", err)
		}
	}

	params := repository.UpdateEpisodeParams{
		ID:         current.ID,
		WorkID:     current.WorkID,
		Number:     fields.Number,
		RawNumber:  fields.RawNumber,
		Title:      fields.Title,
		TitleEn:    fields.TitleEn,
		SortNumber: fields.SortNumber,
		Version:    fields.UpdatedAt,
		UserID:     input.User.ID,
	}

	return uc.updateEpisode(ctx, params, current, existingAnime)
}

// updateEpisode persists the submit across episodes and, when the episode is already mapped,
// its anime / anime_classification in a single transaction. episodes stays the source of truth
// during the migration, so the anime writes are kept in one block that the cutover can remove
// wholesale.
//
// The episodes write comes first because it carries the version match: a submit made against a
// stale read stops there and the anime is never touched.
//
// [Ja] updateEpisode は送信を episodes に、そしてエピソードが既にマッピング済みならその anime /
// anime_classification にも 1 トランザクションで永続化する。移行期間中は episodes が正本のため、
// anime への書き込みは正本切り替えでまるごと外せるよう 1 ブロックにまとめてある。
//
// episodes への書き込みを先に行うのは、そこに版の照合が乗っているため。古い読み取りに対する
// 送信はそこで止まり、anime には触れない。
func (uc *UpdateEpisodeUsecase) updateEpisode(
	ctx context.Context,
	params repository.UpdateEpisodeParams,
	current *model.Episode,
	existingAnime *model.Anime,
) (*UpdateEpisodeOutput, error) {
	// Build the anime / classification mapping before opening the transaction, mirroring
	// createEpisodes and keeping the transaction body to persistence only. Projecting the
	// submitted values onto a *model.Episode (preserving the columns the form does not touch)
	// and reusing the phase 2 sync mapping helpers keeps the episode -> anime / classification
	// mapping single-sourced, so a sync run right after this update reports Unchanged.
	//
	// [Ja] anime / 分類の写像はトランザクションを開く前に組み立てる (createEpisodes と対称、
	// トランザクション内は永続化のみとする)。送信された値を *model.Episode に射影し
	// (フォームが触れないカラムは保持)、フェーズ 2 同期の写像ヘルパーを再利用して
	// episode -> anime / 分類 の写像の正本を 1 つに保つ。これにより更新直後の同期は Unchanged を
	// 報告する。
	var animeParams repository.UpdateAnimeParams
	var classificationParams repository.CreateAnimeClassificationParams
	if existingAnime != nil {
		episode := episodeFromUpdateEpisodeParams(params, current)
		animeParams = animeUpdateParamsFromEpisode(episode, existingAnime)
		classificationParams = classificationCreateParamsFromEpisode(episode, existingAnime.ID)
	}

	output, err := uc.retryEpisodeUpdateLock(ctx, func() (*UpdateEpisodeOutput, error) {
		return uc.updateEpisodeAttempt(ctx, params, current, existingAnime, animeParams, classificationParams)
	})
	if !errors.Is(err, repository.ErrEpisodeLockUnavailable) {
		return output, err
	}

	// Every attempt found a row it needed already locked. Nothing was written and the version
	// the submit was made against still matches, so this is not the version conflict a
	// concurrent edit produces: the same submit succeeds once the other writer commits. Saying
	// so keeps the editor from looking for a difference against stored values that are still
	// the ones their form was opened on.
	//
	// [Ja] どの試行も、必要な行が既にロックされているのを見つけた。何も書かれておらず、送信が
	// 前提とした版も一致したままであるため、これは並行編集が生む版の競合ではない (相手が
	// commit すれば同じ送信で成功する)。そう伝えることで、編集者が、フォームを開いたときのまま
	// である保存済みの値との差分を探しに行かずに済む。
	return nil, &model.AppError{
		Code:     model.AppErrCodeBusy,
		UserMsg:  i18n.T(ctx, "validation_record_busy"),
		Internal: err,
		Metadata: map[string]string{"episode_id": params.ID.String()},
	}
}

// updateEpisodeAttempt persists one complete transaction attempt. A NOWAIT lock miss aborts
// the transaction in PostgreSQL, so retries must begin here rather than around only the failed
// repository call.
//
// [Ja] updateEpisodeAttempt は 1 回分の完全なトランザクションを永続化する。NOWAIT のロック取得
// 失敗は PostgreSQL 上でトランザクションを中断するため、失敗した Repository 呼び出しだけでなく
// ここから再試行する必要がある。
func (uc *UpdateEpisodeUsecase) updateEpisodeAttempt(
	ctx context.Context,
	params repository.UpdateEpisodeParams,
	current *model.Episode,
	existingAnime *model.Anime,
	animeParams repository.UpdateAnimeParams,
	classificationParams repository.CreateAnimeClassificationParams,
) (*UpdateEpisodeOutput, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	updated, err := uc.episodeRepo.WithTx(tx).Update(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("エピソードの更新に失敗しました: %w", err)
	}
	if !updated {
		return nil, &model.AppError{
			Code:     model.AppErrCodeConflict,
			UserMsg:  i18n.T(ctx, "validation_version_conflict"),
			Metadata: map[string]string{"episode_id": params.ID.String()},
		}
	}

	if existingAnime != nil {
		if err := uc.animeRepo.WithTx(tx).Update(ctx, animeParams); err != nil {
			return nil, fmt.Errorf("anime の更新に失敗しました: %w", err)
		}
		if err := uc.animeClassificationRepo.WithTx(tx).Upsert(ctx, classificationParams); err != nil {
			return nil, fmt.Errorf("anime_classification の保存に失敗しました: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return &UpdateEpisodeOutput{EpisodeID: params.ID, WorkID: current.WorkID}, nil
}

// retryEpisodeUpdateLock reruns a complete transaction attempt only for the NOWAIT lock miss
// that breaks the cross-application lock cycle. Other errors keep their original single-attempt
// behaviour. Once the attempts are used up the last lock miss is returned, and the caller turns
// it into the response the editor sees.
//
// [Ja] retryEpisodeUpdateLock は、アプリ間のロック循環を断つ NOWAIT のロック取得失敗の場合だけ、
// 完全なトランザクション試行をやり直す。それ以外のエラーは従来どおり 1 回で返す。試行を使い切った
// 場合は最後のロック取得失敗を返し、呼び出し側がそれを編集者に見せる応答へ変換する。
func (uc *UpdateEpisodeUsecase) retryEpisodeUpdateLock(ctx context.Context, attempt func() (*UpdateEpisodeOutput, error)) (*UpdateEpisodeOutput, error) {
	var lastErr error
	for i := 0; i < uc.lockRetryLimit; i++ {
		output, err := attempt()
		if !errors.Is(err, repository.ErrEpisodeLockUnavailable) {
			return output, err
		}
		lastErr = err
		if i == uc.lockRetryLimit-1 {
			break
		}
		if err := uc.waitForEpisodeUpdateRetry(ctx, i); err != nil {
			return nil, err
		}
	}
	return nil, lastErr
}

// waitForEpisodeUpdateRetry waits without ignoring cancellation between whole-transaction
// retries.
//
// [Ja] waitForEpisodeUpdateRetry は、トランザクション全体の再試行間でキャンセルを無視せず待つ。
func (uc *UpdateEpisodeUsecase) waitForEpisodeUpdateRetry(ctx context.Context, attempt int) error {
	delay := uc.lockRetryBaseDelay * time.Duration(1<<attempt)
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// episodeFromUpdateEpisodeParams projects an UpdateEpisodeParams onto a *model.Episode, adding
// the anime-mapped columns the edit form does not submit from the stored row: title_ro, the
// parent work's anime, and the episode-state source (unpublished_at / deleted_at), from which
// animeUpdateParamsFromEpisode derives anime.status. Carrying the state timestamps over is what
// keeps a content edit from clobbering an archived anime back to published. The update never
// changes those columns, so the updated anime mirrors the post-update episodes row and the sync
// right after the update reports Unchanged.
//
// [Ja] episodeFromUpdateEpisodeParams は UpdateEpisodeParams を *model.Episode に射影し、編集
// フォームが送信しない anime 写像カラムを保存済みの行から足す: title_ro、親作品の anime、および
// animeUpdateParamsFromEpisode が anime.status を導出するエピソード状態の source (unpublished_at
// / deleted_at)。状態タイムスタンプを引き継ぐことが、内容編集でアーカイブ済みの anime を
// published に戻してしまうのを防ぐ。更新はこれらのカラムを変えないため、更新後の anime が更新後の
// episodes 行を写し、更新直後の同期は Unchanged を報告する。
func episodeFromUpdateEpisodeParams(params repository.UpdateEpisodeParams, current *model.Episode) *model.Episode {
	return &model.Episode{
		ID:            current.ID,
		WorkID:        current.WorkID,
		Title:         params.Title,
		TitleRo:       current.TitleRo,
		TitleEn:       params.TitleEn,
		Number:        params.Number,
		RawNumber:     params.RawNumber,
		SortNumber:    params.SortNumber,
		AnimeID:       current.AnimeID,
		ParentAnimeID: current.ParentAnimeID,
		UnpublishedAt: current.UnpublishedAt,
		DeletedAt:     current.DeletedAt,
	}
}
