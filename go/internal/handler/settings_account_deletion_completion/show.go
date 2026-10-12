package settings_account_deletion_completion

import (
	"bytes"
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/settings_account_deletion"
	"github.com/annict/annict/go/internal/viewmodel"
)

// settingsAccountDeletionCompletionPathは退会の完了ページの代表GETパス。
const settingsAccountDeletionCompletionPath = "/settings/account/deletion/completed"

// Showは退会の完了ページを表示します (GET /settings/account/deletion/completed)。
// 受付後はセッションを破棄しているため、ログインを求めない。
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, settingsAccountDeletionCompletionPath)
	meta.SetTitle(ctx, "settings_account_deletion_completion_show_title")

	component := layouts.Default(
		ctx,
		meta,
		viewmodel.NewUserForSidebar(middleware.GetUserFromContext(ctx), h.imageHelper),
		viewmodel.NewSeasons(h.cfg),
		h.cfg.GetAssetVersion(),
		settings_account_deletion.CompletionShow(),
	)
	var body bytes.Buffer
	if err := component.Render(ctx, &body); err != nil {
		slog.ErrorContext(ctx, "テンプレートのレンダリングエラー", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(body.Bytes()); err != nil {
		slog.ErrorContext(ctx, "退会の完了ページのレスポンスの書き込みに失敗", "error", err)
	}
}
