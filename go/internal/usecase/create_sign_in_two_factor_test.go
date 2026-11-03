package usecase

import (
	"context"
	"database/sql"
	"strings"
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

// signInTwoFactorFixtureは、ログインの二要素認証のUseCaseと、照合の相手を用意するのに使うものをまとめる。
// UseCaseはトランザクションを開かないため、準備した行をコミットせずにテスト用トランザクションの中で使える。
type signInTwoFactorFixture struct {
	t        *testing.T
	tx       *sql.Tx
	uc       *CreateSignInTwoFactorUsecase
	authRepo *repository.UserTwoFactorAuthRepository
}

func newSignInTwoFactorFixture(t *testing.T) *signInTwoFactorFixture {
	t.Helper()

	db, tx := testutil.SetupTx(t)
	queries := query.New(db).WithTx(tx)
	key := testutil.NewTestTwoFactorKey(t)
	authRepo := repository.NewUserTwoFactorAuthRepository(queries)
	return &signInTwoFactorFixture{
		t:        t,
		tx:       tx,
		uc:       NewCreateSignInTwoFactorUsecase(key, validator.NewSignInTwoFactorCreateValidator(), repository.NewUserRepository(queries), authRepo),
		authRepo: authRepo,
	}
}

// newUserはパスワードのハッシュを持つユーザーを作る。
func (f *signInTwoFactorFixture) newUser() model.UserID {
	f.t.Helper()
	return testutil.NewUserBuilder(f.t, f.tx).WithEncryptedPassword("encrypted-password").Build()
}

// enableは、testutil.EnableTwoFactorAuthForTestでユーザーの二要素認証を有効にし、秘密鍵を返す。
// UseCaseが照合できることで、testutilの付加データの作り方がUseCaseと食い違わないことも確かめる。
func (f *signInTwoFactorFixture) enable(userID model.UserID) string {
	f.t.Helper()
	return testutil.EnableTwoFactorAuthForTest(f.t, f.tx, userID)
}

// TestCreateSignInTwoFactorUsecaseは、認証アプリのコードが合えばそのタイムステップを使用済みにして、
// セッションの作成に要るユーザーの情報を返すことを検証する。
func TestCreateSignInTwoFactorUsecase(t *testing.T) {
	t.Parallel()

	f := newSignInTwoFactorFixture(t)
	ctx := i18n.SetLocale(context.Background(), "ja")
	userID := f.newUser()
	secret := f.enable(userID)

	now := time.Now()
	code := testutil.TOTPCode(t, secret, now)
	output, err := f.uc.Execute(ctx, CreateSignInTwoFactorInput{UserID: userID, Code: " " + code[:3] + "-" + code[3:]})
	if err != nil {
		t.Fatalf("Execute()のエラー = %v", err)
	}
	if output.UserID != userID || output.EncryptedPassword != "encrypted-password" || output.Username == "" {
		t.Errorf("Execute() = %+v、ユーザーの情報を期待", output)
	}

	saved, err := f.authRepo.FindByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("FindByUserID()のエラー = %v", err)
	}
	if saved.LastUsedStep == 0 {
		t.Error("LastUsedStep = 0、照合したステップの記録を期待")
	}

	// 同じコードは使い回せない。
	_, err = f.uc.Execute(ctx, CreateSignInTwoFactorInput{UserID: userID, Code: code})
	assertCodeFieldError(t, err, "コードが正しくありません。認証アプリに表示されている最新のコードを入力してください")
}

// TestCreateSignInTwoFactorUsecase_InvalidCodeは、形の違うコードと一致しないコードを
// コードの欄のValidationErrorで拒むことを検証する。
func TestCreateSignInTwoFactorUsecase_InvalidCode(t *testing.T) {
	t.Parallel()

	f := newSignInTwoFactorFixture(t)
	ctx := i18n.SetLocale(context.Background(), "ja")
	userID := f.newUser()
	secret := f.enable(userID)

	tests := []struct {
		name    string
		code    string
		wantMsg string
	}{
		{name: "空", code: "", wantMsg: "入力してください"},
		{name: "5桁", code: "12345", wantMsg: "認証アプリに表示された6桁の数字を入力してください"},
		{name: "一致しない", code: testutil.NonMatchingTOTPCode(t, secret, time.Now()), wantMsg: "コードが正しくありません。認証アプリに表示されている最新のコードを入力してください"},
	}
	for _, tt := range tests {
		_, err := f.uc.Execute(ctx, CreateSignInTwoFactorInput{UserID: userID, Code: tt.code})
		assertCodeFieldError(t, err, tt.wantMsg)
	}
}

// TestCreateSignInTwoFactorUsecase_Unavailableは、第1の要素を確かめたあとに二要素認証を無効にした・
// 登録の途中しか無い・退会したユーザーにAppErrCodeConflictを返し、ログインをやり直させることを検証する。
func TestCreateSignInTwoFactorUsecase_Unavailable(t *testing.T) {
	t.Parallel()

	f := newSignInTwoFactorFixture(t)
	ctx := i18n.SetLocale(context.Background(), "ja")

	noSettingUserID := f.newUser()
	pendingUserID := f.newUser()
	if _, err := f.authRepo.UpsertPending(ctx, pendingUserID, []byte("pending-secret")); err != nil {
		t.Fatalf("UpsertPending()のエラー = %v", err)
	}
	deletedUserID := f.newUser()
	deletedUserSecret := f.enable(deletedUserID)
	if _, err := f.tx.ExecContext(ctx, "UPDATE users SET deleted_at = NOW() WHERE id = $1", int64(deletedUserID)); err != nil {
		t.Fatalf("ユーザーの論理削除に失敗: %v", err)
	}

	tests := []struct {
		name   string
		userID model.UserID
		code   string
	}{
		{name: "設定が無い", userID: noSettingUserID, code: "123456"},
		{name: "登録の途中", userID: pendingUserID, code: "123456"},
		{name: "退会した", userID: deletedUserID, code: testutil.TOTPCode(t, deletedUserSecret, time.Now())},
	}
	for _, tt := range tests {
		if _, err := f.uc.Execute(ctx, CreateSignInTwoFactorInput{UserID: tt.userID, Code: tt.code}); !isConflictForTest(err) {
			t.Errorf("%s: Execute()のエラー = %v、AppErrCodeConflictを期待", tt.name, err)
		}
	}
}

