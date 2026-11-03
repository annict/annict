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

// TestGetPendingTwoFactorAuthUsecaseは、登録の途中の秘密鍵を準備したときと同じ形で返し、
// 設定が無い・有効にしたときはSetupをnilで返すことを検証する。
func TestGetPendingTwoFactorAuthUsecase(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewUserTwoFactorAuthRepository(query.New(db).WithTx(tx))
	key := testutil.NewTestTwoFactorKey(t)
	uc := NewGetPendingTwoFactorAuthUsecase(key, repo)
	ctx := context.Background()

	pendingUser := &model.User{ID: testutil.NewUserBuilder(t, tx).Build(), Username: "pending_user"}
	prepared, err := NewPrepareTwoFactorAuthUsecase(key, repo).Execute(ctx, PrepareTwoFactorAuthInput{User: pendingUser})
	if err != nil {
		t.Fatalf("PrepareTwoFactorAuthUsecase.Execute()のエラー = %v", err)
	}

	got, err := uc.Execute(ctx, GetPendingTwoFactorAuthInput{User: pendingUser})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if got.Setup == nil || *got.Setup != *prepared.Setup {
		t.Errorf("Setup = %+v、期待値 = %+v", got.Setup, prepared.Setup)
	}

	noSettingUser := &model.User{ID: testutil.NewUserBuilder(t, tx).Build(), Username: "no_setting_user"}
	enabledUser := &model.User{ID: testutil.NewUserBuilder(t, tx).Build(), Username: "enabled_user"}
	testutil.NewUserTwoFactorAuthBuilder(t, tx, enabledUser.ID).WithEnabledAt(time.Now()).Build()
	for _, user := range []*model.User{noSettingUser, enabledUser} {
		got, err := uc.Execute(ctx, GetPendingTwoFactorAuthInput{User: user})
		if err != nil {
			t.Fatalf("%s: Execute()のエラー = %v", user.Username, err)
		}
		if got.Setup != nil {
			t.Errorf("%s: Setup = %+v、nilを期待", user.Username, got.Setup)
		}
	}
}

// TestGetPendingTwoFactorAuthUsecase_OtherUsersCiphertextは、別のユーザーの行へ書き写した暗号文を復号できず、
// エラーを返すことを検証する。暗号化の付加データにユーザーIDを渡しているため。
func TestGetPendingTwoFactorAuthUsecase_OtherUsersCiphertext(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewUserTwoFactorAuthRepository(query.New(db).WithTx(tx))
	key := testutil.NewTestTwoFactorKey(t)
	ctx := context.Background()

	owner := &model.User{ID: testutil.NewUserBuilder(t, tx).Build(), Username: "owner"}
	if _, err := NewPrepareTwoFactorAuthUsecase(key, repo).Execute(ctx, PrepareTwoFactorAuthInput{User: owner}); err != nil {
		t.Fatalf("PrepareTwoFactorAuthUsecase.Execute()のエラー = %v", err)
	}
	ownerSetting, err := repo.FindByUserID(ctx, owner.ID)
	if err != nil {
		t.Fatalf("FindByUserID()のエラー = %v", err)
	}

	other := &model.User{ID: testutil.NewUserBuilder(t, tx).Build(), Username: "other"}
	testutil.NewUserTwoFactorAuthBuilder(t, tx, other.ID).WithSecretCiphertext(ownerSetting.SecretCiphertext).Build()

	if _, err := NewGetPendingTwoFactorAuthUsecase(key, repo).Execute(ctx, GetPendingTwoFactorAuthInput{User: other}); err == nil {
		t.Error("Execute()のエラー = nil、復号の失敗を期待")
	}
}
