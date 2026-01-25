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
	ID            int64
	StartedAt     time.Time
	WorkID        int64
	WorkTitle     string
	WorkTitleEn   string
	EpisodeID     int64
	EpisodeTitle  string
	EpisodeNumber string
	ChannelName   string
}

// CalendarWork はカレンダーに表示する作品（放送開始日）を表します
type CalendarWork struct {
	ID        int64
	Title     string
	TitleEn   string
	StartedOn time.Time
}
