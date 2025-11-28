package seed

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/annict/annict/internal/seed"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/schollz/progressbar/v3"
)

// CleanupProfileImagesUsecase プロフィール画像クリーンアップUsecase（シード専用）
// Cloudflare R2上の shrine/profile/ プレフィックス配下のすべての画像を削除します
type CleanupProfileImagesUsecase struct {
	endpoint        string
	accessKeyID     string
	secretAccessKey string
	region          string
	bucketName      string
}

// NewCleanupProfileImagesUsecase 新しいCleanupProfileImagesUsecaseを作成
func NewCleanupProfileImagesUsecase(
	endpoint string,
	accessKeyID string,
	secretAccessKey string,
	region string,
	bucketName string,
) *CleanupProfileImagesUsecase {
	return &CleanupProfileImagesUsecase{
		endpoint:        endpoint,
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
		region:          region,
		bucketName:      bucketName,
	}
}

// Execute Cloudflare R2上の shrine/profile/ プレフィックス配下のすべての画像を削除します
// データベース上の profiles テーブルのimage_dataカラムは既に cmd/seed/main.go の cleanupExistingData で削除済みです
func (uc *CleanupProfileImagesUsecase) Execute(ctx context.Context) error {
	// S3設定がない場合はスキップ
	if uc.endpoint == "" || uc.accessKeyID == "" || uc.secretAccessKey == "" || uc.bucketName == "" {
		slog.Info("S3設定がないため、プロフィール画像のクリーンアップをスキップします")
		return nil
	}

	// S3クライアントを作成（Cloudflare R2はS3互換API）
	cfg := aws.Config{
		Region: uc.region,
		Credentials: credentials.NewStaticCredentialsProvider(
			uc.accessKeyID,
			uc.secretAccessKey,
			"",
		),
		BaseEndpoint: aws.String(uc.endpoint),
	}
	client := s3.NewFromConfig(cfg)

	// shrine/profile/ プレフィックス配下のすべてのオブジェクトを取得
	slog.Info("S3バケット内のプロフィール画像を検索しています...")
	objects, err := uc.listAllObjects(ctx, client)
	if err != nil {
		return fmt.Errorf("オブジェクト一覧取得エラー: %w", err)
	}

	if len(objects) == 0 {
		slog.Info("削除対象のプロフィール画像が見つかりませんでした")
		return nil
	}

	slog.Info("削除対象のプロフィール画像", "count", len(objects))

	// 進捗バー
	bar := progressbar.NewOptions(len(objects),
		progressbar.OptionSetDescription("プロフィール画像削除"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	// DeleteObjects APIは最大1000件まで削除可能なので、バッチで削除
	deleteBatchSize := 1000
	for i := 0; i < len(objects); i += deleteBatchSize {
		end := i + deleteBatchSize
		if end > len(objects) {
			end = len(objects)
		}
		batch := objects[i:end]

		if err := uc.deleteObjectsBatch(ctx, client, batch); err != nil {
			return fmt.Errorf("オブジェクト削除エラー: %w", err)
		}

		// 進捗表示を更新
		bar.Add(len(batch))
	}

	fmt.Println() // プログレスバーの後に改行
	slog.Info("プロフィール画像のクリーンアップが完了しました", "deleted_count", len(objects))

	return nil
}

// listAllObjects S3バケット内の shrine/profile/ プレフィックス配下のすべてのオブジェクトを取得します
func (uc *CleanupProfileImagesUsecase) listAllObjects(ctx context.Context, client *s3.Client) ([]types.Object, error) {
	var allObjects []types.Object
	var continuationToken *string

	// ListObjectsV2をページネーションで実行
	for {
		output, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(uc.bucketName),
			Prefix:            aws.String(seed.ShrineProfilePathPrefix + "/"),
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return nil, fmt.Errorf("s3 ListObjectsV2エラー: %w", err)
		}

		allObjects = append(allObjects, output.Contents...)

		// 次のページがあるか確認
		if output.IsTruncated != nil && *output.IsTruncated {
			continuationToken = output.NextContinuationToken
		} else {
			break
		}
	}

	return allObjects, nil
}

// deleteObjectsBatch 複数のオブジェクトをバッチで削除します（最大1000件）
func (uc *CleanupProfileImagesUsecase) deleteObjectsBatch(ctx context.Context, client *s3.Client, objects []types.Object) error {
	// DeleteObjects用のObjectIdentifierリストを作成
	identifiers := make([]types.ObjectIdentifier, len(objects))
	for i, obj := range objects {
		identifiers[i] = types.ObjectIdentifier{
			Key: obj.Key,
		}
	}

	// バッチ削除を実行
	_, err := client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(uc.bucketName),
		Delete: &types.Delete{
			Objects: identifiers,
			Quiet:   aws.Bool(true), // エラーのみ返す
		},
	})
	if err != nil {
		return fmt.Errorf("s3 DeleteObjectsエラー: %w", err)
	}

	return nil
}
