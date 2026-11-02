package repository_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// createPendingTwoFactorAuthはユーザーの登録の途中の二要素認証の設定を作る。
func createPendingTwoFactorAuth(t *testing.T, repo *repository.UserTwoFactorAuthRepository, userID model.UserID, secretCiphertext []byte) *model.UserTwoFactorAuth {
	t.Helper()

	setting, err := repo.UpsertPending(context.Background(), userID, secretCiphertext)
	if err != nil || setting == nil {
		t.Fatalf("UpsertPending() = (%+v, %v)、設定とnilを期待", setting, err)
	}

	return setting
}

// TestUserTwoFactorAuthRepository_UpsertPendingは、登録の途中の設定を作り、開き直したときは秘密鍵を差し替え、
// 有効にした設定は書き換えないことを検証する。
func TestUserTwoFactorAuthRepository_UpsertPending(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewUserTwoFactorAuthRepository(query.New(db).WithTx(tx))
	userID := testutil.NewUserBuilder(t, tx).Build()
	ctx := context.Background()

	created := createPendingTwoFactorAuth(t, repo, userID, []byte("first"))
	if created.UserID != userID || !bytes.Equal(created.SecretCiphertext, []byte("first")) || created.IsEnabled() || created.LastUsedStep != 0 {
		t.Errorf("作った設定 = %+v、入力のユーザーと秘密鍵の登録の途中の設定を期待", created)
	}

	replaced := createPendingTwoFactorAuth(t, repo, userID, []byte("second"))
	if replaced.ID != created.ID || !bytes.Equal(replaced.SecretCiphertext, []byte("second")) {
		t.Errorf("差し替えた設定 = %+v、同じ行で秘密鍵が second になることを期待", replaced)
	}

	if ok, err := repo.Enable(ctx, userID, replaced.SecretCiphertext, 100); err != nil || !ok {
		t.Fatalf("Enable() = (%t, %v)、期待値 = (true, nil)", ok, err)
	}
	if got, err := repo.UpsertPending(ctx, userID, []byte("third")); err != nil || got != nil {
		t.Errorf("有効にした後のUpsertPending() = (%+v, %v)、期待値 = (nil, nil)", got, err)
	}

	found, err := repo.FindByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("FindByUserID()のエラー = %v", err)
	}
	if !bytes.Equal(found.SecretCiphertext, []byte("second")) {
		t.Errorf("有効にした設定の秘密鍵 = %q、期待値 = %q", found.SecretCiphertext, "second")
	}
}

// TestUserTwoFactorAuthRepository_FindByUserIDは、設定の無いユーザーにnilを返すことを検証する。
func TestUserTwoFactorAuthRepository_FindByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewUserTwoFactorAuthRepository(query.New(db).WithTx(tx))
	userID := testutil.NewUserBuilder(t, tx).Build()

	got, err := repo.FindByUserID(context.Background(), userID)
	if err != nil || got != nil {
		t.Errorf("FindByUserID() = (%+v, %v)、期待値 = (nil, nil)", got, err)
	}
}

// TestUserTwoFactorAuthRepository_FindByUserID_Enabledは、有効な設定を列の値どおりのモデルで返すことを検証する。
func TestUserTwoFactorAuthRepository_FindByUserID_Enabled(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewUserTwoFactorAuthRepository(query.New(db).WithTx(tx))
	userID := testutil.NewUserBuilder(t, tx).Build()
	enabledAt := time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)
	id := testutil.NewUserTwoFactorAuthBuilder(t, tx, userID).
		WithSecretCiphertext([]byte("enabled-secret")).
		WithLastUsedStep(42).
		WithEnabledAt(enabledAt).
		Build()

	got, err := repo.FindByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("FindByUserID()のエラー = %v", err)
	}
	if got == nil {
		t.Fatal("FindByUserID() = nil、設定を期待")
	}
	if got.ID != id || got.UserID != userID || !bytes.Equal(got.SecretCiphertext, []byte("enabled-secret")) || got.LastUsedStep != 42 {
		t.Errorf("FindByUserID() = %+v、ビルダーで作った値を期待", got)
	}
	if !got.IsEnabled() || !got.EnabledAt.Time.Equal(enabledAt) {
		t.Errorf("EnabledAt = %v、期待値 = %v (有効)", got.EnabledAt, enabledAt)
	}
}

