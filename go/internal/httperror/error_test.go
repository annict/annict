package httperror

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
)

func TestNotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		locale        string
		wantTitle     string
		wantMessage   string
		wantBackLabel string
	}{
		{
			name:          "日本語",
			locale:        "ja",
			wantTitle:     "ページが見つかりません",
			wantMessage:   "ページが移動または削除された可能性があります。",
			wantBackLabel: "ホームに戻る",
		},
		{
			name:          "英語",
			locale:        "en",
			wantTitle:     "Page not found",
			wantMessage:   "The page may have been moved or deleted.",
			wantBackLabel: "Back to Home",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/missing", nil)
			req = req.WithContext(i18n.SetLocale(req.Context(), tt.locale))
			rr := httptest.NewRecorder()

			NotFound(rr, req)

			assertErrorResponse(t, rr, http.StatusNotFound, tt.locale, tt.wantTitle, tt.wantMessage, tt.wantBackLabel)
		})
	}
}

func TestInternalServerError(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/broken", nil)
	req = req.WithContext(i18n.SetLocale(req.Context(), "ja"))
	rr := httptest.NewRecorder()

	InternalServerError(rr, req)

	assertErrorResponse(
		t,
		rr,
		http.StatusInternalServerError,
		"ja",
		"問題が発生しました",
		"しばらくしてから、もう一度お試しください。",
		"ホームに戻る",
	)
	if strings.Contains(rr.Body.String(), "Internal Server Error") {
		t.Error("500 レスポンスに内部エラーの文言を含めてはいけません")
	}
}

func assertErrorResponse(
	t *testing.T,
	rr *httptest.ResponseRecorder,
	wantStatus int,
	wantLocale string,
	wantTitle string,
	wantMessage string,
	wantBackLabel string,
) {
	t.Helper()

	if rr.Code != wantStatus {
		t.Errorf("status = %d, want %d", rr.Code, wantStatus)
	}
	if contentType := rr.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/html; charset=utf-8", contentType)
	}

	body := rr.Body.String()
	for _, expected := range []string{
		`<html lang="` + wantLocale + `">`,
		"<title>" + wantTitle + " | Annict</title>",
		"<h1",
		wantTitle,
		wantMessage,
		`href="/"`,
		wantBackLabel,
		`class="error-link"`,
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("response body does not contain %q", expected)
		}
	}
}
