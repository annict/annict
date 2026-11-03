package usecase

import (
	"bytes"
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// TestPrepareTwoFactorAuthUsecaseは、登録の途中の設定に暗号化した秘密鍵を保存し、
// 画面には復号できる同じ秘密鍵とユーザー名入りのotpauth URIを返すことを検証する。
// 開き直したときは新しい秘密鍵に差し替える。
func TestPrepareTwoFactorAuthUsecase(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewUserTwoFactorAuthRepository(query.New(db).WithTx(tx))
	key := testutil.NewTestTwoFactorKey(t)
	uc := NewPrepareTwoFactorAuthUsecase(key, repo)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	user := &model.User{ID: userID, Username: "prepare_user"}

	first, err := uc.Execute(ctx, PrepareTwoFactorAuthInput{User: user})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if first.Setup == nil {
		t.Fatal("Setup = nil、秘密鍵を期待")
	}

	saved, err := repo.FindByUserID(ctx, userID)
	if err != nil || saved == nil {
		t.Fatalf("FindByUserID() = (%+v, %v)、登録の途中の設定を期待", saved, err)
	}
	if saved.IsEnabled() {
		t.Error("保存した設定が有効になっている、登録の途中を期待")
	}
	if bytes.Contains(saved.SecretCiphertext, []byte(first.Setup.Secret)) {
		t.Error("保存した値に平文の秘密鍵が含まれている")
	}
	decrypted, err := decryptTwoFactorSecret(key, saved)
	if err != nil {
		t.Fatalf("decryptTwoFactorSecret()のエラー = %v", err)
	}
	if decrypted != first.Setup.Secret {
		t.Errorf("復号した秘密鍵 = %q、期待値 = %q", decrypted, first.Setup.Secret)
	}

	u, err := url.Parse(first.Setup.OTPAuthURL)
	if err != nil {
		t.Fatalf("otpauth URIの解析のエラー = %v", err)
	}
	if u.Path != "/Annict:prepare_user" || u.Query().Get("secret") != first.Setup.Secret {
		t.Errorf("OTPAuthURL = %q、ユーザー名と秘密鍵を含むことを期待", first.Setup.OTPAuthURL)
	}

	second, err := uc.Execute(ctx, PrepareTwoFactorAuthInput{User: user})
	if err != nil {
		t.Fatalf("2回目のExecute()のエラー = %v", err)
	}
	if second.Setup == nil || second.Setup.Secret == first.Setup.Secret {
		t.Errorf("2回目のSetup = %+v、新しい秘密鍵を期待", second.Setup)
	}
}

// TestPrepareTwoFactorAuthUsecase_Enabledは、既に有効にしているときに秘密鍵を差し替えず、Setupをnilで返すことを検証する。
func TestPrepareTwoFactorAuthUsecase_Enabled(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewUserTwoFactorAuthRepository(query.New(db).WithTx(tx))
	uc := NewPrepareTwoFactorAuthUsecase(testutil.NewTestTwoFactorKey(t), repo)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	testutil.NewUserTwoFactorAuthBuilder(t, tx, userID).
		WithSecretCiphertext([]byte("enabled-secret")).
		WithEnabledAt(time.Now()).
		Build()

	output, err := uc.Execute(ctx, PrepareTwoFactorAuthInput{User: &model.User{ID: userID, Username: "enabled_user"}})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if output.Setup != nil {
		t.Errorf("Setup = %+v、nilを期待", output.Setup)
	}

	saved, err := repo.FindByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("FindByUserID()のエラー = %v", err)
	}
	if !bytes.Equal(saved.SecretCiphertext, []byte("enabled-secret")) {
		t.Errorf("秘密鍵 = %q、差し替えないことを期待", saved.SecretCiphertext)
	}
}