// TestUserTwoFactorAuthRepository_DeletedWithUserは、ユーザーの行を削除すると設定とリカバリーコードも消えることを検証する。
// 退会のジョブはusersの行を削除するだけで、二要素認証の行の後始末を外部キーのON DELETE CASCADEに任せている。
func TestUserTwoFactorAuthRepository_DeletedWithUser(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	repo := repository.NewUserTwoFactorAuthRepository(queries)
	recoveryCodeRepo := repository.NewUserTwoFactorRecoveryCodeRepository(queries)
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	testutil.NewUserTwoFactorAuthBuilder(t, tx, userID).WithEnabledAt(time.Now()).Build()
	if err := recoveryCodeRepo.CreateAll(ctx, userID, []string{"digest-1"}); err != nil {
		t.Fatalf("CreateAll()のエラー = %v", err)
	}

	// NewUserBuilderが併せて作る行は外部キーにON DELETEが無いため、先に消しておく。
	for _, statement := range []string{
		`DELETE FROM email_notifications WHERE user_id = $1`,
		`DELETE FROM settings WHERE user_id = $1`,
		`DELETE FROM profiles WHERE user_id = $1`,
		`DELETE FROM users WHERE id = $1`,
	} {
		if _, err := tx.ExecContext(ctx, statement, int64(userID)); err != nil {
			t.Fatalf("ユーザーの削除に失敗 (%s): %v", statement, err)
		}
	}

	if got := testutil.CountRows(t, tx, "user_two_factor_auths", "user_id = $1", int64(userID)); got != 0 {
		t.Errorf("user_two_factor_authsの行数 = %d、期待値 = 0", got)
	}
	if got := testutil.CountRows(t, tx, "user_two_factor_recovery_codes", "user_id = $1", int64(userID)); got != 0 {
		t.Errorf("user_two_factor_recovery_codesの行数 = %d、期待値 = 0", got)
	}
	if got, err := repo.FindByUserID(ctx, userID); err != nil || got != nil {
		t.Errorf("FindByUserID() = (%+v, %v)、期待値 = (nil, nil)", got, err)
	}
}

// TestUserTwoFactorAuthRepository_Enableは、登録の途中の設定だけを有効にして照合したステップを記録し、
// 設定が無い・既に有効・秘密鍵が違うときは更新しないことを検証する。
func TestUserTwoFactorAuthRepository_Enable(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewUserTwoFactorAuthRepository(query.New(db).WithTx(tx))
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	setting := createPendingTwoFactorAuth(t, repo, userID, []byte("secret"))

	if ok, err := repo.Enable(ctx, userID, []byte("other"), 100); err != nil || ok {
		t.Errorf("秘密鍵が違うEnable() = (%t, %v)、期待値 = (false, nil)", ok, err)
	}

	for i, want := range []bool{true, false} {
		ok, err := repo.Enable(ctx, userID, setting.SecretCiphertext, 100)
		if err != nil || ok != want {
			t.Errorf("%d回目のEnable() = (%t, %v)、期待値 = (%t, nil)", i+1, ok, err, want)
		}
	}

	found, err := repo.FindByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("FindByUserID()のエラー = %v", err)
	}
	if !found.IsEnabled() || found.LastUsedStep != 100 {
		t.Errorf("有効にした設定 = %+v、有効でステップ100を記録していることを期待", found)
	}

	noAuthUserID := testutil.NewUserBuilder(t, tx).Build()
	if ok, err := repo.Enable(ctx, noAuthUserID, setting.SecretCiphertext, 100); err != nil || ok {
		t.Errorf("設定の無いユーザーのEnable() = (%t, %v)、期待値 = (false, nil)", ok, err)
	}
}

