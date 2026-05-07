package model

import (
	"encoding/json"
	"time"
)

// Session はセッションのドメインエンティティ
type Session struct {
	ID        int64
	SessionID string
	Data      json.RawMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}
