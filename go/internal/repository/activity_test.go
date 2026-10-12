package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestActivityRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewActivityRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	workID := testutil.NewWorkBuilder(t, tx).Build()
	testutil.NewActivityBuilder(t, tx, userID, workID).Build()
	testutil.NewActivityBuilder(t, tx, userID, workID).Build()
	otherActivityID := testutil.NewActivityBuilder(t, tx, otherUserID, workID).Build()

	// 上限を超える分は残し、次の呼び出しで削除する。
	for i, want := range []int64{1, 1, 0} {
		got, err := repo.DeleteByUserID(ctx, userID, 1)
		if err != nil {
			t.Fatalf("%d回目のDeleteByUserIDのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の削除件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	if got := testutil.CountRows(t, tx, "activities", "user_id = $1", int64(userID)); got != 0 {
		t.Errorf("ユーザーのアクティビティの件数 = %d、期待値 = 0", got)
	}
	if got := testutil.CountRows(t, tx, "activities", "id = $1", otherActivityID); got != 1 {
		t.Errorf("他のユーザーのアクティビティの件数 = %d、期待値 = 1", got)
	}
}
