package main

import (
	"context"
	"reflect"
	"testing"

	"github.com/annict/annict/go/internal/config"
)

// TestSeedMasterTask_Registeredはseed-masterというタスク名が環境ガードと本体に解決する
// ことを確認する。運用者が打ち込むのはタスク名だけであるため、登録はDB接続前の安全確認と
// 要求された処理の両方を保たなければならない。
func TestSeedMasterTask_Registered(t *testing.T) {
	t.Parallel()

	task, ok := tasks["seed-master"]
	if !ok {
		t.Fatal(`タスク"seed-master"が登録されていない`)
	}
	if task.guard == nil {
		t.Fatal(`タスク"seed-master"にguardが無い`)
	}
	if got, want := reflect.ValueOf(task.guard).Pointer(), reflect.ValueOf(guardSeed).Pointer(); got != want {
		t.Error(`タスク"seed-master"が誤ったguardで登録されている`)
	}
	if got, want := reflect.ValueOf(task.body).Pointer(), reflect.ValueOf(seedMaster).Pointer(); got != want {
		t.Error(`タスク"seed-master"が誤ったbodyで登録されている`)
	}
}

// TestSeedMaster_RejectsNonSeedableEnvは意図的にnilのデータベースハンドルを渡す。投入を
// 行うかどうかを決めるのは本体が受け渡す環境であり、拒否された場合はデータベースに到達する
// 前に戻らなければならない。渡された設定を転送せず自前で環境を読む実装であれば、ここでは
// テスト環境が見えてガードを通過し、nilのハンドルでpanicする。
func TestSeedMaster_RejectsNonSeedableEnv(t *testing.T) {
	t.Parallel()

	if err := seedMaster(context.Background(), &config.Config{Env: "prod"}, nil, nil); err == nil {
		t.Fatal("seedMaster() = nil、期待値 = エラーあり")
	}
}
