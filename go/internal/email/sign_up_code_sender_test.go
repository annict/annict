package email

import (
	"bytes"
	"context"
	"testing"
)

func TestSignUpCodeSender_Send(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		locale string
	}{
		{"Japanese", "ja"},
		{"English", "en"},
		{"UnknownLocale falls back to Ja", "fr"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			noop := NewNoopSender()
			sender := NewSignUpCodeSender(noop)

			ctx := context.Background()
			err := sender.Send(ctx, "test@example.com", "123456", tt.locale)
			if err != nil {
				t.Fatalf("Send() error = %v", err)
			}

			if len(noop.SentEmails) != 1 {
				t.Fatalf("SentEmails length = %d, want 1", len(noop.SentEmails))
			}

			sent := noop.SentEmails[0]
			if sent.To != "test@example.com" {
				t.Errorf("To = %s, want test@example.com", sent.To)
			}
			if sent.Subject == "" {
				t.Error("Subject が空です")
			}

			if sent.HTMLBody == nil {
				t.Fatal("HTMLBody が nil です")
			}
			var htmlBuf bytes.Buffer
			if err := sent.HTMLBody.Render(ctx, &htmlBuf); err != nil {
				t.Fatalf("HTMLBody.Render() error = %v", err)
			}
			if htmlBuf.Len() == 0 {
				t.Error("HTMLBody の出力が空です")
			}

			if sent.TextBody == nil {
				t.Fatal("TextBody が nil です")
			}
			var textBuf bytes.Buffer
			if err := sent.TextBody.Render(ctx, &textBuf); err != nil {
				t.Fatalf("TextBody.Render() error = %v", err)
			}
			if textBuf.Len() == 0 {
				t.Error("TextBody の出力が空です")
			}
		})
	}
}
