package model

// StripeSubscriptionStatusはStripeサブスクリプションの状態を表す型です
// https://docs.stripe.com/api/subscriptions/object#subscription_object-status
type StripeSubscriptionStatus string

const (
	// StripeSubscriptionStatusActiveは通常のアクティブ状態
	StripeSubscriptionStatusActive StripeSubscriptionStatus = "active"
	// StripeSubscriptionStatusPastDueは支払い遅延中の状態 (リトライ中)
	StripeSubscriptionStatusPastDue StripeSubscriptionStatus = "past_due"
	// StripeSubscriptionStatusUnpaidは未払い状態
	StripeSubscriptionStatusUnpaid StripeSubscriptionStatus = "unpaid"
	// StripeSubscriptionStatusCanceledはキャンセル済みの状態
	StripeSubscriptionStatusCanceled StripeSubscriptionStatus = "canceled"
	// StripeSubscriptionStatusIncompleteは初回支払い未完了の状態
	StripeSubscriptionStatusIncomplete StripeSubscriptionStatus = "incomplete"
	// StripeSubscriptionStatusIncompleteExpiredは初回支払い期限切れの状態
	StripeSubscriptionStatusIncompleteExpired StripeSubscriptionStatus = "incomplete_expired"
	// StripeSubscriptionStatusTrialingはトライアル期間中の状態
	StripeSubscriptionStatusTrialing StripeSubscriptionStatus = "trialing"
	// StripeSubscriptionStatusPausedは一時停止中の状態
	StripeSubscriptionStatusPaused StripeSubscriptionStatus = "paused"
)

// StringはStripeSubscriptionStatusの文字列表現を返します
func (s StripeSubscriptionStatus) String() string {
	return string(s)
}

// IsValidはステータスが有効な値かどうかを判定します
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

// IsActiveはサブスクリプションがアクティブかどうかを判定します
// activeまたはpast_due状態をアクティブとして扱います
// past_dueは支払い遅延中だが、Stripeがリトライ中のため猶予期間として利用可能
func (s StripeSubscriptionStatus) IsActive() bool {
	return s == StripeSubscriptionStatusActive || s == StripeSubscriptionStatusPastDue
}
