package worker

import (
	"testing"

	"github.com/riverqueue/river"
)

// TestArgsInsertOptsは各ジョブArgsのInsertOptsメソッドをテストします
func TestArgsInsertOpts(t *testing.T) {
	tests := []struct {
		name            string
		opts            river.InsertOpts
		wantQueue       string
		wantMaxAttempts int
	}{
		{
			name:            "CleanupExpiredTokensArgs",
			opts:            CleanupExpiredTokensArgs{}.InsertOpts(),
			wantQueue:       river.QueueDefault,
			wantMaxAttempts: 3,
		},
		{
			name:            "CleanupExpiredSignInCodesArgs",
			opts:            CleanupExpiredSignInCodesArgs{}.InsertOpts(),
			wantQueue:       river.QueueDefault,
			wantMaxAttempts: 3,
		},
		{
			name:            "CleanupExpiredSessionsArgs",
			opts:            CleanupExpiredSessionsArgs{}.InsertOpts(),
			wantQueue:       river.QueueDefault,
			wantMaxAttempts: 3,
		},
		{
			name:            "SyncAnimesArgs",
			opts:            SyncAnimesArgs{}.InsertOpts(),
			wantQueue:       river.QueueDefault,
			wantMaxAttempts: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.opts.Queue != tt.wantQueue {
				t.Errorf("Queue = %q、期待値 = %q", tt.opts.Queue, tt.wantQueue)
			}
			if tt.opts.MaxAttempts != tt.wantMaxAttempts {
				t.Errorf("MaxAttempts = %d、期待値 = %d", tt.opts.MaxAttempts, tt.wantMaxAttempts)
			}
		})
	}
}
