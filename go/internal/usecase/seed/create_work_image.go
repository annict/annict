package seed

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/seed"
	"github.com/schollz/progressbar/v3"
)

// CreateWorkImageParams 作品画像作成のパラメータ
type CreateWorkImageParams struct {
	WorkID int64
	UserID int64
}

// CreateWorkImageResult 作品画像作成の結果
type CreateWorkImageResult struct {
	WorkImageID int64
	ImagePath   string
}

// CreateWorkImageUsecase 作品画像生成Usecase（シード専用、バルクインサート対応）
type CreateWorkImageUsecase struct {
	db              *sql.DB
	queries         *query.Queries
	endpoint        string
	accessKeyID     string
	secretAccessKey string
	region          string
	bucketName      string
}

// NewCreateWorkImageUsecase 新しいCreateWorkImageUsecaseを作成
func NewCreateWorkImageUsecase(
	db *sql.DB,
	queries *query.Queries,
	endpoint string,
	accessKeyID string,
	secretAccessKey string,
	region string,
	bucketName string,
) *CreateWorkImageUsecase {
	return &CreateWorkImageUsecase{
		db:              db,
		queries:         queries,
		endpoint:        endpoint,
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
		region:          region,
		bucketName:      bucketName,
	}
}

// ExecuteBatch 複数の作品画像をバッチで作成します
// 各作品画像について：
// 1. ランダム画像を生成
// 2. Cloudflare R2にアップロード
// 3. work_imagesテーブルにレコードを作成
func (uc *CreateWorkImageUsecase) ExecuteBatch(ctx context.Context, params []CreateWorkImageParams, progressBar *progressbar.ProgressBar) ([]CreateWorkImageResult, error) {
	return uc.executeBatchWithTx(ctx, nil, params, progressBar)
}

// ExecuteBatchWithTx 複数の作品画像をバッチで作成します（テスト用：既存トランザクションを使用）
// txがnilの場合は内部でトランザクションを作成します
func (uc *CreateWorkImageUsecase) ExecuteBatchWithTx(ctx context.Context, tx *sql.Tx, params []CreateWorkImageParams, progressBar *progressbar.ProgressBar) ([]CreateWorkImageResult, error) {
	return uc.executeBatchWithTx(ctx, tx, params, progressBar)
}

// executeBatchWithTx 内部実装：トランザクションを受け取るか新規作成する
func (uc *CreateWorkImageUsecase) executeBatchWithTx(ctx context.Context, existingTx *sql.Tx, params []CreateWorkImageParams, progressBar *progressbar.ProgressBar) ([]CreateWorkImageResult, error) {
	// 既存トランザクションがある場合は、1件ずつ処理（テスト用）
	if existingTx != nil {
		results := make([]CreateWorkImageResult, 0, len(params))
		queries := uc.queries.WithTx(existingTx)
		for _, param := range params {
			result, err := uc.createSingleWorkImage(ctx, queries, param)
			if err != nil {
				return nil, fmt.Errorf("作品画像作成エラー (work_id=%d): %w", param.WorkID, err)
			}
			results = append(results, *result)

			// 進捗表示を更新
			if progressBar != nil {
				progressBar.Add(1)
			}
		}
		return results, nil
	}

	// 既存トランザクションがない場合は、並列処理で高速化
	// Worker Poolパターンで並列処理を実装（10並列）
	const numWorkers = 10

	// ワーカージョブの定義
	type workJob struct {
		param CreateWorkImageParams
		index int
	}

	// ワーカー結果の定義
	type workResult struct {
		result CreateWorkImageResult
		index  int
		err    error
	}

	// ジョブチャネルと結果チャネル
	jobs := make(chan workJob, len(params))
	resultsCh := make(chan workResult, len(params))

	// ワーカーを起動（goroutineで並列処理）
	for w := 0; w < numWorkers; w++ {
		go func() {
			// 各ワーカーはqueriesを持つ（トランザクションなし）
			queries := uc.queries
			for job := range jobs {
				result, err := uc.createSingleWorkImage(ctx, queries, job.param)
				if err != nil {
					resultsCh <- workResult{index: job.index, err: err}
				} else {
					resultsCh <- workResult{result: *result, index: job.index}
				}
			}
		}()
	}

	// ジョブを投入
	for i, param := range params {
		jobs <- workJob{param: param, index: i}
	}
	close(jobs)

	// 結果を回収（順序を保持するためにマップを使用）
	resultMap := make(map[int]CreateWorkImageResult)
	errorMap := make(map[int]error)

	for i := 0; i < len(params); i++ {
		res := <-resultsCh
		if res.err != nil {
			errorMap[res.index] = res.err
		} else {
			resultMap[res.index] = res.result
		}

		// 進捗表示を更新
		if progressBar != nil {
			progressBar.Add(1)
		}
	}

	// エラーがあれば最初のエラーを返す
	if len(errorMap) > 0 {
		// 最初のエラーを返す
		for _, err := range errorMap {
			return nil, err
		}
	}

	// 結果を順序通りに並べて返す
	orderedResults := make([]CreateWorkImageResult, len(params))
	for i := 0; i < len(params); i++ {
		orderedResults[i] = resultMap[i]
	}

	return orderedResults, nil
}

// createSingleWorkImage 単一の作品画像を作成します
func (uc *CreateWorkImageUsecase) createSingleWorkImage(ctx context.Context, queries *query.Queries, param CreateWorkImageParams) (*CreateWorkImageResult, error) {
	// 1. ランダム画像を生成（Shrineのpretty_locationプラグインの仕様に合わせて作品IDを渡す）
	img, err := seed.GenerateRandomWorkImage(param.WorkID)
	if err != nil {
		return nil, fmt.Errorf("ランダム画像生成エラー: %w", err)
	}

	// 2. Cloudflare R2にアップロード（R2設定がある場合のみ）
	if uc.endpoint != "" && uc.accessKeyID != "" && uc.secretAccessKey != "" && uc.bucketName != "" {
		if err := seed.UploadToR2(ctx, img, uc.endpoint, uc.accessKeyID, uc.secretAccessKey, uc.region, uc.bucketName); err != nil {
			return nil, fmt.Errorf("r2へのアップロードエラー: %w", err)
		}
	}

	// 3. Shrine形式のimage_data JSONを生成
	imageData, err := seed.GenerateShrineImageData(img)
	if err != nil {
		return nil, fmt.Errorf("shrine JSON生成エラー: %w", err)
	}

	// 4. work_imagesテーブルにレコードを作成
	now := time.Now()
	workImageID, err := queries.CreateWorkImage(ctx, query.CreateWorkImageParams{
		WorkID:    param.WorkID,
		UserID:    param.UserID,
		ImageData: imageData,
		Copyright: "",            // デフォルト空文字
		Asin:      "",            // デフォルト空文字
		ColorRgb:  "255,255,255", // デフォルト白色
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return nil, fmt.Errorf("work_imagesテーブルへの挿入エラー: %w", err)
	}

	return &CreateWorkImageResult{
		WorkImageID: workImageID,
		ImagePath:   img.Path,
	}, nil
}
