package model

import "time"

// EmailNotification はユーザーのメール通知設定のドメインエンティティ
type EmailNotification struct {
	ID                      EmailNotificationID
	UserID                  UserID
	UnsubscriptionKey       string
	EventFollowedUser       bool
	EventLikedEpisodeRecord bool
	CreatedAt               time.Time
	UpdatedAt               time.Time
}
