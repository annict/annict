package usecase

import (
	"context"
	"database/sql"
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
			//
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

// marshalCheckoutSessionEvent builds a checkout.session.completed stripe.Event
// whose Data.Raw is the JSON of the given session, mirroring how Stripe delivers
// the event.
//
// [Ja] 与えられた session を JSON 化して Data.Raw に持つ checkout.session.completed の
// stripe.Event を組み立てる。Stripe がイベントを配信する形を模している。
func marshalCheckoutSessionEvent(t *testing.T, eventID string, session stripe.CheckoutSession) *stripe.Event {
	t.Helper()

	raw, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("チェックアウトセッションのJSON変換に失敗: %v", err)
	}

	return &stripe.Event{
		ID:   eventID,
		Type: stripe.EventTypeCheckoutSessionCompleted,
		Data: &stripe.EventData{Raw: raw},
	}
}

// TestProcessStripeWebhookUsecase_HandleCheckoutSessionCompleted exercises the
// checkout.session.completed path: the happy path creates a StripeSubscriber and
// links the user, while malformed sessions are either skipped (missing
// subscription / customer) or failed (missing / invalid metadata user_id). The
// missing-customer case is the regression guard for a panic on session.Customer.ID.
//
// [Ja] checkout.session.completed 経路を検証する。正常系では StripeSubscriber を作成し
// ユーザーを紐付ける。不正なセッションは skipped (subscription / customer 欠落) または
// failed (metadata の user_id 欠落 / 不正) になる。customer 欠落ケースは
// session.Customer.ID 参照での panic に対する回帰ガードである。
func TestProcessStripeWebhookUsecase_HandleCheckoutSessionCompleted(t *testing.T) {
	t.Parallel()

	// handleCheckoutSessionCompleted invokes CreateStripeSubscriberUsecase, which
	// opens its own transaction, so use the shared DB directly and commit test data
	// (GetTestDB) rather than an outer rollback transaction whose data the inner
	// transaction could not see.
	//
	// [Ja] handleCheckoutSessionCompleted は内部でトランザクションを開く
	// CreateStripeSubscriberUsecase を呼ぶため、内側のトランザクションから見えない
	// 外側のロールバック用トランザクションではなく、共有 DB を直接使ってテストデータを
	// コミットする (GetTestDB)。
	db := testutil.GetTestDB()
	queries := query.New(db)
	stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
	userRepo := repository.NewUserRepository(queries)
	stripeWebhookEventRepo := repository.NewStripeWebhookEventRepository(queries)

	updateUC := NewUpdateStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo)
	deleteUC := NewDeleteStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo)

	t.Run("正常系: StripeSubscriber作成・ユーザー紐付け・processed", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		userID := insertStripeTestUser(t, db)
		customerID := "cus_completed_" + randomString(8)
		subscriptionID := "sub_completed_" + randomString(8)
		eventID := "evt_completed_" + randomString(8)

		t.Cleanup(func() {
			_, _ = db.Exec("DELETE FROM stripe_webhook_events WHERE stripe_event_id = $1", eventID)
			_, _ = db.Exec("DELETE FROM users WHERE id = $1", int64(userID))
			_, _ = db.Exec("DELETE FROM stripe_subscribers WHERE stripe_subscription_id = $1", subscriptionID)
		})

		createUC := NewCreateStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo, &fakeSubscriptionRetriever{
			subscription: validSubscription(),
		})
		uc := NewProcessStripeWebhookUsecase(stripeWebhookEventRepo, createUC, updateUC, deleteUC)

		event := marshalCheckoutSessionEvent(t, eventID, stripe.CheckoutSession{
			Subscription: &stripe.Subscription{ID: subscriptionID},
			Customer:     &stripe.Customer{ID: customerID},
			Metadata:     map[string]string{"user_id": userID.String()},
		})

		if _, err := uc.Execute(ctx, ProcessStripeWebhookInput{Event: event}); err != nil {
			t.Fatalf("Webhook処理で予期しないエラー: %v", err)
		}

		gotEvent, err := stripeWebhookEventRepo.GetByStripeEventID(ctx, eventID)
		if err != nil {
			t.Fatalf("Webhookイベントの取得に失敗: %v", err)
		}
		if gotEvent.Status != model.WebhookEventStatusProcessed.String() {
			t.Errorf("イベントステータス: got %s, want %s", gotEvent.Status, model.WebhookEventStatusProcessed)
		}

		// The StripeSubscriber must be created and the customer ID from the session
		// carried onto the record.
		//
		// [Ja] StripeSubscriber が作成され、セッションの customer ID がレコードへ
		// 引き継がれることを確認する。
		subscriber, err := stripeSubscriberRepo.GetByStripeSubscriptionID(ctx, subscriptionID)
		if err != nil {
			t.Fatalf("StripeSubscriber取得エラー: %v", err)
		}
		if subscriber == nil {
			t.Fatal("StripeSubscriber が作成されていません")
		}
		if subscriber.StripeCustomerID != customerID {
			t.Errorf("StripeCustomerID: got %s, want %s", subscriber.StripeCustomerID, customerID)
		}

		linked, err := userRepo.GetByID(ctx, userID)
		if err != nil {
			t.Fatalf("ユーザー取得エラー: %v", err)
		}
		if !linked.StripeSubscriberID.Valid {
			t.Fatal("ユーザーに StripeSubscriber が紐付けられていません")
		}
		if linked.StripeSubscriberID.Int64 != int64(subscriber.ID) {
			t.Errorf("紐付けられた StripeSubscriberID: got %d, want %d", linked.StripeSubscriberID.Int64, int64(subscriber.ID))
		}
	})

	// These cases short-circuit before CreateStripeSubscriberUsecase is reached, so
	// they neither create a subscriber nor touch a user; only the webhook event row
	// is written and asserted.
	//
	// [Ja] これらのケースは CreateStripeSubscriberUsecase 到達前に短絡するため、
	// subscriber も作らずユーザーにも触れない。書き込まれて検証されるのは Webhook
	// イベント行のみ。
	malformedCases := []struct {
		name         string
		buildSession func() stripe.CheckoutSession
		wantStatus   model.WebhookEventStatus
	}{
		{
			name: "subscription欠落はskipped (一回限りの支払いなど)",
			buildSession: func() stripe.CheckoutSession {
				return stripe.CheckoutSession{
					Customer: &stripe.Customer{ID: "cus_nosub_" + randomString(8)},
					Metadata: map[string]string{"user_id": "1"},
				}
			},
			wantStatus: model.WebhookEventStatusSkipped,
		},
		{
			name: "customer欠落はskipped (nilガードの検証)",
			buildSession: func() stripe.CheckoutSession {
				return stripe.CheckoutSession{
					Subscription: &stripe.Subscription{ID: "sub_nocus_" + randomString(8)},
					Metadata:     map[string]string{"user_id": "1"},
				}
			},
			wantStatus: model.WebhookEventStatusSkipped,
		},
		{
			name: "metadataのuser_id欠落はfailed",
			buildSession: func() stripe.CheckoutSession {
				return stripe.CheckoutSession{
					Subscription: &stripe.Subscription{ID: "sub_noid_" + randomString(8)},
					Customer:     &stripe.Customer{ID: "cus_noid_" + randomString(8)},
					Metadata:     map[string]string{},
				}
			},
			wantStatus: model.WebhookEventStatusFailed,
		},
		{
			name: "metadataのuser_idが不正はfailed",
			buildSession: func() stripe.CheckoutSession {
				return stripe.CheckoutSession{
					Subscription: &stripe.Subscription{ID: "sub_badid_" + randomString(8)},
					Customer:     &stripe.Customer{ID: "cus_badid_" + randomString(8)},
					Metadata:     map[string]string{"user_id": "not-a-number"},
				}
			},
			wantStatus: model.WebhookEventStatusFailed,
		},
	}

	for _, tc := range malformedCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			eventID := "evt_completed_malformed_" + randomString(8)
			t.Cleanup(func() {
				_, _ = db.Exec("DELETE FROM stripe_webhook_events WHERE stripe_event_id = $1", eventID)
			})

			// A nil Stripe client is sufficient: none of these cases reach the create
			// usecase.
			//
			// [Ja] これらのケースは create ユースケースに到達しないため、Stripe
			// クライアントは nil で十分。
			createUC := NewCreateStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo, nil)
			uc := NewProcessStripeWebhookUsecase(stripeWebhookEventRepo, createUC, updateUC, deleteUC)

			event := marshalCheckoutSessionEvent(t, eventID, tc.buildSession())
			if _, err := uc.Execute(ctx, ProcessStripeWebhookInput{Event: event}); err != nil {
				t.Fatalf("Webhook処理で予期しないエラー: %v", err)
			}

			got, err := stripeWebhookEventRepo.GetByStripeEventID(ctx, eventID)
			if err != nil {
				t.Fatalf("Webhookイベントの取得に失敗: %v", err)
			}
			if got.Status != tc.wantStatus.String() {
				t.Errorf("イベントステータス: got %s, want %s", got.Status, tc.wantStatus)
			}
		})
	}
}

