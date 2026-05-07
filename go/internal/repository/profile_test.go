package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// createUserWithoutRelations はプロフィール等の関連レコードなしでユーザーを作成するヘルパー
func createUserWithoutRelations(t *testing.T, queries *query.Queries, email string, username string) model.UserID {
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

// TestProfileRepository_Create はプロフィールを正常に作成し、Modelとして返却されることをテスト
func TestProfileRepository_Create(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewProfileRepository(queries)

	userID := createUserWithoutRelations(t, queries, "profile-create@example.com", "profilecreateuser")

	profile, err := repo.Create(context.Background(), userID, "testuser")
	if err != nil {
		t.Fatalf("Createに失敗: %v", err)
	}

	if profile == nil {
		t.Fatal("profileがnilです")
	}
	if profile.ID == 0 {
		t.Error("ProfileIDがゼロ値です")
	}
	if profile.UserID != userID {
		t.Errorf("UserIDが一致しません: got %v, want %v", profile.UserID, userID)
	}
	if profile.Name != "testuser" {
		t.Errorf("Nameが一致しません: got %v, want %v", profile.Name, "testuser")
	}
	// Createは description を空文字列で初期化する
	if profile.Description != "" {
		t.Errorf("Descriptionが空文字列ではありません: got %v", profile.Description)
	}
	if !profile.CreatedAt.Valid {
		t.Error("CreatedAtがセットされていません")
	}
}

// TestProfileRepository_WithTx はWithTxで取得したRepositoryがトランザクション内で動作することをテスト
func TestProfileRepository_WithTx(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db)
	repo := repository.NewProfileRepository(queries)

	repoWithTx := repo.WithTx(tx)

	queriesWithTx := queries.WithTx(tx)
	userID := createUserWithoutRelations(t, queriesWithTx, "profile-withtx@example.com", "profilewithtxuser")

	profile, err := repoWithTx.Create(context.Background(), userID, "withtxuser")
	if err != nil {
		t.Fatalf("WithTxで取得したRepositoryでCreateに失敗: %v", err)
	}

	if profile.Name != "withtxuser" {
		t.Errorf("Nameが一致しません: got %v, want %v", profile.Name, "withtxuser")
	}
	if profile.UserID != userID {
		t.Errorf("UserIDが一致しません: got %v, want %v", profile.UserID, userID)
	}
}
