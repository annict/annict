package usecase

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

// enableTwoFactorAuthFixtureは有効化のUseCaseと、その検証に使うリポジトリをまとめる。
type enableTwoFactorAuthFixture struct {
	uc               *EnableTwoFactorAuthUsecase
	key              *auth.TwoFactorKey
	authRepo         *repository.UserTwoFactorAuthRepository
	recoveryCodeRepo *repository.UserTwoFactorRecoveryCodeRepository
	// notificationsは通知メールのジョブの投入を記録する
	notifications *recordingJobInserter
	dispatcher    *dispatcher.Dispatcher
	userRepo      *repository.UserRepository
}

// newEnableTwoFactorAuthFixtureは共有テストDBに対して有効化のUseCaseを組み立てる。
// UseCaseが開くトランザクションから見えるよう、テストは準備した行をコミットしておく。
func newEnableTwoFactorAuthFixture(t *testing.T, db *sql.DB) *enableTwoFactorAuthFixture {
	t.Helper()

	queries := query.New(db)
	key := testutil.NewTestTwoFactorKey(t)
	authRepo := repository.NewUserTwoFactorAuthRepository(queries)
	recoveryCodeRepo := repository.NewUserTwoFactorRecoveryCodeRepository(queries)
	userRepo := repository.NewUserRepository(queries)
	notifications := &recordingJobInserter{}
	d := dispatcher.NewDispatcher(notifications, nil)
	return &enableTwoFactorAuthFixture{
		uc:               NewEnableTwoFactorAuthUsecase(db, key, validator.NewSettingsTwoFactorAuthCreateValidator(), authRepo, recoveryCodeRepo, userRepo, d),
		key:              key,
		authRepo:         authRepo,
		recoveryCodeRepo: recoveryCodeRepo,
		notifications:    notifications,
		dispatcher:       d,
		userRepo:         userRepo,
	}
}

// preparePendingは登録の途中の設定を作り、その秘密鍵を返す。
func (f *enableTwoFactorAuthFixture) preparePending(t *testing.T, userID model.UserID) string {
	t.Helper()

	output, err := NewPrepareTwoFactorAuthUsecase(f.key, f.authRepo).Execute(context.Background(), PrepareTwoFactorAuthInput{
		User: &model.User{ID: userID, Username: "enable_user"},
	})
	if err != nil || output.Setup == nil {
		t.Fatalf("PrepareTwoFactorAuthUsecase.Execute() = (%+v, %v)、秘密鍵を期待", output, err)
	}
	return output.Setup.Secret
}

// TestEnableTwoFactorAuthUsecaseは、認証アプリのコードが合えば設定を有効にして照合したステップを記録し、
// 10個のリカバリーコードの平文を返してダイジェストだけを保存することを検証する。
// 残っていた古いリカバリーコードは消える。
func TestEnableTwoFactorAuthUsecase(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	f := newEnableTwoFactorAuthFixture(t, db)
	ctx := i18n.SetLocale(context.Background(), "ja")

	userID, _ := createCommittedUserForTest(t, db, false)
	secret := f.preparePending(t, userID)
	if err := f.recoveryCodeRepo.CreateAll(ctx, userID, []string{"stale-digest"}); err != nil {
		t.Fatalf("CreateAll()のエラー = %v", err)
	}

	now := time.Now()
	// 貼り付けで混じる半角と全角の空白も受け付ける。
	code := testutil.TOTPCode(t, secret, now)
	output, err := f.uc.Execute(ctx, EnableTwoFactorAuthInput{UserID: userID, Code: " " + code[:3] + "　" + code[3:]})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}

	if len(output.RecoveryCodes) != auth.RecoveryCodeCount {
		t.Errorf("リカバリーコードの数 = %d、期待値 = %d", len(output.RecoveryCodes), auth.RecoveryCodeCount)
	}
	pattern := regexp.MustCompile(`^[a-z2-9]{4}-[a-z2-9]{4}$`)
	for _, recoveryCode := range output.RecoveryCodes {
		if !pattern.MatchString(recoveryCode) {
			t.Errorf("リカバリーコード = %q、xxxx-xxxxの形を期待", recoveryCode)
		}
	}

	saved, err := f.authRepo.FindByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("FindByUserID()のエラー = %v", err)
	}
	if !saved.IsEnabled() {
		t.Error("設定が有効になっていない")
	}
	if step := auth.TOTPStep(now); saved.LastUsedStep < step-1 || saved.LastUsedStep > step+1 {
		t.Errorf("LastUsedStep = %d、照合したステップ (%d前後) を期待", saved.LastUsedStep, step)
	}

	count, err := f.recoveryCodeRepo.CountUnused(ctx, userID)
	if err != nil {
		t.Fatalf("CountUnused()のエラー = %v", err)
	}
	if count != auth.RecoveryCodeCount {
		t.Errorf("未使用のリカバリーコードの数 = %d、期待値 = %d (古いコードは消える)", count, auth.RecoveryCodeCount)
	}
	assertTwoFactorAuthNotifications(t, f.notifications, db, userID, model.TwoFactorAuthNotificationEventEnabled)
	if ok, err := f.recoveryCodeRepo.Use(ctx, userID, "stale-digest"); err != nil || ok {
		t.Errorf("古いコードのUse() = (%t, %v)、期待値 = (false, nil)", ok, err)
	}
	digest := f.key.RecoveryCodeDigest(auth.NormalizeRecoveryCode(output.RecoveryCodes[0]))
	if ok, err := f.recoveryCodeRepo.Use(ctx, userID, digest); err != nil || !ok {
		t.Errorf("返したコードのダイジェストのUse() = (%t, %v)、期待値 = (true, nil)", ok, err)
	}
}

