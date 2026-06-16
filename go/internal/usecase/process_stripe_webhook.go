package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/stripe/stripe-go/v84"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/repository"
	annictSentry "github.com/annict/annict/go/internal/sentry"
	annictstripe "github.com/annict/annict/go/internal/stripe"
)

// 処理対象のイベントタイプ
const (
	EventCheckoutSessionCompleted    = "checkout.session.completed"
	EventCustomerSubscriptionUpdated = "customer.subscription.updated"
	EventCustomerSubscriptionDeleted = "customer.subscription.deleted"
	EventInvoicePaymentSucceeded     = "invoice.payment_succeeded"
	EventInvoicePaymentFailed        = "invoice.payment_failed"
)

// ProcessStripeWebhookUsecase はStripe Webhookイベントを処理するユースケースです
type ProcessStripeWebhookUsecase struct {
	stripeWebhookEventRepo   *repository.StripeWebhookEventRepository
	createStripeSubscriberUC *CreateStripeSubscriberUsecase
	updateStripeSubscriberUC *UpdateStripeSubscriberUsecase
	deleteStripeSubscriberUC *DeleteStripeSubscriberUsecase
}

// NewProcessStripeWebhookUsecase は新しいProcessStripeWebhookUsecaseを作成します
func NewProcessStripeWebhookUsecase(
	stripeWebhookEventRepo *repository.StripeWebhookEventRepository,
	createStripeSubscriberUC *CreateStripeSubscriberUsecase,
	updateStripeSubscriberUC *UpdateStripeSubscriberUsecase,
	deleteStripeSubscriberUC *DeleteStripeSubscriberUsecase,
) *ProcessStripeWebhookUsecase {
	return &ProcessStripeWebhookUsecase{
		stripeWebhookEventRepo:   stripeWebhookEventRepo,
		createStripeSubscriberUC: createStripeSubscriberUC,
		updateStripeSubscriberUC: updateStripeSubscriberUC,
		deleteStripeSubscriberUC: deleteStripeSubscriberUC,
	}
}

// ProcessStripeWebhookInput はユースケースの入力です
type ProcessStripeWebhookInput struct {
	Event *stripe.Event
}

// ProcessStripeWebhookOutput はユースケースの出力です
type ProcessStripeWebhookOutput struct {
	Skipped bool // 冪等性チェックによりスキップされた場合true
}

