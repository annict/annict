package stripe

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	stripego "github.com/stripe/stripe-go/v84"
)

// TestNewSubscriptionFromStripeはstripe-goのサブスクリプションからドメイン形
// Subscriptionへの変換を検証する。ステータスのパススルー、アイテムのマッピング
// (price IDとUnix時刻変換)、キャンセル時刻のNullTimeへの変換を対象とする。
func TestNewSubscriptionFromStripe(t *testing.T) {
	t.Parallel()

	const (
		periodStart = int64(1_700_000_000)
		periodEnd   = int64(1_702_592_000)
		cancelAt    = int64(1_705_000_000)
		canceledAt  = int64(1_705_100_000)
	)

	t.Run("正常系: itemsとキャンセル情報を変換できる", func(t *testing.T) {
		t.Parallel()

		sub := &stripego.Subscription{
			Status: stripego.SubscriptionStatusActive,
			Items: &stripego.SubscriptionItemList{
				Data: []*stripego.SubscriptionItem{
					{
						Price:              &stripego.Price{ID: "price_monthly"},
						CurrentPeriodStart: periodStart,
						CurrentPeriodEnd:   periodEnd,
					},
				},
			},
			CancelAt:   cancelAt,
			CanceledAt: canceledAt,
		}

		got := newSubscriptionFromStripe(sub)

		if got.Status != "active" {
			t.Errorf("Status = %q、期待値 = %q", got.Status, "active")
		}
		if len(got.Items) != 1 {
			t.Fatalf("len(Items) = %d、期待値 = 1", len(got.Items))
		}
		item := got.Items[0]
		if item.PriceID != "price_monthly" {
			t.Errorf("Items[0].PriceID = %q、期待値 = %q", item.PriceID, "price_monthly")
		}
		if !item.CurrentPeriodStart.Equal(time.Unix(periodStart, 0)) {
			t.Errorf("Items[0].CurrentPeriodStart = %v、期待値 = %v", item.CurrentPeriodStart, time.Unix(periodStart, 0))
		}
		if !item.CurrentPeriodEnd.Equal(time.Unix(periodEnd, 0)) {
			t.Errorf("Items[0].CurrentPeriodEnd = %v、期待値 = %v", item.CurrentPeriodEnd, time.Unix(periodEnd, 0))
		}
		if !got.CancelAt.Valid || !got.CancelAt.Time.Equal(time.Unix(cancelAt, 0)) {
			t.Errorf("CancelAt = %+v、期待値 = Validがtrueで%v", got.CancelAt, time.Unix(cancelAt, 0))
		}
		if !got.CanceledAt.Valid || !got.CanceledAt.Time.Equal(time.Unix(canceledAt, 0)) {
			t.Errorf("CanceledAt = %+v、期待値 = Validがtrueで%v", got.CanceledAt, time.Unix(canceledAt, 0))
		}
	})

	t.Run("正常系: itemsが空のときは空スライスになる", func(t *testing.T) {
		t.Parallel()

		sub := &stripego.Subscription{
			Status: stripego.SubscriptionStatusActive,
			Items:  &stripego.SubscriptionItemList{Data: []*stripego.SubscriptionItem{}},
		}

		got := newSubscriptionFromStripe(sub)

		if len(got.Items) != 0 {
			t.Errorf("len(Items) = %d、期待値 = 0", len(got.Items))
		}
	})

	t.Run("正常系: キャンセル時刻が0のときNullTimeはinvalidになる", func(t *testing.T) {
		t.Parallel()

		sub := &stripego.Subscription{
			Status: stripego.SubscriptionStatusCanceled,
			Items:  &stripego.SubscriptionItemList{Data: []*stripego.SubscriptionItem{}},
		}

		got := newSubscriptionFromStripe(sub)

		if got.CancelAt.Valid {
			t.Errorf("CancelAt.Valid = true、期待値 = false")
		}
		if got.CanceledAt.Valid {
			t.Errorf("CanceledAt.Valid = true、期待値 = false")
		}
	})
}

// newTestAdapterはStripe APIの代わりにhandlerへリクエストを送るAdapterを作成する。
func newTestAdapter(t *testing.T, handler http.HandlerFunc) *Adapter {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	backends := stripego.NewBackendsWithConfig(&stripego.BackendConfig{
		URL:               stripego.String(server.URL),
		HTTPClient:        server.Client(),
		MaxNetworkRetries: stripego.Int64(0),
		LeveledLogger:     &stripego.LeveledLogger{Level: stripego.LevelNull},
	})
	return NewAdapter(stripego.NewClient("sk_test_dummy", stripego.WithBackends(backends)))
}

// writeStripeSubscriptionは指定したステータスのサブスクリプションをレスポンスとして書き込む。
func writeStripeSubscription(w http.ResponseWriter, status string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprintf(w, `{"id":"sub_test","object":"subscription","status":%q,"items":{"object":"list","data":[]}}`, status)
}

