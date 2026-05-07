package password_reset

import (
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates/layouts"
	passwordpages "github.com/annict/annict/go/internal/templates/pages/password"
	"github.com/annict/annict/go/internal/viewmodel"
)

// New はパスワードリセット申請フォームを表示します (GET /password/reset)
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	h.renderNewForm(w, r, http.StatusOK, nil, "")
}

// renderNewForm はパスワードリセット申請フォームをレンダリングします。
// バリデーションエラーが存在する場合は status に http.StatusUnprocessableEntity を渡してください。
func (h *Handler) renderNewForm(w http.ResponseWriter, r *http.Request, status int, formErrors *model.ValidationError, email string) {
	ctx := r.Context()

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg)
	meta.SetTitle(ctx, "password_reset_title")
	meta.OGURL = h.cfg.AppURL() + "/password/reset"

	csrfToken := middleware.GetOrCreateCSRFToken(w, r, h.sessionMgr)

	data := passwordpages.ResetPageData{
		CSRFToken:        csrfToken,
		TurnstileSiteKey: h.cfg.TurnstileSiteKey,
		FormErrors:       formErrors,
		Email:            email,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	component := layouts.Simple(ctx, meta, h.cfg.GetAssetVersion(), passwordpages.Reset(data))
	if err := component.Render(ctx, w); err != nil {
		slog.ErrorContext(ctx, "テンプレート実行エラー", "error", err)
	}
}
