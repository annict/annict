package model

import (
	"database/sql"
	"time"
)

// SignInCode はメールでログインする際の 6 桁コードのドメインエンティティ
type SignInCode struct {
	ID         SignInCodeID
	UserID     UserID
	CodeDigest string
	Attempts   int32
	ExpiresAt  time.Time
	UsedAt     sql.NullTime
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
