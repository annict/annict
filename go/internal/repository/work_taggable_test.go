package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestWorkTaggableRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewWorkTaggableRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	workTagID := testutil.CreateTestWorkTag(t, tx)
	testutil.CreateTestWorkTaggable(t, tx, userID, workTagID)
	testutil.CreateTestWorkTaggable(t, tx, userID, testutil.CreateTestWorkTag(t, tx))
	// 他人のタグの記録と、他人と共有するタグそのものは残す。
	otherWorkTaggableID := testutil.CreateTestWorkTaggable(t, tx, otherUserID, workTagID)

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

	if got := testutil.CountRows(t, tx, "work_taggables", "user_id = $1", int64(userID)); got != 0 {
		t.Errorf("ユーザーのタグの記録の件数 = %d、期待値 = 0", got)
	}
	if got := testutil.CountRows(t, tx, "work_taggables", "id = $1", otherWorkTaggableID); got != 1 {
		t.Errorf("他人のタグの記録の件数 = %d、期待値 = 1", got)
	}
	if got := testutil.CountRows(t, tx, "work_tags", "id = $1", workTagID); got != 1 {
		t.Errorf("タグの件数 = %d、期待値 = 1", got)
	}
}
