package mail

import (
	"testing"
)

func TestResendClient_from(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		fromEmail string
		fromName  string
		want      string
	}{
		{
			name:      "fromNameあり",
			fromEmail: "noreply@example.com",
			fromName:  "Annict",
			want:      "Annict <noreply@example.com>",
		},
		{
			name:      "fromNameなし",
			fromEmail: "noreply@example.com",
			fromName:  "",
			want:      "noreply@example.com",
		},
		{
			name:      "fromNameが空白のみ",
			fromEmail: "noreply@example.com",
			fromName:  "  ",
			want:      "   <noreply@example.com>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &ResendClient{
				fromEmail: tt.fromEmail,
				fromName:  tt.fromName,
			}

			got := client.from()
			if got != tt.want {
				t.Errorf("from() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewResendClient(t *testing.T) {
	t.Parallel()

	apiKey := "test-api-key"
	fromEmail := "noreply@example.com"
	fromName := "Annict"

	client := NewResendClient(apiKey, fromEmail, fromName)

	if client == nil {
		t.Fatal("NewResendClient returned nil")
	}

	if client.fromEmail != fromEmail {
		t.Errorf("fromEmail = %q, want %q", client.fromEmail, fromEmail)
	}

	if client.fromName != fromName {
		t.Errorf("fromName = %q, want %q", client.fromName, fromName)
	}

	if client.client == nil {
		t.Error("client.client is nil")
	}
}

func TestResendClient_ImplementsMailSender(t *testing.T) {
	t.Parallel()

	// ResendClientがMailSenderインターフェースを実装していることを確認
	var _ MailSender = (*ResendClient)(nil)
}
