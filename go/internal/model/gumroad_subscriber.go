package model

import (
	"database/sql"
	"time"
)

// GumroadSubscriber はGumroadサブスクライバーのドメインエンティティ
type GumroadSubscriber struct {
	ID                                 GumroadSubscriberID
	GumroadID                          string
	GumroadProductID                   string
	GumroadProductName                 string
	GumroadUserID                      sql.NullString
	GumroadUserEmail                   sql.NullString
	GumroadPurchaseIds                 []string
	GumroadCreatedAt                   time.Time
	GumroadCancelledAt                 sql.NullTime
	GumroadUserRequestedCancellationAt sql.NullTime
	GumroadChargeOccurrenceCount       sql.NullTime
	GumroadEndedAt                     sql.NullTime
	CreatedAt                          time.Time
	UpdatedAt                          time.Time
}
