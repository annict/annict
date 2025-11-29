package password_reset

import (
	"log/slog"
	"net/http"

	"github.com/annict/annict/internal/middleware"
	"github.com/annict/annict/internal/templates/layouts"
	passwordpages "github.com/annict/annict/internal/templates/pages/password"
	"github.com/annict/annict/internal/viewmodel"
)

// New はパスワードリセット申請フォームを表示します (GET /password/reset)
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// セッションからフラッシュメッセージとフォームエラーを取得
	flash, _ := h.sessionManager.GetFlash(ctx, r)
	formErrors, _ := h.sessionManager.GetFormErrors(ctx, r)

	// メタ情報を設定
	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "password_reset_title")
	meta.OGURL = h.cfg.AppURL() + "/password/reset"

	// CSRFトークンを取得（セッションが存在しない場合は新規作成）
	csrfToken := middleware.GetOrCreateCSRFToken(w, r, h.sessionManager)

	// テンプレートをレンダリング
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := layouts.Simple(ctx, meta, flash, h.cfg.GetAssetVersion(), passwordpages.Reset(ctx, formErrors, csrfToken, "", h.cfg.TurnstileSiteKey))
	if err := component.Render(ctx, w); err != nil {
		slog.Error("テンプレート実行エラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "パスワードリセット申請フォームを表示しました")
}
