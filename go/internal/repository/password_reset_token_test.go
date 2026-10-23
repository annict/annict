package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/password_reset"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestPasswordResetTokenRepository_Createはトークンを正常に作成し、Modelとして返却されることをテスト
func TestPasswordResetTokenRepository_Create(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
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

	if token == nil {
		t.Fatal("tokenがnilです")
	}
	if token.ID == 0 {
		t.Error("PasswordResetTokenIDがゼロ値です")
	}
	if token.UserID != userID {
		t.Errorf("UserID = %v、期待値 = %v", token.UserID, userID)
	}
	if token.TokenDigest != tokenDigest {
		t.Errorf("TokenDigest = %v、期待値 = %v", token.TokenDigest, tokenDigest)
	}
	if token.UsedAt.Valid {
		t.Error("UsedAtが有効値です (未使用のはず)")
	}
	if token.CreatedAt.IsZero() {
		t.Error("CreatedAtがゼロ値です")
	}
}

// TestPasswordResetTokenRepository_GetValidByToken_DeletedUserはトークンの発行後に削除されたユーザーのトークンが無効になることをテスト
func TestPasswordResetTokenRepository_GetValidByToken_DeletedUser(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewPasswordResetTokenRepository(queries)

	userID := testutil.NewUserBuilder(t, tx).Build()

	token := "deleted-user-reset-token"
	if _, err := repo.Create(context.Background(), userID, password_reset.HashToken(token), time.Now().Add(1*time.Hour)); err != nil {
		t.Fatalf("Createに失敗: %v", err)
	}

	if _, err := repo.GetValidByToken(context.Background(), token); err != nil {
		t.Fatalf("削除前のGetValidByTokenのエラー = %v", err)
	}

	if _, err := tx.Exec(`UPDATE users SET deleted_at = NOW() WHERE id = $1`, int64(userID)); err != nil {
		t.Fatalf("ユーザーの論理削除に失敗: %v", err)
	}

	got, err := repo.GetValidByToken(context.Background(), token)
	if !errors.Is(err, repository.ErrInvalidPasswordResetToken) {
		t.Errorf("エラー = %v、期待値 = ErrInvalidPasswordResetToken", err)
	}
	if got != nil {
		t.Errorf("トークン = %+v、期待値 = nil", got)
	}
}

// TestPasswordResetTokenRepository_DeleteUnusedByUserIDは未使用トークンが削除されることをテスト
func TestPasswordResetTokenRepository_DeleteUnusedByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewPasswordResetTokenRepository(queries)

	userID := testutil.NewUserBuilder(t, tx).
		WithEmail("reset-delete@example.com").
		Build()

	// 未使用トークンを2つ作成
	expiresAt := time.Now().Add(1 * time.Hour)
	if _, err := repo.Create(context.Background(), userID, "digest1", expiresAt); err != nil {
		t.Fatalf("トークン1の作成に失敗: %v", err)
	}
	if _, err := repo.Create(context.Background(), userID, "digest2", expiresAt); err != nil {
		t.Fatalf("トークン2の作成に失敗: %v", err)
	}

	// 未使用トークンを削除
	if err := repo.DeleteUnusedByUserID(context.Background(), userID); err != nil {
		t.Fatalf("DeleteUnusedByUserIDに失敗: %v", err)
	}

	// トークンが削除されたことを確認
	tokens, err := repo.GetByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetByUserIDに失敗: %v", err)
	}

	if len(tokens) != 0 {
		t.Errorf("未使用トークンの件数 = %d、期待値 = 0", len(tokens))
	}
}

// TestPasswordResetTokenRepository_DeleteUnusedByUserID_KeepsUsedTokensは使用済みトークンが残ることをテスト
func TestPasswordResetTokenRepository_DeleteUnusedByUserID_KeepsUsedTokens(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewPasswordResetTokenRepository(queries)

	userID := testutil.NewUserBuilder(t, tx).
		WithEmail("reset-keep@example.com").
		Build()

	expiresAt := time.Now().Add(1 * time.Hour)

	// 未使用トークンを作成
	if _, err := repo.Create(context.Background(), userID, "unused_digest", expiresAt); err != nil {
		t.Fatalf("未使用トークンの作成に失敗: %v", err)
	}

	// 使用済みトークンを作成してマーク
	usedToken, err := repo.Create(context.Background(), userID, "used_digest", expiresAt)
	if err != nil {
		t.Fatalf("使用済みトークンの作成に失敗: %v", err)
	}
	if err := repo.MarkAsUsed(context.Background(), usedToken.ID); err != nil {
		t.Fatalf("トークンの使用済みマークに失敗: %v", err)
	}

	// 未使用トークンを削除
	if err := repo.DeleteUnusedByUserID(context.Background(), userID); err != nil {
		t.Fatalf("DeleteUnusedByUserIDに失敗: %v", err)
	}

	// 使用済みトークンのみ残っていることを確認
	tokens, err := repo.GetByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetByUserIDに失敗: %v", err)
	}

	if len(tokens) != 1 {
		t.Fatalf("トークン数 = %d、期待値 = 1", len(tokens))
	}

	if tokens[0].TokenDigest != "used_digest" {
		t.Errorf("残っているトークン = %v、期待値 = used_digest", tokens[0].TokenDigest)
	}
	if !tokens[0].UsedAt.Valid {
		t.Error("使用済みトークンのUsedAtが有効値ではありません")
	}
}

// TestPasswordResetTokenRepository_WithTxはWithTxで取得したRepositoryがトランザクション内で動作することをテスト
func TestPasswordResetTokenRepository_WithTx(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
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
		t.Errorf("UserID = %v、期待値 = %v", token.UserID, userID)
	}

	// トランザクション内でGetByDigestできることを確認
	fetched, err := repoWithTx.GetByDigest(context.Background(), "withtx_digest")
	if err != nil {
		t.Fatalf("WithTxで取得したRepositoryでGetByDigestに失敗: %v", err)
	}

	if fetched.ID != token.ID {
		t.Errorf("取得したトークンID = %v、期待値 = %v", fetched.ID, token.ID)
	}
}
