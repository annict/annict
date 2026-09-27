package query_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/testutil"
)

// TestGetWorkByIDはGetWorkByIDメソッドのテスト
// sqlcが生成したコードが実際のDBスキーマと正しく連携するかを確認する
func TestGetWorkByID(t *testing.T) {
	// テストDBとトランザクションをセットアップ
	db, tx := testutil.SetupTx(t)

	// テストデータを作成
	workID := testutil.NewWorkBuilder(t, tx).
		WithTitle("テスト作品").
		WithSeason(2024, testutil.SeasonSpring).
		Build()

	// エピソードを追加 (関連データの確認用)
	testutil.NewEpisodeBuilder(t, tx, workID).
		WithNumber("1").
		WithTitle("第1話 始まり").
		Build()

	testutil.NewEpisodeBuilder(t, tx, workID).
		WithNumber("2").
		WithTitle("第2話 出会い").
		Build()

	// sqlcリポジトリを作成 (トランザクションを使用)
	queries := query.New(db).WithTx(tx)

	// GetWorkByIDメソッドをテスト
	work, err := queries.GetWorkByID(context.Background(), int64(workID))
	if err != nil {
		t.Fatalf("IDによる作品の取得エラー = %v", err)
	}

	// 基本的なアサーション
	if work.Title != "テスト作品" {
		t.Errorf("work.Title = %s、期待値 = テスト作品", work.Title)
	}

	if work.SeasonYear.Valid && work.SeasonYear.Int32 != 2024 {
		t.Errorf("SeasonYear = %d、期待値 = 2024", work.SeasonYear.Int32)
	}

	// SeasonNameはenum値 (整数) として格納されている
	// winter=1, spring=2, summer=3, autumn=4
	if work.SeasonName.Valid && work.SeasonName.Int32 != 2 {
		t.Errorf("SeasonName = %d、期待値 = 2 (spring)", work.SeasonName.Int32)
	}

	// IDが正しく設定されているか確認
	if work.ID != int64(workID) {
		t.Errorf("作品IDの期待値 = %d、実測値 = %d", workID, work.ID)
	}
}

// TestGetWorkByID_NotFoundは存在しないIDでの取得テスト
func TestGetWorkByID_NotFound(t *testing.T) {
	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)

	// 存在しないIDで取得を試みる
	_, err := queries.GetWorkByID(context.Background(), 999999)
	if err == nil {
		t.Error("存在しない作品IDでエラーを期待したが、nilだった")
	}
}
