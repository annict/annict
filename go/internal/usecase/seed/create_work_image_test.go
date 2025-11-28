package seed

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/seed"
	"github.com/annict/annict/internal/testutil"
)

// TestCreateWorkImageUsecase_ExecuteBatchWithTx はExecuteBatchWithTxメソッドのテスト（トランザクションあり、シーケンシャル処理）
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
			db, tx := testutil.SetupTestDB(t)
			queries := query.New(db)

			// Usecaseを作成（R2設定は空にしてアップロードをスキップ）
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
				t.Errorf("ExecuteBatchWithTx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			// 作成された作品画像の数を確認
			if len(results) != tt.numImages {
				t.Errorf("作成された作品画像の数が期待値と異なります: got %d, want %d", len(results), tt.numImages)
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
					t.Errorf("results[%d]: ImagePathのプレフィックスが期待値と異なります: got %s, want prefix %s", i, result.ImagePath, expectedPrefix)
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
				if workID != params[0].WorkID {
					t.Errorf("work_idが期待値と異なります: got %d, want %d", workID, params[0].WorkID)
				}
				if userID != params[0].UserID {
					t.Errorf("user_idが期待値と異なります: got %d, want %d", userID, params[0].UserID)
				}

				// image_dataのJSON形式を確認
				var shrineData seed.ShrineImageData
				if err := json.Unmarshal([]byte(imageData), &shrineData); err != nil {
					t.Fatalf("image_dataのJSONパースエラー: %v", err)
				}

				// Shrine形式のフィールドを確認
				if shrineData.Master.Storage != "store" {
					t.Errorf("storage が期待値と異なります: got %s, want store", shrineData.Master.Storage)
				}
				if shrineData.Master.Metadata.MimeType != "image/png" {
					t.Errorf("mime_type が期待値と異なります: got %s, want image/png", shrineData.Master.Metadata.MimeType)
				}
				if shrineData.Master.Metadata.Width != seed.WorkImageWidth {
					t.Errorf("width が期待値と異なります: got %d, want %d", shrineData.Master.Metadata.Width, seed.WorkImageWidth)
				}
				if shrineData.Master.Metadata.Height != seed.WorkImageHeight {
					t.Errorf("height が期待値と異なります: got %d, want %d", shrineData.Master.Metadata.Height, seed.WorkImageHeight)
				}
				if shrineData.Master.Metadata.Size <= 0 {
					t.Errorf("size が0以下です: got %d", shrineData.Master.Metadata.Size)
				}
			}
		})
	}
}

// TestCreateWorkImageUsecase_ExecuteBatch はExecuteBatchメソッドのテスト（トランザクションなし、並列処理）
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
			// テストDBをセットアップ（トランザクションはコミット前に準備データを作成）
			db, tx := testutil.SetupTestDB(t)
			queries := query.New(db)

			// Usecaseを作成（R2設定は空にしてアップロードをスキップ）
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

			// トランザクションをコミット（並列処理で参照するため）
			if err := tx.Commit(); err != nil {
				t.Fatalf("トランザクションのコミットエラー: %v", err)
			}

			// ExecuteBatchを実行（トランザクションなし、並列処理パス）
			results, err := uc.ExecuteBatch(context.Background(), params, nil)

			// エラーチェック
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			// 作成された作品画像の数を確認
			if len(results) != tt.numImages {
				t.Errorf("作成された作品画像の数が期待値と異なります: got %d, want %d", len(results), tt.numImages)
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
					t.Errorf("results[%d]: ImagePathのプレフィックスが期待値と異なります: got %s, want prefix %s", i, result.ImagePath, expectedPrefix)
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
				if workID != params[i].WorkID {
					t.Errorf("results[%d]: work_idが期待値と異なります: got %d, want %d", i, workID, params[i].WorkID)
				}
				if userID != params[i].UserID {
					t.Errorf("results[%d]: user_idが期待値と異なります: got %d, want %d", i, userID, params[i].UserID)
				}

				// image_dataのJSON形式を確認
				var shrineData seed.ShrineImageData
				if err := json.Unmarshal([]byte(imageData), &shrineData); err != nil {
					t.Errorf("results[%d]: image_dataのJSONパースエラー: %v", i, err)
					continue
				}

				// Shrine形式のフィールドを確認
				if shrineData.Master.Storage != "store" {
					t.Errorf("results[%d]: storage が期待値と異なります: got %s, want store", i, shrineData.Master.Storage)
				}
				if shrineData.Master.Metadata.MimeType != "image/png" {
					t.Errorf("results[%d]: mime_type が期待値と異なります: got %s, want image/png", i, shrineData.Master.Metadata.MimeType)
				}
				if shrineData.Master.Metadata.Width != seed.WorkImageWidth {
					t.Errorf("results[%d]: width が期待値と異なります: got %d, want %d", i, shrineData.Master.Metadata.Width, seed.WorkImageWidth)
				}
				if shrineData.Master.Metadata.Height != seed.WorkImageHeight {
					t.Errorf("results[%d]: height が期待値と異なります: got %d, want %d", i, shrineData.Master.Metadata.Height, seed.WorkImageHeight)
				}
				if shrineData.Master.Metadata.Size <= 0 {
					t.Errorf("results[%d]: size が0以下です: got %d", i, shrineData.Master.Metadata.Size)
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
