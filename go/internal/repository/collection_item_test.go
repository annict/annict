package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestCollectionItemRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewCollectionItemRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	collectionID := testutil.CreateTestCollection(t, tx, userID)
	otherCollectionID := testutil.CreateTestCollection(t, tx, otherUserID)
	testutil.CreateTestCollectionItem(t, tx, userID, collectionID, testutil.NewWorkBuilder(t, tx).Build())
	// ユーザーのコレクションに入っている項目は、項目の作成者を問わず削除する。
	itemInCollectionID := testutil.CreateTestCollectionItem(t, tx, otherUserID, collectionID, testutil.NewWorkBuilder(t, tx).Build())
	otherItemID := testutil.CreateTestCollectionItem(t, tx, otherUserID, otherCollectionID, testutil.NewWorkBuilder(t, tx).Build())

	for i, want := range []int64{1, 1, 0} {
		got, err := repo.DeleteByUserID(ctx, userID, 1)
		if err != nil {
			t.Fatalf("%d回目のDeleteByUserIDのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の削除件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	if got := testutil.CountRows(t, tx, "collection_items", "collection_id = $1", collectionID); got != 0 {
		t.Errorf("ユーザーのコレクションの項目の件数 = %d、期待値 = 0", got)
	}
	if got := testutil.CountRows(t, tx, "collection_items", "id = $1", itemInCollectionID); got != 0 {
		t.Errorf("ユーザーのコレクションに入っている他人の項目の件数 = %d、期待値 = 0", got)
	}
	if got := testutil.CountRows(t, tx, "collection_items", "id = $1", otherItemID); got != 1 {
		t.Errorf("他人のコレクションの項目の件数 = %d、期待値 = 1", got)
	}
}
