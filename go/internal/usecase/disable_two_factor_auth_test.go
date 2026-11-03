package usecase

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/validator"
)

// disableTestPasswordは無効化のテストのユーザーに設定するパスワード。
const disableTestPassword = "disable-password-123"

// disableTwoFactorAuthFixtureは無効化のUseCaseと、二要素認証を有効にした状態を作る有効化のフィクスチャをまとめる。
type disableTwoFactorAuthFixture struct {
	*enableTwoFactorAuthFixture
	uc *DisableTwoFactorAuthUsecase
}

// newDisableTwoFactorAuthFixtureは共有テストDBに対して無効化のUseCaseを組み立てる。
// UseCaseが開くトランザクションから見えるよう、テストは準備した行をコミットしておく。
func newDisableTwoFactorAuthFixture(t *testing.T, db *sql.DB) *disableTwoFactorAuthFixture {
	t.Helper()

	enable := newEnableTwoFactorAuthFixture(t, db)
	return &disableTwoFactorAuthFixture{
		enableTwoFactorAuthFixture: enable,
		uc: NewDisableTwoFactorAuthUsecase(
			db,
			enable.key,
			validator.NewSettingsTwoFactorAuthDeleteValidator(),
			repository.NewUserRepository(query.New(db)),
			enable.authRepo,
			enable.recoveryCodeRepo,
			enable.dispatcher,
		),
	}
}

// enabledUserは、encryptedPasswordのユーザーを作って二要素認証を有効にし、秘密鍵とリカバリーコードと、
// 有効にしたときに使ったコードのタイムステップを返す。
type enabledUser struct {
	id            model.UserID
	secret        string
	recoveryCodes []string
}

func (f *disableTwoFactorAuthFixture) createEnabledUser(t *testing.T, db *sql.DB, encryptedPassword string) *enabledUser {
	t.Helper()

	userID, _ := createCommittedUserWithBuilderForTest(t, db, func(tx *sql.Tx) *testutil.UserBuilder {
		return testutil.NewUserBuilder(t, tx).WithEncryptedPassword(encryptedPassword)
	}, false)
	secret := f.preparePending(t, userID)
	// 有効にするコードは1つ前のステップのものにし、無効化で今のステップのコードを使えるようにする。
	output, err := f.enableTwoFactorAuthFixture.uc.Execute(context.Background(), EnableTwoFactorAuthInput{
		UserID: userID,
		Code:   testutil.TOTPCode(t, secret, time.Now().Add(-30*time.Second)),
	})
	if err != nil {
		t.Fatalf("EnableTwoFactorAuthUsecase.Execute()のエラー = %v", err)
	}
	return &enabledUser{id: userID, secret: secret, recoveryCodes: output.RecoveryCodes}
}

// assertTwoFactorAuthDeletedは、設定とリカバリーコードがすべて消えたかどうかを確かめる。
func (f *disableTwoFactorAuthFixture) assertTwoFactorAuthDeleted(t *testing.T, userID model.UserID, want bool) {
	t.Helper()

	ctx := context.Background()
	setting, err := f.authRepo.FindByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("FindByUserID()のエラー = %v", err)
	}
	count, err := f.recoveryCodeRepo.CountUnused(ctx, userID)
	if err != nil {
		t.Fatalf("CountUnused()のエラー = %v", err)
	}
	if deleted := setting == nil && count == 0; deleted != want {
		t.Errorf("設定とリカバリーコードを削除したか = %t、期待値 = %t (設定 = %+v、未使用のコード = %d)", deleted, want, setting, count)
	}
}

// hashPasswordForTestはパスワードをbcryptでハッシュにする。
func hashPasswordForTest(t *testing.T, password string) string {
	t.Helper()

	hashed, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword()のエラー = %v", err)
	}
	return hashed
}