// TestUserTwoFactorAuthRepository_Enable_SecretReplacedは、コード照合後に秘密鍵が差し替わったときに
// 古い秘密鍵での有効化を拒み、現在の秘密鍵は有効化できることを検証する。
func TestUserTwoFactorAuthRepository_Enable_SecretReplaced(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewUserTwoFactorAuthRepository(query.New(db).WithTx(tx))
	userID := testutil.NewUserBuilder(t, tx).Build()
	ctx := context.Background()

	previous := createPendingTwoFactorAuth(t, repo, userID, []byte("previous"))
	current := createPendingTwoFactorAuth(t, repo, userID, []byte("current"))

	if ok, err := repo.Enable(ctx, userID, previous.SecretCiphertext, 100); err != nil || ok {
		t.Errorf("差し替え前の秘密鍵でのEnable() = (%t, %v)、期待値 = (false, nil)", ok, err)
	}

	found, err := repo.FindByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("FindByUserID()のエラー = %v", err)
	}
	if found.IsEnabled() || found.LastUsedStep != 0 {
		t.Errorf("古い秘密鍵での照合後の設定 = %+v、登録の途中でステップ未使用を期待", found)
	}

	if ok, err := repo.Enable(ctx, userID, current.SecretCiphertext, 100); err != nil || !ok {
		t.Errorf("現在の秘密鍵でのEnable() = (%t, %v)、期待値 = (true, nil)", ok, err)
	}
}

// TestUserTwoFactorAuthRepository_DeleteByUserIDは、指定したユーザーの設定だけを消して削除した設定を返し、
// 消す設定が無いときはnilを返すことを検証する。
func TestUserTwoFactorAuthRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewUserTwoFactorAuthRepository(query.New(db).WithTx(tx))
	ctx := context.Background()

	target := testutil.NewUserBuilder(t, tx).Build()
	other := testutil.NewUserBuilder(t, tx).Build()
	created := createPendingTwoFactorAuth(t, repo, target, []byte("target"))
	createPendingTwoFactorAuth(t, repo, other, []byte("other"))
	if enabled, err := repo.Enable(ctx, target, created.SecretCiphertext, 100); err != nil || !enabled {
		t.Fatalf("Enable() = (%t, %v)、期待値 = (true, nil)", enabled, err)
	}

	for i, want := range []bool{true, false} {
		deleted, err := repo.DeleteByUserID(ctx, target)
		if err != nil || (deleted != nil) != want {
			t.Errorf("%d回目のDeleteByUserID() = (%+v, %v)、削除した設定の有無 = %tを期待", i+1, deleted, err, want)
		}
		if deleted != nil && (deleted.ID != created.ID || deleted.UserID != target || !deleted.IsEnabled() || deleted.LastUsedStep != 100) {
			t.Errorf("削除した設定 = %+v、対象ユーザーの有効な設定を期待", deleted)
		}
	}

	if got, err := repo.FindByUserID(ctx, target); err != nil || got != nil {
		t.Errorf("対象のユーザーのFindByUserID() = (%+v, %v)、期待値 = (nil, nil)", got, err)
	}
	if got, err := repo.FindByUserID(ctx, other); err != nil || got == nil {
		t.Errorf("ほかのユーザーのFindByUserID() = (%+v, %v)、設定が残ることを期待", got, err)
	}
}

