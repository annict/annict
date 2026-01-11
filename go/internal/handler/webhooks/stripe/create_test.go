package stripe

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
)

// generateStripeSignature はテスト用のStripe Webhook署名を生成します
func generateStripeSignature(payload []byte, secret string, timestamp int64) string {
	signedPayload := fmt.Sprintf("%d.%s", timestamp, string(payload))
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(signedPayload))
	signature := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("t=%d,v1=%s", timestamp, signature)
}

func TestCreate_SignatureValidation(t *testing.T) {
	t.Parallel()

	// テストDBをセットアップ
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(tx)

	// テスト用の設定
	webhookSecret := "whsec_test_secret_12345"
	cfg := &config.Config{
		StripeWebhookSecret: webhookSecret,
	}

	// リポジトリの作成
	stripeWebhookEventRepo := repository.NewStripeWebhookEventRepository(queries)
	stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
	userRepo := repository.NewUserRepository(queries)

	// UseCaseの作成
	createStripeSubscriberUC := usecase.NewCreateStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo)

	// ハンドラーの作成
	handler := NewHandler(cfg, stripeWebhookEventRepo, stripeSubscriberRepo, userRepo, createStripeSubscriberUC)

	tests := []struct {
		name           string
		payload        string
		signature      string
		wantStatusCode int
	}{
		{
			name: "署名がないリクエストは400を返す",
			payload: `{
				"id": "evt_test_123",
				"type": "checkout.session.completed",
				"data": {}
			}`,
			signature:      "",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "不正な署名は400を返す",
			payload: `{
				"id": "evt_test_123",
				"type": "checkout.session.completed",
				"data": {}
			}`,
			signature:      "t=12345,v1=invalid_signature",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "正しい署名は200を返す",
			payload: `{
				"id": "evt_test_valid_signature",
				"object": "event",
				"type": "checkout.session.completed",
				"data": {"object": {}},
				"livemode": false,
				"api_version": "2025-12-15.clover",
				"created": 1234567890,
				"pending_webhooks": 0,
				"request": {"id": null, "idempotency_key": null}
			}`,
			signature:      "valid", // 後で正しい署名に置き換え
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature := tt.signature

			// 正しい署名が必要な場合は生成
			if tt.signature == "valid" {
				timestamp := time.Now().Unix()
				signature = generateStripeSignature([]byte(tt.payload), webhookSecret, timestamp)
			}

			// リクエストを作成
			req := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", strings.NewReader(tt.payload))
			req.Header.Set("Content-Type", "application/json")
			if signature != "" {
				req.Header.Set("Stripe-Signature", signature)
			}

			// レスポンスを記録
			rr := httptest.NewRecorder()

			// ハンドラーを実行
			handler.Create(rr, req)

			// ステータスコードを確認
			if rr.Code != tt.wantStatusCode {
				t.Errorf("ステータスコード: got %d, want %d", rr.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestCreate_Idempotency(t *testing.T) {
	t.Parallel()

	// テストDBをセットアップ
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(tx)

	// テスト用の設定
	webhookSecret := "whsec_test_secret_12345"
	cfg := &config.Config{
		StripeWebhookSecret: webhookSecret,
	}

	// リポジトリの作成
	stripeWebhookEventRepo := repository.NewStripeWebhookEventRepository(queries)
	stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
	userRepo := repository.NewUserRepository(queries)

	// UseCaseの作成
	createStripeSubscriberUC := usecase.NewCreateStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo)

	// ハンドラーの作成
	handler := NewHandler(cfg, stripeWebhookEventRepo, stripeSubscriberRepo, userRepo, createStripeSubscriberUC)

	// 既にイベントを登録
	existingEventID := "evt_existing_event_123"
	testutil.NewStripeWebhookEventBuilder(t, tx).
		WithStripeEventID(existingEventID).
		WithStripeEventType("checkout.session.completed").
		WithStatus("processed").
		Build()

	// 同じイベントIDで再度リクエストを送信
	payload := fmt.Sprintf(`{
		"id": "%s",
		"object": "event",
		"type": "checkout.session.completed",
		"data": {"object": {}},
		"livemode": false,
		"api_version": "2025-12-15.clover",
		"created": 1234567890,
		"pending_webhooks": 0,
		"request": {"id": null, "idempotency_key": null}
	}`, existingEventID)

	timestamp := time.Now().Unix()
	signature := generateStripeSignature([]byte(payload), webhookSecret, timestamp)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Stripe-Signature", signature)

	rr := httptest.NewRecorder()
	handler.Create(rr, req)

	// 200を返す（冪等性による重複スキップ）
	if rr.Code != http.StatusOK {
		t.Errorf("冪等性チェック後のステータスコード: got %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestCreate_EventProcessing(t *testing.T) {
	t.Parallel()

	// テストDBをセットアップ
	db, tx := testutil.SetupTestDB(t)
	queries := query.New(tx)

	// テスト用の設定
	webhookSecret := "whsec_test_secret_12345"
	cfg := &config.Config{
		StripeWebhookSecret: webhookSecret,
	}

	// リポジトリの作成
	stripeWebhookEventRepo := repository.NewStripeWebhookEventRepository(queries)
	stripeSubscriberRepo := repository.NewStripeSubscriberRepository(queries)
	userRepo := repository.NewUserRepository(queries)

	// UseCaseの作成
	createStripeSubscriberUC := usecase.NewCreateStripeSubscriberUsecase(db, stripeSubscriberRepo, userRepo)

	// ハンドラーの作成
	handler := NewHandler(cfg, stripeWebhookEventRepo, stripeSubscriberRepo, userRepo, createStripeSubscriberUC)

	tests := []struct {
		name           string
		eventType      string
		wantStatusCode int
	}{
		{
			name:           "checkout.session.completedイベントを処理",
			eventType:      "checkout.session.completed",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "customer.subscription.updatedイベントを処理",
			eventType:      "customer.subscription.updated",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "customer.subscription.deletedイベントを処理",
			eventType:      "customer.subscription.deleted",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "invoice.payment_succeededイベントを処理",
			eventType:      "invoice.payment_succeeded",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "invoice.payment_failedイベントを処理",
			eventType:      "invoice.payment_failed",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "処理対象外のイベントはスキップ",
			eventType:      "product.created",
			wantStatusCode: http.StatusOK,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 一意のイベントIDを生成
			eventID := fmt.Sprintf("evt_test_%s_%d", tt.eventType, i)

			payload := fmt.Sprintf(`{
				"id": "%s",
				"object": "event",
				"type": "%s",
				"data": {"object": {}},
				"livemode": false,
				"api_version": "2025-12-15.clover",
				"created": 1234567890,
				"pending_webhooks": 0,
				"request": {"id": null, "idempotency_key": null}
			}`, eventID, tt.eventType)

			timestamp := time.Now().Unix()
			signature := generateStripeSignature([]byte(payload), webhookSecret, timestamp)

			req := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Stripe-Signature", signature)

			rr := httptest.NewRecorder()
			handler.Create(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Errorf("ステータスコード: got %d, want %d", rr.Code, tt.wantStatusCode)
			}

			// イベントがDBに保存されていることを確認
			savedEvent, err := stripeWebhookEventRepo.GetByStripeEventID(req.Context(), eventID)
			if err != nil {
				t.Errorf("イベントがDBに保存されていません: %v", err)
			}
			if savedEvent.StripeEventType != tt.eventType {
				t.Errorf("イベントタイプ: got %s, want %s", savedEvent.StripeEventType, tt.eventType)
			}
		})
	}
}
