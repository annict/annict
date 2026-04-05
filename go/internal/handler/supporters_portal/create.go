package supporters_portal

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/annict/annict/go/internal/i18n"
	authMiddleware "github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/session"
	"github.com/annict/annict/go/internal/usecase"
)

// Create POST /supporters/portal - Stripe Customer Portalへリダイレクト
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// ログインユーザーの取得（認証必須）
	user := authMiddleware.GetUserFromContext(ctx)
	if user == nil {
		http.Redirect(w, r, "/sign_in", http.StatusSeeOther)
		return
	}

	// UseCaseの実行
	result, err := h.createPortalSessionUC.Execute(ctx, usecase.CreatePortalSessionInput{
		User:   user,
		Locale: i18n.GetLocale(ctx),
	})
	if err != nil {
		if usecase.IsNotStripeSubscriberError(err) {
			slog.WarnContext(ctx, "Stripeサポーターではないユーザーがポータルにアクセスしようとしました", "user_id", user.ID)
			h.redirectWithError(w, r, ctx, "supporters_portal_not_supporter")
			return
		}

		slog.ErrorContext(ctx, "Customer Portalセッションの作成に失敗しました", "error", err, "user_id", user.ID)
		h.redirectWithError(w, r, ctx, "supporters_portal_error")
		return
	}

	slog.InfoContext(ctx, "Stripe Customer Portalセッションを作成しました", "user_id", user.ID)

	// Stripe Customer Portalページへリダイレクト
	http.Redirect(w, r, result.PortalURL, http.StatusSeeOther)
}

// redirectWithError はエラーメッセージをフラッシュに設定してリダイレクトします
func (h *Handler) redirectWithError(w http.ResponseWriter, r *http.Request, ctx context.Context, messageKey string) {
	// フラッシュメッセージを設定（"error"タイプ）
	h.sessionManager.SetFlash(w, session.FlashError, i18n.T(ctx, messageKey))
	http.Redirect(w, r, "/supporters", http.StatusSeeOther)
}
