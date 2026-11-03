package main

import (
	"context"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/testutil"
)

// TestDisableTwoFactorAuthTask_Registeredは、disable-two-factor-authというタスク名が無効化の本体と、
// ユーザーIDを1つ受け取る引数の定義に解決することを確認する。
func TestDisableTwoFactorAuthTask_Registered(t *testing.T) {
	t.Parallel()

	task, ok := tasks["disable-two-factor-auth"]
	if !ok {
		t.Fatal(`タスク"disable-two-factor-auth"が登録されていない`)
	}
	if got, want := reflect.ValueOf(task.body).Pointer(), reflect.ValueOf(disableTwoFactorAuth).Pointer(); got != want {
		t.Error(`タスク"disable-two-factor-auth"が誤ったbodyで登録されている`)
	}
	if want := []string{"user_id"}; !slices.Equal(task.params, want) {
		t.Errorf(`タスク"disable-two-factor-auth"の引数 = %q、期待値 = %q`, task.params, want)
	}
}

// TestDisableTwoFactorAuth_InvalidUserIDは、正の整数でないユーザーIDを、データベースを使う前に拒否することを確認する。
func TestDisableTwoFactorAuth_InvalidUserID(t *testing.T) {
	t.Parallel()

	for _, arg := range []string{"abc", "0", "-1", ""} {
		err := disableTwoFactorAuth(context.Background(), nil, nil, nil, []string{arg})
		if err == nil || !strings.Contains(err.Error(), "ユーザーIDには正の整数を指定してください") {
			t.Errorf("disableTwoFactorAuth(%q)のエラー = %v、期待値 = ユーザーIDの形式のエラー", arg, err)
		}
	}
}

// TestDisableTwoFactorAuth_Wiringは、タスク本体をテスト用データベースに対して実行し、
// 有効にしていたユーザーの設定とリカバリーコードを削除することを確かめる。
// 無効化のUseCaseはdbから自前のトランザクションを開くため、準備した行はコミットしておく。
func TestDisableTwoFactorAuth_Wiring(t *testing.T) {
	t.Parallel()

	db := testutil.GetTestDB()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("準備用トランザクションの開始に失敗: %v", err)
	}
	userID := testutil.NewUserBuilder(t, tx).Build()
	if err := tx.Commit(); err != nil {
		t.Fatalf("準備用トランザクションのコミットに失敗: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM user_two_factor_recovery_codes WHERE user_id = $1", int64(userID))
		_, _ = db.Exec("DELETE FROM user_two_factor_auths WHERE user_id = $1", int64(userID))
		_, _ = db.Exec("DELETE FROM users WHERE id = $1", int64(userID))
	})
	if _, err := db.Exec("INSERT INTO user_two_factor_auths (user_id, secret_ciphertext, enabled_at) VALUES ($1, $2, NOW())", int64(userID), []byte("secret")); err != nil {
		t.Fatalf("二要素認証の設定の作成に失敗: %v", err)
	}
	if _, err := db.Exec("INSERT INTO user_two_factor_recovery_codes (user_id, code_digest) VALUES ($1, 'digest')", int64(userID)); err != nil {
		t.Fatalf("リカバリーコードの作成に失敗: %v", err)
	}

	// 2回目は無効にする設定が無いが、エラーにはしない。
	for i := range 2 {
		if err := disableTwoFactorAuth(context.Background(), nil, db, query.New(db), []string{userID.String()}); err != nil {
			t.Fatalf("%d回目のdisableTwoFactorAuth()のエラー = %v", i+1, err)
		}
	}

	var rows int
	if err := db.QueryRow(`SELECT (SELECT COUNT(*) FROM user_two_factor_auths WHERE user_id = $1) + (SELECT COUNT(*) FROM user_two_factor_recovery_codes WHERE user_id = $1)`, int64(userID)).Scan(&rows); err != nil {
		t.Fatalf("行数の取得に失敗: %v", err)
	}
	if rows != 0 {
		t.Errorf("残った二要素認証の行数 = %d、期待値 = 0", rows)
	}
}
