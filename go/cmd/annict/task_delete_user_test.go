package main

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
)

// TestDeleteUserTask_Registeredは、delete-userというタスク名が削除の本体と、ユーザーIDを
// 1つ受け取る引数の定義に解決することを確認する。運用者はユーザーIDを渡して実行するため、
// 引数の定義が欠けると、ディスパッチャがユーザーIDを余分な引数として拒否してしまう。
func TestDeleteUserTask_Registered(t *testing.T) {
	t.Parallel()

	task, ok := tasks["delete-user"]
	if !ok {
		t.Fatal(`タスク"delete-user"が登録されていない`)
	}
	if got, want := reflect.ValueOf(task.body).Pointer(), reflect.ValueOf(deleteUser).Pointer(); got != want {
		t.Error(`タスク"delete-user"が誤ったbodyで登録されている`)
	}
	if want := []string{"user_id"}; !slices.Equal(task.params, want) {
		t.Errorf(`タスク"delete-user"の引数 = %q、期待値 = %q`, task.params, want)
	}
}

// TestDeleteUser_InvalidUserIDは、正の整数でないユーザーIDを、削除を始める前に拒否する
// ことを確認する。設定とデータベースをnilにしているのは、値の解釈がそれらを使うより前に
// 行われることも併せて確かめるため。
func TestDeleteUser_InvalidUserID(t *testing.T) {
	t.Parallel()

	for _, arg := range []string{"abc", "0", "-1", "1.5", ""} {
		t.Run(arg, func(t *testing.T) {
			t.Parallel()

			err := deleteUser(context.Background(), nil, nil, nil, []string{arg})
			if err == nil {
				t.Fatal("deleteUser()のエラー = nil、期待値 = エラーあり")
			}
			if !strings.Contains(err.Error(), "ユーザーIDには正の整数を指定してください") {
				t.Errorf("deleteUser()のエラー = %q、期待値 = ユーザーIDの形式のエラー", err)
			}
		})
	}
}

// TestDeleteUser_Wiringは、タスク本体をテスト用データベースに対して実行する。何が削除
// されるかはUseCase側の関心事のため、CLI側では、配線した削除の段階が実スキーマに対して
// 最後まで進むことと、論理削除されていないユーザーを拒否するUseCaseの判定が呼び出し元へ
// 届くことを確かめる。
//
// クエリはテスト用トランザクション上に組み立て、削除がロールバックされてテスト用データベースを
// 共有する他パッケージから見えないようにする。ユーザー行の削除の段階は自前のトランザクションを
// 開くため、このトランザクションで作ったユーザー行は見えず、削除されない。ユーザー行の削除は
// UseCaseのテストが確かめる。
func TestDeleteUser_Wiring(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	queries := testutil.NewQueriesWithTx(db, tx)
	// 作成するユーザーはStripeの購読もプロフィール画像も持たないため、アダプタは外部へ接続しない。
	cfg := &config.Config{}

	deletedUserID := testutil.NewUserBuilder(t, tx).WithDeletedAt(time.Now()).Build()
	if err := deleteUser(context.Background(), cfg, db, queries, []string{deletedUserID.String()}); err != nil {
		t.Errorf("論理削除済みのユーザーでのdeleteUser()のエラー = %v", err)
	}

	activeUserID := testutil.NewUserBuilder(t, tx).Build()
	err := deleteUser(context.Background(), cfg, db, queries, []string{activeUserID.String()})
	if !errors.Is(err, usecase.ErrUserNotDeleted) {
		t.Errorf("論理削除されていないユーザーでのdeleteUser()のエラー = %v、期待値 = %v", err, usecase.ErrUserNotDeleted)
	}
}
