package templates

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestDBSidebarStateMiddlewareは有効・未設定・不正なCookie値から、SSR用の
// デスクトップサイドバー設定が期待どおり生成されることを検証する。
func TestDBSidebarStateMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cookieValue string
		wantOpen    bool
	}{
		{name: "未設定", wantOpen: true},
		{name: "開いている", cookieValue: "true", wantOpen: true},
		{name: "閉じている", cookieValue: "false", wantOpen: false},
		{name: "不正な値", cookieValue: "invalid", wantOpen: true},
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
				t.Errorf("IsDBSidebarOpen() = %v、期待値 = %v", got, tt.wantOpen)
			}
		})
	}
}