// TestDisableTwoFactorAuthUsecaseは、認証アプリのコード・リカバリーコード・パスワードのいずれでも本人確認でき、
// 設定とリカバリーコードを削除して、受け付けた入力の種類を返すことを検証する。
func TestDisableTwoFactorAuthUsecase(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	f := newDisableTwoFactorAuthFixture(t, db)
	ctx := i18n.SetLocale(context.Background(), "ja")
	hashedPassword := hashPasswordForTest(t, disableTestPassword)

	tests := []struct {
		name       string
		credential func(user *enabledUser) string
		wantMethod TwoFactorAuthDisableMethod
	}{
		{
			name: "認証アプリのコード (全角の空白を含む)",
			credential: func(user *enabledUser) string {
				code := testutil.TOTPCode(t, user.secret, time.Now())
				return code[:3] + "　" + code[3:]
			},
			wantMethod: TwoFactorAuthDisableMethodTOTP,
		},
		{
			name:       "リカバリーコード (大文字・区切り無し)",
			credential: func(user *enabledUser) string { return "  " + toUpperWithoutHyphen(user.recoveryCodes[0]) },
			wantMethod: TwoFactorAuthDisableMethodRecoveryCode,
		},
		{
			name:       "パスワード",
			credential: func(*enabledUser) string { return disableTestPassword },
			wantMethod: TwoFactorAuthDisableMethodPassword,
		},
	}

	for _, tt := range tests {
		user := f.createEnabledUser(t, db, hashedPassword)

		output, err := f.uc.Execute(ctx, DisableTwoFactorAuthInput{UserID: user.id, Credential: tt.credential(user)})
		if err != nil {
			t.Fatalf("%s: Execute()のエラー = %v", tt.name, err)
		}
		if output.Method != tt.wantMethod {
			t.Errorf("%s: Method = %q、期待値 = %q", tt.name, output.Method, tt.wantMethod)
		}
		f.assertTwoFactorAuthDeleted(t, user.id, true)
		assertTwoFactorAuthNotifications(t, f.notifications, db, user.id, model.TwoFactorAuthNotificationEventEnabled, model.TwoFactorAuthNotificationEventDisabled)
	}
}

// TestDisableTwoFactorAuthUsecase_Rejectedは、どれとしても一致しない入力・使用済みのコードを
// 本人確認の欄の同じ文面のValidationErrorで拒み、設定を残すことを検証する。
func TestDisableTwoFactorAuthUsecase_Rejected(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	f := newDisableTwoFactorAuthFixture(t, db)
	ctx := i18n.SetLocale(context.Background(), "ja")

	withPassword := f.createEnabledUser(t, db, hashPasswordForTest(t, disableTestPassword))
	withoutPassword := f.createEnabledUser(t, db, "")
	setting, err := f.authRepo.FindByUserID(ctx, withPassword.id)
	if err != nil {
		t.Fatalf("FindByUserID()のエラー = %v", err)
	}
	// 記録済みのステップのコードを、そのステップの時刻から作る。テストの途中で30秒の境界をまたいでも同じコードになる。
	lastUsedCode := testutil.TOTPCode(t, withPassword.secret, time.Unix(setting.LastUsedStep*30, 0))
	usedRecoveryCode := withPassword.recoveryCodes[1]
	if _, err := f.recoveryCodeRepo.Use(ctx, withPassword.id, f.key.RecoveryCodeDigest(auth.NormalizeRecoveryCode(usedRecoveryCode))); err != nil {
		t.Fatalf("Use()のエラー = %v", err)
	}

	tests := []struct {
		name       string
		user       *enabledUser
		credential string
		wantMsg    string
	}{
		{name: "空", user: withPassword, credential: "", wantMsg: "入力してください"},
		{name: "違うパスワード", user: withPassword, credential: "wrong-password", wantMsg: "パスワードまたはコードが正しくありません"},
		{name: "照合窓の外のコード", user: withPassword, credential: testutil.NonMatchingTOTPCode(t, withPassword.secret, time.Now()), wantMsg: "パスワードまたはコードが正しくありません"},
		// 有効にしたときに使ったコードは、記録済みのステップのため使い回せない。
		{name: "有効にしたときのコード", user: withPassword, credential: lastUsedCode, wantMsg: "パスワードまたはコードが正しくありません"},
		{name: "使用済みのリカバリーコード", user: withPassword, credential: usedRecoveryCode, wantMsg: "パスワードまたはコードが正しくありません"},
		{name: "パスワードの無いユーザー", user: withoutPassword, credential: "", wantMsg: "入力してください"},
		{name: "パスワードの無いユーザーの任意の文字列", user: withoutPassword, credential: "anything", wantMsg: "パスワードまたはコードが正しくありません"},
	}

	for _, tt := range tests {
		_, err := f.uc.Execute(ctx, DisableTwoFactorAuthInput{UserID: tt.user.id, Credential: tt.credential})
		ve := model.AsValidationError(err)
		if ve == nil {
			t.Fatalf("%s: Execute()のエラー = %v、ValidationErrorを期待", tt.name, err)
		}
		if got := ve.GetFieldErrors("credential"); len(got) != 1 || got[0] != tt.wantMsg {
			t.Errorf("%s: credentialのフィールドエラー = %v、期待値 = [%q]", tt.name, got, tt.wantMsg)
		}
	}

	f.assertTwoFactorAuthDeleted(t, withPassword.id, false)
	f.assertTwoFactorAuthDeleted(t, withoutPassword.id, false)
	// 本人確認を受け付けなかったときは、無効にしたことを知らせない。
	assertTwoFactorAuthNotifications(t, f.notifications, db, withPassword.id, model.TwoFactorAuthNotificationEventEnabled)
}

