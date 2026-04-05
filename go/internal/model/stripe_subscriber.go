package model

import (
	"database/sql"
	"time"
)

// StripeSubscriber はStripeサブスクライバーのドメインエンティティ
type StripeSubscriber struct {
	ID                       int64
	StripeCustomerID         string
	StripeSubscriptionID     string
	StripePriceID            string
	StripeStatus             string
	StripeCurrentPeriodStart time.Time
	StripeCurrentPeriodEnd   time.Time
	StripeCancelAt           sql.NullTime
	StripeCanceledAt         sql.NullTime
	CreatedAt                time.Time
	UpdatedAt                time.Time
}
