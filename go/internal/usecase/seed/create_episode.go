package seed

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/schollz/progressbar/v3"
)

// CreateEpisodeParams エピソード作成のパラメータ
type CreateEpisodeParams struct {
	WorkID        int64
	Number        string
	Title         string
	SortNumber    int32
	PrevEpisodeID *int64 // 前のエピソードのID（最初のエピソードはnil）
}

// CreateEpisodeResult エピソード作成の結果
type CreateEpisodeResult struct {
	EpisodeID int64
}

// CreateEpisodeUsecase エピソード生成Usecase（シード専用、バルクインサート対応）
type CreateEpisodeUsecase struct {
	db *sql.DB
}

// NewCreateEpisodeUsecase 新しいCreateEpisodeUsecaseを作成
func NewCreateEpisodeUsecase(db *sql.DB) *CreateEpisodeUsecase {
	return &CreateEpisodeUsecase{
		db: db,
	}
}

// ExecuteBatch 複数のエピソードをバッチで作成します
// 1000件ごとにコミットしてパフォーマンスを最適化します
func (uc *CreateEpisodeUsecase) ExecuteBatch(ctx context.Context, episodes []CreateEpisodeParams, progressBar *progressbar.ProgressBar) ([]CreateEpisodeResult, error) {
	return uc.executeBatchWithTx(ctx, nil, episodes, progressBar)
}

// ExecuteBatchWithTx 複数のエピソードをバッチで作成します（テスト用：既存トランザクションを使用）
// txがnilの場合は内部でトランザクションを作成します
func (uc *CreateEpisodeUsecase) ExecuteBatchWithTx(ctx context.Context, tx *sql.Tx, episodes []CreateEpisodeParams, progressBar *progressbar.ProgressBar) ([]CreateEpisodeResult, error) {
	return uc.executeBatchWithTx(ctx, tx, episodes, progressBar)
}

// executeBatchWithTx 内部実装：トランザクションを受け取るか新規作成する
//
// 注意: エピソードはprev_episode_idの依存関係があるため、マルチ行INSERTは使用せず、
// 1件ずつ順番に処理します。これにより、work_idごとにエピソード連鎖が正しく維持されます。
func (uc *CreateEpisodeUsecase) executeBatchWithTx(ctx context.Context, existingTx *sql.Tx, episodes []CreateEpisodeParams, progressBar *progressbar.ProgressBar) ([]CreateEpisodeResult, error) {
	results := make([]CreateEpisodeResult, 0, len(episodes))

	// work_idごとに前のエピソードIDを追跡するマップ
	prevEpisodeIDMap := make(map[int64]int64)

	// 既存トランザクションがある場合は、バッチサイズを無視して全件処理
	if existingTx != nil {
		for _, episodeParam := range episodes {
			// work_idごとに前のエピソードIDを設定
			if prevID, exists := prevEpisodeIDMap[episodeParam.WorkID]; exists {
				episodeParam.PrevEpisodeID = &prevID
			}

			result, err := uc.createSingleEpisode(ctx, existingTx, episodeParam)
			if err != nil {
				return nil, fmt.Errorf("エピソード作成エラー（work_id: %d, number: %s）: %w", episodeParam.WorkID, episodeParam.Number, err)
			}
			results = append(results, *result)

			// 次のエピソードのために、作成したエピソードIDを保存
			prevEpisodeIDMap[episodeParam.WorkID] = result.EpisodeID

			// 進捗表示を更新
			if progressBar != nil {
				progressBar.Add(1)
			}
		}
		return results, nil
	}

	// 既存トランザクションがない場合は、3000件ごとにコミット
	batchSize := 3000
	for i := 0; i < len(episodes); i += batchSize {
		end := i + batchSize
		if end > len(episodes) {
			end = len(episodes)
		}
		batch := episodes[i:end]

		// トランザクション開始
		tx, err := uc.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("トランザクション開始エラー: %w", err)
		}
		defer tx.Rollback()

		// バッチ内のエピソードを1件ずつ作成
		for _, episodeParam := range batch {
			// work_idごとに前のエピソードIDを設定
			if prevID, exists := prevEpisodeIDMap[episodeParam.WorkID]; exists {
				episodeParam.PrevEpisodeID = &prevID
			}

			result, err := uc.createSingleEpisode(ctx, tx, episodeParam)
			if err != nil {
				return nil, fmt.Errorf("エピソード作成エラー（work_id: %d, number: %s）: %w", episodeParam.WorkID, episodeParam.Number, err)
			}
			results = append(results, *result)

			// 次のエピソードのために、作成したエピソードIDを保存
			prevEpisodeIDMap[episodeParam.WorkID] = result.EpisodeID

			// 進捗表示を更新
			if progressBar != nil {
				progressBar.Add(1)
			}
		}

		// コミット
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("トランザクションコミットエラー: %w", err)
		}
	}

	return results, nil
}

