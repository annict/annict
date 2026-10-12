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

func TestEpisodeRecordRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewEpisodeRecordRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	workID := testutil.NewWorkBuilder(t, tx).Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).Build()
	hiddenEpisodeID := testutil.NewEpisodeBuilder(t, tx, workID).Build()
	if _, err := tx.ExecContext(ctx,
		"UPDATE episodes SET episode_records_count = 10, episode_record_bodies_count = 10 WHERE work_id = $1",
		int64(workID),
	); err != nil {
		t.Fatalf("エピソードの記録数の設定に失敗: %v", err)
	}

	// 本文の有無によって、本文のある記録数を減算するかが変わる。
	withBody := testutil.NewEpisodeRecordBuilder(t, tx, userID, workID, episodeID).WithBody("感想").Build()
	testutil.NewEpisodeRecordBuilder(t, tx, userID, workID, episodeID).Build()
	testutil.NewEpisodeRecordBuilder(t, tx, userID, workID, episodeID).WithBody("").Build()
	// 論理削除された記録は記録数に含まれていないため、減算しない。
	testutil.NewEpisodeRecordBuilder(t, tx, userID, workID, hiddenEpisodeID).
		WithBody("感想").
		WithDeletedAt(time.Now()).
		Build()
	otherEpisodeRecord := testutil.NewEpisodeRecordBuilder(t, tx, otherUserID, workID, episodeID).Build()

	// ユーザーの記録に付いた他人のいいねと、そのいいねでユーザーに届いた通知は削除する。
	likeID := testutil.CreateTestLike(t, tx, otherUserID, "EpisodeRecord", withBody.EpisodeRecordID)
	notificationID := testutil.CreateTestNotification(t, tx, userID, otherUserID, "Like", likeID)
	// 同じIDを持つ別の種類へのいいねは残す。
	likeToStatusID := testutil.CreateTestLike(t, tx, otherUserID, "Status", withBody.EpisodeRecordID)

	// 上限を超える分は残し、次の呼び出しで削除する。
	for i, want := range []int64{3, 1, 0} {
		got, err := repo.DeleteByUserID(ctx, userID, 3)
		if err != nil {
			t.Fatalf("%d回目のDeleteByUserIDのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の削除件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	if got := testutil.CountRows(t, tx, "episode_records", "user_id = $1", int64(userID)); got != 0 {
		t.Errorf("ユーザーのエピソードの記録の件数 = %d、期待値 = 0", got)
	}
	rows := []struct {
		name  string
		table string
		id    int64
		want  int
	}{
		{name: "他のユーザーのエピソードの記録", table: "episode_records", id: otherEpisodeRecord.EpisodeRecordID, want: 1},
		{name: "記録へのいいね", table: "likes", id: likeID, want: 0},
		{name: "記録へのいいねの通知", table: "notifications", id: notificationID, want: 0},
		{name: "記録以外へのいいね", table: "likes", id: likeToStatusID, want: 1},
	}
	for _, row := range rows {
		if got := testutil.CountRows(t, tx, row.table, "id = $1", row.id); got != row.want {
			t.Errorf("%sの件数 = %d、期待値 = %d", row.name, got, row.want)
		}
	}

	tests := []struct {
		name           string
		episodeID      model.EpisodeID
		wantRecords    int32
		wantWithBodies int32
	}{
		{name: "記録を削除したエピソード", episodeID: episodeID, wantRecords: 7, wantWithBodies: 9},
		{name: "論理削除された記録だけを削除したエピソード", episodeID: hiddenEpisodeID, wantRecords: 10, wantWithBodies: 10},
	}
	for _, tt := range tests {
		var recordsCount, bodiesCount int32
		if err := tx.QueryRowContext(ctx,
			"SELECT episode_records_count, episode_record_bodies_count FROM episodes WHERE id = $1",
			int64(tt.episodeID),
		).Scan(&recordsCount, &bodiesCount); err != nil {
			t.Fatalf("エピソードの記録数の取得に失敗: %v", err)
		}
		if recordsCount != tt.wantRecords {
			t.Errorf("%sの記録数 = %d、期待値 = %d", tt.name, recordsCount, tt.wantRecords)
		}
		if bodiesCount != tt.wantWithBodies {
			t.Errorf("%sの本文のある記録数 = %d、期待値 = %d", tt.name, bodiesCount, tt.wantWithBodies)
		}
	}
}

func TestEpisodeRecordRepository_DetachDeletedOAuthApplicationsByOwnerID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewEpisodeRecordRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	workID := testutil.NewWorkBuilder(t, tx).Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).Build()
	deletedApplicationID := testutil.CreateTestOAuthApplication(t, tx, userID, true)
	activeApplicationID := testutil.CreateTestOAuthApplication(t, tx, userID, false)
	create := func(applicationID int64) int64 {
		id := testutil.NewEpisodeRecordBuilder(t, tx, otherUserID, workID, episodeID).Build().EpisodeRecordID
		if _, err := tx.Exec("UPDATE episode_records SET oauth_application_id = $1 WHERE id = $2", applicationID, id); err != nil {
			t.Fatalf("エピソードの記録の投稿元のアプリケーションの設定に失敗: %v", err)
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
		if got := testutil.CountRows(t, tx, "episode_records", "id = $1 AND oauth_application_id IS NULL", id); got != 1 {
			t.Errorf("論理削除済みのアプリケーションから投稿されたエピソードの記録の、紐づきを外した件数 = %d、期待値 = 1", got)
		}
	}
	if got := testutil.CountRows(t, tx, "episode_records", "id = $1 AND oauth_application_id = $2", keptID, activeApplicationID); got != 1 {
		t.Errorf("有効なアプリケーションから投稿されたエピソードの記録の件数 = %d、期待値 = 1", got)
	}
}
