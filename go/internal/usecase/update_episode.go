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

// UpdateEpisodeUsecaseはAnnict DBのエピソード編集フォームの1回の送信を適用する。
// 一括作成と同じくanimesを基点とし、マッピング済みのanimeとそのkind='episode' の分類を
// episodesの行と同一トランザクションで両書きする。移行期間中の正本はepisodes側。
type UpdateEpisodeUsecase struct {
	db                      *sql.DB
	episodeRepo             *repository.EpisodeRepository
	animeRepo               *repository.AnimeRepository
	animeClassificationRepo *repository.AnimeClassificationRepository
	validator               *validator.DBEpisodeUpdateValidator
	// lockRetryLimitとlockRetryBaseDelayは、NOWAITのロック取得失敗が引き起こす
	// トランザクション全体の再試行を制限する。定数ではなくフィールドにしているのは、この
	// UseCaseに到達する全ケースが実際のbackoffを払わずに、テストから待ち時間を縮められる
	// ようにするため。
	lockRetryLimit     int
	lockRetryBaseDelay time.Duration
}

// NewUpdateEpisodeUsecaseはUpdateEpisodeUsecaseを生成する。
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

// UpdateEpisodeInputは送信された編集フォームを保持する。値が文字列なのはフォームが送信
// するものが文字列であるため。UpdatedAtはフォームを開いた時点の版。
type UpdateEpisodeInput struct {
	EpisodeID model.EpisodeID
	// Userは変更の帰属先となる編集者であり、送信を受け付けるかを決めるロールのsource
	// でもある。解決済みのフラグではなくユーザーのまま運ぶことで、このUseCaseに到達するどの
	// 経路でもロールの解釈が揃う。
	User       *model.User
	Number     string
	RawNumber  string
	SortNumber string
	Title      string
	TitleEn    string
	UpdatedAt  string
}

// UpdateEpisodeOutputは更新されたエピソードと、その所属作品 (呼び出し元がリダイレクト
// 先にする) を報告する。
type UpdateEpisodeOutput struct {
	EpisodeID model.EpisodeID
	WorkID    model.WorkID
}

// Executeは送信された値をエピソードに適用する。
//
// 最初に認可、次にバリデーションを行い、エピソードの読み込みは最後に行う。読み込みは送信された
// 値に加えて両書きが必要とするものだけを供給するため、却下される送信でその費用を払わない。
func (uc *UpdateEpisodeUsecase) Execute(ctx context.Context, input UpdateEpisodeInput) (*UpdateEpisodeOutput, error) {
	// 送信はcommitterに限られ、変更者に帰属するため、ユーザーを伴わない呼び出しと
	// 一般ユーザーの呼び出しは何かを読む前に拒否する。webのルートはcommitterの
	// ミドルウェアで塞がれていてここには来ないが、この確認によりルールが呼び出し元の1つでは
	// なくUseCase側に残る。
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
	// 結果が空の場合、そのエピソードは編集できない。存在しなかったか、編集のGETと本送信の
	// 間にエピソードまたはその作品が削除された。
	if current == nil {
		return nil, &model.AppError{
			Code:     model.AppErrCodeResourceNotFound,
			UserMsg:  i18n.T(ctx, "error_episode_not_found"),
			Metadata: map[string]string{"episode_id": input.EpisodeID.String()},
		}
	}

	// エピソード側と親側の写像がともに完成している場合だけ、両書きがanimes由来でない
	// カラム (archive_messageなど) を引き継げるよう、マッピング済みのanimeを読み込む
	// (UpdateWorkUsecaseと対称)。親の写像がnil、エピソードの写像がnil、またはepisodeの
	// anime_idが存在しないanimeを指す場合はいずれもanimeへの両書きをスキップし、親の
	// マッピング後にフェーズ2の同期が追いつくようにする。
	var existingAnime *model.Anime
	if current.AnimeID != nil && current.ParentAnimeID != nil {
		existingAnime, err = uc.animeRepo.GetByID(ctx, *current.AnimeID)
		if err != nil {
			return nil, fmt.Errorf("animeの取得に失敗しました: %w", err)
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

// updateEpisodeは送信をepisodesに、そしてエピソードが既にマッピング済みならそのanime /
// anime_classificationにも1トランザクションで永続化する。移行期間中はepisodesが正本のため、
// animeへの書き込みは正本切り替えでまるごと外せるよう1ブロックにまとめてある。
//
// episodesへの書き込みを先に行うのは、そこに版の照合が乗っているため。古い読み取りに対する
// 送信はそこで止まり、animeには触れない。
func (uc *UpdateEpisodeUsecase) updateEpisode(
	ctx context.Context,
	params repository.UpdateEpisodeParams,
	current *model.Episode,
	existingAnime *model.Anime,
) (*UpdateEpisodeOutput, error) {
	// anime / 分類の写像はトランザクションを開く前に組み立てる (createEpisodesと対称、
	// トランザクション内は永続化のみとする)。送信された値を *model.Episodeに射影し
	// (フォームが触れないカラムは保持)、フェーズ2同期の写像ヘルパーを再利用して
	// episode -> anime / 分類 の写像の正本を1つに保つ。これにより更新直後の同期はUnchangedを
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

	// どの試行も、必要な行が既にロックされているのを見つけた。何も書かれておらず、送信が
	// 前提とした版も一致したままであるため、これは並行編集が生む版の競合ではない (相手が
	// commitすれば同じ送信で成功する)。そう伝えることで、編集者が、フォームを開いたときのまま
	// である保存済みの値との差分を探しに行かずに済む。
	return nil, &model.AppError{
		Code:     model.AppErrCodeBusy,
		UserMsg:  i18n.T(ctx, "validation_record_busy"),
		Internal: err,
		Metadata: map[string]string{"episode_id": params.ID.String()},
	}
}

// updateEpisodeAttemptは1回分の完全なトランザクションを永続化する。NOWAITのロック取得
// 失敗はPostgreSQL上でトランザクションを中断するため、失敗したRepository呼び出しだけでなく
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
			return nil, fmt.Errorf("animeの更新に失敗しました: %w", err)
		}
		if err := uc.animeClassificationRepo.WithTx(tx).Upsert(ctx, classificationParams); err != nil {
			return nil, fmt.Errorf("anime_classificationの保存に失敗しました: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return &UpdateEpisodeOutput{EpisodeID: params.ID, WorkID: current.WorkID}, nil
}

// retryEpisodeUpdateLockは、アプリ間のロック循環を断つNOWAITのロック取得失敗の場合だけ、
// 完全なトランザクション試行をやり直す。それ以外のエラーは従来どおり1回で返す。試行を使い切った
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

// waitForEpisodeUpdateRetryは、トランザクション全体の再試行間でキャンセルを無視せず待つ。
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

// episodeFromUpdateEpisodeParamsはUpdateEpisodeParamsを *model.Episodeに射影し、編集
// フォームが送信しないanime写像カラムを保存済みの行から足す: title_ro、親作品のanime、および
// animeUpdateParamsFromEpisodeがanime.statusを導出するエピソード状態のsource (unpublished_at
// / deleted_at)。状態タイムスタンプを引き継ぐことが、内容編集でアーカイブ済みのanimeを
// publishedに戻してしまうのを防ぐ。更新はこれらのカラムを変えないため、更新後のanimeが更新後の
// episodes行を写し、更新直後の同期はUnchangedを報告する。
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