// signInTwoFactorStepDBは、ステップのUPDATEの直前に別の操作を挟み、照合後の状態変更を再現する。
// SQLの実行はすべて実DBへ渡し、テスト用トランザクションの終了時にロールバックする。
type signInTwoFactorStepDB struct {
	query.DBTX
	beforeUpdate func()
}

func (d *signInTwoFactorStepDB) ExecContext(ctx context.Context, statement string, args ...any) (sql.Result, error) {
	if strings.Contains(statement, "UPDATE user_two_factor_auths") && d.beforeUpdate != nil {
		beforeUpdate := d.beforeUpdate
		d.beforeUpdate = nil
		beforeUpdate()
	}
	return d.DBTX.ExecContext(ctx, statement, args...)
}

// TestCreateSignInTwoFactorUsecase_SettingChangedは、コードの照合後に設定が削除・再登録されたときに、
// 古い設定のコードでログインさせず、再登録した設定のステップを変えないことを検証する。
func TestCreateSignInTwoFactorUsecase_SettingChanged(t *testing.T) {
	t.Parallel()

	for _, state := range []string{"削除", "再登録", "登録の途中"} {
		t.Run(state, func(t *testing.T) {
			t.Parallel()

			f := newSignInTwoFactorFixture(t)
			ctx := context.Background()
			userID := f.newUser()
			secret := f.enable(userID)
			now := time.Now()
			code := testutil.TOTPCode(t, secret, now.Add(30*time.Second))
			var replacement *model.UserTwoFactorAuth
			changed := false
			db := &signInTwoFactorStepDB{DBTX: f.tx}
			db.beforeUpdate = func() {
				previous, err := f.authRepo.DeleteByUserID(ctx, userID)
				if err != nil || previous == nil {
					t.Fatalf("設定の削除 = (%+v, %v)、元の設定の削除を期待", previous, err)
				}
				switch state {
				case "再登録":
					// 元のコードが新しい秘密鍵に偶然一致しないようにする。
					var newSecret string
					for {
						newSecret, err = auth.GenerateTOTPSecret()
						if err != nil {
							t.Fatalf("秘密鍵の生成に失敗: %v", err)
						}
						if _, matches := auth.MatchTOTPCode(newSecret, code, time.Now()); !matches {
							break
						}
					}
					ciphertext, err := testutil.NewTestTwoFactorKey(t).EncryptTOTPSecret(newSecret, []byte(userID.String()))
					if err != nil {
						t.Fatalf("秘密鍵の暗号化に失敗: %v", err)
					}
					testutil.NewUserTwoFactorAuthBuilder(t, f.tx, userID).
						WithSecretCiphertext(ciphertext).
						WithEnabledAt(now).
						WithLastUsedStep(now.Unix() / 30).
						Build()
				case "登録の途中":
					if _, err := f.authRepo.UpsertPending(ctx, userID, []byte("new-pending-secret")); err != nil {
						t.Fatalf("登録途中の設定の作成に失敗: %v", err)
					}
				}
				replacement, err = f.authRepo.FindByUserID(ctx, userID)
				if err != nil {
					t.Fatalf("差し替えた設定の取得に失敗: %v", err)
				}
				if replacement != nil && replacement.ID == previous.ID {
					t.Fatal("設定のIDが入れ替わっていない")
				}
				changed = true
			}
			uc := NewCreateSignInTwoFactorUsecase(
				testutil.NewTestTwoFactorKey(t),
				validator.NewSignInTwoFactorCreateValidator(),
				repository.NewUserRepository(query.New(f.tx)),
				repository.NewUserTwoFactorAuthRepository(query.New(db)),
			)

			output, err := uc.Execute(ctx, CreateSignInTwoFactorInput{UserID: userID, Code: code})
			if !changed {
				t.Fatal("照合後のステップの記録に到達しなかった")
			}
			if output != nil || !isConflictForTest(err) {
				t.Errorf("設定が変わった後のExecute() = (%+v, %v)、結果なしとConflictを期待", output, err)
			}
			saved, err := f.authRepo.FindByUserID(ctx, userID)
			if err != nil {
				t.Fatalf("照合後の設定の取得に失敗: %v", err)
			}
			if replacement == nil {
				if saved != nil {
					t.Errorf("削除した設定が復元された: %+v", saved)
				}
			} else if saved == nil || saved.ID != replacement.ID || saved.LastUsedStep != replacement.LastUsedStep {
				t.Errorf("差し替えた設定のステップが変わった: 変更前=%+v、照合後=%+v", replacement, saved)
			}
		})
	}
}

// assertCodeFieldErrorは、エラーがコードの欄に1つだけメッセージを持つValidationErrorであることを確かめる。
func assertCodeFieldError(t *testing.T, err error, wantMsg string) {
	t.Helper()

	ve := model.AsValidationError(err)
	if ve == nil {
		t.Fatalf("エラー = %v、ValidationErrorを期待", err)
	}
	if got := ve.GetFieldErrors("code"); len(got) != 1 || got[0] != wantMsg {
		t.Errorf("codeのフィールドエラー = %v、期待値 = [%q]", got, wantMsg)
	}
}
