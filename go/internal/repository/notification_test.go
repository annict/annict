package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestNotificationRepository_DeleteByActionUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewNotificationRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	unreadRecipientID := testutil.NewUserBuilder(t, tx).Build()
	readRecipientID := testutil.NewUserBuilder(t, tx).Build()
	for _, id := range []model.UserID{userID, unreadRecipientID, readRecipientID} {
		if _, err := tx.ExecContext(ctx, "UPDATE users SET notifications_count = 5 WHERE id = $1", int64(id)); err != nil {
			t.Fatalf("未読数の設定に失敗: %v", err)
		}
	}

	// ユーザーが起こした通知は、宛先や既読かどうかを問わず削除する。
	deletedIDs := []int64{
		testutil.CreateTestNotification(t, tx, unreadRecipientID, userID, "Like", 1),
		testutil.CreateTestNotification(t, tx, unreadRecipientID, userID, "Comment", 2),
		testutil.CreateTestNotification(t, tx, readRecipientID, userID, "Follow", 3),
		// 宛先がユーザー自身の通知は、ユーザー行ごと消えるため未読数を減算しない。
		testutil.CreateTestNotification(t, tx, userID, userID, "Like", 4),
	}
	if _, err := tx.ExecContext(ctx, "UPDATE notifications SET read = true WHERE id = $1", deletedIDs[2]); err != nil {
		t.Fatalf("通知の既読化に失敗: %v", err)
	}

	// 他人が起こした通知は、ユーザー宛てでも残す (別の段階で削除する)。
	keptIDs := []int64{
		testutil.CreateTestNotification(t, tx, userID, unreadRecipientID, "Like", 5),
		testutil.CreateTestNotification(t, tx, readRecipientID, unreadRecipientID, "Like", 6),
	}

	// 上限を超える分は残し、次の呼び出しで削除する。
	for i, want := range []int64{3, 1} {
		got, err := repo.DeleteByActionUserID(ctx, userID, 3)
		if err != nil {
			t.Fatalf("%d回目のDeleteByActionUserIDのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の削除件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	for _, id := range deletedIDs {
		if got := testutil.CountRows(t, tx, "notifications", "id = $1", id); got != 0 {
			t.Errorf("ユーザーが起こした通知 %d の件数 = %d、期待値 = 0", id, got)
		}
	}
	for _, id := range keptIDs {
		if got := testutil.CountRows(t, tx, "notifications", "id = $1", id); got != 1 {
			t.Errorf("他人が起こした通知 %d の件数 = %d、期待値 = 1", id, got)
		}
	}

	for id, want := range map[model.UserID]int32{
		userID:            5,
		unreadRecipientID: 3,
		readRecipientID:   5,
	} {
		var notificationsCount int32
		if err := tx.QueryRowContext(ctx, "SELECT notifications_count FROM users WHERE id = $1", int64(id)).Scan(&notificationsCount); err != nil {
			t.Fatalf("未読数の取得に失敗: %v", err)
		}
		if notificationsCount != want {
			t.Errorf("ユーザー %d の未読数 = %d、期待値 = %d", id, notificationsCount, want)
		}
	}
}

func TestNotificationRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewNotificationRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()

	deletedIDs := []int64{
		testutil.CreateTestNotification(t, tx, userID, otherUserID, "Like", 1),
		testutil.CreateTestNotification(t, tx, userID, otherUserID, "Follow", 2),
	}
	// ユーザーが起こした他人宛ての通知は残す (別の段階で削除する)。
	keptID := testutil.CreateTestNotification(t, tx, otherUserID, userID, "Like", 3)

	for i, want := range []int64{1, 1, 0} {
		got, err := repo.DeleteByUserID(ctx, userID, 1)
		if err != nil {
			t.Fatalf("%d回目のDeleteByUserIDのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の削除件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	for _, id := range deletedIDs {
		if got := testutil.CountRows(t, tx, "notifications", "id = $1", id); got != 0 {
			t.Errorf("ユーザー宛ての通知 %d の件数 = %d、期待値 = 0", id, got)
		}
	}
	if got := testutil.CountRows(t, tx, "notifications", "id = $1", keptID); got != 1 {
		t.Errorf("ユーザーが起こした他人宛ての通知の件数 = %d、期待値 = 1", got)
	}
}
