package repository_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

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

func TestOAuthAccessTokenRepository_FindByToken(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewOAuthAccessTokenRepository(query.New(db)).WithTx(tx)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	createdAt := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	revokedAt := time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)

	t.Run("個人用アクセストークンの項目をモデルに写す", func(t *testing.T) {
		token := uuid.New().String()
		id := testutil.NewOAuthAccessTokenBuilder(t, tx, userID).
			WithToken(token).
			WithScopes("read write").
			WithExpiresIn(7200).
			WithRevokedAt(revokedAt).
			WithCreatedAt(createdAt).
			Build()

		got, err := repo.FindByToken(ctx, token)
		if err != nil {
			t.Fatalf("FindByToken()のエラー = %v", err)
		}
		if got == nil {
			t.Fatal("FindByToken() = nil、トークンを期待")
		}
		if got.ID != id || got.ResourceOwnerID != userID {
			t.Errorf("ID = %v、ResourceOwnerID = %v、期待値 = %v・%v", got.ID, got.ResourceOwnerID, id, userID)
		}
		if !got.Personal {
			t.Error("Personal = false、期待値 = true")
		}
		if !slices.Equal(got.Scopes, []string{"read", "write"}) {
			t.Errorf("Scopes = %v、期待値 = [read write]", got.Scopes)
		}
		if got.ExpiresIn == nil || *got.ExpiresIn != 2*time.Hour {
			t.Errorf("ExpiresIn = %v、期待値 = 2h", got.ExpiresIn)
		}
		if got.RevokedAt == nil || !got.RevokedAt.Equal(revokedAt) {
			t.Errorf("RevokedAt = %v、期待値 = %v", got.RevokedAt, revokedAt)
		}
		if !got.CreatedAt.Equal(createdAt) {
			t.Errorf("CreatedAt = %v、期待値 = %v", got.CreatedAt, createdAt)
		}
	})

	t.Run("OAuthで発行したトークンは個人用アクセストークンとして扱わない", func(t *testing.T) {
		token := uuid.New().String()
		applicationID := testutil.CreateTestOAuthApplication(t, tx, userID, false)
		testutil.NewOAuthAccessTokenBuilder(t, tx, userID).WithToken(token).WithApplicationID(applicationID).Build()

		got, err := repo.FindByToken(ctx, token)
		if err != nil {
			t.Fatalf("FindByToken()のエラー = %v", err)
		}
		if got == nil || got.Personal {
			t.Errorf("FindByToken() = %+v、期待値 = 個人用でないトークン", got)
		}
		if got != nil && (got.ExpiresIn != nil || got.RevokedAt != nil) {
			t.Errorf("ExpiresIn = %v、RevokedAt = %v、期待値 = どちらもnil", got.ExpiresIn, got.RevokedAt)
		}
	})

	t.Run("無いトークンは (nil, nil)", func(t *testing.T) {
		got, err := repo.FindByToken(ctx, uuid.New().String())
		if err != nil || got != nil {
			t.Errorf("FindByToken() = %v, %v、期待値 = nil, nil", got, err)
		}
	})

	t.Run("持ち主が退会した利用者のトークンは (nil, nil)", func(t *testing.T) {
		token := uuid.New().String()
		deletedUserID := testutil.NewUserBuilder(t, tx).WithDeletedAt(time.Now()).Build()
		testutil.NewOAuthAccessTokenBuilder(t, tx, deletedUserID).WithToken(token).Build()

		got, err := repo.FindByToken(ctx, token)
		if err != nil || got != nil {
			t.Errorf("FindByToken() = %v, %v、期待値 = nil, nil", got, err)
		}
	})
}
