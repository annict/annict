package repository_test

import (
	"context"
	"testing"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// createRecoveryCodesはユーザーのリカバリーコードをダイジェストから作る。
func createRecoveryCodes(t *testing.T, repo *repository.UserTwoFactorRecoveryCodeRepository, userID model.UserID, digests ...string) {
	t.Helper()

	if err := repo.CreateAll(context.Background(), userID, digests); err != nil {
		t.Fatalf("CreateAll()のエラー = %v", err)
	}
}

// assertUnusedRecoveryCodesはユーザーの未使用のリカバリーコードの数を確かめる。
func assertUnusedRecoveryCodes(t *testing.T, repo *repository.UserTwoFactorRecoveryCodeRepository, userID model.UserID, want int) {
	t.Helper()

	got, err := repo.CountUnused(context.Background(), userID)
	if err != nil {
		t.Fatalf("CountUnused()のエラー = %v", err)
	}
	if got != want {
		t.Errorf("未使用のコードの数 = %d、期待値 = %d", got, want)
	}
}

// TestUserTwoFactorRecoveryCodeRepository_CreateAllは、渡したダイジェストの数だけ未使用のコードを作ることを検証する。
func TestUserTwoFactorRecoveryCodeRepository_CreateAll(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewUserTwoFactorRecoveryCodeRepository(query.New(db).WithTx(tx))

	userID := testutil.NewUserBuilder(t, tx).Build()
	createRecoveryCodes(t, repo, userID, "digest-1", "digest-2", "digest-3")

	assertUnusedRecoveryCodes(t, repo, userID, 3)
	assertUnusedRecoveryCodes(t, repo, testutil.NewUserBuilder(t, tx).Build(), 0)
}

// TestUserTwoFactorRecoveryCodeRepository_CreateAll_Emptyは、ダイジェストが空のときに何も作らずエラーも返さないことを検証する。
func TestUserTwoFactorRecoveryCodeRepository_CreateAll_Empty(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewUserTwoFactorRecoveryCodeRepository(query.New(db).WithTx(tx))

	userID := testutil.NewUserBuilder(t, tx).Build()
	createRecoveryCodes(t, repo, userID)

	assertUnusedRecoveryCodes(t, repo, userID, 0)
}

// TestUserTwoFactorRecoveryCodeRepository_Useは、一致する未使用のコードを1度だけ使え、
// 一致しないコードやほかのユーザーのコードは使えないことを検証する。
func TestUserTwoFactorRecoveryCodeRepository_Use(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewUserTwoFactorRecoveryCodeRepository(query.New(db).WithTx(tx))
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	otherID := testutil.NewUserBuilder(t, tx).Build()
	createRecoveryCodes(t, repo, userID, "digest-1", "digest-2")
	createRecoveryCodes(t, repo, otherID, "digest-other")

	tests := []struct {
		name   string
		digest string
		want   bool
	}{
		{name: "未使用のコード", digest: "digest-1", want: true},
		{name: "使ったコード", digest: "digest-1", want: false},
		{name: "一致しないコード", digest: "digest-none", want: false},
		{name: "ほかのユーザーのコード", digest: "digest-other", want: false},
	}

	for _, tt := range tests {
		ok, err := repo.Use(ctx, userID, tt.digest)
		if err != nil || ok != tt.want {
			t.Errorf("%s: Use() = (%t, %v)、期待値 = (%t, nil)", tt.name, ok, err, tt.want)
		}
	}

	assertUnusedRecoveryCodes(t, repo, userID, 1)
	assertUnusedRecoveryCodes(t, repo, otherID, 1)
}

// TestUserTwoFactorRecoveryCodeRepository_DeleteByUserIDは、指定したユーザーのコードだけを消すことを検証する。
func TestUserTwoFactorRecoveryCodeRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewUserTwoFactorRecoveryCodeRepository(query.New(db).WithTx(tx))

	target := testutil.NewUserBuilder(t, tx).Build()
	other := testutil.NewUserBuilder(t, tx).Build()
	createRecoveryCodes(t, repo, target, "digest-1", "digest-2")
	createRecoveryCodes(t, repo, other, "digest-1")

	if err := repo.DeleteByUserID(context.Background(), target); err != nil {
		t.Fatalf("DeleteByUserID()のエラー = %v", err)
	}

	assertUnusedRecoveryCodes(t, repo, target, 0)
	assertUnusedRecoveryCodes(t, repo, other, 1)
}
