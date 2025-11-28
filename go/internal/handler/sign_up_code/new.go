package sign_up_code

import (
	"log/slog"
	"net/http"

	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/middleware"
	"github.com/annict/annict/internal/templates/layouts"
	signUpCodePage "github.com/annict/annict/internal/templates/pages/sign_up_code"
	"github.com/annict/annict/internal/viewmodel"
)

// New は新規登録確認コード入力フォームを表示します (GET /sign_up/code)
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// セッションからメールアドレスを取得
	email, err := h.sessionMgr.GetValue(ctx, r, "sign_up_email")
	if err != nil {
		slog.Error("セッション値取得エラー", "error", err, "key", "sign_up_email")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// メールアドレスがセッションにない場合は /sign_up にリダイレクト
	if email == "" {
		http.Redirect(w, r, "/sign_up", http.StatusSeeOther)
		return
	}

	// Flashメッセージを取得
	flash, _ := h.sessionMgr.GetFlash(ctx, r)
	formErrors, _ := h.sessionMgr.GetFormErrors(ctx, r)

	// メタ情報を設定
	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "sign_up_code_title")
	meta.Description = i18n.T(ctx, "sign_up_code_description")
	meta.OGURL = h.cfg.AppURL() + "/sign_up/code"

	// CSRFトークンを取得
	csrfToken := middleware.GetCSRFToken(r, h.sessionMgr)

	// テンプレートをレンダリング
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := layouts.Simple(ctx, meta, flash, h.cfg.GetAssetVersion(), signUpCodePage.SignUpCodeNew(ctx, email, formErrors, csrfToken))
	if err := component.Render(ctx, w); err != nil {
		slog.Error("テンプレート実行エラー", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
