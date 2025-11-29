package sign_in_code

import (
	"log/slog"
	"net/http"

	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/middleware"
	"github.com/annict/annict/internal/templates/layouts"
	signInCodePage "github.com/annict/annict/internal/templates/pages/sign_in_code"
	"github.com/annict/annict/internal/viewmodel"
)

// Show は6桁コード入力フォームを表示します (GET /sign_in/code)
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
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

	// セッションからメールアドレスを取得
	email, err := h.sessionMgr.GetValue(ctx, r, "sign_in_email")
	if err != nil {
		slog.Error("セッション値取得エラー", "error", err, "key", "sign_in_email")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// メールアドレスがセッションにない場合は /sign_in にリダイレクト
	if email == "" {
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}

	// backパラメータを取得（ログイン後のリダイレクト先）
	backURL := r.URL.Query().Get("back")

	// Flashメッセージを取得
	flash, _ := h.sessionMgr.GetFlash(ctx, r)
	formErrors, _ := h.sessionMgr.GetFormErrors(ctx, r)

	// メタ情報を設定
	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "sign_in_code_title")
	meta.Description = i18n.T(ctx, "sign_in_code_description")
	meta.OGURL = h.cfg.AppURL() + "/sign_in/code"

	// CSRFトークンを取得
	csrfToken := middleware.GetCSRFToken(r, h.sessionMgr)

	// テンプレートをレンダリング
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := layouts.Simple(ctx, meta, flash, h.cfg.GetAssetVersion(), signInCodePage.SignInCodeShow(ctx, email, formErrors, csrfToken, backURL))
	if err := component.Render(ctx, w); err != nil {
		slog.Error("テンプレート実行エラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
