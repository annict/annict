// Package storageはS3互換オブジェクトストレージ (Cloudflare R2) との連携機能を提供します
package storage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// deleteObjectsLimitはDeleteObjects APIが1回のリクエストで削除できるオブジェクト数の上限。
const deleteObjectsLimit = 1000

// NewClientはS3互換オブジェクトストレージのクライアントを作成します
func NewClient(endpoint, accessKeyID, secretAccessKey, region string) *s3.Client {
	return s3.NewFromConfig(aws.Config{
		Region:       region,
		Credentials:  credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
		BaseEndpoint: aws.String(endpoint),
	})
}

// Adapterはaws-sdk-goのS3クライアントをラップし、呼び出し側 (UseCase) のinterfaceに
// 適合させて、aws-sdk-goの型をUseCaseシグネチャから締め出す。
type Adapter struct {
	client     *s3.Client
	bucketName string
}

// NewAdapterは渡されたS3クライアントとバケットを使うAdapterを作成する。
func NewAdapter(client *s3.Client, bucketName string) *Adapter {
	return &Adapter{client: client, bucketName: bucketName}
}

// DeleteObjectsは指定したキーのオブジェクトを削除する。存在しないキーは削除済みとして扱われる
// (S3の仕様) ため、同じキーで再度呼んでも失敗しない。
func (a *Adapter) DeleteObjects(ctx context.Context, keys []string) error {
	for start := 0; start < len(keys); start += deleteObjectsLimit {
		end := min(start+deleteObjectsLimit, len(keys))

		identifiers := make([]types.ObjectIdentifier, 0, end-start)
		for _, key := range keys[start:end] {
			identifiers = append(identifiers, types.ObjectIdentifier{Key: aws.String(key)})
		}

		output, err := a.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(a.bucketName),
			Delete: &types.Delete{
				Objects: identifiers,
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			return fmt.Errorf("オブジェクトの削除に失敗: %w", err)
		}
		// DeleteObjectsはリクエストが成功しても、オブジェクトごとの失敗を結果に含めて返す。
		if len(output.Errors) > 0 {
			first := output.Errors[0]
			return fmt.Errorf("%d件のオブジェクトの削除に失敗 (最初の失敗: キー %s, コード %s, %s)",
				len(output.Errors), aws.ToString(first.Key), aws.ToString(first.Code), aws.ToString(first.Message))
		}
	}
	return nil
}
