package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestOAuthAccessTokenRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewOAuthAccessTokenRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherUserID := testutil.NewUserBuilder(t, tx).Build()
	activeApplicationID := testutil.CreateTestOAuthApplication(t, tx, userID, false)
	deletedApplicationID := testutil.CreateTestOAuthApplication(t, tx, userID, true)
	otherApplicationID := testutil.CreateTestOAuthApplication(t, tx, otherUserID, false)

	rows := []struct {
		name string
		id   int64
		want int
	}{
		{name: "ユーザーへのアクセストークン", id: testutil.CreateTestOAuthAccessToken(t, tx, userID, otherApplicationID), want: 0},
		{name: "ユーザーの論理削除済みのアプリケーションに残る他人へのアクセストークン", id: testutil.CreateTestOAuthAccessToken(t, tx, otherUserID, deletedApplicationID), want: 0},
		{name: "ユーザーの有効なアプリケーションでの他人へのアクセストークン", id: testutil.CreateTestOAuthAccessToken(t, tx, otherUserID, activeApplicationID), want: 1},
		{name: "他人のアプリケーションでの他人へのアクセストークン", id: testutil.CreateTestOAuthAccessToken(t, tx, otherUserID, otherApplicationID), want: 1},
	}

	for i, want := range []int64{1, 1, 0} {
		got, err := repo.DeleteByUserID(ctx, userID, 1)
		if err != nil {
			t.Fatalf("%d回目のDeleteByUserIDのエラー = %v", i+1, err)
		}
		if got != want {
			t.Errorf("%d回目の削除件数 = %d、期待値 = %d", i+1, got, want)
		}
	}

	for _, row := range rows {
		if got := testutil.CountRows(t, tx, "oauth_access_tokens", "id = $1", row.id); got != row.want {
			t.Errorf("%sの件数 = %d、期待値 = %d", row.name, got, row.want)
		}
	}
}
