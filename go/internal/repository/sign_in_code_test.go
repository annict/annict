package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestSignInCodeRepository_Create はサインインコードを正常に作成し、Modelとして返却されることをテスト
func TestSignInCodeRepository_Create(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewSignInCodeRepository(queries)

	userID := testutil.NewUserBuilder(t, tx).
		WithEmail("signincode-create@example.com").
		Build()

	codeDigest := "bcrypt-hashed-digest"
	expiresAt := time.Now().Add(15 * time.Minute)

	code, err := repo.Create(context.Background(), repository.SignInCodeCreateParams{
		UserID:     userID,
		CodeDigest: codeDigest,
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		t.Fatalf("Createに失敗: %v", err)
	}

	if code == nil {
		t.Fatal("codeがnilです")
	}
	if code.ID == 0 {
		t.Error("SignInCodeIDがゼロ値です")
	}
	if code.UserID != userID {
		t.Errorf("UserIDが一致しません: got %v, want %v", code.UserID, userID)
	}
	if code.CodeDigest != codeDigest {
		t.Errorf("CodeDigestが一致しません: got %v, want %v", code.CodeDigest, codeDigest)
	}
	if code.Attempts != 0 {
		t.Errorf("Attemptsの初期値が0ではありません: got %v", code.Attempts)
	}
	if code.UsedAt.Valid {
		t.Error("UsedAtが有効値です（未使用のはず）")
	}
	if code.CreatedAt.IsZero() {
		t.Error("CreatedAtがゼロ値です")
	}
	if code.UpdatedAt.IsZero() {
		t.Error("UpdatedAtがゼロ値です")
	}
}

// TestSignInCodeRepository_GetValidByUserID は有効なコードが取得できることをテスト
func TestSignInCodeRepository_GetValidByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewSignInCodeRepository(queries)

	userID := testutil.NewUserBuilder(t, tx).
		WithEmail("signincode-getvalid@example.com").
		Build()

	expiresAt := time.Now().Add(15 * time.Minute)
	created, err := repo.Create(context.Background(), repository.SignInCodeCreateParams{
		UserID:     userID,
		CodeDigest: "valid-digest",
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		t.Fatalf("Createに失敗: %v", err)
	}

	got, err := repo.GetValidByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetValidByUserIDに失敗: %v", err)
	}

	if got == nil {
		t.Fatal("コードが取得できませんでした")
	}
	if got.ID != created.ID {
		t.Errorf("IDが一致しません: got %v, want %v", got.ID, created.ID)
	}
	if got.UserID != userID {
		t.Errorf("UserIDが一致しません: got %v, want %v", got.UserID, userID)
	}
	if got.CodeDigest != "valid-digest" {
		t.Errorf("CodeDigestが一致しません: got %v, want valid-digest", got.CodeDigest)
	}
}

// TestSignInCodeRepository_WithTx はWithTxで取得したRepositoryがトランザクション内で動作することをテスト
func TestSignInCodeRepository_WithTx(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db)
	repo := repository.NewSignInCodeRepository(queries)

	repoWithTx := repo.WithTx(tx)

	userID := testutil.NewUserBuilder(t, tx).
		WithEmail("signincode-withtx@example.com").
		Build()

	code, err := repoWithTx.Create(context.Background(), repository.SignInCodeCreateParams{
		UserID:     userID,
		CodeDigest: "withtx-digest",
		ExpiresAt:  time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		t.Fatalf("WithTxで取得したRepositoryでCreateに失敗: %v", err)
	}

	if code.UserID != userID {
		t.Errorf("UserIDが一致しません: got %v, want %v", code.UserID, userID)
	}
}
