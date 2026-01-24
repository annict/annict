package supporters_portal

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/stripe/stripe-go/v84"

	"github.com/annict/annict/go/internal/i18n"
	authMiddleware "github.com/annict/annict/go/internal/middleware"
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

	// Stripeサポーターのみアクセス可能
	if !user.StripeSubscriberID.Valid || h.stripeSubscriberRepo == nil {
		slog.WarnContext(ctx, "Stripeサポーターではないユーザーがポータルにアクセスしようとしました", "user_id", user.ID)
		h.redirectWithError(w, r, ctx, "supporters_portal_not_supporter")
		return
	}

	stripeSubscriber, err := h.stripeSubscriberRepo.GetByID(ctx, user.StripeSubscriberID.Int64)
	if err != nil {
		slog.ErrorContext(ctx, "Stripeサブスクライバーの取得に失敗しました", "error", err, "user_id", user.ID)
		h.redirectWithError(w, r, ctx, "supporters_portal_error")
		return
	}

	// アクティブなサブスクリプションのみ許可
	if !h.stripeSubscriberRepo.IsActive(&stripeSubscriber) {
		slog.WarnContext(ctx, "非アクティブなサブスクリプションでポータルにアクセスしようとしました", "user_id", user.ID, "status", stripeSubscriber.StripeStatus)
		h.redirectWithError(w, r, ctx, "supporters_portal_not_supporter")
		return
	}

	// Stripe Customer Portal セッションを作成
	returnURL := h.cfg.AppURL() + "/supporters"

	params := &stripe.BillingPortalSessionCreateParams{
		Customer:  stripe.String(stripeSubscriber.StripeCustomerID),
		ReturnURL: stripe.String(returnURL),
	}

	// ロケールの設定
	locale := i18n.GetLocale(ctx)
	if locale == "ja" {
		params.Locale = stripe.String("ja")
	} else {
		params.Locale = stripe.String("en")
	}

	// Stripeクライアントが設定されていない場合はエラー
	if h.stripeClient == nil {
		slog.ErrorContext(ctx, "Stripeクライアントが設定されていません", "user_id", user.ID)
		h.redirectWithError(w, r, ctx, "supporters_portal_error")
		return
	}

	portalSession, err := h.stripeClient.V1BillingPortalSessions.Create(ctx, params)
	if err != nil {
		slog.ErrorContext(ctx, "Stripe Customer Portalセッションの作成に失敗しました", "error", err, "user_id", user.ID)
		h.redirectWithError(w, r, ctx, "supporters_portal_error")
		return
	}

	slog.InfoContext(ctx, "Stripe Customer Portalセッションを作成しました", "user_id", user.ID)

	// Stripe Customer Portalページへリダイレクト
	http.Redirect(w, r, portalSession.URL, http.StatusSeeOther)
}

// redirectWithError はエラーメッセージをフラッシュに設定してリダイレクトします
func (h *Handler) redirectWithError(w http.ResponseWriter, r *http.Request, ctx context.Context, messageKey string) {
	// フラッシュメッセージを設定（"error"タイプ）
	_ = h.sessionManager.SetFlash(ctx, w, r, "error", i18n.T(ctx, messageKey))
	http.Redirect(w, r, "/supporters", http.StatusSeeOther)
}
