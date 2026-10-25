package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestLibraryEntryRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewLibraryEntryRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()

	// 見たい・見てる・見たの作品だけ視聴者数が減り、それ以外の作品は変わらない。
	tests := []struct {
		status            string
		wantWatchersCount int32
	}{
		{status: "wanna_watch", wantWatchersCount: 9},
		{status: "watching", wantWatchersCount: 9},
		{status: "watched", wantWatchersCount: 9},
		{status: "on_hold", wantWatchersCount: 10},
		{status: "stop_watching", wantWatchersCount: 10},
	}
	workIDs := make([]model.WorkID, len(tests))
	for i, tt := range tests {
		workIDs[i] = testutil.NewWorkBuilder(t, tx).WithWatchersCount(10).Build()
		testutil.NewLibraryEntryBuilder(t, tx).
			WithUserID(userID).
			WithWorkID(workIDs[i]).
			WithStatus(tt.status).
			Build()
	}
	otherLibraryEntryID := testutil.NewLibraryEntryBuilder(t, tx).
		WithUserID(otherUserID).
		WithWorkID(workIDs[0]).
		Build()

	// 上限を超える分は残し、次の呼び出しで削除する。
	for i, want := range []int64{3, 2, 0} {
		got, err := repo.DeleteByUserID(ctx, userID, 3)
		if err != nil {
			t.Fatalf("%d回目のDeleteByUserIDのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の削除件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	if got := testutil.CountRows(t, tx, "library_entries", "user_id = $1", int64(userID)); got != 0 {
		t.Errorf("ユーザーのライブラリエントリの件数 = %d、期待値 = 0", got)
	}
	if got := testutil.CountRows(t, tx, "library_entries", "id = $1", otherLibraryEntryID); got != 1 {
		t.Errorf("他のユーザーのライブラリエントリの件数 = %d、期待値 = 1", got)
	}
	for i, tt := range tests {
		var watchersCount int32
		if err := tx.QueryRowContext(ctx, "SELECT watchers_count FROM works WHERE id = $1", int64(workIDs[i])).Scan(&watchersCount); err != nil {
			t.Fatalf("作品の視聴者数の取得に失敗: %v", err)
		}
		if watchersCount != tt.wantWatchersCount {
			t.Errorf("%sの作品の視聴者数 = %d、期待値 = %d", tt.status, watchersCount, tt.wantWatchersCount)
		}
	}
}
