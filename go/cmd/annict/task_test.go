package main

import (
	"bytes"
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"
)

// taskNamePatternはタスク名の命名規則で、英小文字と数字の語をハイフンで繋いだ形。
// 運用者がシェルで打ち込む名前であるため、アンダースコアやキャメルケースを混在させず
// 形を1つに揃える。
var taskNamePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

func TestTasks_Registry(t *testing.T) {
	t.Parallel()

	for name, task := range tasks {
		if !taskNamePattern.MatchString(name) {
			t.Errorf("タスク名%qが%sに一致しない", name, taskNamePattern)
		}
		if name == listTaskName {
			t.Errorf("タスク名%qは一覧表示コマンドの予約語", name)
		}
		if task.desc == "" {
			t.Errorf("タスク%qに説明が無い", name)
		}
		if task.body == nil {
			t.Errorf("タスク%qにbodyが無い", name)
		}
		if task.run == nil {
			t.Errorf("タスク%qにrun関数が無い", name)
		}
	}
}

func TestRunTask_List(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	if err := runTask(context.Background(), &out, []string{"list"}, tasks); err != nil {
		t.Fatalf("runTask()のエラー = %v", err)
	}

	listing := out.String()
	if !strings.Contains(listing, listTaskName) {
		t.Errorf("一覧に含まれていない文字列 = %q:\n%s", listTaskName, listing)
	}
	for name, task := range tasks {
		if !strings.Contains(listing, name) {
			t.Errorf("一覧に含まれていないタスク = %q:\n%s", name, listing)
		}
		if !strings.Contains(listing, task.desc) {
			t.Errorf("一覧に含まれていないタスク%qの説明:\n%s", name, listing)
		}
	}
}

func TestRunTask_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		wantErrs []string
	}{
		{
			name:     "タスク名なし",
			args:     nil,
			wantErrs: []string{"タスク名を指定してください"},
		},
		{
			name:     "未知のタスク名",
			args:     []string{"no-such-task"},
			wantErrs: []string{`不明なタスクです: "no-such-task"`},
		},
		{
			name:     "タスク名の後ろに余分な引数",
			args:     []string{"sync-animes", "--dry-run"},
			wantErrs: []string{`タスク "sync-animes"は引数を取りません`},
		},
		{
			name:     "listの後ろに余分な引数",
			args:     []string{"list", "extra"},
			wantErrs: []string{`タスク "list"は引数を取りません`},
		},
		{
			name:     "未知のタスク名の後ろに余分な引数",
			args:     []string{"no-such-task", "--dry-run"},
			wantErrs: []string{`不明なタスクです: "no-such-task"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer
			err := runTask(context.Background(), &out, tt.args, tasks)
			if err == nil {
				t.Fatal("runTask()のエラー = nil、期待値 = エラーあり")
			}
			for _, want := range tt.wantErrs {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("runTask()のエラー = %q、期待値 = %qを含むこと", err, want)
				}
			}
			// エラーはusageテキストを含むため、呼び出し元が再実行しなくても済む。
			if !strings.Contains(err.Error(), "使い方: annict task <name>") {
				t.Errorf("runTask()のエラー = %q、期待値 = タスクの使い方を含むこと", err)
			}
			if out.Len() != 0 {
				t.Errorf("runTask()がoutへ書き出した内容 = %q、期待値 = 空", out.String())
			}
		})
	}
}

func TestRun_TaskList(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	if err := run(context.Background(), &out, []string{"task", "list"}, tasks); err != nil {
		t.Fatalf("run()のエラー = %v", err)
	}

	if got, want := out.String(), taskListing(tasks); got != want {
		t.Errorf("run()がoutへ書き出した内容 = %q、期待値 = %q", got, want)
	}
}

func TestRun_TaskDispatch(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wantErr := errors.New("task failed")
	called := false
	registry := map[string]taskDef{
		"test-task": {
			desc: "test task",
			run: func(gotCtx context.Context) error {
				called = true
				if gotCtx != ctx {
					t.Errorf("タスクに渡されたcontext = %v、期待値 = %v", gotCtx, ctx)
				}
				return wantErr
			},
		},
	}

	err := run(ctx, &bytes.Buffer{}, []string{"task", "test-task"}, registry)
	if !called {
		t.Error("タスクが呼ばれなかった")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("run()のエラー = %v、期待値 = %v", err, wantErr)
	}
}

func TestRun_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "サブコマンドなし",
			args:    nil,
			wantErr: "サブコマンドを指定してください",
		},
		{
			name:    "未知のサブコマンド",
			args:    []string{"no-such-command"},
			wantErr: `不明なサブコマンドです: "no-such-command"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer
			err := run(context.Background(), &out, tt.args, tasks)
			if err == nil {
				t.Fatal("run()のエラー = nil、期待値 = エラーあり")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("run()のエラー = %q、期待値 = %qを含むこと", err, tt.wantErr)
			}
			if !strings.Contains(err.Error(), "使い方: annict <command>") {
				t.Errorf("run()のエラー = %q、期待値 = 使い方を含むこと", err)
			}
			for name := range tasks {
				if !strings.Contains(err.Error(), name) {
					t.Errorf("run()のエラー = %q、期待値 = タスク%qを含むこと", err, name)
				}
			}
		})
	}
}
