package usecase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/validator"
)

// episodeSortNumberStep is the gap the bulk create leaves between the sort_numbers it
// assigns. The spacing is what lets an editor slot a later addition (a recap aired between
// two episodes, say) between two existing rows without renumbering the work.
//
// [Ja] episodeSortNumberStep は一括作成が振る sort_number の間隔。この間隔があることで、
// 後から追加されたエピソード (2 話の間に挟まる総集編など) を、作品全体を振り直さずに既存の
// 2 行の間へ入れられる。
const episodeSortNumberStep int32 = 100

// CreateEpisodesUsecase creates the episodes of one bulk-create submit. One submit carries
// many rows and creates them together: either every row of the submit is stored or none is.
//
// Writes are anchored on animes, as the work create is: each row becomes an anime plus its
// kind='episode' classification, and the episodes row is dual-written in the same
// transaction. episodes stays the source of truth during the migration.
//
// [Ja] CreateEpisodesUsecase は一括作成の 1 回の送信で作られるエピソードを作成する。1 回の
// 送信は複数行を運び、それらをまとめて作成する (送信された全行が保存されるか、1 行も保存され
// ないかのどちらか)。
//
// 書き込みは作品作成と同じく animes を基点とする。各行は anime とその kind='episode' の分類に
// なり、episodes の行は同一トランザクションで両書きする。移行期間中の正本は episodes 側。
type CreateEpisodesUsecase struct {
	db                      *sql.DB
	workRepo                *repository.WorkRepository
	episodeRepo             *repository.EpisodeRepository
	animeRepo               *repository.AnimeRepository
	animeClassificationRepo *repository.AnimeClassificationRepository
	validator               *validator.DBEpisodeCreateValidator
}

// NewCreateEpisodesUsecase constructs a CreateEpisodesUsecase.
//
// [Ja] NewCreateEpisodesUsecase は CreateEpisodesUsecase を生成する。
func NewCreateEpisodesUsecase(
	db *sql.DB,
	workRepo *repository.WorkRepository,
	episodeRepo *repository.EpisodeRepository,
	animeRepo *repository.AnimeRepository,
	animeClassificationRepo *repository.AnimeClassificationRepository,
	validator *validator.DBEpisodeCreateValidator,
) *CreateEpisodesUsecase {
	return &CreateEpisodesUsecase{
		db:                      db,
		workRepo:                workRepo,
		episodeRepo:             episodeRepo,
		animeRepo:               animeRepo,
		animeClassificationRepo: animeClassificationRepo,
		validator:               validator,
	}
}

// CreateEpisodesInput carries the submitted form: the work the episodes belong to and the
// raw lines, one episode per line, as typed into the textarea.
//
// [Ja] CreateEpisodesInput は送信されたフォームを保持する。エピソードが属する作品と、
// textarea に入力されたままの行 (1 行 1 エピソード)。
type CreateEpisodesInput struct {
	WorkID model.WorkID
	// User is the editor the submit is attributed to, and the source of the role that decides
	// whether the manual-creation restriction applies. The user travels whole rather than as a
	// pre-resolved flag so every entry point reaching this use case gets the same reading of
	// the role.
	//
	// [Ja] User は送信の帰属先となる編集者であり、手動作成制限を適用するかを決めるロールの
	// source でもある。解決済みのフラグではなくユーザーのまま運ぶことで、この UseCase に
	// 到達するどの経路でもロールの解釈が揃う。
	User *model.User
	Rows string
}

// CreateEpisodesOutput reports the episodes created by the submit, in input order.
//
// [Ja] CreateEpisodesOutput は送信によって作成されたエピソードを入力順で報告する。
type CreateEpisodesOutput struct {
	EpisodeIDs []model.EpisodeID
}

