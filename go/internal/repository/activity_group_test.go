package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestActivityGroupRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewActivityGroupRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	testutil.CreateTestActivityGroup(t, tx, userID)
	testutil.CreateTestActivityGroup(t, tx, userID)
	otherActivityGroupID := testutil.CreateTestActivityGroup(t, tx, otherUserID)

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

	if got := testutil.CountRows(t, tx, "activity_groups", "user_id = $1", int64(userID)); got != 0 {
		t.Errorf("ユーザーのアクティビティグループの件数 = %d、期待値 = 0", got)
	}
	if got := testutil.CountRows(t, tx, "activity_groups", "id = $1", otherActivityGroupID); got != 1 {
		t.Errorf("他のユーザーのアクティビティグループの件数 = %d、期待値 = 1", got)
	}
}
