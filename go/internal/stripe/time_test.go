package stripe

import (
	"testing"
)

func TestNullTimeFromUnix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		timestamp int64
		wantValid bool
	}{
		{
			name:      "0の場合はValidがfalse",
			timestamp: 0,
			wantValid: false,
		},
		{
			name:      "正の値の場合はValidがtrue",
			timestamp: 1609459200, // 2021-01-01 00:00:00 UTC
			wantValid: true,
		},
		{
			name:      "負の値の場合もValidがtrue",
			timestamp: -1,
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := NullTimeFromUnix(tt.timestamp)

			if result.Valid != tt.wantValid {
				t.Errorf("Valid: got %v, want %v", result.Valid, tt.wantValid)
			}

			if tt.wantValid && result.Time.Unix() != tt.timestamp {
				t.Errorf("Time.Unix(): got %d, want %d", result.Time.Unix(), tt.timestamp)
			}
		})
	}
}
