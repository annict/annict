package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/testutil"
)

// TestForceDisableTwoFactorAuthUsecaseは、本人確認なしで設定とリカバリーコードを削除し、
// 有効だったときだけDisabledをtrueにすることを検証する。
func TestForceDisableTwoFactorAuthUsecase(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	f := newDisableTwoFactorAuthFixture(t, db)
	uc := NewForceDisableTwoFactorAuthUsecase(db, f.authRepo, f.recoveryCodeRepo, f.userRepo, f.dispatcher)
	ctx := context.Background()

	enabled := f.createEnabledUser(t, db, "")
	output, err := uc.Execute(ctx, ForceDisableTwoFactorAuthInput{UserID: enabled.id})
	if err != nil {
		t.Fatalf("有効なユーザーのExecute()のエラー = %v", err)
	}
	if !output.Disabled {
		t.Error("有効なユーザーのDisabled = false、trueを期待")
	}
	f.assertTwoFactorAuthDeleted(t, enabled.id, true)
	assertTwoFactorAuthNotifications(t, f.notifications, db, enabled.id, model.TwoFactorAuthNotificationEventEnabled, model.TwoFactorAuthNotificationEventDisabledByAdmin)

	pendingUserID, _ := createCommittedUserForTest(t, db, false)
	f.preparePending(t, pendingUserID)
	output, err = uc.Execute(ctx, ForceDisableTwoFactorAuthInput{UserID: pendingUserID})
	if err != nil {
		t.Fatalf("登録の途中のユーザーのExecute()のエラー = %v", err)
	}
	if output.Disabled {
		t.Error("登録の途中のユーザーのDisabled = true、falseを期待")
	}
	f.assertTwoFactorAuthDeleted(t, pendingUserID, true)
	// 登録の途中の設定を消しただけのときは知らせない。
	assertTwoFactorAuthNotifications(t, f.notifications, db, pendingUserID)

	noSettingUserID, _ := createCommittedUserForTest(t, db, false)
	output, err = uc.Execute(ctx, ForceDisableTwoFactorAuthInput{UserID: noSettingUserID})
	if err != nil || output.Disabled {
		t.Errorf("設定の無いユーザーのExecute() = (%+v, %v)、期待値 = (Disabled:false, nil)", output, err)
	}
}

// TestForceDisableTwoFactorAuthUsecase_EnabledWhileWaitingは、削除のロック待ちの間に有効になった設定を
// 無効にしたときもDisabledをtrueにし、リカバリーコードも消すことを検証する。
func TestForceDisableTwoFactorAuthUsecase_EnabledWhileWaiting(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	f := newDisableTwoFactorAuthFixture(t, db)
	userID, _ := createCommittedUserForTest(t, db, false)
	f.preparePending(t, userID)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	pending, err := f.authRepo.FindByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("登録の途中の設定のFindByUserID()のエラー = %v", err)
	}
	blocker, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("ブロッカーのトランザクションの開始に失敗: %v", err)
	}
	defer func() { _ = blocker.Rollback() }()
	var blockerPID int
	if err := blocker.QueryRowContext(ctx, "SELECT pg_backend_pid()").Scan(&blockerPID); err != nil {
		t.Fatalf("ブロッカーのバックエンドのPIDの取得に失敗: %v", err)
	}
	var id int64
	if err := blocker.QueryRowContext(ctx, `SELECT id FROM user_two_factor_auths WHERE user_id = $1 FOR UPDATE`, int64(userID)).Scan(&id); err != nil {
		t.Fatalf("ブロッカーでの設定の行のロックに失敗: %v", err)
	}
	type result struct {
		output *ForceDisableTwoFactorAuthOutput
		err    error
	}
	results := make(chan result, 1)
	go func() {
		output, err := NewForceDisableTwoFactorAuthUsecase(db, f.authRepo, f.recoveryCodeRepo, f.userRepo, f.dispatcher).Execute(ctx, ForceDisableTwoFactorAuthInput{UserID: userID})
		results <- result{output, err}
	}()
	awaitBlockedByBackend(t, ctx, db, blockerPID, results)
	if enabled, err := f.authRepo.WithTx(blocker).Enable(ctx, userID, pending.SecretCiphertext, auth.TOTPStep(time.Now())); err != nil || !enabled {
		t.Fatalf("ロック待ちの間の有効化 = (%t, %v)、成功を期待", enabled, err)
	}
	if err := f.recoveryCodeRepo.WithTx(blocker).CreateAll(ctx, userID, []string{"enabled-while-waiting"}); err != nil {
		t.Fatalf("ロック待ちの間のリカバリーコードの作成に失敗: %v", err)
	}
	if err := blocker.Commit(); err != nil {
		t.Fatalf("ブロッカーのトランザクションのコミットに失敗: %v", err)
	}
	got := <-results
	if got.err != nil || got.output == nil || !got.output.Disabled {
		t.Fatalf("Execute() = (%+v, %v)、Disabled=trueを期待", got.output, got.err)
	}
	f.assertTwoFactorAuthDeleted(t, userID, true)
}
