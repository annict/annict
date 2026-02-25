package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// createUserWithoutRelations はプロフィール等の関連レコードなしでユーザーを作成するヘルパー
func createUserWithoutRelations(t *testing.T, queries *query.Queries, email string, username string) int64 {
	t.Helper()

	repo := repository.NewUserRepository(queries)
	user, err := repo.Create(context.Background(), repository.UserCreateParams{
		Username:          username,
		Email:             email,
		EncryptedPassword: "",
		Locale:            "ja",
	})
	if err != nil {
		t.Fatalf("ユーザーの作成に失敗: %v", err)
	}
	return user.ID
}

// TestProfileRepository_Create はプロフィールを正常に作成できることをテスト
func TestProfileRepository_Create(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewProfileRepository(queries)

	userID := createUserWithoutRelations(t, queries, "profile-create@example.com", "profilecreateuser")

	err := repo.Create(context.Background(), userID, "testuser")
	if err != nil {
		t.Fatalf("Createに失敗: %v", err)
	}

	// 作成されたプロフィールを確認
	profile, err := queries.GetProfileByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("プロフィールの取得に失敗: %v", err)
	}

	if profile.UserID != userID {
		t.Errorf("UserIDが一致しません: got %v, want %v", profile.UserID, userID)
	}
	if profile.Name != "testuser" {
		t.Errorf("Nameが一致しません: got %v, want %v", profile.Name, "testuser")
	}
}

// TestProfileRepository_WithTx はWithTxで取得したRepositoryがトランザクション内で動作することをテスト
func TestProfileRepository_WithTx(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db)
	repo := repository.NewProfileRepository(queries)

	repoWithTx := repo.WithTx(tx)

	queriesWithTx := queries.WithTx(tx)
	userID := createUserWithoutRelations(t, queriesWithTx, "profile-withtx@example.com", "profilewithtxuser")

	err := repoWithTx.Create(context.Background(), userID, "withtxuser")
	if err != nil {
		t.Fatalf("WithTxで取得したRepositoryでCreateに失敗: %v", err)
	}

	// トランザクション内で取得できることを確認
	profile, err := queriesWithTx.GetProfileByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("トランザクション内でプロフィールの取得に失敗: %v", err)
	}

	if profile.Name != "withtxuser" {
		t.Errorf("Nameが一致しません: got %v, want %v", profile.Name, "withtxuser")
	}
}
