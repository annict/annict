package settings_two_factor_auth

import (
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	page "github.com/annict/annict/go/internal/templates/pages/settings_two_factor_auth"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// Showは二要素認証の状態を表示します (GET /settings/two_factor_auth)
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	h.renderShow(w, r, http.StatusOK, nil)
}

// renderShowは設定画面を指定したステータスで描画します。
// 表示 (200) と、無効にするフォームの本人確認を受け付けなかったときの再描画 (422) で共有する。
func (h *Handler) renderShow(w http.ResponseWriter, r *http.Request, status int, formErrors *model.ValidationError) {
	ctx := r.Context()
	user := middleware.GetUserFromContext(ctx)

	output, err := h.getTwoFactorAuthStatusUC.Execute(ctx, usecase.GetTwoFactorAuthStatusInput{UserID: user.ID})
	if err != nil {
		slog.ErrorContext(ctx, "二要素認証の状態の取得に失敗", "error", err, "user_id", user.ID)
		setNoStore(w)
		httperror.InternalServerError(w, r)
		return
	}

	h.renderPage(w, r, status, "settings_two_factor_auth_show_title", page.ShowPath, page.Show(page.ShowPageData{
		Enabled:                 output.Enabled,
		UnusedRecoveryCodeCount: output.UnusedRecoveryCodeCount,
		CSRFToken:               middleware.GetCSRFToken(r, h.sessionManager),
		FormErrors:              viewmodel.NewFormErrors(formErrors),
	}))
}
