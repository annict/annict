package settings_two_factor_auth

import "testing"

// TestGroupSecretは、手入力用のキーを4文字ずつ空白で区切ることを検証する。
func TestGroupSecret(t *testing.T) {
	t.Parallel()

	tests := []struct {
		secret string
		want   string
	}{
		{secret: "ABCDEFGH", want: "ABCD EFGH"},
		{secret: "ABCDEFGHIJ", want: "ABCD EFGH IJ"},
		{secret: "ABC", want: "ABC"},
	}
	for _, tt := range tests {
		if got := groupSecret(tt.secret); got != tt.want {
			t.Errorf("groupSecret(%q) = %q、期待値 = %q", tt.secret, got, tt.want)
		}
	}
}
