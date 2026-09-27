package seed

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/seed"
	"github.com/annict/annict/go/internal/testutil"
)

// TestCreateWorkImageUsecase_ExecuteBatchWithTxはExecuteBatchWithTxメソッドのテスト (トランザクションあり、シーケンシャル処理)
func TestCreateWorkImageUsecase_ExecuteBatchWithTx(t *testing.T) {
	// テストケース
	tests := []struct {
		name      string
		numImages int
		wantErr   bool
	}{
		{
			name:      "正常系: 1つの作品画像を作成",
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
			uc := NewCreateWorkImageUsecase(db, queries, "", "", "", "", "")

			// テスト用ユーザーを作成
			userID := testutil.NewUserBuilder(t, tx).Build()

			// テスト用作品を作成
			params := make([]CreateWorkImageParams, tt.numImages)
			for i := 0; i < tt.numImages; i++ {
				workID := testutil.NewWorkBuilder(t, tx).Build()
				params[i] = CreateWorkImageParams{
					WorkID: workID,
					UserID: userID,
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

			// 作成された作品画像の数を確認
			if len(results) != tt.numImages {
				t.Errorf("作成された作品画像の数 = %d、期待値 = %d", len(results), tt.numImages)
			}

			// 各結果を検証
			for i, result := range results {
				// work_imagesテーブルにレコードが作成されたか確認
				if result.WorkImageID == 0 {
					t.Errorf("results[%d]: WorkImageIDが0です", i)
				}

				// 画像パスが生成されているか確認
				if result.ImagePath == "" {
					t.Errorf("results[%d]: ImagePathが空です", i)
				}

				// 画像パスのプレフィックスを確認
				expectedPrefix := seed.ShrinePathPrefix
				if len(result.ImagePath) < len(expectedPrefix) || result.ImagePath[:len(expectedPrefix)] != expectedPrefix {
					t.Errorf("results[%d]: ImagePathのプレフィックス = %s、期待値 = %sで始まること", i, result.ImagePath, expectedPrefix)
				}
			}

			// work_imagesテーブルから1件目のレコードを取得して検証
			if len(results) > 0 {
				row := tx.QueryRow("SELECT work_id, user_id, image_data FROM work_images WHERE id = $1", results[0].WorkImageID)
				var workID, userID int64
				var imageData string
				if err := row.Scan(&workID, &userID, &imageData); err != nil {
					t.Fatalf("work_imagesテーブルからの取得エラー: %v", err)
				}

				// work_idとuser_idを確認
				if model.WorkID(workID) != params[0].WorkID {
					t.Errorf("work_id = %d、期待値 = %d", workID, params[0].WorkID)
				}
				if model.UserID(userID) != params[0].UserID {
					t.Errorf("user_id = %d、期待値 = %d", userID, params[0].UserID)
				}

				// image_dataのJSON形式を確認
				var shrineData seed.ShrineImageData
				if err := json.Unmarshal([]byte(imageData), &shrineData); err != nil {
					t.Fatalf("image_dataのJSONパースエラー: %v", err)
				}

				// Shrine形式のフィールドを確認
				if shrineData.Master.Storage != "store" {
					t.Errorf("storage = %s、期待値 = store", shrineData.Master.Storage)
				}
				if shrineData.Master.Metadata.MimeType != "image/png" {
					t.Errorf("mime_type = %s、期待値 = image/png", shrineData.Master.Metadata.MimeType)
				}
				if shrineData.Master.Metadata.Width != seed.WorkImageWidth {
					t.Errorf("width = %d、期待値 = %d", shrineData.Master.Metadata.Width, seed.WorkImageWidth)
				}
				if shrineData.Master.Metadata.Height != seed.WorkImageHeight {
					t.Errorf("height = %d、期待値 = %d", shrineData.Master.Metadata.Height, seed.WorkImageHeight)
				}
				if shrineData.Master.Metadata.Size <= 0 {
					t.Errorf("size = %d、期待値 = 1以上", shrineData.Master.Metadata.Size)
				}
			}
		})
	}
}

// TestCreateWorkImageUsecase_ExecuteBatchはExecuteBatchメソッドのテスト (トランザクションなし、並列処理)
// このテストは並列処理パスがコンパイルされ、基本的に動作することを確認します
func TestCreateWorkImageUsecase_ExecuteBatch(t *testing.T) {
	// テストケース
	tests := []struct {
		name      string
		numImages int
		wantErr   bool
	}{
		{
			name:      "正常系: 3つの作品画像を並列処理で作成",
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
			uc := NewCreateWorkImageUsecase(db, queries, "", "", "", "", "")

			// テスト用ユーザーを作成
			userID := testutil.NewUserBuilder(t, tx).Build()

			// テスト用作品を作成
			params := make([]CreateWorkImageParams, tt.numImages)
			for i := 0; i < tt.numImages; i++ {
				workID := testutil.NewWorkBuilder(t, tx).Build()
				params[i] = CreateWorkImageParams{
					WorkID: workID,
					UserID: userID,
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

			// 作成された作品画像の数を確認
			if len(results) != tt.numImages {
				t.Errorf("作成された作品画像の数 = %d、期待値 = %d", len(results), tt.numImages)
			}

			// 各結果を検証
			for i, result := range results {
				// work_imagesテーブルにレコードが作成されたか確認
				if result.WorkImageID == 0 {
					t.Errorf("results[%d]: WorkImageIDが0です", i)
				}

				// 画像パスが生成されているか確認
				if result.ImagePath == "" {
					t.Errorf("results[%d]: ImagePathが空です", i)
				}

				// 画像パスのプレフィックスを確認
				expectedPrefix := seed.ShrinePathPrefix
				if len(result.ImagePath) < len(expectedPrefix) || result.ImagePath[:len(expectedPrefix)] != expectedPrefix {
					t.Errorf("results[%d]: ImagePathのプレフィックス = %s、期待値 = %sで始まること", i, result.ImagePath, expectedPrefix)
				}

				// work_imagesテーブルからレコードを取得して検証
				row := db.QueryRow("SELECT work_id, user_id, image_data FROM work_images WHERE id = $1", result.WorkImageID)
				var workID, userID int64
				var imageData string
				if err := row.Scan(&workID, &userID, &imageData); err != nil {
					t.Errorf("results[%d]: work_imagesテーブルからの取得エラー: %v", i, err)
					continue
				}

				// work_idとuser_idを確認
				if model.WorkID(workID) != params[i].WorkID {
					t.Errorf("results[%d]: work_id = %d、期待値 = %d", i, workID, params[i].WorkID)
				}
				if model.UserID(userID) != params[i].UserID {
					t.Errorf("results[%d]: user_id = %d、期待値 = %d", i, userID, params[i].UserID)
				}

				// image_dataのJSON形式を確認
				var shrineData seed.ShrineImageData
				if err := json.Unmarshal([]byte(imageData), &shrineData); err != nil {
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
				if shrineData.Master.Metadata.Width != seed.WorkImageWidth {
					t.Errorf("results[%d]: width = %d、期待値 = %d", i, shrineData.Master.Metadata.Width, seed.WorkImageWidth)
				}
				if shrineData.Master.Metadata.Height != seed.WorkImageHeight {
					t.Errorf("results[%d]: height = %d、期待値 = %d", i, shrineData.Master.Metadata.Height, seed.WorkImageHeight)
				}
				if shrineData.Master.Metadata.Size <= 0 {
					t.Errorf("results[%d]のsize = %d、期待値 = 1以上", i, shrineData.Master.Metadata.Size)
				}
			}

			// クリーンアップ: テスト後に作成されたwork_imagesレコードを削除
			for _, result := range results {
				if _, err := db.Exec("DELETE FROM work_images WHERE id = $1", result.WorkImageID); err != nil {
					t.Logf("警告: work_imagesレコードの削除エラー (id=%d): %v", result.WorkImageID, err)
				}
			}
		})
	}
}
