package model

// StripeSubscriptionStatus はStripeサブスクリプションの状態を表す型です
// https://docs.stripe.com/api/subscriptions/object#subscription_object-status
type StripeSubscriptionStatus string

const (
	// StripeSubscriptionStatusActive は通常のアクティブ状態
	StripeSubscriptionStatusActive StripeSubscriptionStatus = "active"
	// StripeSubscriptionStatusPastDue は支払い遅延中の状態（リトライ中）
	StripeSubscriptionStatusPastDue StripeSubscriptionStatus = "past_due"
	// StripeSubscriptionStatusUnpaid は未払い状態
	StripeSubscriptionStatusUnpaid StripeSubscriptionStatus = "unpaid"
	// StripeSubscriptionStatusCanceled はキャンセル済みの状態
	StripeSubscriptionStatusCanceled StripeSubscriptionStatus = "canceled"
	// StripeSubscriptionStatusIncomplete は初回支払い未完了の状態
	StripeSubscriptionStatusIncomplete StripeSubscriptionStatus = "incomplete"
	// StripeSubscriptionStatusIncompleteExpired は初回支払い期限切れの状態
	StripeSubscriptionStatusIncompleteExpired StripeSubscriptionStatus = "incomplete_expired"
	// StripeSubscriptionStatusTrialing はトライアル期間中の状態
	StripeSubscriptionStatusTrialing StripeSubscriptionStatus = "trialing"
	// StripeSubscriptionStatusPaused は一時停止中の状態
	StripeSubscriptionStatusPaused StripeSubscriptionStatus = "paused"
)

// String はStripeSubscriptionStatusの文字列表現を返します
func (s StripeSubscriptionStatus) String() string {
	return string(s)
}

// IsValid はステータスが有効な値かどうかを判定します
func (s StripeSubscriptionStatus) IsValid() bool {
	switch s {
	case StripeSubscriptionStatusActive,
		StripeSubscriptionStatusPastDue,
		StripeSubscriptionStatusUnpaid,
		StripeSubscriptionStatusCanceled,
		StripeSubscriptionStatusIncomplete,
		StripeSubscriptionStatusIncompleteExpired,
		StripeSubscriptionStatusTrialing,
		StripeSubscriptionStatusPaused:
		return true
	}
	return false
}

// IsActive はサブスクリプションがアクティブかどうかを判定します
// active または past_due 状態をアクティブとして扱います
// past_due は支払い遅延中だが、Stripeがリトライ中のため猶予期間として利用可能
func (s StripeSubscriptionStatus) IsActive() bool {
	return s == StripeSubscriptionStatusActive || s == StripeSubscriptionStatusPastDue
}
