package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestForumCommentRepository_DeleteOnForumPostsByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewForumCommentRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	forumPostID := testutil.CreateTestForumPost(t, tx, userID)
	otherForumPostID := testutil.CreateTestForumPost(t, tx, otherUserID)
	// ユーザーのスレッドのコメントは、書いた人を問わず削除する。
	testutil.CreateTestForumComment(t, tx, userID, forumPostID)
	testutil.CreateTestForumComment(t, tx, otherUserID, forumPostID)
	// 他人のスレッドのコメントは、ユーザーが書いたものでも残す。
	userCommentID := testutil.CreateTestForumComment(t, tx, userID, otherForumPostID)

	for i, want := range []int64{1, 1, 0} {
		got, err := repo.DeleteOnForumPostsByUserID(ctx, userID, 1)
		if err != nil {
			t.Fatalf("%d回目のDeleteOnForumPostsByUserIDのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の削除件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	if got := testutil.CountRows(t, tx, "forum_comments", "forum_post_id = $1", forumPostID); got != 0 {
		t.Errorf("ユーザーのスレッドのコメントの件数 = %d、期待値 = 0", got)
	}
	if got := testutil.CountRows(t, tx, "forum_comments", "id = $1", userCommentID); got != 1 {
		t.Errorf("他人のスレッドのユーザーのコメントの件数 = %d、期待値 = 1", got)
	}
}

func TestForumCommentRepository_DetachFromUser(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewForumCommentRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	forumPostID := testutil.CreateTestForumPost(t, tx, otherUserID)
	commentIDs := []int64{
		testutil.CreateTestForumComment(t, tx, userID, forumPostID),
		testutil.CreateTestForumComment(t, tx, userID, forumPostID),
	}
	otherCommentID := testutil.CreateTestForumComment(t, tx, otherUserID, forumPostID)

	for i, want := range []int64{1, 1, 0} {
		got, err := repo.DetachFromUser(ctx, userID, 1)
		if err != nil {
			t.Fatalf("%d回目のDetachFromUserのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	// コメントは本文ごと残し、投稿者との紐づきだけを外す。
	for _, id := range commentIDs {
		var commenterID sql.NullInt64
		var body string
		if err := tx.QueryRow(`SELECT user_id, body FROM forum_comments WHERE id = $1`, id).Scan(&commenterID, &body); err != nil {
			t.Fatalf("コメントの取得に失敗: %v", err)
		}
		if commenterID.Valid {
			t.Errorf("コメントのuser_id = %d、期待値 = NULL", commenterID.Int64)
		}
		if body == "" {
			t.Error("コメントの本文が空になった")
		}
	}
	if got := testutil.CountRows(t, tx, "forum_comments", "id = $1 AND user_id = $2", otherCommentID, int64(otherUserID)); got != 1 {
		t.Errorf("他人のコメントの件数 = %d、期待値 = 1", got)
	}
}
