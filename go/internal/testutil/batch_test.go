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
		t.Fatalf("BatchBuildWorks failed: %v", err)
	}

	// IDが正しく返されることを確認
	if len(ids) != count {
		t.Errorf("wrong number of IDs: got %d, want %d", len(ids), count)
	}

	// 進捗コールバックが正しく呼ばれたことを確認
	if callbackCalled != count {
		t.Errorf("callback not called correctly: got %d, want %d", callbackCalled, count)
	}

	// Verify they were persisted. Count only the ids created here: `make test` runs
	// `go test ./...` without resetting the shared DB between packages, so other
	// packages' usecase tests commit real works via GetTestDB and an unscoped
	// COUNT(*) would no longer equal count.
	//
	// [Ja] 永続化されたことを確認する。ここで作成した id のみを数える。`make test` は
	// パッケージ間で共有 DB をリセットせずに `go test ./...` を実行するため、他パッケージの
	// usecase テストが GetTestDB で works をコミットし、限定しない COUNT(*) はもはや count と
	// 一致しなくなる。
	workIDs := make([]int64, len(ids))
	for i, id := range ids {
		workIDs[i] = int64(id)
	}
	var actualCount int
	err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM works WHERE id = ANY($1::bigint[])", pq.Array(workIDs)).Scan(&actualCount)
	if err != nil {
		t.Fatalf("failed to count works: %v", err)
	}
	if actualCount != count {
		t.Errorf("wrong number of works in DB: got %d, want %d", actualCount, count)
	}

	// 各作品が正しく作成されたことを確認
	for _, id := range ids {
		var title string
		err = tx.QueryRowContext(ctx, "SELECT title FROM works WHERE id = $1", int64(id)).Scan(&title)
		if err != nil {
			t.Fatalf("failed to get work %d: %v", id, err)
		}
		// タイトルが設定されていることを確認
		if title == "" {
			t.Errorf("work %d has empty title", id)
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
		t.Fatalf("BatchBuildUsers failed: %v", err)
	}

	// IDが正しく返されることを確認
	if len(ids) != count {
		t.Errorf("wrong number of IDs: got %d, want %d", len(ids), count)
	}

	// 各ユーザーが正しく作成されたことを確認
	for i, id := range ids {
		var username string
		err = tx.QueryRowContext(ctx, "SELECT username FROM users WHERE id = $1", id).Scan(&username)
		if err != nil {
			t.Fatalf("failed to get user %d: %v", id, err)
		}
		// ユーザー名が設定されていることを確認
		expectedUsername := fmt.Sprintf("user_%d", i+1)
		if username != expectedUsername {
			t.Errorf("user %d has wrong username: got %s, want %s", id, username, expectedUsername)
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
		t.Fatalf("BatchBuildEpisodes failed: %v", err)
	}

	// IDが正しく返されることを確認
	if len(ids) != count {
		t.Errorf("wrong number of IDs: got %d, want %d", len(ids), count)
	}

	// データベースに正しく作成されたことを確認
	var actualCount int
	err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM episodes WHERE work_id = $1", int64(workID)).Scan(&actualCount)
	if err != nil {
		t.Fatalf("failed to count episodes: %v", err)
	}
	if actualCount != count {
		t.Errorf("wrong number of episodes in DB: got %d, want %d", actualCount, count)
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
		t.Error("expected error with cancelled context, got nil")
	}
}
