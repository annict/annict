package stripe

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/webhook"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	annictSentry "github.com/annict/annict/go/internal/sentry"
	annictstripe "github.com/annict/annict/go/internal/stripe"
	"github.com/annict/annict/go/internal/usecase"
)

// 処理対象のイベントタイプ
const (
	EventCheckoutSessionCompleted    = "checkout.session.completed"
	EventCustomerSubscriptionUpdated = "customer.subscription.updated"
	EventCustomerSubscriptionDeleted = "customer.subscription.deleted"
	EventInvoicePaymentSucceeded     = "invoice.payment_succeeded"
	EventInvoicePaymentFailed        = "invoice.payment_failed"
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

	// 冪等性チェック: 既に処理済みのイベントかどうか確認
	exists, err := h.stripeWebhookEventRepo.Exists(ctx, event.ID)
	if err != nil {
		slog.ErrorContext(ctx, "Webhookイベント存在チェックに失敗",
			"error", err,
			"stripe_event_id", event.ID,
		)
		annictSentry.CaptureError(ctx, fmt.Errorf("webhookイベント存在チェックに失敗: %w", err))
		// 内部エラーでも200を返す（Stripeのリトライを防ぐため）
		w.WriteHeader(http.StatusOK)
		return
	}

	if exists {
		slog.InfoContext(ctx, "既に処理済みのイベントのためスキップ",
			"stripe_event_id", event.ID,
		)
		w.WriteHeader(http.StatusOK)
		return
	}

	// イベントをDBに保存（status=pending）
	payloadJSON, err := json.Marshal(event)
	if err != nil {
		slog.ErrorContext(ctx, "イベントペイロードのJSON変換に失敗",
			"error", err,
			"stripe_event_id", event.ID,
		)
		annictSentry.CaptureError(ctx, fmt.Errorf("イベントペイロードのJSON変換に失敗: %w", err))
		w.WriteHeader(http.StatusOK)
		return
	}

	webhookEvent, err := h.stripeWebhookEventRepo.Create(ctx, repository.CreateStripeWebhookEventParams{
		StripeEventID:   event.ID,
		StripeEventType: string(event.Type),
		StripePayload:   payloadJSON,
		Status:          model.WebhookEventStatusPending.String(),
		ReceivedAt:      time.Now(),
	})
	if err != nil {
		slog.ErrorContext(ctx, "Webhookイベントの保存に失敗",
			"error", err,
			"stripe_event_id", event.ID,
		)
		annictSentry.CaptureError(ctx, fmt.Errorf("webhookイベントの保存に失敗: %w", err))
		w.WriteHeader(http.StatusOK)
		return
	}

	// イベントタイプに応じた処理を実行
	if err := h.processEvent(ctx, &event, webhookEvent.ID); err != nil {
		slog.ErrorContext(ctx, "Webhookイベント処理に失敗",
			"error", err,
			"stripe_event_id", event.ID,
			"event_type", event.Type,
		)
		annictSentry.CaptureError(ctx, fmt.Errorf("webhookイベント処理に失敗 (event_type=%s, stripe_event_id=%s): %w",
			event.Type, event.ID, err))

		// 処理失敗としてマーク
		if markErr := h.stripeWebhookEventRepo.MarkAsFailed(ctx, webhookEvent.ID, err.Error()); markErr != nil {
			slog.ErrorContext(ctx, "Webhookイベントの失敗マークに失敗",
				"error", markErr,
				"stripe_event_id", event.ID,
			)
		}
	}

	// Stripeには常に200を返す（リトライを防ぐため）
	w.WriteHeader(http.StatusOK)
}

