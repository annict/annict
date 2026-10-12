package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestFollowRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewFollowRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	followingIDs := []model.UserID{
		testutil.NewUserBuilder(t, tx).Build(),
		testutil.NewUserBuilder(t, tx).Build(),
	}
	for _, followingID := range followingIDs {
		setFollowCounts(t, tx, followingID)
		testutil.CreateTestFollow(t, tx, userID, followingID)
	}
	// ユーザーへのフォローと、他人どうしのフォローは残す。
	followerFollowID := testutil.CreateTestFollow(t, tx, followingIDs[0], userID)
	otherFollowID := testutil.CreateTestFollow(t, tx, followingIDs[1], followingIDs[0])

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

	if got := testutil.CountRows(t, tx, "follows", "user_id = $1", int64(userID)); got != 0 {
		t.Errorf("ユーザーのフォローの件数 = %d、期待値 = 0", got)
	}
	for _, id := range []int64{followerFollowID, otherFollowID} {
		if got := testutil.CountRows(t, tx, "follows", "id = $1", id); got != 1 {
			t.Errorf("フォロー %d の件数 = %d、期待値 = 1", id, got)
		}
	}
	for _, followingID := range followingIDs {
		// フォロー先のフォロワー数だけを減らし、フォロー数は変えない。
		followers, following := getFollowCounts(t, tx, followingID)
		if followers != 4 || following != 5 {
			t.Errorf("フォロー先 %d の (フォロワー数, フォロー数) = (%d, %d)、期待値 = (4, 5)", followingID, followers, following)
		}
	}
}

func TestFollowRepository_DeleteByFollowingID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewFollowRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	followerIDs := []model.UserID{
		testutil.NewUserBuilder(t, tx).Build(),
		testutil.NewUserBuilder(t, tx).Build(),
	}
	for _, followerID := range followerIDs {
		setFollowCounts(t, tx, followerID)
		testutil.CreateTestFollow(t, tx, followerID, userID)
	}
	// ユーザーがしたフォローは残す。
	userFollowID := testutil.CreateTestFollow(t, tx, userID, followerIDs[0])

	for i, want := range []int64{1, 1, 0} {
		got, err := repo.DeleteByFollowingID(ctx, userID, 1)
		if err != nil {
			t.Fatalf("%d回目のDeleteByFollowingIDのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の削除件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	if got := testutil.CountRows(t, tx, "follows", "following_id = $1", int64(userID)); got != 0 {
		t.Errorf("ユーザーへのフォローの件数 = %d、期待値 = 0", got)
	}
	if got := testutil.CountRows(t, tx, "follows", "id = $1", userFollowID); got != 1 {
		t.Errorf("ユーザーがしたフォローの件数 = %d、期待値 = 1", got)
	}
	for _, followerID := range followerIDs {
		// フォローしていた人のフォロー数だけを減らし、フォロワー数は変えない。
		followers, following := getFollowCounts(t, tx, followerID)
		if followers != 5 || following != 4 {
			t.Errorf("フォローしていた人 %d の (フォロワー数, フォロー数) = (%d, %d)、期待値 = (5, 4)", followerID, followers, following)
		}
	}
}

// setFollowCountsはユーザーのフォロワー数・フォロー数をどちらも5にする。
func setFollowCounts(t *testing.T, tx *sql.Tx, userID model.UserID) {
	t.Helper()
	if _, err := tx.Exec("UPDATE users SET followers_count = 5, following_count = 5 WHERE id = $1", int64(userID)); err != nil {
		t.Fatalf("フォロワー数・フォロー数の設定に失敗: %v", err)
	}
}

// getFollowCountsはユーザーのフォロワー数とフォロー数を返す。
func getFollowCounts(t *testing.T, tx *sql.Tx, userID model.UserID) (followers, following int32) {
	t.Helper()
	if err := tx.QueryRow("SELECT followers_count, following_count FROM users WHERE id = $1", int64(userID)).Scan(&followers, &following); err != nil {
		t.Fatalf("フォロワー数・フォロー数の取得に失敗: %v", err)
	}
	return followers, following
}
