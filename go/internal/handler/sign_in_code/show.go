package sign_in_code

import (
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates/layouts"
	signInCodePage "github.com/annict/annict/go/internal/templates/pages/sign_in_code"
	"github.com/annict/annict/go/internal/viewmodel"
)

// Show は6桁コード入力フォームを表示します (GET /sign_in/code)
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// すでにログイン済みの場合はホームにリダイレクト
	currentUser, err := h.sessionMgr.GetCurrentUser(ctx, r)
	if err != nil {
		slog.ErrorContext(ctx, "セッション取得エラー", "error", err)
	}
	if currentUser != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// セッションからメールアドレスを取得
	email, err := h.sessionMgr.GetValue(ctx, r, "sign_in_email")
	if err != nil {
		slog.ErrorContext(ctx, "セッション値取得エラー", "error", err, "key", "sign_in_email")
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

	h.renderShowForm(w, r, http.StatusOK, nil, email, backURL)
}

// renderShowForm は6桁コード入力フォームをレンダリングします。
// バリデーションエラーが存在する場合は status に http.StatusUnprocessableEntity を渡してください。
func (h *Handler) renderShowForm(w http.ResponseWriter, r *http.Request, status int, formErrors *model.ValidationError, email string, backURL string) {
	ctx := r.Context()

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "sign_in_code_title")
	meta.Description = i18n.T(ctx, "sign_in_code_description")
	meta.OGURL = h.cfg.AppURL() + "/sign_in/code"

	csrfToken := middleware.GetOrCreateCSRFToken(w, r, h.sessionMgr)

	data := signInCodePage.ShowPageData{
		CSRFToken:  csrfToken,
		FormErrors: formErrors,
		Email:      email,
		BackURL:    backURL,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	component := layouts.Simple(ctx, meta, h.cfg.GetAssetVersion(), signInCodePage.Show(data))
	if err := component.Render(ctx, w); err != nil {
		slog.ErrorContext(ctx, "テンプレート実行エラー", "error", err)
	}
}