// Execute creates every submitted row under the given work.
//
// Authorization is checked first. The work is then loaded before validation, so a submit
// against a work that does not exist (or was deleted) is reported as such rather than as a
// failed validation, matching the Rails create action that finds the work before it touches
// the form.
//
// [Ja] Execute は指定作品の配下に、送信された全行を作成する。
//
// 最初に認可を確認する。その後は作品をバリデーションより先に読み込むため、存在しない
// (あるいは削除済みの) 作品への送信はバリデーション失敗ではなくそのものとして報告される。
// フォームに触れる前に作品を引く Rails の create アクションと同じ順序。
func (uc *CreateEpisodesUsecase) Execute(ctx context.Context, input CreateEpisodesInput) (*CreateEpisodesOutput, error) {
	// The submit is restricted to committers and attributed to its author, so a call carrying
	// no user or an ordinary user is refused before anything is read. The web route is gated
	// by the committer middleware and never gets here, but the check keeps the rule with the
	// use case rather than with one caller.
	//
	// [Ja] 送信は committer に限られ、作成者に帰属するため、ユーザーを伴わない呼び出しと
	// 一般ユーザーの呼び出しは何かを読む前に拒否する。web のルートは committer の
	// ミドルウェアで塞がれていてここには来ないが、この確認によりルールが呼び出し元の 1 つでは
	// なく UseCase 側に残る。
	if input.User == nil || !input.User.IsCommitter() {
		return nil, &model.AppError{
			Code:     model.AppErrCodeForbidden,
			UserMsg:  i18n.T(ctx, "error_forbidden"),
			Metadata: map[string]string{"work_id": input.WorkID.String()},
		}
	}

	exists, err := uc.workRepo.ExistsForEpisodeCreateByID(ctx, input.WorkID)
	if err != nil {
		return nil, fmt.Errorf("作品の存在確認に失敗: %w", err)
	}
	if !exists {
		return nil, workNotFoundError(ctx, input.WorkID)
	}

	rows, err := uc.validator.Validate(ctx, validator.DBEpisodeCreateValidatorInput{Rows: input.Rows})
	if err != nil {
		return nil, err
	}

	return uc.createEpisodes(ctx, input, rows)
}

func workNotFoundError(ctx context.Context, workID model.WorkID) *model.AppError {
	return &model.AppError{
		Code:     model.AppErrCodeResourceNotFound,
		UserMsg:  i18n.T(ctx, "error_work_not_found"),
		Metadata: map[string]string{"work_id": workID.String()},
	}
}

