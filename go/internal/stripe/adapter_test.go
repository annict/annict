package stripe

import (
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
