package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestPersonFavoriteRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewPersonFavoriteRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	targetIDs := []int64{testutil.CreateTestPerson(t, tx), testutil.CreateTestPerson(t, tx)}
	for _, targetID := range targetIDs {
		if _, err := tx.ExecContext(ctx, "UPDATE people SET favorite_users_count = 5 WHERE id = $1", targetID); err != nil {
			t.Fatalf("人物のお気に入り数の設定に失敗: %v", err)
		}
		testutil.CreateTestPersonFavorite(t, tx, userID, targetID)
	}
	// 他人のお気に入りは残す。
	otherFavoriteID := testutil.CreateTestPersonFavorite(t, tx, otherUserID, targetIDs[0])

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

	if got := testutil.CountRows(t, tx, "person_favorites", "user_id = $1", int64(userID)); got != 0 {
		t.Errorf("ユーザーのお気に入りの件数 = %d、期待値 = 0", got)
	}
	if got := testutil.CountRows(t, tx, "person_favorites", "id = $1", otherFavoriteID); got != 1 {
		t.Errorf("他人のお気に入りの件数 = %d、期待値 = 1", got)
	}
	for _, targetID := range targetIDs {
		var count int32
		if err := tx.QueryRowContext(ctx, "SELECT favorite_users_count FROM people WHERE id = $1", targetID).Scan(&count); err != nil {
			t.Fatalf("人物のお気に入り数の取得に失敗: %v", err)
		}
		if count != 4 {
			t.Errorf("人物 %d のお気に入り数 = %d、期待値 = 4", targetID, count)
		}
	}
}