// processEvent はイベントタイプに応じた処理を実行します
func (h *Handler) processEvent(ctx context.Context, event *stripe.Event, webhookEventID int64) error {
	switch event.Type {
	case EventCheckoutSessionCompleted:
		return h.handleCheckoutSessionCompleted(ctx, event, webhookEventID)

	case EventCustomerSubscriptionUpdated:
		return h.handleCustomerSubscriptionUpdated(ctx, event, webhookEventID)

	case EventCustomerSubscriptionDeleted:
		return h.handleCustomerSubscriptionDeleted(ctx, event, webhookEventID)

	case EventInvoicePaymentSucceeded:
		// ログ記録のみ（追加処理不要）
		slog.InfoContext(ctx, "支払い成功イベントを受信",
			"stripe_event_id", event.ID,
		)
		if err := h.stripeWebhookEventRepo.MarkAsProcessed(ctx, webhookEventID); err != nil {
			return fmt.Errorf("イベント処理完了のマークに失敗: %w", err)
		}
		return nil

	case EventInvoicePaymentFailed:
		// ログ記録のみ（Stripe側で自動リトライされるため、特別な処理は不要）
		slog.WarnContext(ctx, "支払い失敗イベントを受信",
			"stripe_event_id", event.ID,
		)
		if err := h.stripeWebhookEventRepo.MarkAsProcessed(ctx, webhookEventID); err != nil {
			return fmt.Errorf("イベント処理完了のマークに失敗: %w", err)
		}
		return nil

	default:
		// 処理対象外のイベント
		slog.InfoContext(ctx, "処理対象外のイベントをスキップ",
			"stripe_event_id", event.ID,
			"event_type", event.Type,
		)
		if err := h.stripeWebhookEventRepo.MarkAsSkipped(ctx, webhookEventID); err != nil {
			return fmt.Errorf("イベントスキップのマークに失敗: %w", err)
		}
		return nil
	}
}

// handleCheckoutSessionCompleted はcheckout.session.completedイベントを処理します
func (h *Handler) handleCheckoutSessionCompleted(ctx context.Context, event *stripe.Event, webhookEventID int64) error {
	slog.InfoContext(ctx, "checkout.session.completedイベントを受信",
		"stripe_event_id", event.ID,
	)

	// イベントデータからチェックアウトセッションを取得
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		return fmt.Errorf("チェックアウトセッションのパースに失敗: %w", err)
	}

	// サブスクリプションIDがない場合はスキップ（一回限りの支払いなど）
	if session.Subscription == nil {
		slog.InfoContext(ctx, "サブスクリプションIDがないためスキップ",
			"stripe_event_id", event.ID,
		)
		if err := h.stripeWebhookEventRepo.MarkAsSkipped(ctx, webhookEventID); err != nil {
			return fmt.Errorf("イベントスキップのマークに失敗: %w", err)
		}
		return nil
	}

	// metadataからユーザーIDを取得
	userID, err := usecase.ParseUserIDFromMetadata(session.Metadata)
	if err != nil {
		return fmt.Errorf("ユーザーID取得に失敗: %w", err)
	}

	// StripeSubscriberを作成してUserと紐付け
	input := usecase.CreateStripeSubscriberInput{
		StripeCustomerID:     session.Customer.ID,
		StripeSubscriptionID: session.Subscription.ID,
		UserID:               userID,
	}

	result, err := h.createStripeSubscriberUC.Execute(ctx, input)
	if err != nil {
		return fmt.Errorf("StripeSubscriber作成に失敗: %w", err)
	}

	slog.InfoContext(ctx, "StripeSubscriberを作成しました",
		"stripe_event_id", event.ID,
		"stripe_subscriber_id", result.StripeSubscriber.ID,
		"user_id", userID,
	)

	// 処理完了としてマーク
	if err := h.stripeWebhookEventRepo.MarkAsProcessed(ctx, webhookEventID); err != nil {
		return fmt.Errorf("イベント処理完了のマークに失敗: %w", err)
	}

	return nil
}

