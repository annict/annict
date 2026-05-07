package model

import (
	"database/sql"
	"time"
)

// PasswordResetToken はパスワードリセットトークンのドメインエンティティ
type PasswordResetToken struct {
	ID          PasswordResetTokenID
	UserID      UserID
	TokenDigest string
	ExpiresAt   time.Time
	UsedAt      sql.NullTime
	CreatedAt   time.Time
}