// newWebhookUsecaseForTest wires a ProcessStripeWebhookUsecase against the shared
// test DB and returns it together with the repositories used to seed and assert.
// The subscription-update / delete / invoice / unhandled paths under test never
// reach CreateStripeSubscriberUsecase, so a nil SubscriptionRetriever is enough.
//
// [Ja] テスト用の共有 DB に対して ProcessStripeWebhookUsecase を組み立て、仕込みと
// 検証に使う Repository と一緒に返す。本テストが対象とする subscription の更新 / 削除
// / invoice / 対象外イベントの経路はいずれも CreateStripeSubscriberUsecase に到達しない
// ため、SubscriptionRetriever は nil で十分。
func newWebhookUsecaseForTest(db *sql.DB) (
	*ProcessStripeWebhookUsecase,
	*repository.StripeSubscriberRepository,
	*repository.UserRepository,
	*repository.StripeWebhookEventRepository,
) {
	queries := query.New(db)
	stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
	userRepo := repository.NewUserRepository(queries)
	stripeWebhookEventRepo := repository.NewStripeWebhookEventRepository(queries)

	createUC := NewCreateStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo, nil)
	updateUC := NewUpdateStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo)
	deleteUC := NewDeleteStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo)
	uc := NewProcessStripeWebhookUsecase(stripeWebhookEventRepo, createUC, updateUC, deleteUC)

	return uc, stripeSubscriberRepo, userRepo, stripeWebhookEventRepo
}

