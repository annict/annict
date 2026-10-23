package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestStatusRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewStatusRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	workID := testutil.NewWorkBuilder(t, tx).Build()

	statusID := testutil.NewStatusBuilder(t, tx, userID, workID).Build()
	testutil.NewStatusBuilder(t, tx, userID, workID).WithKind(testutil.StatusKindWatched).Build()
	otherStatusID := testutil.NewStatusBuilder(t, tx, otherUserID, workID).Build()

	// ユーザーのステータスに付いた他人のいいねと、そのいいねでユーザーに届いた通知は削除する。
	likeToStatusID := testutil.CreateTestLike(t, tx, otherUserID, "Status", statusID)
	notificationOfLikeID := testutil.CreateTestNotification(t, tx, userID, otherUserID, "Like", likeToStatusID)

	// ユーザーが他人のステータスに付けたいいねと、その通知は残す (別の段階で削除する)。
	likeByUserID := testutil.CreateTestLike(t, tx, userID, "Status", otherStatusID)
	notificationToOtherID := testutil.CreateTestNotification(t, tx, otherUserID, userID, "Like", likeByUserID)

	// 同じIDを持つ別の種類へのいいねと、ステータス以外へのいいねの通知は残す。
	likeToRecordID := testutil.CreateTestLike(t, tx, otherUserID, "EpisodeRecord", statusID)
	notificationOfRecordLikeID := testutil.CreateTestNotification(t, tx, userID, otherUserID, "Like", likeToRecordID)

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

	if got := testutil.CountRows(t, tx, "statuses", "user_id = $1", int64(userID)); got != 0 {
		t.Errorf("ユーザーのステータスの件数 = %d、期待値 = 0", got)
	}

	rows := []struct {
		name  string
		table string
		id    int64
		want  int
	}{
		{name: "他のユーザーのステータス", table: "statuses", id: otherStatusID, want: 1},
		{name: "ユーザーのステータスへのいいね", table: "likes", id: likeToStatusID, want: 0},
		{name: "ユーザーのステータスへのいいねの通知", table: "notifications", id: notificationOfLikeID, want: 0},
		{name: "ユーザーが付けたいいね", table: "likes", id: likeByUserID, want: 1},
		{name: "ユーザーが付けたいいねの通知", table: "notifications", id: notificationToOtherID, want: 1},
		{name: "ステータス以外へのいいね", table: "likes", id: likeToRecordID, want: 1},
		{name: "ステータス以外へのいいねの通知", table: "notifications", id: notificationOfRecordLikeID, want: 1},
	}
	for _, row := range rows {
		if got := testutil.CountRows(t, tx, row.table, "id = $1", row.id); got != row.want {
			t.Errorf("%sの件数 = %d、期待値 = %d", row.name, got, row.want)
		}
	}
}

func TestStatusRepository_DetachDeletedOAuthApplicationsByOwnerID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewStatusRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	workID := testutil.NewWorkBuilder(t, tx).Build()
	deletedApplicationID := testutil.CreateTestOAuthApplication(t, tx, userID, true)
	activeApplicationID := testutil.CreateTestOAuthApplication(t, tx, userID, false)
	create := func(applicationID int64) int64 {
		id := testutil.NewStatusBuilder(t, tx, otherUserID, workID).Build()
		if _, err := tx.Exec("UPDATE statuses SET oauth_application_id = $1 WHERE id = $2", applicationID, id); err != nil {
			t.Fatalf("ステータスの投稿元のアプリケーションの設定に失敗: %v", err)
		}
		return id
	}
	detachedIDs := []int64{create(deletedApplicationID), create(deletedApplicationID)}
	// 論理削除されていないアプリケーションとの紐づきは残す。
	keptID := create(activeApplicationID)

	for i, want := range []int64{1, 1, 0} {
		got, err := repo.DetachDeletedOAuthApplicationsByOwnerID(ctx, userID, 1)
		if err != nil {
			t.Fatalf("%d回目のDetachDeletedOAuthApplicationsByOwnerIDのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	for _, id := range detachedIDs {
		if got := testutil.CountRows(t, tx, "statuses", "id = $1 AND oauth_application_id IS NULL", id); got != 1 {
			t.Errorf("論理削除済みのアプリケーションから投稿されたステータスの、紐づきを外した件数 = %d、期待値 = 1", got)
		}
	}
	if got := testutil.CountRows(t, tx, "statuses", "id = $1 AND oauth_application_id = $2", keptID, activeApplicationID); got != 1 {
		t.Errorf("有効なアプリケーションから投稿されたステータスの件数 = %d、期待値 = 1", got)
	}
}
