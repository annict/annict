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

// TestGetSignInTwoFactorStatusUsecaseは、有効なユーザーに有効な設定があるときだけ利用できることと、
// 状態を取得してもコードのステップを消費しないことを検証する。
func TestGetSignInTwoFactorStatusUsecase(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	authRepo := repository.NewUserTwoFactorAuthRepository(queries)
	uc := NewGetSignInTwoFactorStatusUsecase(repository.NewUserRepository(queries), authRepo)
	ctx := context.Background()

	noSettingUserID := testutil.NewUserBuilder(t, tx).Build()
	pendingUserID := testutil.NewUserBuilder(t, tx).Build()
	testutil.NewUserTwoFactorAuthBuilder(t, tx, pendingUserID).Build()
	enabledUserID := testutil.NewUserBuilder(t, tx).Build()
	testutil.NewUserTwoFactorAuthBuilder(t, tx, enabledUserID).WithEnabledAt(time.Now()).WithLastUsedStep(123).Build()
	deletedUserID := testutil.NewUserBuilder(t, tx).Build()
	testutil.NewUserTwoFactorAuthBuilder(t, tx, deletedUserID).WithEnabledAt(time.Now()).Build()
	if _, err := tx.ExecContext(ctx, "UPDATE users SET deleted_at = NOW() WHERE id = $1", int64(deletedUserID)); err != nil {
		t.Fatalf("ユーザーの論理削除に失敗: %v", err)
	}

	for _, tt := range []struct {
		name          string
		userID        model.UserID
		wantAvailable bool
	}{
		{name: "設定が無い", userID: noSettingUserID},
		{name: "登録の途中", userID: pendingUserID},
		{name: "有効", userID: enabledUserID, wantAvailable: true},
		{name: "退会した", userID: deletedUserID},
		{name: "存在しないユーザー", userID: 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			output, err := uc.Execute(ctx, GetSignInTwoFactorStatusInput{UserID: tt.userID})
			if err != nil {
				t.Fatalf("Execute()のエラー = %v", err)
			}
			if output.Available != tt.wantAvailable {
				t.Errorf("Available = %t、期待値 = %t", output.Available, tt.wantAvailable)
			}
		})
	}
	setting, err := authRepo.FindByUserID(ctx, enabledUserID)
	if err != nil || setting == nil {
		t.Fatalf("有効な設定の取得 = (%+v, %v)、設定が残っていることを期待", setting, err)
	}
	if setting.LastUsedStep != 123 {
		t.Errorf("LastUsedStep = %d、期待値 = 123", setting.LastUsedStep)
	}
}
