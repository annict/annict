package email

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/model"
)

// TestTwoFactorAuthNotificationSender_Sendは、通知の種類とロケールに応じた件名と本文で送り、
// どの種類でも心当たりが無いときの連絡先を書くことを検証する。
func TestTwoFactorAuthNotificationSender_Send(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		locale      string
		event       model.TwoFactorAuthNotificationEvent
		wantSubject string
		wantBody    string
	}{
		{
			name:        "日本語の有効化",
			locale:      "ja",
			event:       model.TwoFactorAuthNotificationEventEnabled,
			wantSubject: "Annict - 二要素認証を有効にしました",
			wantBody:    "Annictのアカウントで二要素認証を有効にしました。",
		},
		{
			name:        "日本語の本人による無効化",
			locale:      "ja",
			event:       model.TwoFactorAuthNotificationEventDisabled,
			wantSubject: "Annict - 二要素認証を無効にしました",
			wantBody:    "Annictのアカウントで二要素認証を無効にしました。",
		},
		{
			name:        "日本語の管理者による無効化",
			locale:      "ja",
			event:       model.TwoFactorAuthNotificationEventDisabledByAdmin,
			wantSubject: "Annict - 管理者が二要素認証を無効にしました",
			wantBody:    "管理者がAnnictのアカウントの二要素認証を無効にしました。",
		},
		{
			name:        "英語の有効化",
			locale:      "en",
			event:       model.TwoFactorAuthNotificationEventEnabled,
			wantSubject: "Annict - Two-factor authentication has been turned on",
			wantBody:    "Two-factor authentication has been turned on for your Annict account.",
		},
		{
			name:        "英語の管理者による無効化",
			locale:      "en",
			event:       model.TwoFactorAuthNotificationEventDisabledByAdmin,
			wantSubject: "Annict - An administrator has turned off two-factor authentication",
			wantBody:    "an administrator has turned off two-factor authentication for your Annict account.",
		},
		{
			name:        "未知のロケールは日本語にフォールバックする",
			locale:      "other",
			event:       model.TwoFactorAuthNotificationEventDisabled,
			wantSubject: "Annict - 二要素認証を無効にしました",
			wantBody:    "Annictのアカウントで二要素認証を無効にしました。",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			noop := NewNoopSender()
			ctx := context.Background()
			if err := NewTwoFactorAuthNotificationSender(noop).Send(ctx, "test@example.com", tt.locale, tt.event); err != nil {
				t.Fatalf("Send()のエラー = %v", err)
			}

			if len(noop.SentEmails) != 1 {
				t.Fatalf("SentEmailsの件数 = %d、期待値 = 1", len(noop.SentEmails))
			}
			sent := noop.SentEmails[0]
			if sent.To != "test@example.com" {
				t.Errorf("To = %q、期待値 = %q", sent.To, "test@example.com")
			}
			if sent.Subject != tt.wantSubject {
				t.Errorf("Subject = %q、期待値 = %q", sent.Subject, tt.wantSubject)
			}

			bodies := map[string]*bytes.Buffer{"HTMLBody": {}, "TextBody": {}}
			if err := sent.HTMLBody.Render(ctx, bodies["HTMLBody"]); err != nil {
				t.Fatalf("HTMLBody.Render()のエラー = %v", err)
			}
			if err := sent.TextBody.Render(ctx, bodies["TextBody"]); err != nil {
				t.Fatalf("TextBody.Render()のエラー = %v", err)
			}
			for name, body := range bodies {
				if !strings.Contains(body.String(), tt.wantBody) {
					t.Errorf("%sに %qが含まれていない: %s", name, tt.wantBody, body.String())
				}
				if !strings.Contains(body.String(), "hello@annict.com") {
					t.Errorf("%sに問い合わせ先が含まれていない", name)
				}
			}
		})
	}
}

// TestTwoFactorAuthNotificationSender_Send_UnknownEventは、未知の通知の種類では送らずにエラーを返すことを検証する。
func TestTwoFactorAuthNotificationSender_Send_UnknownEvent(t *testing.T) {
	t.Parallel()

	noop := NewNoopSender()
	if err := NewTwoFactorAuthNotificationSender(noop).Send(context.Background(), "test@example.com", "ja", "unknown"); err == nil {
		t.Error("Send()のエラー = nil、エラーを期待")
	}
	if len(noop.SentEmails) != 0 {
		t.Errorf("SentEmailsの件数 = %d、期待値 = 0", len(noop.SentEmails))
	}
}
