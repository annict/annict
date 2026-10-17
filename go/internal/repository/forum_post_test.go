package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestForumPostRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewForumPostRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	testutil.CreateTestForumPost(t, tx, userID)
	testutil.CreateTestForumPost(t, tx, userID)
	otherForumPostID := testutil.CreateTestForumPost(t, tx, otherUserID)

	for i, want := range []int64{1, 1, 0} {
		got, err := repo.DeleteByUserID(ctx, userID, 1)
		if err != nil {
			t.Fatalf("%d回目のDeleteByUserIDのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の削除件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	if got := testutil.CountRows(t, tx, "forum_posts", "user_id = $1", int64(userID)); got != 0 {
		t.Errorf("ユーザーのスレッドの件数 = %d、期待値 = 0", got)
	}
	if got := testutil.CountRows(t, tx, "forum_posts", "id = $1", otherForumPostID); got != 1 {
		t.Errorf("他人のスレッドの件数 = %d、期待値 = 1", got)
	}
}
