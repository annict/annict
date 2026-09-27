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

	// 実行 (エラーなく完了するはず)
	ctx := context.Background()
	err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("S3の設定が無いときのエラー = %v、期待値 = nil", err)
	}
}

// TestCleanupWorkImagesUsecase_EmptyBucketバケットが空の場合のテスト
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
	// 	t.Fatalf("想定外のエラー = %v", err)
	// }
}

// TestCleanupWorkImagesUsecase_WithObjectsオブジェクトが存在する場合のテスト
// 注: このテストは実際のS3接続が必要なため、統合テストとしてスキップします
func TestCleanupWorkImagesUsecase_WithObjects(t *testing.T) {
	t.Skip("統合テスト: 実際のS3接続が必要なため、ローカル環境ではスキップします")

	// 実際のテスト手順:
	// 1. テスト用の画像をS3にアップロード
	// 2. CleanupWorkImagesUsecaseを実行
	// 3. S3に画像が残っていないことを確認
}

// TestNewCleanupWorkImagesUsecaseコンストラクタのテスト
func TestNewCleanupWorkImagesUsecase(t *testing.T) {
	t.Parallel()

	endpoint := "https://test-endpoint.com"
	accessKeyID := "test-access-key"
	secretAccessKey := "test-secret-key"
	region := "auto"
	bucketName := "test-bucket"

	uc := NewCleanupWorkImagesUsecase(endpoint, accessKeyID, secretAccessKey, region, bucketName)

	if uc == nil {
		t.Fatal("usecaseがnilだった")
	}

	// 構造体のフィールドが正しく設定されているか確認
	if uc.endpoint != endpoint {
		t.Errorf("endpointの期待値 = %q、実測値 = %q", endpoint, uc.endpoint)
	}
	if uc.accessKeyID != accessKeyID {
		t.Errorf("accessKeyIDの期待値 = %q、実測値 = %q", accessKeyID, uc.accessKeyID)
	}
	if uc.secretAccessKey != secretAccessKey {
		t.Errorf("secretAccessKeyの期待値 = %q、実測値 = %q", secretAccessKey, uc.secretAccessKey)
	}
	if uc.region != region {
		t.Errorf("regionの期待値 = %q、実測値 = %q", region, uc.region)
	}
	if uc.bucketName != bucketName {
		t.Errorf("bucketNameの期待値 = %q、実測値 = %q", bucketName, uc.bucketName)
	}
}

// TestCleanupWorkImagesUsecase_PartialS3Config一部のS3設定のみが設定されている場合
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

			// 一部の設定のみの場合はスキップされるはず (エラーなし)
			if err != nil {
				t.Fatalf("S3の設定が一部だけのときのエラー = %v、期待値 = nil", err)
			}
		})
	}
}
