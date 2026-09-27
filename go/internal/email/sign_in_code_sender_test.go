package email

import (
	"bytes"
	"context"
	"testing"
)

func TestSignInCodeSender_Send(t *testing.T) {
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
			sender := NewSignInCodeSender(noop)

			ctx := context.Background()
			err := sender.Send(ctx, "test@example.com", "123456", tt.locale)
			if err != nil {
				t.Fatalf("Send()のエラー = %v", err)
			}

			if len(noop.SentEmails) != 1 {
				t.Fatalf("SentEmailsの件数 = %d、期待値 = 1", len(noop.SentEmails))
			}

			sent := noop.SentEmails[0]
			if sent.To != "test@example.com" {
				t.Errorf("To = %s、期待値 = test@example.com", sent.To)
			}
			if sent.Subject == "" {
				t.Error("Subjectが空です")
			}

			if sent.HTMLBody == nil {
				t.Fatal("HTMLBodyがnilです")
			}
			var htmlBuf bytes.Buffer
			if err := sent.HTMLBody.Render(ctx, &htmlBuf); err != nil {
				t.Fatalf("HTMLBody.Render()のエラー = %v", err)
			}
			if htmlBuf.Len() == 0 {
				t.Error("HTMLBodyの出力が空です")
			}

			if sent.TextBody == nil {
				t.Fatal("TextBodyがnilです")
			}
			var textBuf bytes.Buffer
			if err := sent.TextBody.Render(ctx, &textBuf); err != nil {
				t.Fatalf("TextBody.Render()のエラー = %v", err)
			}
			if textBuf.Len() == 0 {
				t.Error("TextBodyの出力が空です")
			}
		})
	}
}