// TestDisableTwoFactorAuthUsecase_NotEnabledは、設定が無い・登録の途中・二重送信で既に無効にしたときに
// AppErrCodeConflictを返すことを検証する。
func TestDisableTwoFactorAuthUsecase_NotEnabled(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	f := newDisableTwoFactorAuthFixture(t, db)
	ctx := i18n.SetLocale(context.Background(), "ja")
	hashedPassword := hashPasswordForTest(t, disableTestPassword)

	noSettingUserID, _ := createCommittedUserForTest(t, db, false)
	if _, err := f.uc.Execute(ctx, DisableTwoFactorAuthInput{UserID: noSettingUserID, Credential: disableTestPassword}); !isConflictForTest(err) {
		t.Errorf("設定の無いユーザーのExecute()のエラー = %v、AppErrCodeConflictを期待", err)
	}

	pendingUserID, _ := createCommittedUserForTest(t, db, false)
	f.preparePending(t, pendingUserID)
	if _, err := f.uc.Execute(ctx, DisableTwoFactorAuthInput{UserID: pendingUserID, Credential: disableTestPassword}); !isConflictForTest(err) {
		t.Errorf("登録の途中のユーザーのExecute()のエラー = %v、AppErrCodeConflictを期待", err)
	}

	user := f.createEnabledUser(t, db, hashedPassword)
	if _, err := f.uc.Execute(ctx, DisableTwoFactorAuthInput{UserID: user.id, Credential: disableTestPassword}); err != nil {
		t.Fatalf("1回目のExecute()のエラー = %v", err)
	}
	if _, err := f.uc.Execute(ctx, DisableTwoFactorAuthInput{UserID: user.id, Credential: disableTestPassword}); !isConflictForTest(err) {
		t.Errorf("2回目のExecute()のエラー = %v、AppErrCodeConflictを期待", err)
	}
	// 二重送信でも、無効にしたことを知らせるメールは1通だけにする。
	assertTwoFactorAuthNotifications(t, f.notifications, db, user.id, model.TwoFactorAuthNotificationEventEnabled, model.TwoFactorAuthNotificationEventDisabled)
}

// toUpperWithoutHyphenはリカバリーコードを大文字にし、区切りのハイフンを取り除く。
func toUpperWithoutHyphen(code string) string {
	out := make([]rune, 0, len(code))
	for _, r := range code {
		switch {
		case r == '-':
			continue
		case r >= 'a' && r <= 'z':
			out = append(out, r-'a'+'A')
		default:
			out = append(out, r)
		}
	}
	return string(out)
}

