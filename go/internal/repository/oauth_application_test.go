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

func TestOAuthApplicationRepository_DeleteDeletedByOwnerID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewOAuthApplicationRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	testutil.CreateTestOAuthApplication(t, tx, userID, true)
	testutil.CreateTestOAuthApplication(t, tx, userID, true)
	// 論理削除されていないアプリケーションと、他人のアプリケーションは残す。
	activeApplicationID := testutil.CreateTestOAuthApplication(t, tx, userID, false)
	otherApplicationID := testutil.CreateTestOAuthApplication(t, tx, otherUserID, true)

	for i, want := range []int64{1, 1, 0} {
		got, err := repo.DeleteDeletedByOwnerID(ctx, userID, 1)
		if err != nil {
			t.Fatalf("%d回目のDeleteDeletedByOwnerIDのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の削除件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	if got := testutil.CountRows(t, tx, "oauth_applications", "owner_id = $1 AND deleted_at IS NOT NULL", int64(userID)); got != 0 {
		t.Errorf("ユーザーの論理削除済みのアプリケーションの件数 = %d、期待値 = 0", got)
	}
	if got := testutil.CountRows(t, tx, "oauth_applications", "id = $1", activeApplicationID); got != 1 {
		t.Errorf("ユーザーの有効なアプリケーションの件数 = %d、期待値 = 1", got)
	}
	if got := testutil.CountRows(t, tx, "oauth_applications", "id = $1", otherApplicationID); got != 1 {
		t.Errorf("他人のアプリケーションの件数 = %d、期待値 = 1", got)
	}
}

func TestOAuthApplicationRepository_ExistsKeptByOwnerID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prepare func(t *testing.T, tx *sql.Tx, userID, otherUserID model.UserID)
		want    bool
	}{
		{
			name:    "アプリケーションが無い",
			prepare: func(t *testing.T, tx *sql.Tx, userID, otherUserID model.UserID) {},
			want:    false,
		},
		{
			name: "論理削除されていないアプリケーションがある",
			prepare: func(t *testing.T, tx *sql.Tx, userID, otherUserID model.UserID) {
				testutil.CreateTestOAuthApplication(t, tx, userID, false)
			},
			want: true,
		},
		{
			name: "論理削除済みのアプリケーションだけがある",
			prepare: func(t *testing.T, tx *sql.Tx, userID, otherUserID model.UserID) {
				testutil.CreateTestOAuthApplication(t, tx, userID, true)
			},
			want: false,
		},
		{
			name: "他人のアプリケーションだけがある",
			prepare: func(t *testing.T, tx *sql.Tx, userID, otherUserID model.UserID) {
				testutil.CreateTestOAuthApplication(t, tx, otherUserID, false)
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, tx := testutil.SetupTx(t)
			repo := repository.NewOAuthApplicationRepository(query.New(db)).WithTx(tx)

			userID := testutil.NewUserBuilder(t, tx).Build()
			otherUserID := testutil.NewUserBuilder(t, tx).Build()
			tt.prepare(t, tx, userID, otherUserID)

			got, err := repo.ExistsKeptByOwnerID(context.Background(), userID)
			if err != nil {
				t.Fatalf("ExistsKeptByOwnerIDのエラー = %v", err)
			}
			if got != tt.want {
				t.Errorf("ExistsKeptByOwnerID = %v、期待値 = %v", got, tt.want)
			}
		})
	}
}
