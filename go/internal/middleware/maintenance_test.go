package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/annict/annict/internal/config"
)

// テスト用のダミーハンドラー（200 OK を返す）
func testHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
}

func TestMaintenanceMiddleware_DisabledMode(t *testing.T) {
	t.Parallel()

	// メンテナンスモードが無効の場合は通常処理
	cfg := &config.Config{
		MaintenanceMode: false,
		AdminIPs:        []string{},
	}

	mw := NewMaintenanceMiddleware(cfg)
	handler := mw.Middleware(testHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("メンテナンスモードOFF時: ステータスコード = %d, want %d", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != "OK" {
		t.Errorf("メンテナンスモードOFF時: レスポンスボディ = %q, want %q", rr.Body.String(), "OK")
	}
}

func TestMaintenanceMiddleware_EnabledMode_AdminIP(t *testing.T) {
	t.Parallel()

	// メンテナンスモードON + 管理者IPの場合は通常処理
	cfg := &config.Config{
		MaintenanceMode: true,
		AdminIPs:        []string{"192.168.1.100"},
	}

	mw := NewMaintenanceMiddleware(cfg)
	handler := mw.Middleware(testHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("メンテナンスモードON+管理者IP: ステータスコード = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestMaintenanceMiddleware_EnabledMode_NonAdminIP(t *testing.T) {
	t.Parallel()

	// メンテナンスモードON + 一般IPの場合は503を返す
	cfg := &config.Config{
		MaintenanceMode: true,
		AdminIPs:        []string{"192.168.1.100"},
	}

	mw := NewMaintenanceMiddleware(cfg)
	handler := mw.Middleware(testHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("メンテナンスモードON+一般IP: ステータスコード = %d, want %d", rr.Code, http.StatusServiceUnavailable)
	}

	// Content-Typeヘッダーの確認
	contentType := rr.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", contentType, "text/html; charset=utf-8")
	}

	// Retry-Afterヘッダーの確認
	retryAfter := rr.Header().Get("Retry-After")
	if retryAfter != "3600" {
		t.Errorf("Retry-After = %q, want %q", retryAfter, "3600")
	}

	// メンテナンスページの内容を確認
	body := rr.Body.String()
	if !strings.Contains(body, "メンテナンス") {
		t.Error("レスポンスボディにメンテナンスページの内容が含まれていません")
	}
}

func TestMaintenanceMiddleware_MultipleAdminIPs(t *testing.T) {
	t.Parallel()

	// 複数の管理者IPに対応
	cfg := &config.Config{
		MaintenanceMode: true,
		AdminIPs:        []string{"192.168.1.100", "10.0.0.50", "172.16.0.1"},
	}

	mw := NewMaintenanceMiddleware(cfg)
	handler := mw.Middleware(testHandler())

	testCases := []struct {
		name     string
		ip       string
		wantCode int
	}{
		{"最初の管理者IP", "192.168.1.100:12345", http.StatusOK},
		{"2番目の管理者IP", "10.0.0.50:12345", http.StatusOK},
		{"3番目の管理者IP", "172.16.0.1:12345", http.StatusOK},
		{"管理者以外のIP", "8.8.8.8:12345", http.StatusServiceUnavailable},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tc.ip
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantCode {
				t.Errorf("%s: ステータスコード = %d, want %d", tc.name, rr.Code, tc.wantCode)
			}
		})
	}
}

func TestMaintenanceMiddleware_XForwardedFor(t *testing.T) {
	t.Parallel()

	// X-Forwarded-Forヘッダーからの IP取得
	cfg := &config.Config{
		MaintenanceMode: true,
		AdminIPs:        []string{"203.0.113.50"},
	}

	mw := NewMaintenanceMiddleware(cfg)
	handler := mw.Middleware(testHandler())

	testCases := []struct {
		name        string
		xff         string
		remoteAddr  string
		wantCode    int
		description string
	}{
		{
			name:        "XFF経由で管理者IP",
			xff:         "203.0.113.50",
			remoteAddr:  "10.0.0.1:12345",
			wantCode:    http.StatusOK,
			description: "X-Forwarded-Forに管理者IPが含まれる場合は通常処理",
		},
		{
			name:        "XFF経由で一般IP",
			xff:         "8.8.8.8",
			remoteAddr:  "10.0.0.1:12345",
			wantCode:    http.StatusServiceUnavailable,
			description: "X-Forwarded-Forが管理者IP以外の場合は503",
		},
		{
			name:        "XFFが複数IP（最初が管理者IP）",
			xff:         "203.0.113.50, 10.0.0.1, 172.16.0.1",
			remoteAddr:  "192.168.1.1:12345",
			wantCode:    http.StatusOK,
			description: "X-Forwarded-Forの最初のIPが管理者IPの場合は通常処理",
		},
		{
			name:        "XFFが複数IP（最初が一般IP）",
			xff:         "8.8.8.8, 203.0.113.50, 10.0.0.1",
			remoteAddr:  "192.168.1.1:12345",
			wantCode:    http.StatusServiceUnavailable,
			description: "X-Forwarded-Forの最初のIPが管理者IP以外の場合は503",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("X-Forwarded-For", tc.xff)
			req.RemoteAddr = tc.remoteAddr
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantCode {
				t.Errorf("%s: ステータスコード = %d, want %d (%s)", tc.name, rr.Code, tc.wantCode, tc.description)
			}
		})
	}
}

