package mail

import (
	"context"
	"testing"
)

func TestNoopSender_ImplementsMailSender(t *testing.T) {
	t.Parallel()

	var _ MailSender = (*NoopSender)(nil)
}

func TestNewNoopSender(t *testing.T) {
	t.Parallel()

	sender := NewNoopSender()

	if sender == nil {
		t.Fatal("NewNoopSender returned nil")
	}

	if len(sender.SentEmails) != 0 {
		t.Errorf("SentEmails length = %d, want 0", len(sender.SentEmails))
	}
}

func TestNoopSender_SendMultipartEmail(t *testing.T) {
	t.Parallel()

	sender := NewNoopSender()
	ctx := context.Background()

	err := sender.SendMultipartEmail(ctx, "test@example.com", "件名", "テキスト本文", "<p>HTML本文</p>")
	if err != nil {
		t.Fatalf("SendMultipartEmail() error = %v", err)
	}

	if len(sender.SentEmails) != 1 {
		t.Fatalf("SentEmails length = %d, want 1", len(sender.SentEmails))
	}

	email := sender.SentEmails[0]
	if email.To != "test@example.com" {
		t.Errorf("To = %q, want %q", email.To, "test@example.com")
	}
	if email.Subject != "件名" {
		t.Errorf("Subject = %q, want %q", email.Subject, "件名")
	}
	if email.Text != "テキスト本文" {
		t.Errorf("Text = %q, want %q", email.Text, "テキスト本文")
	}
	if email.HTML != "<p>HTML本文</p>" {
		t.Errorf("HTML = %q, want %q", email.HTML, "<p>HTML本文</p>")
	}
}

func TestNoopSender_SendMultipartEmail_Multiple(t *testing.T) {
	t.Parallel()

	sender := NewNoopSender()
	ctx := context.Background()

	for i := range 3 {
		err := sender.SendMultipartEmail(ctx, "test@example.com", "件名", "本文", "<p>本文</p>")
		if err != nil {
			t.Fatalf("SendMultipartEmail() call %d error = %v", i, err)
		}
	}

	if len(sender.SentEmails) != 3 {
		t.Errorf("SentEmails length = %d, want 3", len(sender.SentEmails))
	}
}

func TestNoopSender_Reset(t *testing.T) {
	t.Parallel()

	sender := NewNoopSender()
	ctx := context.Background()

	_ = sender.SendMultipartEmail(ctx, "test@example.com", "件名", "本文", "<p>本文</p>")
	_ = sender.SendMultipartEmail(ctx, "test2@example.com", "件名2", "本文2", "<p>本文2</p>")

	if len(sender.SentEmails) != 2 {
		t.Fatalf("SentEmails length = %d, want 2", len(sender.SentEmails))
	}

	sender.Reset()

	if len(sender.SentEmails) != 0 {
		t.Errorf("SentEmails length after Reset = %d, want 0", len(sender.SentEmails))
	}
}