// TestProcessStripeWebhookUsecase_HandleCustomerSubscriptionUpdated exercises the
// customer.subscription.updated path: an event carrying items updates the existing
// subscriber and is marked processed, an event without items is skipped before the
// update usecase is reached, and an event whose status is invalid is rejected by the
// update usecase and recorded as failed (the processing-failure path that ends in
// MarkAsFailed). The not-found subscriber case is covered by
// TestProcessStripeWebhookUsecase_SkipsWhenSubscriberNotFound.
//
// [Ja] customer.subscription.updated 経路を検証する。items 付きイベントは既存
// subscriber を更新して processed になり、items を持たないイベントは update ユースケース
// 到達前に skipped となり、ステータスが無効なイベントは update ユースケースに弾かれて
// failed として記録される (MarkAsFailed に至る処理失敗経路)。未存在 subscriber のケースは
// TestProcessStripeWebhookUsecase_SkipsWhenSubscriberNotFound が担当する。
func TestProcessStripeWebhookUsecase_HandleCustomerSubscriptionUpdated(t *testing.T) {
	t.Parallel()

	// UpdateStripeSubscriberUsecase reads and writes the subscriber directly without
	// opening an inner transaction, and the fail case rejects the event before any
	// lookup. Both rely on committed data, so use GetTestDB and clean up added rows.
	//
	// [Ja] UpdateStripeSubscriberUsecase は内部トランザクションを開かず subscriber を
	// 直接読み書きし、処理失敗ケースは参照前にイベントを弾く。どちらもコミット済みの
	// データに依存するため、GetTestDB を使い追加した行をクリーンアップする。
	db := testutil.GetTestDB()
	uc, stripeSubscriberRepo, _, stripeWebhookEventRepo := newWebhookUsecaseForTest(db)

	t.Run("正常系: items付きペイロードでupdate + processed", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		subscriptionID := "sub_updated_" + randomString(8)
		eventID := "evt_updated_" + randomString(8)

		seeded, err := stripeSubscriberRepo.Create(ctx, query.CreateStripeSubscriberParams{
			StripeCustomerID:         "cus_updated_" + randomString(8),
			StripeSubscriptionID:     subscriptionID,
			StripePriceID:            "price_monthly",
			StripeStatus:             string(model.StripeSubscriptionStatusActive),
			StripeCurrentPeriodStart: time.Now(),
			StripeCurrentPeriodEnd:   time.Now().AddDate(0, 1, 0),
		})
		if err != nil {
			t.Fatalf("シード用 StripeSubscriber の作成に失敗: %v", err)
		}
		t.Cleanup(func() {
			_, _ = db.Exec("DELETE FROM stripe_webhook_events WHERE stripe_event_id = $1", eventID)
			_, _ = db.Exec("DELETE FROM stripe_subscribers WHERE id = $1", int64(seeded.ID))
		})

		// The updated event carries a new price; the persisted record must reflect it.
		//
		// [Ja] updated イベントは新しい price を載せる。永続化レコードがそれを反映する必要がある。
		event := marshalSubscriptionEvent(t, eventID, stripe.EventTypeCustomerSubscriptionUpdated, stripe.Subscription{
			ID:     subscriptionID,
			Status: stripe.SubscriptionStatusActive,
			Items: &stripe.SubscriptionItemList{
				Data: []*stripe.SubscriptionItem{
					{
						Price:              &stripe.Price{ID: "price_yearly"},
						CurrentPeriodStart: time.Now().Unix(),
						CurrentPeriodEnd:   time.Now().AddDate(1, 0, 0).Unix(),
					},
				},
			},
		})

		if _, err := uc.Execute(ctx, ProcessStripeWebhookInput{Event: event}); err != nil {
			t.Fatalf("Webhook処理で予期しないエラー: %v", err)
		}

		gotEvent, err := stripeWebhookEventRepo.GetByStripeEventID(ctx, eventID)
		if err != nil {
			t.Fatalf("Webhookイベントの取得に失敗: %v", err)
		}
		if gotEvent.Status != model.WebhookEventStatusProcessed.String() {
			t.Errorf("イベントステータス: got %s, want %s", gotEvent.Status, model.WebhookEventStatusProcessed)
		}

		updated, err := stripeSubscriberRepo.GetByStripeSubscriptionID(ctx, subscriptionID)
		if err != nil {
			t.Fatalf("StripeSubscriber取得エラー: %v", err)
		}
		if updated == nil {
			t.Fatal("StripeSubscriber が見つかりません")
		}
		if updated.StripePriceID != "price_yearly" {
			t.Errorf("StripePriceID: got %s, want price_yearly", updated.StripePriceID)
		}
	})

	t.Run("処理失敗: 無効なステータスはfailed (MarkAsFailed)", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		eventID := "evt_updated_fail_" + randomString(8)
		t.Cleanup(func() {
			_, _ = db.Exec("DELETE FROM stripe_webhook_events WHERE stripe_event_id = $1", eventID)
		})

		// An invalid status is rejected by UpdateStripeSubscriberUsecase before any
		// repository lookup, so no subscriber is seeded. The error is not
		// ErrStripeSubscriberNotFound, so the webhook records failed (not skipped).
		//
		// [Ja] 無効なステータスは UpdateStripeSubscriberUsecase がリポジトリ参照前に弾く
		// ため、subscriber は仕込まない。このエラーは ErrStripeSubscriberNotFound では
		// ないため、Webhook は (skipped ではなく) failed として記録する。
		event := marshalSubscriptionEvent(t, eventID, stripe.EventTypeCustomerSubscriptionUpdated, stripe.Subscription{
			ID:     "sub_updated_fail_" + randomString(8),
			Status: stripe.SubscriptionStatus("invalid_status"),
			Items: &stripe.SubscriptionItemList{
				Data: []*stripe.SubscriptionItem{
					{
						Price:              &stripe.Price{ID: "price_monthly"},
						CurrentPeriodStart: time.Now().Unix(),
						CurrentPeriodEnd:   time.Now().AddDate(0, 1, 0).Unix(),
					},
				},
			},
		})

		if _, err := uc.Execute(ctx, ProcessStripeWebhookInput{Event: event}); err != nil {
			t.Fatalf("Webhook処理で予期しないエラー: %v", err)
		}

		got, err := stripeWebhookEventRepo.GetByStripeEventID(ctx, eventID)
		if err != nil {
			t.Fatalf("Webhookイベントの取得に失敗: %v", err)
		}
		if got.Status != model.WebhookEventStatusFailed.String() {
			t.Errorf("イベントステータス: got %s, want %s", got.Status, model.WebhookEventStatusFailed)
		}
	})

	t.Run("items空はskipped", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		eventID := "evt_updated_noitems_" + randomString(8)
		t.Cleanup(func() {
			_, _ = db.Exec("DELETE FROM stripe_webhook_events WHERE stripe_event_id = $1", eventID)
		})

		// An updated event without items is skipped by handleCustomerSubscriptionUpdated
		// before the update usecase (and any repository lookup) is reached, so no
		// subscriber is seeded.
		//
		// [Ja] items を持たない updated イベントは handleCustomerSubscriptionUpdated が
		// update ユースケース (およびリポジトリ参照) 到達前にスキップするため、subscriber は
		// 仕込まない。
		event := marshalSubscriptionEvent(t, eventID, stripe.EventTypeCustomerSubscriptionUpdated, stripe.Subscription{
			ID:     "sub_updated_noitems_" + randomString(8),
			Status: stripe.SubscriptionStatusActive,
		})

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

// TestProcessStripeWebhookUsecase_HandleCustomerSubscriptionDeleted exercises the
// customer.subscription.deleted path: the existing subscriber is moved to canceled,
// the linked user is unlinked, and the event is marked processed. The not-found
// subscriber case is covered by TestProcessStripeWebhookUsecase_SkipsWhenSubscriberNotFound.
//
// [Ja] customer.subscription.deleted 経路を検証する。既存 subscriber は canceled に
// 更新され、紐付くユーザーは紐付け解除され、イベントは processed になる。未存在
// subscriber のケースは TestProcessStripeWebhookUsecase_SkipsWhenSubscriberNotFound
// が担当する。
func TestProcessStripeWebhookUsecase_HandleCustomerSubscriptionDeleted(t *testing.T) {
	t.Parallel()

	// DeleteStripeSubscriberUsecase opens its own transaction, so the seed data must
	// be committed (GetTestDB) for that inner transaction to see it.
	//
	// [Ja] DeleteStripeSubscriberUsecase は自前のトランザクションを開くため、前提データは
	// コミット (GetTestDB) して内側のトランザクションから見えるようにする必要がある。
	db := testutil.GetTestDB()
	uc, stripeSubscriberRepo, userRepo, stripeWebhookEventRepo := newWebhookUsecaseForTest(db)

	t.Run("正常系: canceled更新 + ユーザー紐付け解除 + processed", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		subscriptionID := "sub_deleted_" + randomString(8)
		eventID := "evt_deleted_" + randomString(8)

		seeded, err := stripeSubscriberRepo.Create(ctx, query.CreateStripeSubscriberParams{
			StripeCustomerID:         "cus_deleted_" + randomString(8),
			StripeSubscriptionID:     subscriptionID,
			StripePriceID:            "price_monthly",
			StripeStatus:             string(model.StripeSubscriptionStatusActive),
			StripeCurrentPeriodStart: time.Now(),
			StripeCurrentPeriodEnd:   time.Now().AddDate(0, 1, 0),
		})
		if err != nil {
			t.Fatalf("シード用 StripeSubscriber の作成に失敗: %v", err)
		}

		// A user linked to the subscriber so we can assert the link is cleared.
		//
		// [Ja] subscriber に紐付くユーザーを作り、紐付けが解除されることを検証できるようにする。
		userID := insertStripeTestUserLinkedTo(t, db, sql.NullInt64{Int64: int64(seeded.ID), Valid: true})
		t.Cleanup(func() {
			_, _ = db.Exec("DELETE FROM stripe_webhook_events WHERE stripe_event_id = $1", eventID)
			_, _ = db.Exec("DELETE FROM users WHERE id = $1", int64(userID))
			_, _ = db.Exec("DELETE FROM stripe_subscribers WHERE id = $1", int64(seeded.ID))
		})

		event := marshalSubscriptionEvent(t, eventID, stripe.EventTypeCustomerSubscriptionDeleted, stripe.Subscription{
			ID:         subscriptionID,
			Status:     stripe.SubscriptionStatusCanceled,
			CanceledAt: time.Now().Unix(),
		})

		if _, err := uc.Execute(ctx, ProcessStripeWebhookInput{Event: event}); err != nil {
			t.Fatalf("Webhook処理で予期しないエラー: %v", err)
		}

		gotEvent, err := stripeWebhookEventRepo.GetByStripeEventID(ctx, eventID)
		if err != nil {
			t.Fatalf("Webhookイベントの取得に失敗: %v", err)
		}
		if gotEvent.Status != model.WebhookEventStatusProcessed.String() {
			t.Errorf("イベントステータス: got %s, want %s", gotEvent.Status, model.WebhookEventStatusProcessed)
		}

		// The subscriber status moves to canceled.
		//
		// [Ja] subscriber のステータスが canceled に更新される。
		canceled, err := stripeSubscriberRepo.GetByStripeSubscriptionID(ctx, subscriptionID)
		if err != nil {
			t.Fatalf("StripeSubscriber取得エラー: %v", err)
		}
		if canceled == nil {
			t.Fatal("StripeSubscriber が見つかりません")
		}
		if canceled.StripeStatus != string(model.StripeSubscriptionStatusCanceled) {
			t.Errorf("StripeStatus: got %s, want %s", canceled.StripeStatus, model.StripeSubscriptionStatusCanceled)
		}

		// The user is unlinked from the subscriber.
		//
		// [Ja] ユーザーの subscriber 紐付けが解除される。
		unlinked, err := userRepo.GetByID(ctx, userID)
		if err != nil {
			t.Fatalf("ユーザー取得エラー: %v", err)
		}
		if unlinked.StripeSubscriberID.Valid {
			t.Errorf("ユーザーの紐付けが解除されていません: stripe_subscriber_id=%d", unlinked.StripeSubscriberID.Int64)
		}
	})
}

