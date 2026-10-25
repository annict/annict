package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestCommentRepository_DeleteOnEpisodeRecordsByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewCommentRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	commenterID := testutil.NewUserBuilder(t, tx).Build()
	likerID := testutil.NewUserBuilder(t, tx).Build()
	workID := testutil.NewWorkBuilder(t, tx).Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).Build()

	episodeRecord := testutil.NewEpisodeRecordBuilder(t, tx, userID, workID, episodeID).Build()
	otherEpisodeRecord := testutil.NewEpisodeRecordBuilder(t, tx, commenterID, workID, episodeID).Build()

	// ユーザーの記録に他人が書いたコメントと、そのコメントで記録の持ち主に届いた通知は削除する。
	commentID := testutil.CreateTestComment(t, tx, commenterID, episodeRecord.EpisodeRecordID)
	notificationOfCommentID := testutil.CreateTestNotification(t, tx, userID, commenterID, "Comment", commentID)
	// ユーザーが自分の記録に書いたコメントも削除する。
	ownCommentID := testutil.CreateTestComment(t, tx, userID, episodeRecord.EpisodeRecordID)

	// コメントに付いたいいねと、そのいいねでコメントを書いた人に届いた通知も削除する。
	// 未読の通知の分だけ、宛先の未読数を減算する。
	unreadLikeID := testutil.CreateTestLike(t, tx, likerID, "Comment", commentID)
	unreadNotificationID := testutil.CreateTestNotification(t, tx, commenterID, likerID, "Like", unreadLikeID)
	readLikeID := testutil.CreateTestLike(t, tx, userID, "Comment", commentID)
	readNotificationID := testutil.CreateTestNotification(t, tx, commenterID, userID, "Like", readLikeID)
	if _, err := tx.ExecContext(ctx, "UPDATE notifications SET read = true WHERE id = $1", readNotificationID); err != nil {
		t.Fatalf("通知の既読化に失敗: %v", err)
	}
	if _, err := tx.ExecContext(ctx, "UPDATE users SET notifications_count = 5 WHERE id = $1", int64(commenterID)); err != nil {
		t.Fatalf("未読数の設定に失敗: %v", err)
	}

	// ユーザーが他人の記録に書いたコメントと、そのいいね・通知は残す (別の段階で削除する)。
	commentByUserID := testutil.CreateTestComment(t, tx, userID, otherEpisodeRecord.EpisodeRecordID)
	likeToCommentByUserID := testutil.CreateTestLike(t, tx, likerID, "Comment", commentByUserID)
	notificationToUserID := testutil.CreateTestNotification(t, tx, userID, likerID, "Like", likeToCommentByUserID)

	// 上限を超える分は残し、次の呼び出しで削除する。
	for i, want := range []int64{1, 1, 0} {
		got, err := repo.DeleteOnEpisodeRecordsByUserID(ctx, userID, 1)
		if err != nil {
			t.Fatalf("%d回目のDeleteOnEpisodeRecordsByUserIDのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の削除件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	rows := []struct {
		name  string
		table string
		id    int64
		want  int
	}{
		{name: "ユーザーの記録への他人のコメント", table: "comments", id: commentID, want: 0},
		{name: "ユーザーの記録への自分のコメント", table: "comments", id: ownCommentID, want: 0},
		{name: "コメントの通知", table: "notifications", id: notificationOfCommentID, want: 0},
		{name: "コメントへのいいね (未読の通知)", table: "likes", id: unreadLikeID, want: 0},
		{name: "コメントへのいいねの未読の通知", table: "notifications", id: unreadNotificationID, want: 0},
		{name: "コメントへのいいね (既読の通知)", table: "likes", id: readLikeID, want: 0},
		{name: "コメントへのいいねの既読の通知", table: "notifications", id: readNotificationID, want: 0},
		{name: "ユーザーが他人の記録に書いたコメント", table: "comments", id: commentByUserID, want: 1},
		{name: "ユーザーが他人の記録に書いたコメントへのいいね", table: "likes", id: likeToCommentByUserID, want: 1},
		{name: "ユーザーが他人の記録に書いたコメントへのいいねの通知", table: "notifications", id: notificationToUserID, want: 1},
	}
	for _, row := range rows {
		if got := testutil.CountRows(t, tx, row.table, "id = $1", row.id); got != row.want {
			t.Errorf("%sの件数 = %d、期待値 = %d", row.name, got, row.want)
		}
	}

	var notificationsCount int32
	if err := tx.QueryRowContext(ctx, "SELECT notifications_count FROM users WHERE id = $1", int64(commenterID)).Scan(&notificationsCount); err != nil {
		t.Fatalf("未読数の取得に失敗: %v", err)
	}
	if notificationsCount != 4 {
		t.Errorf("コメントを書いた人の未読数 = %d、期待値 = 4", notificationsCount)
	}
}

func TestCommentRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewCommentRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	workID := testutil.NewWorkBuilder(t, tx).Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).Build()

	episodeRecord := testutil.NewEpisodeRecordBuilder(t, tx, otherUserID, workID, episodeID).Build()
	anotherEpisodeRecord := testutil.NewEpisodeRecordBuilder(t, tx, otherUserID, workID, episodeID).Build()
	for _, id := range []int64{episodeRecord.EpisodeRecordID, anotherEpisodeRecord.EpisodeRecordID} {
		if _, err := tx.ExecContext(ctx, "UPDATE episode_records SET comments_count = 5 WHERE id = $1", id); err != nil {
			t.Fatalf("コメント数の設定に失敗: %v", err)
		}
	}

	// ユーザーが他人の記録に書いたコメントと、そのコメントに付いたいいねを削除する。
	commentIDs := []int64{
		testutil.CreateTestComment(t, tx, userID, episodeRecord.EpisodeRecordID),
		testutil.CreateTestComment(t, tx, userID, episodeRecord.EpisodeRecordID),
		testutil.CreateTestComment(t, tx, userID, anotherEpisodeRecord.EpisodeRecordID),
	}
	likeID := testutil.CreateTestLike(t, tx, otherUserID, "Comment", commentIDs[0])

	// 他人のコメントとそのいいねは残す。
	otherCommentID := testutil.CreateTestComment(t, tx, otherUserID, episodeRecord.EpisodeRecordID)
	otherLikeID := testutil.CreateTestLike(t, tx, userID, "Comment", otherCommentID)
	// 非公開の記録に付いたコメントを消す場合も、再公開後に備えてコメント数を減らす。
	if _, err := tx.ExecContext(ctx, "UPDATE episode_records SET deleted_at = NOW() WHERE id = $1", anotherEpisodeRecord.EpisodeRecordID); err != nil {
		t.Fatalf("エピソードの記録の論理削除に失敗: %v", err)
	}

	// 上限を超える分は残し、次の呼び出しで削除する。
	for i, want := range []int64{2, 1} {
		got, err := repo.DeleteByUserID(ctx, userID, 2)
		if err != nil {
			t.Fatalf("%d回目のDeleteByUserIDのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の削除件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	rows := []struct {
		name  string
		table string
		id    int64
		want  int
	}{
		{name: "ユーザーのコメント1", table: "comments", id: commentIDs[0], want: 0},
		{name: "ユーザーのコメント2", table: "comments", id: commentIDs[1], want: 0},
		{name: "ユーザーのコメント3", table: "comments", id: commentIDs[2], want: 0},
		{name: "ユーザーのコメントへのいいね", table: "likes", id: likeID, want: 0},
		{name: "他人のコメント", table: "comments", id: otherCommentID, want: 1},
		{name: "他人のコメントへのいいね", table: "likes", id: otherLikeID, want: 1},
	}
	for _, row := range rows {
		if got := testutil.CountRows(t, tx, row.table, "id = $1", row.id); got != row.want {
			t.Errorf("%sの件数 = %d、期待値 = %d", row.name, got, row.want)
		}
	}

	for id, want := range map[int64]int32{
		episodeRecord.EpisodeRecordID:        3,
		anotherEpisodeRecord.EpisodeRecordID: 4,
	} {
		var commentsCount int32
		if err := tx.QueryRowContext(ctx, "SELECT comments_count FROM episode_records WHERE id = $1", id).Scan(&commentsCount); err != nil {
			t.Fatalf("コメント数の取得に失敗: %v", err)
		}
		if commentsCount != want {
			t.Errorf("エピソードの記録 %d のコメント数 = %d、期待値 = %d", id, commentsCount, want)
		}
	}
}
