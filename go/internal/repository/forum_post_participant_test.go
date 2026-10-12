package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestForumPostParticipantRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewForumPostParticipantRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	forumPostID := testutil.CreateTestForumPost(t, tx, userID)
	otherForumPostID := testutil.CreateTestForumPost(t, tx, otherUserID)
	// ユーザーの他人のスレッドへの参加と、ユーザーのスレッドへの他人の参加を削除する。
	participationID := testutil.CreateTestForumPostParticipant(t, tx, userID, otherForumPostID)
	participantID := testutil.CreateTestForumPostParticipant(t, tx, otherUserID, forumPostID)
	otherParticipantID := testutil.CreateTestForumPostParticipant(t, tx, otherUserID, otherForumPostID)

	for i, want := range []int64{1, 1, 0} {
		got, err := repo.DeleteByUserID(ctx, userID, 1)
		if err != nil {
			t.Fatalf("%d回目のDeleteByUserIDのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の削除件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	for _, c := range []struct {
		name string
		id   int64
		want int
	}{
		{name: "ユーザーの他人のスレッドへの参加", id: participationID, want: 0},
		{name: "ユーザーのスレッドへの他人の参加", id: participantID, want: 0},
		{name: "他人のスレッドへの他人の参加", id: otherParticipantID, want: 1},
	} {
		if got := testutil.CountRows(t, tx, "forum_post_participants", "id = $1", c.id); got != c.want {
			t.Errorf("%sの件数 = %d、期待値 = %d", c.name, got, c.want)
		}
	}
}
