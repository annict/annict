package supporters

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/stripe/stripe-go/v84"

	"github.com/annict/annict/go/internal/i18n"
	authMiddleware "github.com/annict/annict/go/internal/middleware"
)

// CheckoutRequest はCheckoutセッション作成のリクエストパラメータ
type CheckoutRequest struct {
	Plan string
}

// Validate はリクエストパラメータのバリデーションを行います
func (r *CheckoutRequest) Validate() map[string]string {
	errors := make(map[string]string)
	if r.Plan != "monthly" && r.Plan != "yearly" {
		errors["plan"] = "invalid_plan"
	}
	return errors
}

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

	req := &CheckoutRequest{
		Plan: r.FormValue("plan"),
	}

	// バリデーション
	if errs := req.Validate(); len(errs) > 0 {
		slog.WarnContext(ctx, "バリデーションエラー", "errors", errs, "user_id", user.ID)
		h.redirectWithError(w, r, ctx, "supporters_checkout_invalid_plan")
		return
	}

	// 重複サブスクリプションチェック
	// 1. ユーザーに紐づくstripe_subscriber_idが存在しアクティブな場合はエラーを返す
	if user.StripeSubscriberID.Valid && h.stripeSubscriberRepo != nil {
		stripeSubscriber, err := h.stripeSubscriberRepo.GetByID(ctx, user.StripeSubscriberID.Int64)
		if err == nil {
			if h.stripeSubscriberRepo.IsActive(&stripeSubscriber) {
				slog.InfoContext(ctx, "既にアクティブなサブスクリプションが存在します", "user_id", user.ID, "stripe_subscriber_id", user.StripeSubscriberID.Int64)
				h.redirectWithError(w, r, ctx, "supporters_checkout_already_active")
				return
			}
		}
	}

	// 価格IDの決定
	var priceID string
	switch req.Plan {
	case "monthly":
		priceID = h.stripeCfg.PriceMonthlyID
	case "yearly":
		priceID = h.stripeCfg.PriceYearlyID
	}

	if priceID == "" {
		slog.ErrorContext(ctx, "Stripe価格IDが設定されていません", "plan", req.Plan)
		h.redirectWithError(w, r, ctx, "supporters_checkout_error")
		return
	}

	// Stripe Checkoutセッションの作成
	successURL := h.cfg.AppURL() + "/supporters?success=true"
	cancelURL := h.cfg.AppURL() + "/supporters?canceled=true"

	params := &stripe.CheckoutSessionCreateParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionCreateLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Metadata: map[string]string{
			"user_id": strconv.FormatInt(user.ID, 10),
		},
	}

	// ロケールの設定
	locale := i18n.GetLocale(ctx)
	if locale == "ja" {
		params.Locale = stripe.String("ja")
	} else {
		params.Locale = stripe.String("en")
	}

	checkoutSession, err := h.stripeClient.V1CheckoutSessions.Create(ctx, params)
	if err != nil {
		slog.ErrorContext(ctx, "Stripe Checkoutセッションの作成に失敗しました", "error", err, "user_id", user.ID)
		h.redirectWithError(w, r, ctx, "supporters_checkout_error")
		return
	}

	slog.InfoContext(ctx, "Stripe Checkoutセッションを作成しました", "session_id", checkoutSession.ID, "user_id", user.ID, "plan", req.Plan)

	// Stripe Checkoutページへリダイレクト
	http.Redirect(w, r, checkoutSession.URL, http.StatusSeeOther)
}

// redirectWithError はエラーメッセージをフラッシュに設定してリダイレクトします
func (h *Handler) redirectWithError(w http.ResponseWriter, r *http.Request, ctx context.Context, messageKey string) {
	// フラッシュメッセージを設定（"error"タイプ）
	_ = h.sessionManager.SetFlash(ctx, w, r, "error", i18n.T(ctx, messageKey))
	http.Redirect(w, r, "/supporters", http.StatusSeeOther)
}
