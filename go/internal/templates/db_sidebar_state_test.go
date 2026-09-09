package templates

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestDBSidebarStateMiddleware verifies valid, missing, and invalid Cookie values produce the
// expected server-rendered desktop sidebar preference.
//
// [Ja] TestDBSidebarStateMiddleware は有効・未設定・不正な Cookie 値から、SSR 用の
// デスクトップサイドバー設定が期待どおり生成されることを検証する。
func TestDBSidebarStateMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cookieValue string
		wantOpen    bool
	}{
		{name: "missing", wantOpen: true},
		{name: "open", cookieValue: "true", wantOpen: true},
		{name: "closed", cookieValue: "false", wantOpen: false},
		{name: "invalid", cookieValue: "invalid", wantOpen: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/db/works", nil)
			if tt.cookieValue != "" {
				req.AddCookie(&http.Cookie{Name: DBSidebarOpenCookieName, Value: tt.cookieValue})
			}

			var got bool
			handler := DBSidebarStateMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = IsDBSidebarOpen(r.Context())
			}))
			handler.ServeHTTP(httptest.NewRecorder(), req)

			if got != tt.wantOpen {
				t.Errorf("IsDBSidebarOpen() = %v, want %v", got, tt.wantOpen)
			}
		})
	}
}
