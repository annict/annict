package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestWorkRecordRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewWorkRecordRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	workID := testutil.NewWorkBuilder(t, tx).Build()
	hiddenWorkID := testutil.NewWorkBuilder(t, tx).Build()
	if _, err := tx.ExecContext(ctx,
		"UPDATE works SET work_records_count = 10, work_records_with_body_count = 10 WHERE id IN ($1, $2)",
		int64(workID), int64(hiddenWorkID),
	); err != nil {
		t.Fatalf("作品の記録数の設定に失敗: %v", err)
	}

	// 本文の有無によって、本文のある記録数を減算するかが変わる。
	withBody := testutil.NewWorkRecordBuilder(t, tx, userID, workID).WithBody("感想").Build()
	withoutBody := testutil.NewWorkRecordBuilder(t, tx, userID, workID).Build()
	// 論理削除された記録は記録数に含まれていないため、減算しない。
	testutil.NewWorkRecordBuilder(t, tx, userID, hiddenWorkID).
		WithBody("感想").
		WithDeletedAt(time.Now()).
		Build()
	otherWorkRecord := testutil.NewWorkRecordBuilder(t, tx, otherUserID, workID).Build()

	// ユーザーの記録に付いた他人のいいね (Rails版の別名で作られたものを含む) と、
	// そのいいねでユーザーに届いた通知は削除する。
	likeID := testutil.CreateTestLike(t, tx, otherUserID, "WorkRecord", withBody.WorkRecordID)
	notificationID := testutil.CreateTestNotification(t, tx, userID, otherUserID, "Like", likeID)
	aliasLikeID := testutil.CreateTestLike(t, tx, otherUserID, "AnimeRecord", withoutBody.WorkRecordID)
	aliasNotificationID := testutil.CreateTestNotification(t, tx, userID, otherUserID, "Like", aliasLikeID)
	// 同じIDを持つ別の種類へのいいねは残す。
	likeToEpisodeRecordID := testutil.CreateTestLike(t, tx, otherUserID, "EpisodeRecord", withBody.WorkRecordID)

	// 上限を超える分は残し、次の呼び出しで削除する。
	for i, want := range []int64{2, 1, 0} {
		got, err := repo.DeleteByUserID(ctx, userID, 2)
		if err != nil {
			t.Fatalf("%d回目のDeleteByUserIDのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の削除件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	if got := testutil.CountRows(t, tx, "work_records", "user_id = $1", int64(userID)); got != 0 {
		t.Errorf("ユーザーの作品の記録の件数 = %d、期待値 = 0", got)
	}
	rows := []struct {
		name  string
		table string
		id    int64
		want  int
	}{
		{name: "他のユーザーの作品の記録", table: "work_records", id: otherWorkRecord.WorkRecordID, want: 1},
		{name: "記録へのいいね", table: "likes", id: likeID, want: 0},
		{name: "記録へのいいねの通知", table: "notifications", id: notificationID, want: 0},
		{name: "別名で作られた記録へのいいね", table: "likes", id: aliasLikeID, want: 0},
		{name: "別名で作られた記録へのいいねの通知", table: "notifications", id: aliasNotificationID, want: 0},
		{name: "記録以外へのいいね", table: "likes", id: likeToEpisodeRecordID, want: 1},
	}
	for _, row := range rows {
		if got := testutil.CountRows(t, tx, row.table, "id = $1", row.id); got != row.want {
			t.Errorf("%sの件数 = %d、期待値 = %d", row.name, got, row.want)
		}
	}

	tests := []struct {
		name         string
		workID       model.WorkID
		wantRecords  int32
		wantWithBody int32
	}{
		// 他のユーザーの記録は残るため、ユーザーの2件の分だけ減る。
		{name: "記録を削除した作品", workID: workID, wantRecords: 8, wantWithBody: 9},
		{name: "論理削除された記録だけを削除した作品", workID: hiddenWorkID, wantRecords: 10, wantWithBody: 10},
	}
	for _, tt := range tests {
		var recordsCount, withBodyCount int32
		if err := tx.QueryRowContext(ctx,
			"SELECT work_records_count, work_records_with_body_count FROM works WHERE id = $1",
			int64(tt.workID),
		).Scan(&recordsCount, &withBodyCount); err != nil {
			t.Fatalf("作品の記録数の取得に失敗: %v", err)
		}
		if recordsCount != tt.wantRecords {
			t.Errorf("%sの記録数 = %d、期待値 = %d", tt.name, recordsCount, tt.wantRecords)
		}
		if withBodyCount != tt.wantWithBody {
			t.Errorf("%sの本文のある記録数 = %d、期待値 = %d", tt.name, withBodyCount, tt.wantWithBody)
		}
	}
}

func TestWorkRecordRepository_DetachDeletedOAuthApplicationsByOwnerID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewWorkRecordRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	workID := testutil.NewWorkBuilder(t, tx).Build()
	deletedApplicationID := testutil.CreateTestOAuthApplication(t, tx, userID, true)
	activeApplicationID := testutil.CreateTestOAuthApplication(t, tx, userID, false)
	create := func(applicationID int64) int64 {
		id := testutil.NewWorkRecordBuilder(t, tx, otherUserID, workID).Build().WorkRecordID
		if _, err := tx.Exec("UPDATE work_records SET oauth_application_id = $1 WHERE id = $2", applicationID, id); err != nil {
			t.Fatalf("作品の記録の投稿元のアプリケーションの設定に失敗: %v", err)
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
		if got := testutil.CountRows(t, tx, "work_records", "id = $1 AND oauth_application_id IS NULL", id); got != 1 {
			t.Errorf("論理削除済みのアプリケーションから投稿された作品の記録の、紐づきを外した件数 = %d、期待値 = 1", got)
		}
	}
	if got := testutil.CountRows(t, tx, "work_records", "id = $1 AND oauth_application_id = $2", keptID, activeApplicationID); got != 1 {
		t.Errorf("有効なアプリケーションから投稿された作品の記録の件数 = %d、期待値 = 1", got)
	}
}
