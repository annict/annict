package sign_in

import (
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/sign_in"
	"github.com/annict/annict/go/internal/viewmodel"
)

// New はメールアドレス入力フォームを表示します (GET /sign_in)
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

	backURL := r.URL.Query().Get("back")

	h.renderNewForm(w, r, http.StatusOK, nil, "", backURL)
}

// renderNewForm はメールアドレス入力フォームをレンダリングします。
// バリデーションエラーが存在する場合は status に http.StatusUnprocessableEntity を渡してください。
func (h *Handler) renderNewForm(w http.ResponseWriter, r *http.Request, status int, formErrors *model.ValidationError, email string, backURL string) {
	ctx := r.Context()

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "sign_in_title")
	meta.Description = i18n.T(ctx, "sign_in_description")
	meta.OGURL = h.cfg.AppURL() + "/sign_in"

	csrfToken := middleware.GetOrCreateCSRFToken(w, r, h.sessionMgr)

	data := sign_in.NewPageData{
		CSRFToken:        csrfToken,
		TurnstileSiteKey: h.cfg.TurnstileSiteKey,
		FormErrors:       formErrors,
		Email:            email,
		BackURL:          backURL,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	component := layouts.Simple(ctx, meta, h.cfg.GetAssetVersion(), sign_in.New(data))
	if err := component.Render(ctx, w); err != nil {
		slog.ErrorContext(ctx, "テンプレート実行エラー", "error", err)
	}
}
