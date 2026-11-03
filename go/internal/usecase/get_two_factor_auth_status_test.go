package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestGetTwoFactorAuthStatusUsecaseは、設定が無い・登録の途中のときは無効として返し、
// 有効のときは未使用のリカバリーコードの数を返すことを検証する。
func TestGetTwoFactorAuthStatusUsecase(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	recoveryCodeRepo := repository.NewUserTwoFactorRecoveryCodeRepository(queries)
	uc := NewGetTwoFactorAuthStatusUsecase(repository.NewUserTwoFactorAuthRepository(queries), recoveryCodeRepo)
	ctx := context.Background()

	noSettingUserID := testutil.NewUserBuilder(t, tx).Build()
	pendingUserID := testutil.NewUserBuilder(t, tx).Build()
	testutil.NewUserTwoFactorAuthBuilder(t, tx, pendingUserID).Build()
	enabledUserID := testutil.NewUserBuilder(t, tx).Build()
	testutil.NewUserTwoFactorAuthBuilder(t, tx, enabledUserID).WithEnabledAt(time.Now()).Build()
	if err := recoveryCodeRepo.CreateAll(ctx, enabledUserID, []string{"digest-1", "digest-2", "digest-3"}); err != nil {
		t.Fatalf("CreateAll()のエラー = %v", err)
	}
	if ok, err := recoveryCodeRepo.Use(ctx, enabledUserID, "digest-1"); err != nil || !ok {
		t.Fatalf("Use() = (%t, %v)、期待値 = (true, nil)", ok, err)
	}

	tests := []struct {
		name        string
		userID      model.UserID
		wantEnabled bool
		wantCount   int
	}{
		{name: "設定が無い", userID: noSettingUserID},
		{name: "登録の途中", userID: pendingUserID},
		{name: "有効", userID: enabledUserID, wantEnabled: true, wantCount: 2},
	}

	for _, tt := range tests {
		output, err := uc.Execute(ctx, GetTwoFactorAuthStatusInput{UserID: tt.userID})
		if err != nil {
			t.Fatalf("%s: Execute()のエラー = %v", tt.name, err)
		}
		if output.Enabled != tt.wantEnabled || output.UnusedRecoveryCodeCount != tt.wantCount {
			t.Errorf("%s: Execute() = %+v、期待値 = {Enabled:%t UnusedRecoveryCodeCount:%d}", tt.name, output, tt.wantEnabled, tt.wantCount)
		}
	}
}
