package testutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/lib/pq"
)

func TestBatchBuildWorks(t *testing.T) {
	t.Parallel()

	_, tx := SetupTx(t)

	ctx := context.Background()

	// 進捗コールバックのテスト
	var callbackCalled int
	callback := func(current, total int) {
		callbackCalled++
		if current > total {
			t.Errorf("進捗が不正: current=%d, total=%d", current, total)
		}
	}

	// 10件の作品データをバッチ作成
	count := 10
	ids, err := BatchBuildWorks(ctx, tx, count, callback)
	if err != nil {
		t.Fatalf("BatchBuildWorksのエラー = %v", err)
	}

	// IDが正しく返されることを確認
	if len(ids) != count {
		t.Errorf("IDの件数 = %d、期待値 = %d", len(ids), count)
	}

	// 進捗コールバックが正しく呼ばれたことを確認
	if callbackCalled != count {
		t.Errorf("コールバックの呼び出し回数 = %d、期待値 = %d", callbackCalled, count)
	}

	// 永続化されたことを確認する。ここで作成したidのみを数える。`make test` は
	// パッケージ間で共有DBをリセットせずに `go test ./...` を実行するため、他パッケージの
	// usecaseテストがGetTestDBでworksをコミットし、限定しないCOUNT(*) はもはやcountと
	// 一致しなくなる。
	workIDs := make([]int64, len(ids))
	for i, id := range ids {
		workIDs[i] = int64(id)
	}
	var actualCount int
	err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM works WHERE id = ANY($1::bigint[])", pq.Array(workIDs)).Scan(&actualCount)
	if err != nil {
		t.Fatalf("作品の件数の取得エラー = %v", err)
	}
	if actualCount != count {
		t.Errorf("DBの作品の件数 = %d、期待値 = %d", actualCount, count)
	}

	// 各作品が正しく作成されたことを確認
	for _, id := range ids {
		var title string
		err = tx.QueryRowContext(ctx, "SELECT title FROM works WHERE id = $1", int64(id)).Scan(&title)
		if err != nil {
			t.Fatalf("作品%dの取得エラー = %v", id, err)
		}
		// タイトルが設定されていることを確認
		if title == "" {
			t.Errorf("作品%dのtitleが空だった", id)
		}
	}
}

func TestBatchBuildUsers(t *testing.T) {
	t.Parallel()

	_, tx := SetupTx(t)

	ctx := context.Background()

	// 進捗コールバックなしでテスト
	count := 5
	ids, err := BatchBuildUsers(ctx, tx, count, nil)
	if err != nil {
		t.Fatalf("BatchBuildUsersのエラー = %v", err)
	}

	// IDが正しく返されることを確認
	if len(ids) != count {
		t.Errorf("IDの件数 = %d、期待値 = %d", len(ids), count)
	}

	// 各ユーザーが正しく作成されたことを確認
	for i, id := range ids {
		var username string
		err = tx.QueryRowContext(ctx, "SELECT username FROM users WHERE id = $1", id).Scan(&username)
		if err != nil {
			t.Fatalf("ユーザー%dの取得エラー = %v", id, err)
		}
		// ユーザー名が設定されていることを確認
		expectedUsername := fmt.Sprintf("user_%d", i+1)
		if username != expectedUsername {
			t.Errorf("ユーザー%dのusername = %s、期待値 = %s", id, username, expectedUsername)
		}
	}
}

func TestBatchBuildEpisodes(t *testing.T) {
	t.Parallel()

	_, tx := SetupTx(t)

	ctx := context.Background()

	// テスト用の作品を作成
	workID := NewWorkBuilder(t, tx).WithTitle("テストアニメ").Build()

	// 12話分のエピソードをバッチ作成
	count := 12
	ids, err := BatchBuildEpisodes(ctx, tx, workID, count, nil)
	if err != nil {
		t.Fatalf("BatchBuildEpisodesのエラー = %v", err)
	}

	// IDが正しく返されることを確認
	if len(ids) != count {
		t.Errorf("IDの件数 = %d、期待値 = %d", len(ids), count)
	}

	// データベースに正しく作成されたことを確認
	var actualCount int
	err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM episodes WHERE work_id = $1", int64(workID)).Scan(&actualCount)
	if err != nil {
		t.Fatalf("エピソードの件数の取得エラー = %v", err)
	}
	if actualCount != count {
		t.Errorf("DBのエピソードの件数 = %d、期待値 = %d", actualCount, count)
	}
}

func TestBatchBuildWithContext(t *testing.T) {
	t.Parallel()

	_, tx := SetupTx(t)

	// コンテキストキャンセルのテスト
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // すぐにキャンセル

	// キャンセルされたコンテキストでは失敗するはず
	_, err := BatchBuildWorks(ctx, tx, 1, nil)
	if err == nil {
		t.Error("キャンセル済みのcontextでエラーを期待したが、nilだった")
	}
}
