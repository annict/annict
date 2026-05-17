package supporters_checkout

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/i18n"
	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

// Create POST /supporters/checkout - Stripe Checkoutセッションを作成
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// ログインユーザーの取得（認証必須）
	user := authMiddleware.GetUserFromContext(ctx)
	if user == nil {
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}

	// リクエストパラメータの解析
	if err := r.ParseForm(); err != nil {
		slog.ErrorContext(ctx, "フォームのパースに失敗しました", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// UseCaseの実行
	output, err := h.createCheckoutSessionUC.Execute(ctx, usecase.CreateCheckoutSessionInput{
		User:   user,
		Plan:   r.FormValue("plan"),
		Locale: i18n.GetLocale(ctx),
	})
	if err != nil {
		if ve := model.AsValidationError(err); ve != nil {
			slog.WarnContext(ctx, "バリデーションエラー", "errors", ve, "user_id", user.ID)
			h.redirectWithError(w, r, ctx, "supporters_checkout_invalid_plan")
			return
		}

		if ae := model.AsAppError(err); ae != nil && ae.Code == model.AppErrCodeConflict {
			slog.InfoContext(ctx, "既にアクティブなサブスクリプションが存在します", "user_id", user.ID)
			h.flashMgr.SetError(w, ae.UserMsg)
			http.Redirect(w, r, "/supporters", http.StatusSeeOther)
			return
		}

		slog.ErrorContext(ctx, "Checkoutセッションの作成に失敗しました", "error", err, "user_id", user.ID)
		h.redirectWithError(w, r, ctx, "supporters_checkout_error")
		return
	}

	slog.InfoContext(ctx, "Stripe Checkoutセッションを作成しました", "user_id", user.ID, "plan", r.FormValue("plan"))

	// Stripe Checkoutページへリダイレクト
	http.Redirect(w, r, output.CheckoutURL, http.StatusSeeOther)
}

// redirectWithError はエラーメッセージをフラッシュに設定してリダイレクトします
func (h *Handler) redirectWithError(w http.ResponseWriter, r *http.Request, ctx context.Context, messageKey string) {
	h.flashMgr.SetError(w, i18n.T(ctx, messageKey))
	http.Redirect(w, r, "/supporters", http.StatusSeeOther)
}