// createSingleEpisode 単一のエピソードを作成します（トランザクション内）
func (uc *CreateEpisodeUsecase) createSingleEpisode(ctx context.Context, tx *sql.Tx, params CreateEpisodeParams) (*CreateEpisodeResult, error) {
	// エピソードを作成
	episodeID, err := uc.createEpisode(ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("エピソードレコード作成エラー: %w", err)
	}

	return &CreateEpisodeResult{
		EpisodeID: episodeID,
	}, nil
}

// createEpisode episodesテーブルにレコードを作成します
func (uc *CreateEpisodeUsecase) createEpisode(ctx context.Context, tx *sql.Tx, params CreateEpisodeParams) (int64, error) {
	query := `
		INSERT INTO episodes (
			work_id, number, title, sort_number, prev_episode_id,
			episode_records_count, aasm_state,
			created_at, updated_at,
			title_ro, title_en, number_en,
			episode_record_bodies_count, ratings_count,
			fetch_syobocal
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7,
			$8, $9,
			$10, $11, $12,
			$13, $14,
			$15
		) RETURNING id
	`

	var episodeID int64
	err := tx.QueryRowContext(
		ctx,
		query,
		params.WorkID,
		params.Number,
		params.Title,
		params.SortNumber,
		params.PrevEpisodeID, // nil可（最初のエピソード）
		0,                    // episode_records_count (デフォルト0)
		"published",          // aasm_state (デフォルト'published')
		time.Now(),           // created_at
		time.Now(),           // updated_at
		"",                   // title_ro (デフォルト空文字)
		"",                   // title_en (デフォルト空文字)
		"",                   // number_en (デフォルト空文字)
		0,                    // episode_record_bodies_count (デフォルト0)
		0,                    // ratings_count (デフォルト0)
		false,                // fetch_syobocal (デフォルトfalse)
	).Scan(&episodeID)

	if err != nil {
		return 0, fmt.Errorf("episodesテーブルへの挿入エラー: %w", err)
	}

	return episodeID, nil
}

// GenerateEpisodeParamsForWork は指定された作品に対してランダムなエピソードパラメータを生成します
// episodeCountは生成するエピソード数（平均12話を想定）
//
// 注意: PrevEpisodeIDはnilで生成されます。ExecuteBatch内でwork_idごとに自動的に設定されます。
// エピソードは作品ごとにまとめて、かつ順番に並べてExecuteBatchに渡す必要があります。
func GenerateEpisodeParamsForWork(r *rand.Rand, workID int64, episodeCount int) []CreateEpisodeParams {
	episodes := make([]CreateEpisodeParams, 0, episodeCount)

	for i := 1; i <= episodeCount; i++ {
		number := fmt.Sprintf("第%d話", i)
		title := generateEpisodeSubtitle(r)
		// エピソード数は現実的にint32の範囲内に収まるため、オーバーフローの心配はない
		sortNumber := int32(i) // #nosec G115

		episodes = append(episodes, CreateEpisodeParams{
			WorkID:        workID,
			Number:        number,
			Title:         title,
			SortNumber:    sortNumber,
			PrevEpisodeID: nil, // ExecuteBatch内で自動的に設定される
		})
	}

	return episodes
}

// generateEpisodeSubtitle はエピソードのサブタイトルを生成します
func generateEpisodeSubtitle(r *rand.Rand) string {
	subtitles := []string{
		"始まりの物語",
		"新たな出会い",
		"試練の時",
		"運命の選択",
		"光と闇",
		"真実の扉",
		"約束の場所",
		"絆の力",
		"決戦の刻",
		"希望の明日",
		"最後の戦い",
		"未来への一歩",
		"秘密の暴露",
		"別れの時",
		"再会の約束",
		"奇跡の瞬間",
		"涙の理由",
		"笑顔のために",
		"勇気の証",
		"愛の形",
	}

	return subtitles[r.Intn(len(subtitles))]
}
