// Package httperror renders shared, localized HTTP error responses for handlers.
//
// [Ja] Package httperror はハンドラー共通のローカライズ済み HTTP エラーレスポンスを描画する。
package httperror

import (
	"bytes"
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/templates/layouts"
	errorpages "github.com/annict/annict/go/internal/templates/pages/errors"
)

// NotFound renders the shared 404 page.
//
// [Ja] NotFound は共通の 404 ページを描画する。
func NotFound(w http.ResponseWriter, r *http.Request) {
	render(w, r, http.StatusNotFound, "error_not_found_title", "error_not_found_message")
}

// InternalServerError renders the shared 500 page without exposing the underlying error.
//
// [Ja] InternalServerError は内部エラーを公開せず、共通の 500 ページを描画する。
func InternalServerError(w http.ResponseWriter, r *http.Request) {
	render(w, r, http.StatusInternalServerError, "error_internal_server_title", "error_internal_server_message")
}

func render(w http.ResponseWriter, r *http.Request, status int, titleKey string, messageKey string) {
	ctx := r.Context()
	title := i18n.T(ctx, titleKey)
	backLink := &errorpages.BackLink{
		URL:  "/",
		Text: i18n.T(ctx, "error_back_to_home"),
	}
	component := layouts.Error(
		title,
		errorpages.HTTPError(title, i18n.T(ctx, messageKey), backLink),
	)

	var body bytes.Buffer
	if err := component.Render(ctx, &body); err != nil {
		slog.ErrorContext(ctx, "HTTPエラーページのレンダリングに失敗", "status", status, "error", err)
		http.Error(w, http.StatusText(status), status)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write(body.Bytes()); err != nil {
		slog.ErrorContext(ctx, "HTTPエラーレスポンスの書き込みに失敗", "status", status, "error", err)
	}
}
