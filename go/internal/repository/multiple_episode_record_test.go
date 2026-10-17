package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestMultipleEpisodeRecordRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewMultipleEpisodeRecordRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	workID := testutil.NewWorkBuilder(t, tx).Build()

	multipleEpisodeRecordID := testutil.CreateTestMultipleEpisodeRecord(t, tx, userID, workID)
	testutil.CreateTestMultipleEpisodeRecord(t, tx, userID, workID)
	otherMultipleEpisodeRecordID := testutil.CreateTestMultipleEpisodeRecord(t, tx, otherUserID, workID)

	// ユーザーの記録に付いた他人のいいねと、そのいいねでユーザーに届いた通知は削除する。
	likeID := testutil.CreateTestLike(t, tx, otherUserID, "MultipleEpisodeRecord", multipleEpisodeRecordID)
	notificationID := testutil.CreateTestNotification(t, tx, userID, otherUserID, "Like", likeID)
	// 同じIDを持つ別の種類へのいいねは残す。
	likeToStatusID := testutil.CreateTestLike(t, tx, otherUserID, "Status", multipleEpisodeRecordID)

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

	if got := testutil.CountRows(t, tx, "multiple_episode_records", "user_id = $1", int64(userID)); got != 0 {
		t.Errorf("ユーザーの複数エピソードの記録の件数 = %d、期待値 = 0", got)
	}
	rows := []struct {
		name  string
		table string
		id    int64
		want  int
	}{
		{name: "他のユーザーの複数エピソードの記録", table: "multiple_episode_records", id: otherMultipleEpisodeRecordID, want: 1},
		{name: "記録へのいいね", table: "likes", id: likeID, want: 0},
		{name: "記録へのいいねの通知", table: "notifications", id: notificationID, want: 0},
		{name: "記録以外へのいいね", table: "likes", id: likeToStatusID, want: 1},
	}
	for _, row := range rows {
		if got := testutil.CountRows(t, tx, row.table, "id = $1", row.id); got != row.want {
			t.Errorf("%sの件数 = %d、期待値 = %d", row.name, got, row.want)
		}
	}
}
