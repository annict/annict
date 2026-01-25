// Package ical はiCalendar形式（RFC 5545）のカレンダーデータを生成する機能を提供します
package ical

import (
	"fmt"
	"strings"
	"time"
)

// Calendar はiCalendar形式のカレンダーを表します
type Calendar struct {
	TimeZone string
	CalName  string
	Events   []Event
}

// Event はカレンダーイベントを表します
type Event struct {
	UID         string
	Summary     string
	Description string
	Start       time.Time
	End         time.Time
	AllDay      bool // trueの場合はDate形式（終日イベント）、falseの場合はDateTime形式
}

// ToICS はiCalendar形式の文字列を生成します（RFC 5545準拠）
func (c *Calendar) ToICS() string {
	var b strings.Builder

	// カレンダーヘッダー
	b.WriteString("BEGIN:VCALENDAR\r\n")
	b.WriteString("VERSION:2.0\r\n")
	b.WriteString("PRODID:-//Annict//Annict Calendar//EN\r\n")
	b.WriteString("CALSCALE:GREGORIAN\r\n")
	b.WriteString("METHOD:PUBLISH\r\n")
	b.WriteString(fmt.Sprintf("X-WR-TIMEZONE:%s\r\n", c.TimeZone))
	b.WriteString(fmt.Sprintf("X-WR-CALNAME:%s\r\n", c.CalName))

	// タイムゾーン情報
	b.WriteString(c.generateVTimezone())

	// イベント
	for _, event := range c.Events {
		b.WriteString(c.generateVEvent(event))
	}

	b.WriteString("END:VCALENDAR\r\n")

	return b.String()
}

// generateVTimezone はVTIMEZONEコンポーネントを生成します
func (c *Calendar) generateVTimezone() string {
	loc, err := time.LoadLocation(c.TimeZone)
	if err != nil {
		// タイムゾーンが無効な場合はUTCとして処理
		loc = time.UTC
	}

	// タイムゾーン名とオフセットを取得
	_, offset := time.Now().In(loc).Zone()
	offsetStr := formatTimezoneOffset(offset)

	// タイムゾーンの略称を取得
	tzName, _ := time.Now().In(loc).Zone()

	var b strings.Builder
	b.WriteString("BEGIN:VTIMEZONE\r\n")
	b.WriteString(fmt.Sprintf("TZID:%s\r\n", c.TimeZone))
	b.WriteString("BEGIN:STANDARD\r\n")
	b.WriteString(fmt.Sprintf("TZOFFSETFROM:%s\r\n", offsetStr))
	b.WriteString(fmt.Sprintf("TZOFFSETTO:%s\r\n", offsetStr))
	b.WriteString(fmt.Sprintf("TZNAME:%s\r\n", tzName))
	b.WriteString("DTSTART:19700101T000000\r\n")
	b.WriteString("END:STANDARD\r\n")
	b.WriteString("END:VTIMEZONE\r\n")

	return b.String()
}

// generateVEvent はVEVENTコンポーネントを生成します
func (c *Calendar) generateVEvent(event Event) string {
	var b strings.Builder

	b.WriteString("BEGIN:VEVENT\r\n")
	b.WriteString(fmt.Sprintf("UID:%s\r\n", event.UID))

	if event.AllDay {
		// 終日イベント（Date形式）
		b.WriteString(fmt.Sprintf("DTSTART;VALUE=DATE:%s\r\n", formatDate(event.Start)))
		b.WriteString(fmt.Sprintf("DTEND;VALUE=DATE:%s\r\n", formatDate(event.End)))
	} else {
		// 時刻指定イベント（DateTime形式）
		b.WriteString(fmt.Sprintf("DTSTART;TZID=%s:%s\r\n", c.TimeZone, formatDateTime(event.Start)))
		b.WriteString(fmt.Sprintf("DTEND;TZID=%s:%s\r\n", c.TimeZone, formatDateTime(event.End)))
	}

	b.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", escapeText(event.Summary)))
	b.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", escapeText(event.Description)))
	b.WriteString("END:VEVENT\r\n")

	return b.String()
}

// formatTimezoneOffset は秒単位のオフセットをiCalendar形式（例: +0900）に変換します
func formatTimezoneOffset(offsetSeconds int) string {
	sign := "+"
	if offsetSeconds < 0 {
		sign = "-"
		offsetSeconds = -offsetSeconds
	}

	hours := offsetSeconds / 3600
	minutes := (offsetSeconds % 3600) / 60

	return fmt.Sprintf("%s%02d%02d", sign, hours, minutes)
}

// formatDate は日付をiCalendar形式（YYYYMMDD）に変換します
func formatDate(t time.Time) string {
	return t.Format("20060102")
}

// formatDateTime は日時をiCalendar形式（YYYYMMDDTHHMMSS）に変換します
func formatDateTime(t time.Time) string {
	return t.Format("20060102T150405")
}

// escapeText はiCalendar形式のテキストをエスケープします
// RFC 5545 Section 3.3.11 に従い、バックスラッシュ、セミコロン、カンマ、改行をエスケープします
func escapeText(s string) string {
	// バックスラッシュを最初にエスケープ
	s = strings.ReplaceAll(s, "\\", "\\\\")
	// セミコロンとカンマをエスケープ
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, ",", "\\,")
	// 改行をエスケープ（\n を \\n に変換）
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "")

	return s
}
