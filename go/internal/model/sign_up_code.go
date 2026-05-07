package model

import (
	"database/sql"
	"time"
)

// SignUpCode は新規登録メール確認の 6 桁コードのドメインエンティティ
type SignUpCode struct {
	ID         SignUpCodeID
	Email      string
	CodeDigest string
	Attempts   int32
	UsedAt     sql.NullTime
	ExpiresAt  time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