// handleCustomerSubscriptionUpdated はcustomer.subscription.updatedイベントを処理します
func (h *Handler) handleCustomerSubscriptionUpdated(ctx context.Context, event *stripe.Event, webhookEventID int64) error {
	slog.InfoContext(ctx, "customer.subscription.updatedイベントを受信",
		"stripe_event_id", event.ID,
	)

	// イベントデータからサブスクリプションを取得
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return fmt.Errorf("サブスクリプションのパースに失敗: %w", err)
	}

	// サブスクリプションにアイテムが含まれていない場合はスキップ
	if subscription.Items == nil || len(subscription.Items.Data) == 0 {
		slog.InfoContext(ctx, "サブスクリプションにアイテムが含まれていないためスキップ",
			"stripe_event_id", event.ID,
			"subscription_id", subscription.ID,
		)
		if err := h.stripeWebhookEventRepo.MarkAsSkipped(ctx, webhookEventID); err != nil {
			return fmt.Errorf("イベントスキップのマークに失敗: %w", err)
		}
		return nil
	}

	// 価格IDと請求期間を取得（最初のアイテムから）
	item := subscription.Items.Data[0]

	// StripeSubscriberを更新
	input := usecase.UpdateStripeSubscriberInput{
		StripeSubscriptionID:     subscription.ID,
		StripePriceID:            item.Price.ID,
		StripeStatus:             string(subscription.Status),
		StripeCurrentPeriodStart: time.Unix(item.CurrentPeriodStart, 0),
		StripeCurrentPeriodEnd:   time.Unix(item.CurrentPeriodEnd, 0),
		StripeCancelAt:           annictstripe.NullTimeFromUnix(subscription.CancelAt),
		StripeCanceledAt:         annictstripe.NullTimeFromUnix(subscription.CanceledAt),
	}

	result, err := h.updateStripeSubscriberUC.Execute(ctx, input)
	if err != nil {
		// サブスクリプションが見つからない場合はスキップ
		if err == sql.ErrNoRows {
			slog.WarnContext(ctx, "対応するStripeSubscriberが見つからないためスキップ",
				"stripe_event_id", event.ID,
				"subscription_id", subscription.ID,
			)
			if markErr := h.stripeWebhookEventRepo.MarkAsSkipped(ctx, webhookEventID); markErr != nil {
				return fmt.Errorf("イベントスキップのマークに失敗: %w", markErr)
			}
			return nil
		}
		return fmt.Errorf("StripeSubscriber更新に失敗: %w", err)
	}

	slog.InfoContext(ctx, "StripeSubscriberを更新しました",
		"stripe_event_id", event.ID,
		"stripe_subscriber_id", result.StripeSubscriber.ID,
		"new_status", subscription.Status,
	)

	// 処理完了としてマーク
	if err := h.stripeWebhookEventRepo.MarkAsProcessed(ctx, webhookEventID); err != nil {
		return fmt.Errorf("イベント処理完了のマークに失敗: %w", err)
	}

	return nil
}

// handleCustomerSubscriptionDeleted はcustomer.subscription.deletedイベントを処理します
func (h *Handler) handleCustomerSubscriptionDeleted(ctx context.Context, event *stripe.Event, webhookEventID int64) error {
	slog.InfoContext(ctx, "customer.subscription.deletedイベントを受信",
		"stripe_event_id", event.ID,
	)

	// イベントデータからサブスクリプションを取得
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return fmt.Errorf("サブスクリプションのパースに失敗: %w", err)
	}

	// StripeSubscriberを削除（ステータス更新＋ユーザー紐付け解除）
	input := usecase.DeleteStripeSubscriberInput{
		StripeSubscriptionID: subscription.ID,
		StripeCanceledAt:     time.Unix(subscription.CanceledAt, 0),
	}

	result, err := h.updateStripeSubscriberUC.ExecuteDelete(ctx, input)
	if err != nil {
		// サブスクリプションが見つからない場合はスキップ
		if err == sql.ErrNoRows {
			slog.WarnContext(ctx, "対応するStripeSubscriberが見つからないためスキップ",
				"stripe_event_id", event.ID,
				"subscription_id", subscription.ID,
			)
			if markErr := h.stripeWebhookEventRepo.MarkAsSkipped(ctx, webhookEventID); markErr != nil {
				return fmt.Errorf("イベントスキップのマークに失敗: %w", markErr)
			}
			return nil
		}
		return fmt.Errorf("StripeSubscriber削除処理に失敗: %w", err)
	}

	logFields := []any{
		"stripe_event_id", event.ID,
		"stripe_subscriber_id", result.StripeSubscriber.ID,
	}
	if result.UserID != nil {
		logFields = append(logFields, "unlinked_user_id", *result.UserID)
	}
	slog.InfoContext(ctx, "StripeSubscriberを削除しました", logFields...)

	// 処理完了としてマーク
	if err := h.stripeWebhookEventRepo.MarkAsProcessed(ctx, webhookEventID); err != nil {
		return fmt.Errorf("イベント処理完了のマークに失敗: %w", err)
	}

	return nil
}
