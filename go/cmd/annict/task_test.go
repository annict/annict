package main

import (
	"bytes"
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"
)

// taskNamePattern is the naming convention for task names: lowercase alphanumerics in
// hyphen-separated words. The names are typed by operators on a shell, so they stay in
// one shape rather than mixing underscores and camelCase.
//
// [Ja] taskNamePattern はタスク名の命名規則で、英小文字と数字の語をハイフンで繋いだ形。
// 運用者がシェルで打ち込む名前であるため、アンダースコアやキャメルケースを混在させず
// 形を 1 つに揃える。
var taskNamePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

func TestTasks_Registry(t *testing.T) {
	t.Parallel()

	for name, task := range tasks {
		if !taskNamePattern.MatchString(name) {
			t.Errorf("task name %q does not match %s", name, taskNamePattern)
		}
		if name == listTaskName {
			t.Errorf("task name %q is reserved for the listing command", name)
		}
		if task.desc == "" {
			t.Errorf("task %q has no description", name)
		}
		if task.body == nil {
			t.Errorf("task %q has no body", name)
		}
		if task.run == nil {
			t.Errorf("task %q has no run function", name)
		}
	}
}

func TestRunTask_List(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	if err := runTask(context.Background(), &out, []string{"list"}, tasks); err != nil {
		t.Fatalf("runTask() error = %v", err)
	}

	listing := out.String()
	if !strings.Contains(listing, listTaskName) {
		t.Errorf("listing does not contain %q:\n%s", listTaskName, listing)
	}
	for name, task := range tasks {
		if !strings.Contains(listing, name) {
			t.Errorf("listing does not contain task %q:\n%s", name, listing)
		}
		if !strings.Contains(listing, task.desc) {
			t.Errorf("listing does not contain the description of task %q:\n%s", name, listing)
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
			wantErrs: []string{`タスク "sync-animes" は引数を取りません`},
		},
		{
			name:     "list の後ろに余分な引数",
			args:     []string{"list", "extra"},
			wantErrs: []string{`タスク "list" は引数を取りません`},
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
				t.Fatal("runTask() error = nil, want error")
			}
			for _, want := range tt.wantErrs {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("runTask() error = %q, want it to contain %q", err, want)
				}
			}
			// The error carries the usage text, so the caller does not need a second run.
			//
			// [Ja] エラーは usage テキストを含むため、呼び出し元が再実行しなくても済む。
			if !strings.Contains(err.Error(), "使い方: annict task <name>") {
				t.Errorf("runTask() error = %q, want it to contain the task usage", err)
			}
			if out.Len() != 0 {
				t.Errorf("runTask() wrote %q to out, want nothing", out.String())
			}
		})
	}
}

func TestRun_TaskList(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	if err := run(context.Background(), &out, []string{"task", "list"}, tasks); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	if got, want := out.String(), taskListing(tasks); got != want {
		t.Errorf("run() output = %q, want %q", got, want)
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
					t.Errorf("task context = %v, want %v", gotCtx, ctx)
				}
				return wantErr
			},
		},
	}

	err := run(ctx, &bytes.Buffer{}, []string{"task", "test-task"}, registry)
	if !called {
		t.Error("task was not called")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("run() error = %v, want %v", err, wantErr)
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
				t.Fatal("run() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("run() error = %q, want it to contain %q", err, tt.wantErr)
			}
			if !strings.Contains(err.Error(), "使い方: annict <command>") {
				t.Errorf("run() error = %q, want it to contain the usage", err)
			}
			for name := range tasks {
				if !strings.Contains(err.Error(), name) {
					t.Errorf("run() error = %q, want it to list task %q", err, name)
				}
			}
		})
	}
}