func TestMaintenanceMiddleware_CFConnectingIP(t *testing.T) {
	t.Parallel()

	// CF-Connecting-IPヘッダー（Cloudflare経由）からのIP取得
	cfg := &config.Config{
		MaintenanceMode: true,
		AdminIPs:        []string{"203.0.113.100"},
	}

	mw := NewMaintenanceMiddleware(cfg)
	handler := mw.Middleware(testHandler())

	testCases := []struct {
		name           string
		cfConnectingIP string
		xff            string
		remoteAddr     string
		wantCode       int
	}{
		{
			name:           "CF-Connecting-IPが管理者IP",
			cfConnectingIP: "203.0.113.100",
			xff:            "8.8.8.8",
			remoteAddr:     "10.0.0.1:12345",
			wantCode:       http.StatusOK,
		},
		{
			name:           "CF-Connecting-IPが一般IP",
			cfConnectingIP: "8.8.8.8",
			xff:            "203.0.113.100",
			remoteAddr:     "10.0.0.1:12345",
			wantCode:       http.StatusServiceUnavailable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("CF-Connecting-IP", tc.cfConnectingIP)
			req.Header.Set("X-Forwarded-For", tc.xff)
			req.RemoteAddr = tc.remoteAddr
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantCode {
				t.Errorf("%s: ステータスコード = %d, want %d", tc.name, rr.Code, tc.wantCode)
			}
		})
	}
}

func TestMaintenanceMiddleware_EmptyAdminIPs(t *testing.T) {
	t.Parallel()

	// 管理者IPが設定されていない場合（空のスライス）
	cfg := &config.Config{
		MaintenanceMode: true,
		AdminIPs:        []string{},
	}

	mw := NewMaintenanceMiddleware(cfg)
	handler := mw.Middleware(testHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// 管理者IPが設定されていない場合は全てのアクセスで503
	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("管理者IP未設定時: ステータスコード = %d, want %d", rr.Code, http.StatusServiceUnavailable)
	}
}

func TestMaintenanceMiddleware_NilAdminIPs(t *testing.T) {
	t.Parallel()

	// 管理者IPがnilの場合
	cfg := &config.Config{
		MaintenanceMode: true,
		AdminIPs:        nil,
	}

	mw := NewMaintenanceMiddleware(cfg)
	handler := mw.Middleware(testHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// 管理者IPがnilの場合も全てのアクセスで503
	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("管理者IPがnil時: ステータスコード = %d, want %d", rr.Code, http.StatusServiceUnavailable)
	}
}

func TestMaintenanceMiddleware_XRealIP(t *testing.T) {
	t.Parallel()

	// X-Real-IPヘッダーからのIP取得
	cfg := &config.Config{
		MaintenanceMode: true,
		AdminIPs:        []string{"203.0.113.200"},
	}

	mw := NewMaintenanceMiddleware(cfg)
	handler := mw.Middleware(testHandler())

	testCases := []struct {
		name       string
		xRealIP    string
		remoteAddr string
		wantCode   int
	}{
		{
			name:       "X-Real-IPが管理者IP",
			xRealIP:    "203.0.113.200",
			remoteAddr: "10.0.0.1:12345",
			wantCode:   http.StatusOK,
		},
		{
			name:       "X-Real-IPが一般IP",
			xRealIP:    "8.8.8.8",
			remoteAddr: "10.0.0.1:12345",
			wantCode:   http.StatusServiceUnavailable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("X-Real-IP", tc.xRealIP)
			req.RemoteAddr = tc.remoteAddr
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantCode {
				t.Errorf("%s: ステータスコード = %d, want %d", tc.name, rr.Code, tc.wantCode)
			}
		})
	}
}
