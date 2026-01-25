package ical

import (
	"strings"
	"testing"
	"time"
)

func TestCalendar_ToICS(t *testing.T) {
	t.Parallel()

	// Asia/Tokyoタイムゾーン
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatalf("failed to load timezone: %v", err)
	}

	tests := []struct {
		name     string
		calendar Calendar
		contains []string // 出力に含まれるべき文字列
	}{
		{
			name: "基本的なカレンダー生成",
			calendar: Calendar{
				TimeZone: "Asia/Tokyo",
				CalName:  "Annict@testuser",
				Events:   []Event{},
			},
			contains: []string{
				"BEGIN:VCALENDAR\r\n",
				"VERSION:2.0\r\n",
				"PRODID:-//Annict//Annict Calendar//EN\r\n",
				"CALSCALE:GREGORIAN\r\n",
				"METHOD:PUBLISH\r\n",
				"X-WR-TIMEZONE:Asia/Tokyo\r\n",
				"X-WR-CALNAME:Annict@testuser\r\n",
				"END:VCALENDAR\r\n",
			},
		},
		{
			name: "時刻指定イベント（放送枠）",
			calendar: Calendar{
				TimeZone: "Asia/Tokyo",
				CalName:  "Annict@testuser",
				Events: []Event{
					{
						UID:         "slot-12345@annict.com",
						Summary:     "テストアニメ 第1話 サブタイトル (TOKYO MX)",
						Description: "テストアニメ 第1話 サブタイトル\nhttps://annict.com/works/123/episodes/456",
						Start:       time.Date(2025, 1, 20, 1, 0, 0, 0, jst),
						End:         time.Date(2025, 1, 20, 1, 30, 0, 0, jst),
						AllDay:      false,
					},
				},
			},
			contains: []string{
				"BEGIN:VEVENT\r\n",
				"UID:slot-12345@annict.com\r\n",
				"DTSTART;TZID=Asia/Tokyo:20250120T010000\r\n",
				"DTEND;TZID=Asia/Tokyo:20250120T013000\r\n",
				"SUMMARY:テストアニメ 第1話 サブタイトル (TOKYO MX)\r\n",
				"DESCRIPTION:テストアニメ 第1話 サブタイトル\\nhttps://annict.com/works/123/episodes/456\r\n",
				"END:VEVENT\r\n",
			},
		},
		{
			name: "終日イベント（作品放送開始日）",
			calendar: Calendar{
				TimeZone: "Asia/Tokyo",
				CalName:  "Annict@testuser",
				Events: []Event{
					{
						UID:         "work-789@annict.com",
						Summary:     "新作アニメタイトル",
						Description: "新作アニメタイトル\nhttps://annict.com/works/789",
						Start:       time.Date(2025, 4, 1, 0, 0, 0, 0, jst),
						End:         time.Date(2025, 4, 2, 0, 0, 0, 0, jst),
						AllDay:      true,
					},
				},
			},
			contains: []string{
				"BEGIN:VEVENT\r\n",
				"UID:work-789@annict.com\r\n",
				"DTSTART;VALUE=DATE:20250401\r\n",
				"DTEND;VALUE=DATE:20250402\r\n",
				"SUMMARY:新作アニメタイトル\r\n",
				"DESCRIPTION:新作アニメタイトル\\nhttps://annict.com/works/789\r\n",
				"END:VEVENT\r\n",
			},
		},
		{
			name: "複数イベント",
			calendar: Calendar{
				TimeZone: "Asia/Tokyo",
				CalName:  "Annict@testuser",
				Events: []Event{
					{
						UID:     "slot-1@annict.com",
						Summary: "アニメA",
						Start:   time.Date(2025, 1, 20, 1, 0, 0, 0, jst),
						End:     time.Date(2025, 1, 20, 1, 30, 0, 0, jst),
						AllDay:  false,
					},
					{
						UID:     "slot-2@annict.com",
						Summary: "アニメB",
						Start:   time.Date(2025, 1, 21, 23, 0, 0, 0, jst),
						End:     time.Date(2025, 1, 21, 23, 30, 0, 0, jst),
						AllDay:  false,
					},
				},
			},
			contains: []string{
				"UID:slot-1@annict.com\r\n",
				"SUMMARY:アニメA\r\n",
				"UID:slot-2@annict.com\r\n",
				"SUMMARY:アニメB\r\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.calendar.ToICS()

			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("ToICS() does not contain %q\nGot:\n%s", want, got)
				}
			}
		})
	}
}

