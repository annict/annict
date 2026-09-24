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

// episodeSortNumberStepは一括作成が振るsort_numberの間隔。この間隔があることで、
// 後から追加されたエピソード (2話の間に挟まる総集編など) を、作品全体を振り直さずに既存の
// 2行の間へ入れられる。
const episodeSortNumberStep int32 = 100

// CreateEpisodesUsecaseは一括作成の1回の送信で作られるエピソードを作成する。1回の
// 送信は複数行を運び、それらをまとめて作成する (送信された全行が保存されるか、1行も保存され
// ないかのどちらか)。
//
// 書き込みは作品作成と同じくanimesを基点とする。各行はanimeとそのkind='episode' の分類に
// なり、episodesの行は同一トランザクションで両書きする。移行期間中の正本はepisodes側。
type CreateEpisodesUsecase struct {
	db                      *sql.DB
	workRepo                *repository.WorkRepository
	episodeRepo             *repository.EpisodeRepository
	animeRepo               *repository.AnimeRepository
	animeClassificationRepo *repository.AnimeClassificationRepository
	validator               *validator.DBEpisodeCreateValidator
}

// NewCreateEpisodesUsecaseはCreateEpisodesUsecaseを生成する。
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

// CreateEpisodesInputは送信されたフォームを保持する。エピソードが属する作品と、
// textareaに入力されたままの行 (1行1エピソード)。
type CreateEpisodesInput struct {
	WorkID model.WorkID
	// Userは送信の帰属先となる編集者であり、手動作成制限を適用するかを決めるロールの
	// sourceでもある。解決済みのフラグではなくユーザーのまま運ぶことで、このUseCaseに
	// 到達するどの経路でもロールの解釈が揃う。
	User *model.User
	Rows string
}

// CreateEpisodesOutputは送信によって作成されたエピソードを入力順で報告する。
type CreateEpisodesOutput struct {
	EpisodeIDs []model.EpisodeID
}

// Executeは指定作品の配下に、送信された全行を作成する。
//
// 最初に認可を確認する。その後は作品をバリデーションより先に読み込むため、存在しない
// (あるいは削除済みの) 作品への送信はバリデーション失敗ではなくそのものとして報告される。
// フォームに触れる前に作品を引くRailsのcreateアクションと同じ順序。
func (uc *CreateEpisodesUsecase) Execute(ctx context.Context, input CreateEpisodesInput) (*CreateEpisodesOutput, error) {
	// 送信はcommitterに限られ、作成者に帰属するため、ユーザーを伴わない呼び出しと
	// 一般ユーザーの呼び出しは何かを読む前に拒否する。webのルートはcommitterの
	// ミドルウェアで塞がれていてここには来ないが、この確認によりルールが呼び出し元の1つでは
	// なくUseCase側に残る。
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

// createEpisodesは送信された行を1トランザクションで永続化する。各行はanime、その
// kind='episode' の分類、episodesの行の順に書く。episodes.anime_idはanimes(id) へのFKの
// ため、エピソードが指す前にanimeが存在している必要がある。まだanimeにマッピングされて
// いない作品ではepisodesの行だけを書き、そのanimeは親作品が同期された後にフェーズ2の
// 同期が作る。
//
// anime / 分類への写像はフェーズ2同期のヘルパーを再利用する。これにより、作成直後の同期は
// 作りたてのものを書き直さずUnchangedを報告する。
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

	// 採番の起点は作品ロックを取得した後にだけ読み取る。既定のREAD COMMITTED分離では、
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

	// 採番はRailsのフォームと同じく、作品のエピソード数 × ステップから始める。件数が
	// int64なのはCOUNTがそう返すためで、1作品が持ちうるエピソード数はsort_numberカラム
	// のint32の範囲に十分収まる。
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
				return nil, fmt.Errorf("animeの作成に失敗しました (%d行目): %w", i+1, err)
			}
			if _, err := classificationRepo.Create(ctx, classificationCreateParamsFromEpisode(episode, anime.ID)); err != nil {
				return nil, fmt.Errorf("anime_classificationの作成に失敗しました (%d行目): %w", i+1, err)
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
			return nil, fmt.Errorf("エピソードの作成に失敗しました (%d行目): %w", i+1, err)
		}
		episodeIDs = append(episodeIDs, episodeID)

		previous = nextSortAnchor(previous, episodeID, sortNumber)
	}

	createdCount := int32(len(rows)) // #nosec G115 -- 1回の送信は検証で100行までに制限される
	if err := workRepo.IncrementEpisodesCount(ctx, input.WorkID, createdCount); err != nil {
		return nil, fmt.Errorf("作品のエピソード件数の更新に失敗しました: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return &CreateEpisodesOutput{EpisodeIDs: episodeIDs}, nil
}

// manualCreationValidationErrorは、作品の状態が許さない送信を報告する。メッセージは
// グローバルに積む。送信された行に問題は無く、textareaを不正と印付けると、直す必要のない
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

// episodeFromCreateRowは送信された1行を、animes / anime_classificationsの写像が読む
// *model.Episodeのフィールドに射影する。これによりcreate経路もフェーズ2同期と同じ
// animeCreateParamsFromEpisode / classificationCreateParamsFromEpisodeヘルパーに通せ、両者が
// 同じエピソードから異なる行を導出することがなくなる。
//
// episodeFromAnimeSyncRowのpartial-loadパターンに倣い、写像対象のフィールドだけをセットして
// 残りの *model.Episodeはゼロ値のまま残す。ParentAnimeIDは親作品のanimeを運び、作品が
// 未マッピングのあいだはnilになる。新規エピソードはUnpublishedAt / DeletedAtをnilのまま
// にするためDerivedStatusはpublishedを報告し、写像が生むanimeもpublishedになる。
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

// nextSortAnchorは、次に作る行が直前のエピソードとして名指しする起点を返す。作成した
// ばかりのエピソードがそこまでで最大のsort_numberを持つならそれを、そうでなければ従来の
// 起点を返す。同値の場合は2つのうち後の行である新規エピソードを採る。
//
// 採番の起点が作品の最大sort_numberではなくエピソード数であるため、この比較が要る。既存の
// エピソードがより広い間隔で並んでいる作品では、そのエピソードが直前のエピソードとして名指し
// され続ける。これはRailsのコールバックの挙動そのもの。
func nextSortAnchor(previous *repository.DBEpisodeSortAnchor, episodeID model.EpisodeID, sortNumber int32) *repository.DBEpisodeSortAnchor {
	if previous != nil && previous.SortNumber > sortNumber {
		return previous
	}

	return &repository.DBEpisodeSortAnchor{ID: episodeID, SortNumber: sortNumber}
}
