package mail

import "context"

// NoopSender はメールを送信しないダミー実装（テスト用）
type NoopSender struct {
	SentEmails    []SendInput
	SentRawEmails []SendRawInput
}

// NewNoopSender は新しいNoopSenderを作成します
func NewNoopSender() *NoopSender {
	return &NoopSender{
		SentEmails:    make([]SendInput, 0),
		SentRawEmails: make([]SendRawInput, 0),
	}
}

// Send はメールを送信せず、記録のみ行います
func (s *NoopSender) Send(_ context.Context, input SendInput) error {
	s.SentEmails = append(s.SentEmails, input)
	return nil
}

// SendRaw はレンダリング済みメールを送信せず、記録のみ行います
func (s *NoopSender) SendRaw(_ context.Context, input SendRawInput) error {
	s.SentRawEmails = append(s.SentRawEmails, input)
	return nil
}

// Reset は送信記録をクリアします
func (s *NoopSender) Reset() {
	s.SentEmails = make([]SendInput, 0)
	s.SentRawEmails = make([]SendRawInput, 0)
}