// createEpisodes persists the submitted rows in a single transaction. Each row is written as
// an anime, its kind='episode' classification and the episodes row itself, in that order:
// episodes.anime_id is an FK to animes(id), so the anime has to exist before the episode can
// name it. A work that is not mapped to an anime yet gets the episodes rows alone, and the
// phase 2 sync creates their animes once the parent work is synced.
//
// The mapping to anime / classification reuses the phase 2 sync helpers, so a sync run right
// after a create reports Unchanged instead of immediately rewriting what was just created.
//
// [Ja] createEpisodes は送信された行を 1 トランザクションで永続化する。各行は anime、その
// kind='episode' の分類、episodes の行の順に書く。episodes.anime_id は animes(id) への FK の
// ため、エピソードが指す前に anime が存在している必要がある。まだ anime にマッピングされて
// いない作品では episodes の行だけを書き、その anime は親作品が同期された後にフェーズ 2 の
// 同期が作る。
//
// anime / 分類への写像はフェーズ 2 同期のヘルパーを再利用する。これにより、作成直後の同期は
// 作りたてのものを書き直さず Unchanged を報告する。
func (uc *CreateEpisodesUsecase) createEpisodes(
	ctx context.Context,
	input CreateEpisodesInput,
	rows []validator.DBEpisodeRow,
) (*CreateEpisodesOutput, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	workRepo := uc.workRepo.WithTx(tx)
	animeRepo := uc.animeRepo.WithTx(tx)
	classificationRepo := uc.animeClassificationRepo.WithTx(tx)
	episodeRepo := uc.episodeRepo.WithTx(tx)

	locked, err := workRepo.LockForEpisodeCreateByID(ctx, input.WorkID)
	if err != nil {
		return nil, fmt.Errorf("作品のロックに失敗しました: %w", err)
	}
	if !locked {
		return nil, workNotFoundError(ctx, input.WorkID)
	}

	// Read the numbering anchors only after the work lock has been acquired. Under the
	// default READ COMMITTED isolation, a waiter then observes the preceding creator's
	// committed episodes rather than the snapshot it had before waiting.
	//
	// [Ja] 採番の起点は作品ロックを取得した後にだけ読み取る。既定の READ COMMITTED 分離では、
	// 待機した処理が待機前のスナップショットではなく、先行作成者がコミットしたエピソードを
	// 参照できる。
	work, err := workRepo.GetForEpisodeCreateByID(ctx, input.WorkID)
	if err != nil {
		return nil, fmt.Errorf("作品の取得に失敗しました: %w", err)
	}
	if work == nil {
		return nil, workNotFoundError(ctx, input.WorkID)
	}
	if !input.User.IsAdmin() && !work.ManualCreationState.Allowed() {
		return nil, manualCreationValidationError(ctx, work.ManualCreationState)
	}

	// The numbering starts at the work's episode count times the step, as the Rails form
	// does. The count is an int64 only because COUNT returns one; the number of episodes a
	// work can have stays far inside the int32 range the sort_number column holds.
	//
	// [Ja] 採番は Rails のフォームと同じく、作品のエピソード数 × ステップから始める。件数が
	// int64 なのは COUNT がそう返すためで、1 作品が持ちうるエピソード数は sort_number カラム
	// の int32 の範囲に十分収まる。
	sortNumber := int32(work.EpisodeCount) * episodeSortNumberStep // #nosec G115
	previous := work.LatestEpisode
	episodeIDs := make([]model.EpisodeID, 0, len(rows))

	for i, row := range rows {
		sortNumber += episodeSortNumberStep
		episode := episodeFromCreateRow(work.Work, row, sortNumber)

		var animeID *model.AnimeID
		if episode.ParentAnimeID != nil {
			anime, err := animeRepo.Create(ctx, animeCreateParamsFromEpisode(episode))
			if err != nil {
				return nil, fmt.Errorf("anime の作成に失敗しました (%d 行目): %w", i+1, err)
			}
			if _, err := classificationRepo.Create(ctx, classificationCreateParamsFromEpisode(episode, anime.ID)); err != nil {
				return nil, fmt.Errorf("anime_classification の作成に失敗しました (%d 行目): %w", i+1, err)
			}
			animeID = &anime.ID
		}

		params := repository.CreateEpisodeParams{
			WorkID:     episode.WorkID,
			Number:     episode.Number,
			RawNumber:  episode.RawNumber,
			Title:      episode.Title,
			SortNumber: episode.SortNumber,
			AnimeID:    animeID,
			UserID:     input.User.ID,
		}
		if previous != nil {
			params.PrevEpisodeID = &previous.ID
		}

		episodeID, err := episodeRepo.Create(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("エピソードの作成に失敗しました (%d 行目): %w", i+1, err)
		}
		episodeIDs = append(episodeIDs, episodeID)

		previous = nextSortAnchor(previous, episodeID, sortNumber)
	}

	createdCount := int32(len(rows)) // #nosec G115 -- validation caps a submit at 100 rows.
	if err := workRepo.IncrementEpisodesCount(ctx, input.WorkID, createdCount); err != nil {
		return nil, fmt.Errorf("作品のエピソード件数の更新に失敗しました: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return &CreateEpisodesOutput{EpisodeIDs: episodeIDs}, nil
}

// manualCreationValidationError reports a submit the work's state does not allow. The message
// is a global one: nothing about the submitted lines is wrong, so marking the textarea invalid
// would tell the editor to correct input that is fine.
//
// [Ja] manualCreationValidationError は、作品の状態が許さない送信を報告する。メッセージは
// グローバルに積む。送信された行に問題は無く、textarea を不正と印付けると、直す必要のない
// 入力を直すよう促すことになるため。
func manualCreationValidationError(ctx context.Context, state model.ManualEpisodeCreationState) *model.ValidationError {
	key := "validation_episode_manual_creation_slots_exist"
	if state.Restriction() == model.ManualEpisodeCreationEpisodesFilled {
		key = "validation_episode_manual_creation_episodes_filled"
	}

	ve := model.NewValidationError()
	ve.AddGlobal(i18n.T(ctx, key))
	return ve
}

// episodeFromCreateRow projects one submitted row onto the *model.Episode fields the animes /
// anime_classifications mapping reads, so the create path feeds the same
// animeCreateParamsFromEpisode / classificationCreateParamsFromEpisode helpers the phase 2
// sync uses and the two cannot derive different rows from the same episode.
//
// It mirrors the partial-load pattern of episodeFromAnimeSyncRow: only the mapped fields are
// set and the rest of *model.Episode stays at its zero value. ParentAnimeID carries the
// parent work's anime, which is nil while the work is not mapped. A new episode leaves
// UnpublishedAt / DeletedAt at nil, so DerivedStatus reports published and the anime the
// mapping produces is published too.
//
// [Ja] episodeFromCreateRow は送信された 1 行を、animes / anime_classifications の写像が読む
// *model.Episode のフィールドに射影する。これにより create 経路もフェーズ 2 同期と同じ
// animeCreateParamsFromEpisode / classificationCreateParamsFromEpisode ヘルパーに通せ、両者が
// 同じエピソードから異なる行を導出することがなくなる。
//
// episodeFromAnimeSyncRow の partial-load パターンに倣い、写像対象のフィールドだけをセットして
// 残りの *model.Episode はゼロ値のまま残す。ParentAnimeID は親作品の anime を運び、作品が
// 未マッピングのあいだは nil になる。新規エピソードは UnpublishedAt / DeletedAt を nil のまま
// にするため DerivedStatus は published を報告し、写像が生む anime も published になる。
func episodeFromCreateRow(work *model.Work, row validator.DBEpisodeRow, sortNumber int32) *model.Episode {
	return &model.Episode{
		WorkID:        work.ID,
		Title:         row.Title,
		Number:        row.Number,
		RawNumber:     row.RawNumber,
		SortNumber:    sortNumber,
		ParentAnimeID: work.AnimeID,
	}
}

// nextSortAnchor returns the anchor the next created row names as its preceding episode: the
// episode just created when it holds the greatest sort_number so far, and the previous anchor
// otherwise. Ties go to the new episode, which is the later row of the two.
//
// The comparison matters because the numbering starts from the work's episode count rather
// than from its greatest sort_number: a work whose existing episodes were spaced further
// apart keeps naming that episode as the preceding one, exactly as the Rails callback does.
//
// [Ja] nextSortAnchor は、次に作る行が直前のエピソードとして名指しする起点を返す。作成した
// ばかりのエピソードがそこまでで最大の sort_number を持つならそれを、そうでなければ従来の
// 起点を返す。同値の場合は 2 つのうち後の行である新規エピソードを採る。
//
// 採番の起点が作品の最大 sort_number ではなくエピソード数であるため、この比較が要る。既存の
// エピソードがより広い間隔で並んでいる作品では、そのエピソードが直前のエピソードとして名指し
// され続ける。これは Rails のコールバックの挙動そのもの。
func nextSortAnchor(previous *repository.DBEpisodeSortAnchor, episodeID model.EpisodeID, sortNumber int32) *repository.DBEpisodeSortAnchor {
	if previous != nil && previous.SortNumber > sortNumber {
		return previous
	}

	return &repository.DBEpisodeSortAnchor{ID: episodeID, SortNumber: sortNumber}
}
