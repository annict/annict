package model

import "testing"

// TestTwoFactorAuthNotificationEvent_IsValidは、定義した3つの種類だけを知っている種類として扱うことを検証する。
func TestTwoFactorAuthNotificationEvent_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		event TwoFactorAuthNotificationEvent
		want  bool
	}{
		{event: TwoFactorAuthNotificationEventEnabled, want: true},
		{event: TwoFactorAuthNotificationEventDisabled, want: true},
		{event: TwoFactorAuthNotificationEventDisabledByAdmin, want: true},
		{event: "", want: false},
		{event: "unknown", want: false},
	}
	for _, tt := range tests {
		if got := tt.event.IsValid(); got != tt.want {
			t.Errorf("TwoFactorAuthNotificationEvent(%q).IsValid() = %t、期待値 = %t", tt.event, got, tt.want)
		}
	}
}
