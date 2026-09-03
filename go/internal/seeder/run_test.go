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
					t.Fatalf("EnsureSeedableEnv(%q) = nil, want error", tt.env)
				}
				if !strings.Contains(err.Error(), tt.env) {
					t.Errorf("EnsureSeedableEnv(%q) error = %v, want it to name the environment", tt.env, err)
				}

				return
			}
			if err != nil {
				t.Fatalf("EnsureSeedableEnv(%q) error = %v", tt.env, err)
			}
		})
	}
}

// TestRun_RejectsNonSeedableEnv passes a nil database handle on purpose: the guard has
// to reject the environment before Run reaches the database, so a nil handle is never
// dereferenced. A guard that ran too late would panic here rather than pass quietly.
//
// [Ja] TestRun_RejectsNonSeedableEnv は意図的に nil のデータベースハンドルを渡す。ガードは
// Run がデータベースに到達する前に環境を拒否しなければならず、nil のハンドルが参照される
// ことは無い。ガードが遅れて動く実装であれば、静かに通るのではなくここで panic する。
func TestRun_RejectsNonSeedableEnv(t *testing.T) {
	t.Parallel()

	if err := Run(context.Background(), &config.Config{Env: "prod"}, nil); err == nil {
		t.Fatal("Run() = nil, want error")
	}
}
