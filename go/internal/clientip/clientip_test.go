package clientip

import (
	"net/http/httptest"
	"testing"
)

func TestGetClientIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		cfConnectingIP string
		xForwardedFor  string
		xRealIP        string
		remoteAddr     string
		expected       string
	}{
		{
			name:           "CF-Connecting-IPが優先される",
			cfConnectingIP: "1.2.3.4",
			xForwardedFor:  "5.6.7.8",
			xRealIP:        "9.10.11.12",
			remoteAddr:     "13.14.15.16:1234",
			expected:       "1.2.3.4",
		},
		{
			name:          "CF-Connecting-IPがない場合はX-Forwarded-Forが使用される",
			xForwardedFor: "5.6.7.8",
			xRealIP:       "9.10.11.12",
			remoteAddr:    "13.14.15.16:1234",
			expected:      "5.6.7.8",
		},
		{
			name:          "X-Forwarded-Forのカンマ区切りの最初のIPが取得される",
			xForwardedFor: "5.6.7.8, 9.10.11.12, 13.14.15.16",
			xRealIP:       "17.18.19.20",
			remoteAddr:    "21.22.23.24:1234",
			expected:      "5.6.7.8",
		},
		{
			name:          "X-Forwarded-Forのスペース付きカンマ区切り",
			xForwardedFor: "5.6.7.8 ,  9.10.11.12",
			xRealIP:       "17.18.19.20",
			remoteAddr:    "21.22.23.24:1234",
			expected:      "5.6.7.8",
		},
		{
			name:       "X-Forwarded-Forがない場合はX-Real-IPが使用される",
			xRealIP:    "9.10.11.12",
			remoteAddr: "13.14.15.16:1234",
			expected:   "9.10.11.12",
		},
		{
			name:       "ヘッダーがない場合はRemoteAddr（ポート番号なし）が使用される",
			remoteAddr: "13.14.15.16:1234",
			expected:   "13.14.15.16",
		},
		{
			name:       "RemoteAddrにポート番号がない場合はそのまま返される",
			remoteAddr: "192.168.1.1",
			expected:   "192.168.1.1",
		},
		{
			name:     "すべて空の場合は空のRemoteAddrが返される",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.cfConnectingIP != "" {
				req.Header.Set("CF-Connecting-IP", tt.cfConnectingIP)
			}
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}
			req.RemoteAddr = tt.remoteAddr

			got := GetClientIP(req)
			if got != tt.expected {
				t.Errorf("GetClientIP() = %v, want %v", got, tt.expected)
			}
		})
	}
}