// TestUserTwoFactorAuthRepository_MatchingOperationsは、照合した有効な設定とID・持ち主・暗号文が一致するときだけ
// ステップを記録・設定を削除し、登録の途中の設定や使用済みのステップを拒むことを検証する。
func TestUserTwoFactorAuthRepository_MatchingOperations(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewUserTwoFactorAuthRepository(query.New(db).WithTx(tx))
	ctx := context.Background()
	userID := testutil.NewUserBuilder(t, tx).Build()
	pending := createPendingTwoFactorAuth(t, repo, userID, []byte("secret"))
	if used, err := repo.UseMatchingStep(ctx, pending, 100); err != nil || used {
		t.Fatalf("登録の途中のUseMatchingStep() = (%t, %v)、期待値 = (false, nil)", used, err)
	}
	if deleted, err := repo.DeleteMatching(ctx, pending); err != nil || deleted {
		t.Fatalf("登録の途中のDeleteMatching() = (%t, %v)、期待値 = (false, nil)", deleted, err)
	}
	if enabled, err := repo.Enable(ctx, userID, pending.SecretCiphertext, 100); err != nil || !enabled {
		t.Fatalf("Enable() = (%t, %v)、期待値 = (true, nil)", enabled, err)
	}
	setting, err := repo.FindByUserID(ctx, userID)
	if err != nil || setting == nil {
		t.Fatalf("FindByUserID() = (%+v, %v)、設定を期待", setting, err)
	}
	otherID := testutil.NewUserBuilder(t, tx).Build()
	for _, change := range []func(*model.UserTwoFactorAuth){
		func(s *model.UserTwoFactorAuth) { s.ID++ },
		func(s *model.UserTwoFactorAuth) { s.UserID = otherID },
		func(s *model.UserTwoFactorAuth) { s.SecretCiphertext = []byte("different") },
	} {
		mismatched := *setting
		change(&mismatched)
		if used, err := repo.UseMatchingStep(ctx, &mismatched, 101); err != nil || used {
			t.Errorf("一致しない設定のUseMatchingStep() = (%t, %v)、期待値 = (false, nil)", used, err)
		}
		if deleted, err := repo.DeleteMatching(ctx, &mismatched); err != nil || deleted {
			t.Errorf("一致しない設定のDeleteMatching() = (%t, %v)、期待値 = (false, nil)", deleted, err)
		}
	}
	for _, tt := range []struct {
		step int64
		want bool
	}{{100, false}, {99, false}, {101, true}, {101, false}} {
		if used, err := repo.UseMatchingStep(ctx, setting, tt.step); err != nil || used != tt.want {
			t.Errorf("UseMatchingStep(%d) = (%t, %v)、期待値 = (%t, nil)", tt.step, used, err, tt.want)
		}
	}
	for _, want := range []bool{true, false} {
		if deleted, err := repo.DeleteMatching(ctx, setting); err != nil || deleted != want {
			t.Errorf("DeleteMatching() = (%t, %v)、期待値 = (%t, nil)", deleted, err, want)
		}
	}
}

// TestUserTwoFactorAuthRepository_LockMatchingは、照合した有効な設定の行だけをロックしてtrueを返し、
// 設定が入れ替わった・登録の途中・削除されたときはfalseを返すことを検証する。
func TestUserTwoFactorAuthRepository_LockMatching(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewUserTwoFactorAuthRepository(query.New(db).WithTx(tx))
	ctx := context.Background()

	userID := testutil.NewUserBuilder(t, tx).Build()
	testutil.NewUserTwoFactorAuthBuilder(t, tx, userID).WithSecretCiphertext([]byte("secret")).WithEnabledAt(time.Now()).Build()
	setting, err := repo.FindByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("FindByUserID()のエラー = %v", err)
	}

	if locked, err := repo.LockMatching(ctx, setting); err != nil || !locked {
		t.Errorf("照合した設定のLockMatching() = (%t, %v)、期待値 = (true, nil)", locked, err)
	}

	replaced := *setting
	replaced.SecretCiphertext = []byte("other-secret")
	if locked, err := repo.LockMatching(ctx, &replaced); err != nil || locked {
		t.Errorf("秘密鍵が違う設定のLockMatching() = (%t, %v)、期待値 = (false, nil)", locked, err)
	}

	pendingUserID := testutil.NewUserBuilder(t, tx).Build()
	testutil.NewUserTwoFactorAuthBuilder(t, tx, pendingUserID).Build()
	pending, err := repo.FindByUserID(ctx, pendingUserID)
	if err != nil {
		t.Fatalf("FindByUserID()のエラー = %v", err)
	}
	if locked, err := repo.LockMatching(ctx, pending); err != nil || locked {
		t.Errorf("登録の途中の設定のLockMatching() = (%t, %v)、期待値 = (false, nil)", locked, err)
	}

	if _, err := repo.DeleteByUserID(ctx, userID); err != nil {
		t.Fatalf("DeleteByUserID()のエラー = %v", err)
	}
	if locked, err := repo.LockMatching(ctx, setting); err != nil || locked {
		t.Errorf("削除した設定のLockMatching() = (%t, %v)、期待値 = (false, nil)", locked, err)
	}
}
