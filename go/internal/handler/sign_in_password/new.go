package sign_in_password

import (
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates/layouts"
	signInPasswordPage "github.com/annict/annict/go/internal/templates/pages/sign_in_password"
	"github.com/annict/annict/go/internal/viewmodel"
)

// New GET /sign_in/password - パスワードログインフォーム
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
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
		slog.ErrorContext(ctx, "セッションからメールアドレスの取得に失敗しました", "error", err)
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}

	// メールアドレスがない場合は /sign_in にリダイレクト
	if email == "" {
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}

	// backパラメータを取得（ログイン後のリダイレクト先）
	backURL := r.URL.Query().Get("back")

	h.renderNewForm(w, r, http.StatusOK, nil, email, backURL)
}

// renderNewForm はパスワードログインフォームをレンダリングします。
// バリデーションエラーが存在する場合は status に http.StatusUnprocessableEntity を渡してください。
func (h *Handler) renderNewForm(w http.ResponseWriter, r *http.Request, status int, formErrors *model.ValidationError, email string, backURL string) {
	ctx := r.Context()

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "sign_in_title")
	meta.Description = i18n.T(ctx, "sign_in_description")
	meta.OGURL = h.cfg.AppURL() + "/sign_in/password"

	csrfToken := middleware.GetOrCreateCSRFToken(w, r, h.sessionMgr)

	data := signInPasswordPage.NewPageData{
		CSRFToken:  csrfToken,
		FormErrors: formErrors,
		Email:      email,
		BackURL:    backURL,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	component := layouts.Simple(ctx, meta, h.cfg.GetAssetVersion(), signInPasswordPage.New(data))
	if err := component.Render(ctx, w); err != nil {
		slog.ErrorContext(ctx, "テンプレート実行エラー", "error", err)
	}
}
