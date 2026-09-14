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

// 与えられたsubscriptionをJSON化してData.Rawに持つstripe.Eventを組み立てる。
// Stripeがcustomer.subscription.* イベントを配信する形を模している。
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

// 対応するStripeSubscriberが存在しないupdate / deletedイベントがfailedでは
// なくskippedとして記録されることを検証する確定バグの回帰テスト。以前はWebhook層が
// ラップ済みエラーを == sql.ErrNoRowsで比較していたため、未存在分岐が到達不能でfailed
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

			// これらのケースではcreateユースケースは実行されないため、Stripe
			// クライアントはnilで十分。
			createUC := NewCreateStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo, nil)
			updateUC := NewUpdateStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo)
			deleteUC := NewDeleteStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo)
			uc := NewProcessStripeWebhookUsecase(stripeWebhookEventRepo, createUC, updateUC, deleteUC)

			ctx := context.Background()

			// 対応するStripeSubscriberが存在しないサブスクリプションID。
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
				t.Errorf("イベントステータス = %s、期待値 = %s", got.Status, model.WebhookEventStatusSkipped)
			}
		})
	}
}

// 与えられたsessionをJSON化してData.Rawに持つcheckout.session.completedの
// stripe.Eventを組み立てる。Stripeがイベントを配信する形を模している。
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

// checkout.session.completed経路を検証する。正常系ではStripeSubscriberを作成し
// ユーザーを紐付ける。不正なセッションはskipped (subscription / customer欠落) または
// failed (metadataのuser_id欠落 / 不正) になる。customer欠落ケースは
// session.Customer.ID参照でのpanicに対する回帰ガードである。
func TestProcessStripeWebhookUsecase_HandleCheckoutSessionCompleted(t *testing.T) {
	t.Parallel()

	// handleCheckoutSessionCompletedは内部でトランザクションを開く
	// CreateStripeSubscriberUsecaseを呼ぶため、内側のトランザクションから見えない
	// 外側のロールバック用トランザクションではなく、共有DBを直接使ってテストデータを
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
			t.Errorf("イベントステータス = %s、期待値 = %s", gotEvent.Status, model.WebhookEventStatusProcessed)
		}

		// StripeSubscriberが作成され、セッションのcustomer IDがレコードへ
		// 引き継がれることを確認する。
		subscriber, err := stripeSubscriberRepo.GetByStripeSubscriptionID(ctx, subscriptionID)
		if err != nil {
			t.Fatalf("StripeSubscriber取得エラー: %v", err)
		}
		if subscriber == nil {
			t.Fatal("StripeSubscriberが作成されていません")
		}
		if subscriber.StripeCustomerID != customerID {
			t.Errorf("StripeCustomerID = %s、期待値 = %s", subscriber.StripeCustomerID, customerID)
		}

		linked, err := userRepo.GetByID(ctx, userID)
		if err != nil {
			t.Fatalf("ユーザー取得エラー: %v", err)
		}
		if !linked.StripeSubscriberID.Valid {
			t.Fatal("ユーザーにStripeSubscriberが紐付けられていません")
		}
		if linked.StripeSubscriberID.Int64 != int64(subscriber.ID) {
			t.Errorf("紐付けられたStripeSubscriberID = %d、期待値 = %d", linked.StripeSubscriberID.Int64, int64(subscriber.ID))
		}
	})

	// これらのケースはCreateStripeSubscriberUsecase到達前に短絡するため、
	// subscriberも作らずユーザーにも触れない。書き込まれて検証されるのはWebhook
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

			// これらのケースはcreateユースケースに到達しないため、Stripe
			// クライアントはnilで十分。
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
				t.Errorf("イベントステータス = %s、期待値 = %s", got.Status, tc.wantStatus)
			}
		})
	}
}

// テスト用の共有DBに対してProcessStripeWebhookUsecaseを組み立て、仕込みと
// 検証に使うRepositoryと一緒に返す。本テストが対象とするsubscriptionの更新 / 削除
// / invoice / 対象外イベントの経路はいずれもCreateStripeSubscriberUsecaseに到達しない
// ため、SubscriptionRetrieverはnilで十分。
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

