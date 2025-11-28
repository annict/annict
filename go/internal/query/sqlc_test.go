package query_test

import (
	"context"
	"testing"

	"github.com/annict/annict/internal/query"
	"github.com/annict/annict/internal/testutil"
)

// TestGetWorkByID はGetWorkByIDメソッドのテスト
// sqlcが生成したコードが実際のDBスキーマと正しく連携するかを確認する
func TestGetWorkByID(t *testing.T) {
	// テストDBとトランザクションをセットアップ
	db, tx := testutil.SetupTestDB(t)

	// テストデータを作成
	workID := testutil.NewWorkBuilder(t, tx).
		WithTitle("テスト作品").
		WithSeason(2024, testutil.SeasonSpring).
		Build()

	// エピソードを追加（関連データの確認用）
	testutil.NewEpisodeBuilder(t, tx, workID).
		WithNumber("1").
		WithTitle("第1話 始まり").
		Build()

	testutil.NewEpisodeBuilder(t, tx, workID).
		WithNumber("2").
		WithTitle("第2話 出会い").
		Build()

	// sqlcリポジトリを作成（トランザクションを使用）
	queries := query.New(db).WithTx(tx)

	// GetWorkByIDメソッドをテスト
	work, err := queries.GetWorkByID(context.Background(), workID)
	if err != nil {
		t.Fatalf("Failed to get work by ID: %v", err)
	}

	// 基本的なアサーション
	if work.Title != "テスト作品" {
		t.Errorf("Expected title 'テスト作品', got %s", work.Title)
	}

	if work.SeasonYear.Valid && work.SeasonYear.Int32 != 2024 {
		t.Errorf("Expected season year 2024, got %d", work.SeasonYear.Int32)
	}

	// SeasonNameはenum値（整数）として格納されている
	// winter=1, spring=2, summer=3, autumn=4
	if work.SeasonName.Valid && work.SeasonName.Int32 != 2 {
		t.Errorf("Expected season name 2 (spring), got %d", work.SeasonName.Int32)
	}

	// IDが正しく設定されているか確認
	if work.ID != workID {
		t.Errorf("Expected work ID %d, got %d", workID, work.ID)
	}
}

// TestGetWorkByID_NotFound は存在しないIDでの取得テスト
func TestGetWorkByID_NotFound(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)

	// 存在しないIDで取得を試みる
	_, err := queries.GetWorkByID(context.Background(), 999999)
	if err == nil {
		t.Error("Expected error for non-existent work ID, but got nil")
	}
}
