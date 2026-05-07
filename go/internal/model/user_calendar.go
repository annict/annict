// Package model はドメインモデルを定義します
package model

import "time"

// UserCalendar はユーザーのカレンダーデータを表します
type UserCalendar struct {
	Username string
	TimeZone string
	Locale   string
	Slots    []CalendarSlot
	Works    []CalendarWork
}

// CalendarSlot はカレンダーに表示する放送枠を表します
type CalendarSlot struct {
	ID            SlotID
	StartedAt     time.Time
	WorkID        WorkID
	WorkTitle     string
	WorkTitleEn   string
	EpisodeID     EpisodeID
	EpisodeTitle  string
	EpisodeNumber string
	ChannelName   string
}

// CalendarWork はカレンダーに表示する作品（放送開始日）を表します
type CalendarWork struct {
	ID        WorkID
	Title     string
	TitleEn   string
	StartedOn time.Time
}
