package main

import (
	"testing"
	"time"
)

func TestHourlySchedule_Next(t *testing.T) {
	t.Parallel()

	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	tests := []struct {
		name     string
		schedule hourlySchedule
		current  time.Time
		want     time.Time
	}{
		{
			name:     "ゼロ値: 毎時0分の途中なら次の時の0分",
			schedule: hourlySchedule{},
			current:  time.Date(2026, 9, 27, 10, 15, 0, 0, jst),
			want:     time.Date(2026, 9, 27, 11, 0, 0, 0, jst),
		},
		{
			name:     "ゼロ値: ちょうど0分なら次の時の0分",
			schedule: hourlySchedule{},
			current:  time.Date(2026, 9, 27, 10, 0, 0, 0, jst),
			want:     time.Date(2026, 9, 27, 11, 0, 0, 0, jst),
		},
		{
			name:     "30分: 30分より前なら同じ時の30分",
			schedule: hourlySchedule{minute: 30},
			current:  time.Date(2026, 9, 27, 10, 15, 0, 0, jst),
			want:     time.Date(2026, 9, 27, 10, 30, 0, 0, jst),
		},
		{
			name:     "30分: ちょうど30分なら次の時の30分",
			schedule: hourlySchedule{minute: 30},
			current:  time.Date(2026, 9, 27, 10, 30, 0, 0, jst),
			want:     time.Date(2026, 9, 27, 11, 30, 0, 0, jst),
		},
		{
			name:     "30分: 30分を過ぎていれば次の時の30分",
			schedule: hourlySchedule{minute: 30},
			current:  time.Date(2026, 9, 27, 10, 30, 0, 1, jst),
			want:     time.Date(2026, 9, 27, 11, 30, 0, 0, jst),
		},
		{
			name:     "30分: 23時台に30分を過ぎていれば翌日0時30分",
			schedule: hourlySchedule{minute: 30},
			current:  time.Date(2026, 9, 27, 23, 45, 0, 0, jst),
			want:     time.Date(2026, 9, 28, 0, 30, 0, 0, jst),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.schedule.Next(tt.current); !got.Equal(tt.want) {
				t.Errorf("Next(%v) = %v、期待値 = %v", tt.current, got, tt.want)
			}
		})
	}
}
