package seed

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/seed"
	"github.com/annict/annict/go/internal/testutil"
)

// TestCreateProfileImageUsecase_ExecuteBatchWithTxはExecuteBatchWithTxメソッドのテスト (トランザクションあり、シーケンシャル処理)
func TestCreateProfileImageUsecase_ExecuteBatchWithTx(t *testing.T) {
	// テストケース
	tests := []struct {
		name      string
		numImages int
		wantErr   bool
	}{
		{
			name:      "正常系: 1つのプロフィール画像を作成",
			numImages: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 各サブテストで新しいトランザクションを作成
			db, tx := testutil.SetupTx(t)
			queries := query.New(db)

			// Usecaseを作成 (R2設定は空にしてアップロードをスキップ)
			uc := NewCreateProfileImageUsecase(db, queries, "", "", "", "", "")

			// テスト用ユーザーを作成 (自動的にプロフィールも作成される)
			params := make([]CreateProfileImageParams, tt.numImages)
			for i := 0; i < tt.numImages; i++ {
				userID := testutil.NewUserBuilder(t, tx).Build()

				// プロフィールIDを取得
				row := tx.QueryRow("SELECT id FROM profiles WHERE user_id = $1", userID)
				var profileID int64
				if err := row.Scan(&profileID); err != nil {
					t.Fatalf("プロフィールIDの取得エラー: %v", err)
				}

				params[i] = CreateProfileImageParams{
					ProfileID: profileID,
					UserID:    userID,
				}
			}

			// ExecuteBatchWithTxを実行
			results, err := uc.ExecuteBatchWithTx(context.Background(), tx, params, nil)

			// エラーチェック
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteBatchWithTx()のエラー = %v、期待値 = %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			// 作成されたプロフィール画像の数を確認
			if len(results) != tt.numImages {
				t.Errorf("作成されたプロフィール画像の数 = %d、期待値 = %d", len(results), tt.numImages)
			}

			// 各結果を検証
			for i, result := range results {
				// ProfileIDが正しいか確認
				if result.ProfileID != params[i].ProfileID {
					t.Errorf("results[%d]: ProfileID = %d、期待値 = %d", i, result.ProfileID, params[i].ProfileID)
				}

				// 画像パスが生成されているか確認
				if result.ImagePath == "" {
					t.Errorf("results[%d]: ImagePathが空です", i)
				}

				// 画像パスのプレフィックスを確認
				expectedPrefix := seed.ShrineProfilePathPrefix
				if len(result.ImagePath) < len(expectedPrefix) || result.ImagePath[:len(expectedPrefix)] != expectedPrefix {
					t.Errorf("results[%d]: ImagePathのプレフィックス = %s、期待値 = %sで始まること", i, result.ImagePath, expectedPrefix)
				}
			}

			// profilesテーブルから1件目のレコードを取得して検証
			if len(results) > 0 {
				row := tx.QueryRow("SELECT user_id, image_data FROM profiles WHERE id = $1", results[0].ProfileID)
				var userID int64
				var imageData *string
				if err := row.Scan(&userID, &imageData); err != nil {
					t.Fatalf("profilesテーブルからの取得エラー: %v", err)
				}

				// user_idを確認
				if userID != int64(params[0].UserID) {
					t.Errorf("user_id = %d、期待値 = %d", userID, params[0].UserID)
				}

				// image_dataが設定されているか確認
				if imageData == nil {
					t.Fatal("image_dataがnullです")
				}

				// image_dataのJSON形式を確認
				var shrineData seed.ShrineImageData
				if err := json.Unmarshal([]byte(*imageData), &shrineData); err != nil {
					t.Fatalf("image_dataのJSONパースエラー: %v", err)
				}

				// Shrine形式のフィールドを確認
				if shrineData.Master.Storage != "store" {
					t.Errorf("storage = %s、期待値 = store", shrineData.Master.Storage)
				}
				if shrineData.Master.Metadata.MimeType != "image/png" {
					t.Errorf("mime_type = %s、期待値 = image/png", shrineData.Master.Metadata.MimeType)
				}
				if shrineData.Master.Metadata.Width != seed.ProfileImageWidth {
					t.Errorf("width = %d、期待値 = %d", shrineData.Master.Metadata.Width, seed.ProfileImageWidth)
				}
				if shrineData.Master.Metadata.Height != seed.ProfileImageHeight {
					t.Errorf("height = %d、期待値 = %d", shrineData.Master.Metadata.Height, seed.ProfileImageHeight)
				}
				if shrineData.Master.Metadata.Size <= 0 {
					t.Errorf("size = %d、期待値 = 1以上", shrineData.Master.Metadata.Size)
				}
			}
		})
	}
}

