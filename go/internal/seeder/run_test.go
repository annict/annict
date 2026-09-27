package seeder

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/config"
)

func TestEnsureSeedableEnv(t *testing.T) {
	t.Parallel()

	tests := []struct {
		env     string
		wantErr bool
	}{
		{env: "dev", wantErr: false},
		{env: "test", wantErr: false},
		{env: "prod", wantErr: true},
		{env: "staging", wantErr: true},
		{env: "", wantErr: true},
		{env: "Dev", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			t.Parallel()

			err := EnsureSeedableEnv(&config.Config{Env: tt.env})
			if tt.wantErr {
				if err == nil {
					t.Fatalf("EnsureSeedableEnv(%q)のエラー = nil、期待値 = エラーあり", tt.env)
				}
				if !strings.Contains(err.Error(), tt.env) {
					t.Errorf("EnsureSeedableEnv(%q)のエラー = %v、期待値 = 環境名を含むエラー", tt.env, err)
				}

				return
			}
			if err != nil {
				t.Fatalf("EnsureSeedableEnv(%q)のエラー = %v", tt.env, err)
			}
		})
	}
}

// TestRun_RejectsNonSeedableEnvは意図的にnilのデータベースハンドルを渡す。ガードは
// Runがデータベースに到達する前に環境を拒否しなければならず、nilのハンドルが参照される
// ことは無い。ガードが遅れて動く実装であれば、静かに通るのではなくここでpanicする。
func TestRun_RejectsNonSeedableEnv(t *testing.T) {
	t.Parallel()

	if err := Run(context.Background(), &config.Config{Env: "prod"}, nil); err == nil {
		t.Fatal("Run() = nil、期待値 = エラーあり")
	}
}
