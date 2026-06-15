package usecase

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stripe/stripe-go/v84"

	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// marshalSubscriptionEvent builds a stripe.Event whose Data.Raw is the JSON of the
// given subscription, mirroring how Stripe delivers customer.subscription.* events.
//
// [Ja] 与えられた subscription を JSON 化して Data.Raw に持つ stripe.Event を組み立てる。
// Stripe が customer.subscription.* イベントを配信する形を模している。
func marshalSubscriptionEvent(t *testing.T, eventID string, eventType stripe.EventType, sub stripe.Subscription) *stripe.Event {
	t.Helper()

	raw, err := json.Marshal(sub)
	if err != nil {
		t.Fatalf("サブスクリプションのJSON変換に失敗: %v", err)
	}

	return &stripe.Event{
		ID:   eventID,
		Type: eventType,
		Data: &stripe.EventData{Raw: raw},
	}
}

// TestProcessStripeWebhookUsecase_SkipsWhenSubscriberNotFound is a regression test
// for a confirmed bug: an update / deleted event whose StripeSubscriber does not
// exist must be recorded as skipped, not failed. The webhook layer previously
// compared the wrapped error with == sql.ErrNoRows, so the not-found branch was
// unreachable and the event was treated as failed.
//
// [Ja] 対応する StripeSubscriber が存在しない update / deleted イベントが failed では
// なく skipped として記録されることを検証する確定バグの回帰テスト。以前は Webhook 層が
// ラップ済みエラーを == sql.ErrNoRows で比較していたため、未存在分岐が到達不能で failed
// 扱いになっていた。
func TestProcessStripeWebhookUsecase_SkipsWhenSubscriberNotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		eventType stripe.EventType
		buildSub  func(subID string) stripe.Subscription
	}{
		{
			name:      "customer.subscription.updated: 未存在subscriberはskipped",
			eventType: stripe.EventTypeCustomerSubscriptionUpdated,
			buildSub: func(subID string) stripe.Subscription {
				return stripe.Subscription{
					ID:     subID,
					Status: stripe.SubscriptionStatusActive,
					Items: &stripe.SubscriptionItemList{
						Data: []*stripe.SubscriptionItem{
							{
								Price:              &stripe.Price{ID: "price_monthly"},
								CurrentPeriodStart: time.Now().Unix(),
								CurrentPeriodEnd:   time.Now().AddDate(0, 1, 0).Unix(),
							},
						},
					},
				}
			},
		},
		{
			name:      "customer.subscription.deleted: 未存在subscriberはskipped",
			eventType: stripe.EventTypeCustomerSubscriptionDeleted,
			buildSub: func(subID string) stripe.Subscription {
				return stripe.Subscription{
					ID:         subID,
					Status:     stripe.SubscriptionStatusCanceled,
					CanceledAt: time.Now().Unix(),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, tx := testutil.SetupTx(t)
			queries := query.New(db).WithTx(tx)

			stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
			userRepo := repository.NewUserRepository(queries)
			stripeWebhookEventRepo := repository.NewStripeWebhookEventRepository(queries)

			// The create usecase is not exercised by these cases, so a nil Stripe
			// client is sufficient.
			//
			// [Ja] これらのケースでは create ユースケースは実行されないため、Stripe
			// クライアントは nil で十分。
			createUC := NewCreateStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo, nil)
			updateUC := NewUpdateStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo)
			deleteUC := NewDeleteStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo)
			uc := NewProcessStripeWebhookUsecase(stripeWebhookEventRepo, createUC, updateUC, deleteUC)

			ctx := context.Background()

			// No StripeSubscriber row exists for this subscription ID.
			// [Ja] 対応する StripeSubscriber が存在しないサブスクリプション ID。
			subID := "sub_webhook_notfound_" + randomString(8)
			eventID := "evt_webhook_notfound_" + randomString(8)
			event := marshalSubscriptionEvent(t, eventID, tt.eventType, tt.buildSub(subID))

			if _, err := uc.Execute(ctx, ProcessStripeWebhookInput{Event: event}); err != nil {
				t.Fatalf("Webhook処理で予期しないエラー: %v", err)
			}

			got, err := stripeWebhookEventRepo.GetByStripeEventID(ctx, eventID)
			if err != nil {
				t.Fatalf("Webhookイベントの取得に失敗: %v", err)
			}
			if got.Status != model.WebhookEventStatusSkipped.String() {
				t.Errorf("イベントステータス: got %s, want %s", got.Status, model.WebhookEventStatusSkipped)
			}
		})
	}
}
