package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestLikeRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewLikeRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	workID := testutil.NewWorkBuilder(t, tx).Build()
	episodeID := testutil.NewEpisodeBuilder(t, tx, workID).Build()

	episodeRecord := testutil.NewEpisodeRecordBuilder(t, tx, otherUserID, workID, episodeID).Build()
	workRecord := testutil.NewWorkRecordBuilder(t, tx, otherUserID, workID).Build()

	// いいね先のテーブルごとに、いいね数を5にしておく。
	recipients := []struct {
		name  string
		table string
		id    int64
		// typesはユーザーが付けるいいねのいいね先の種類。
		types []string
		want  int32
	}{
		{name: "エピソードの記録", table: "episode_records", id: episodeRecord.EpisodeRecordID, types: []string{"EpisodeRecord"}, want: 4},
		// Rails版の別名 (AnimeRecord) で作られたいいねも作品の記録のいいねとして減算する。
		{name: "作品の記録", table: "work_records", id: workRecord.WorkRecordID, types: []string{"WorkRecord", "AnimeRecord"}, want: 3},
		{name: "ステータス", table: "statuses", id: testutil.NewStatusBuilder(t, tx, otherUserID, workID).Build(), types: []string{"Status"}, want: 4},
		{name: "複数エピソードの記録", table: "multiple_episode_records", id: testutil.CreateTestMultipleEpisodeRecord(t, tx, otherUserID, workID), types: []string{"MultipleEpisodeRecord"}, want: 4},
		{name: "コメント", table: "comments", id: testutil.CreateTestComment(t, tx, otherUserID, episodeRecord.EpisodeRecordID), types: []string{"Comment"}, want: 4},
		{name: "コレクション", table: "collections", id: testutil.CreateTestCollection(t, tx, otherUserID), types: []string{"Collection"}, want: 4},
	}
	var likeIDs []int64
	for _, r := range recipients {
		if _, err := tx.ExecContext(ctx, "UPDATE "+r.table+" SET likes_count = 5 WHERE id = $1", r.id); err != nil {
			t.Fatalf("%sのいいね数の設定に失敗: %v", r.name, err)
		}
		for _, recipientType := range r.types {
			likeIDs = append(likeIDs, testutil.CreateTestLike(t, tx, userID, recipientType, r.id))
		}
	}

	// 他人のいいねは残す。
	otherLikeID := testutil.CreateTestLike(t, tx, otherUserID, "EpisodeRecord", episodeRecord.EpisodeRecordID)
	// 非公開の記録に付いたいいねを消す場合も、再公開後に備えていいね数を減らす。
	if _, err := tx.ExecContext(ctx, "UPDATE work_records SET deleted_at = NOW() WHERE id = $1", workRecord.WorkRecordID); err != nil {
		t.Fatalf("作品の記録の論理削除に失敗: %v", err)
	}

	// 上限を超える分は残し、次の呼び出しで削除する。
	for i, want := range []int64{3, 3, 1} {
		got, err := repo.DeleteByUserID(ctx, userID, 3)
		if err != nil {
			t.Fatalf("%d回目のDeleteByUserIDのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の削除件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	for _, id := range likeIDs {
		if got := testutil.CountRows(t, tx, "likes", "id = $1", id); got != 0 {
			t.Errorf("ユーザーのいいね %d の件数 = %d、期待値 = 0", id, got)
		}
	}
	if got := testutil.CountRows(t, tx, "likes", "id = $1", otherLikeID); got != 1 {
		t.Errorf("他人のいいねの件数 = %d、期待値 = 1", got)
	}

	for _, r := range recipients {
		var likesCount int32
		if err := tx.QueryRowContext(ctx, "SELECT likes_count FROM "+r.table+" WHERE id = $1", r.id).Scan(&likesCount); err != nil {
			t.Fatalf("%sのいいね数の取得に失敗: %v", r.name, err)
		}
		if likesCount != r.want {
			t.Errorf("%sのいいね数 = %d、期待値 = %d", r.name, likesCount, r.want)
		}
	}
}
