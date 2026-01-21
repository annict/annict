package viewmodel

import "time"

// SupporterStatus はサポーターの状態を表します
type SupporterStatus int

const (
	// SupporterStatusNone は非サポーター
	SupporterStatusNone SupporterStatus = iota
	// SupporterStatusGumroad はGumroadサポーター（アクティブ）
	SupporterStatusGumroad
	// SupporterStatusStripe はStripeサポーター（アクティブ）
	SupporterStatusStripe
	// SupporterStatusBoth はGumroadとStripe両方アクティブ
	SupporterStatusBoth
)

// StripeSubscriberView はStripeサブスクライバーのビューモデルです
type StripeSubscriberView struct {
	CustomerID       string
	Status           string
	CurrentPeriodEnd time.Time
	CancelAt         *time.Time
}

// GumroadSubscriberView はGumroadサブスクライバーのビューモデルです
type GumroadSubscriberView struct {
	GumroadID string
	CreatedAt time.Time
	EndedAt   *time.Time
}

// SupporterPageData はサポーターページのビューモデルです
type SupporterPageData struct {
	IsLoggedIn          bool
	Status              SupporterStatus
	StripeSubscriber    *StripeSubscriberView
	GumroadSubscriber   *GumroadSubscriberView
	ShowSuccessMessage  bool
	ShowCanceledMessage bool
}