// Execute はStripe Webhookイベントを処理します
//
// 処理フロー:
// 1. 冪等性チェック（既に処理済みのイベントはスキップ）
// 2. 新規イベントをDBに保存（status=pending）
// 3. イベントタイプに応じた処理を実行
// 4. 処理結果に応じてステータスを更新
func (uc *ProcessStripeWebhookUsecase) Execute(ctx context.Context, input ProcessStripeWebhookInput) (*ProcessStripeWebhookOutput, error) {
	event := input.Event

	// 冪等性チェック: 既存のイベントを確認し、ステータスに応じて処理を分岐
	var webhookEventID model.StripeWebhookEventID
	existingEvent, err := uc.stripeWebhookEventRepo.GetByStripeEventID(ctx, event.ID)
	if err != nil && err != sql.ErrNoRows {
		slog.ErrorContext(ctx, "Webhookイベント取得に失敗",
			"error", err,
			"stripe_event_id", event.ID,
		)
		annictSentry.CaptureError(ctx, fmt.Errorf("webhookイベント取得に失敗: %w", err))
		return &ProcessStripeWebhookOutput{}, nil
	}

	if err == nil {
		// 既存のイベントが見つかった場合
		status := model.WebhookEventStatus(existingEvent.Status)

		// processed または skipped の場合はスキップ
		if status == model.WebhookEventStatusProcessed || status == model.WebhookEventStatusSkipped {
			slog.InfoContext(ctx, "既に処理済みのイベントのためスキップ",
				"stripe_event_id", event.ID,
				"status", existingEvent.Status,
			)
			return &ProcessStripeWebhookOutput{Skipped: true}, nil
		}

		// pending または failed の場合は再処理を試みる
		slog.InfoContext(ctx, "未完了のイベントを再処理します",
			"stripe_event_id", event.ID,
			"status", existingEvent.Status,
		)
		webhookEventID = model.StripeWebhookEventID(existingEvent.ID)
	} else {
		// 新規イベントの場合はDBに保存（status=pending）
		payloadJSON, err := json.Marshal(event)
		if err != nil {
			slog.ErrorContext(ctx, "イベントペイロードのJSON変換に失敗",
				"error", err,
				"stripe_event_id", event.ID,
			)
			annictSentry.CaptureError(ctx, fmt.Errorf("イベントペイロードのJSON変換に失敗: %w", err))
			return &ProcessStripeWebhookOutput{}, nil
		}

		webhookEvent, err := uc.stripeWebhookEventRepo.Create(ctx, repository.CreateStripeWebhookEventParams{
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
			return &ProcessStripeWebhookOutput{}, nil
		}
		webhookEventID = model.StripeWebhookEventID(webhookEvent.ID)
	}

	// イベントタイプに応じた処理を実行
	if err := uc.processEvent(ctx, event, webhookEventID); err != nil {
		slog.ErrorContext(ctx, "Webhookイベント処理に失敗",
			"error", err,
			"stripe_event_id", event.ID,
			"event_type", event.Type,
		)
		annictSentry.CaptureError(ctx, fmt.Errorf("webhookイベント処理に失敗 (event_type=%s, stripe_event_id=%s): %w",
			event.Type, event.ID, err))

		// 処理失敗としてマーク
		if markErr := uc.stripeWebhookEventRepo.MarkAsFailed(ctx, webhookEventID, err.Error()); markErr != nil {
			slog.ErrorContext(ctx, "Webhookイベントの失敗マークに失敗",
				"error", markErr,
				"stripe_event_id", event.ID,
			)
		}
	}

	return &ProcessStripeWebhookOutput{}, nil
}

// processEvent はイベントタイプに応じた処理を実行します
func (uc *ProcessStripeWebhookUsecase) processEvent(ctx context.Context, event *stripe.Event, webhookEventID model.StripeWebhookEventID) error {
	switch event.Type {
	case EventCheckoutSessionCompleted:
		return uc.handleCheckoutSessionCompleted(ctx, event, webhookEventID)

	case EventCustomerSubscriptionUpdated:
		return uc.handleCustomerSubscriptionUpdated(ctx, event, webhookEventID)

	case EventCustomerSubscriptionDeleted:
		return uc.handleCustomerSubscriptionDeleted(ctx, event, webhookEventID)

	case EventInvoicePaymentSucceeded:
		slog.InfoContext(ctx, "支払い成功イベントを受信",
			"stripe_event_id", event.ID,
		)
		if err := uc.stripeWebhookEventRepo.MarkAsProcessed(ctx, webhookEventID); err != nil {
			return fmt.Errorf("イベント処理完了のマークに失敗: %w", err)
		}
		return nil

	case EventInvoicePaymentFailed:
		// Stripe側で自動リトライされるため、特別な処理は不要
		slog.WarnContext(ctx, "支払い失敗イベントを受信",
			"stripe_event_id", event.ID,
		)
		if err := uc.stripeWebhookEventRepo.MarkAsProcessed(ctx, webhookEventID); err != nil {
			return fmt.Errorf("イベント処理完了のマークに失敗: %w", err)
		}
		return nil

	default:
		slog.InfoContext(ctx, "処理対象外のイベントをスキップ",
			"stripe_event_id", event.ID,
			"event_type", event.Type,
		)
		if err := uc.stripeWebhookEventRepo.MarkAsSkipped(ctx, webhookEventID); err != nil {
			return fmt.Errorf("イベントスキップのマークに失敗: %w", err)
		}
		return nil
	}
}

// handleCheckoutSessionCompleted はcheckout.session.completedイベントを処理します
func (uc *ProcessStripeWebhookUsecase) handleCheckoutSessionCompleted(ctx context.Context, event *stripe.Event, webhookEventID model.StripeWebhookEventID) error {
	slog.InfoContext(ctx, "checkout.session.completedイベントを受信",
		"stripe_event_id", event.ID,
	)

	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		return fmt.Errorf("チェックアウトセッションのパースに失敗: %w", err)
	}

	// サブスクリプションIDがない場合はスキップ（一回限りの支払いなど）
	if session.Subscription == nil {
		slog.InfoContext(ctx, "サブスクリプションIDがないためスキップ",
			"stripe_event_id", event.ID,
		)
		if err := uc.stripeWebhookEventRepo.MarkAsSkipped(ctx, webhookEventID); err != nil {
			return fmt.Errorf("イベントスキップのマークに失敗: %w", err)
		}
		return nil
	}

	// Stripe may omit customer on a session, so guard before dereferencing
	// session.Customer.ID below. A missing customer means we cannot link the
	// subscription to a user, so skip the event safely instead of panicking.
	//
	// [Ja] Stripe はセッションで customer を欠落させて送ることがあるため、後段の
	// session.Customer.ID 参照より前にガードする。customer が無いとサブスクリプションを
	// ユーザーに紐付けられないため、panic させずイベントを安全にスキップする。
	if session.Customer == nil {
		slog.InfoContext(ctx, "顧客IDがないためスキップ",
			"stripe_event_id", event.ID,
		)
		if err := uc.stripeWebhookEventRepo.MarkAsSkipped(ctx, webhookEventID); err != nil {
			return fmt.Errorf("イベントスキップのマークに失敗: %w", err)
		}
		return nil
	}

	userID, err := ParseUserIDFromMetadata(session.Metadata)
	if err != nil {
		return fmt.Errorf("ユーザーID取得に失敗: %w", err)
	}

	input := CreateStripeSubscriberInput{
		StripeCustomerID:     session.Customer.ID,
		StripeSubscriptionID: session.Subscription.ID,
		UserID:               userID,
	}

	result, err := uc.createStripeSubscriberUC.Execute(ctx, input)
	if err != nil {
		return fmt.Errorf("StripeSubscriber作成に失敗: %w", err)
	}

	slog.InfoContext(ctx, "StripeSubscriberを作成しました",
		"stripe_event_id", event.ID,
		"stripe_subscriber_id", result.StripeSubscriber.ID,
		"user_id", userID,
	)

	if err := uc.stripeWebhookEventRepo.MarkAsProcessed(ctx, webhookEventID); err != nil {
		return fmt.Errorf("イベント処理完了のマークに失敗: %w", err)
	}

	return nil
}

