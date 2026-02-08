package mail

import "context"

// SentEmail はNoopSenderが記録する送信済みメールの情報
type SentEmail struct {
	To      string
	Subject string
	Text    string
	HTML    string
}

// NoopSender はメールを送信しないダミー実装（テスト用）
type NoopSender struct {
	SentEmails []SentEmail
}

// NewNoopSender は新しいNoopSenderを作成します
func NewNoopSender() *NoopSender {
	return &NoopSender{
		SentEmails: make([]SentEmail, 0),
	}
}

// SendMultipartEmail はメールを送信せず、記録のみ行います
func (s *NoopSender) SendMultipartEmail(_ context.Context, to, subject, text, html string) error {
	s.SentEmails = append(s.SentEmails, SentEmail{
		To:      to,
		Subject: subject,
		Text:    text,
		HTML:    html,
	})
	return nil
}

// Reset は送信記録をクリアします
func (s *NoopSender) Reset() {
	s.SentEmails = make([]SentEmail, 0)
}