// TestProcessStripeWebhookUsecase_MarksInvoiceAndUnhandledEvents covers the event
// types that only update the webhook event row without touching a subscriber:
// invoice.payment_succeeded / invoice.payment_failed are marked processed, and an
// unhandled event type is marked skipped.
//
// [Ja] subscriber に触れず Webhook イベント行だけを更新するイベント種別を検証する。
// invoice.payment_succeeded / invoice.payment_failed は processed、処理対象外の
// イベント種別は skipped としてマークされる。
func TestProcessStripeWebhookUsecase_MarksInvoiceAndUnhandledEvents(t *testing.T) {
	t.Parallel()

	// These branches never read or write a subscriber and do not parse the payload,
	// so an empty Data.Raw is sufficient.
	//
	// [Ja] これらの分岐は subscriber を読み書きせず、ペイロードもパースしないため、
	// Data.Raw は空で十分。
	db := testutil.GetTestDB()
	uc, _, _, stripeWebhookEventRepo := newWebhookUsecaseForTest(db)

	tests := []struct {
		name       string
		eventType  stripe.EventType
		wantStatus model.WebhookEventStatus
	}{
		{
			name:       "invoice.payment_succeeded は processed",
			eventType:  stripe.EventTypeInvoicePaymentSucceeded,
			wantStatus: model.WebhookEventStatusProcessed,
		},
		{
			name:       "invoice.payment_failed は processed (Stripe側で自動リトライ)",
			eventType:  stripe.EventTypeInvoicePaymentFailed,
			wantStatus: model.WebhookEventStatusProcessed,
		},
		{
			name:       "処理対象外イベントは skipped",
			eventType:  stripe.EventTypeCustomerCreated,
			wantStatus: model.WebhookEventStatusSkipped,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			eventID := "evt_misc_" + randomString(8)
			t.Cleanup(func() {
				_, _ = db.Exec("DELETE FROM stripe_webhook_events WHERE stripe_event_id = $1", eventID)
			})

			event := &stripe.Event{
				ID:   eventID,
				Type: tt.eventType,
				Data: &stripe.EventData{Raw: []byte("{}")},
			}

			if _, err := uc.Execute(ctx, ProcessStripeWebhookInput{Event: event}); err != nil {
				t.Fatalf("Webhook処理で予期しないエラー: %v", err)
			}

			got, err := stripeWebhookEventRepo.GetByStripeEventID(ctx, eventID)
			if err != nil {
				t.Fatalf("Webhookイベントの取得に失敗: %v", err)
			}
			if got.Status != tt.wantStatus.String() {
				t.Errorf("イベントステータス: got %s, want %s", got.Status, tt.wantStatus)
			}
		})
	}
}
