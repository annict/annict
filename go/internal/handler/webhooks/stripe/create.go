package stripe

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/stripe/stripe-go/v84/webhook"

	"github.com/annict/annict/go/internal/usecase"
)

// Create はStripe Webhookを受信して処理します (POST /webhooks/stripe)
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// リクエストボディを読み取り（Stripe署名検証に必要）
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.ErrorContext(ctx, "Webhookリクエストボディの読み取りに失敗", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Stripe署名を取得
	sigHeader := r.Header.Get("Stripe-Signature")
	if sigHeader == "" {
		slog.WarnContext(ctx, "Stripe-Signatureヘッダーがありません")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Stripe署名検証
	event, err := webhook.ConstructEvent(body, sigHeader, h.cfg.StripeWebhookSecret)
	if err != nil {
		slog.WarnContext(ctx, "Webhook署名検証に失敗", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	slog.InfoContext(ctx, "Stripe Webhookを受信",
		"event_id", event.ID,
		"event_type", event.Type,
	)

	// UseCaseに処理を委譲
	_, err = h.processStripeWebhookUC.Execute(ctx, usecase.ProcessStripeWebhookInput{
		Event: &event,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Webhook処理に失敗", "error", err)
	}

	// Stripeには常に200を返す（リトライを防ぐため）
	w.WriteHeader(http.StatusOK)
}