// handleCustomerSubscriptionUpdated はcustomer.subscription.updatedイベントを処理します
func (uc *ProcessStripeWebhookUsecase) handleCustomerSubscriptionUpdated(ctx context.Context, event *stripe.Event, webhookEventID model.StripeWebhookEventID) error {
	slog.InfoContext(ctx, "customer.subscription.updatedイベントを受信",
		"stripe_event_id", event.ID,
	)

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
		if err := uc.stripeWebhookEventRepo.MarkAsSkipped(ctx, webhookEventID); err != nil {
			return fmt.Errorf("イベントスキップのマークに失敗: %w", err)
		}
		return nil
	}

	item := subscription.Items.Data[0]

	input := UpdateStripeSubscriberInput{
		StripeSubscriptionID:     subscription.ID,
		StripePriceID:            item.Price.ID,
		StripeStatus:             string(subscription.Status),
		StripeCurrentPeriodStart: time.Unix(item.CurrentPeriodStart, 0),
		StripeCurrentPeriodEnd:   time.Unix(item.CurrentPeriodEnd, 0),
		StripeCancelAt:           annictstripe.NullTimeFromUnix(subscription.CancelAt),
		StripeCanceledAt:         annictstripe.NullTimeFromUnix(subscription.CanceledAt),
	}

	result, err := uc.updateStripeSubscriberUC.Execute(ctx, input)
	if err != nil {
		// The update usecase wraps repository errors, so the not-found case must be
		// detected via errors.Is on the sentinel rather than == sql.ErrNoRows.
		//
		// [Ja] update ユースケースはリポジトリのエラーをラップするため、未存在判定は
		// == sql.ErrNoRows ではなく sentinel に対する errors.Is で行う。
		if errors.Is(err, ErrStripeSubscriberNotFound) {
			slog.WarnContext(ctx, "対応するStripeSubscriberが見つからないためスキップ",
				"stripe_event_id", event.ID,
				"subscription_id", subscription.ID,
			)
			if markErr := uc.stripeWebhookEventRepo.MarkAsSkipped(ctx, webhookEventID); markErr != nil {
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

	if err := uc.stripeWebhookEventRepo.MarkAsProcessed(ctx, webhookEventID); err != nil {
		return fmt.Errorf("イベント処理完了のマークに失敗: %w", err)
	}

	return nil
}

// handleCustomerSubscriptionDeleted はcustomer.subscription.deletedイベントを処理します
func (uc *ProcessStripeWebhookUsecase) handleCustomerSubscriptionDeleted(ctx context.Context, event *stripe.Event, webhookEventID model.StripeWebhookEventID) error {
	slog.InfoContext(ctx, "customer.subscription.deletedイベントを受信",
		"stripe_event_id", event.ID,
	)

	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return fmt.Errorf("サブスクリプションのパースに失敗: %w", err)
	}

	input := DeleteStripeSubscriberInput{
		StripeSubscriptionID: subscription.ID,
		StripeCanceledAt:     time.Unix(subscription.CanceledAt, 0),
	}

	result, err := uc.deleteStripeSubscriberUC.Execute(ctx, input)
	if err != nil {
		// The delete usecase wraps repository errors, so the not-found case must be
		// detected via errors.Is on the sentinel rather than == sql.ErrNoRows.
		//
		// [Ja] delete ユースケースはリポジトリのエラーをラップするため、未存在判定は
		// == sql.ErrNoRows ではなく sentinel に対する errors.Is で行う。
		if errors.Is(err, ErrStripeSubscriberNotFound) {
			slog.WarnContext(ctx, "対応するStripeSubscriberが見つからないためスキップ",
				"stripe_event_id", event.ID,
				"subscription_id", subscription.ID,
			)
			if markErr := uc.stripeWebhookEventRepo.MarkAsSkipped(ctx, webhookEventID); markErr != nil {
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

	if err := uc.stripeWebhookEventRepo.MarkAsProcessed(ctx, webhookEventID); err != nil {
		return fmt.Errorf("イベント処理完了のマークに失敗: %w", err)
	}

	return nil
}
