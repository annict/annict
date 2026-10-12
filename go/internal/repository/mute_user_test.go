package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestMuteUserRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewMuteUserRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	thirdUserID := testutil.NewUserBuilder(t, tx).Build()
	testutil.CreateTestMuteUser(t, tx, userID, otherUserID)
	testutil.CreateTestMuteUser(t, tx, userID, thirdUserID)
	// ユーザーがされたミュートは残す。
	mutedID := testutil.CreateTestMuteUser(t, tx, otherUserID, userID)

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

	if got := testutil.CountRows(t, tx, "mute_users", "user_id = $1", int64(userID)); got != 0 {
		t.Errorf("ユーザーがしたミュートの件数 = %d、期待値 = 0", got)
	}
	if got := testutil.CountRows(t, tx, "mute_users", "id = $1", mutedID); got != 1 {
		t.Errorf("ユーザーがされたミュートの件数 = %d、期待値 = 1", got)
	}
}

func TestMuteUserRepository_DeleteByMutedUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewMuteUserRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	thirdUserID := testutil.NewUserBuilder(t, tx).Build()
	testutil.CreateTestMuteUser(t, tx, otherUserID, userID)
	testutil.CreateTestMuteUser(t, tx, thirdUserID, userID)
	// ユーザーがしたミュートは残す。
	muteID := testutil.CreateTestMuteUser(t, tx, userID, otherUserID)

	for i, want := range []int64{1, 1, 0} {
		got, err := repo.DeleteByMutedUserID(ctx, userID, 1)
		if err != nil {
			t.Fatalf("%d回目のDeleteByMutedUserIDのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の削除件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	if got := testutil.CountRows(t, tx, "mute_users", "muted_user_id = $1", int64(userID)); got != 0 {
		t.Errorf("ユーザーがされたミュートの件数 = %d、期待値 = 0", got)
	}
	if got := testutil.CountRows(t, tx, "mute_users", "id = $1", muteID); got != 1 {
		t.Errorf("ユーザーがしたミュートの件数 = %d、期待値 = 1", got)
	}
}