func TestCalendar_ToICS_VTimezone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		timezone string
		contains []string
	}{
		{
			name:     "Asia/Tokyo",
			timezone: "Asia/Tokyo",
			contains: []string{
				"BEGIN:VTIMEZONE\r\n",
				"TZID:Asia/Tokyo\r\n",
				"BEGIN:STANDARD\r\n",
				"TZOFFSETFROM:+0900\r\n",
				"TZOFFSETTO:+0900\r\n",
				"TZNAME:JST\r\n",
				"DTSTART:19700101T000000\r\n",
				"END:STANDARD\r\n",
				"END:VTIMEZONE\r\n",
			},
		},
		{
			name:     "UTC",
			timezone: "UTC",
			contains: []string{
				"TZID:UTC\r\n",
				"TZOFFSETFROM:+0000\r\n",
				"TZOFFSETTO:+0000\r\n",
			},
		},
		{
			name:     "America/New_York",
			timezone: "America/New_York",
			contains: []string{
				"TZID:America/New_York\r\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cal := Calendar{
				TimeZone: tt.timezone,
				CalName:  "Test",
				Events:   []Event{},
			}

			got := cal.ToICS()

			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("ToICS() does not contain %q\nGot:\n%s", want, got)
				}
			}
		})
	}
}

func TestEscapeText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "エスケープ不要",
			input: "Simple text",
			want:  "Simple text",
		},
		{
			name:  "改行をエスケープ",
			input: "Line1\nLine2",
			want:  "Line1\\nLine2",
		},
		{
			name:  "バックスラッシュをエスケープ",
			input: "Path\\to\\file",
			want:  "Path\\\\to\\\\file",
		},
		{
			name:  "セミコロンをエスケープ",
			input: "Item1;Item2",
			want:  "Item1\\;Item2",
		},
		{
			name:  "カンマをエスケープ",
			input: "Item1,Item2",
			want:  "Item1\\,Item2",
		},
		{
			name:  "複合エスケープ",
			input: "Test\\n;,",
			want:  "Test\\\\n\\;\\,",
		},
		{
			name:  "日本語テキスト",
			input: "テストアニメ 第1話\nサブタイトル",
			want:  "テストアニメ 第1話\\nサブタイトル",
		},
		{
			name:  "CRLFをLFに正規化",
			input: "Line1\r\nLine2",
			want:  "Line1\\nLine2",
		},
		{
			name:  "URLはエスケープ不要",
			input: "https://annict.com/works/123",
			want:  "https://annict.com/works/123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := escapeText(tt.input)
			if got != tt.want {
				t.Errorf("escapeText(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatTimezoneOffset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		offsetSeconds int
		want          string
	}{
		{
			name:          "UTC",
			offsetSeconds: 0,
			want:          "+0000",
		},
		{
			name:          "JST (+9:00)",
			offsetSeconds: 9 * 3600,
			want:          "+0900",
		},
		{
			name:          "EST (-5:00)",
			offsetSeconds: -5 * 3600,
			want:          "-0500",
		},
		{
			name:          "IST (+5:30)",
			offsetSeconds: 5*3600 + 30*60,
			want:          "+0530",
		},
		{
			name:          "NPT (+5:45)",
			offsetSeconds: 5*3600 + 45*60,
			want:          "+0545",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := formatTimezoneOffset(tt.offsetSeconds)
			if got != tt.want {
				t.Errorf("formatTimezoneOffset(%d) = %q, want %q", tt.offsetSeconds, got, tt.want)
			}
		})
	}
}

