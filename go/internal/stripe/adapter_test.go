package stripe

import (
	"testing"
	"time"

	stripego "github.com/stripe/stripe-go/v84"
)

// TestNewSubscriptionFromStripe verifies the conversion from a stripe-go
// subscription to the domain-shaped Subscription: status passthrough, item
// mapping (price ID and Unix-time conversion), and NullTime handling for the
// cancel timestamps.
//
// [Ja] TestNewSubscriptionFromStripe は stripe-go のサブスクリプションからドメイン形
// Subscription への変換を検証する。ステータスのパススルー、アイテムのマッピング
// (price ID と Unix 時刻変換)、キャンセル時刻の NullTime への変換を対象とする。
func TestNewSubscriptionFromStripe(t *testing.T) {
	t.Parallel()

	const (
		periodStart = int64(1_700_000_000)
		periodEnd   = int64(1_702_592_000)
		cancelAt    = int64(1_705_000_000)
		canceledAt  = int64(1_705_100_000)
	)

	t.Run("正常系: items とキャンセル情報を変換できる", func(t *testing.T) {
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
			t.Errorf("Status = %q, want %q", got.Status, "active")
		}
		if len(got.Items) != 1 {
			t.Fatalf("len(Items) = %d, want 1", len(got.Items))
		}
		item := got.Items[0]
		if item.PriceID != "price_monthly" {
			t.Errorf("Items[0].PriceID = %q, want %q", item.PriceID, "price_monthly")
		}
		if !item.CurrentPeriodStart.Equal(time.Unix(periodStart, 0)) {
			t.Errorf("Items[0].CurrentPeriodStart = %v, want %v", item.CurrentPeriodStart, time.Unix(periodStart, 0))
		}
		if !item.CurrentPeriodEnd.Equal(time.Unix(periodEnd, 0)) {
			t.Errorf("Items[0].CurrentPeriodEnd = %v, want %v", item.CurrentPeriodEnd, time.Unix(periodEnd, 0))
		}
		if !got.CancelAt.Valid || !got.CancelAt.Time.Equal(time.Unix(cancelAt, 0)) {
			t.Errorf("CancelAt = %+v, want valid %v", got.CancelAt, time.Unix(cancelAt, 0))
		}
		if !got.CanceledAt.Valid || !got.CanceledAt.Time.Equal(time.Unix(canceledAt, 0)) {
			t.Errorf("CanceledAt = %+v, want valid %v", got.CanceledAt, time.Unix(canceledAt, 0))
		}
	})

	t.Run("正常系: items が空のときは空スライスになる", func(t *testing.T) {
		t.Parallel()

		sub := &stripego.Subscription{
			Status: stripego.SubscriptionStatusActive,
			Items:  &stripego.SubscriptionItemList{Data: []*stripego.SubscriptionItem{}},
		}

		got := newSubscriptionFromStripe(sub)

		if len(got.Items) != 0 {
			t.Errorf("len(Items) = %d, want 0", len(got.Items))
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
			t.Errorf("CancelAt.Valid = true, want false")
		}
		if got.CanceledAt.Valid {
			t.Errorf("CanceledAt.Valid = true, want false")
		}
	})
}
