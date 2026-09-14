package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestSignUpCodeRepository_Createはサインアップコードを正常に作成し、Modelとして返却されることをテスト
func TestSignUpCodeRepository_Create(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewSignUpCodeRepository(queries)

	email := "signupcode-create@example.com"
	codeDigest := "bcrypt-hashed-digest"
	expiresAt := time.Now().Add(15 * time.Minute)

	code, err := repo.Create(context.Background(), repository.SignUpCodeCreateParams{
		Email:      email,
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
		t.Error("SignUpCodeIDがゼロ値です")
	}
	if code.Email != email {
		t.Errorf("Email = %v、期待値 = %v", code.Email, email)
	}
	if code.CodeDigest != codeDigest {
		t.Errorf("CodeDigest = %v、期待値 = %v", code.CodeDigest, codeDigest)
	}
	if code.Attempts != 0 {
		t.Errorf("Attemptsの初期値 = %v、期待値 = 0", code.Attempts)
	}
	if code.UsedAt.Valid {
		t.Error("UsedAtが有効値です (未使用のはず)")
	}
	if code.CreatedAt.IsZero() {
		t.Error("CreatedAtがゼロ値です")
	}
	if code.UpdatedAt.IsZero() {
		t.Error("UpdatedAtがゼロ値です")
	}
}

// TestSignUpCodeRepository_GetValidByEmailは有効なコードが取得できることをテスト
func TestSignUpCodeRepository_GetValidByEmail(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewSignUpCodeRepository(queries)

	email := "signupcode-getvalid@example.com"
	expiresAt := time.Now().Add(15 * time.Minute)

	created, err := repo.Create(context.Background(), repository.SignUpCodeCreateParams{
		Email:      email,
		CodeDigest: "valid-digest",
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		t.Fatalf("Createに失敗: %v", err)
	}

	got, err := repo.GetValidByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("GetValidByEmailに失敗: %v", err)
	}

	if got == nil {
		t.Fatal("コードが取得できませんでした")
	}
	if got.ID != created.ID {
		t.Errorf("ID = %v、期待値 = %v", got.ID, created.ID)
	}
	if got.Email != email {
		t.Errorf("Email = %v、期待値 = %v", got.Email, email)
	}
	if got.CodeDigest != "valid-digest" {
		t.Errorf("CodeDigest = %v、期待値 = valid-digest", got.CodeDigest)
	}
}

// TestSignUpCodeRepository_WithTxはWithTxで取得したRepositoryがトランザクション内で動作することをテスト
func TestSignUpCodeRepository_WithTx(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db)
	repo := repository.NewSignUpCodeRepository(queries)

	repoWithTx := repo.WithTx(tx)

	email := "signupcode-withtx@example.com"
	code, err := repoWithTx.Create(context.Background(), repository.SignUpCodeCreateParams{
		Email:      email,
		CodeDigest: "withtx-digest",
		ExpiresAt:  time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		t.Fatalf("WithTxで取得したRepositoryでCreateに失敗: %v", err)
	}

	if code.Email != email {
		t.Errorf("Email = %v、期待値 = %v", code.Email, email)
	}
}