// TestEnableTwoFactorAuthUsecase_InvalidCodeは、形の違うコードと一致しないコードを
// コードの欄のValidationErrorで拒み、設定を有効にしないことを検証する。
func TestEnableTwoFactorAuthUsecase_InvalidCode(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	f := newEnableTwoFactorAuthFixture(t, db)
	ctx := i18n.SetLocale(context.Background(), "ja")

	userID, _ := createCommittedUserForTest(t, db, false)
	secret := f.preparePending(t, userID)
	incorrectCode := testutil.NonMatchingTOTPCode(t, secret, time.Now())

	tests := []struct {
		name    string
		code    string
		wantMsg string
	}{
		{name: "空", code: "", wantMsg: "入力してください"},
		{name: "5桁", code: "12345", wantMsg: "認証アプリに表示された6桁の数字を入力してください"},
		{name: "一致しない", code: incorrectCode, wantMsg: "コードが正しくありません。認証アプリに表示されている最新のコードを入力してください"},
	}

	for _, tt := range tests {
		_, err := f.uc.Execute(ctx, EnableTwoFactorAuthInput{UserID: userID, Code: tt.code})
		ve := model.AsValidationError(err)
		if ve == nil {
			t.Fatalf("%s: Execute()のエラー = %v、ValidationErrorを期待", tt.name, err)
		}
		if got := ve.GetFieldErrors("code"); len(got) != 1 || got[0] != tt.wantMsg {
			t.Errorf("%s: codeのフィールドエラー = %v、期待値 = [%q]", tt.name, got, tt.wantMsg)
		}
	}

	saved, err := f.authRepo.FindByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("FindByUserID()のエラー = %v", err)
	}
	if saved.IsEnabled() {
		t.Error("コードを受け付けなかったのに設定が有効になっている")
	}
	assertTwoFactorAuthNotifications(t, f.notifications, db, userID)
}

// TestEnableTwoFactorAuthUsecase_NotPendingは、登録の途中の設定が無い・既に有効にしたときに
// AppErrCodeConflictを返し、リカバリーコードを作らないことを検証する。
func TestEnableTwoFactorAuthUsecase_NotPending(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	f := newEnableTwoFactorAuthFixture(t, db)
	ctx := i18n.SetLocale(context.Background(), "ja")

	noSettingUserID, _ := createCommittedUserForTest(t, db, false)
	if _, err := f.uc.Execute(ctx, EnableTwoFactorAuthInput{UserID: noSettingUserID, Code: "123456"}); !isConflictForTest(err) {
		t.Errorf("設定の無いユーザーのExecute()のエラー = %v、AppErrCodeConflictを期待", err)
	}

	// 同じコードの二重送信では、2回目が既に有効になった設定に当たる。
	enabledUserID, _ := createCommittedUserForTest(t, db, false)
	secret := f.preparePending(t, enabledUserID)
	code := testutil.TOTPCode(t, secret, time.Now())
	if _, err := f.uc.Execute(ctx, EnableTwoFactorAuthInput{UserID: enabledUserID, Code: code}); err != nil {
		t.Fatalf("1回目のExecute()のエラー = %v", err)
	}
	if _, err := f.uc.Execute(ctx, EnableTwoFactorAuthInput{UserID: enabledUserID, Code: code}); !isConflictForTest(err) {
		t.Errorf("2回目のExecute()のエラー = %v、AppErrCodeConflictを期待", err)
	}

	count, err := f.recoveryCodeRepo.CountUnused(ctx, enabledUserID)
	if err != nil {
		t.Fatalf("CountUnused()のエラー = %v", err)
	}
	if count != auth.RecoveryCodeCount {
		t.Errorf("未使用のリカバリーコードの数 = %d、期待値 = %d (2回目では作り直さない)", count, auth.RecoveryCodeCount)
	}
	// 二重送信でも、有効にしたことを知らせるメールは1通だけにする。
	assertTwoFactorAuthNotifications(t, f.notifications, db, enabledUserID, model.TwoFactorAuthNotificationEventEnabled)
	assertTwoFactorAuthNotifications(t, f.notifications, db, noSettingUserID)
}

// isConflictForTestはエラーがAppErrCodeConflictの*model.AppErrorかを返す。
func isConflictForTest(err error) bool {
	ae := model.AsAppError(err)
	return ae != nil && ae.Code == model.AppErrCodeConflict
}