// customer.subscription.updated経路を検証する。items付きイベントは既存
// subscriberを更新してprocessedになり、itemsを持たないイベントはupdateユースケース
// 到達前にskippedとなり、ステータスが無効なイベントはupdateユースケースに弾かれて
// failedとして記録される (MarkAsFailedに至る処理失敗経路)。未存在subscriberのケースは
// TestProcessStripeWebhookUsecase_SkipsWhenSubscriberNotFoundが担当する。
func TestProcessStripeWebhookUsecase_HandleCustomerSubscriptionUpdated(t *testing.T) {
	t.Parallel()

	// UpdateStripeSubscriberUsecaseは内部トランザクションを開かずsubscriberを
	// 直接読み書きし、処理失敗ケースは参照前にイベントを弾く。どちらもコミット済みの
	// データに依存するため、GetTestDBを使い追加した行をクリーンアップする。
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
			t.Fatalf("シード用StripeSubscriberの作成に失敗: %v", err)
		}
		t.Cleanup(func() {
			_, _ = db.Exec("DELETE FROM stripe_webhook_events WHERE stripe_event_id = $1", eventID)
			_, _ = db.Exec("DELETE FROM stripe_subscribers WHERE id = $1", int64(seeded.ID))
		})

		// updatedイベントは新しいpriceを載せる。永続化レコードがそれを反映する必要がある。
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
			t.Errorf("イベントステータス = %s、期待値 = %s", gotEvent.Status, model.WebhookEventStatusProcessed)
		}

		updated, err := stripeSubscriberRepo.GetByStripeSubscriptionID(ctx, subscriptionID)
		if err != nil {
			t.Fatalf("StripeSubscriber取得エラー: %v", err)
		}
		if updated == nil {
			t.Fatal("StripeSubscriberが見つかりません")
		}
		if updated.StripePriceID != "price_yearly" {
			t.Errorf("StripePriceID = %s、期待値 = price_yearly", updated.StripePriceID)
		}
	})

	t.Run("処理失敗: 無効なステータスはfailed (MarkAsFailed)", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		eventID := "evt_updated_fail_" + randomString(8)
		t.Cleanup(func() {
			_, _ = db.Exec("DELETE FROM stripe_webhook_events WHERE stripe_event_id = $1", eventID)
		})

		// 無効なステータスはUpdateStripeSubscriberUsecaseがリポジトリ参照前に弾く
		// ため、subscriberは仕込まない。このエラーはErrStripeSubscriberNotFoundでは
		// ないため、Webhookは (skippedではなく) failedとして記録する。
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
			t.Errorf("イベントステータス = %s、期待値 = %s", got.Status, model.WebhookEventStatusFailed)
		}
	})

	t.Run("items空はskipped", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		eventID := "evt_updated_noitems_" + randomString(8)
		t.Cleanup(func() {
			_, _ = db.Exec("DELETE FROM stripe_webhook_events WHERE stripe_event_id = $1", eventID)
		})

		// itemsを持たないupdatedイベントはhandleCustomerSubscriptionUpdatedが
		// updateユースケース (およびリポジトリ参照) 到達前にスキップするため、subscriberは
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
			t.Errorf("イベントステータス = %s、期待値 = %s", got.Status, model.WebhookEventStatusSkipped)
		}
	})
}

// customer.subscription.deleted経路を検証する。既存subscriberはcanceledに
// 更新され、紐付くユーザーは紐付け解除され、イベントはprocessedになる。未存在
// subscriberのケースはTestProcessStripeWebhookUsecase_SkipsWhenSubscriberNotFound
// が担当する。
func TestProcessStripeWebhookUsecase_HandleCustomerSubscriptionDeleted(t *testing.T) {
	t.Parallel()

	// DeleteStripeSubscriberUsecaseは自前のトランザクションを開くため、前提データは
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
			t.Fatalf("シード用StripeSubscriberの作成に失敗: %v", err)
		}

		// subscriberに紐付くユーザーを作り、紐付けが解除されることを検証できるようにする。
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
			t.Errorf("イベントステータス = %s、期待値 = %s", gotEvent.Status, model.WebhookEventStatusProcessed)
		}

		// subscriberのステータスがcanceledに更新される。
		canceled, err := stripeSubscriberRepo.GetByStripeSubscriptionID(ctx, subscriptionID)
		if err != nil {
			t.Fatalf("StripeSubscriber取得エラー: %v", err)
		}
		if canceled == nil {
			t.Fatal("StripeSubscriberが見つかりません")
		}
		if canceled.StripeStatus != string(model.StripeSubscriptionStatusCanceled) {
			t.Errorf("StripeStatus = %s、期待値 = %s", canceled.StripeStatus, model.StripeSubscriptionStatusCanceled)
		}

		// ユーザーのsubscriber紐付けが解除される。
		unlinked, err := userRepo.GetByID(ctx, userID)
		if err != nil {
			t.Fatalf("ユーザー取得エラー: %v", err)
		}
		if unlinked.StripeSubscriberID.Valid {
			t.Errorf("ユーザーの紐付けが解除されていません: stripe_subscriber_id=%d", unlinked.StripeSubscriberID.Int64)
		}
	})
}

// subscriberに触れずWebhookイベント行だけを更新するイベント種別を検証する。
// invoice.payment_succeeded / invoice.payment_failedはprocessed、処理対象外の
// イベント種別はskippedとしてマークされる。
func TestProcessStripeWebhookUsecase_MarksInvoiceAndUnhandledEvents(t *testing.T) {
	t.Parallel()

	// これらの分岐はsubscriberを読み書きせず、ペイロードもパースしないため、
	// Data.Rawは空で十分。
	db := testutil.GetTestDB()
	uc, _, _, stripeWebhookEventRepo := newWebhookUsecaseForTest(db)

	tests := []struct {
		name       string
		eventType  stripe.EventType
		wantStatus model.WebhookEventStatus
	}{
		{
			name:       "invoice.payment_succeededはprocessed",
			eventType:  stripe.EventTypeInvoicePaymentSucceeded,
			wantStatus: model.WebhookEventStatusProcessed,
		},
		{
			name:       "invoice.payment_failedはprocessed (Stripe側で自動リトライ)",
			eventType:  stripe.EventTypeInvoicePaymentFailed,
			wantStatus: model.WebhookEventStatusProcessed,
		},
		{
			name:       "処理対象外イベントはskipped",
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
				t.Errorf("イベントステータス = %s、期待値 = %s", got.Status, tt.wantStatus)
			}
		})
	}
}