// TestCreateProfileImageUsecase_ExecuteBatchはExecuteBatchメソッドのテスト (トランザクションなし、並列処理)
// このテストは並列処理パスがコンパイルされ、基本的に動作することを確認します
func TestCreateProfileImageUsecase_ExecuteBatch(t *testing.T) {
	// テストケース
	tests := []struct {
		name      string
		numImages int
		wantErr   bool
	}{
		{
			name:      "正常系: 3つのプロフィール画像を並列処理で作成",
			numImages: 3,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テストDBをセットアップ (トランザクションはコミット前に準備データを作成)
			db, tx := testutil.SetupTx(t)
			queries := query.New(db)

			// Usecaseを作成 (R2設定は空にしてアップロードをスキップ)
			uc := NewCreateProfileImageUsecase(db, queries, "", "", "", "", "")

			// テスト用ユーザーとプロフィールを作成
			params := make([]CreateProfileImageParams, tt.numImages)
			for i := 0; i < tt.numImages; i++ {
				userID := testutil.NewUserBuilder(t, tx).Build()

				// プロフィールIDを取得
				row := tx.QueryRow("SELECT id FROM profiles WHERE user_id = $1", userID)
				var profileID int64
				if err := row.Scan(&profileID); err != nil {
					t.Fatalf("プロフィールIDの取得エラー: %v", err)
				}

				params[i] = CreateProfileImageParams{
					ProfileID: profileID,
					UserID:    userID,
				}
			}

			// トランザクションをコミット (並列処理で参照するため)
			if err := tx.Commit(); err != nil {
				t.Fatalf("トランザクションのコミットエラー: %v", err)
			}

			// ExecuteBatchを実行 (トランザクションなし、並列処理パス)
			results, err := uc.ExecuteBatch(context.Background(), params, nil)

			// エラーチェック
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteBatch()のエラー = %v、期待値 = %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			// 作成されたプロフィール画像の数を確認
			if len(results) != tt.numImages {
				t.Errorf("作成されたプロフィール画像の数 = %d、期待値 = %d", len(results), tt.numImages)
			}

			// 各結果を検証
			for i, result := range results {
				// ProfileIDが正しいか確認
				if result.ProfileID != params[i].ProfileID {
					t.Errorf("results[%d]: ProfileID = %d、期待値 = %d", i, result.ProfileID, params[i].ProfileID)
				}

				// 画像パスが生成されているか確認
				if result.ImagePath == "" {
					t.Errorf("results[%d]: ImagePathが空です", i)
				}

				// 画像パスのプレフィックスを確認
				expectedPrefix := seed.ShrineProfilePathPrefix
				if len(result.ImagePath) < len(expectedPrefix) || result.ImagePath[:len(expectedPrefix)] != expectedPrefix {
					t.Errorf("results[%d]: ImagePathのプレフィックス = %s、期待値 = %sで始まること", i, result.ImagePath, expectedPrefix)
				}

				// profilesテーブルからレコードを取得して検証
				row := db.QueryRow("SELECT user_id, image_data FROM profiles WHERE id = $1", result.ProfileID)
				var userID int64
				var imageData *string
				if err := row.Scan(&userID, &imageData); err != nil {
					t.Errorf("results[%d]: profilesテーブルからの取得エラー: %v", i, err)
					continue
				}

				// user_idを確認
				if userID != int64(params[i].UserID) {
					t.Errorf("results[%d]: user_id = %d、期待値 = %d", i, userID, params[i].UserID)
				}

				// image_dataが設定されているか確認
				if imageData == nil {
					t.Errorf("results[%d]: image_dataがnullです", i)
					continue
				}

				// image_dataのJSON形式を確認
				var shrineData seed.ShrineImageData
				if err := json.Unmarshal([]byte(*imageData), &shrineData); err != nil {
					t.Errorf("results[%d]: image_dataのJSONパースエラー: %v", i, err)
					continue
				}

				// Shrine形式のフィールドを確認
				if shrineData.Master.Storage != "store" {
					t.Errorf("results[%d]: storage = %s、期待値 = store", i, shrineData.Master.Storage)
				}
				if shrineData.Master.Metadata.MimeType != "image/png" {
					t.Errorf("results[%d]: mime_type = %s、期待値 = image/png", i, shrineData.Master.Metadata.MimeType)
				}
				if shrineData.Master.Metadata.Width != seed.ProfileImageWidth {
					t.Errorf("results[%d]: width = %d、期待値 = %d", i, shrineData.Master.Metadata.Width, seed.ProfileImageWidth)
				}
				if shrineData.Master.Metadata.Height != seed.ProfileImageHeight {
					t.Errorf("results[%d]: height = %d、期待値 = %d", i, shrineData.Master.Metadata.Height, seed.ProfileImageHeight)
				}
				if shrineData.Master.Metadata.Size <= 0 {
					t.Errorf("results[%d]のsize = %d、期待値 = 1以上", i, shrineData.Master.Metadata.Size)
				}
			}

			// クリーンアップ: テスト後に作成されたprofilesレコードのimage_dataをクリア
			for _, result := range results {
				if _, err := db.Exec("UPDATE profiles SET image_data = NULL WHERE id = $1", result.ProfileID); err != nil {
					t.Logf("警告: profilesレコードのimage_dataクリアエラー (id=%d): %v", result.ProfileID, err)
				}
			}
		})
	}
}
