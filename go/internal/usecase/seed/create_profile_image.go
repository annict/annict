package seed

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/seed"
	"github.com/schollz/progressbar/v3"
)

// CreateProfileImageParams プロフィール画像作成のパラメータ
type CreateProfileImageParams struct {
	ProfileID int64
	UserID    int64
}

// CreateProfileImageResult プロフィール画像作成の結果
type CreateProfileImageResult struct {
	ProfileID int64
	ImagePath string
}

// CreateProfileImageUsecase プロフィール画像生成Usecase（シード専用、バルクインサート対応）
type CreateProfileImageUsecase struct {
	db              *sql.DB
	queries         *query.Queries
	endpoint        string
	accessKeyID     string
	secretAccessKey string
	region          string
	bucketName      string
}

// NewCreateProfileImageUsecase 新しいCreateProfileImageUsecaseを作成
func NewCreateProfileImageUsecase(
	db *sql.DB,
	queries *query.Queries,
	endpoint string,
	accessKeyID string,
	secretAccessKey string,
	region string,
	bucketName string,
) *CreateProfileImageUsecase {
	return &CreateProfileImageUsecase{
		db:              db,
		queries:         queries,
		endpoint:        endpoint,
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
		region:          region,
		bucketName:      bucketName,
	}
}

// ExecuteBatch 複数のプロフィール画像をバッチで作成します
// 各プロフィール画像について：
// 1. ランダム画像を生成
// 2. Cloudflare R2にアップロード
// 3. profilesテーブルのimage_dataカラムを更新
func (uc *CreateProfileImageUsecase) ExecuteBatch(ctx context.Context, params []CreateProfileImageParams, progressBar *progressbar.ProgressBar) ([]CreateProfileImageResult, error) {
	return uc.executeBatchWithTx(ctx, nil, params, progressBar)
}

// ExecuteBatchWithTx 複数のプロフィール画像をバッチで作成します（テスト用：既存トランザクションを使用）
// txがnilの場合は内部でトランザクションを作成します
func (uc *CreateProfileImageUsecase) ExecuteBatchWithTx(ctx context.Context, tx *sql.Tx, params []CreateProfileImageParams, progressBar *progressbar.ProgressBar) ([]CreateProfileImageResult, error) {
	return uc.executeBatchWithTx(ctx, tx, params, progressBar)
}

// executeBatchWithTx 内部実装：トランザクションを受け取るか新規作成する
func (uc *CreateProfileImageUsecase) executeBatchWithTx(ctx context.Context, existingTx *sql.Tx, params []CreateProfileImageParams, progressBar *progressbar.ProgressBar) ([]CreateProfileImageResult, error) {
	// 既存トランザクションがある場合は、1件ずつ処理（テスト用）
	if existingTx != nil {
		results := make([]CreateProfileImageResult, 0, len(params))
		queries := uc.queries.WithTx(existingTx)
		for _, param := range params {
			result, err := uc.createSingleProfileImage(ctx, queries, param)
			if err != nil {
				return nil, fmt.Errorf("プロフィール画像作成エラー (profile_id=%d): %w", param.ProfileID, err)
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
		param CreateProfileImageParams
		index int
	}

	// ワーカー結果の定義
	type workResult struct {
		result CreateProfileImageResult
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
				result, err := uc.createSingleProfileImage(ctx, queries, job.param)
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
	resultMap := make(map[int]CreateProfileImageResult)
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
	orderedResults := make([]CreateProfileImageResult, len(params))
	for i := 0; i < len(params); i++ {
		orderedResults[i] = resultMap[i]
	}

	return orderedResults, nil
}

// createSingleProfileImage 単一のプロフィール画像を作成します
func (uc *CreateProfileImageUsecase) createSingleProfileImage(ctx context.Context, queries *query.Queries, param CreateProfileImageParams) (*CreateProfileImageResult, error) {
	// 1. ランダム画像を生成（Shrineのpretty_locationプラグインの仕様に合わせてプロフィールIDを渡す）
	img, err := seed.GenerateRandomProfileImage(param.ProfileID)
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

	// 4. profilesテーブルのimage_dataカラムを更新
	err = queries.UpdateProfileImageData(ctx, query.UpdateProfileImageDataParams{
		ImageData: sql.NullString{String: imageData, Valid: true},
		ID:        param.ProfileID,
	})
	if err != nil {
		return nil, fmt.Errorf("profilesテーブルの更新エラー: %w", err)
	}

	return &CreateProfileImageResult{
		ProfileID: param.ProfileID,
		ImagePath: img.Path,
	}, nil
}
