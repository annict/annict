package settings_account_deletion_completion

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/i18n"
)

// TestShowは、受付後のセッションが無い状態でも完了ページを表示することを検証する。
func TestShow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		locale   string
		expected []string
	}{
		{
			name:   "日本語",
			locale: "ja",
			expected: []string{
				"<title>退会を受け付けました | Annict</title>",
				"データの削除は順次行います",
				`<a href="/" class="underline">`,
			},
		},
		{
			name:   "英語",
			locale: "en",
			expected: []string{
				"<title>Your Account Deletion Has Been Accepted | Annict</title>",
				"Thank you for using Annict.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := NewHandler(&config.Config{Env: "test", Domain: "test.annict.com"}, nil)

			req := httptest.NewRequest(http.MethodGet, "/settings/account/deletion/completed", nil)
			req = req.WithContext(i18n.SetLocale(req.Context(), tt.locale))
			rr := httptest.NewRecorder()
			handler.Show(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("ステータスコード = %d、期待値 = %d", rr.Code, http.StatusOK)
			}
			if ct := rr.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
				t.Errorf("Content-Type = %q、期待値 = text/html; charset=utf-8", ct)
			}

			body := rr.Body.String()
			for _, expected := range tt.expected {
				if !strings.Contains(body, expected) {
					t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
				}
			}
		})
	}
}
