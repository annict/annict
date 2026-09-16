package email

import (
	"context"
	"testing"
)

// ResendSenderがSenderインターフェースを実装していることを確認
var _ Sender = (*ResendSender)(nil)

// NoopSenderがSenderインターフェースを実装していることを確認
var _ Sender = (*NoopSender)(nil)

func TestNewResendSender(t *testing.T) {
	t.Parallel()

	sender := NewResendSender("test-api-key", "test@example.com", "Annict")

	if sender.client == nil {
		t.Error("clientがnilだった")
	}
	if sender.fromEmail != "test@example.com" {
		t.Errorf("fromEmail = %s、期待値 = test@example.com", sender.fromEmail)
	}
	if sender.fromName != "Annict" {
		t.Errorf("fromName = %s、期待値 = Annict", sender.fromName)
	}
}

func TestResendSender_from(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		fromEmail string
		fromName  string
		want      string
	}{
		{
			name:      "fromNameが設定されている場合",
			fromEmail: "noreply@annict.com",
			fromName:  "Annict",
			want:      "Annict <noreply@annict.com>",
		},
		{
			name:      "fromNameが空の場合",
			fromEmail: "noreply@annict.com",
			fromName:  "",
			want:      "noreply@annict.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sender := NewResendSender("test-api-key", tt.fromEmail, tt.fromName)
			got := sender.from()
			if got != tt.want {
				t.Errorf("from() = %s、期待値 = %s", got, tt.want)
			}
		})
	}
}

func TestNewNoopSender(t *testing.T) {
	t.Parallel()

	sender := NewNoopSender()

	if sender.SentEmails == nil {
		t.Error("SentEmailsがnilだった")
	}
	if len(sender.SentEmails) != 0 {
		t.Errorf("SentEmailsの件数 = %d、期待値 = 0", len(sender.SentEmails))
	}
}

func TestNoopSender_Send(t *testing.T) {
	t.Parallel()

	sender := NewNoopSender()
	ctx := context.Background()

	input1 := SendInput{
		To:      "user1@example.com",
		Subject: "Subject 1",
	}
	err := sender.Send(ctx, input1)
	if err != nil {
		t.Fatalf("Send()のエラー = %v", err)
	}

	if len(sender.SentEmails) != 1 {
		t.Fatalf("SentEmailsの件数 = %d、期待値 = 1", len(sender.SentEmails))
	}
	if sender.SentEmails[0].To != "user1@example.com" {
		t.Errorf("SentEmails[0].To = %s、期待値 = user1@example.com", sender.SentEmails[0].To)
	}

	input2 := SendInput{
		To:      "user2@example.com",
		Subject: "Subject 2",
	}
	err = sender.Send(ctx, input2)
	if err != nil {
		t.Fatalf("Send()のエラー = %v", err)
	}

	if len(sender.SentEmails) != 2 {
		t.Fatalf("SentEmailsの件数 = %d、期待値 = 2", len(sender.SentEmails))
	}
	if sender.SentEmails[1].To != "user2@example.com" {
		t.Errorf("SentEmails[1].To = %s、期待値 = user2@example.com", sender.SentEmails[1].To)
	}
}

func TestNoopSender_Reset(t *testing.T) {
	t.Parallel()

	sender := NewNoopSender()
	ctx := context.Background()

	err := sender.Send(ctx, SendInput{To: "test@example.com", Subject: "Test"})
	if err != nil {
		t.Fatalf("Send()のエラー = %v", err)
	}
	if len(sender.SentEmails) != 1 {
		t.Fatalf("SentEmailsの件数 = %d、期待値 = 1", len(sender.SentEmails))
	}

	sender.Reset()

	if len(sender.SentEmails) != 0 {
		t.Errorf("Reset()後のSentEmailsの件数 = %d、期待値 = 0", len(sender.SentEmails))
	}
}
