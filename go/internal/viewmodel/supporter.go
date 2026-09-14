package viewmodel

import "time"

// SupporterStatusはサポーターの状態を表します
type SupporterStatus int

const (
	// SupporterStatusNoneは非サポーター
	SupporterStatusNone SupporterStatus = iota
	// SupporterStatusGumroadはGumroadサポーター (アクティブ)
	SupporterStatusGumroad
	// SupporterStatusStripeはStripeサポーター (アクティブ)
	SupporterStatusStripe
	// SupporterStatusBothはGumroadとStripe両方アクティブ
	SupporterStatusBoth
)

// StripeSubscriberViewはStripeサブスクライバーのビューモデルです
type StripeSubscriberView struct {
	CustomerID       string
	Status           string
	CurrentPeriodEnd time.Time
	CancelAt         *time.Time
}

// GumroadSubscriberViewはGumroadサブスクライバーのビューモデルです
type GumroadSubscriberView struct {
	GumroadID   string
	CreatedAt   time.Time
	CancelledAt *time.Time // 契約終了予定日 (gumroad_cancelled_at)
	EndedAt     *time.Time // 実際の終了日 (gumroad_ended_at)
}

// SupporterPageDataはサポーターページのビューモデルです
type SupporterPageData struct {
	IsLoggedIn          bool
	Status              SupporterStatus
	StripeSubscriber    *StripeSubscriberView
	GumroadSubscriber   *GumroadSubscriberView
	ShowSuccessMessage  bool
	ShowCanceledMessage bool
	CSRFToken           string
	Location            *time.Location
}
