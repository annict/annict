package mail

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/a-h/templ"
)

// testComponent はテスト用のtempl.Componentを作成します
func testComponent(content string) templ.Component {
	return templ.ComponentFunc(func(_ context.Context, w io.Writer) error {
		_, err := io.WriteString(w, content)
		return err
	})
}

func TestNoopSender_ImplementsSender(t *testing.T) {
	t.Parallel()

	var _ Sender = (*NoopSender)(nil)
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

func TestNoopSender_Send(t *testing.T) {
	t.Parallel()

	sender := NewNoopSender()
	ctx := context.Background()

	htmlBody := testComponent("<p>HTML本文</p>")
	textBody := testComponent("テキスト本文")

	err := sender.Send(ctx, SendInput{
		To:       "test@example.com",
		Subject:  "件名",
		HTMLBody: htmlBody,
		TextBody: textBody,
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
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

	// templ.Componentをレンダリングして内容を検証
	var htmlBuf bytes.Buffer
	if err := email.HTMLBody.Render(ctx, &htmlBuf); err != nil {
		t.Fatalf("HTMLBody.Render() error = %v", err)
	}
	if htmlBuf.String() != "<p>HTML本文</p>" {
		t.Errorf("HTMLBody = %q, want %q", htmlBuf.String(), "<p>HTML本文</p>")
	}

	var textBuf bytes.Buffer
	if err := email.TextBody.Render(ctx, &textBuf); err != nil {
		t.Fatalf("TextBody.Render() error = %v", err)
	}
	if textBuf.String() != "テキスト本文" {
		t.Errorf("TextBody = %q, want %q", textBuf.String(), "テキスト本文")
	}
}

func TestNoopSender_Send_HTMLOnly(t *testing.T) {
	t.Parallel()

	sender := NewNoopSender()
	ctx := context.Background()

	err := sender.Send(ctx, SendInput{
		To:       "test@example.com",
		Subject:  "件名",
		HTMLBody: testComponent("<p>HTML</p>"),
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	email := sender.SentEmails[0]
	if email.TextBody != nil {
		t.Error("TextBody should be nil")
	}
}

func TestNoopSender_Send_Multiple(t *testing.T) {
	t.Parallel()

	sender := NewNoopSender()
	ctx := context.Background()

	for i := range 3 {
		err := sender.Send(ctx, SendInput{
			To:       "test@example.com",
			Subject:  "件名",
			HTMLBody: testComponent("<p>本文</p>"),
			TextBody: testComponent("本文"),
		})
		if err != nil {
			t.Fatalf("Send() call %d error = %v", i, err)
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

	_ = sender.Send(ctx, SendInput{
		To:       "test@example.com",
		Subject:  "件名",
		HTMLBody: testComponent("<p>本文</p>"),
	})
	_ = sender.Send(ctx, SendInput{
		To:       "test2@example.com",
		Subject:  "件名2",
		HTMLBody: testComponent("<p>本文2</p>"),
	})

	if len(sender.SentEmails) != 2 {
		t.Fatalf("SentEmails length = %d, want 2", len(sender.SentEmails))
	}

	sender.Reset()

	if len(sender.SentEmails) != 0 {
		t.Errorf("SentEmails length after Reset = %d, want 0", len(sender.SentEmails))
	}
}
