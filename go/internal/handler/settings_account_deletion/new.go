package settings_account_deletion

import (
	"bytes"
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/httperror"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/templates/layouts"
	"github.com/annict/annict/go/internal/templates/pages/settings_account_deletion"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/viewmodel"
)

// settingsAccountDeletionNewPathは退会の確認ページの代表GETパス。
const settingsAccountDeletionNewPath = "/settings/account/deletion/new"

// Newは退会の確認ページを表示します (GET /settings/account/deletion/new)
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	h.renderNew(w, r, http.StatusOK, nil, "")
}

// renderNewは退会の確認ページを描画します。
// 受付のバリデーションエラーを示すときはstatusにhttp.StatusUnprocessableEntityを渡し、
// inputUsernameに利用者が入力した値を渡してください。
func (h *Handler) renderNew(w http.ResponseWriter, r *http.Request, status int, formErrors *model.ValidationError, inputUsername string) {
	ctx := r.Context()
	user := middleware.GetUserFromContext(ctx)

	output, err := h.getAccountDeletionNewUC.Execute(ctx, usecase.GetAccountDeletionNewInput{User: user})
	if err != nil {
		slog.ErrorContext(ctx, "退会の確認ページのデータの取得に失敗", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	meta := viewmodel.DefaultPageMeta(ctx, h.cfg, settingsAccountDeletionNewPath)
	meta.SetTitle(ctx, "settings_account_deletion_new_title")

	component := layouts.Default(
		ctx,
		meta,
		viewmodel.NewUserForSidebar(user, h.imageHelper),
		viewmodel.NewSeasons(h.cfg),
		h.cfg.GetAssetVersion(),
		settings_account_deletion.New(settings_account_deletion.NewPageData{
			CSRFToken:                middleware.GetCSRFToken(r, h.sessionManager),
			Username:                 user.Username,
			IsStripeSupporter:        output.IsStripeSupporter,
			IsGumroadSupporter:       output.IsGumroadSupporter,
			HasKeptOAuthApplications: output.HasKeptOAuthApplications,
			FormErrors:               viewmodel.NewFormErrors(formErrors),
			InputUsername:            inputUsername,
		}),
	)
	var body bytes.Buffer
	if err := component.Render(ctx, &body); err != nil {
		slog.ErrorContext(ctx, "テンプレートのレンダリングエラー", "error", err)
		httperror.InternalServerError(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write(body.Bytes()); err != nil {
		slog.ErrorContext(ctx, "退会の確認ページのレスポンスの書き込みに失敗", "error", err)
	}
}
