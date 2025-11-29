package seed

import (
	"context"
	"testing"
)

// TestCleanupWorkImagesUsecase_NoS3Config S3設定がない場合はスキップする
func TestCleanupWorkImagesUsecase_NoS3Config(t *testing.T) {
	t.Parallel()

	// S3設定なしでUsecaseを作成
	uc := NewCleanupWorkImagesUsecase("", "", "", "", "")

	// 実行（エラーなく完了するはず）
	ctx := context.Background()
	err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("expected no error when S3 config is not set, got: %v", err)
	}
}

// TestCleanupWorkImagesUsecase_EmptyBucket バケットが空の場合のテスト
// 注: このテストは実際のS3接続が必要なため、統合テストとしてスキップします
func TestCleanupWorkImagesUsecase_EmptyBucket(t *testing.T) {
	t.Skip("統合テスト: 実際のS3接続が必要なため、ローカル環境ではスキップします")

	// 実際のCloudflare R2設定でテストする場合のサンプルコード
	// uc := NewCleanupWorkImagesUsecase(
	// 	"https://your-account-id.r2.cloudflarestorage.com",
	// 	"your-access-key-id",
	// 	"your-secret-access-key",
	// 	"auto",
	// 	"your-bucket-name",
	// )
	//
	// ctx := context.Background()
	// err := uc.Execute(ctx)
	// if err != nil {
	// 	t.Fatalf("unexpected error: %v", err)
	// }
}

// TestCleanupWorkImagesUsecase_WithObjects オブジェクトが存在する場合のテスト
// 注: このテストは実際のS3接続が必要なため、統合テストとしてスキップします
func TestCleanupWorkImagesUsecase_WithObjects(t *testing.T) {
	t.Skip("統合テスト: 実際のS3接続が必要なため、ローカル環境ではスキップします")

	// 実際のテスト手順:
	// 1. テスト用の画像をS3にアップロード
	// 2. CleanupWorkImagesUsecaseを実行
	// 3. S3に画像が残っていないことを確認
}

// TestNewCleanupWorkImagesUsecase コンストラクタのテスト
func TestNewCleanupWorkImagesUsecase(t *testing.T) {
	t.Parallel()

	endpoint := "https://test-endpoint.com"
	accessKeyID := "test-access-key"
	secretAccessKey := "test-secret-key"
	region := "auto"
	bucketName := "test-bucket"

	uc := NewCleanupWorkImagesUsecase(endpoint, accessKeyID, secretAccessKey, region, bucketName)

	if uc == nil {
		t.Fatal("expected non-nil usecase")
	}

	// 構造体のフィールドが正しく設定されているか確認
	if uc.endpoint != endpoint {
		t.Errorf("expected endpoint %q, got %q", endpoint, uc.endpoint)
	}
	if uc.accessKeyID != accessKeyID {
		t.Errorf("expected accessKeyID %q, got %q", accessKeyID, uc.accessKeyID)
	}
	if uc.secretAccessKey != secretAccessKey {
		t.Errorf("expected secretAccessKey %q, got %q", secretAccessKey, uc.secretAccessKey)
	}
	if uc.region != region {
		t.Errorf("expected region %q, got %q", region, uc.region)
	}
	if uc.bucketName != bucketName {
		t.Errorf("expected bucketName %q, got %q", bucketName, uc.bucketName)
	}
}

// TestCleanupWorkImagesUsecase_PartialS3Config 一部のS3設定のみが設定されている場合
func TestCleanupWorkImagesUsecase_PartialS3Config(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		endpoint        string
		accessKeyID     string
		secretAccessKey string
		region          string
		bucketName      string
	}{
		{
			name:            "endpointのみ設定",
			endpoint:        "https://test.com",
			accessKeyID:     "",
			secretAccessKey: "",
			region:          "",
			bucketName:      "",
		},
		{
			name:            "accessKeyIDのみ設定",
			endpoint:        "",
			accessKeyID:     "key",
			secretAccessKey: "",
			region:          "",
			bucketName:      "",
		},
		{
			name:            "secretAccessKeyのみ設定",
			endpoint:        "",
			accessKeyID:     "",
			secretAccessKey: "secret",
			region:          "",
			bucketName:      "",
		},
		{
			name:            "bucketNameのみ設定",
			endpoint:        "",
			accessKeyID:     "",
			secretAccessKey: "",
			region:          "",
			bucketName:      "bucket",
		},
		{
			name:            "endpoint+accessKeyIDのみ設定",
			endpoint:        "https://test.com",
			accessKeyID:     "key",
			secretAccessKey: "",
			region:          "",
			bucketName:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := NewCleanupWorkImagesUsecase(
				tt.endpoint,
				tt.accessKeyID,
				tt.secretAccessKey,
				tt.region,
				tt.bucketName,
			)

			ctx := context.Background()
			err := uc.Execute(ctx)

			// 一部の設定のみの場合はスキップされるはず（エラーなし）
			if err != nil {
				t.Fatalf("expected no error when partial S3 config is set, got: %v", err)
			}
		})
	}
}