// TestDisableTwoFactorAuthUsecase_SettingRecreatedは、TOTPの照合後に無効化と再登録を挟むと、
// 古い秘密鍵の照合結果を拒否し、新しい設定とリカバリーコードを残すことを検証する。
func TestDisableTwoFactorAuthUsecase_SettingRecreated(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	f := newDisableTwoFactorAuthFixture(t, db)
	ctx := context.Background()
	user := f.createEnabledUser(t, db, "")
	previous, err := f.authRepo.FindByUserID(ctx, user.id)
	if err != nil {
		t.Fatalf("古い設定のFindByUserID()のエラー = %v", err)
	}
	credential, err := f.uc.matchCredential(ctx, previous, testutil.TOTPCode(t, user.secret, time.Now().Add(30*time.Second)))
	if err != nil || credential.totpStep == 0 {
		t.Fatalf("古い設定の照合に失敗: %v", err)
	}
	if _, err := NewForceDisableTwoFactorAuthUsecase(db, f.authRepo, f.recoveryCodeRepo, f.userRepo, f.dispatcher).Execute(ctx, ForceDisableTwoFactorAuthInput{UserID: user.id}); err != nil {
		t.Fatalf("照合後の強制無効化のExecute()のエラー = %v", err)
	}
	secret := f.preparePending(t, user.id)
	if _, err := f.enableTwoFactorAuthFixture.uc.Execute(ctx, EnableTwoFactorAuthInput{UserID: user.id, Code: testutil.TOTPCode(t, secret, time.Now())}); err != nil {
		t.Fatalf("再登録した設定の有効化のExecute()のエラー = %v", err)
	}
	current, err := f.authRepo.FindByUserID(ctx, user.id)
	if err != nil {
		t.Fatalf("再登録した設定のFindByUserID()のエラー = %v", err)
	}
	if _, err := f.uc.disable(ctx, previous, credential); !isConflictForTest(err) {
		t.Fatalf("古い設定での無効化のエラー = %v、競合を期待", err)
	}
	saved, err := f.authRepo.FindByUserID(ctx, user.id)
	if err != nil || saved == nil || saved.ID != current.ID || !saved.IsEnabled() {
		t.Fatalf("現在の設定が残っていない: %+v、エラー = %v", saved, err)
	}
	if count, err := f.recoveryCodeRepo.CountUnused(ctx, user.id); err != nil || count != auth.RecoveryCodeCount {
		t.Errorf("残ったリカバリーコード = (%d, %v)、期待値 = (%d, nil)", count, err, auth.RecoveryCodeCount)
	}
}

// TestDisableTwoFactorAuthUsecase_ConcurrentMethodsは、リカバリーコードとTOTPによる無効化が
// 同じ順序でロックを取り、先に無効にした方だけが成功し、他方はデッドロックせず競合を返すことを検証する。
func TestDisableTwoFactorAuthUsecase_ConcurrentMethods(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	f := newDisableTwoFactorAuthFixture(t, db)
	user := f.createEnabledUser(t, db, "")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	setting, err := f.authRepo.FindByUserID(ctx, user.id)
	if err != nil {
		t.Fatalf("FindByUserID()のエラー = %v", err)
	}
	recovery, err := f.uc.matchCredential(ctx, setting, user.recoveryCodes[0])
	if err != nil {
		t.Fatalf("リカバリーコードのmatchCredential()のエラー = %v", err)
	}
	totp, err := f.uc.matchCredential(ctx, setting, testutil.TOTPCode(t, user.secret, time.Now()))
	if err != nil {
		t.Fatalf("TOTPのmatchCredential()のエラー = %v", err)
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
	if err := blocker.QueryRowContext(ctx, `SELECT id FROM user_two_factor_recovery_codes WHERE user_id = $1 AND code_digest = $2 FOR UPDATE`, int64(user.id), recovery.recoveryCodeDigest).Scan(&id); err != nil {
		t.Fatalf("ブロッカーでのリカバリーコードの行のロックに失敗: %v", err)
	}
	type result struct {
		method TwoFactorAuthDisableMethod
		err    error
	}
	recoveryResults := make(chan result, 1)
	go func() { method, err := f.uc.disable(ctx, setting, recovery); recoveryResults <- result{method, err} }()
	awaitBlockedByBackend(t, ctx, db, blockerPID, recoveryResults)
	var recoveryPID int
	if err := db.QueryRowContext(ctx, `SELECT pid FROM pg_stat_activity WHERE $1 = ANY(pg_blocking_pids(pid)) AND wait_event_type = 'Lock'`, blockerPID).Scan(&recoveryPID); err != nil {
		t.Fatalf("リカバリーコードの無効化のバックエンドのPIDの取得に失敗: %v", err)
	}
	totpResults := make(chan result, 1)
	go func() { method, err := f.uc.disable(ctx, setting, totp); totpResults <- result{method, err} }()
	awaitBlockedByBackend(t, ctx, db, recoveryPID, totpResults)
	if err := blocker.Commit(); err != nil {
		t.Fatalf("ブロッカーのトランザクションのコミットに失敗: %v", err)
	}
	gotRecovery, gotTOTP := <-recoveryResults, <-totpResults
	if gotRecovery.err != nil || gotRecovery.method != TwoFactorAuthDisableMethodRecoveryCode {
		t.Errorf("先行するリカバリーコードの無効化 = %+v、成功を期待", gotRecovery)
	}
	if !isConflictForTest(gotTOTP.err) {
		t.Errorf("後続するTOTPの無効化のエラー = %v、競合を期待", gotTOTP.err)
	}
	f.assertTwoFactorAuthDeleted(t, user.id, true)
}
