package api_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestOASDiffCompatibility(t *testing.T) {
	t.Parallel()

	// 列挙値追加の許可と、削除・型変更の拒否を同じAPIの差分で確かめる。
	const description = `{
  "openapi": "3.1.1",
  "info": {"title": "互換性検査用API", "version": "2.0.0"},
  "paths": {
    "/probe": {
      "get": {
        "responses": {
          "200": {
            "description": "成功",
            "content": {"application/json": {"schema": %s}}
          }
        }
      }
    }
  }
}`
	const baseSchema = `{"type":"object","required":["state"],"properties":{"state":{"type":"string","enum":["pending"]}}}`
	basePath := filepath.Join(t.TempDir(), "base.json")
	if err := os.WriteFile(basePath, []byte(fmt.Sprintf(description, baseSchema)), 0o600); err != nil {
		t.Fatalf("基準の仕様を書き出せない: %v", err)
	}

	tests := []struct {
		name     string
		schema   string
		wantExit int
		wantRule string
	}{
		{
			name:     "レスポンスの列挙値追加は警告で通る",
			schema:   `{"type":"object","required":["state"],"properties":{"state":{"type":"string","enum":["pending","released"]}}}`,
			wantExit: 0,
			wantRule: "response-property-enum-value-added",
		},
		{
			name:     "必須レスポンスフィールドの削除は失敗する",
			schema:   `{"type":"object","properties":{}}`,
			wantExit: 1,
			wantRule: "response-required-property-removed",
		},
		{
			name:     "レスポンスフィールドの型変更は失敗する",
			schema:   `{"type":"object","required":["state"],"properties":{"state":{"type":"integer"}}}`,
			wantExit: 1,
			wantRule: "response-property-type-changed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			revisionPath := filepath.Join(t.TempDir(), "revision.json")
			if err := os.WriteFile(revisionPath, []byte(fmt.Sprintf(description, tt.schema)), 0o600); err != nil {
				t.Fatalf("変更後の仕様を書き出せない: %v", err)
			}

			cmd := exec.CommandContext(t.Context(), "go", "tool", "oasdiff", "breaking",
				"--fail-on", "ERR",
				"--severity-levels", "oasdiff-levels.txt",
				"--err-ignore", "oasdiff-err-ignore.txt",
				basePath, revisionPath,
			)
			output, err := cmd.CombinedOutput()
			exitCode := 0
			if err != nil {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) {
					t.Fatalf("oasdiffを実行できない: %v", err)
				}
				exitCode = exitErr.ExitCode()
			}
			if exitCode != tt.wantExit {
				t.Errorf("終了コード = %d、期待値 = %d\n%s", exitCode, tt.wantExit, output)
			}
			if !strings.Contains(string(output), tt.wantRule) {
				t.Errorf("出力に検査ルール %qが含まれていない:\n%s", tt.wantRule, output)
			}
		})
	}
}
