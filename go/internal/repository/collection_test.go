package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestCollectionRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewCollectionRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	collectionID := testutil.CreateTestCollection(t, tx, userID)
	testutil.CreateTestCollection(t, tx, userID)
	otherCollectionID := testutil.CreateTestCollection(t, tx, otherUserID)
	// ユーザーのコレクションに付いた他人のいいねは一緒に削除し、他人のコレクションへのいいねは残す。
	likeID := testutil.CreateTestLike(t, tx, otherUserID, "Collection", collectionID)
	otherLikeID := testutil.CreateTestLike(t, tx, userID, "Collection", otherCollectionID)

	for i, want := range []int64{1, 1, 0} {
		got, err := repo.DeleteByUserID(ctx, userID, 1)
		if err != nil {
			t.Fatalf("%d回目のDeleteByUserIDのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の削除件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	for _, c := range []struct {
		name  string
		table string
		where string
		arg   int64
		want  int
	}{
		{name: "ユーザーのコレクション", table: "collections", where: "user_id = $1", arg: int64(userID), want: 0},
		{name: "他人のコレクション", table: "collections", where: "id = $1", arg: otherCollectionID, want: 1},
		{name: "ユーザーのコレクションへのいいね", table: "likes", where: "id = $1", arg: likeID, want: 0},
		{name: "他人のコレクションへのいいね", table: "likes", where: "id = $1", arg: otherLikeID, want: 1},
	} {
		if got := testutil.CountRows(t, tx, c.table, c.where, c.arg); got != c.want {
			t.Errorf("%sの件数 = %d、期待値 = %d", c.name, got, c.want)
		}
	}
}