func TestFormatDate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "標準的な日付",
			time: time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
			want: "20250401",
		},
		{
			name: "年初",
			time: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			want: "20250101",
		},
		{
			name: "年末",
			time: time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			want: "20251231",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := formatDate(tt.time)
			if got != tt.want {
				t.Errorf("formatDate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatDateTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "標準的な日時",
			time: time.Date(2025, 1, 20, 1, 0, 0, 0, time.UTC),
			want: "20250120T010000",
		},
		{
			name: "深夜帯",
			time: time.Date(2025, 1, 20, 1, 30, 0, 0, time.UTC),
			want: "20250120T013000",
		},
		{
			name: "正午",
			time: time.Date(2025, 1, 20, 12, 0, 0, 0, time.UTC),
			want: "20250120T120000",
		},
		{
			name: "秒も含む",
			time: time.Date(2025, 1, 20, 23, 59, 59, 0, time.UTC),
			want: "20250120T235959",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := formatDateTime(tt.time)
			if got != tt.want {
				t.Errorf("formatDateTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCalendar_ToICS_CompleteOutput(t *testing.T) {
	t.Parallel()

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatalf("failed to load timezone: %v", err)
	}

	cal := Calendar{
		TimeZone: "Asia/Tokyo",
		CalName:  "Annict@testuser",
		Events: []Event{
			{
				UID:         "slot-12345@annict.com",
				Summary:     "作品タイトル 第1話 サブタイトル (TOKYO MX)",
				Description: "作品タイトル 第1話 サブタイトル\nhttps://annict.com/works/123/episodes/456",
				Start:       time.Date(2025, 1, 20, 1, 0, 0, 0, jst),
				End:         time.Date(2025, 1, 20, 1, 30, 0, 0, jst),
				AllDay:      false,
			},
			{
				UID:         "work-789@annict.com",
				Summary:     "新作アニメタイトル",
				Description: "新作アニメタイトル\nhttps://annict.com/works/789",
				Start:       time.Date(2025, 4, 1, 0, 0, 0, 0, jst),
				End:         time.Date(2025, 4, 2, 0, 0, 0, 0, jst),
				AllDay:      true,
			},
		},
	}

	got := cal.ToICS()

	// 出力がBEGIN:VCALENDARで始まりEND:VCALENDARで終わることを確認
	if !strings.HasPrefix(got, "BEGIN:VCALENDAR\r\n") {
		t.Error("ToICS() should start with BEGIN:VCALENDAR")
	}
	if !strings.HasSuffix(got, "END:VCALENDAR\r\n") {
		t.Error("ToICS() should end with END:VCALENDAR")
	}

	// VTIMEZONEが含まれることを確認
	if !strings.Contains(got, "BEGIN:VTIMEZONE") {
		t.Error("ToICS() should contain VTIMEZONE")
	}

	// 2つのVEVENTが含まれることを確認
	eventCount := strings.Count(got, "BEGIN:VEVENT")
	if eventCount != 2 {
		t.Errorf("ToICS() should contain 2 VEVENTs, got %d", eventCount)
	}
}

func TestCalendar_ToICS_EmptyEvents(t *testing.T) {
	t.Parallel()

	cal := Calendar{
		TimeZone: "Asia/Tokyo",
		CalName:  "Annict@testuser",
		Events:   []Event{},
	}

	got := cal.ToICS()

	// イベントがなくてもカレンダーは有効
	if !strings.Contains(got, "BEGIN:VCALENDAR") {
		t.Error("ToICS() should contain BEGIN:VCALENDAR even with no events")
	}
	if strings.Contains(got, "BEGIN:VEVENT") {
		t.Error("ToICS() should not contain VEVENT when no events")
	}
}