// writeStripeErrorはStripe APIのエラーをレスポンスとして書き込む。
func writeStripeError(w http.ResponseWriter, statusCode int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = fmt.Fprintf(w, `{"error":{"type":"invalid_request_error","code":%q,"message":"テスト用のエラー"}}`, code)
}

func TestAdapter_CancelSubscription(t *testing.T) {
	t.Parallel()

	const path = "/v1/subscriptions/sub_test"

	t.Run("正常系: 有効なサブスクリプションを日割り返金・即時請求なしで解約する", func(t *testing.T) {
		t.Parallel()

		var canceled atomic.Bool
		adapter := newTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != path {
				t.Errorf("リクエストのパス = %q、期待値 = %q", r.URL.Path, path)
			}
			switch r.Method {
			case http.MethodGet:
				writeStripeSubscription(w, "active")
			case http.MethodDelete:
				canceled.Store(true)
				if err := r.ParseForm(); err != nil {
					t.Errorf("ParseForm()のエラー = %v", err)
				}
				if got := r.Form.Get("prorate"); got != "false" {
					t.Errorf("prorate = %q、期待値 = %q", got, "false")
				}
				if got := r.Form.Get("invoice_now"); got != "false" {
					t.Errorf("invoice_now = %q、期待値 = %q", got, "false")
				}
				writeStripeSubscription(w, "canceled")
			default:
				t.Errorf("想定外のメソッド = %s", r.Method)
			}
		})

		got, err := adapter.CancelSubscription(context.Background(), "sub_test")
		if err != nil {
			t.Fatalf("CancelSubscription()のエラー = %v、期待値 = nil", err)
		}
		if got != CancelSubscriptionResultCanceled {
			t.Errorf("CancelSubscription()の結果 = %v、期待値 = CancelSubscriptionResultCanceled", got)
		}
		if !canceled.Load() {
			t.Error("解約のリクエストが送られていない")
		}
	})

	for _, status := range []string{"canceled", "incomplete_expired"} {
		t.Run(fmt.Sprintf("正常系: %sのサブスクリプションには解約を送らない", status), func(t *testing.T) {
			t.Parallel()

			adapter := newTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("想定外のメソッド = %s、期待値 = GETのみ", r.Method)
				}
				writeStripeSubscription(w, status)
			})

			got, err := adapter.CancelSubscription(context.Background(), "sub_test")
			if err != nil {
				t.Fatalf("CancelSubscription()のエラー = %v、期待値 = nil", err)
			}
			if got != CancelSubscriptionResultAlreadyEnded {
				t.Errorf("CancelSubscription()の結果 = %v、期待値 = CancelSubscriptionResultAlreadyEnded", got)
			}
		})
	}

	t.Run("正常系: 取得時に存在しなければNotFoundを返す", func(t *testing.T) {
		t.Parallel()

		adapter := newTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("想定外のメソッド = %s、期待値 = GETのみ", r.Method)
			}
			writeStripeError(w, http.StatusNotFound, "resource_missing")
		})

		got, err := adapter.CancelSubscription(context.Background(), "sub_test")
		if err != nil {
			t.Fatalf("CancelSubscription()のエラー = %v、期待値 = nil", err)
		}
		if got != CancelSubscriptionResultNotFound {
			t.Errorf("CancelSubscription()の結果 = %v、期待値 = CancelSubscriptionResultNotFound", got)
		}
	})

	t.Run("正常系: 解約時に存在しなければNotFoundを返す", func(t *testing.T) {
		t.Parallel()

		adapter := newTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				writeStripeSubscription(w, "active")
			default:
				writeStripeError(w, http.StatusNotFound, "resource_missing")
			}
		})

		got, err := adapter.CancelSubscription(context.Background(), "sub_test")
		if err != nil {
			t.Fatalf("CancelSubscription()のエラー = %v、期待値 = nil", err)
		}
		if got != CancelSubscriptionResultNotFound {
			t.Errorf("CancelSubscription()の結果 = %v、期待値 = CancelSubscriptionResultNotFound", got)
		}
	})

	t.Run("異常系: 取得に失敗したらエラーを返す", func(t *testing.T) {
		t.Parallel()

		adapter := newTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("想定外のメソッド = %s、期待値 = GETのみ", r.Method)
			}
			writeStripeError(w, http.StatusInternalServerError, "")
		})

		if _, err := adapter.CancelSubscription(context.Background(), "sub_test"); err == nil {
			t.Error("CancelSubscription()のエラー = nil、期待値 = エラーあり")
		}
	})

	t.Run("異常系: 解約に失敗したらエラーを返す", func(t *testing.T) {
		t.Parallel()

		adapter := newTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				writeStripeSubscription(w, "active")
			default:
				writeStripeError(w, http.StatusInternalServerError, "")
			}
		})

		if _, err := adapter.CancelSubscription(context.Background(), "sub_test"); err == nil {
			t.Error("CancelSubscription()のエラー = nil、期待値 = エラーあり")
		}
	})
}
