package email

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// TestAccountDeletionSender_Sendは、ロケールに応じた件名と本文で退会完了メールを送り、
// Stripeのサポーターにだけ解約の案内を書くことをテストする。
func TestAccountDeletionSender_Send(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		locale            string
		isStripeSupporter bool
		wantSubject       string
		wantBody          string
		cancelNotice      string
	}{
		{
			name:         "日本語のサポーターでないユーザー",
			locale:       "ja",
			wantSubject:  "Annict - 退会を受け付けました",
			wantBody:     "Annictの退会を受け付けました。",
			cancelNotice: "サブスクリプションは即時に解約します",
		},
		{
			name:              "日本語のStripeのサポーター",
			locale:            "ja",
			isStripeSupporter: true,
			wantSubject:       "Annict - 退会を受け付けました",
			wantBody:          "Annictの退会を受け付けました。",
			cancelNotice:      "サブスクリプションは即時に解約します",
		},
		{
			name:         "英語のサポーターでないユーザー",
			locale:       "en",
			wantSubject:  "Annict - Your account deletion request has been received",
			wantBody:     "We have received your request to delete your Annict account.",
			cancelNotice: "subscription will be canceled immediately",
		},
		{
			name:              "英語のStripeのサポーター",
			locale:            "en",
			isStripeSupporter: true,
			wantSubject:       "Annict - Your account deletion request has been received",
			wantBody:          "We have received your request to delete your Annict account.",
			cancelNotice:      "subscription will be canceled immediately",
		},
		{
			name:         "未知のロケールは日本語にフォールバックする",
			locale:       "other",
			wantSubject:  "Annict - 退会を受け付けました",
			wantBody:     "Annictの退会を受け付けました。",
			cancelNotice: "サブスクリプションは即時に解約します",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			noop := NewNoopSender()
			sender := NewAccountDeletionSender(noop)

			ctx := context.Background()
			if err := sender.Send(ctx, "test@example.com", tt.locale, tt.isStripeSupporter); err != nil {
				t.Fatalf("Send()のエラー = %v", err)
			}

			if len(noop.SentEmails) != 1 {
				t.Fatalf("SentEmailsの件数 = %d、期待値 = 1", len(noop.SentEmails))
			}
			sent := noop.SentEmails[0]
			if sent.To != "test@example.com" {
				t.Errorf("To = %s、期待値 = test@example.com", sent.To)
			}
			if sent.Subject != tt.wantSubject {
				t.Errorf("Subject = %q、期待値 = %q", sent.Subject, tt.wantSubject)
			}

			bodies := map[string]*bytes.Buffer{"HTMLBody": {}, "TextBody": {}}
			if sent.HTMLBody == nil || sent.TextBody == nil {
				t.Fatal("HTMLBodyまたはTextBodyがnilです")
			}
			if err := sent.HTMLBody.Render(ctx, bodies["HTMLBody"]); err != nil {
				t.Fatalf("HTMLBody.Render()のエラー = %v", err)
			}
			if err := sent.TextBody.Render(ctx, bodies["TextBody"]); err != nil {
				t.Fatalf("TextBody.Render()のエラー = %v", err)
			}
			for name, body := range bodies {
				if !strings.Contains(body.String(), tt.wantBody) {
					t.Errorf("%sに %q が含まれていません: %s", name, tt.wantBody, body.String())
				}
				if !strings.Contains(body.String(), "hello@annict.com") {
					t.Errorf("%sに問い合わせ先が含まれていません: %s", name, body.String())
				}
				if got := strings.Contains(body.String(), tt.cancelNotice); got != tt.isStripeSupporter {
					t.Errorf("%sの解約の案内の有無 = %v、期待値 = %v", name, got, tt.isStripeSupporter)
				}
			}
		})
	}
}
