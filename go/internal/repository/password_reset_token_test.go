package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestPasswordResetTokenRepository_Create はトークンを正常に作成できることをテスト
func TestPasswordResetTokenRepository_Create(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewPasswordResetTokenRepository(queries)

	userID := testutil.NewUserBuilder(t, tx).
		WithEmail("reset-create@example.com").
		Build()

	tokenDigest := "abc123digest"
	expiresAt := time.Now().Add(1 * time.Hour)

	token, err := repo.Create(context.Background(), userID, tokenDigest, expiresAt)
	if err != nil {
		t.Fatalf("Createに失敗: %v", err)
	}

	if token.UserID != userID {
		t.Errorf("UserIDが一致しません: got %v, want %v", token.UserID, userID)
	}

	if token.TokenDigest != tokenDigest {
		t.Errorf("TokenDigestが一致しません: got %v, want %v", token.TokenDigest, tokenDigest)
	}

	if token.ID == 0 {
		t.Error("IDが0です")
	}
}

// TestPasswordResetTokenRepository_DeleteUnusedByUserID は未使用トークンが削除されることをテスト
func TestPasswordResetTokenRepository_DeleteUnusedByUserID(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewPasswordResetTokenRepository(queries)

	userID := testutil.NewUserBuilder(t, tx).
		WithEmail("reset-delete@example.com").
		Build()

	// 未使用トークンを2つ作成
	expiresAt := time.Now().Add(1 * time.Hour)
	_, err := repo.Create(context.Background(), userID, "digest1", expiresAt)
	if err != nil {
		t.Fatalf("トークン1の作成に失敗: %v", err)
	}
	_, err = repo.Create(context.Background(), userID, "digest2", expiresAt)
	if err != nil {
		t.Fatalf("トークン2の作成に失敗: %v", err)
	}

	// 未使用トークンを削除
	err = repo.DeleteUnusedByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("DeleteUnusedByUserIDに失敗: %v", err)
	}

	// トークンが削除されたことを確認
	tokens, err := queries.GetPasswordResetTokensByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("トークン一覧の取得に失敗: %v", err)
	}

	if len(tokens) != 0 {
		t.Errorf("未使用トークンが残っています: got %d, want 0", len(tokens))
	}
}

// TestPasswordResetTokenRepository_DeleteUnusedByUserID_KeepsUsedTokens は使用済みトークンが残ることをテスト
func TestPasswordResetTokenRepository_DeleteUnusedByUserID_KeepsUsedTokens(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewPasswordResetTokenRepository(queries)

	userID := testutil.NewUserBuilder(t, tx).
		WithEmail("reset-keep@example.com").
		Build()

	expiresAt := time.Now().Add(1 * time.Hour)

	// 未使用トークンを作成
	_, err := repo.Create(context.Background(), userID, "unused_digest", expiresAt)
	if err != nil {
		t.Fatalf("未使用トークンの作成に失敗: %v", err)
	}

	// 使用済みトークンを作成してマーク
	usedToken, err := repo.Create(context.Background(), userID, "used_digest", expiresAt)
	if err != nil {
		t.Fatalf("使用済みトークンの作成に失敗: %v", err)
	}
	err = queries.MarkPasswordResetTokenAsUsed(context.Background(), usedToken.ID)
	if err != nil {
		t.Fatalf("トークンの使用済みマークに失敗: %v", err)
	}

	// 未使用トークンを削除
	err = repo.DeleteUnusedByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("DeleteUnusedByUserIDに失敗: %v", err)
	}

	// 使用済みトークンのみ残っていることを確認
	tokens, err := queries.GetPasswordResetTokensByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("トークン一覧の取得に失敗: %v", err)
	}

	if len(tokens) != 1 {
		t.Fatalf("トークン数が正しくありません: got %d, want 1", len(tokens))
	}

	if tokens[0].TokenDigest != "used_digest" {
		t.Errorf("残っているトークンが正しくありません: got %v, want used_digest", tokens[0].TokenDigest)
	}
}

// TestPasswordResetTokenRepository_WithTx はWithTxで取得したRepositoryがトランザクション内で動作することをテスト
func TestPasswordResetTokenRepository_WithTx(t *testing.T) {
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(db)
	repo := repository.NewPasswordResetTokenRepository(queries)

	repoWithTx := repo.WithTx(tx)

	userID := testutil.NewUserBuilder(t, tx).
		WithEmail("reset-withtx@example.com").
		Build()

	expiresAt := time.Now().Add(1 * time.Hour)

	// トランザクション内でトークンを作成
	token, err := repoWithTx.Create(context.Background(), userID, "withtx_digest", expiresAt)
	if err != nil {
		t.Fatalf("WithTxで取得したRepositoryでCreateに失敗: %v", err)
	}

	if token.UserID != userID {
		t.Errorf("UserIDが一致しません: got %v, want %v", token.UserID, userID)
	}

	// トランザクション内でGetByDigestできることを確認
	fetched, err := repoWithTx.GetByDigest(context.Background(), "withtx_digest")
	if err != nil {
		t.Fatalf("WithTxで取得したRepositoryでGetByDigestに失敗: %v", err)
	}

	if fetched.ID != token.ID {
		t.Errorf("取得したトークンIDが一致しません: got %v, want %v", fetched.ID, token.ID)
	}
}
