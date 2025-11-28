package sign_up

import (
	"log/slog"
	"net/http"

	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/middleware"
	"github.com/annict/annict/internal/templates/layouts"
	"github.com/annict/annict/internal/templates/pages/sign_up"
	"github.com/annict/annict/internal/viewmodel"
)

// New は新規登録フォームを表示します (GET /sign_up)
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// すでにログイン済みの場合はホームにリダイレクト
	currentUser, err := h.sessionMgr.GetCurrentUser(ctx, r)
	if err != nil {
		slog.Error("セッション取得エラー", "error", err)
	}
	if currentUser != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Flashメッセージを取得
	flash, _ := h.sessionMgr.GetFlash(ctx, r)
	formErrors, _ := h.sessionMgr.GetFormErrors(ctx, r)

	// メタ情報を設定
	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "sign_up_title")
	meta.Description = i18n.T(ctx, "sign_up_description")
	meta.OGURL = h.cfg.AppURL() + "/sign_up"

	// CSRFトークンを取得（セッションが存在しない場合は新規作成）
	csrfToken := middleware.GetOrCreateCSRFToken(w, r, h.sessionMgr)

	// テンプレートをレンダリング
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := layouts.Simple(ctx, meta, flash, h.cfg.GetAssetVersion(), sign_up.New(ctx, formErrors, csrfToken, h.cfg.TurnstileSiteKey))
	if err := component.Render(ctx, w); err != nil {
		slog.Error("テンプレート実行エラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
