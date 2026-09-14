package model

import (
	"database/sql"
	"time"
)

// PasswordResetTokenはパスワードリセットトークンのドメインエンティティ
type PasswordResetToken struct {
	ID          PasswordResetTokenID
	UserID      UserID
	TokenDigest string
	ExpiresAt   time.Time
	UsedAt      sql.NullTime
	CreatedAt   time.Time
}
